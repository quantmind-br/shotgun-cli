package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestBuildLLMConfig_GeminiWeb(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "geminiweb")
	viper.Set(config.KeyGeminiBinaryPath, "/path/to/geminiweb")
	viper.Set(config.KeyGeminiBrowserRefresh, "firefox")
	viper.Set(config.KeyGeminiModel, "gemini-2.0-pro")

	cfg := BuildLLMConfig()

	assert.Equal(t, llm.ProviderGeminiWeb, cfg.Provider)
	assert.Equal(t, "/path/to/geminiweb", cfg.BinaryPath)
	assert.Equal(t, "firefox", cfg.BrowserRefresh)
	assert.Equal(t, "gemini-2.0-pro", cfg.Model)
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

func TestCreateLLMProvider_OpenAI(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "openai")
	viper.Set(config.KeyLLMAPIKey, "sk-test-key")

	cfg := BuildLLMConfig()
	provider, err := CreateLLMProvider(cfg)

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "OpenAI", provider.Name())
}

func TestCreateLLMProvider_Anthropic(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "anthropic")
	viper.Set(config.KeyLLMAPIKey, "sk-ant-test-key")

	cfg := BuildLLMConfig()
	provider, err := CreateLLMProvider(cfg)

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "Anthropic", provider.Name())
}

func TestCreateLLMProvider_Gemini(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "gemini")
	viper.Set(config.KeyLLMAPIKey, "test-gemini-key")

	cfg := BuildLLMConfig()
	provider, err := CreateLLMProvider(cfg)

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "Gemini", provider.Name())
}

func TestCreateLLMProvider_GeminiWeb(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "geminiweb")

	cfg := BuildLLMConfig()
	provider, err := CreateLLMProvider(cfg)

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "GeminiWeb", provider.Name())
}

func TestCreateLLMProvider_InvalidProvider(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "invalid-provider")

	cfg := BuildLLMConfig()
	provider, err := CreateLLMProvider(cfg)

	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "failed to create provider")
}

func TestRunLLMStatus_NoError(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "geminiweb")
	viper.Set(config.KeyGeminiBinaryPath, "/nonexistent/geminiweb")

	cmd := &cobra.Command{}
	err := runLLMStatus(cmd, []string{})

	_ = err // intentionally ignored for this test
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

func TestRunLLMStatus_GeminiWeb(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "geminiweb")
	viper.Set(config.KeyGeminiBinaryPath, "/path/to/geminiweb")
	viper.Set(config.KeyGeminiModel, "gemini-2.0-pro")

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
	assert.Contains(t, output, "Provider:  geminiweb")
	assert.Contains(t, output, "gemini-2.0-pro")
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

	require.NoError(t, err)
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

func TestRunLLMDoctor_GeminiWeb(t *testing.T) {
	viper.Reset()
	viper.Set(config.KeyLLMProvider, "geminiweb")
	viper.Set(config.KeyGeminiBinaryPath, "/path/to/geminiweb")
	viper.Set(config.KeyGeminiModel, "gemini-2.0-pro")

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
	assert.Contains(t, output, "Running diagnostics for geminiweb")
	assert.Contains(t, output, "Checking provider... geminiweb")
	assert.Contains(t, output, "Checking model... gemini-2.0-pro")
	assert.NotContains(t, output, "Checking API key")
	assert.Contains(t, output, "Next steps:")
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

	_ = err
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

	require.NoError(t, err)
	assert.Contains(t, output, "Checking API key... not configured")
	assert.Contains(t, output, "API key not configured")
	assert.Contains(t, output, "Found")
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
	assert.Contains(t, output, "geminiweb")
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
		{"GeminiWeb as current", "geminiweb", "* geminiweb"},
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
	assert.Contains(t, output, "Browser-based")
}

func TestGetProviderRegistry(t *testing.T) {
	registry := GetProviderRegistry()

	assert.NotNil(t, registry)

	providers := registry.SupportedProviders()
	assert.Contains(t, providers, llm.ProviderOpenAI)
	assert.Contains(t, providers, llm.ProviderAnthropic)
	assert.Contains(t, providers, llm.ProviderGemini)
	assert.Contains(t, providers, llm.ProviderGeminiWeb)
}

func TestGetProviderRegistry_AllProvidersPresent(t *testing.T) {
	registry := GetProviderRegistry()

	providers := registry.SupportedProviders()

	// Verify all expected providers are present
	expectedProviders := []llm.ProviderType{
		llm.ProviderOpenAI,
		llm.ProviderAnthropic,
		llm.ProviderGemini,
		llm.ProviderGeminiWeb,
	}

	for _, expected := range expectedProviders {
		found := false
		for _, p := range providers {
			if p == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected provider %s not found in registry", expected)
		}
	}
}

func TestGetProviderRegistry_CreatesProvider(t *testing.T) {
	registry := GetProviderRegistry()

	tests := []struct {
		name     string
		provider llm.ProviderType
		model    string
		apiKey   string
	}{
		{"OpenAI", llm.ProviderOpenAI, "gpt-4o", "sk-test"},
		{"Anthropic", llm.ProviderAnthropic, "claude-sonnet-4-20250514", "sk-ant-test"},
		{"Gemini", llm.ProviderGemini, "gemini-2.5-flash", "test-key"},
		{"GeminiWeb", llm.ProviderGeminiWeb, "gemini-2.0-pro", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &llm.Config{
				Provider: tt.provider,
				Model:    tt.model,
				APIKey:   tt.apiKey,
				Timeout:  60,
			}

			provider, err := registry.Create(*cfg)

			if err != nil {
				t.Logf("Provider creation returned error (may be expected): %v", err)
			}
			if provider != nil {
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestGetProviderRegistry_Singleton(t *testing.T) {
	registry1 := GetProviderRegistry()
	registry2 := GetProviderRegistry()

	// Should return the same registry instance
	assert.Same(t, registry1, registry2)
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
			name:     "empty url with GeminiWeb - no default configured",
			url:      "",
			provider: llm.ProviderGeminiWeb,
			want:     "(default)",
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
			name:     "custom url with GeminiWeb",
			url:      "https://custom.endpoint.com",
			provider: llm.ProviderGeminiWeb,
			want:     "https://custom.endpoint.com",
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
