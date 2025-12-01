package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestConfigKeyCompletion(t *testing.T) {
	t.Run("no args - return all config keys", func(t *testing.T) {
		results, directive := configKeyCompletion(nil, nil, "")

		assert.NotEmpty(t, results)
		assert.Contains(t, results, "scanner.max-files\tMaximum number of files to scan")
		assert.Contains(t, results, "scanner.max-file-size\tMaximum size per file (e.g., 1MB)")
		assert.Contains(t, results, "scanner.respect-gitignore\tRespect .gitignore files (true/false)")
		assert.Contains(t, results, "scanner.skip-binary\tSkip binary files (true/false)")
		assert.Contains(t, results, "context.max-size\tMaximum context size (e.g., 10MB)")
		assert.Contains(t, results, "context.include-tree\tInclude directory tree (true/false)")
		assert.Contains(t, results, "context.include-summary\tInclude file summaries (true/false)")
		assert.Contains(t, results, "template.custom-path\tPath to custom templates")
		assert.Contains(t, results, "output.format\tOutput format (markdown/text)")
		assert.Contains(t, results, "output.clipboard\tCopy to clipboard (true/false)")
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("with args - no completion", func(t *testing.T) {
		results, directive := configKeyCompletion(nil, []string{"scanner.max-files"}, "")

		assert.Empty(t, results)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("with multiple args - no completion", func(t *testing.T) {
		results, directive := configKeyCompletion(nil, []string{"arg1", "arg2", "arg3"}, "")

		assert.Empty(t, results)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})
}

func TestBoolValueCompletion(t *testing.T) {
	t.Run("boolean keys - return true/false", func(t *testing.T) {
		booleanKeys := []string{
			"scanner.respect-gitignore",
			"scanner.skip-binary",
			"context.include-tree",
			"context.include-summary",
			"output.clipboard",
		}

		for _, key := range booleanKeys {
			t.Run("key: "+key, func(t *testing.T) {
				results, directive := boolValueCompletion(nil, []string{key}, "")

				assert.Equal(t, []string{"true", "false"}, results)
				assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
			})
		}
	})

	t.Run("output.format - return format options", func(t *testing.T) {
		results, directive := boolValueCompletion(nil, []string{"output.format"}, "")

		assert.Equal(t, []string{"markdown", "text"}, results)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("template.custom-path - enable file completion", func(t *testing.T) {
		results, directive := boolValueCompletion(nil, []string{"template.custom-path"}, "")

		assert.Empty(t, results)
		assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
	})

	t.Run("invalid number of args - no completion", func(t *testing.T) {
		results, directive := boolValueCompletion(nil, nil, "")

		assert.Empty(t, results)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("too many args - no completion", func(t *testing.T) {
		results, directive := boolValueCompletion(nil, []string{"key", "value", "extra"}, "")

		assert.Empty(t, results)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("non-boolean/non-special key - no completion", func(t *testing.T) {
		results, directive := boolValueCompletion(nil, []string{"scanner.max-files"}, "")

		assert.Empty(t, results)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("empty key - no completion", func(t *testing.T) {
		results, directive := boolValueCompletion(nil, []string{""}, "")

		assert.Empty(t, results)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})
}

func TestCompletionCommand(t *testing.T) {
	t.Run("bash completion", func(t *testing.T) {
		cmd := &cobra.Command{Use: "shotgun-cli"}
		cmd.AddCommand(completionCmd)

		// Just verify the command is valid and has correct structure
		assert.Equal(t, "completion [bash|zsh|fish|powershell]", completionCmd.Use)
		assert.Equal(t, "Generate completion script", completionCmd.Short)
		assert.Contains(t, completionCmd.Long, "Generate shell completion scripts")
		assert.Equal(t, []string{"bash", "zsh", "fish", "powershell"}, completionCmd.ValidArgs)
	})

	t.Run("valid shells", func(t *testing.T) {
		validShells := []string{"bash", "zsh", "fish", "powershell"}
		for _, shell := range validShells {
			t.Run("shell: "+shell, func(t *testing.T) {
				// Verify each shell is in ValidArgs
				assert.Contains(t, completionCmd.ValidArgs, shell)
			})
		}
	})
}
