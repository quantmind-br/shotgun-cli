package geminiapi

import (
	"context"
	"encoding/json"
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
	base, err := llmbase.NewBaseClientWithDefaults(cfg, llmbase.DefaultConfig{
		BaseURL:   defaultBaseURL,
		Model:     "gemini-2.5-flash",
		MaxTokens: defaultMaxTokens,
		Timeout:   300 * time.Second,
	}, "Gemini")
	if err != nil {
		return nil, err
	}

	return &Client{BaseClient: base}, nil
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

// handleError parses the error body Gemini returns on a non-200 status.
//
// It used to return "" unconditionally, so HandleHTTPError fell back to printing
// the raw body: the user saw the code and message wrapped in JSON noise, while
// OpenAI and Anthropic showed a clean message for the same class of failure.
// ParseResponse already parsed this shape for errors delivered inside a 200.
func (c *Client) handleError(err error) error {
	return c.HandleHTTPError(err, func(body []byte) string {
		var errResp struct {
			Error *APIError `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != nil && errResp.Error.Message != "" {
			return errResp.Error.Message
		}

		return ""
	})
}
