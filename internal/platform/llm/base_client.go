package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/http",
)

// BaseClient provides common functionality for HTTP-based LLM providers using composition pattern
type BaseClient struct {
	JSONClient   *http.JSONClient
	APIKey       string
	Model        string
	MaxTokens    int
	ProviderName string
	BaseURL      string
	Timeout      time.Duration
	sender       Sender
}

// NewBaseClient creates a new BaseClient with configuration
func NewBaseClient(jsonClient *http.JSONClient, cfg Config) *BaseClient {
	return &BaseClient{
		JSONClient:   jsonClient,
		APIKey:       cfg.APIKey,
		Model:        cfg.Model,
		MaxTokens:    cfg.MaxTokens,
		ProviderName: cfg.ProviderName,
		BaseURL:      cfg.BaseURL,
		Timeout:      cfg.Timeout,
	}
}

// Name returns the provider name
func (c *BaseClient) Name() string {
	return c.ProviderName
}

// IsAvailable checks if the provider is available
func (c *BaseClient) IsAvailable() bool {
	return c.JSONClient != nil
}

// IsConfigured checks if the provider has required configuration
func (c *BaseClient) IsConfigured() bool {
	return c.APIKey != "" && c.Model != ""
}

// ValidateConfig validates the provider configuration
func (c *BaseClient) ValidateConfig() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required for %s provider", c.ProviderName)
	}
	if c.Model == "" {
		return fmt.Errorf("model is required for %s provider", c.ProviderName)
	}
	return nil
}

// Send sends content to the LLM provider
func (c *BaseClient) Send(ctx context.Context, content string) (*Result, error) {
	request, err := c.sender.BuildRequest(content)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	endpoint := c.sender.GetEndpoint()
	headers := c.sender.GetHeaders()

	responseData, err := c.JSONClient.PostJSON(ctx, endpoint, headers, request, c.sender.GetResponseType())
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return c.sender.ParseResponse(responseData)
}

// SendWithProgress sends content with progress reporting
func (c *BaseClient) SendWithProgress(ctx context.Context, content string, progressFn ProgressCallback) (*Result, error) {
	request, err := c.sender.BuildRequest(content)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	endpoint := c.sender.GetEndpoint()
	headers := c.sender.GetHeaders()

	responseData, err := c.JSONClient.PostJSONWithProgress(ctx, endpoint, headers, request, c.sender.GetResponseType(), func(stage string, current, total int64) {
		if progressFn != nil {
			progressFn(stage, fmt.Sprintf("%s (%d/%d)", stage, current, total)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return c.sender.ParseResponse(responseData)
}
