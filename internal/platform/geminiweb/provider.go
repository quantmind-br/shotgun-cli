package geminiweb

import (
	"context"
	"fmt"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

// WebProvider adapts the existing Executor to the llm.Provider interface.
type WebProvider struct {
	executor *Executor
	config   Config
}

// NewWebProvider creates a GeminiWeb provider.
func NewWebProvider(cfg llm.Config) (*WebProvider, error) {
	execCfg := Config{
		BinaryPath:     cfg.BinaryPath,
		Model:          cfg.Model,
		Timeout:        cfg.Timeout,
		BrowserRefresh: cfg.BrowserRefresh,
	}

	// Apply defaults if not set.
	if execCfg.Model == "" {
		execCfg.Model = "gemini-2.5-flash"
	}
	if execCfg.Timeout == 0 {
		execCfg.Timeout = 300
	}

	return &WebProvider{
		executor: NewExecutor(execCfg),
		config:   execCfg,
	}, nil
}

// Send sends a prompt and returns the response.
func (p *WebProvider) Send(ctx context.Context, content string) (*llm.Result, error) {
	result, err := p.executor.Send(ctx, content)
	if err != nil {
		return nil, err
	}

	return &llm.Result{
		Response:    result.Response,
		RawResponse: result.RawResponse,
		Model:       result.Model,
		Provider:    "GeminiWeb",
		Duration:    result.Duration,
	}, nil
}

// SendWithProgress sends with progress callback.
func (p *WebProvider) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	result, err := p.executor.SendWithProgress(ctx, content, progress)
	if err != nil {
		return nil, err
	}

	return &llm.Result{
		Response:    result.Response,
		RawResponse: result.RawResponse,
		Model:       result.Model,
		Provider:    "GeminiWeb",
		Duration:    result.Duration,
	}, nil
}

// Name returns the provider name.
func (p *WebProvider) Name() string {
	return "GeminiWeb"
}

// IsAvailable checks if geminiweb binary is available.
func (p *WebProvider) IsAvailable() bool {
	return IsAvailable()
}

// IsConfigured checks if geminiweb is configured.
func (p *WebProvider) IsConfigured() bool {
	return IsConfigured()
}

// ValidateConfig validates the configuration.
func (p *WebProvider) ValidateConfig() error {
	if !p.IsAvailable() {
		return fmt.Errorf("geminiweb binary not found. Install with: go install github.com/diogo/geminiweb/cmd/geminiweb@latest")
	}
	if !p.IsConfigured() {
		return fmt.Errorf("geminiweb not configured. Run: geminiweb auto-login")
	}
	return nil
}
