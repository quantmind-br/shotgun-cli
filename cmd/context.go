package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/app"
	cfgkeys "github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/core/tokens"
	"github.com/quantmind-br/shotgun-cli/internal/platform/gemini"
	"github.com/quantmind-br/shotgun-cli/internal/utils"
)

// ProgressMode defines how progress is reported.
type ProgressMode string

const (
	ProgressNone  ProgressMode = "none"
	ProgressHuman ProgressMode = "human"
	ProgressJSON  ProgressMode = "json"
)

// ProgressOutput represents a progress event for output
type ProgressOutput struct {
	Timestamp string  `json:"timestamp"`
	Stage     string  `json:"stage"`
	Message   string  `json:"message"`
	Current   int64   `json:"current,omitempty"`
	Total     int64   `json:"total,omitempty"`
	Percent   float64 `json:"percent,omitempty"`
}

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
	// Template configuration
	Template   string            // Template name to use
	Task       string            // Task description for LLM
	Rules      string            // Rules/constraints for LLM
	CustomVars map[string]string // Custom template variables (KEY=VALUE)
	// Scanner overrides (0/false = use config)
	Workers        int
	IncludeHidden  bool
	IncludeIgnored bool
	// Progress output
	ProgressMode ProgressMode
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

	// Template flags
	templateName, _ := cmd.Flags().GetString("template")
	task, _ := cmd.Flags().GetString("task")
	rules, _ := cmd.Flags().GetString("rules")
	varFlags, _ := cmd.Flags().GetStringArray("var")

	// Scanner override flags
	workers, _ := cmd.Flags().GetInt("workers")
	includeHidden, _ := cmd.Flags().GetBool("include-hidden")
	includeIgnored, _ := cmd.Flags().GetBool("include-ignored")

	// Progress flag
	progressStr, _ := cmd.Flags().GetString("progress")
	var progressMode ProgressMode
	switch progressStr {
	case "none", "":
		progressMode = ProgressNone
	case "human":
		progressMode = ProgressHuman
	case "json":
		progressMode = ProgressJSON
	default:
		return GenerateConfig{}, fmt.Errorf("invalid --progress value: %q (expected: none, human, json)", progressStr)
	}

	// Parse custom variables
	customVars := make(map[string]string)
	for _, v := range varFlags {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return GenerateConfig{}, fmt.Errorf("invalid --var format: %q (expected KEY=VALUE)", v)
		}
		customVars[parts[0]] = parts[1]
	}

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
		geminiModel = viper.GetString(cfgkeys.KeyGeminiModel)
	}
	if geminiTimeout == 0 {
		geminiTimeout = viper.GetInt(cfgkeys.KeyGeminiTimeout)
	}

	// Enable gemini via config if auto-send is enabled AND gemini is enabled
	if !sendGemini && viper.GetBool(cfgkeys.KeyGeminiAutoSend) && viper.GetBool(cfgkeys.KeyGeminiEnabled) {
		sendGemini = true
	}

	return GenerateConfig{
		RootPath:       absPath,
		Include:        include,
		Exclude:        exclude,
		Output:         output,
		MaxSize:        maxSize,
		EnforceLimit:   enforceLimit,
		SendGemini:     sendGemini,
		GeminiModel:    geminiModel,
		GeminiOutput:   geminiOutput,
		GeminiTimeout:  geminiTimeout,
		Template:       templateName,
		Task:           task,
		Rules:          rules,
		CustomVars:     customVars,
		Workers:        workers,
		IncludeHidden:  includeHidden,
		IncludeIgnored: includeIgnored,
		ProgressMode:   progressMode,
	}, nil
}

