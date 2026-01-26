// Package llm provides a unified interface for interacting with various LLM providers.
package llm

import (
	"context"
	"time"
)

// Result represents the result of an LLM call.
type Result struct {
	Response    string        // Processed/cleaned response
	RawResponse string        // Raw response from API
	Model       string        // Model used
	Provider    string        // Provider name
	Duration    time.Duration // Execution time
	Usage       *Usage        // Usage metrics (tokens, etc)
}

// Usage contains API usage metrics.
type Usage struct {
	PromptTokens     int // Tokens in the prompt
	CompletionTokens int // Tokens in the response
	TotalTokens      int // Total tokens
}

// Provider defines the common interface for LLM providers.
type Provider interface {
	// Send sends a prompt and returns the response.
	Send(ctx context.Context, content string) (*Result, error)

	// SendWithProgress sends with progress callback (for TUI).
	SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*Result, error)

	// Name returns the provider name (e.g., "OpenAI", "Anthropic", "Gemini").
	Name() string

	// IsAvailable checks if the provider is available (e.g., binary exists).
	IsAvailable() bool

	// IsConfigured checks if the provider is configured (e.g., API key present).
	IsConfigured() bool

	// ValidateConfig validates the configuration before use.
	ValidateConfig() error
}

// ProviderType identifies the provider type.
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderGemini    ProviderType = "gemini"
)

// AllProviders returns all supported providers.
func AllProviders() []ProviderType {
	return []ProviderType{
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderGemini,
	}
}

// IsValidProvider checks if the provider is valid.
func IsValidProvider(p string) bool {
	for _, valid := range AllProviders() {
		if string(valid) == p {
			return true
		}
	}
	return false
}

// String returns the string representation of the provider type.
func (p ProviderType) String() string {
	return string(p)
}
