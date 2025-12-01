package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ctxgen "github.com/quantmind-br/shotgun-cli/internal/core/context"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/tokens"
	"github.com/quantmind-br/shotgun-cli/internal/platform/clipboard"
	"github.com/quantmind-br/shotgun-cli/internal/platform/gemini"
	"github.com/quantmind-br/shotgun-cli/internal/utils"
)

type GenerateConfig struct {
	RootPath      string
	Include       []string
	Exclude       []string
	Output        string
	MaxSize       int64
	EnforceLimit  bool
	SendGemini    bool
	GeminiModel   string
	GeminiOutput  string
	GeminiTimeout int
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
		if _, err := utils.ParseSize(maxSizeStr); err != nil {
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

	// Gemini flags
	sendGemini, _ := cmd.Flags().GetBool("send-gemini")
	geminiModel, _ := cmd.Flags().GetString("gemini-model")
	geminiOutput, _ := cmd.Flags().GetString("gemini-output")
	geminiTimeout, _ := cmd.Flags().GetInt("gemini-timeout")

	// Convert root to absolute path
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return GenerateConfig{}, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Parse max size
	maxSize, err := utils.ParseSize(maxSizeStr)
	if err != nil {
		return GenerateConfig{}, fmt.Errorf("failed to parse max-size: %w", err)
	}

	// Generate default output filename if not specified
	if output == "" {
		timestamp := time.Now().Format("20060102-150405")
		output = fmt.Sprintf("shotgun-prompt-%s.md", timestamp)
	}

	// Use config defaults for Gemini if not specified via flags
	if geminiModel == "" {
		geminiModel = viper.GetString("gemini.model")
	}
	if geminiTimeout == 0 {
		geminiTimeout = viper.GetInt("gemini.timeout")
	}

	// Enable gemini via config if auto-send is enabled
	if !sendGemini && viper.GetBool("gemini.auto-send") {
		sendGemini = true
	}

	return GenerateConfig{
		RootPath:      absPath,
		Include:       include,
		Exclude:       exclude,
		Output:        output,
		MaxSize:       maxSize,
		EnforceLimit:  enforceLimit,
		SendGemini:    sendGemini,
		GeminiModel:   geminiModel,
		GeminiOutput:  geminiOutput,
		GeminiTimeout: geminiTimeout,
	}, nil
}

func generateContextHeadless(config GenerateConfig) error {
	// Initialize scanner with configuration
	scannerConfig := scanner.ScanConfig{
		MaxFiles:        viper.GetInt64("scanner.max-files"),
		MaxFileSize:     utils.ParseSizeWithDefault(viper.GetString("scanner.max-file-size"), 1024*1024),
		SkipBinary:      viper.GetBool("scanner.skip-binary"),
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

	// Create selection map for all valid files
	selections := make(map[string]bool)
	collectAllSelections(tree, selections)

	// Extract file list from tree
	fileCount := countFilesInTree(tree)
	log.Info().Int("files", fileCount).Msg("Files scanned")

	// Initialize context generator
	generator := ctxgen.NewDefaultContextGenerator()

	contextConfig := ctxgen.GenerateConfig{
		MaxTotalSize: config.MaxSize,
		TemplateVars: map[string]string{
			"TASK":           "Context generation",
			"RULES":          "",
			"FILE_STRUCTURE": "",
			"CURRENT_DATE":   time.Now().Format("2006-01-02"),
		},
		SkipBinary: viper.GetBool("scanner.skip-binary"),
	}

	// Generate context
	content, err := generator.Generate(tree, selections, contextConfig)
	if err != nil {
		return fmt.Errorf("failed to generate context: %w", err)
	}

	// Check size limits
	contentSize := int64(len(content))
	if contentSize > config.MaxSize {
		if config.EnforceLimit {
			return fmt.Errorf(
				"generated context size (%s) exceeds limit (%s). "+
					"Use --no-enforce-limit to allow truncation or generation without enforcement",
				utils.FormatBytes(contentSize), utils.FormatBytes(config.MaxSize))
		} else {
			log.Warn().
				Int64("actual", contentSize).
				Int64("limit", config.MaxSize).
				Msg("Generated context exceeds size limit - continuing without enforcement")
		}
	}

	// Write output file
	if err := os.WriteFile(config.Output, []byte(content), 0600); err != nil {
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

	// Print summary with token estimate
	contentBytes := int64(len(content))
	estimatedTokens := tokens.EstimateFromBytes(contentBytes)
	fmt.Printf("✅ Context generated successfully!\n")
	fmt.Printf("📁 Root path: %s\n", config.RootPath)
	fmt.Printf("📄 Output file: %s\n", config.Output)
	fmt.Printf("📊 Files processed: %d\n", fileCount)
	fmt.Printf("📏 Total size: %s (~%s tokens)\n", utils.FormatBytes(contentBytes), tokens.FormatTokens(estimatedTokens))
	fmt.Printf("🎯 Size limit: %s\n", utils.FormatBytes(config.MaxSize))

	// Send to Gemini if requested
	if config.SendGemini {
		if err := sendToGemini(config, content); err != nil {
			log.Error().Err(err).Msg("Failed to send to Gemini")
			fmt.Printf("❌ Gemini: %v\n", err)
		}
	}

	return nil
}

func sendToGemini(config GenerateConfig, content string) error {
	// Check availability
	if !gemini.IsAvailable() {
		return fmt.Errorf("geminiweb not found. Install with: go install github.com/diogo/geminiweb/cmd/geminiweb@latest")
	}

	if !gemini.IsConfigured() {
		return fmt.Errorf("geminiweb not configured. Run: geminiweb auto-login")
	}

	// Build gemini config
	geminiCfg := gemini.Config{
		BinaryPath:     viper.GetString("gemini.binary-path"),
		Model:          config.GeminiModel,
		Timeout:        config.GeminiTimeout,
		BrowserRefresh: viper.GetString("gemini.browser-refresh"),
		Verbose:        viper.GetBool("verbose"),
	}

	executor := gemini.NewExecutor(geminiCfg)

	fmt.Printf("\n🤖 Sending to Gemini (%s)...\n", geminiCfg.Model)

	ctx := context.Background()
	result, err := executor.Send(ctx, content)
	if err != nil {
		return err
	}

	// Determine output file
	geminiOutput := config.GeminiOutput
	if geminiOutput == "" {
		geminiOutput = strings.TrimSuffix(config.Output, ".md") + "_response.md"
	}

	// Save response
	if viper.GetBool("gemini.save-response") {
		if err := os.WriteFile(geminiOutput, []byte(result.Response), 0600); err != nil {
			return fmt.Errorf("failed to save response: %w", err)
		}
		fmt.Printf("✅ Gemini response saved to: %s\n", geminiOutput)
	} else {
		fmt.Printf("\n--- Gemini Response ---\n%s\n", result.Response)
	}

	fmt.Printf("⏱️  Response time: %s\n", gemini.FormatDuration(result.Duration))

	return nil
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

// collectAllSelections recursively collects all non-ignored files into the selections map
func collectAllSelections(node *scanner.FileNode, selections map[string]bool) {
	if node == nil {
		return
	}

	// Mark file as selected if it's not ignored
	if !node.IsIgnored() {
		selections[node.Path] = true
	}

	// Recursively select children
	if node.IsDir {
		for _, child := range node.Children {
			collectAllSelections(child, selections)
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

	// Gemini integration flags
	contextGenerateCmd.Flags().Bool("send-gemini", false, "Send generated context to Gemini")
	contextGenerateCmd.Flags().String("gemini-model", "", "Gemini model to use (gemini-2.5-flash, gemini-2.5-pro, gemini-3.0-pro)")
	contextGenerateCmd.Flags().String("gemini-output", "", "File to save Gemini response (default: <output>_response.md)")
	contextGenerateCmd.Flags().Int("gemini-timeout", 0, "Timeout in seconds for Gemini request (default: from config)")

	// Mark root as required would be too restrictive since we have a default
	// But we validate it in PreRunE instead

	contextCmd.AddCommand(contextGenerateCmd)
	rootCmd.AddCommand(contextCmd)
}
