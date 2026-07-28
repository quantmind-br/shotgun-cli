package cmd

import (
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

// BuildLLMConfig builds the LLM configuration from Viper.
func BuildLLMConfig() llm.Config {
	cfg := llm.Config{
		Provider: llm.ProviderType(viper.GetString(config.KeyLLMProvider)),
		APIKey:   viper.GetString(config.KeyLLMAPIKey),
		BaseURL:  viper.GetString(config.KeyLLMBaseURL),
		Model:    viper.GetString(config.KeyLLMModel),
		Timeout:  viper.GetInt(config.KeyLLMTimeout),
	}
	cfg.WithDefaults()

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
