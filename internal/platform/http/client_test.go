package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRequest struct {
	Name string `json:"name"`
}

type testResponse struct {
	ID   string `json:"id"`
	Echo string `json:"echo"`
}

func TestNewJSONClient(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		client := NewJSONClient(ClientConfig{BaseURL: "http://example.com"})
		assert.Equal(t, 300*time.Second, client.httpClient.Timeout)
		assert.Equal(t, "http://example.com", client.baseURL)
	})

	t.Run("custom config", func(t *testing.T) {
		client := NewJSONClient(ClientConfig{
			BaseURL: "http://test.com",
			Timeout: 10 * time.Second,
		})
		assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
		assert.Equal(t, "http://test.com", client.baseURL)
	})
}

func TestPostJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockHandler    func(w nethttp.ResponseWriter, r *nethttp.Request)
		expectedResult *testResponse
		expectError    bool
		errorContains  string
		setup          func(client *JSONClient, server *httptest.Server)
	}{
		{
			name:        "success",
			requestBody: testRequest{Name: "test"},
			mockHandler: func(w nethttp.ResponseWriter, r *nethttp.Request) {
				assert.Equal(t, nethttp.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "test-header-val", r.Header.Get("X-Test-Header"))

				var req testRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.WriteHeader(nethttp.StatusBadRequest)
					return
				}

				resp := testResponse{ID: "123", Echo: req.Name}
				w.WriteHeader(nethttp.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			},
			expectedResult: &testResponse{ID: "123", Echo: "test"},
			expectError:    false,
		},
		{
			name:          "marshal error",
			requestBody:   make(chan int),
			mockHandler:   func(w nethttp.ResponseWriter, r *nethttp.Request) {},
			expectError:   true,
			errorContains: "failed to marshal request",
		},
		{
			name:        "server error",
			requestBody: testRequest{Name: "fail"},
			mockHandler: func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.WriteHeader(nethttp.StatusInternalServerError)
				_, _ = w.Write([]byte("internal error"))
			},
			expectedResult: nil,
			expectError:    true,
			errorContains:  "HTTP 500: internal error",
		},
		{
			name:        "invalid response json",
			requestBody: testRequest{Name: "badjson"},
			mockHandler: func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.WriteHeader(nethttp.StatusOK)
				_, _ = w.Write([]byte("{invalid-json"))
			},
			expectedResult: nil,
			expectError:    true,
			errorContains:  "failed to parse response",
		},
		{
			name:        "connection error",
			requestBody: testRequest{Name: "test"},
			mockHandler: func(w nethttp.ResponseWriter, r *nethttp.Request) {},
			setup: func(client *JSONClient, server *httptest.Server) {
				server.Close()
			},
			expectError:   true,
			errorContains: "request failed",
		},
		{
			name:        "invalid url",
			requestBody: testRequest{Name: "test"},
			mockHandler: func(w nethttp.ResponseWriter, r *nethttp.Request) {},
			setup: func(client *JSONClient, server *httptest.Server) {
				client.baseURL = "http://\x7f"
			},
			expectError:   true,
			errorContains: "failed to create request",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(nethttp.HandlerFunc(tt.mockHandler))

			closed := false
			defer func() {
				if !closed {
					server.Close()
				}
			}()

			client := NewJSONClient(ClientConfig{
				BaseURL: server.URL,
				Timeout: 1 * time.Second,
			})

			if tt.setup != nil {
				tt.setup(client, server)
				if tt.name == "connection error" {
					closed = true
				}
			}

			var result testResponse
			headers := map[string]string{"X-Test-Header": "test-header-val"}
			err := client.PostJSON(context.Background(), "/api/test", headers, tt.requestBody, &result)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, &result)
			}
		})
	}
}
