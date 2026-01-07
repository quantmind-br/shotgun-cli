package anthropic

import (
	"context"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/platform/anthropic"
	"github.com/quantmind-br/shotgun-cli/internal/platform/http"
	"github.com/quantmind-br/shotgun-cli/internal/platform/llm"
)

type client struct {
	*llm.BaseClient
	sender *sender
}

// NewClient creates a new Anthropic client
func NewClient(cfg llm.Config) (*client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key is required")
	}

	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	sender := &sender{
		client: llm.NewBaseClient(
			&http.JSONClient{
				Timeout: timeout,
				BaseURL: cfg.BaseURL,
			},
			llm.Config{
				APIKey:       cfg.APIKey,
				Model:        cfg.Model,
				MaxTokens:    cfg.MaxTokens,
				ProviderName: "Anthropic",
				BaseURL:      cfg.BaseURL,
			},
		),
		model: cfg.Model,
	}

	return &client{
		BaseClient: sender.client,
		sender:     sender,
	}
}

func (c *client) Send(ctx context.Context, content string) (*llm.Result, error) {
	return c.BaseClient.Send(ctx, content)
}

func (c *client) SendWithProgress(ctx context.Context, content string, progress llm.ProgressCallback) (*llm.Result, error) {
	return c.BaseClient.SendWithProgress(ctx, content, progress)
}

func (c *client) Name() string {
	return "Anthropic"
}

func (c *client) IsAvailable() bool {
	return c.BaseClient.IsAvailable()
}

func (c *client) IsConfigured() bool {
	return c.BaseClient.IsConfigured()
}

func (c *client) ValidateConfig() error {
	return c.BaseClient.ValidateConfig()
}
