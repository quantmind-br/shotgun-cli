package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantmind-br/shotgun-cli/internal/config"
)

// TestConfigKeyCompletion_MatchesRegistry is the load-bearing assertion of
// CI-009: the completion output must be exactly the metadata registry's key set.
// Asserting literal key strings here would just relocate the duplication the
// item removes, so the expectation is derived from the registry instead.
func TestConfigKeyCompletion_MatchesRegistry(t *testing.T) {
	t.Parallel()

	results, directive := configKeyCompletion(nil, nil, "")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

	metadata := config.AllConfigMetadata()
	require.Len(t, results, len(metadata), "completion must offer every registered key and nothing else")

	gotKeys := make([]string, 0, len(results))
	for _, r := range results {
		key, description, found := strings.Cut(r, "\t")
		require.True(t, found, "completion entry %q must carry a tab-separated description", r)
		assert.NotEmpty(t, description, "key %q completes with an empty description", key)
		gotKeys = append(gotKeys, key)
	}

	wantKeys := make([]string, 0, len(metadata))
	for _, m := range metadata {
		wantKeys = append(wantKeys, m.Key)
	}

	assert.ElementsMatch(t, wantKeys, gotKeys)
}

// TestConfigKeyCompletion_MatchesValidKeys pins the other half of the contract:
// a completable key must also be settable. These drifted before -- the hardcoded
// completion list was missing all six llm.* keys.
func TestConfigKeyCompletion_MatchesValidKeys(t *testing.T) {
	t.Parallel()

	results, _ := configKeyCompletion(nil, nil, "")
	for _, r := range results {
		key, _, _ := strings.Cut(r, "\t")
		assert.True(t, config.IsValidKey(key), "completed key %q is rejected by IsValidKey", key)
	}
}

func TestConfigKeyCompletion(t *testing.T) {
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

// TestBoolValueCompletion_DrivenByType walks every registered key and asserts
// the completion matches its declared type, so a new bool or enum key is
// completable with no edit here.
func TestBoolValueCompletion_DrivenByType(t *testing.T) {
	t.Parallel()

	for _, m := range config.AllConfigMetadata() {
		t.Run(m.Key, func(t *testing.T) {
			t.Parallel()
			results, directive := boolValueCompletion(nil, []string{m.Key}, "")

			switch m.Type {
			case config.TypeBool:
				assert.Equal(t, []string{"true", "false"}, results)
				assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
			case config.TypeEnum:
				require.NotEmpty(t, m.EnumOptions, "enum key %q declares no options", m.Key)
				assert.Equal(t, m.EnumOptions, results)
				assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
			case config.TypePath:
				assert.Empty(t, results)
				assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
			default:
				assert.Empty(t, results)
				assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
			}
		})
	}
}

func TestBoolValueCompletion(t *testing.T) {
	t.Run("output.format - return format options", func(t *testing.T) {
		results, directive := boolValueCompletion(nil, []string{"output.format"}, "")

		assert.Equal(t, []string{"markdown", "text"}, results)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("llm.provider - return provider options", func(t *testing.T) {
		results, directive := boolValueCompletion(nil, []string{"llm.provider"}, "")

		assert.Equal(t, []string{"openai", "anthropic", "gemini"}, results)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("unknown key - no completion", func(t *testing.T) {
		results, directive := boolValueCompletion(nil, []string{"nonsense.key"}, "")

		assert.Empty(t, results)
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
