package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/core/context"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/platform/clipboard"
)

type GenerateConfig struct {
	RootPath     string
	Include      []string
	Exclude      []string
	Output       string
	MaxSize      int64
	EnforceLimit bool
}

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Context generation tools",
	Long:  "Commands for generating and managing codebase context for LLMs",
}

var contextGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate context from codebase",
	Long: `Generate a structured text representation of your codebase within LLM token limits.

This command scans your codebase, applies ignore patterns, and generates an optimized
context file suitable for LLM consumption. The output includes a directory tree,
file summaries, and file contents within the specified size limits.

By default, the context generation will fail if the output exceeds the max-size limit.
Use --no-enforce-limit to allow generation that exceeds the limit with a warning.

Examples:
  shotgun-cli context generate --root . --include "*.go"
  shotgun-cli context generate --exclude "vendor/*,*.test.go" --max-size 5MB
  shotgun-cli context generate --output my-context.md --root ./src
  shotgun-cli context generate --include "*.py,*.js" --exclude "node_modules/*"
  shotgun-cli context generate --no-enforce-limit --max-size 5MB`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate root path
		rootPath, _ := cmd.Flags().GetString("root")
		if rootPath == "" {
			return fmt.Errorf("root path cannot be empty")
		}

		// Convert to absolute path
		absPath, err := filepath.Abs(rootPath)
		if err != nil {
			return fmt.Errorf("invalid root path '%s': %w", rootPath, err)
		}

		// Check if path exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("root path does not exist: %s", absPath)
		}

		// Check if path is a directory
		if info, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("cannot access root path '%s': %w", absPath, err)
		} else if !info.IsDir() {
			return fmt.Errorf("root path must be a directory: %s", absPath)
		}

		// Validate max-size format
		maxSizeStr, _ := cmd.Flags().GetString("max-size")
		if _, err := parseSize(maxSizeStr); err != nil {
			return fmt.Errorf("invalid max-size format '%s': %w (use formats like 1MB, 5GB, 500KB)", maxSizeStr, err)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build configuration from flags
		config, err := buildGenerateConfig(cmd)
		if err != nil {
			return fmt.Errorf("failed to build configuration: %w", err)
		}

		// Generate context
		log.Info().Str("root", config.RootPath).Msg("Starting context generation...")

		if err := generateContextHeadless(config); err != nil {
			return fmt.Errorf("context generation failed: %w", err)
		}

		log.Info().Msg("Context generated successfully")
		return nil
	},
}

func buildGenerateConfig(cmd *cobra.Command) (GenerateConfig, error) {
	rootPath, _ := cmd.Flags().GetString("root")
	include, _ := cmd.Flags().GetStringSlice("include")
	exclude, _ := cmd.Flags().GetStringSlice("exclude")
	output, _ := cmd.Flags().GetString("output")
	maxSizeStr, _ := cmd.Flags().GetString("max-size")
	enforceLimit, _ := cmd.Flags().GetBool("enforce-limit")

	// Convert root to absolute path
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return GenerateConfig{}, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Parse max size
	maxSize, err := parseSize(maxSizeStr)
	if err != nil {
		return GenerateConfig{}, fmt.Errorf("failed to parse max-size: %w", err)
	}

	// Generate default output filename if not specified
	if output == "" {
		timestamp := time.Now().Format("20060102-150405")
		output = fmt.Sprintf("shotgun-prompt-%s.md", timestamp)
	}

	return GenerateConfig{
		RootPath:     absPath,
		Include:      include,
		Exclude:      exclude,
		Output:       output,
		MaxSize:      maxSize,
		EnforceLimit: enforceLimit,
	}, nil
}

