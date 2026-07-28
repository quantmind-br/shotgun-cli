package app

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

func TestBuildGeneratorConfig_ForwardsEveryLimit(t *testing.T) {
	t.Parallel()

	in := GeneratorConfigInput{
		Template:       "tpl",
		TemplateVars:   map[string]string{"TASK": "t"},
		MaxTotalSize:   50 * 1024 * 1024,
		MaxFileSize:    2 * 1024 * 1024,
		MaxFiles:       7500,
		SkipBinary:     true,
		IncludeTree:    true,
		IncludeSummary: true,
		IncludeIgnored: true,
	}

	got := BuildGeneratorConfig(in)

	assert.Equal(t, in.MaxTotalSize, got.MaxTotalSize)
	assert.Equal(t, in.MaxFileSize, got.MaxFileSize)
	assert.Equal(t, in.MaxFiles, got.MaxFiles)
	assert.True(t, got.SkipBinary)
	assert.True(t, got.IncludeTree)
	assert.True(t, got.IncludeSummary)
	assert.True(t, got.IncludeIgnored)
	assert.Equal(t, "tpl", got.Template)
	assert.Equal(t, map[string]string{"TASK": "t"}, got.TemplateVars)
}

func TestBuildGeneratorConfig_NilTemplateVars(t *testing.T) {
	t.Parallel()

	got := BuildGeneratorConfig(GeneratorConfigInput{})

	require.NotNil(t, got.TemplateVars, "a nil map would panic on the first variable write")
	assert.Empty(t, got.TemplateVars)
}

func TestScannerLimits_CopiesSharedFields(t *testing.T) {
	t.Parallel()

	cfg := &scanner.ScanConfig{
		MaxFileSize:    4096,
		MaxFiles:       9000,
		SkipBinary:     true,
		IncludeIgnored: true,
	}

	got := ScannerLimits(cfg, GeneratorConfigInput{Template: "keep me"})

	assert.Equal(t, int64(4096), got.MaxFileSize)
	assert.Equal(t, 9000, got.MaxFiles)
	assert.True(t, got.SkipBinary)
	assert.True(t, got.IncludeIgnored)
	assert.Equal(t, "keep me", got.Template, "unrelated fields must survive")
}

func TestScannerLimits_NilConfigIsIdentity(t *testing.T) {
	t.Parallel()

	in := GeneratorConfigInput{MaxFiles: 42, Template: "tpl"}

	assert.Equal(t, in, ScannerLimits(nil, in))
}

// TestScannerLimits_ClampsMaxFiles guards the int64 -> int narrowing. A wrapped
// negative would read as "no limit" once CI-013 lands, which is the opposite of
// what a caller asking for a huge ceiling wants.
func TestScannerLimits_ClampsMaxFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   int64
		want int
	}{
		{"zero stays zero", 0, 0},
		{"negative clamps to zero", -5, 0},
		{"ordinary value passes through", 10000, 10000},
		// math.MaxInt64 fits an int on 64-bit and overflows it on 32-bit; both
		// must land on maxInt, never on a wrapped negative.
		{"largest value clamps to max int", math.MaxInt64, int(^uint(0) >> 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ScannerLimits(&scanner.ScanConfig{MaxFiles: tt.in}, GeneratorConfigInput{})
			assert.Equal(t, tt.want, got.MaxFiles)
		})
	}
}
