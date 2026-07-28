package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantmind-br/shotgun-cli/internal/app"
	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

func TestBuildLLMConfig_CustomValues(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "openai")
	viper.Set(config.KeyLLMAPIKey, "sk-test-key")
	viper.Set(config.KeyLLMModel, "gpt-4o")
	viper.Set(config.KeyLLMTimeout, 120)
	viper.Set(config.KeyLLMBaseURL, "https://custom.api.com")

	cfg := BuildLLMConfig()

	assert.Equal(t, llm.ProviderOpenAI, cfg.Provider)
	assert.Equal(t, "sk-test-key", cfg.APIKey)
	assert.Equal(t, "gpt-4o", cfg.Model)
	assert.Equal(t, 120, cfg.Timeout)
	assert.Equal(t, "https://custom.api.com", cfg.BaseURL)
}

func TestBuildLLMConfigWithOverrides(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "anthropic")
	viper.Set(config.KeyLLMModel, "claude-sonnet-4-20250514")
	viper.Set(config.KeyLLMTimeout, 60)

	cfg := BuildLLMConfigWithOverrides("claude-opus-4-20250514", 180)

	assert.Equal(t, llm.ProviderAnthropic, cfg.Provider)
	assert.Equal(t, "claude-opus-4-20250514", cfg.Model)
	assert.Equal(t, 180, cfg.Timeout)
}

func TestBuildLLMConfigWithOverrides_OnlyModel(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "openai")
	viper.Set(config.KeyLLMModel, "gpt-4o")
	viper.Set(config.KeyLLMTimeout, 60)

	cfg := BuildLLMConfigWithOverrides("gpt-4o-mini", 0)

	assert.Equal(t, "gpt-4o-mini", cfg.Model)
	assert.Equal(t, 60, cfg.Timeout)
}

func TestDefaultProviderRegistry_CreateOpenAI(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "openai")
	viper.Set(config.KeyLLMAPIKey, "sk-test-key")

	provider, err := app.DefaultProviderRegistry.Create(BuildLLMConfig())

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "OpenAI", provider.Name())
}

func TestDefaultProviderRegistry_CreateAnthropic(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "anthropic")
	viper.Set(config.KeyLLMAPIKey, "sk-ant-test-key")

	provider, err := app.DefaultProviderRegistry.Create(BuildLLMConfig())

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "Anthropic", provider.Name())
}

func TestDefaultProviderRegistry_CreateGemini(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "gemini")
	viper.Set(config.KeyLLMAPIKey, "test-gemini-key")

	provider, err := app.DefaultProviderRegistry.Create(BuildLLMConfig())

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "Gemini", provider.Name())
}

func TestDefaultProviderRegistry_CreateInvalidProvider(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "invalid-provider")

	provider, err := app.DefaultProviderRegistry.Create(BuildLLMConfig())

	require.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "unsupported provider")
}

func TestRunLLMStatus_OpenAI_Configured(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "openai")
	viper.Set(config.KeyLLMAPIKey, "sk-test-key-12345")
	viper.Set(config.KeyLLMModel, "gpt-4o")
	viper.Set(config.KeyLLMTimeout, 60)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMStatus(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "=== LLM Configuration ===")
	assert.Contains(t, output, "Provider:  openai")
	assert.Contains(t, output, "Model:     gpt-4o")
	assert.Contains(t, output, "Timeout:   60s")
	assert.Contains(t, output, "sk-t")
}

func TestRunLLMStatus_Anthropic_WithCustomURL(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "anthropic")
	viper.Set(config.KeyLLMAPIKey, "sk-ant-test-key")
	viper.Set(config.KeyLLMModel, "claude-sonnet-4-20250514")
	viper.Set(config.KeyLLMBaseURL, "https://custom.anthropic.proxy.com")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMStatus(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Provider:  anthropic")
	assert.Contains(t, output, "https://custom.anthropic.proxy.com")
	assert.Contains(t, output, "claude-sonnet-4-20250514")
}

