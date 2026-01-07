package llm

import (
	"testing"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseClient(t *testing.T) {
	t.Parallel()

	jsonClient := http.NewTestClient()
	cfg := ClientConfig{
		APIKey:       "test-key",
		Model:        "gpt-4",
		MaxTokens:    1000,
		ProviderName: "test-provider",
		BaseURL:      "https://api.test.com",
		Timeout:      30 * time.Second,
	}

	client := NewBaseClient(jsonClient, cfg)

	require.Equal(t, "test-provider", client.Name())
	require.Equal(t, "test-key", client.APIKey)
	require.Equal(t, "gpt-4", client.Model)
	require.Equal(t, 1000, client.MaxTokens)
	require.Equal(t, "https://api.test.com", client.BaseURL)
	require.Equal(t, 30*time.Second, client.Timeout)
	require.NotNil(t, client.JSONClient)
	require.NotNil(t, client.sender)
}

func TestBaseClient_Name(t *testing.T) {
	t.Parallel()

	jsonClient := http.NewTestClient()
	client := NewBaseClient(jsonClient, ClientConfig{ProviderName: "test"})

	assert.Equal(t, "test", client.Name())
}

func TestBaseClient_IsAvailable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		jsonClient *http.JSONClient
		expected   bool
	}{
		{
			name:       "client available",
			jsonClient: http.NewTestClient(),
			expected:   true,
		},
		{
			name:       "client not available",
			jsonClient: nil,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewBaseClient(tt.jsonClient, ClientConfig{ProviderName: "test"})
			result := client.IsAvailable()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseClient_IsConfigured(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      Config
		expected bool
	}{
		{
			name:     "configured with API key and model",
			cfg:      Config{APIKey: "key", Model: "model", ProviderName: "test"},
			expected: true,
		},
		{
			name:     "missing API key",
			cfg:      Config{APIKey: "", Model: "model", ProviderName: "test"},
			expected: false,
		},
		{
			name:     "missing model",
			cfg:      Config{APIKey: "key", Model: "", ProviderName: "test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonClient := http.NewTestClient()
			client := NewBaseClient(jsonClient, tt.cfg)
			result := client.IsConfigured()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseClient_ValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     Config{APIKey: "key", Model: "model", ProviderName: "test"},
			wantErr: false,
		},
		{
			name:    "missing API key",
			cfg:     Config{APIKey: "", Model: "model", ProviderName: "test"},
			wantErr: true,
		},
		{
			name:    "missing model",
			cfg:     Config{APIKey: "key", Model: "", ProviderName: "test"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonClient := http.NewTestClient()
			client := NewBaseClient(jsonClient, tt.cfg)
			err := client.ValidateConfig()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
