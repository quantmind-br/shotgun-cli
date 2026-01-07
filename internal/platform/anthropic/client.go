package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/http"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llm"
)

const (
	defaultBaseURL   = "https://api.anthropic.com"
	anthropicVersion = "2023-06-01"
	defaultMaxTokens = 8192
)

// Client implements llm.Provider for Anthropic API.
type Client struct {
	jsonClient *platformhttp.JSONClient
	apiKey     string
	model      string
	maxTokens  int
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
		jsonClient: platformhttp.NewJSONClient(platformhttp.ClientConfig{
			BaseURL: baseURL,
			Timeout: timeout,
		}),
		apiKey:    cfg.APIKey,
		model:     model,
		maxTokens: maxTokens,
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

	headers := map[string]string{
		"x-api-key":         c.apiKey,
		"anthropic-version": anthropicVersion,
	}

	var msgResp MessagesResponse
	err := c.jsonClient.PostJSON(ctx, "/v1/messages", headers, req, &msgResp)
	if err != nil {
		return nil, c.handleError(err)
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

	rawResp, _ := json.Marshal(msgResp)

	return &llm.Result{
		Response:    responseText,
		RawResponse: string(rawResp),
		Model:       c.model,
		Provider:    "Anthropic",
		Duration:    time.Since(startTime),
		Usage:       usage,
	}, nil
}

func (c *Client) handleError(err error) error {
	if httpErr, ok := err.(*platformhttp.HTTPError); ok {
		var errResp ErrorResponse
		if json.Unmarshal(httpErr.Body, &errResp) == nil && errResp.Error.Message != "" {
			return fmt.Errorf("API error [%d]: %s", httpErr.StatusCode, errResp.Error.Message)
		}
		return fmt.Errorf("API error [%d]: %s", httpErr.StatusCode, string(httpErr.Body))
	}
	return err
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
