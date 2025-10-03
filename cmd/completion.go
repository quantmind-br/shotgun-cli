package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	configKeyTemplateCustomPath = "template.custom-path"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `Generate shell completion scripts for shotgun-cli.

The completion script provides intelligent tab completion for commands, flags,
and dynamic values like template names and configuration keys.

Installation instructions:

Bash:
  # Linux:
  shotgun-cli completion bash | sudo tee /etc/bash_completion.d/shotgun-cli > /dev/null

  # macOS:
  shotgun-cli completion bash | sudo tee /usr/local/etc/bash_completion.d/shotgun-cli > /dev/null

Zsh:
  # Add to ~/.zshrc:
  autoload -U compinit; compinit
  source <(shotgun-cli completion zsh)

  # Or generate to file:
  shotgun-cli completion zsh > "${fpath[1]}/_shotgun-cli"

Fish:
  shotgun-cli completion fish | source

  # Or generate to file:
  shotgun-cli completion fish > ~/.config/fish/completions/shotgun-cli.fish

PowerShell:
  # Add to PowerShell profile:
  shotgun-cli completion powershell | Out-String | Invoke-Expression

  # Or save to file and source in profile:
  shotgun-cli completion powershell > shotgun-cli.ps1

After installing completion, restart your shell or source the completion file.`,

	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]

		switch shell {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", shell)
		}
	},
}

// Custom completion functions for dynamic values
func configKeyCompletion(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Only complete the first argument (config key)
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	configKeys := []string{
		"scanner.max-files\tMaximum number of files to scan",
		"scanner.max-file-size\tMaximum size per file (e.g., 1MB)",
		"scanner.respect-gitignore\tRespect .gitignore files (true/false)",
		"context.max-size\tMaximum context size (e.g., 10MB)",
		"context.include-tree\tInclude directory tree (true/false)",
		"context.include-summary\tInclude file summaries (true/false)",
		configKeyTemplateCustomPath + "\tPath to custom templates",
		"output.format\tOutput format (markdown/text)",
		"output.clipboard\tCopy to clipboard (true/false)",
	}

	return configKeys, cobra.ShellCompDirectiveNoFileComp
}

func boolValueCompletion(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Only complete the second argument (config value) for boolean keys
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := args[0]
	boolKeys := []string{
		"scanner.respect-gitignore",
		"context.include-tree",
		"context.include-summary",
		"output.clipboard",
	}

	for _, boolKey := range boolKeys {
		if key == boolKey {
			return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
		}
	}

	// For output.format, provide format options
	if key == "output.format" {
		return []string{"markdown", "text"}, cobra.ShellCompDirectiveNoFileComp
	}

	// For path-based configs, enable file completion
	if key == configKeyTemplateCustomPath {
		return nil, cobra.ShellCompDirectiveDefault
	}

	// For other keys, no specific completion
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	// Register custom completion functions
	if configSetCmd != nil {
		configSetCmd.ValidArgsFunction = func(
			cmd *cobra.Command,
			args []string,
			toComplete string,
		) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return configKeyCompletion(cmd, args, toComplete)
			}
			if len(args) == 1 {
				return boolValueCompletion(cmd, args, toComplete)
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}

	rootCmd.AddCommand(completionCmd)
}