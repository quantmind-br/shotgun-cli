package llm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/http"
	"github.com/quantmind-br/shotgun-cli/internal/platform/http"
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
	cfg := ClientConfig{ProviderName: "test-provider"}
	client := NewBaseClient(jsonClient, cfg)

	assert.Equal(t, "test-provider", client.Name())
}

func TestBaseClient_IsAvailable(t *testing.T) {
	t.Parallel()

	t.Run("client available", func(t *testing.T) {
		jsonClient := http.NewTestClient()
		client := NewBaseClient(jsonClient, ClientConfig{ProviderName: "test"})
		assert.True(t, client.IsAvailable())
	})

	t.Run("client not available", func(t *testing.T) {
		client := NewBaseClient(nil, ClientConfig{ProviderName: "test"})
		assert.False(t, client.IsAvailable())
	})
}

func TestBaseClient_IsConfigured(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      ClientConfig
		expected bool
	}{
		{
			name:     "configured with API key and model",
			cfg:     ClientConfig{APIKey: "test-key", Model: "gpt-4"},
			expected: true,
		},
		{
			name:     "missing API key",
			cfg:     ClientConfig{APIKey: "", Model: "gpt-4"},
			expected: false,
		},
		{
			name:     "missing model",
			cfg:     ClientConfig{APIKey: "test-key", Model: ""},
			expected: false,
		},
		{
			name:     "missing both",
			cfg:     ClientConfig{APIKey: "", Model: ""},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonClient := http.NewTestClient()
			client := NewBaseClient(jsonClient, Config{APIKey: tt.apiKey, Model: tt.model, ProviderName: "test"})
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

func TestBaseClient_Send(t *testing.T) {
	t.Parallel()

	mockSender := &mockSender{
		response: &llm.Result{
			Response:    "test response",
			Model:       "gpt-4",
			Usage: &llm.Usage{
				PromptTokens:     100,
				CompletionTokens: 200,
				TotalTokens:      300,
			},
			Duration: 500 * time.Millisecond,
		}

	jsonClient := http.NewTestClient()
	cfg := ClientConfig{
		APIKey:       "test-key",
		Model:        "gpt-4",
		ProviderName: "test-provider",
	}
	client := NewBaseClient(jsonClient, cfg)
	client.sender = mockSender

	ctx := context.Background()
	result, err := client.Send(ctx, "test content")

	require.NoError(t, err)
	require.Equal(t, "test response", result.Response)
	require.Equal(t, "gpt-4", result.Model)
	require.NotNil(t, result.Usage)
	require.Equal(t, 500*time.Millisecond, result.Duration)
	require.True(t, mockSender.buildCalled)
	require.Equal(t, "test content", mockSender.receivedContent)
}

	jsonClient := http.NewTestClient()
	cfg := Config{
		APIKey:       "test-key",
		Model:        "gpt-4",
		ProviderName: "test-provider",
	}
	client := NewBaseClient(jsonClient, cfg)
	client.sender = mockSender

	ctx := context.Background()
	result, err := client.Send(ctx, "test content")

	require.NoError(t, err)
	require.Equal(t, "test response", result.Response)
	require.Equal(t, "gpt-4", result.Model)
	require.NotNil(t, result.Usage)
	require.Equal(t, 500*time.Millisecond, result.Duration)
	require.True(t, mockSender.buildCalled)
	require.Equal(t, "test content", mockSender.receivedContent)
}

func TestBaseClient_SendWithProgress(t *testing.T) {
	t.Parallel()

	progressStages := []string{}
	mockSender := &mockSender{
		response: &llm.Result{
			Response: "test response",
			Model:    "gpt-4",
			Usage: &llm.Usage{
				TotalTokens: 300,
			},
			Duration: 500 * time.Millisecond,
		},
	}

	jsonClient := http.NewTestClient()
	cfg := Config{
		APIKey:       "test-key",
		Model:        "gpt-4",
		ProviderName: "test-provider",
	}
	client := NewBaseClient(jsonClient, cfg)
	client.sender = mockSender

	ctx := context.Background()
	result, err := client.SendWithProgress(ctx, "test content", func(stage, msg string, cur, total int64) {
		progressStages = append(progressStages, stage)
	})

	require.NoError(t, err)
	require.Equal(t, "test response", result.Response)
	require.Equal(t, 3, len(progressStages))
	require.Contains(t, progressStages, "uploading")
	require.Contains(t, progressStages, "processing")
	require.Contains(t, progressStages, "complete")
}

type mockSender struct {
	buildCalled     bool
	receivedContent string
	response        *llm.Result
}

func (m *mockSender) BuildRequest(content string) (interface{}, error) {
	m.buildCalled = true
	m.receivedContent = content
	return map[string]interface{}{"content": content}, nil
}

func (m *mockSender) GetEndpoint() string {
	return "/v1/chat/completions"
}

func (m *mockSender) GetHeaders() map[string]string {
	return map[string]string{"Authorization": "Bearer test-key", "Content-Type": "application/json"}
}

func (m *mockSender) ParseResponse(data interface{}) (*llm.Result, error) {
	return m.response, nil
}

func (m *mockSender) GetResponseType() interface{} {
	return &llm.Result{}
}
