package llmbase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	platformhttp "github.com/quantmind-br/shotgun-cli/internal/platform/http"
)

// BaseClient provides common functionality for HTTP-based LLM providers.
type BaseClient struct {
	JSONClient   *platformhttp.JSONClient
	APIKey       string
	Model        string
	MaxTokens    int
	ProviderName string
}

// Config holds the configuration for creating a BaseClient.
type Config struct {
	APIKey    string
	BaseURL   string
	Model     string
	MaxTokens int
	Timeout   time.Duration
}

// NewBaseClient creates a new BaseClient with the given configuration.
func NewBaseClient(cfg Config, providerName string) *BaseClient {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	return &BaseClient{
		JSONClient: platformhttp.NewJSONClient(platformhttp.ClientConfig{
			BaseURL: cfg.BaseURL,
			Timeout: timeout,
		}),
		APIKey:       cfg.APIKey,
		Model:        cfg.Model,
		MaxTokens:    cfg.MaxTokens,
		ProviderName: providerName,
	}
}

// Name returns the provider name.
func (c *BaseClient) Name() string {
	return c.ProviderName
}

// IsAvailable returns true for HTTP-based providers (always available if there's internet).
func (c *BaseClient) IsAvailable() bool {
	return true
}

// IsConfigured checks if the provider is configured with required fields.
func (c *BaseClient) IsConfigured() bool {
	return c.APIKey != "" && c.Model != ""
}

// ValidateConfig validates the configuration.
func (c *BaseClient) ValidateConfig() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}

// Send sends content to the LLM using the provided Sender strategy.
func (c *BaseClient) Send(ctx context.Context, content string, sender Sender) (*llm.Result, error) {
	startTime := time.Now()

	reqBody, err := sender.BuildRequest(content)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	headers := sender.GetHeaders()
	endpoint := sender.GetEndpoint()
	response := sender.NewResponse()

	err = c.JSONClient.PostJSON(ctx, endpoint, headers, reqBody, response)
	if err != nil {
		return nil, err
	}

	rawJSON, _ := json.Marshal(response)

	result, err := sender.ParseResponse(response, rawJSON)
	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// SendWithProgress sends content with progress callback.
func (c *BaseClient) SendWithProgress(ctx context.Context, content string, sender Sender, progress func(stage string)) (*llm.Result, error) {
	progress(fmt.Sprintf("Connecting to %s...", c.ProviderName))
	result, err := c.Send(ctx, content, sender)
	if err == nil {
		progress("Response received")
	}
	return result, err
}

// HandleHTTPError converts platformhttp.HTTPError to a formatted error message.
func (c *BaseClient) HandleHTTPError(err error, parseBody func([]byte) string) error {
	if httpErr, ok := err.(*platformhttp.HTTPError); ok {
		if msg := parseBody(httpErr.Body); msg != "" {
			return fmt.Errorf("API error [%d]: %s", httpErr.StatusCode, msg)
		}
		return fmt.Errorf("API error [%d]: %s", httpErr.StatusCode, string(httpErr.Body))
	}
	return err
}
