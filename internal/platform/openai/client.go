package openai

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

const defaultBaseURL = "https://api.openai.com/v1"

// Client implements llm.Provider for OpenAI-compatible APIs.
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	timeout    time.Duration
	maxTokens  int
	httpClient *http.Client
}

// NewClient creates a new OpenAI client.
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
		model = "gpt-4o"
	}

	return &Client{
		apiKey:    cfg.APIKey,
		baseURL:   baseURL,
		model:     model,
		timeout:   timeout,
		maxTokens: cfg.MaxTokens,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Send sends a prompt and returns the response.
func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error) {
	startTime := time.Now()

	req := ChatCompletionRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "user", Content: content},
		},
	}

	if c.maxTokens > 0 {
		req.MaxTokens = c.maxTokens
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

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

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	var usage *llm.Usage
	if chatResp.Usage.TotalTokens > 0 {
		usage = &llm.Usage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		}
	}

	return &llm.Result{
		Response:    chatResp.Choices[0].Message.Content,
		RawResponse: string(respBody),
		Model:       c.model,
		Provider:    "OpenAI",
		Duration:    time.Since(startTime),
		Usage:       usage,
	}, nil
}

// SendWithProgress sends with progress callback.
func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	progress("Connecting to OpenAI...")
	result, err := c.Send(ctx, content)
	if err == nil {
		progress("Response received")
	}
	return result, err
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "OpenAI"
}

// IsAvailable checks if the provider is available.
func (c *Client) IsAvailable() bool {
	return true // API always available if there's internet
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
