package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLIConfig_AllFields(t *testing.T) {
	t.Parallel()

	cfg := CLIConfig{
		RootPath:     "/tmp/test",
		Include:       []string{"*.go"},
		Exclude:       []string{"vendor/*"},
		Output:        "output.md",
		MaxSize:       1024 * 1024,
		EnforceLimit:  true,
		SendGemini:    true,
		GeminiModel:   "gemini-2.5-flash",
		GeminiOutput:  "response.md",
		GeminiTimeout: 300,
		Template:       "makePlan",
		Task:           "Refactor code",
		Rules:          "Keep it clean",
		CustomVars:     map[string]string{"FOO": "bar"},
		Workers:        4,
		IncludeHidden:  true,
		IncludeIgnored: false,
		ProgressMode:   ProgressHuman,
	}

	require.Equal(t, "/tmp/test", cfg.RootPath)
	require.Equal(t, []string{"*.go"}, cfg.Include)
	require.Equal(t, []string{"vendor/*"}, cfg.Exclude)
	require.Equal(t, "output.md", cfg.Output)
	require.Equal(t, int64(1024*1024), cfg.MaxSize)
	require.True(t, cfg.EnforceLimit)
	require.True(t, cfg.SendGemini)
	require.Equal(t, "gemini-2.5-flash", cfg.GeminiModel)
	require.Equal(t, "response.md", cfg.GeminiOutput)
	require.Equal(t, 300, cfg.GeminiTimeout)
	require.Equal(t, "makePlan", cfg.Template)
	require.Equal(t, "Refactor code", cfg.Task)
	require.Equal(t, "Keep it clean", cfg.Rules)
	require.Equal(t, map[string]string{"FOO": "bar"}, cfg.CustomVars)
	require.Equal(t, 4, cfg.Workers)
	require.True(t, cfg.IncludeHidden)
	require.False(t, cfg.IncludeIgnored)
	require.Equal(t, ProgressHuman, cfg.ProgressMode)
}

func TestProgressMode_Constants(t *testing.T) {
	t.Parallel()

	require.Equal(t, ProgressMode("none"), ProgressNone)
	require.Equal(t, ProgressMode("human"), ProgressHuman)
	require.Equal(t, ProgressMode("json"), ProgressJSON)
}
