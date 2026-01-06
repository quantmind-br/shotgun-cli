package app

import (
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/anthropic"
	"github.com/quantmind-br/shotgun-cli/internal/platform/gemini"
	"github.com/quantmind-br/shotgun-cli/internal/platform/geminiapi"
	"github.com/quantmind-br/shotgun-cli/internal/platform/openai"
)

var DefaultProviderRegistry = llm.NewRegistry()

func init() {
	DefaultProviderRegistry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
		return openai.NewClient(cfg)
	})
	DefaultProviderRegistry.Register(llm.ProviderAnthropic, func(cfg llm.Config) (llm.Provider, error) {
		return anthropic.NewClient(cfg)
	})
	DefaultProviderRegistry.Register(llm.ProviderGemini, func(cfg llm.Config) (llm.Provider, error) {
		return geminiapi.NewClient(cfg)
	})
	DefaultProviderRegistry.Register(llm.ProviderGeminiWeb, func(cfg llm.Config) (llm.Provider, error) {
		return gemini.NewWebProvider(cfg)
	})
}
