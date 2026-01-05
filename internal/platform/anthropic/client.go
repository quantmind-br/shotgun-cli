package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

const (
	defaultBaseURL   = "https://api.anthropic.com"
	anthropicVersion = "2023-06-01"
	defaultMaxTokens = 8192
)

// Client implements llm.Provider for Anthropic API.
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	timeout    time.Duration
	maxTokens  int
	httpClient *http.Client
}

// NewClient creates a new Anthropic client.
func NewClient(cfg llm.Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	model := cfg.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}

	return &Client{
		apiKey:    cfg.APIKey,
		baseURL:   baseURL,
		model:     model,
		timeout:   timeout,
		maxTokens: maxTokens,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Send sends a prompt and returns the response.
func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error) {
	startTime := time.Now()

	req := MessagesRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages: []Message{
			{Role: "user", Content: content},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, string(respBody))
	}

	var msgResp MessagesResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text from content blocks.
	var responseText string
	for _, block := range msgResp.Content {
		if block.Type == "text" {
			responseText += block.Text
		}
	}

	var usage *llm.Usage
	if msgResp.Usage.InputTokens > 0 || msgResp.Usage.OutputTokens > 0 {
		usage = &llm.Usage{
			PromptTokens:     msgResp.Usage.InputTokens,
			CompletionTokens: msgResp.Usage.OutputTokens,
			TotalTokens:      msgResp.Usage.InputTokens + msgResp.Usage.OutputTokens,
		}
	}

	return &llm.Result{
		Response:    responseText,
		RawResponse: string(respBody),
		Model:       c.model,
		Provider:    "Anthropic",
		Duration:    time.Since(startTime),
		Usage:       usage,
	}, nil
}

// SendWithProgress sends with progress callback.
func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	progress("Connecting to Anthropic...")
	result, err := c.Send(ctx, content)
	if err == nil {
		progress("Response received")
	}
	return result, err
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "Anthropic"
}

// IsAvailable checks if the provider is available.
func (c *Client) IsAvailable() bool {
	return true
}

// IsConfigured checks if the provider is configured.
func (c *Client) IsConfigured() bool {
	return c.apiKey != "" && c.model != ""
}

// ValidateConfig validates the configuration.
func (c *Client) ValidateConfig() error {
	if c.apiKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.model == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}
