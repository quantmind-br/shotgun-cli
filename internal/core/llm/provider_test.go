package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidProvider(t *testing.T) {
	tests := []struct {
		provider string
		want     bool
	}{
		{"openai", true},
		{"anthropic", true},
		{"gemini", true},
		{"invalid", false},
		{"", false},
		{"OPENAI", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := IsValidProvider(tt.provider)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllProviders(t *testing.T) {
	providers := AllProviders()
	assert.Len(t, providers, 3)
	assert.Contains(t, providers, ProviderOpenAI)
	assert.Contains(t, providers, ProviderAnthropic)
	assert.Contains(t, providers, ProviderGemini)
}

func TestProviderTypeString(t *testing.T) {
	assert.Equal(t, "openai", ProviderOpenAI.String())
	assert.Equal(t, "anthropic", ProviderAnthropic.String())
	assert.Equal(t, "gemini", ProviderGemini.String())
}
