package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// Required UI dependencies as per verification comment
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/lipgloss"
	_ "github.com/sabhiram/go-gitignore"
)

var (
	version = "dev" // Will be set during build
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "shotgun-cli",
	Short: "Generate LLM-optimized codebase contexts",
	Long: `shotgun-cli is a cross-platform CLI tool that generates LLM-optimized
codebase contexts with both TUI wizard and headless CLI modes.

When called without arguments, it launches an interactive 5-step wizard.
When called with arguments, it runs in headless CLI mode.`,
	Version: version,
	Run:     runRootCommand,
}

func runRootCommand(cmd *cobra.Command, args []string) {
	// Check for version flag
	if v, _ := cmd.Flags().GetBool("version"); v {
		fmt.Printf("shotgun-cli version %s\n", version)
		return
	}

	// If no subcommands and no flags (except global ones), launch TUI wizard
	if len(args) == 0 && len(os.Args) == 1 {
		log.Info().Msg("Launching TUI wizard...")
		launchTUIWizard()
		return
	}

	// If args provided but no valid subcommand, show help
	if len(args) == 0 {
		cmd.Help()
	}
}

func launchTUIWizard() {
	// TODO: This will be implemented in Phase 2
	// For now, just print a placeholder message
	fmt.Println("🎯 shotgun-cli TUI Wizard")
	fmt.Println("Interactive 5-step wizard will be available in the next phase.")
	fmt.Println("For now, use 'shotgun-cli --help' to see available commands.")
	fmt.Println("")
	fmt.Println("Available commands:")
	fmt.Println("  context generate - Generate LLM-optimized context")
	fmt.Println("  template list    - List available templates")
	fmt.Println("  template render  - Render templates with variables")
	fmt.Println("  diff split       - Split large diff files")
	fmt.Println("  config show      - Show current configuration")
	fmt.Println("  config set       - Set configuration values")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.config/shotgun-cli/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet output")

	// Local flags
	rootCmd.Flags().BoolP("version", "", false, "show version information")

	// Hide completion command from help
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

func initConfig() {
	// Initialize logging first with basic setup
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05",
	})

	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in multiple locations
		home, err := os.UserHomeDir()
		if err != nil {
			log.Debug().Err(err).Msg("Could not determine home directory")
			return
		}

		// Add config paths based on platform
		viper.AddConfigPath(getConfigDir())
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetEnvPrefix("SHOTGUN")

	// Set default values
	setConfigDefaults()

	// Bind persistent flags to viper
	if rootCmd.PersistentFlags().Lookup("verbose") != nil {
		viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	}
	if rootCmd.PersistentFlags().Lookup("quiet") != nil {
		viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	}

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use defaults
			log.Debug().Msg("No config file found, using defaults")
		} else {
			// Config file was found but another error was produced
			log.Debug().Err(err).Msg("Error reading config file")
		}
	} else {
		log.Debug().Str("config", viper.ConfigFileUsed()).Msg("Using config file")
	}

	// Update logging based on final configuration
	updateLoggingLevel()
}

func getConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "shotgun-cli")
		}
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "shotgun-cli")
	default: // Linux and others
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			return filepath.Join(xdgConfig, "shotgun-cli")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "shotgun-cli")
	}
	return "."
}

func setConfigDefaults() {
	// Scanner defaults
	viper.SetDefault("scanner.max-files", 10000)
	viper.SetDefault("scanner.max-file-size", "1MB")
	viper.SetDefault("scanner.respect-gitignore", true)

	// Context generation defaults
	viper.SetDefault("context.max-size", "10MB")
	viper.SetDefault("context.include-tree", true)
	viper.SetDefault("context.include-summary", true)

	// Template defaults
	viper.SetDefault("template.custom-path", "")

	// Output defaults
	viper.SetDefault("output.format", "markdown")
	viper.SetDefault("output.clipboard", true)
}

func updateLoggingLevel() {
	level := zerolog.InfoLevel

	if viper.GetBool("quiet") {
		level = zerolog.ErrorLevel
	} else if viper.GetBool("verbose") {
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)
}