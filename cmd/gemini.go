package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/platform/gemini"
)

const (
	iconCheck = "✓"
	iconCross = "✗"
)

var geminiCmd = &cobra.Command{
	Use:   "gemini",
	Short: "Gemini integration commands",
	Long:  `Commands for managing and diagnosing the Gemini integration via geminiweb.`,
}

var geminiStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Gemini integration status",
	Long: `Check and display the status of the Gemini integration, including:
- Whether the geminiweb binary is available
- Whether cookies are configured
- Current configuration settings
- Next steps if something is missing

Examples:
  shotgun-cli gemini status
  shotgun-cli gemini status --json`,

	RunE: runGeminiStatus,
}

var geminiDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose Gemini integration issues",
	Long: `Run diagnostics on the Gemini integration and suggest fixes.

This command checks:
- geminiweb binary availability
- Cookie configuration
- Config settings validity
- Model configuration

Examples:
  shotgun-cli gemini doctor`,

	RunE: runGeminiDoctor,
}

func init() {
	geminiStatusCmd.Flags().Bool("json", false, "Output in JSON format")

	geminiCmd.AddCommand(geminiStatusCmd)
	geminiCmd.AddCommand(geminiDoctorCmd)
	rootCmd.AddCommand(geminiCmd)
}

func runGeminiStatus(cmd *cobra.Command, args []string) error {
	status := gemini.GetStatus()
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		return printGeminiStatusJSON(status)
	}

	return printGeminiStatusHuman(status)
}

func printGeminiStatusHuman(status gemini.Status) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(w, "=== Gemini Integration Status ===")
	_, _ = fmt.Fprintln(w)

	// Enabled status
	enabled := viper.GetBool(config.KeyGeminiEnabled)
	enabledIcon := iconCross
	if enabled {
		enabledIcon = iconCheck
	}
	_, _ = fmt.Fprintf(w, "Enabled:\t%s %v\n", enabledIcon, enabled)

	// Binary availability
	availIcon := iconCross
	if status.Available {
		availIcon = iconCheck
	}
	binaryInfo := "not found"
	if status.BinaryPath != "" {
		binaryInfo = status.BinaryPath
	}
	_, _ = fmt.Fprintf(w, "Binary (geminiweb):\t%s %s\n", availIcon, binaryInfo)

	// Cookies configuration
	configIcon := iconCross
	if status.Configured {
		configIcon = iconCheck
	}
	_, _ = fmt.Fprintf(w, "Cookies configured:\t%s %s\n", configIcon, status.CookiesPath)

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "=== Configuration ===")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "Model:\t%s\n", viper.GetString(config.KeyGeminiModel))
	_, _ = fmt.Fprintf(w, "Timeout:\t%ds\n", viper.GetInt(config.KeyGeminiTimeout))
	_, _ = fmt.Fprintf(w, "Auto-send:\t%v\n", viper.GetBool(config.KeyGeminiAutoSend))
	_, _ = fmt.Fprintf(w, "Save response:\t%v\n", viper.GetBool(config.KeyGeminiSaveResponse))
	_, _ = fmt.Fprintf(w, "Browser refresh:\t%s\n", viper.GetString(config.KeyGeminiBrowserRefresh))

	_ = w.Flush()

	// Status summary and next steps
	fmt.Println()

	switch {
	case status.Error != "":
		fmt.Println("⚠️  Issue detected:", status.Error)
		printGeminiNextSteps(status)
	case !enabled:
		fmt.Println("💡 Gemini is configured but disabled.")
		fmt.Println("   To enable: shotgun-cli config set gemini.enabled true")
	default:
		fmt.Println("✅ Gemini is ready to use!")
	}

	return nil
}

func printGeminiNextSteps(status gemini.Status) {
	fmt.Println()
	fmt.Println("📋 Next steps:")

	if !status.Available {
		fmt.Println("   1. Install geminiweb:")
		fmt.Println("      go install github.com/diogo/geminiweb/cmd/geminiweb@latest")
		fmt.Println()
		fmt.Println("   2. Or configure a custom path:")
		fmt.Println("      shotgun-cli config set gemini.binary-path /path/to/geminiweb")
	} else if !status.Configured {
		fmt.Println("   1. Authenticate with Google:")
		fmt.Println("      geminiweb auto-login")
		fmt.Println()
		fmt.Println("   2. Follow the browser instructions to complete authentication")
	}
}

func printGeminiStatusJSON(status gemini.Status) error {
	output := struct {
		gemini.Status
		Enabled        bool   `json:"enabled"`
		Model          string `json:"model"`
		Timeout        int    `json:"timeout"`
		AutoSend       bool   `json:"autoSend"`
		SaveResponse   bool   `json:"saveResponse"`
		BrowserRefresh string `json:"browserRefresh"`
	}{
		Status:         status,
		Enabled:        viper.GetBool(config.KeyGeminiEnabled),
		Model:          viper.GetString(config.KeyGeminiModel),
		Timeout:        viper.GetInt(config.KeyGeminiTimeout),
		AutoSend:       viper.GetBool(config.KeyGeminiAutoSend),
		SaveResponse:   viper.GetBool(config.KeyGeminiSaveResponse),
		BrowserRefresh: viper.GetString(config.KeyGeminiBrowserRefresh),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	fmt.Println(string(data))

	return nil
}

func runGeminiDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 Running Gemini integration diagnostics...")
	fmt.Println()

	status := gemini.GetStatus()
	issues := []string{}

	// Check 1: Binary
	fmt.Print("Checking geminiweb binary... ")
	if status.Available {
		fmt.Printf("✓ found at %s\n", status.BinaryPath)
	} else {
		fmt.Println("✗ not found")
		issues = append(issues, "geminiweb binary not found")
	}

	// Check 2: Cookies
	fmt.Print("Checking cookies... ")
	if status.Configured {
		fmt.Printf("✓ found at %s\n", status.CookiesPath)
	} else {
		fmt.Println("✗ not found or empty")
		issues = append(issues, "Cookies file not found or empty")
	}

	// Check 3: Enabled config
	fmt.Print("Checking configuration... ")
	if viper.GetBool(config.KeyGeminiEnabled) {
		fmt.Println("✓ enabled")
	} else {
		fmt.Println("⚠ disabled")
		issues = append(issues, "Gemini integration is disabled in configuration")
	}

	// Check 4: Model configuration
	fmt.Print("Checking model... ")
	model := viper.GetString(config.KeyGeminiModel)
	if model != "" {
		fmt.Printf("✓ %s\n", model)
	} else {
		fmt.Println("⚠ not set (will use default)")
	}

	// Check 5: Timeout
	fmt.Print("Checking timeout... ")
	timeout := viper.GetInt(config.KeyGeminiTimeout)
	if timeout > 0 && timeout <= 3600 {
		fmt.Printf("✓ %ds\n", timeout)
	} else {
		fmt.Printf("⚠ %ds (recommended: 60-600)\n", timeout)
	}

	// Summary
	fmt.Println()
	if len(issues) == 0 {
		fmt.Println("✅ No issues found! Gemini integration is ready.")
		return nil
	}

	fmt.Printf("⚠️  Found %d issue(s):\n", len(issues))
	for i, issue := range issues {
		fmt.Printf("   %d. %s\n", i+1, issue)
	}

	printGeminiNextSteps(status)

	return nil
}
