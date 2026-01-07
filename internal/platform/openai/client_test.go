package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

func TestClient_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		resp := ChatCompletionResponse{
			ID:    "test-id",
			Model: "gpt-4o",
			Choices: []Choice{
				{
					Message:      Message{Role: "assistant", Content: "Hello!"},
					FinishReason: "stop",
				},
			},
			Usage: UsageAPI{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(llm.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-4o",
		Timeout: 30,
	})
	require.NoError(t, err)

	result, err := client.Send(context.Background(), "test prompt")
	require.NoError(t, err)

	assert.Equal(t, "Hello!", result.Response)
	assert.Equal(t, "OpenAI", result.Provider)
	assert.Equal(t, "gpt-4o", result.Model)
	assert.NotNil(t, result.Usage)
	assert.Equal(t, 15, result.Usage.TotalTokens)
	assert.Equal(t, 10, result.Usage.PromptTokens)
	assert.Equal(t, 5, result.Usage.CompletionTokens)
}

func TestClient_Send_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ErrorResponse{}
		resp.Error.Message = "Invalid API key"
		resp.Error.Type = "invalid_request_error"

		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(llm.Config{
		APIKey:  "bad-key",
		BaseURL: server.URL,
		Model:   "gpt-4o",
		Timeout: 30,
	})
	require.NoError(t, err)

	_, err = client.Send(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid API key")
	assert.Contains(t, err.Error(), "401")
}

func TestClient_Send_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatCompletionResponse{
			ID:      "test-id",
			Model:   "gpt-4o",
			Choices: []Choice{},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(llm.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-4o",
		Timeout: 30,
	})
	require.NoError(t, err)

	_, err = client.Send(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no choices")
}

func TestClient_NewClient_Validation(t *testing.T) {
	// Missing API key
	_, err := NewClient(llm.Config{
		Model:   "gpt-4o",
		Timeout: 30,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api key is required")

	// Valid config
	client, err := NewClient(llm.Config{
		APIKey:  "test-key",
		Timeout: 30,
	})
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o", client.model) // default model
}

func TestClient_IsConfigured(t *testing.T) {
	client := &Client{apiKey: "key", model: "model"}
	assert.True(t, client.IsConfigured())

	client = &Client{apiKey: "", model: "model"}
	assert.False(t, client.IsConfigured())

	client = &Client{apiKey: "key", model: ""}
	assert.False(t, client.IsConfigured())
}

func TestClient_ValidateConfig(t *testing.T) {
	client := &Client{apiKey: "key", model: "model"}
	assert.NoError(t, client.ValidateConfig())

	client = &Client{apiKey: "", model: "model"}
	assert.Error(t, client.ValidateConfig())

	client = &Client{apiKey: "key", model: ""}
	assert.Error(t, client.ValidateConfig())
}

func TestClient_SendWithProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatCompletionResponse{
			ID:    "test-id",
			Model: "gpt-4o",
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: "Hello!"}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(llm.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-4o",
		Timeout: 30,
	})

	var stages []string
	result, err := client.SendWithProgress(context.Background(), "test", func(stage string) {
		stages = append(stages, stage)
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello!", result.Response)
	assert.Contains(t, stages, "Connecting to OpenAI...")
	assert.Contains(t, stages, "Response received")
}
