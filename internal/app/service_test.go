package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/core/contextgen"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/selection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockScanner struct {
	tree *scanner.FileNode
	err  error
}

func (m *mockScanner) Scan(rootPath string, config *scanner.ScanConfig) (*scanner.FileNode, error) {
	return m.tree, m.err
}

func (m *mockScanner) ScanWithProgress(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
	return m.tree, m.err
}

type mockGenerator struct {
	content string
	err     error
}

func (m *mockGenerator) Generate(tree *scanner.FileNode, selections map[string]bool, config contextgen.GenerateConfig) (string, error) {
	return m.content, m.err
}

func (m *mockGenerator) GenerateWithProgress(tree *scanner.FileNode, selections map[string]bool, config contextgen.GenerateConfig, progress func(string)) (string, error) {
	return m.content, m.err
}

func (m *mockGenerator) GenerateWithProgressEx(tree *scanner.FileNode, selections map[string]bool, config contextgen.GenerateConfig, progress func(contextgen.GenProgress)) (string, error) {
	return m.content, m.err
}

type mockProvider struct {
	name       string
	available  bool
	configured bool
	result     *llm.Result
	err        error
}

func (m *mockProvider) Name() string          { return m.name }
func (m *mockProvider) IsAvailable() bool     { return m.available }
func (m *mockProvider) IsConfigured() bool    { return m.configured }
func (m *mockProvider) ValidateConfig() error { return nil }
func (m *mockProvider) Send(ctx context.Context, content string) (*llm.Result, error) {
	return m.result, m.err
}
func (m *mockProvider) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	return m.result, m.err
}

func TestNewContextService_Default(t *testing.T) {
	svc := NewContextService()
	require.NotNil(t, svc)
	assert.NotNil(t, svc.scanner)
	assert.NotNil(t, svc.generator)
}

func TestNewContextService_WithScanner(t *testing.T) {
	mock := &mockScanner{}
	svc := NewContextService(WithScanner(mock))
	assert.Equal(t, mock, svc.scanner)
}

func TestNewContextService_WithGenerator(t *testing.T) {
	mock := &mockGenerator{}
	svc := NewContextService(WithGenerator(mock))
	assert.Equal(t, mock, svc.generator)
}

func TestDefaultContextService_Scanner(t *testing.T) {
	mock := &mockScanner{}
	svc := NewContextService(WithScanner(mock))
	assert.Equal(t, mock, svc.Scanner())
}

