package anthropic

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
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, anthropicVersion, r.Header.Get("anthropic-version"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		resp := MessagesResponse{
			ID:   "msg_test",
			Type: "message",
			Role: "assistant",
			Content: []ContentBlock{
				{Type: "text", Text: "Hello from Claude!"},
			},
			Model:      "claude-sonnet-4-20250514",
			StopReason: "end_turn",
			Usage: UsageAPI{
				InputTokens:  10,
				OutputTokens: 5,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(llm.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "claude-sonnet-4-20250514",
		Timeout: 30,
	})
	require.NoError(t, err)

	result, err := client.Send(context.Background(), "test prompt")
	require.NoError(t, err)

	assert.Equal(t, "Hello from Claude!", result.Response)
	assert.Equal(t, "Anthropic", result.Provider)
	assert.Equal(t, "claude-sonnet-4-20250514", result.Model)
	assert.NotNil(t, result.Usage)
	assert.Equal(t, 15, result.Usage.TotalTokens)
	assert.Equal(t, 10, result.Usage.PromptTokens)
	assert.Equal(t, 5, result.Usage.CompletionTokens)
}

func TestClient_Send_MultipleContentBlocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MessagesResponse{
			ID:   "msg_test",
			Role: "assistant",
			Content: []ContentBlock{
				{Type: "text", Text: "First part. "},
				{Type: "text", Text: "Second part."},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(llm.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "claude-sonnet-4-20250514",
		Timeout: 30,
	})

	result, err := client.Send(context.Background(), "test")
	require.NoError(t, err)
	assert.Equal(t, "First part. Second part.", result.Response)
}

func TestClient_Send_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ErrorResponse{Type: "error"}
		resp.Error.Type = "authentication_error"
		resp.Error.Message = "Invalid API key"

		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(llm.Config{
		APIKey:  "bad-key",
		BaseURL: server.URL,
		Model:   "claude-sonnet-4-20250514",
		Timeout: 30,
	})

	_, err := client.Send(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid API key")
	assert.Contains(t, err.Error(), "401")
}

func TestClient_NewClient_Validation(t *testing.T) {
	// Missing API key
	_, err := NewClient(llm.Config{
		Model:   "claude-sonnet-4-20250514",
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
	assert.Equal(t, "claude-sonnet-4-20250514", client.model)
	assert.Equal(t, defaultBaseURL, client.baseURL)
	assert.Equal(t, defaultMaxTokens, client.maxTokens)
}

func TestClient_IsConfigured(t *testing.T) {
	client := &Client{apiKey: "key", model: "model"}
	assert.True(t, client.IsConfigured())

	client = &Client{apiKey: "", model: "model"}
	assert.False(t, client.IsConfigured())
}

func TestClient_SendWithProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MessagesResponse{
			ID:      "msg_test",
			Content: []ContentBlock{{Type: "text", Text: "Hello!"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(llm.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "claude-sonnet-4-20250514",
		Timeout: 30,
	})

	var stages []string
	result, err := client.SendWithProgress(context.Background(), "test", func(stage string) {
		stages = append(stages, stage)
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello!", result.Response)
	assert.Contains(t, stages, "Connecting to Anthropic...")
	assert.Contains(t, stages, "Response received")
}
