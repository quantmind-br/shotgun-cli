package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantmind-br/shotgun-cli/internal/app"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
)

// TestBuildGeneratorConfig_ParityWithHeadless is the load-bearing assertion of
// CI-012. The TUI and the headless path used to build contextgen.GenerateConfig
// independently and had drifted: the TUI forwarded none of the five limits, so
// the generator substituted its own ceilings and `scanner.max-files: 10000`
// still aborted at 1000.
//
// Both sides now go through app.BuildGeneratorConfig, and equivalent input must
// produce an identical config.
func TestBuildGeneratorConfig_ParityWithHeadless(t *testing.T) {
	t.Parallel()

	scanCfg := &scanner.ScanConfig{
		MaxFileSize:    2 * 1024 * 1024,
		MaxFiles:       9000,
		SkipBinary:     true,
		IncludeIgnored: true,
	}
	const maxTotal = int64(50 * 1024 * 1024)

	uiCfg := &GenerateConfig{
		Template:       &template.Template{Content: "tpl"},
		TaskDesc:       "task",
		Rules:          "rules",
		MaxTotalSize:   maxTotal,
		MaxFileSize:    scanCfg.MaxFileSize,
		MaxFiles:       int(scanCfg.MaxFiles),
		SkipBinary:     scanCfg.SkipBinary,
		IncludeTree:    true,
		IncludeSummary: true,
		IncludeIgnored: scanCfg.IncludeIgnored,
	}
	fromTUI := buildGeneratorConfig(uiCfg)

	headlessInput := app.ScannerLimits(scanCfg, app.GeneratorConfigInput{
		Template:       "tpl",
		MaxTotalSize:   maxTotal,
		IncludeTree:    true,
		IncludeSummary: true,
	})
	fromHeadless := app.BuildGeneratorConfig(headlessInput)

	// TemplateVars are legitimately front-end specific; every other field must match.
	fromTUI.TemplateVars = nil
	fromHeadless.TemplateVars = nil

	assert.Equal(t, fromHeadless, fromTUI)
}

// TestBuildGeneratorConfig_ForwardsLimits pins the concrete regression: none of
// these five may arrive as a zero value when the request carries them.
func TestBuildGeneratorConfig_ForwardsLimits(t *testing.T) {
	t.Parallel()

	got := buildGeneratorConfig(&GenerateConfig{
		Template:       &template.Template{Content: "tpl"},
		MaxTotalSize:   50 * 1024 * 1024,
		MaxFileSize:    2 * 1024 * 1024,
		MaxFiles:       7500,
		SkipBinary:     true,
		IncludeIgnored: true,
	})

	assert.Equal(t, int64(50*1024*1024), got.MaxTotalSize)
	assert.Equal(t, int64(2*1024*1024), got.MaxFileSize)
	assert.Equal(t, 7500, got.MaxFiles)
	assert.True(t, got.SkipBinary)
	assert.True(t, got.IncludeIgnored, "the ignored-file toggle drives tree rendering")
}

// TestHandleStartGeneration_PlumbsLimits checks the other half: the wizard must
// actually populate the request. A perfect builder fed an empty request would
// reproduce the bug exactly.
func TestHandleStartGeneration_PlumbsLimits(t *testing.T) {
	t.Parallel()

	scanCfg := &scanner.ScanConfig{
		MaxFileSize:    3 * 1024 * 1024,
		MaxFiles:       8000,
		SkipBinary:     true,
		IncludeIgnored: true,
	}
	wizardCfg := &WizardConfig{
		Context: ContextConfig{IncludeTree: true, IncludeSummary: true, MaxSize: "50MB"},
	}

	m := NewWizard("/tmp/parity", scanCfg, wizardCfg, nil)
	m.generateCoordinator = NewGenerateCoordinator(&mockGenerator{})

	cmd := m.handleStartGeneration(startGenerationMsg{
		fileTree:      &scanner.FileNode{Name: "root", IsDir: true},
		selectedFiles: map[string]bool{},
		template:      &template.Template{Content: "tpl"},
		rootPath:      "/tmp/parity",
	})
	require.NotNil(t, cmd)

	got := m.generateCoordinator.config
	require.NotNil(t, got, "Start should have recorded the request")

	assert.Equal(t, int64(50*1024*1024), got.MaxTotalSize, "context.max-size must reach generation")
	assert.Equal(t, scanCfg.MaxFileSize, got.MaxFileSize)
	assert.Equal(t, int(scanCfg.MaxFiles), got.MaxFiles)
	assert.True(t, got.SkipBinary)
	assert.True(t, got.IncludeIgnored)
}

func TestWizardContextMaxSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		maxSize string
		want    int64
	}{
		{"parsed", "50MB", 50 * 1024 * 1024},
		{"decimal", "10.5MB", 11010048},
		{"empty means no limit", "", 0},
		{"unparseable means no limit", "nonsense", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := NewWizard("/tmp/x", nil, &WizardConfig{Context: ContextConfig{MaxSize: tt.maxSize}}, nil)
			assert.Equal(t, tt.want, m.contextMaxSize())
		})
	}
}