func generateContextHeadless(cfg GenerateConfig) error {
	scannerConfig := buildScannerConfig(cfg)
	log.Debug().Interface("config", scannerConfig).Msg("Scanner configuration")

	templateVars := buildTemplateVars(cfg)
	templateContent, err := loadTemplateContent(cfg.Template)
	if err != nil {
		return err
	}

	svc := app.NewContextService()
	svcCfg := app.GenerateConfig{
		RootPath:        cfg.RootPath,
		ScanConfig:      &scannerConfig,
		Template:        templateContent,
		TemplateVars:    templateVars,
		MaxSize:         cfg.MaxSize,
		EnforceLimit:    cfg.EnforceLimit,
		OutputPath:      cfg.Output,
		CopyToClipboard: viper.GetBool(cfgkeys.KeyOutputClipboard),
		IncludeTree:     viper.GetBool(cfgkeys.KeyContextIncludeTree),
		IncludeSummary:  viper.GetBool(cfgkeys.KeyContextIncludeSummary),
		SkipBinary:      viper.GetBool(cfgkeys.KeyScannerSkipBinary),
	}

	var result *app.GenerateResult
	ctx := context.Background()

	if cfg.ProgressMode != ProgressNone {
		result, err = svc.GenerateWithProgress(ctx, svcCfg, func(stage, msg string, cur, total int64) {
			var percent float64
			if total > 0 {
				percent = float64(cur) / float64(total) * 100
			}
			renderProgress(cfg.ProgressMode, ProgressOutput{
				Timestamp: time.Now().Format(time.RFC3339),
				Stage:     stage,
				Message:   msg,
				Current:   cur,
				Total:     total,
				Percent:   percent,
			})
		})
		clearProgressLine(cfg.ProgressMode)
	} else {
		result, err = svc.Generate(ctx, svcCfg)
	}

	if err != nil {
		return fmt.Errorf("context generation failed: %w", err)
	}

	log.Info().Int("files", result.FileCount).Msg("Files scanned")
	if result.CopiedToClipboard {
		log.Info().Msg("Context copied to clipboard")
	}

	printGenerationSummary(result, cfg)

	if cfg.SendGemini {
		if err := sendToGemini(cfg, result.Content); err != nil {
			log.Error().Err(err).Msg("Failed to send to Gemini")
			fmt.Printf("❌ Gemini: %v\n", err)
		}
	}

	return nil
}

func buildScannerConfig(cfg GenerateConfig) scanner.ScanConfig {
	scannerConfig := scanner.ScanConfig{
		MaxFiles:             viper.GetInt64(cfgkeys.KeyScannerMaxFiles),
		MaxFileSize:          utils.ParseSizeWithDefault(viper.GetString(cfgkeys.KeyScannerMaxFileSize), 1024*1024),
		MaxMemory:            utils.ParseSizeWithDefault(viper.GetString(cfgkeys.KeyScannerMaxMemory), 500*1024*1024),
		SkipBinary:           viper.GetBool(cfgkeys.KeyScannerSkipBinary),
		IncludeHidden:        viper.GetBool(cfgkeys.KeyScannerIncludeHidden),
		IncludeIgnored:       viper.GetBool(cfgkeys.KeyScannerIncludeIgnored),
		Workers:              viper.GetInt(cfgkeys.KeyScannerWorkers),
		RespectGitignore:     viper.GetBool(cfgkeys.KeyScannerRespectGitignore),
		RespectShotgunignore: viper.GetBool(cfgkeys.KeyScannerRespectShotgunignore),
		IgnorePatterns:       cfg.Exclude,
		IncludePatterns:      cfg.Include,
	}

	if cfg.Workers > 0 {
		scannerConfig.Workers = cfg.Workers
	}
	if cfg.IncludeHidden {
		scannerConfig.IncludeHidden = true
	}
	if cfg.IncludeIgnored {
		scannerConfig.IncludeIgnored = true
	}

	return scannerConfig
}

func buildTemplateVars(cfg GenerateConfig) map[string]string {
	taskValue := cfg.Task
	if taskValue == "" {
		taskValue = "Context generation"
	}

	templateVars := map[string]string{
		"TASK":           taskValue,
		"RULES":          cfg.Rules,
		"FILE_STRUCTURE": "",
		"CURRENT_DATE":   time.Now().Format("2006-01-02"),
	}

	for k, v := range cfg.CustomVars {
		templateVars[k] = v
	}

	return templateVars
}

func loadTemplateContent(templateName string) (string, error) {
	if templateName == "" {
		return "", nil
	}

	tmplMgr, err := template.NewManager(template.ManagerConfig{
		CustomPath: viper.GetString(cfgkeys.KeyTemplateCustomPath),
	})
	if err != nil {
		return "", fmt.Errorf("failed to initialize template manager: %w", err)
	}

	tmpl, err := tmplMgr.GetTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to load template %q: %w", templateName, err)
	}

	log.Debug().Str("template", templateName).Msg("Using custom template")
	return tmpl.Content, nil
}

func printGenerationSummary(result *app.GenerateResult, cfg GenerateConfig) {
	fmt.Printf("✅ Context generated successfully!\n")
	fmt.Printf("📁 Root path: %s\n", cfg.RootPath)
	fmt.Printf("📄 Output file: %s\n", result.OutputPath)
	fmt.Printf("📊 Files processed: %d\n", result.FileCount)
	fmt.Printf("📏 Total size: %s (~%s tokens)\n",
		utils.FormatBytes(result.ContentSize),
		tokens.FormatTokens(int(result.TokenEstimate)))
	fmt.Printf("🎯 Size limit: %s\n", utils.FormatBytes(cfg.MaxSize))
}