func TestDefaultContextService_Generate_InvalidConfig(t *testing.T) {
	svc := NewContextService()
	cfg := GenerateConfig{}

	_, err := svc.Generate(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestDefaultContextService_Generate_ScanError(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &mockScanner{err: assert.AnError}
	svc := NewContextService(WithScanner(mock))

	cfg := GenerateConfig{RootPath: tmpDir}
	_, err := svc.Generate(context.Background(), cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scan failed")
}

func TestDefaultContextService_Generate_GenerationError(t *testing.T) {
	tmpDir := t.TempDir()
	mockScan := &mockScanner{
		tree: &scanner.FileNode{Name: "root", IsDir: true, Path: tmpDir},
	}
	mockGen := &mockGenerator{err: assert.AnError}
	svc := NewContextService(WithScanner(mockScan), WithGenerator(mockGen))

	cfg := GenerateConfig{RootPath: tmpDir}
	_, err := svc.Generate(context.Background(), cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generation failed")
}

func TestDefaultContextService_Generate_EnforceLimitExceeded(t *testing.T) {
	tmpDir := t.TempDir()
	mockScan := &mockScanner{
		tree: &scanner.FileNode{Name: "root", IsDir: true, Path: tmpDir},
	}
	mockGen := &mockGenerator{content: "this is some content that exceeds the limit"}
	svc := NewContextService(WithScanner(mockScan), WithGenerator(mockGen))

	cfg := GenerateConfig{
		RootPath:     tmpDir,
		MaxSize:      10,
		EnforceLimit: true,
	}
	_, err := svc.Generate(context.Background(), cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds limit")
}

func TestDefaultContextService_Generate_Success(t *testing.T) {
	tmpDir := t.TempDir()
	mockScan := &mockScanner{
		tree: &scanner.FileNode{Name: "root", IsDir: true, Path: tmpDir},
	}
	mockGen := &mockGenerator{content: "generated content"}
	svc := NewContextService(WithScanner(mockScan), WithGenerator(mockGen))

	outputFile := filepath.Join(tmpDir, "output.md")
	cfg := GenerateConfig{
		RootPath:   tmpDir,
		OutputPath: outputFile,
	}
	result, err := svc.Generate(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, "generated content", result.Content)
	assert.Equal(t, outputFile, result.OutputPath)
	assert.Equal(t, int64(len("generated content")), result.ContentSize)

	savedContent, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, "generated content", string(savedContent))
}

func TestDefaultContextService_GenerateWithProgress(t *testing.T) {
	tmpDir := t.TempDir()
	mockScan := &mockScanner{
		tree: &scanner.FileNode{Name: "root", IsDir: true, Path: tmpDir},
	}
	mockGen := &mockGenerator{content: "content"}
	svc := NewContextService(WithScanner(mockScan), WithGenerator(mockGen))

	var progressCalls []string
	progress := func(stage, message string, current, total int64) {
		progressCalls = append(progressCalls, stage)
	}

	outputFile := filepath.Join(tmpDir, "output.md")
	cfg := GenerateConfig{
		RootPath:   tmpDir,
		OutputPath: outputFile,
	}

	result, err := svc.GenerateWithProgress(context.Background(), cfg, progress)
	require.NoError(t, err)
	assert.NotNil(t, result)

	assert.Contains(t, progressCalls, "scanning")
	assert.Contains(t, progressCalls, "generating")
	assert.Contains(t, progressCalls, "saving")
	assert.Contains(t, progressCalls, "complete")
}

func TestDefaultContextService_Generate_WithCustomSelections(t *testing.T) {
	tmpDir := t.TempDir()
	mockScan := &mockScanner{
		tree: &scanner.FileNode{Name: "root", IsDir: true, Path: tmpDir},
	}
	mockGen := &mockGenerator{content: "content"}
	svc := NewContextService(WithScanner(mockScan), WithGenerator(mockGen))

	outputFile := filepath.Join(tmpDir, "output.md")
	cfg := GenerateConfig{
		RootPath:   tmpDir,
		OutputPath: outputFile,
		Selections: map[string]bool{"/custom/path": true},
	}

	result, err := svc.Generate(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDefaultContextService_SendToLLM_Unavailable(t *testing.T) {
	svc := NewContextService()
	provider := &mockProvider{name: "test", available: false}

	_, err := svc.SendToLLM(context.Background(), "content", provider)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestDefaultContextService_SendToLLM_Success(t *testing.T) {
	svc := NewContextService()
	provider := &mockProvider{
		name:      "test",
		available: true,
		result:    &llm.Result{Response: "response"},
	}

	result, err := svc.SendToLLM(context.Background(), "content", provider)
	require.NoError(t, err)
	assert.Equal(t, "response", result.Response)
}

func TestDefaultContextService_SendToLLM_Error(t *testing.T) {
	svc := NewContextService()
	provider := &mockProvider{
		name:      "test",
		available: true,
		err:       assert.AnError,
	}

	_, err := svc.SendToLLM(context.Background(), "content", provider)
	assert.Error(t, err)
}

// capturingGenerator records the selections map it is handed so tests can assert
// which files the service resolved before generation.
type capturingGenerator struct {
	content    string
	selections map[string]bool
}

func (m *capturingGenerator) Generate(tree *scanner.FileNode, selections map[string]bool, config contextgen.GenerateConfig) (string, error) {
	m.selections = selections
	return m.content, nil
}

func (m *capturingGenerator) GenerateWithProgress(tree *scanner.FileNode, selections map[string]bool, config contextgen.GenerateConfig, progress func(string)) (string, error) {
	m.selections = selections
	return m.content, nil
}

func (m *capturingGenerator) GenerateWithProgressEx(tree *scanner.FileNode, selections map[string]bool, config contextgen.GenerateConfig, progress func(contextgen.GenProgress)) (string, error) {
	m.selections = selections
	return m.content, nil
}

// twoFileTree builds a root directory holding two non-ignored files with distinct
// absolute Path and slash-form RelPath values, rooted at tmpDir.
func twoFileTree(tmpDir string) *scanner.FileNode {
	return &scanner.FileNode{
		Name:    "root",
		Path:    tmpDir,
		RelPath: ".",
		IsDir:   true,
		Children: []*scanner.FileNode{
			{Name: "file1.go", Path: filepath.Join(tmpDir, "file1.go"), RelPath: "file1.go"},
			{Name: "file2.go", Path: filepath.Join(tmpDir, "file2.go"), RelPath: "file2.go"},
		},
	}
}

// A configured store with a saved deselection must drop exactly that file from
// the selection map the service hands the generator, leaving siblings selected.
func TestDefaultContextService_Generate_HonorsSavedDeselection(t *testing.T) {
	tmpDir := t.TempDir()
	tree := twoFileTree(tmpDir)
	file1Path := filepath.Join(tmpDir, "file1.go")
	file2Path := filepath.Join(tmpDir, "file2.go")

	store := selection.NewStore(filepath.Join(t.TempDir(), "selections.json"))
	cfg := GenerateConfig{
		RootPath:   tmpDir,
		OutputPath: filepath.Join(tmpDir, "output.md"),
	}
	require.NoError(t, store.Save(cfg.RootPath, []string{"file2.go"}))

	gen := &capturingGenerator{content: "content"}
	svc := NewContextService(
		WithScanner(&mockScanner{tree: tree}),
		WithGenerator(gen),
		WithSelectionStore(store),
	)

	_, err := svc.Generate(context.Background(), cfg)
	require.NoError(t, err)

	require.NotNil(t, gen.selections)
	assert.True(t, gen.selections[file1Path], "file1 must stay selected")
	assert.NotContains(t, gen.selections, file2Path, "saved deselection must exclude file2")
}

// With no store wired, nil cfg.Selections must resolve to the full select-all set,
// identical to scanner.NewSelectAll.
func TestDefaultContextService_Generate_NoStoreSelectsAll(t *testing.T) {
	tmpDir := t.TempDir()
	tree := twoFileTree(tmpDir)

	gen := &capturingGenerator{content: "content"}
	svc := NewContextService(
		WithScanner(&mockScanner{tree: tree}),
		WithGenerator(gen),
	)

	cfg := GenerateConfig{
		RootPath:   tmpDir,
		OutputPath: filepath.Join(tmpDir, "output.md"),
	}
	_, err := svc.Generate(context.Background(), cfg)
	require.NoError(t, err)

	require.NotNil(t, gen.selections)
	assert.Equal(t, scanner.NewSelectAll(tree), gen.selections)
	assert.True(t, gen.selections[filepath.Join(tmpDir, "file1.go")])
	assert.True(t, gen.selections[filepath.Join(tmpDir, "file2.go")])
}

// A malformed store on disk makes Load fail; the service must swallow that error
// and fall back to the full select-all set rather than propagating or emptying it.
func TestDefaultContextService_Generate_MalformedStoreFallsBack(t *testing.T) {
	tmpDir := t.TempDir()
	tree := twoFileTree(tmpDir)

	storePath := filepath.Join(t.TempDir(), "selections.json")
	require.NoError(t, os.WriteFile(storePath, []byte("{not valid json"), 0o600))

	gen := &capturingGenerator{content: "content"}
	svc := NewContextService(
		WithScanner(&mockScanner{tree: tree}),
		WithGenerator(gen),
		WithSelectionStore(selection.NewStore(storePath)),
	)

	cfg := GenerateConfig{
		RootPath:   tmpDir,
		OutputPath: filepath.Join(tmpDir, "output.md"),
	}
	_, err := svc.Generate(context.Background(), cfg)
	require.NoError(t, err)

	require.NotNil(t, gen.selections)
	assert.Equal(t, scanner.NewSelectAll(tree), gen.selections, "malformed store must fall back to full select-all")
	assert.True(t, gen.selections[filepath.Join(tmpDir, "file1.go")])
	assert.True(t, gen.selections[filepath.Join(tmpDir, "file2.go")])
}
