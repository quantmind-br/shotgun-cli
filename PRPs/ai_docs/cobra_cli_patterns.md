# Cobra CLI Framework Patterns for Shotgun-CLI

## Critical URLs and Documentation

### Official Documentation
- **Main Site**: https://cobra.dev
- **GitHub Repository**: https://github.com/spf13/cobra
- **User Guide**: https://github.com/spf13/cobra/blob/main/site/content/user_guide.md
- **Getting Started**: https://cobra.dev/#getting-started
- **API Reference**: https://pkg.go.dev/github.com/spf13/cobra

### Key Version Requirements
- **Cobra**: v1.10.1+ (latest stable)
- **Viper**: v1.19+ (for configuration)

## Root Command with TUI Launch Pattern

### Entry Point Implementation
```go
package main

import (
    "os"
    "github.com/spf13/cobra"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

var (
    verbose    bool
    configFile string
)

var rootCmd = &cobra.Command{
    Use:   "shotgun-cli",
    Short: "Shotgun CLI - Context-aware development tool",
    Long: `Shotgun CLI generates structured text representations of codebases for LLMs.
It features both an interactive TUI wizard and headless CLI modes for automation.`,

    // Launch TUI when no arguments provided
    Run: func(cmd *cobra.Command, args []string) {
        if len(args) == 0 && len(os.Args) == 1 {
            log.Info().Msg("Launching TUI wizard...")
            launchTUIWizard()
            return
        }

        // Show help if args provided but no subcommand
        cmd.Help()
    },
}

func init() {
    // Global persistent flags
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
    rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file (default is $HOME/.shotgun-cli.yaml)")

    // Initialize configuration
    cobra.OnInitialize(initConfig)
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        log.Fatal().Err(err).Msg("Command execution failed")
        os.Exit(1)
    }
}
```

## Subcommand Structure Pattern

### Context Commands
```go
package cmd

import (
    "fmt"
    "path/filepath"
    "github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
    Use:   "context",
    Short: "Context generation tools",
    Long:  "Commands for generating and managing codebase context for LLMs",
}

var contextGenerateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Generate context from codebase",
    Long: `Generate a structured text representation of your codebase within LLM token limits.

Examples:
  shotgun-cli context generate --root . --include "*.go"
  shotgun-cli context generate --exclude "vendor/*" --max-size 5MB`,

    RunE: func(cmd *cobra.Command, args []string) error {
        rootPath, _ := cmd.Flags().GetString("root")
        include, _ := cmd.Flags().GetStringSlice("include")
        exclude, _ := cmd.Flags().GetStringSlice("exclude")
        output, _ := cmd.Flags().GetString("output")
        maxSize, _ := cmd.Flags().GetString("max-size")

        // Validate root path
        if !filepath.IsAbs(rootPath) {
            var err error
            rootPath, err = filepath.Abs(rootPath)
            if err != nil {
                return fmt.Errorf("invalid root path: %w", err)
            }
        }

        config := GenerateConfig{
            RootPath: rootPath,
            Include:  include,
            Exclude:  exclude,
            Output:   output,
            MaxSize:  parseSize(maxSize),
        }

        return generateContextHeadless(config)
    },
}

func init() {
    // Context generate flags
    contextGenerateCmd.Flags().StringP("root", "r", ".", "Root directory to scan")
    contextGenerateCmd.Flags().StringSliceP("include", "i", []string{"*"}, "File patterns to include")
    contextGenerateCmd.Flags().StringSliceP("exclude", "e", []string{}, "File patterns to exclude")
    contextGenerateCmd.Flags().StringP("output", "o", "", "Output file (default: shotgun-prompt-YYYYMMDD-HHMMSS.md)")
    contextGenerateCmd.Flags().String("max-size", "10MB", "Maximum context size (default: 10MB)")

    contextCmd.AddCommand(contextGenerateCmd)
    rootCmd.AddCommand(contextCmd)
}
```

