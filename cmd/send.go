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

	"github.com/quantmind-br/shotgun-cli/internal/app"
	"github.com/quantmind-br/shotgun-cli/internal/config"
)

var contextSendCmd = &cobra.Command{
	Use:   "send [file]",
	Short: "Send a context file to the configured LLM provider",
	Long: `Send an existing context file (or stdin) directly to the configured LLM provider.

This command sends the content of a file or stdin to the configured LLM provider
and captures the response.

Examples:
  shotgun-cli context send prompt.md
  shotgun-cli context send prompt.md -o response.md
  cat prompt.md | shotgun-cli context send
  shotgun-cli context send prompt.md -m gemini-2.0-pro
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

// newSendService is a variable so tests can substitute a registry-injected service.
var newSendService = func() app.ContextService {
	return app.NewContextService()
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

	// Get flag overrides
	model, _ := cmd.Flags().GetString("model")
	timeout, _ := cmd.Flags().GetInt("timeout")
	outputFile, _ := cmd.Flags().GetString("output")
	raw, _ := cmd.Flags().GetBool("raw")

	// Check save-response config if no output file specified
	saveResponse := viper.GetBool(config.KeyLLMSaveResponse)
	if outputFile == "" && saveResponse {
		// Auto-generate output filename
		timestamp := time.Now().Format("20060102-150405")
		outputFile = fmt.Sprintf("llm-response-%s.md", timestamp)
	}

	cfg := BuildLLMConfigWithOverrides(model, timeout)

	// The service writes result.Response itself; --raw needs result.RawResponse,
	// so that case opts out of the service write and saves below instead.
	sendCfg := app.LLMSendConfig{
		Provider:     cfg.Provider,
		APIKey:       cfg.APIKey,
		BaseURL:      cfg.BaseURL,
		Model:        cfg.Model,
		Timeout:      cfg.Timeout,
		SaveResponse: outputFile != "" && !raw,
		OutputPath:   outputFile,
	}

	log.Info().
		Str("provider", string(cfg.Provider)).
		Str("model", cfg.Model).
		Int("content_length", len(content)).
		Msg("Sending to LLM")

	fmt.Printf("Sending to %s (%s)...\n", cfg.Provider, cfg.Model)

	result, err := newSendService().SendToLLMWithProgress(context.Background(), content, sendCfg, nil)
	if err != nil {
		return err
	}

	response := result.Response
	if raw {
		response = result.RawResponse
	}

	if outputFile != "" {
		if raw {
			// #nosec G703 -- o caminho vem de --output, informado pelo operador que já roda o binário.
			if writeErr := os.WriteFile(outputFile, []byte(response), 0600); writeErr != nil {
				return fmt.Errorf("failed to save response to '%s': %w", outputFile, writeErr)
			}
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
	contextSendCmd.Flags().StringP("output", "o", "", "Output file for the LLM response")
	contextSendCmd.Flags().StringP("model", "m", "", "Model to use (default: from config)")
	contextSendCmd.Flags().Int("timeout", 0, "Timeout in seconds (default: from config)")
	contextSendCmd.Flags().Bool("raw", false, "Output raw response without processing")

	contextCmd.AddCommand(contextSendCmd)
}