func generateContextHeadless(config GenerateConfig) error {
	// Initialize scanner with configuration
	scannerConfig := scanner.ScanConfig{
		MaxFiles:        viper.GetInt64("scanner.max-files"),
		MaxFileSize:     parseConfigSize(viper.GetString("scanner.max-file-size")),
		SkipBinary:      false,
		IncludeHidden:   false,
		Workers:         1,
		IgnorePatterns:  config.Exclude,
		IncludePatterns: config.Include,
	}

	log.Debug().Interface("config", scannerConfig).Msg("Scanner configuration")

	// Create and run scanner
	fs := scanner.NewFileSystemScanner()
	tree, err := fs.Scan(config.RootPath, &scannerConfig)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	// Mark all non-ignored files as selected for content generation
	selectAllFiles(tree)

	// Extract file list from tree
	fileCount := countFilesInTree(tree)
	log.Info().Int("files", fileCount).Msg("Files scanned")

	// Initialize context generator
	generator := context.NewDefaultContextGenerator()

	contextConfig := context.GenerateConfig{
		MaxTotalSize: config.MaxSize,
		TemplateVars: map[string]string{
			"TASK":          "Context generation",
			"RULES":         "",
			"FILE_STRUCTURE": "",
			"CURRENT_DATE":   time.Now().Format("2006-01-02"),
		},
	}

	// Generate context
	content, err := generator.Generate(tree, contextConfig)
	if err != nil {
		return fmt.Errorf("failed to generate context: %w", err)
	}

	// Check size limits
	contentSize := int64(len(content))
	if contentSize > config.MaxSize {
		if config.EnforceLimit {
			return fmt.Errorf("generated context size (%s) exceeds limit (%s). Use --no-enforce-limit to allow truncation or generation without enforcement",
				formatBytes(contentSize), formatBytes(config.MaxSize))
		} else {
			log.Warn().
				Int64("actual", contentSize).
				Int64("limit", config.MaxSize).
				Msg("Generated context exceeds size limit - continuing without enforcement")
		}
	}

	// Write output file
	if err := os.WriteFile(config.Output, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write output file '%s': %w", config.Output, err)
	}

	// Copy to clipboard if enabled
	if viper.GetBool("output.clipboard") {
		if err := clipboard.Copy(content); err != nil {
			log.Warn().Err(err).Msg("Failed to copy to clipboard")
		} else {
			log.Info().Msg("Context copied to clipboard")
		}
	}

	// Print summary
	fmt.Printf("✅ Context generated successfully!\n")
	fmt.Printf("📁 Root path: %s\n", config.RootPath)
	fmt.Printf("📄 Output file: %s\n", config.Output)
	fmt.Printf("📊 Files processed: %d\n", fileCount)
	fmt.Printf("📏 Total size: %s\n", formatBytes(int64(len(content))))
	fmt.Printf("🎯 Size limit: %s\n", formatBytes(config.MaxSize))

	return nil
}

func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Handle numeric-only input (assume bytes)
	if val, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return val, nil
	}

	// Parse with units
	var multiplier int64 = 1
	var numStr string

	switch {
	case strings.HasSuffix(sizeStr, "GB"):
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "GB")
	case strings.HasSuffix(sizeStr, "MB"):
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "MB")
	case strings.HasSuffix(sizeStr, "KB"):
		multiplier = 1024
		numStr = strings.TrimSuffix(sizeStr, "KB")
	case strings.HasSuffix(sizeStr, "B"):
		multiplier = 1
		numStr = strings.TrimSuffix(sizeStr, "B")
	default:
		return 0, fmt.Errorf("invalid size format, use KB, MB, GB, or B")
	}

	val, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %w", err)
	}

	return int64(val * float64(multiplier)), nil
}

func parseConfigSize(sizeStr string) int64 {
	size, err := parseSize(sizeStr)
	if err != nil {
		log.Debug().Err(err).Str("size", sizeStr).Msg("Failed to parse config size, using default")
		return 1024 * 1024 // 1MB default
	}
	return size
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func countFilesInTree(node *scanner.FileNode) int {
	count := 0
	if !node.IsDir {
		return 1
	}
	for _, child := range node.Children {
		count += countFilesInTree(child)
	}
	return count
}

// selectAllFiles recursively marks all non-ignored files as selected
func selectAllFiles(node *scanner.FileNode) {
	if node == nil {
		return
	}

	// Mark node as selected if it's not ignored
	if !node.IsIgnored() {
		node.Selected = true
	}

	// Recursively select children
	if node.IsDir {
		for _, child := range node.Children {
			selectAllFiles(child)
		}
	}
}

func init() {
	// Context generate flags
	contextGenerateCmd.Flags().StringP("root", "r", ".", "Root directory to scan")
	contextGenerateCmd.Flags().StringSliceP("include", "i", []string{"*"}, "File patterns to include (glob patterns)")
	contextGenerateCmd.Flags().StringSliceP("exclude", "e", []string{}, "File patterns to exclude (glob patterns)")
	contextGenerateCmd.Flags().StringP("output", "o", "", "Output file (default: shotgun-prompt-YYYYMMDD-HHMMSS.md)")
	contextGenerateCmd.Flags().String("max-size", "10MB", "Maximum context size (e.g., 5MB, 1GB, 500KB)")
	contextGenerateCmd.Flags().Bool("enforce-limit", true, "Enforce context size limit (default: true)")

	// Mark root as required would be too restrictive since we have a default
	// But we validate it in PreRunE instead

	contextCmd.AddCommand(contextGenerateCmd)
	rootCmd.AddCommand(contextCmd)
}