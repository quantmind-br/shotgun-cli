package geminiapi

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
	defaultBaseURL   = "https://generativelanguage.googleapis.com/v1beta"
	defaultMaxTokens = 8192
)

// Client implements llm.Provider for Google Gemini API.
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	timeout    time.Duration
	maxTokens  int
	httpClient *http.Client
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

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		c.baseURL, c.model, c.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var genResp GenerateResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
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

	return &llm.Result{
		Response:    responseText,
		RawResponse: string(respBody),
		Model:       c.model,
		Provider:    "Gemini",
		Duration:    time.Since(startTime),
		Usage:       usage,
	}, nil
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
