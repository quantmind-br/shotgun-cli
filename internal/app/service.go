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

type DefaultContextService struct {
	scanner   scanner.Scanner
	generator ctxgen.ContextGenerator
}

type ServiceOption func(*DefaultContextService)

func NewContextService(opts ...ServiceOption) *DefaultContextService {
	svc := &DefaultContextService{
		scanner:   scanner.NewFileSystemScanner(),
		generator: ctxgen.NewDefaultContextGenerator(),
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func WithScanner(s scanner.Scanner) ServiceOption {
	return func(svc *DefaultContextService) {
		svc.scanner = s
	}
}

func WithGenerator(g ctxgen.ContextGenerator) ServiceOption {
	return func(svc *DefaultContextService) {
		svc.generator = g
	}
}

func (s *DefaultContextService) Generate(ctx context.Context, cfg GenerateConfig) (*GenerateResult, error) {
	return s.GenerateWithProgress(ctx, cfg, nil)
}

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

func (s *DefaultContextService) Scanner() scanner.Scanner {
	return s.scanner
}
