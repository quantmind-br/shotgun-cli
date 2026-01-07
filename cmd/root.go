package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/ui"
	"github.com/quantmind-br/shotgun-cli/internal/utils"
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
		_ = cmd.Help()
	}
}

func launchTUIWizard() {
	// Detect current working directory as scan root
	rootPath, err := os.Getwd()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get current working directory")
		fmt.Fprintf(os.Stderr, "Error: Could not determine current directory: %v\n", err)
		os.Exit(1)
	}

	scanConfig := &scanner.ScanConfig{
		MaxFiles:             viper.GetInt64(config.KeyScannerMaxFiles),
		MaxFileSize:          utils.ParseSizeWithDefault(viper.GetString(config.KeyScannerMaxFileSize), 1024*1024),
		MaxMemory:            utils.ParseSizeWithDefault(viper.GetString(config.KeyScannerMaxMemory), 500*1024*1024),
		SkipBinary:           viper.GetBool(config.KeyScannerSkipBinary),
		IncludeHidden:        viper.GetBool(config.KeyScannerIncludeHidden),
		IncludeIgnored:       true,
		Workers:              viper.GetInt(config.KeyScannerWorkers),
		RespectGitignore:     viper.GetBool(config.KeyScannerRespectGitignore),
		RespectShotgunignore: viper.GetBool(config.KeyScannerRespectShotgunignore),
	}

	wizardConfig := &ui.WizardConfig{
		LLM: ui.LLMConfig{
			Provider:       viper.GetString(config.KeyLLMProvider),
			APIKey:         viper.GetString(config.KeyLLMAPIKey),
			BaseURL:        viper.GetString(config.KeyLLMBaseURL),
			Model:          viper.GetString(config.KeyLLMModel),
			Timeout:        viper.GetInt(config.KeyLLMTimeout),
			SaveResponse:   viper.GetBool(config.KeyGeminiSaveResponse),
			BinaryPath:     viper.GetString(config.KeyGeminiBinaryPath),
			BrowserRefresh: viper.GetString(config.KeyGeminiBrowserRefresh),
		},
		Gemini: ui.GeminiConfig{
			BinaryPath:     viper.GetString(config.KeyGeminiBinaryPath),
			Model:          viper.GetString(config.KeyGeminiModel),
			Timeout:        viper.GetInt(config.KeyGeminiTimeout),
			BrowserRefresh: viper.GetString(config.KeyGeminiBrowserRefresh),
			SaveResponse:   viper.GetBool(config.KeyGeminiSaveResponse),
		},
		Context: ui.ContextConfig{
			IncludeTree:    viper.GetBool(config.KeyContextIncludeTree),
			IncludeSummary: viper.GetBool(config.KeyContextIncludeSummary),
			MaxSize:        viper.GetString(config.KeyContextMaxSize),
		},
	}

	wizard := ui.NewWizard(rootPath, scanConfig, wizardConfig, nil)

	// Configure Bubble Tea program
	program := tea.NewProgram(
		wizard,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Handle terminal size detection
	if _, err := program.Run(); err != nil {
		log.Error().Err(err).Msg("Failed to start TUI wizard")
		fmt.Fprintf(os.Stderr, "Error starting wizard: %v\n", err)
		os.Exit(1)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "", "config file (default is ~/.config/shotgun-cli/config.yaml)")
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
		_ = viper.BindPFlag(config.KeyVerbose, rootCmd.PersistentFlags().Lookup("verbose"))
	}
	if rootCmd.PersistentFlags().Lookup("quiet") != nil {
		_ = viper.BindPFlag(config.KeyQuiet, rootCmd.PersistentFlags().Lookup("quiet"))
	}

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
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
	viper.SetDefault(config.KeyScannerMaxFiles, 10000)
	viper.SetDefault(config.KeyScannerMaxFileSize, "1MB")
	viper.SetDefault(config.KeyScannerRespectGitignore, true)
	viper.SetDefault(config.KeyScannerSkipBinary, true)
	viper.SetDefault(config.KeyScannerWorkers, 1)
	viper.SetDefault(config.KeyScannerIncludeHidden, false)
	viper.SetDefault(config.KeyScannerIncludeIgnored, false)
	viper.SetDefault(config.KeyScannerRespectShotgunignore, true)
	viper.SetDefault(config.KeyScannerMaxMemory, "500MB")

	viper.SetDefault(config.KeyContextMaxSize, "10MB")
	viper.SetDefault(config.KeyContextIncludeTree, true)
	viper.SetDefault(config.KeyContextIncludeSummary, true)

	viper.SetDefault(config.KeyTemplateCustomPath, "")

	viper.SetDefault(config.KeyOutputFormat, "markdown")
	viper.SetDefault(config.KeyOutputClipboard, true)

	viper.SetDefault(config.KeyLLMProvider, "geminiweb")
	viper.SetDefault(config.KeyLLMAPIKey, "")
	viper.SetDefault(config.KeyLLMBaseURL, "")
	viper.SetDefault(config.KeyLLMModel, "")
	viper.SetDefault(config.KeyLLMTimeout, 300)

	viper.SetDefault(config.KeyGeminiEnabled, false)
	viper.SetDefault(config.KeyGeminiBinaryPath, "")
	viper.SetDefault(config.KeyGeminiModel, "gemini-2.5-flash")
	viper.SetDefault(config.KeyGeminiTimeout, 300)
	viper.SetDefault(config.KeyGeminiBrowserRefresh, "auto")
	viper.SetDefault(config.KeyGeminiAutoSend, false)
	viper.SetDefault(config.KeyGeminiSaveResponse, true)
}

func updateLoggingLevel() {
	level := zerolog.InfoLevel

	if viper.GetBool(config.KeyQuiet) {
		level = zerolog.ErrorLevel
	} else if viper.GetBool(config.KeyVerbose) {
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)
}
