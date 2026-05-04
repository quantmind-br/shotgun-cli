package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements Provider interface for testing.
type mockProvider struct {
	name string
}

func (m *mockProvider) Send(ctx context.Context, content string) (*Result, error) {
	return &Result{Response: "mock response", Provider: m.name}, nil
}

func (m *mockProvider) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*Result, error) {
	return m.Send(ctx, content)
}

func (m *mockProvider) Name() string         { return m.name }
func (m *mockProvider) IsAvailable() bool    { return true }
func (m *mockProvider) IsConfigured() bool   { return true }
func (m *mockProvider) ValidateConfig() error { return nil }

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	// Test empty registry
	assert.Empty(t, r.SupportedProviders())
	assert.False(t, r.IsRegistered(ProviderOpenAI))

	// Register a provider
	r.Register(ProviderOpenAI, func(cfg Config) (Provider, error) {
		return &mockProvider{name: "OpenAI"}, nil
	})

	assert.True(t, r.IsRegistered(ProviderOpenAI))
	assert.Len(t, r.SupportedProviders(), 1)

	// Create provider
	cfg := Config{Provider: ProviderOpenAI}
	provider, err := r.Create(cfg)
	require.NoError(t, err)
	assert.Equal(t, "OpenAI", provider.Name())

	// Try to create unregistered provider
	cfg.Provider = ProviderAnthropic
	_, err = r.Create(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider")
}

func TestRegistryMultipleProviders(t *testing.T) {
	r := NewRegistry()

	r.Register(ProviderOpenAI, func(cfg Config) (Provider, error) {
		return &mockProvider{name: "OpenAI"}, nil
	})
	r.Register(ProviderAnthropic, func(cfg Config) (Provider, error) {
		return &mockProvider{name: "Anthropic"}, nil
	})

	assert.Len(t, r.SupportedProviders(), 2)
	assert.True(t, r.IsRegistered(ProviderOpenAI))
	assert.True(t, r.IsRegistered(ProviderAnthropic))
	assert.False(t, r.IsRegistered(ProviderGemini))
}
