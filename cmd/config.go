package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/ui"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long: `Commands for viewing and modifying shotgun-cli configuration.

When called without subcommands, launches an interactive TUI for managing
all configuration settings organized by category.

Subcommands:
  show    Display current configuration values
  set     Set a specific configuration value

Examples:
  # Launch interactive configuration TUI
  shotgun-cli config

  # Show current configuration
  shotgun-cli config show

  # Set a configuration value
  shotgun-cli config set llm.provider openai`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchConfigTUI()
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Display the current configuration values with their sources.

Shows all configuration values including defaults, values from config files,
environment variables, and command-line flags. The source of each value
is indicated to help understand the configuration precedence.

Example:
  shotgun-cli config show`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return showCurrentConfig()
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration value",
	Long: `Set a configuration value in the config file.

Updates the configuration file with the new value. If no config file exists,
it will be created in the appropriate location for your platform.

Supported configuration keys:

  LLM Provider:
    llm.provider              - LLM provider: openai, anthropic, gemini, geminiweb (default: "geminiweb")
    llm.api-key               - API key for the provider (required for openai, anthropic, gemini)
    llm.base-url              - Custom API endpoint URL (for OpenRouter, Azure, etc.)
    llm.model                 - Model to use (e.g., gpt-4o, claude-sonnet-4-20250514, gemini-2.5-flash)
    llm.timeout               - Request timeout in seconds (default: 300)

  Scanner:
    scanner.max-files         - Maximum number of files to scan (default: 10000)
    scanner.max-file-size     - Maximum size per file (default: "1MB")
    scanner.max-memory        - Maximum memory usage (default: "500MB")
    scanner.respect-gitignore - Respect .gitignore files (default: true)
    scanner.skip-binary       - Skip binary files (default: true)
    scanner.workers           - Number of parallel workers (default: 1)
    scanner.include-hidden    - Include hidden files (default: false)
    scanner.respect-shotgunignore - Respect .shotgunignore files (default: true)

  Context:
    context.max-size          - Maximum context size (default: "10MB")
    context.include-tree      - Include directory tree (default: true)
    context.include-summary   - Include file summaries (default: true)

  Template:
    template.custom-path      - Path to custom templates (default: "")

  Output:
    output.format             - Output format: markdown, text (default: "markdown")
    output.clipboard          - Copy to clipboard (default: true)

  GeminiWeb (legacy):
    gemini.enabled            - Enable GeminiWeb integration (default: false)
    gemini.binary-path        - Path to geminiweb binary
    gemini.model              - Model for GeminiWeb (default: "gemini-2.5-flash")
    gemini.timeout            - Timeout in seconds (default: 300)
    gemini.browser-refresh    - Browser refresh method: auto, chrome, firefox, edge

Examples:
  # Configure OpenAI
  shotgun-cli config set llm.provider openai
  shotgun-cli config set llm.api-key sk-your-api-key
  shotgun-cli config set llm.model gpt-4o

  # Configure Anthropic Claude
  shotgun-cli config set llm.provider anthropic
  shotgun-cli config set llm.api-key sk-ant-your-api-key
  shotgun-cli config set llm.model claude-sonnet-4-20250514

  # Configure Google Gemini API
  shotgun-cli config set llm.provider gemini
  shotgun-cli config set llm.api-key your-gemini-api-key
  shotgun-cli config set llm.model gemini-2.5-flash

  # Use custom endpoint (OpenRouter, Azure, etc.)
  shotgun-cli config set llm.provider openai
  shotgun-cli config set llm.base-url https://openrouter.ai/api/v1
  shotgun-cli config set llm.api-key your-openrouter-key

  # Scanner settings
  shotgun-cli config set scanner.max-files 5000
  shotgun-cli config set context.max-size 5MB`,

	Args: cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if !config.IsValidKey(key) {
			return fmt.Errorf("invalid configuration key '%s'. Use 'shotgun-cli config show' to see available keys", key)
		}

		if err := config.ValidateValue(key, value); err != nil {
			return fmt.Errorf("invalid value for '%s': %w", key, err)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if err := setConfigValue(key, value); err != nil {
			return fmt.Errorf("failed to set configuration: %w", err)
		}

		fmt.Printf("✅ Configuration updated successfully!\n")
		fmt.Printf("📝 Set %s = %s\n", key, value)

		// Show where config was written
		configPath := viper.ConfigFileUsed()
		if configPath == "" {
			configPath = getDefaultConfigPath()
		}
		fmt.Printf("📁 Config file: %s\n", configPath)

		// Add helpful message for template.custom-path
		if key == "template.custom-path" && value != "" {
			fmt.Println("\n💡 The custom template directory will be created automatically on first use.")
		}

		return nil
	},
}

