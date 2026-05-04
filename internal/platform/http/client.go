package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"time"
)

// JSONClient is a shared HTTP client for making JSON API requests.
// It provides a standardized way to handle HTTP interactions for LLM providers.
type JSONClient struct {
	httpClient *nethttp.Client
	baseURL    string
}

// ClientConfig holds configuration for the JSONClient.
type ClientConfig struct {
	// BaseURL is the base URL for the API (e.g. "https://api.openai.com/v1").
	BaseURL string
	// Timeout is the maximum duration for a request.
	Timeout time.Duration
}

// NewJSONClient creates a new JSONClient with the given configuration.
// If timeout is 0, it defaults to 300 seconds.
func NewJSONClient(cfg ClientConfig) *JSONClient {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 300 * time.Second
	}
	return &JSONClient{
		httpClient: &nethttp.Client{Timeout: timeout},
		baseURL:    cfg.BaseURL,
	}
}

// PostJSON sends a POST request with the given body marshaled as JSON.
// It sets common headers, handles the request, and unmarshals the response into target.
// Returns an error if the request fails or the status code is not OK.
func (c *JSONClient) PostJSON(ctx context.Context, path string, headers map[string]string, body interface{}, target interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + path
	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != nethttp.StatusOK {
		return &HTTPError{StatusCode: resp.StatusCode, Body: respBody}
	}

	if err := json.Unmarshal(respBody, target); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

// PostJSONWithProgress posts JSON with progress reporting.
// It behaves like PostJSON but invokes the progressFn callback during the operation.
func (c *JSONClient) PostJSONWithProgress(ctx context.Context, path string, headers map[string]string, body interface{}, target interface{}, progressFn ProgressCallback) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + path
	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != nethttp.StatusOK {
		return respBody, &HTTPError{StatusCode: resp.StatusCode, Body: respBody}
	}

	if err := json.Unmarshal(respBody, target); err != nil {
		return respBody, fmt.Errorf("failed to parse response: %w", err)
	}

	if progressFn != nil {
		// Calculate content size for progress
		contentSize := len(jsonBody)
		progressFn("uploading", fmt.Sprintf("sending %d bytes", contentSize), int64(contentSize), int64(contentSize))
	}

	return respBody, nil
}

// ProgressCallback is a function type for progress reporting.
// stage: The current operation stage (e.g., "uploading", "processing").
// message: A human-readable progress message.
// current: The current progress count.
// total: The total expected count (or -1 if unknown).
type ProgressCallback func(stage string, message string, current, total int64)

// HTTPError represents an error from an HTTP request.
// It captures the status code and the response body.
type HTTPError struct {
	StatusCode int
	Body       []byte
}

// Error returns the error string including status code and body.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, string(e.Body))
}
