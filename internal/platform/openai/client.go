package openai

import (
	"context"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/http"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/openai"
)

type Client struct {
	*llm.BaseClient
	sender *sender
}

// NewClient creates a new OpenAI client
func NewClient(cfg llm.Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key is required")
	}

	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	baseClient := llm.NewBaseClient(
		&http.JSONClient{
			Timeout: timeout,
		},
		llm.Config{
			APIKey:       cfg.APIKey,
			Model:        cfg.Model,
			MaxTokens:    cfg.MaxTokens,
			ProviderName: "OpenAI",
			BaseURL:      cfg.BaseURL,
		},
	)

	sender := &sender{
		client: baseClient,
		model:  cfg.Model,
	}

	return &Client{
		BaseClient: baseClient,
		sender:     sender,
	}
}

func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error) {
	return c.BaseClient.Send(ctx, content)
}

func (c *Client) SendWithProgress(ctx context.Context, content string, progress llm.ProgressCallback) (*llm.Result, error) {
	return c.BaseClient.SendWithProgress(ctx, content, progress)
}

func (c *Client) Name() string {
	return "OpenAI"
}

func (c *Client) IsAvailable() bool {
	return c.BaseClient.IsAvailable()
}

func (c *Client) IsConfigured() bool {
	return c.BaseClient.IsConfigured()
}

func (c *Client) ValidateConfig() error {
	return c.BaseClient.ValidateConfig()
}
