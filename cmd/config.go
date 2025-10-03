package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  "Commands for viewing and modifying shotgun-cli configuration",
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
  scanner.max-files           - Maximum number of files to scan (default: 10000)
  scanner.max-file-size       - Maximum size per file (default: "1MB")
  scanner.respect-gitignore   - Respect .gitignore files (default: true)
  context.max-size           - Maximum context size (default: "10MB")
  context.include-tree       - Include directory tree (default: true)
  context.include-summary    - Include file summaries (default: true)
  template.custom-path       - Path to custom templates (default: "")
  output.format              - Output format (default: "markdown")
  output.clipboard           - Copy to clipboard (default: true)

Examples:
  shotgun-cli config set scanner.max-files 5000
  shotgun-cli config set context.max-size 5MB
  shotgun-cli config set output.clipboard false
  shotgun-cli config set scanner.respect-gitignore true`,

	Args: cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// Validate configuration key
		if !isValidConfigKey(key) {
			return fmt.Errorf("invalid configuration key '%s'. Use 'shotgun-cli config show' to see available keys", key)
		}

		// Validate value format for specific keys
		if err := validateConfigValue(key, value); err != nil {
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
	categoryOrder := []string{"scanner", "context", "template", "output"}

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

	return nil
}

func setConfigValue(key, value string) error {
	// Convert string value to appropriate type
	convertedValue, err := convertConfigValue(key, value)
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
	if err := os.MkdirAll(configDir, 0755); err != nil {
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

func isValidConfigKey(key string) bool {
	validKeys := []string{
		"scanner.max-files",
		"scanner.max-file-size",
		"scanner.respect-gitignore",
		"context.max-size",
		"context.include-tree",
		"context.include-summary",
		"template.custom-path",
		"output.format",
		"output.clipboard",
	}

	for _, validKey := range validKeys {
		if key == validKey {
			return true
		}
	}
	return false
}

func validateConfigValue(key, value string) error {
	switch key {
	case "scanner.max-files":
		return validateMaxFiles(value)
	case "scanner.max-file-size", "context.max-size":
		return validateSizeFormat(value)
	case "scanner.respect-gitignore", "context.include-tree", "context.include-summary", "output.clipboard":
		return validateBooleanValue(value)
	case "output.format":
		return validateOutputFormat(value)
	case "template.custom-path":
		return validateTemplatePath(value)
	}
	return nil
}

func validateMaxFiles(value string) error {
	if _, err := parseSize(value); err == nil {
		return fmt.Errorf("expected a number, got size format")
	}
	var dummy int
	if _, err := fmt.Sscanf(value, "%d", &dummy); err != nil {
		return fmt.Errorf("expected a positive integer")
	}
	if dummy <= 0 {
		return fmt.Errorf("must be positive, got %d", dummy)
	}
	return nil
}

func validateSizeFormat(value string) error {
	if _, err := parseSize(value); err != nil {
		return fmt.Errorf("expected size format (e.g., 1MB, 500KB): %w", err)
	}
	return nil
}

func validateBooleanValue(value string) error {
	lower := strings.ToLower(value)
	if lower != "true" && lower != "false" {
		return fmt.Errorf("expected 'true' or 'false', got '%s'", value)
	}
	return nil
}

func validateOutputFormat(value string) error {
	if value != "markdown" && value != "text" {
		return fmt.Errorf("expected 'markdown' or 'text', got '%s'", value)
	}
	return nil
}

func validateTemplatePath(value string) error {
	if value == "" {
		return nil
	}

	expandedValue := value
	if strings.HasPrefix(value, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to expand home directory: %w", err)
		}
		expandedValue = filepath.Join(home, value[2:])
	}

	parentDir := filepath.Dir(expandedValue)
	if parentDir != "." && parentDir != "/" {
		if info, err := os.Stat(parentDir); err == nil {
			if !info.IsDir() {
				return fmt.Errorf("parent path exists but is not a directory: %s", parentDir)
			}
		}
	}
	return nil
}

func convertConfigValue(key, value string) (interface{}, error) {
	switch key {
	case "scanner.max-files":
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err != nil {
			return nil, err
		}
		return intVal, nil

	case "scanner.respect-gitignore", "context.include-tree", "context.include-summary", "output.clipboard":
		return strings.ToLower(value) == "true", nil

	default:
		// String values
		return value, nil
	}
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

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}