### Template Commands
```go
var templateCmd = &cobra.Command{
    Use:   "template",
    Short: "Template management",
    Long:  "Commands for listing, rendering, and managing prompt templates",
}

var templateListCmd = &cobra.Command{
    Use:   "list",
    Short: "List available templates",
    RunE: func(cmd *cobra.Command, args []string) error {
        templates := listAvailableTemplates()
        for _, tmpl := range templates {
            fmt.Printf("%-20s %s\n", tmpl.Name, tmpl.Description)
        }
        return nil
    },
}

var templateRenderCmd = &cobra.Command{
    Use:   "render [template-name]",
    Short: "Render a template with variables",
    Args:  cobra.ExactArgs(1),
    ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        templates := listAvailableTemplates()
        names := make([]string, len(templates))
        for i, t := range templates {
            names[i] = t.Name
        }
        return names, cobra.ShellCompDirectiveNoFileComp
    },
    RunE: func(cmd *cobra.Command, args []string) error {
        templateName := args[0]
        variables, _ := cmd.Flags().GetStringToString("var")
        output, _ := cmd.Flags().GetString("output")

        return renderTemplate(templateName, variables, output)
    },
}

func init() {
    templateRenderCmd.Flags().StringToStringP("var", "", nil, "Template variables (key=value)")
    templateRenderCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")

    templateCmd.AddCommand(templateListCmd)
    templateCmd.AddCommand(templateRenderCmd)
    rootCmd.AddCommand(templateCmd)
}
```

### Diff Commands
```go
var diffCmd = &cobra.Command{
    Use:   "diff",
    Short: "Diff management tools",
    Long:  "Commands for splitting and managing large diff files",
}

var diffSplitCmd = &cobra.Command{
    Use:   "split",
    Short: "Split large diff into chunks",
    RunE: func(cmd *cobra.Command, args []string) error {
        input, _ := cmd.Flags().GetString("input")
        outputDir, _ := cmd.Flags().GetString("output-dir")
        approxLines, _ := cmd.Flags().GetInt("approx-lines")

        if input == "" {
            return fmt.Errorf("input file is required")
        }

        return splitDiffFile(input, outputDir, approxLines)
    },
}

func init() {
    diffSplitCmd.Flags().StringP("input", "i", "", "Input diff file (required)")
    diffSplitCmd.Flags().StringP("output-dir", "o", "chunks", "Output directory for chunks")
    diffSplitCmd.Flags().Int("approx-lines", 500, "Approximate lines per chunk")
    diffSplitCmd.MarkFlagRequired("input")

    diffCmd.AddCommand(diffSplitCmd)
    rootCmd.AddCommand(diffCmd)
}
```

### Config Commands
```go
var configCmd = &cobra.Command{
    Use:   "config",
    Short: "Configuration management",
    Long:  "Commands for viewing and modifying shotgun-cli configuration",
}

var configShowCmd = &cobra.Command{
    Use:   "show",
    Short: "Show current configuration",
    RunE: func(cmd *cobra.Command, args []string) error {
        config := getCurrentConfig()
        return printConfig(config)
    },
}

var configSetCmd = &cobra.Command{
    Use:   "set [key] [value]",
    Short: "Set configuration value",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        key, value := args[0], args[1]
        return setConfigValue(key, value)
    },
}

func init() {
    configCmd.AddCommand(configShowCmd)
    configCmd.AddCommand(configSetCmd)
    rootCmd.AddCommand(configCmd)
}
```

## Configuration Integration with Viper

```go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "github.com/spf13/viper"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func initConfig() {
    if configFile != "" {
        // Use config file from the flag
        viper.SetConfigFile(configFile)
    } else {
        // Find home directory
        home, err := os.UserHomeDir()
        cobra.CheckErr(err)

        // Search config in home directory with name ".shotgun-cli"
        viper.AddConfigPath(home)
        viper.AddConfigPath(".")
        viper.SetConfigType("yaml")
        viper.SetConfigName(".shotgun-cli")
    }

    // Environment variables
    viper.AutomaticEnv()
    viper.SetEnvPrefix("SHOTGUN")

    // Bind flags to viper
    viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

    // Read config file
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // Config file not found; use defaults
            log.Debug().Msg("No config file found, using defaults")
        } else {
            // Config file was found but another error was produced
            log.Fatal().Err(err).Msg("Error reading config file")
        }
    } else {
        log.Info().Str("config", viper.ConfigFileUsed()).Msg("Using config file")
    }

    // Initialize logging based on config/flags
    initLogging()
}

func initLogging() {
    level := zerolog.InfoLevel
    if viper.GetBool("verbose") {
        level = zerolog.DebugLevel
    }

    zerolog.SetGlobalLevel(level)
    log.Logger = log.Output(zerolog.ConsoleWriter{
        Out:        os.Stderr,
        TimeFormat: "15:04:05",
    })
}
```

## Error Handling Pattern

