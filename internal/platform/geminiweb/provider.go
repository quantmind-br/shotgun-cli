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
	runner   CommandRunner // Optional custom runner for testing
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

	exec := NewExecutor(execCfg)
	return &WebProvider{
		executor: exec,
		config:   execCfg,
		runner:   nil,
	}, nil
}

// NewWebProviderWithRunner creates a GeminiWeb provider with a custom command runner.
// This is primarily used for testing.
func NewWebProviderWithRunner(cfg llm.Config, runner CommandRunner) (*WebProvider, error) {
	provider, err := NewWebProvider(cfg)
	if err != nil {
		return nil, err
	}
	provider.runner = runner
	provider.executor = NewExecutorWithRunner(provider.config, runner)
	return provider, nil
}

// getRunner returns the provider's runner, or the default if none is set.
func (p *WebProvider) getRunner() CommandRunner {
	if p.runner == nil {
		return GetDefaultRunner()
	}
	return p.runner
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
	_, err := p.config.FindBinary(p.getRunner())
	return err == nil
}

// IsConfigured checks if geminiweb is configured.
func (p *WebProvider) IsConfigured() bool {
	return IsConfigured()
}

// ValidateConfig validates the configuration.
func (p *WebProvider) ValidateConfig() error {
	if !p.IsAvailable() {
		return fmt.Errorf("geminiweb binary not found. " +
			"Install with: go install github.com/diogo/geminiweb/cmd/geminiweb@latest")
	}
	if !p.IsConfigured() {
		return fmt.Errorf("geminiweb not configured. Run: geminiweb auto-login")
	}
	return nil
}
