package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

// ContextService defines the interface for the application service layer.
// It provides high-level operations for context generation and LLM interaction,
// bridging the gap between the presentation layer (CLI/TUI) and the core domain logic.
type ContextService interface {
	// Generate generates a codebase context based on the provided configuration.
	// It scans files, applies templates, and produces a single string output.
	Generate(ctx context.Context, cfg GenerateConfig) (*GenerateResult, error)

	// GenerateWithProgress generates a codebase context and reports progress via the callback.
	// This is useful for interactive UIs that need to show scanning/generating status.
	GenerateWithProgress(
		ctx context.Context, cfg GenerateConfig, progress ProgressCallback,
	) (*GenerateResult, error)

	// SendToLLM sends the provided content to an LLM provider.
	// It handles provider validation and configuration before sending.
	SendToLLM(ctx context.Context, content string, provider llm.Provider) (*llm.Result, error)

	// SendToLLMWithProgress sends content to an LLM provider and reports progress.
	// It creates a new provider instance based on the configuration.
	SendToLLMWithProgress(
		ctx context.Context, content string, cfg LLMSendConfig, progress LLMProgressCallback,
	) (*llm.Result, error)
}

// LLMSendConfig holds configuration for sending content to an LLM provider.
type LLMSendConfig struct {
	Provider     llm.ProviderType
	APIKey       string
	BaseURL      string
	Model        string
	Timeout      int
	SaveResponse bool
	OutputPath   string
}

// LLMProgressCallback is a function type for receiving progress updates during LLM operations.
type LLMProgressCallback func(stage string)

// GenerateConfig holds configuration for the context generation process.
type GenerateConfig struct {
	RootPath        string
	ScanConfig      *scanner.ScanConfig
	Selections      map[string]bool
	Template        string
	TemplateVars    map[string]string
	MaxSize         int64
	EnforceLimit    bool
	OutputPath      string
	CopyToClipboard bool
	IncludeTree     bool
	IncludeSummary  bool
	SkipBinary      bool
}

// GenerateResult represents the result of a context generation operation.
type GenerateResult struct {
	Content           string
	OutputPath        string
	FileCount         int
	ContentSize       int64
	TokenEstimate     int64
	CopiedToClipboard bool
}

// ProgressCallback is a function type for receiving detailed progress updates
// including current/total counts and stage messages.
type ProgressCallback func(stage string, message string, current, total int64)

// Validate checks if the configuration is valid.
// It ensures RootPath is set and points to a valid directory.
func (c *GenerateConfig) Validate() error {
	if c.RootPath == "" {
		return fmt.Errorf("root path is required")
	}
	absPath, err := filepath.Abs(c.RootPath)
	if err != nil {
		return fmt.Errorf("invalid root path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("cannot access root path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("root path must be a directory")
	}
	c.RootPath = absPath
	return nil
}

// GenerateOutputPath returns the configured output path or generates a default one
// based on the current timestamp.
func (c *GenerateConfig) GenerateOutputPath() string {
	if c.OutputPath != "" {
		return c.OutputPath
	}
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("shotgun-prompt-%s.md", timestamp)
}