func TestRunLLMStatus_Gemini_DefaultBaseURL(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "gemini")
	viper.Set(config.KeyLLMAPIKey, "test-gemini-key")
	viper.Set(config.KeyLLMModel, "gemini-2.5-flash")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMStatus(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Provider:  gemini")
	assert.Contains(t, output, "https://generativelanguage.googleapis.com/v1beta")
	assert.Contains(t, output, "gemini-2.5-flash")
}

func TestRunLLMStatus_MissingAPIKey(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "openai")
	// Don't set API key

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMStatus(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	_ = err
	assert.Contains(t, output, "Not ready")
}

func TestRunLLMStatus_InvalidProvider(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "invalid-provider")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMStatus(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	_ = err
	assert.Contains(t, output, "Not ready")
}

func TestRunLLMDoctor_OpenAI(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "openai")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMDoctor(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// No API key and no model configured, so doctor must report a failure.
	require.Error(t, err)
	assert.Contains(t, output, "Running diagnostics for openai")
	assert.Contains(t, output, "Checking provider...")
	assert.Contains(t, output, "Checking API key...")
	assert.Contains(t, output, "Checking model...")
	assert.Contains(t, output, "Found")
}

func TestRunLLMDoctor_OpenAI_Configured(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "openai")
	viper.Set(config.KeyLLMAPIKey, "sk-test-key")
	viper.Set(config.KeyLLMModel, "gpt-4o")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMDoctor(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Checking provider... openai")
	assert.Contains(t, output, "Checking API key... configured")
	assert.Contains(t, output, "Checking model... gpt-4o")
	assert.Contains(t, output, "No issues found")
}

func TestRunLLMDoctor_Anthropic(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "anthropic")
	viper.Set(config.KeyLLMAPIKey, "sk-ant-test-key")
	viper.Set(config.KeyLLMModel, "claude-sonnet-4-20250514")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMDoctor(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Running diagnostics for anthropic")
	assert.Contains(t, output, "Checking provider... anthropic")
	assert.Contains(t, output, "Checking API key... configured")
	assert.Contains(t, output, "Checking model... claude-sonnet-4-20250514")
	assert.Contains(t, output, "No issues found")
}

func TestRunLLMDoctor_Gemini(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "gemini")
	viper.Set(config.KeyLLMAPIKey, "test-gemini-key")
	viper.Set(config.KeyLLMModel, "gemini-2.5-flash")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMDoctor(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Running diagnostics for gemini")
	assert.Contains(t, output, "Checking provider... gemini")
	assert.Contains(t, output, "Checking API key... configured")
	assert.Contains(t, output, "Checking model... gemini-2.5-flash")
	assert.Contains(t, output, "No issues found")
}

func TestRunLLMDoctor_InvalidProvider(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "invalid-provider")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMDoctor(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Error(t, err)
	assert.Contains(t, output, "Running diagnostics for invalid-provider")
	assert.Contains(t, output, "Found")
	assert.Contains(t, output, "Invalid provider")
}

func TestRunLLMDoctor_NoAPIKey(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "openai")
	// Don't set API key

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMDoctor(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Error(t, err)
	assert.Contains(t, output, "Checking API key... not configured")
	assert.Contains(t, output, "API key not configured")
	assert.Contains(t, output, "Found")
}

// TestRunLLMDoctor_ErrorReportsIssueCount pins the contract CI-010 adds: the
// returned error exists only to carry a non-zero exit code, and must not repeat
// the issue list that already went to stdout.
func TestRunLLMDoctor_ErrorReportsIssueCount(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "openai")
	// Neither API key nor model: two issues.

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runLLMDoctor(&cobra.Command{}, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration issue(s)")
	assert.NotContains(t, err.Error(), "API key not configured",
		"the error must not duplicate the issue list already printed to stdout")

	// The issue list is printed exactly once.
	assert.Equal(t, 1, strings.Count(output, "API key not configured"))
}

