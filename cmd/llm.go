package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "LLM provider management",
	Long:  "Commands for managing and diagnosing LLM provider configuration",
}

var llmStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show LLM provider status",
	Long: `Display the current LLM provider configuration and status.

Shows the active provider, model, API key status, and whether the
provider is ready to use.

Example:
  shotgun-cli llm status`,
	RunE: runLLMStatus,
}

var llmDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose LLM configuration",
	Long: `Run diagnostics on the LLM provider configuration.

Checks all aspects of the provider configuration and provides
specific guidance on how to fix any issues found.

Example:
  shotgun-cli llm doctor`,
	RunE: runLLMDoctor,
}

var llmListCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported providers",
	Long: `List all supported LLM providers and their status.

Shows which providers are available and how to configure them.

Example:
  shotgun-cli llm list`,
	RunE: runLLMList,
}

func runLLMStatus(cmd *cobra.Command, args []string) error {
	cfg := BuildLLMConfig()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "=== LLM Configuration ===")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Provider:\t%s\n", cfg.Provider)
	fmt.Fprintf(w, "Model:\t%s\n", cfg.Model)
	fmt.Fprintf(w, "Base URL:\t%s\n", displayURL(cfg.BaseURL, cfg.Provider))
	fmt.Fprintf(w, "API Key:\t%s\n", cfg.MaskAPIKey())
	fmt.Fprintf(w, "Timeout:\t%ds\n", cfg.Timeout)

	w.Flush()

	// Test provider
	fmt.Println()
	provider, err := CreateLLMProvider(cfg)
	if err != nil {
		fmt.Printf("Status: Not ready - %s\n", err)
		return nil
	}

	if err := provider.ValidateConfig(); err != nil {
		fmt.Printf("Status: Not configured - %s\n", err)
		return nil
	}

	if !provider.IsAvailable() {
		fmt.Printf("Status: Not available\n")
		return nil
	}

	if !provider.IsConfigured() {
		fmt.Printf("Status: Not configured\n")
		return nil
	}

	fmt.Printf("Status: Ready\n")
	return nil
}

func runLLMDoctor(cmd *cobra.Command, args []string) error {
	cfg := BuildLLMConfig()

	fmt.Printf("Running diagnostics for %s...\n\n", cfg.Provider)

	issues := []string{}

	// Check 1: Provider type
	fmt.Print("Checking provider... ")
	if llm.IsValidProvider(string(cfg.Provider)) {
		fmt.Printf("%s\n", cfg.Provider)
	} else {
		fmt.Printf("invalid: %s\n", cfg.Provider)
		issues = append(issues, fmt.Sprintf("Invalid provider: %s", cfg.Provider))
	}

	// Check 2: API Key (except GeminiWeb)
	if cfg.Provider != llm.ProviderGeminiWeb {
		fmt.Print("Checking API key... ")
		if cfg.APIKey != "" {
			fmt.Println("configured")
		} else {
			fmt.Println("not configured")
			issues = append(issues, "API key not configured")
		}
	}

	// Check 3: Model
	fmt.Print("Checking model... ")
	if cfg.Model != "" {
		fmt.Printf("%s\n", cfg.Model)
	} else {
		fmt.Println("not configured")
		issues = append(issues, "Model not configured")
	}

	// Check 4: Provider-specific
	provider, err := CreateLLMProvider(cfg)
	if err == nil {
		fmt.Print("Checking provider availability... ")
		if provider.IsAvailable() {
			fmt.Println("OK")
		} else {
			fmt.Println("not available")
			issues = append(issues, fmt.Sprintf("%s is not available", provider.Name()))
		}

		fmt.Print("Checking provider configuration... ")
		if provider.IsConfigured() {
			fmt.Println("OK")
		} else {
			fmt.Println("not configured")
			issues = append(issues, fmt.Sprintf("%s is not fully configured", provider.Name()))
		}
	}

	// Summary
	fmt.Println()
	if len(issues) == 0 {
		fmt.Printf("No issues found! %s is ready.\n", cfg.Provider)
		return nil
	}

	fmt.Printf("Found %d issue(s):\n", len(issues))
	for i, issue := range issues {
		fmt.Printf("  %d. %s\n", i+1, issue)
	}

	// Provider-specific help
	fmt.Println("\nNext steps:")
	switch cfg.Provider {
	case llm.ProviderOpenAI:
		fmt.Println("  1. Get API key from: https://platform.openai.com/api-keys")
		fmt.Println("  2. Configure: shotgun-cli config set llm.api-key YOUR_KEY")
		fmt.Println("  3. (Optional) Set model: shotgun-cli config set llm.model gpt-4o")
	case llm.ProviderAnthropic:
		fmt.Println("  1. Get API key from: https://console.anthropic.com/settings/keys")
		fmt.Println("  2. Configure: shotgun-cli config set llm.api-key YOUR_KEY")
		fmt.Println("  3. (Optional) Set model: shotgun-cli config set llm.model claude-sonnet-4-20250514")
	case llm.ProviderGemini:
		fmt.Println("  1. Get API key from: https://aistudio.google.com/app/apikey")
		fmt.Println("  2. Configure: shotgun-cli config set llm.api-key YOUR_KEY")
		fmt.Println("  3. (Optional) Set model: shotgun-cli config set llm.model gemini-2.5-flash")
	case llm.ProviderGeminiWeb:
		fmt.Println("  1. Install: go install github.com/diogo/geminiweb/cmd/geminiweb@latest")
		fmt.Println("  2. Configure: geminiweb auto-login")
		fmt.Println("  3. Enable: shotgun-cli config set gemini.enabled true")
	}

	return nil
}

