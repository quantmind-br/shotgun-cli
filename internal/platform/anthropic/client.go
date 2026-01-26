package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llmbase"
)

const (
	defaultBaseURL   = "https://api.anthropic.com"
	anthropicVersion = "2023-06-01"
	defaultMaxTokens = 8192
)

// Client is the Anthropic implementation of the LLM provider.
type Client struct {
	*llmbase.BaseClient
}

// NewClient creates a new Anthropic client with the given configuration.
// It validates the API key and sets default values for BaseURL, Model, MaxTokens, and Timeout if not provided.
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
		BaseClient: llmbase.NewBaseClient(llmbase.Config{
			APIKey:    cfg.APIKey,
			BaseURL:   baseURL,
			Model:     model,
			MaxTokens: maxTokens,
			Timeout:   timeout,
		}, "Anthropic"),
	}, nil
}

// Send sends a content string to the Anthropic API and returns the result.
func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error) {
	result, err := c.BaseClient.Send(ctx, content, c)
	if err != nil {
		return nil, c.handleError(err)
	}
	return result, nil
}

// SendWithProgress sends a content string to the Anthropic API and reports progress via the callback.
func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	result, err := c.BaseClient.SendWithProgress(ctx, content, c, progress)
	if err != nil {
		return nil, c.handleError(err)
	}
	return result, nil
}

// BuildRequest constructs the Anthropic-specific request body.
func (c *Client) BuildRequest(content string) (interface{}, error) {
	return MessagesRequest{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		Messages: []Message{
			{Role: "user", Content: content},
		},
	}, nil
}

// ParseResponse extracts the result from the Anthropic API response.
func (c *Client) ParseResponse(response interface{}, rawJSON []byte) (*llm.Result, error) {
	msgResp, ok := response.(*MessagesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

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
		RawResponse: string(rawJSON),
		Model:       c.Model,
		Provider:    c.ProviderName,
		Usage:       usage,
	}, nil
}

// GetEndpoint returns the Anthropic Messages API endpoint.
func (c *Client) GetEndpoint() string {
	return "/v1/messages"
}

// GetHeaders returns the necessary headers for Anthropic API requests, including the API key and version.
func (c *Client) GetHeaders() map[string]string {
	return map[string]string{
		"x-api-key":         c.APIKey,
		"anthropic-version": anthropicVersion,
	}
}

// NewResponse returns a new instance of the Anthropic response structure for unmarshaling.
func (c *Client) NewResponse() interface{} {
	return &MessagesResponse{}
}

// GetProviderName returns the display name of the provider ("Anthropic").
func (c *Client) GetProviderName() string {
	return c.ProviderName
}

func (c *Client) handleError(err error) error {
	return c.BaseClient.HandleHTTPError(err, func(body []byte) string {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return errResp.Error.Message
		}
		return ""
	})
}
