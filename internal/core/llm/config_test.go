package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid openai config",
			cfg: Config{
				Provider: ProviderOpenAI,
				APIKey:   "sk-test",
				Model:    "gpt-4o",
				Timeout:  300,
			},
			wantErr: false,
		},
		{
			name: "valid anthropic config",
			cfg: Config{
				Provider: ProviderAnthropic,
				APIKey:   "sk-ant-test",
				Model:    "claude-sonnet-4-20250514",
				Timeout:  300,
			},
			wantErr: false,
		},
		{
			name: "valid gemini config",
			cfg: Config{
				Provider: ProviderGemini,
				APIKey:   "AIza-test",
				Model:    "gemini-2.5-flash",
				Timeout:  300,
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			cfg: Config{
				APIKey:  "key",
				Model:   "model",
				Timeout: 300,
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "invalid provider",
			cfg: Config{
				Provider: "invalid",
				APIKey:   "key",
				Model:    "model",
				Timeout:  300,
			},
			wantErr: true,
			errMsg:  "invalid provider",
		},
		{
			name: "missing api key for openai",
			cfg: Config{
				Provider: ProviderOpenAI,
				Model:    "gpt-4o",
				Timeout:  300,
			},
			wantErr: true,
			errMsg:  "api-key is required",
		},
		{
			name: "missing model",
			cfg: Config{
				Provider: ProviderOpenAI,
				APIKey:   "key",
				Timeout:  300,
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "invalid timeout",
			cfg: Config{
				Provider: ProviderOpenAI,
				APIKey:   "key",
				Model:    "model",
				Timeout:  0,
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "negative timeout",
			cfg: Config{
				Provider: ProviderOpenAI,
				APIKey:   "key",
				Model:    "model",
				Timeout:  -1,
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigMaskAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{"empty", "", "(not configured)"},
		{"short", "abc", "***"},
		{"8 chars", "12345678", "***"},
		{"normal", "sk-1234567890abcdef", "sk-1...cdef"},
		{"long", "sk-ant-api3_very_long_api_key_here", "sk-a...here"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{APIKey: tt.apiKey}
			assert.Equal(t, tt.want, cfg.MaskAPIKey())
		})
	}
}

func TestDefaultConfigs(t *testing.T) {
	defaults := DefaultConfigs()

	assert.Len(t, defaults, 3)

	// OpenAI
	openai := defaults[ProviderOpenAI]
	assert.Equal(t, ProviderOpenAI, openai.Provider)
	assert.Equal(t, "https://api.openai.com/v1", openai.BaseURL)
	assert.Equal(t, "gpt-4o", openai.Model)
	assert.Equal(t, 300, openai.Timeout)

	// Anthropic
	anthropic := defaults[ProviderAnthropic]
	assert.Equal(t, ProviderAnthropic, anthropic.Provider)
	assert.Equal(t, "https://api.anthropic.com", anthropic.BaseURL)
	assert.Equal(t, "claude-sonnet-4-20250514", anthropic.Model)

	// Gemini
	gemini := defaults[ProviderGemini]
	assert.Equal(t, ProviderGemini, gemini.Provider)
	assert.Contains(t, gemini.BaseURL, "generativelanguage.googleapis.com")
	assert.Equal(t, "gemini-2.5-flash", gemini.Model)
}

func TestConfigWithDefaults(t *testing.T) {
	cfg := Config{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
	}

	cfg.WithDefaults()

	assert.Equal(t, "https://api.openai.com/v1", cfg.BaseURL)
	assert.Equal(t, "gpt-4o", cfg.Model)
	assert.Equal(t, 300, cfg.Timeout)
}
