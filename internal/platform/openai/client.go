package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	platformhttp "github.com/quantmind-br/shotgun-cli/internal/platform/http"
)

const defaultBaseURL = "https://api.openai.com/v1"

// Client implements llm.Provider for OpenAI-compatible APIs.
type Client struct {
	jsonClient *platformhttp.JSONClient
	apiKey     string
	model      string
	maxTokens  int
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
		jsonClient: platformhttp.NewJSONClient(platformhttp.ClientConfig{
			BaseURL: baseURL,
			Timeout: timeout,
		}),
		apiKey:    cfg.APIKey,
		model:     model,
		maxTokens: cfg.MaxTokens,
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

	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	var chatResp ChatCompletionResponse
	err := c.jsonClient.PostJSON(ctx, "/chat/completions", headers, req, &chatResp)
	if err != nil {
		return nil, c.handleError(err)
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

	rawResp, _ := json.Marshal(chatResp)

	return &llm.Result{
		Response:    chatResp.Choices[0].Message.Content,
		RawResponse: string(rawResp),
		Model:       c.model,
		Provider:    "OpenAI",
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
