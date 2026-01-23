package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/config"
)

var contextSendCmd = &cobra.Command{
	Use:   "send [file]",
	Short: "Send a context file to Gemini",
	Long: `Send an existing context file (or stdin) directly to Google Gemini.

This command sends the content of a file or stdin to Gemini and captures
the response. It requires geminiweb to be installed and configured.

Prerequisites:
  1. Install geminiweb: go install github.com/diogo/geminiweb/cmd/geminiweb@latest
  2. Configure authentication: geminiweb auto-login

Examples:
  shotgun-cli context send prompt.md
  shotgun-cli context send prompt.md -o response.md
  cat prompt.md | shotgun-cli context send
  shotgun-cli context send prompt.md -m gemini-3.0-pro
  shotgun-cli context send prompt.md --raw`,

	Args: cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if file exists when specified
		if len(args) > 0 {
			if _, err := os.Stat(args[0]); os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", args[0])
			}
		}
		return nil
	},
	RunE: runContextSend,
}

func runContextSend(cmd *cobra.Command, args []string) error {
	var content string

	// Read content from file or stdin
	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file '%s': %w", args[0], err)
		}
		content = string(data)
		log.Debug().Str("file", args[0]).Int("size", len(content)).Msg("Read content from file")
	} else {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("no input provided. Specify a file or pipe content via stdin")
		}

		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		content = string(data)
		log.Debug().Int("size", len(content)).Msg("Read content from stdin")
	}

	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("no content to send (file or stdin is empty)")
	}

	// Check if any LLM provider is enabled/configured
	provider := viper.GetString(config.KeyLLMProvider)

	// Backward compatibility: if llm.provider is geminiweb, check gemini.enabled
	if provider == "" || provider == "geminiweb" {
		if !viper.GetBool(config.KeyGeminiEnabled) {
			return fmt.Errorf("LLM integration is disabled. Enable with: " +
				"shotgun-cli config set gemini.enabled true\n" +
				"Or configure a different provider: shotgun-cli llm list")
		}
	}

	// Get flag overrides
	model, _ := cmd.Flags().GetString("model")
	timeout, _ := cmd.Flags().GetInt("timeout")
	outputFile, _ := cmd.Flags().GetString("output")
	raw, _ := cmd.Flags().GetBool("raw")

	// Check save-response config if no output file specified
	saveResponse := viper.GetBool(config.KeyLLMSaveResponse) || viper.GetBool(config.KeyGeminiSaveResponse)
	if outputFile == "" && saveResponse {
		// Auto-generate output filename
		timestamp := time.Now().Format("20060102-150405")
		outputFile = fmt.Sprintf("llm-response-%s.md", timestamp)
	}

	// Build config
	cfg := BuildLLMConfigWithOverrides(model, timeout)

	// Create provider
	llmProvider, err := CreateLLMProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	if !llmProvider.IsAvailable() {
		return fmt.Errorf("%s not available. Run 'shotgun-cli llm doctor' for help", llmProvider.Name())
	}

	if err := llmProvider.ValidateConfig(); err != nil {
		return fmt.Errorf("%s configuration error: %w. Run 'shotgun-cli llm doctor' for help", llmProvider.Name(), err)
	}

	// Send
	log.Info().
		Str("provider", llmProvider.Name()).
		Str("model", cfg.Model).
		Int("content_length", len(content)).
		Msg("Sending to LLM")

	fmt.Printf("Sending to %s (%s)...\n", llmProvider.Name(), cfg.Model)

	ctx := context.Background()
	result, err := llmProvider.Send(ctx, content)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	// Get response
	response := result.Response
	if raw {
		response = result.RawResponse
	}

	// Output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(response), 0600); err != nil {
			return fmt.Errorf("failed to save response to '%s': %w", outputFile, err)
		}
		fmt.Printf("Response saved to: %s\n", outputFile)
	} else {
		fmt.Println(response)
	}

	// Show usage if available
	if result.Usage != nil {
		fmt.Printf("Tokens: %d (prompt: %d, completion: %d)\n",
			result.Usage.TotalTokens,
			result.Usage.PromptTokens,
			result.Usage.CompletionTokens)
	}
	fmt.Printf("Duration: %s\n", formatDuration(result.Duration))

	return nil
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func init() {
	contextSendCmd.Flags().StringP("output", "o", "", "Output file for Gemini response")
	contextSendCmd.Flags().StringP("model", "m", "", "Gemini model to use (default: from config)")
	contextSendCmd.Flags().Int("timeout", 0, "Timeout in seconds (default: from config)")
	contextSendCmd.Flags().Bool("raw", false, "Output raw response without processing")

	contextCmd.AddCommand(contextSendCmd)
}
