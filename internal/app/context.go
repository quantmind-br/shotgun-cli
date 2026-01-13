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

type ContextService interface {
	Generate(ctx context.Context, cfg GenerateConfig) (*GenerateResult, error)
	GenerateWithProgress(
		ctx context.Context, cfg GenerateConfig, progress ProgressCallback,
	) (*GenerateResult, error)
	SendToLLM(ctx context.Context, content string, provider llm.Provider) (*llm.Result, error)
	SendToLLMWithProgress(
		ctx context.Context, content string, cfg LLMSendConfig, progress LLMProgressCallback,
	) (*llm.Result, error)
}

type LLMSendConfig struct {
	Provider       llm.ProviderType
	APIKey         string
	BaseURL        string
	Model          string
	Timeout        int
	SaveResponse   bool
	OutputPath     string
	BinaryPath     string // GeminiWeb legacy
	BrowserRefresh string // GeminiWeb legacy
}

type LLMProgressCallback func(stage string)

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

type GenerateResult struct {
	Content           string
	OutputPath        string
	FileCount         int
	ContentSize       int64
	TokenEstimate     int64
	CopiedToClipboard bool
}

type ProgressCallback func(stage string, message string, current, total int64)

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

func (c *GenerateConfig) GenerateOutputPath() string {
	if c.OutputPath != "" {
		return c.OutputPath
	}
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("shotgun-prompt-%s.md", timestamp)
}