// TestRunLLMDoctor_SilencesUsage guards against Cobra burying the remediation
// steps under a full usage dump when doctor exits non-zero.
func TestRunLLMDoctor_SilencesUsage(t *testing.T) {
	assert.True(t, llmDoctorCmd.SilenceUsage, "llm doctor must not print usage on failure")
}

func TestRunLLMList(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "anthropic")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMList(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Supported LLM Providers")
	assert.Contains(t, output, "openai")
	assert.Contains(t, output, "anthropic")
	assert.Contains(t, output, "gemini")
	assert.Contains(t, output, "GPT-4o")
	assert.Contains(t, output, "Claude 4")
	assert.Contains(t, output, "Configure with:")
}

func TestRunLLMList_CurrentProviderMarker(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		expectedMarker string
	}{
		{"OpenAI as current", "openai", "* openai"},
		{"Anthropic as current", "anthropic", "* anthropic"},
		{"Gemini as current", "gemini", "* gemini"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set(config.KeyLLMProvider, tt.provider)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd := &cobra.Command{}
			err := runLLMList(cmd, []string{})

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			require.NoError(t, err)
			assert.Contains(t, output, tt.expectedMarker)
		})
	}
}

func TestRunLLMList_ProviderDescriptions(t *testing.T) {
	viper.Reset()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err := runLLMList(cmd, []string{})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "GPT-4o, GPT-4, o1, o3")
	assert.Contains(t, output, "Claude 4, Claude 3.5")
	assert.Contains(t, output, "Gemini 2.5, Gemini 2.0")
}

func TestDefaultProviderRegistry_SupportedProviders(t *testing.T) {
	providers := app.DefaultProviderRegistry.SupportedProviders()

	assert.Contains(t, providers, llm.ProviderOpenAI)
	assert.Contains(t, providers, llm.ProviderAnthropic)
	assert.Contains(t, providers, llm.ProviderGemini)
}

func TestDisplayURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		provider llm.ProviderType
		want     string
	}{
		{
			name:     "empty url with OpenAI - shows default",
			url:      "",
			provider: llm.ProviderOpenAI,
			want:     "(default: https://api.openai.com/v1)",
		},
		{
			name:     "empty url with Anthropic - shows default",
			url:      "",
			provider: llm.ProviderAnthropic,
			want:     "(default: https://api.anthropic.com)",
		},
		{
			name:     "empty url with Gemini - shows default",
			url:      "",
			provider: llm.ProviderGemini,
			want:     "(default: https://generativelanguage.googleapis.com/v1beta)",
		},
		{
			name:     "custom url with OpenAI",
			url:      "https://custom.openai.proxy.com/v1",
			provider: llm.ProviderOpenAI,
			want:     "https://custom.openai.proxy.com/v1",
		},
		{
			name:     "custom url with Anthropic",
			url:      "https://custom.anthropic.proxy.com",
			provider: llm.ProviderAnthropic,
			want:     "https://custom.anthropic.proxy.com",
		},
		{
			name:     "custom url with Gemini",
			url:      "https://custom.gemini.proxy.com/v1beta",
			provider: llm.ProviderGemini,
			want:     "https://custom.gemini.proxy.com/v1beta",
		},
		{
			name:     "localhost url",
			url:      "http://localhost:8080/v1",
			provider: llm.ProviderOpenAI,
			want:     "http://localhost:8080/v1",
		},
		{
			name:     "openrouter url",
			url:      "https://openrouter.ai/api/v1",
			provider: llm.ProviderAnthropic,
			want:     "https://openrouter.ai/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := displayURL(tt.url, tt.provider)
			assert.Equal(t, tt.want, got, "displayURL(%q, %v) = %q, want %q", tt.url, tt.provider, got, tt.want)
		})
	}
}
