package geminiapi

import (
	"context"
	"fmt"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llmbase"
)

const (
	defaultBaseURL   = "https://generativelanguage.googleapis.com/v1beta"
	defaultMaxTokens = 8192
)

// Client is the Google Gemini API implementation of the LLM provider.
type Client struct {
	*llmbase.BaseClient
}

// NewClient creates a new Gemini API client with the given configuration.
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
		model = "gemini-2.5-flash"
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
		}, "Gemini"),
	}, nil
}

// Send sends a content string to the Gemini API and returns the result.
func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error) {
	result, err := c.BaseClient.Send(ctx, content, c)
	if err != nil {
		return nil, c.handleError(err)
	}
	return result, nil
}

// SendWithProgress sends a content string to the Gemini API and reports progress via the callback.
func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	result, err := c.BaseClient.SendWithProgress(ctx, content, c, progress)
	if err != nil {
		return nil, c.handleError(err)
	}
	return result, nil
}

// BuildRequest constructs the Gemini-specific request body.
func (c *Client) BuildRequest(content string) (interface{}, error) {
	return GenerateRequest{
		Contents: []Content{
			{
				Parts: []Part{{Text: content}},
			},
		},
		GenerationConfig: &GenerationConfig{
			MaxOutputTokens: c.MaxTokens,
		},
	}, nil
}

// ParseResponse extracts the result from the Gemini API response.
func (c *Client) ParseResponse(response interface{}, rawJSON []byte) (*llm.Result, error) {
	genResp, ok := response.(*GenerateResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	if genResp.Error != nil {
		return nil, fmt.Errorf("API error [%d]: %s", genResp.Error.Code, genResp.Error.Message)
	}

	if len(genResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	var responseText string
	for _, part := range genResp.Candidates[0].Content.Parts {
		responseText += part.Text
	}

	var usage *llm.Usage
	if genResp.UsageMetadata != nil {
		usage = &llm.Usage{
			PromptTokens:     genResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: genResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      genResp.UsageMetadata.TotalTokenCount,
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

// GetEndpoint returns the Gemini API endpoint including the model and API key.
func (c *Client) GetEndpoint() string {
	return fmt.Sprintf("/models/%s:generateContent?key=%s", c.Model, c.APIKey)
}

// GetHeaders returns the necessary headers for Gemini API requests.
// For Gemini, authentication is handled via query parameters, so this returns an empty map.
func (c *Client) GetHeaders() map[string]string {
	return map[string]string{}
}

// NewResponse returns a new instance of the Gemini response structure for unmarshaling.
func (c *Client) NewResponse() interface{} {
	return &GenerateResponse{}
}

// GetProviderName returns the display name of the provider ("Gemini").
func (c *Client) GetProviderName() string {
	return c.ProviderName
}

func (c *Client) handleError(err error) error {
	return c.BaseClient.HandleHTTPError(err, func(body []byte) string {
		return ""
	})
}
