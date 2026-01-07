package geminiapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

func TestClient_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "generateContent"))
		assert.Equal(t, "test-key", r.URL.Query().Get("key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		resp := GenerateResponse{
			Candidates: []Candidate{
				{
					Content: Content{
						Parts: []Part{{Text: "Hello from Gemini!"}},
					},
					FinishReason: "STOP",
				},
			},
			UsageMetadata: &UsageMetadata{
				PromptTokenCount:     10,
				CandidatesTokenCount: 5,
				TotalTokenCount:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(llm.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gemini-2.5-flash",
		Timeout: 30,
	})
	require.NoError(t, err)

	result, err := client.Send(context.Background(), "test prompt")
	require.NoError(t, err)

	assert.Equal(t, "Hello from Gemini!", result.Response)
	assert.Equal(t, "Gemini", result.Provider)
	assert.Equal(t, "gemini-2.5-flash", result.Model)
	assert.NotNil(t, result.Usage)
	assert.Equal(t, 15, result.Usage.TotalTokens)
	assert.Equal(t, 10, result.Usage.PromptTokens)
	assert.Equal(t, 5, result.Usage.CompletionTokens)
}

func TestClient_Send_MultipleParts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GenerateResponse{
			Candidates: []Candidate{
				{
					Content: Content{
						Parts: []Part{
							{Text: "First part. "},
							{Text: "Second part."},
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(llm.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gemini-2.5-flash",
		Timeout: 30,
	})

	result, err := client.Send(context.Background(), "test")
	require.NoError(t, err)
	assert.Equal(t, "First part. Second part.", result.Response)
}

func TestClient_Send_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GenerateResponse{
			Error: &APIError{
				Code:    400,
				Message: "API key not valid",
				Status:  "INVALID_ARGUMENT",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(llm.Config{
		APIKey:  "bad-key",
		BaseURL: server.URL,
		Model:   "gemini-2.5-flash",
		Timeout: 30,
	})

	_, err := client.Send(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key not valid")
	assert.Contains(t, err.Error(), "400")
}

func TestClient_Send_NoCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GenerateResponse{
			Candidates: []Candidate{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(llm.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gemini-2.5-flash",
		Timeout: 30,
	})

	_, err := client.Send(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no candidates")
}

func TestClient_NewClient_Validation(t *testing.T) {
	_, err := NewClient(llm.Config{
		Model:   "gemini-2.5-flash",
		Timeout: 30,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api key is required")

	client, err := NewClient(llm.Config{
		APIKey:  "test-key",
		Timeout: 30,
	})
	require.NoError(t, err)
	assert.Equal(t, "gemini-2.5-flash", client.model)
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
		resp := GenerateResponse{
			Candidates: []Candidate{
				{Content: Content{Parts: []Part{{Text: "Hello!"}}}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient(llm.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gemini-2.5-flash",
		Timeout: 30,
	})

	var stages []string
	result, err := client.SendWithProgress(context.Background(), "test", func(stage string) {
		stages = append(stages, stage)
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello!", result.Response)
	assert.Contains(t, stages, "Connecting to Gemini API...")
	assert.Contains(t, stages, "Response received")
}
