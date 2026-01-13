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

type Client struct {
	*llmbase.BaseClient
}

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

func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error) {
	result, err := c.BaseClient.Send(ctx, content, c)
	if err != nil {
		return nil, c.handleError(err)
	}
	return result, nil
}

func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	result, err := c.BaseClient.SendWithProgress(ctx, content, c, progress)
	if err != nil {
		return nil, c.handleError(err)
	}
	return result, nil
}

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

func (c *Client) GetEndpoint() string {
	return fmt.Sprintf("/models/%s:generateContent?key=%s", c.Model, c.APIKey)
}

func (c *Client) GetHeaders() map[string]string {
	return map[string]string{}
}

func (c *Client) NewResponse() interface{} {
	return &GenerateResponse{}
}

func (c *Client) GetProviderName() string {
	return c.ProviderName
}

func (c *Client) handleError(err error) error {
	return c.BaseClient.HandleHTTPError(err, func(body []byte) string {
		return ""
	})
}
