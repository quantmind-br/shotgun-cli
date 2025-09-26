package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

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
	if v, _ := cmd.Flags().GetBool("version"); v {
		fmt.Println(version)
		return
	}

	// If no arguments provided, launch TUI wizard
	if len(args) == 0 && !cmd.Flags().Changed("help") {
		launchTUIWizard()
		return
	}

	// Otherwise show help
	cmd.Help()
}

func launchTUIWizard() {
	// TODO: This will be implemented in Phase 2
	// For now, just print a placeholder message
	fmt.Println("🎯 shotgun-cli TUI Wizard")
	fmt.Println("Interactive 5-step wizard will be available in the next phase.")
	fmt.Println("For now, use 'shotgun-cli --help' to see available commands.")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.shotgun-cli.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet output")

	// Local flags
	rootCmd.Flags().BoolP("version", "", false, "show version information")
}

func initConfig() {
	// TODO: Configuration initialization will be implemented in later phases
	log.Debug().Msg("Configuration initialization placeholder")
}