func showCurrentConfig() error {
	fmt.Println("Current Configuration:")
	fmt.Println("=====================")
	fmt.Println()

	// Get config file path
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		fmt.Println("Config file: Not found (using defaults)")
	} else {
		fmt.Printf("Config file: %s\n", configPath)
	}
	fmt.Println()

	// Get all configuration keys and organize them
	allKeys := viper.AllKeys()
	sort.Strings(allKeys)

	// Group keys by category
	categories := make(map[string][]string)
	for _, key := range allKeys {
		parts := strings.Split(key, ".")
		category := parts[0]
		categories[category] = append(categories[category], key)
	}

	// Display by category
	categoryOrder := []string{"scanner", "context", "template", "output", "llm", "gemini"}

	for _, category := range categoryOrder {
		if keys, exists := categories[category]; exists {
			fmt.Printf("[%s]\n", strings.ToUpper(category))
			for _, key := range keys {
				value := viper.Get(key)
				source := getConfigSource(key)
				fmt.Printf("  %-25s = %-15v (%s)\n", key, formatValue(value), source)
			}
			fmt.Println()
		}
	}

	// Show any remaining categories not in the predefined order
	for category, keys := range categories {
		found := false
		for _, predefined := range categoryOrder {
			if category == predefined {
				found = true

				break
			}
		}
		if !found {
			fmt.Printf("[%s]\n", strings.ToUpper(category))
			for _, key := range keys {
				value := viper.Get(key)
				source := getConfigSource(key)
				fmt.Printf("  %-25s = %-15v (%s)\n", key, formatValue(value), source)
			}
			fmt.Println()
		}
	}

	// Show Gemini integration status
	fmt.Println("[GEMINI STATUS]")
	status := getGeminiStatusSummary()
	fmt.Printf("  %-25s = %s\n", "integration", status)
	fmt.Println()

	return nil
}

// getGeminiStatusSummary returns a brief status summary for Gemini integration.
func getGeminiStatusSummary() string {
	if !viper.GetBool(config.KeyGeminiEnabled) {
		return "disabled"
	}

	// Lazy import to avoid circular dependency - check directly
	home, _ := os.UserHomeDir()
	cookiesPath := filepath.Join(home, ".geminiweb", "cookies.json")

	// Check if geminiweb exists in PATH
	_, err := exec.LookPath("geminiweb")
	if err != nil {
		return "✗ geminiweb not found (run: shotgun-cli gemini doctor)"
	}

	// Check cookies
	info, err := os.Stat(cookiesPath)
	if err != nil || info.Size() <= 2 {
		return "⚠ needs configuration (run: shotgun-cli gemini doctor)"
	}

	return "✓ ready"
}

func setConfigValue(key, value string) error {
	convertedValue, err := config.ConvertValue(key, value)
	if err != nil {
		return err
	}

	// Set the value in viper
	viper.Set(key, convertedValue)

	// Determine config file path
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		configPath = getDefaultConfigPath()
		viper.SetConfigFile(configPath)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the configuration
	if err := viper.WriteConfig(); err != nil {
		// If config file doesn't exist, create it
		if os.IsNotExist(err) {
			if err := viper.SafeWriteConfig(); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
		} else {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	log.Debug().Str("key", key).Interface("value", convertedValue).Str("path", configPath).Msg("Configuration updated")

	return nil
}

func getConfigSource(key string) string {
	// This is a simplified version - viper doesn't expose the actual source
	// We make educated guesses based on common patterns

	if viper.IsSet(key) {
		if viper.ConfigFileUsed() != "" {
			return "config file"
		}

		// Check if it might be from environment
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		envKey = "SHOTGUN_" + envKey
		if os.Getenv(envKey) != "" {
			return "environment"
		}

		return "flag/default"
	}

	return "default"
}

func formatValue(value interface{}) string {
	if value == nil {
		return "<nil>"
	}

	// Handle different types
	switch v := value.(type) {
	case string:
		if v == "" {
			return `""`
		}

		return fmt.Sprintf(`"%s"`, v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getDefaultConfigPath() string {
	return filepath.Join(getConfigDir(), "config.yaml")
}

func launchConfigTUI() error {
	wizard := ui.NewConfigWizard()

	program := tea.NewProgram(
		wizard,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("failed to start config TUI: %w", err)
	}

	return nil
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
