package app

import (
	"context"
	"fmt"
	"os"

	ctxgen "github.com/quantmind-br/shotgun-cli/internal/core/context"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/tokens"
	"github.com/quantmind-br/shotgun-cli/internal/platform/clipboard"
)

// DefaultContextService implements the ContextService interface.
// It orchestrates the context generation workflow by coordinating the scanner,
// generator, and LLM provider components.
type DefaultContextService struct {
	scanner   scanner.Scanner
	generator ctxgen.ContextGenerator
	registry  *llm.Registry
}

// ServiceOption defines a functional option for configuring the DefaultContextService.
type ServiceOption func(*DefaultContextService)

// NewContextService creates a new DefaultContextService with the provided options.
// If no options are provided, it uses default implementations for scanner, generator, and registry.
func NewContextService(opts ...ServiceOption) *DefaultContextService {
	svc := &DefaultContextService{
		scanner:   scanner.NewFileSystemScanner(),
		generator: ctxgen.NewDefaultContextGenerator(),
		registry:  DefaultProviderRegistry,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// WithRegistry configures the service to use a specific LLM provider registry.
func WithRegistry(r *llm.Registry) ServiceOption {
	return func(svc *DefaultContextService) {
		svc.registry = r
	}
}

// WithScanner configures the service to use a specific scanner implementation.
func WithScanner(s scanner.Scanner) ServiceOption {
	return func(svc *DefaultContextService) {
		svc.scanner = s
	}
}

// WithGenerator configures the service to use a specific context generator implementation.
func WithGenerator(g ctxgen.ContextGenerator) ServiceOption {
	return func(svc *DefaultContextService) {
		svc.generator = g
	}
}

// Generate generates a codebase context synchronously.
// It delegates to GenerateWithProgress with a nil callback.
func (s *DefaultContextService) Generate(ctx context.Context, cfg GenerateConfig) (*GenerateResult, error) {
	return s.GenerateWithProgress(ctx, cfg, nil)
}

// GenerateWithProgress generates a codebase context and reports progress via the callback.
// It performs the following steps:
// 1. Validates configuration
// 2. Scans the filesystem (reporting progress)
// 3. Applies selections (defaulting to all if none provided)
// 4. Generates context content (reporting progress)
// 5. Enforces size limits
// 6. Saves output to file
// 7. Optionally copies to clipboard
func (s *DefaultContextService) GenerateWithProgress(
	ctx context.Context,
	cfg GenerateConfig,
	progress ProgressCallback,
) (*GenerateResult, error) {
	report := func(stage, msg string, cur, total int64) {
		if progress != nil {
			progress(stage, msg, cur, total)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	scanConfig := cfg.ScanConfig
	if scanConfig == nil {
		scanConfig = scanner.DefaultScanConfig()
	}

	report("scanning", "Scanning files...", 0, 0)

	var tree *scanner.FileNode
	var err error

	if progress != nil {
		progressCh := make(chan scanner.Progress, 100)
		done := make(chan struct{})

		go func() {
			defer close(done)
			for p := range progressCh {
				report("scanning", p.Message, p.Current, p.Total)
			}
		}()

		tree, err = s.scanner.ScanWithProgress(cfg.RootPath, scanConfig, progressCh)
		close(progressCh)
		<-done
	} else {
		tree, err = s.scanner.Scan(cfg.RootPath, scanConfig)
	}

	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	selections := cfg.Selections
	if selections == nil {
		selections = scanner.NewSelectAll(tree)
	}

	report("generating", "Generating context...", 0, 0)

	genConfig := ctxgen.GenerateConfig{
		MaxTotalSize:   cfg.MaxSize,
		TemplateVars:   cfg.TemplateVars,
		Template:       cfg.Template,
		SkipBinary:     cfg.SkipBinary,
		IncludeTree:    cfg.IncludeTree,
		IncludeSummary: cfg.IncludeSummary,
	}

	var content string
	if progress != nil {
		content, err = s.generator.GenerateWithProgressEx(tree, selections, genConfig, func(p ctxgen.GenProgress) {
			report("generating", p.Message, 0, 0)
		})
	} else {
		content, err = s.generator.Generate(tree, selections, genConfig)
	}

	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	contentSize := int64(len(content))
	if cfg.EnforceLimit && cfg.MaxSize > 0 && contentSize > cfg.MaxSize {
		return nil, fmt.Errorf("content size (%d) exceeds limit (%d)", contentSize, cfg.MaxSize)
	}

	report("saving", "Saving output...", 0, 0)

	outputPath := cfg.GenerateOutputPath()
	if err := os.WriteFile(outputPath, []byte(content), 0600); err != nil {
		return nil, fmt.Errorf("failed to save output: %w", err)
	}

	copied := false
	if cfg.CopyToClipboard {
		if err := clipboard.Copy(content); err == nil {
			copied = true
		}
	}

	report("complete", "Done", 1, 1)

	return &GenerateResult{
		Content:           content,
		OutputPath:        outputPath,
		FileCount:         tree.CountFiles(),
		ContentSize:       contentSize,
		TokenEstimate:     int64(tokens.EstimateFromBytes(contentSize)),
		CopiedToClipboard: copied,
	}, nil
}

// SendToLLM sends content to an LLM provider synchronously.
// It checks provider availability and configuration before sending.
func (s *DefaultContextService) SendToLLM(
	ctx context.Context,
	content string,
	provider llm.Provider,
) (*llm.Result, error) {
	if !provider.IsAvailable() {
		return nil, fmt.Errorf("%s not available", provider.Name())
	}
	if err := provider.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid provider config: %w", err)
	}
	return provider.Send(ctx, content)
}

// SendToLLMWithProgress sends content to an LLM provider with progress reporting.
// It handles:
// 1. Provider creation from config
// 2. Availability and config validation
// 3. Sending with progress callback
// 4. Optionally saving the response to a file
func (s *DefaultContextService) SendToLLMWithProgress(
	ctx context.Context,
	content string,
	cfg LLMSendConfig,
	progress LLMProgressCallback,
) (*llm.Result, error) {
	llmCfg := llm.Config{
		Provider: cfg.Provider,
		APIKey:   cfg.APIKey,
		BaseURL:  cfg.BaseURL,
		Model:    cfg.Model,
		Timeout:  cfg.Timeout,
	}
	llmCfg.WithDefaults()

	provider, err := s.registry.Create(llmCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	if !provider.IsAvailable() {
		return nil, fmt.Errorf("%s not available", provider.Name())
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid provider config: %w", err)
	}

	var result *llm.Result
	if progress != nil {
		result, err = provider.SendWithProgress(ctx, content, progress)
	} else {
		result, err = provider.Send(ctx, content)
	}

	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}

	if cfg.SaveResponse && cfg.OutputPath != "" {
		if writeErr := os.WriteFile(cfg.OutputPath, []byte(result.Response), 0600); writeErr != nil {
			return result, fmt.Errorf("failed to save response: %w", writeErr)
		}
	}

	return result, nil
}

// Scanner returns the underlying scanner instance.
// This is useful for tests or components that need direct access to the scanner.
func (s *DefaultContextService) Scanner() scanner.Scanner {
	return s.scanner
}