func sendToGemini(cfg GenerateConfig, content string) error {
	// Check if Gemini is enabled
	if !viper.GetBool(cfgkeys.KeyGeminiEnabled) {
		return fmt.Errorf("gemini integration is disabled, enable with: shotgun-cli config set gemini.enabled true")
	}

	// Check availability
	if !gemini.IsAvailable() {
		return fmt.Errorf("geminiweb not found. Run 'shotgun-cli gemini doctor' for help")
	}

	if !gemini.IsConfigured() {
		return fmt.Errorf("geminiweb not configured. Run 'shotgun-cli gemini doctor' for help")
	}

	// Build gemini config
	geminiCfg := gemini.Config{
		BinaryPath:     viper.GetString(cfgkeys.KeyGeminiBinaryPath),
		Model:          cfg.GeminiModel,
		Timeout:        cfg.GeminiTimeout,
		BrowserRefresh: viper.GetString(cfgkeys.KeyGeminiBrowserRefresh),
		Verbose:        viper.GetBool(cfgkeys.KeyVerbose),
	}

	executor := gemini.NewExecutor(geminiCfg)

	fmt.Printf("\n🤖 Sending to Gemini (%s)...\n", geminiCfg.Model)

	ctx := context.Background()
	result, err := executor.Send(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to send to Gemini: %w", err)
	}

	// Determine output file
	geminiOutput := cfg.GeminiOutput
	if geminiOutput == "" {
		geminiOutput = strings.TrimSuffix(cfg.Output, ".md") + "_response.md"
	}

	// Save response
	if viper.GetBool(cfgkeys.KeyGeminiSaveResponse) {
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

// renderProgressHuman renders progress for humans
func renderProgressHuman(p ProgressOutput) {
	if p.Total > 0 {
		fmt.Fprintf(os.Stderr, "\r[%s] %s: %d/%d (%.1f%%)  ",
			p.Stage, p.Message, p.Current, p.Total, p.Percent)
	} else {
		fmt.Fprintf(os.Stderr, "\r[%s] %s  ", p.Stage, p.Message)
	}
}

// renderProgressJSON renders progress as JSON (one line per event)
func renderProgressJSON(p ProgressOutput) {
	data, _ := json.Marshal(p)
	fmt.Fprintln(os.Stderr, string(data))
}

// renderProgress renders progress in the specified mode
func renderProgress(mode ProgressMode, p ProgressOutput) {
	switch mode {
	case ProgressHuman:
		renderProgressHuman(p)
	case ProgressJSON:
		renderProgressJSON(p)
	case ProgressNone:
		// Silent
	}
}

// clearProgressLine clears the progress line (for human mode)
func clearProgressLine(mode ProgressMode) {
	if mode == ProgressHuman {
		fmt.Fprint(os.Stderr, "\r\033[K")
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
	contextGenerateCmd.Flags().String("gemini-model", "", "Gemini model (gemini-2.5-flash, gemini-2.5-pro)")
	contextGenerateCmd.Flags().String("gemini-output", "", "File to save Gemini response")
	contextGenerateCmd.Flags().Int("gemini-timeout", 0, "Timeout in seconds for Gemini request")

	// Template configuration flags
	contextGenerateCmd.Flags().StringP("template", "t", "", "Template name (e.g., makePlan, analyzeBug)")
	contextGenerateCmd.Flags().String("task", "", "Task description for the LLM")
	contextGenerateCmd.Flags().String("rules", "", "Rules/constraints for the LLM")
	contextGenerateCmd.Flags().StringArrayP("var", "V", []string{}, "Custom template vars KEY=VALUE (repeatable)")

	// Scanner override flags
	contextGenerateCmd.Flags().Int("workers", 0, "Number of parallel workers (0 = use config)")
	contextGenerateCmd.Flags().Bool("include-hidden", false, "Include hidden files")
	contextGenerateCmd.Flags().Bool("include-ignored", false, "Include ignored files")

	// Progress output flag
	contextGenerateCmd.Flags().String("progress", "none", "Progress output mode: none, human, json")

	// Mark root as required would be too restrictive since we have a default
	// But we validate it in PreRunE instead

	contextCmd.AddCommand(contextGenerateCmd)
	rootCmd.AddCommand(contextCmd)
}
