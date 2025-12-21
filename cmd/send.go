package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/platform/gemini"
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

	// Check if Gemini is enabled
	if !viper.GetBool("gemini.enabled") {
		return fmt.Errorf("gemini integration is disabled, enable with: shotgun-cli config set gemini.enabled true")
	}

	// Check availability
	if !gemini.IsAvailable() {
		return fmt.Errorf("geminiweb not found. Run 'shotgun-cli gemini doctor' for help")
	}

	if !gemini.IsConfigured() {
		return fmt.Errorf("geminiweb not configured. Run 'shotgun-cli gemini doctor' for help")
	}

	// Get flags
	model, _ := cmd.Flags().GetString("model")
	if model == "" {
		model = viper.GetString("gemini.model")
	}

	timeout, _ := cmd.Flags().GetInt("timeout")
	if timeout == 0 {
		timeout = viper.GetInt("gemini.timeout")
	}

	outputFile, _ := cmd.Flags().GetString("output")
	raw, _ := cmd.Flags().GetBool("raw")

	// Build config
	cfg := gemini.Config{
		BinaryPath:     viper.GetString("gemini.binary-path"),
		Model:          model,
		Timeout:        timeout,
		BrowserRefresh: viper.GetString("gemini.browser-refresh"),
		Verbose:        viper.GetBool("verbose"),
	}

	executor := gemini.NewExecutor(cfg)

	log.Info().Str("model", model).Int("content_length", len(content)).Msg("Sending to Gemini")
	fmt.Printf("🤖 Sending to Gemini (%s)...\n", model)

	ctx := context.Background()
	result, err := executor.Send(ctx, content)
	if err != nil {
		return fmt.Errorf("gemini request failed: %w", err)
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
		fmt.Printf("✅ Response saved to: %s\n", outputFile)
		fmt.Printf("⏱️  Response time: %s\n", gemini.FormatDuration(result.Duration))
	} else {
		fmt.Println(response)
	}

	return nil
}

func init() {
	contextSendCmd.Flags().StringP("output", "o", "", "Output file for Gemini response")
	contextSendCmd.Flags().StringP("model", "m", "", "Gemini model to use (default: from config)")
	contextSendCmd.Flags().Int("timeout", 0, "Timeout in seconds (default: from config)")
	contextSendCmd.Flags().Bool("raw", false, "Output raw response without processing")

	contextCmd.AddCommand(contextSendCmd)
}
