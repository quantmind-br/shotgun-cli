package cmd

import (
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

// BuildLLMConfig builds the LLM configuration from Viper.
func BuildLLMConfig() llm.Config {
	provider := llm.ProviderType(viper.GetString("llm.provider"))

	// Get defaults for the provider.
	defaults := llm.DefaultConfigs()[provider]

	cfg := llm.Config{
		Provider: provider,
		APIKey:   viper.GetString("llm.api-key"),
		BaseURL:  viper.GetString("llm.base-url"),
		Model:    viper.GetString("llm.model"),
		Timeout:  viper.GetInt("llm.timeout"),
	}

	// Apply defaults if not configured.
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaults.BaseURL
	}
	if cfg.Model == "" {
		cfg.Model = defaults.Model
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaults.Timeout
	}

	// GeminiWeb-specific configurations.
	if provider == llm.ProviderGeminiWeb {
		cfg.BinaryPath = viper.GetString("gemini.binary-path")
		cfg.BrowserRefresh = viper.GetString("gemini.browser-refresh")
		// Use gemini model if llm.model is not set.
		if viper.GetString("llm.model") == "" {
			cfg.Model = viper.GetString("gemini.model")
		}
	}

	return cfg
}

// BuildLLMConfigWithOverrides builds config with flag overrides.
func BuildLLMConfigWithOverrides(model string, timeout int) llm.Config {
	cfg := BuildLLMConfig()

	if model != "" {
		cfg.Model = model
	}
	if timeout > 0 {
		cfg.Timeout = timeout
	}

	return cfg
}
