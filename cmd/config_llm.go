package cmd

import (
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

// BuildLLMConfig builds the LLM configuration from Viper.
func BuildLLMConfig() llm.Config {
	provider := llm.ProviderType(viper.GetString(config.KeyLLMProvider))

	// Get defaults for the provider.
	defaults := llm.DefaultConfigs()[provider]

	cfg := llm.Config{
		Provider: provider,
		APIKey:   viper.GetString(config.KeyLLMAPIKey),
		BaseURL:  viper.GetString(config.KeyLLMBaseURL),
		Model:    viper.GetString(config.KeyLLMModel),
		Timeout:  viper.GetInt(config.KeyLLMTimeout),
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
		cfg.BinaryPath = viper.GetString(config.KeyGeminiBinaryPath)
		cfg.BrowserRefresh = viper.GetString(config.KeyGeminiBrowserRefresh)
		// Use gemini model if llm.model is not set.
		if viper.GetString(config.KeyLLMModel) == "" {
			cfg.Model = viper.GetString(config.KeyGeminiModel)
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
