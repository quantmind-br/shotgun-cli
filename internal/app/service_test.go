package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/core/contextgen"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
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
