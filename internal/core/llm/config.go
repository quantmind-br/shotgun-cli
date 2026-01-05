package llm

import (
	"fmt"
	"net/url"
)

// Config contains unified configuration for any LLM provider.
type Config struct {
	// Provider specifies which provider to use.
	Provider ProviderType

	// Common configurations for all providers.
	APIKey  string // API key
	BaseURL string // Base URL for API (allows custom endpoints)
	Model   string // Model to use
	Timeout int    // Timeout in seconds

	// GeminiWeb-specific configurations (legacy).
	BinaryPath     string
	BrowserRefresh string

	// Optional configurations.
	MaxTokens   int     // Max tokens in response
	Temperature float64 // Temperature (0.0 - 2.0)
}

// DefaultConfigs returns default configurations per provider.
func DefaultConfigs() map[ProviderType]Config {
	return map[ProviderType]Config{
		ProviderOpenAI: {
			Provider: ProviderOpenAI,
			BaseURL:  "https://api.openai.com/v1",
			Model:    "gpt-4o",
			Timeout:  300,
		},
		ProviderAnthropic: {
			Provider: ProviderAnthropic,
			BaseURL:  "https://api.anthropic.com",
			Model:    "claude-sonnet-4-20250514",
			Timeout:  300,
		},
		ProviderGemini: {
			Provider: ProviderGemini,
			BaseURL:  "https://generativelanguage.googleapis.com/v1beta",
			Model:    "gemini-2.5-flash",
			Timeout:  300,
		},
		ProviderGeminiWeb: {
			Provider: ProviderGeminiWeb,
			Model:    "gemini-2.5-flash",
			Timeout:  300,
		},
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if !IsValidProvider(string(c.Provider)) {
		return fmt.Errorf("invalid provider: %s", c.Provider)
	}

	// GeminiWeb doesn't require API key.
	if c.Provider != ProviderGeminiWeb {
		if c.APIKey == "" {
			return fmt.Errorf("api-key is required for provider %s", c.Provider)
		}
	}

	if c.Model == "" {
		return fmt.Errorf("model is required")
	}

	if c.BaseURL != "" && c.Provider != ProviderGeminiWeb {
		if _, err := url.Parse(c.BaseURL); err != nil {
			return fmt.Errorf("invalid base-url: %w", err)
		}
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// MaskAPIKey returns the masked API key for display.
func (c *Config) MaskAPIKey() string {
	if c.APIKey == "" {
		return "(not configured)"
	}
	if len(c.APIKey) <= 8 {
		return "***"
	}
	return c.APIKey[:4] + "..." + c.APIKey[len(c.APIKey)-4:]
}

// WithDefaults applies default values from the provider defaults.
func (c *Config) WithDefaults() *Config {
	defaults := DefaultConfigs()[c.Provider]

	if c.BaseURL == "" {
		c.BaseURL = defaults.BaseURL
	}
	if c.Model == "" {
		c.Model = defaults.Model
	}
	if c.Timeout == 0 {
		c.Timeout = defaults.Timeout
	}

	return c
}
