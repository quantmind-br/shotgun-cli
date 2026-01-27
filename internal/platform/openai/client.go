package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llmbase"
)

const defaultBaseURL = "https://api.openai.com/v1"

// Client implements llm.Provider for OpenAI-compatible APIs.
type Client struct {
	*llmbase.BaseClient
}

// NewClient creates a new OpenAI client.
func NewClient(cfg llm.Config) (*Client, error) {
	base, err := llmbase.NewBaseClientWithDefaults(cfg, llmbase.DefaultConfig{
		BaseURL: defaultBaseURL,
		Model:   "gpt-4o",
		Timeout: 300 * time.Second,
	}, "OpenAI")
	if err != nil {
		return nil, err
	}

	return &Client{BaseClient: base}, nil
}

// Send sends a prompt and returns the response.
func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error) {
	result, err := c.BaseClient.Send(ctx, content, c)
	if err != nil {
		return nil, c.handleError(err)
	}
	return result, nil
}

// SendWithProgress sends with progress callback.
func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	result, err := c.BaseClient.SendWithProgress(ctx, content, c, progress)
	if err != nil {
		return nil, c.handleError(err)
	}
	return result, nil
}

// BuildRequest creates the OpenAI-specific request payload.
func (c *Client) BuildRequest(content string) (interface{}, error) {
	req := ChatCompletionRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "user", Content: content},
		},
	}
	if c.MaxTokens > 0 {
		req.MaxTokens = c.MaxTokens
	}
	return req, nil
}

// ParseResponse extracts llm.Result from OpenAI's response.
func (c *Client) ParseResponse(response interface{}, rawJSON []byte) (*llm.Result, error) {
	chatResp, ok := response.(*ChatCompletionResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
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
		RawResponse: string(rawJSON),
		Model:       c.Model,
		Provider:    c.ProviderName,
		Usage:       usage,
	}, nil
}

// GetEndpoint returns the API endpoint path.
func (c *Client) GetEndpoint() string {
	return "/chat/completions"
}

// GetHeaders returns OpenAI-specific HTTP headers.
func (c *Client) GetHeaders() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + c.APIKey,
	}
}

// NewResponse returns a new ChatCompletionResponse for unmarshaling.
func (c *Client) NewResponse() interface{} {
	return &ChatCompletionResponse{}
}

// GetProviderName returns the provider display name.
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