func runLLMList(cmd *cobra.Command, args []string) error {
	fmt.Println("Supported LLM Providers:")
	fmt.Println()

	providers := []struct {
		id        llm.ProviderType
		name      string
		desc      string
		apiKeyURL string
	}{
		{llm.ProviderOpenAI, "OpenAI", "GPT-4o, GPT-4, o1, o3", "https://platform.openai.com/api-keys"},
		{llm.ProviderAnthropic, "Anthropic", "Claude 4, Claude 3.5", "https://console.anthropic.com/settings/keys"},
		{llm.ProviderGemini, "Google Gemini", "Gemini 2.5, Gemini 2.0", "https://aistudio.google.com/app/apikey"},
		{llm.ProviderGeminiWeb, "GeminiWeb", "Browser-based (no API key)", "N/A"},
	}

	current := viper.GetString(config.KeyLLMProvider)

	for _, p := range providers {
		marker := "  "
		if string(p.id) == current {
			marker = "* "
		}
		fmt.Printf("%s%-12s - %s (%s)\n", marker, p.id, p.name, p.desc)
	}

	fmt.Println()
	fmt.Println("Configure with:")
	fmt.Println("  shotgun-cli config set llm.provider <provider>")
	fmt.Println("  shotgun-cli config set llm.api-key <your-api-key>")
	fmt.Println()
	fmt.Println("For custom endpoints (OpenRouter, Azure, etc.):")
	fmt.Println("  shotgun-cli config set llm.base-url https://openrouter.ai/api/v1")

	return nil
}

func displayURL(url string, provider llm.ProviderType) string {
	if url == "" {
		defaults := llm.DefaultConfigs()
		if d, ok := defaults[provider]; ok && d.BaseURL != "" {
			return fmt.Sprintf("(default: %s)", d.BaseURL)
		}
		return "(default)"
	}
	return url
}

func init() {
	llmCmd.AddCommand(llmStatusCmd)
	llmCmd.AddCommand(llmDoctorCmd)
	llmCmd.AddCommand(llmListCmd)
	rootCmd.AddCommand(llmCmd)
}
