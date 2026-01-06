package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/app"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultTemplateVars() map[string]string {
	return map[string]string{
		"TASK":  "Test task",
		"RULES": "Test rules",
	}
}

func TestContextServiceIntegration_BasicGeneration(t *testing.T) {
	root := repoRoot()
	fixture := filepath.Join(root, "test", "fixtures", "sample-project")
	output := filepath.Join(t.TempDir(), "output.md")

	svc := app.NewContextService()
	cfg := app.GenerateConfig{
		RootPath:     fixture,
		OutputPath:   output,
		ScanConfig:   scanner.DefaultScanConfig(),
		TemplateVars: defaultTemplateVars(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := svc.Generate(ctx, cfg)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Content)
	assert.Equal(t, output, result.OutputPath)
	assert.Greater(t, result.FileCount, 0)
	assert.Greater(t, result.ContentSize, int64(0))
	assert.Greater(t, result.TokenEstimate, int64(0))

	savedContent, err := os.ReadFile(output)
	require.NoError(t, err)
	assert.Equal(t, result.Content, string(savedContent))
}

func TestContextServiceIntegration_WithProgress(t *testing.T) {
	root := repoRoot()
	fixture := filepath.Join(root, "test", "fixtures", "sample-project")
	output := filepath.Join(t.TempDir(), "output.md")

	svc := app.NewContextService()
	cfg := app.GenerateConfig{
		RootPath:     fixture,
		OutputPath:   output,
		ScanConfig:   scanner.DefaultScanConfig(),
		TemplateVars: defaultTemplateVars(),
	}

	var stages []string
	progress := func(stage, message string, current, total int64) {
		stages = append(stages, stage)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := svc.GenerateWithProgress(ctx, cfg, progress)
	require.NoError(t, err)
	assert.NotNil(t, result)

	assert.Contains(t, stages, "scanning")
	assert.Contains(t, stages, "generating")
	assert.Contains(t, stages, "saving")
	assert.Contains(t, stages, "complete")
}

func TestContextServiceIntegration_WithSizeLimit(t *testing.T) {
	root := repoRoot()
	fixture := filepath.Join(root, "test", "fixtures", "sample-project")
	output := filepath.Join(t.TempDir(), "output.md")

	svc := app.NewContextService()
	cfg := app.GenerateConfig{
		RootPath:     fixture,
		OutputPath:   output,
		ScanConfig:   scanner.DefaultScanConfig(),
		MaxSize:      100,
		EnforceLimit: true,
		TemplateVars: defaultTemplateVars(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := svc.Generate(ctx, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds")
}

func TestContextServiceIntegration_WithCustomSelections(t *testing.T) {
	root := repoRoot()
	fixture := filepath.Join(root, "test", "fixtures", "sample-project")
	output := filepath.Join(t.TempDir(), "output.md")

	svc := app.NewContextService()

	scanCfg := scanner.DefaultScanConfig()
	fs := scanner.NewFileSystemScanner()
	tree, err := fs.Scan(fixture, scanCfg)
	require.NoError(t, err)

	selections := make(map[string]bool)
	if len(tree.Children) > 0 {
		selections[tree.Children[0].Path] = true
	}

	cfg := app.GenerateConfig{
		RootPath:     fixture,
		OutputPath:   output,
		ScanConfig:   scanCfg,
		Selections:   selections,
		TemplateVars: defaultTemplateVars(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := svc.Generate(ctx, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Content)
}

func TestContextServiceIntegration_InvalidPath(t *testing.T) {
	svc := app.NewContextService()
	cfg := app.GenerateConfig{
		RootPath: "/nonexistent/path/that/does/not/exist",
	}

	_, err := svc.Generate(context.Background(), cfg)
	assert.Error(t, err)
}

func TestContextServiceIntegration_EmptyDirectory(t *testing.T) {
	emptyDir := t.TempDir()
	output := filepath.Join(t.TempDir(), "output.md")

	svc := app.NewContextService()
	cfg := app.GenerateConfig{
		RootPath:     emptyDir,
		OutputPath:   output,
		ScanConfig:   scanner.DefaultScanConfig(),
		TemplateVars: defaultTemplateVars(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := svc.Generate(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.FileCount)
}

func TestContextServiceIntegration_WithIncludePatterns(t *testing.T) {
	root := repoRoot()
	fixture := filepath.Join(root, "test", "fixtures", "sample-project")
	output := filepath.Join(t.TempDir(), "output.md")

	svc := app.NewContextService()
	cfg := app.GenerateConfig{
		RootPath:   fixture,
		OutputPath: output,
		ScanConfig: &scanner.ScanConfig{
			IncludePatterns: []string{"*.go"},
		},
		TemplateVars: defaultTemplateVars(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := svc.Generate(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestContextServiceIntegration_OutputPathAutoGenerated(t *testing.T) {
	root := repoRoot()
	fixture := filepath.Join(root, "test", "fixtures", "sample-project")

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	svc := app.NewContextService()
	cfg := app.GenerateConfig{
		RootPath:     fixture,
		ScanConfig:   scanner.DefaultScanConfig(),
		TemplateVars: defaultTemplateVars(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := svc.Generate(ctx, cfg)
	require.NoError(t, err)

	assert.Contains(t, result.OutputPath, "shotgun-prompt-")
	assert.FileExists(t, result.OutputPath)

	os.Remove(result.OutputPath)
}
