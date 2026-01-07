package geminiapi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	platformhttp "github.com/quantmind-br/shotgun-cli/internal/platform/http"
)

const (
	defaultBaseURL   = "https://generativelanguage.googleapis.com/v1beta"
	defaultMaxTokens = 8192
)

// Client implements llm.Provider for Google Gemini API.
type Client struct {
	jsonClient *platformhttp.JSONClient
	apiKey     string
	model      string
	maxTokens  int
}

// NewClient creates a new Gemini API client.
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

	req := GenerateRequest{
		Contents: []Content{
			{
				Parts: []Part{{Text: content}},
			},
		},
		GenerationConfig: &GenerationConfig{
			MaxOutputTokens: c.maxTokens,
		},
	}

	path := fmt.Sprintf("/models/%s:generateContent?key=%s", c.model, c.apiKey)

	var genResp GenerateResponse
	err := c.jsonClient.PostJSON(ctx, path, nil, req, &genResp)
	if err != nil {
		return nil, c.handleError(err)
	}

	if genResp.Error != nil {
		return nil, fmt.Errorf("API error [%d]: %s", genResp.Error.Code, genResp.Error.Message)
	}

	if len(genResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	// Extract text from parts.
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

	rawResp, _ := json.Marshal(genResp)

	return &llm.Result{
		Response:    responseText,
		RawResponse: string(rawResp),
		Model:       c.model,
		Provider:    "Gemini",
		Duration:    time.Since(startTime),
		Usage:       usage,
	}, nil
}

func (c *Client) handleError(err error) error {
	if httpErr, ok := err.(*platformhttp.HTTPError); ok {
		var genResp GenerateResponse
		if json.Unmarshal(httpErr.Body, &genResp) == nil && genResp.Error != nil {
			return fmt.Errorf("API error [%d]: %s", genResp.Error.Code, genResp.Error.Message)
		}
		return fmt.Errorf("API error [%d]: %s", httpErr.StatusCode, string(httpErr.Body))
	}
	return err
}

// SendWithProgress sends with progress callback.
func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	progress("Connecting to Gemini API...")
	result, err := c.Send(ctx, content)
	if err == nil {
		progress("Response received")
	}
	return result, err
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "Gemini"
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
