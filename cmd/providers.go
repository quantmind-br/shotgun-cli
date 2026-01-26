package cmd

import (
	"fmt"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/anthropic"
	"github.com/quantmind-br/shotgun-cli/internal/platform/geminiapi"
	"github.com/quantmind-br/shotgun-cli/internal/platform/openai"
)

// providerRegistry is the global provider registry.
var providerRegistry *llm.Registry

func init() {
	providerRegistry = llm.NewRegistry()

	// Register OpenAI
	providerRegistry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
		return openai.NewClient(cfg)
	})

	// Register Anthropic
	providerRegistry.Register(llm.ProviderAnthropic, func(cfg llm.Config) (llm.Provider, error) {
		return anthropic.NewClient(cfg)
	})

	// Register Gemini API
	providerRegistry.Register(llm.ProviderGemini, func(cfg llm.Config) (llm.Provider, error) {
		return geminiapi.NewClient(cfg)
	})
}

// CreateLLMProvider creates a provider based on configuration.
func CreateLLMProvider(cfg llm.Config) (llm.Provider, error) {
	provider, err := providerRegistry.Create(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

// GetProviderRegistry returns the registry for external use.
func GetProviderRegistry() *llm.Registry {
	return providerRegistry
}
