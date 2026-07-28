package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/quantmind-br/shotgun-cli/internal/config"
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

// configKeyCompletion offers every registered configuration key, derived from
// the metadata registry rather than a hand-maintained list. The previous
// hardcoded slice had already drifted: it was missing all six llm.* keys.
func configKeyCompletion(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	// Only complete the first argument (config key)
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	metadata := config.AllConfigMetadata()
	configKeys := make([]string, 0, len(metadata))
	for _, m := range metadata {
		configKeys = append(configKeys, m.Key+"\t"+m.Description)
	}

	return configKeys, cobra.ShellCompDirectiveNoFileComp
}

// boolValueCompletion offers the candidate values for a key, driven by its
// declared type in the metadata registry.
func boolValueCompletion(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	// Only complete the second argument (config value)
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	m, found := config.GetMetadata(args[0])
	if !found {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	switch m.Type {
	case config.TypeBool:
		return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
	case config.TypeEnum:
		return m.EnumOptions, cobra.ShellCompDirectiveNoFileComp
	case config.TypePath:
		// Paths are the one case where the shell's own file completion is better
		// than anything we can enumerate.
		return nil, cobra.ShellCompDirectiveDefault
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
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