```go
// Custom error types for better UX
type ValidationError struct {
    Field   string
    Value   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error for %s=%s: %s", e.Field, e.Value, e.Message)
}

// Command with proper error handling
var contextGenerateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Generate context from codebase",
    PreRunE: func(cmd *cobra.Command, args []string) error {
        // Validate flags before execution
        rootPath, _ := cmd.Flags().GetString("root")
        if _, err := os.Stat(rootPath); os.IsNotExist(err) {
            return ValidationError{
                Field:   "root",
                Value:   rootPath,
                Message: "directory does not exist",
            }
        }

        maxSize, _ := cmd.Flags().GetString("max-size")
        if _, err := parseSize(maxSize); err != nil {
            return ValidationError{
                Field:   "max-size",
                Value:   maxSize,
                Message: "invalid size format (use KB, MB, GB)",
            }
        }

        return nil
    },
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation with proper error context
        config, err := buildGenerateConfig(cmd)
        if err != nil {
            return fmt.Errorf("failed to build configuration: %w", err)
        }

        if err := generateContextHeadless(config); err != nil {
            return fmt.Errorf("context generation failed: %w", err)
        }

        log.Info().Msg("Context generated successfully")
        return nil
    },
}
```

## Shell Completion Pattern

```go
func init() {
    // Add completion command
    completionCmd := &cobra.Command{
        Use:   "completion [bash|zsh|fish|powershell]",
        Short: "Generate completion script",
        Long: `Generate shell completion scripts for shotgun-cli.

To configure bash completion:
  shotgun-cli completion bash > /etc/bash_completion.d/shotgun-cli

To configure zsh completion:
  shotgun-cli completion zsh > "${fpath[1]}/_shotgun-cli"`,
        DisableFlagsInUseLine: true,
        ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
        Args:                  cobra.ExactValidArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            switch args[0] {
            case "bash":
                cmd.Root().GenBashCompletion(os.Stdout)
            case "zsh":
                cmd.Root().GenZshCompletion(os.Stdout)
            case "fish":
                cmd.Root().GenFishCompletion(os.Stdout, true)
            case "powershell":
                cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
            }
        },
    }

    rootCmd.AddCommand(completionCmd)
}

// Custom completion for template names
func templateCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    templates := listAvailableTemplates()
    completions := make([]string, 0, len(templates))

    for _, template := range templates {
        if strings.HasPrefix(template.Name, toComplete) {
            completions = append(completions, fmt.Sprintf("%s\t%s", template.Name, template.Description))
        }
    }

    return completions, cobra.ShellCompDirectiveNoFileComp
}
```

## Platform-Specific Handling

```go
// Cross-platform path validation
func validatePath(path string) error {
    if runtime.GOOS == "windows" {
        // Windows path validation
        if matched, _ := regexp.MatchString(`^[a-zA-Z]:\\`, path); !matched && filepath.IsAbs(path) {
            return fmt.Errorf("invalid Windows path format: %s", path)
        }
    }

    // Check path exists and is accessible
    if _, err := os.Stat(path); err != nil {
        return fmt.Errorf("path not accessible: %w", err)
    }

    return nil
}

// Platform-specific default configuration
func getDefaultConfigPath() string {
    switch runtime.GOOS {
    case "windows":
        return filepath.Join(os.Getenv("APPDATA"), "shotgun-cli", "config.yaml")
    case "darwin":
        home, _ := os.UserHomeDir()
        return filepath.Join(home, "Library", "Application Support", "shotgun-cli", "config.yaml")
    default: // Linux and others
        if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
            return filepath.Join(xdgConfig, "shotgun-cli", "config.yaml")
        }
        home, _ := os.UserHomeDir()
        return filepath.Join(home, ".config", "shotgun-cli", "config.yaml")
    }
}
```

## Performance Optimization Pattern

```go
// Lazy initialization for heavy dependencies
var (
    contextGenerator *ContextGenerator
    templateManager  *TemplateManager
    once            sync.Once
)

func getContextGenerator() *ContextGenerator {
    once.Do(func() {
        contextGenerator = NewContextGenerator()
    })
    return contextGenerator
}

// Graceful interruption handling
func withCancellation(fn func(context.Context) error) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle interrupt signals
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-c
        log.Info().Msg("Received interrupt signal, cancelling operation...")
        cancel()
    }()

    return fn(ctx)
}
```

## Key Implementation Gotchas

### Flag Binding with Viper
- Call `viper.BindPFlag()` after adding flags but before `cobra.OnInitialize()`
- Use persistent flags for global options like `--verbose`

### Command Structure
- Use `RunE` instead of `Run` for better error handling
- Implement `PreRunE` for flag validation
- Use `Args: cobra.ExactArgs(n)` for argument validation

### Error Messages
- Provide actionable error messages
- Use structured errors for different error types
- Include examples in command help text

### Configuration Management
- Support multiple config file locations
- Use environment variables with prefix
- Provide sane defaults for all configuration options

This documentation provides all the essential patterns for implementing a robust CLI application with Cobra that supports both interactive TUI mode and headless automation commands.