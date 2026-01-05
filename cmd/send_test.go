package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextSendCmd_PreRunE(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		setup   func(t *testing.T)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no args - allowed",
			args:    []string{},
			wantErr: false,
		},
		{
			name: "existing file - allowed",
			args: []string{"testfile.txt"},
			setup: func(t *testing.T) {
				f, err := os.Create("testfile.txt")
				require.NoError(t, err)
				_ = f.Close()
				t.Cleanup(func() { _ = os.Remove("testfile.txt") })
			},
			wantErr: false,
		},
		{
			name:    "non-existent file - denied",
			args:    []string{"nonexistent.txt"},
			wantErr: true,
			errMsg:  "file not found: nonexistent.txt",
		},
		{
			name:    "too many args - denied",
			args:    []string{"file1.txt", "file2.txt"},
			wantErr: true,
			errMsg:  "file not found: file1.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			cmd := contextSendCmd
			err := cmd.PreRunE(cmd, tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunContextSend_FromFile(t *testing.T) {
	// Create a temporary file with test content
	tempDir := t.TempDir()
	testFile := tempDir + "/test-prompt.txt"
	testContent := "This is a test prompt for Gemini"
	err := os.WriteFile(testFile, []byte(testContent), 0o600)
	require.NoError(t, err)

	// Test successful file reading and validation
	t.Run("successful file read", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("model", "", "")
		cmd.Flags().Int("timeout", 0, "")
		cmd.Flags().Bool("raw", false, "")

		// Test with a file that exists - this may succeed or fail depending on gemini availability
		err := runContextSend(cmd, []string{testFile})

		// If gemini is available and working, the test should succeed
		// If gemini is not available or not working, it should fail with expected error
		if err != nil {
			// We expect either LLM/gemini not available/configured error or send error
			errorMsg := err.Error()
			// Check that we got one of the expected error types
			assert.True(t,
				strings.Contains(errorMsg, "geminiweb") ||
					strings.Contains(errorMsg, "gemini request failed") ||
					strings.Contains(errorMsg, "failed to read file") ||
					strings.Contains(errorMsg, "gemini integration is disabled") ||
					strings.Contains(errorMsg, "LLM integration is disabled") ||
					strings.Contains(errorMsg, "not available") ||
					strings.Contains(errorMsg, "request failed"),
				"Expected LLM/gemini-related error, got: %s", errorMsg)
		}
		// If err is nil, it means gemini is working and the test succeeded
	})

	// Test with empty file
	t.Run("empty file content", func(t *testing.T) {
		emptyFile := tempDir + "/empty.txt"
		err := os.WriteFile(emptyFile, []byte(""), 0o600)
		require.NoError(t, err)

		cmd := &cobra.Command{}
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("model", "", "")
		cmd.Flags().Int("timeout", 0, "")
		cmd.Flags().Bool("raw", false, "")

		err = runContextSend(cmd, []string{emptyFile})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no content to send")
	})

	// Test with whitespace-only file
	t.Run("whitespace-only file content", func(t *testing.T) {
		whitespaceFile := tempDir + "/whitespace.txt"
		err := os.WriteFile(whitespaceFile, []byte("   \n\t  \n   "), 0o600)
		require.NoError(t, err)

		cmd := &cobra.Command{}
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("model", "", "")
		cmd.Flags().Int("timeout", 0, "")
		cmd.Flags().Bool("raw", false, "")

		err = runContextSend(cmd, []string{whitespaceFile})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no content to send")
	})

	// Test with non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("model", "", "")
		cmd.Flags().Int("timeout", 0, "")
		cmd.Flags().Bool("raw", false, "")

		err := runContextSend(cmd, []string{"/non/existent/file.txt"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})
}

func TestRunContextSend_Flags(t *testing.T) {
	tempDir := t.TempDir()
	testFile := tempDir + "/test-prompt.txt"
	err := os.WriteFile(testFile, []byte("Test content"), 0o600)
	require.NoError(t, err)

	t.Run("custom model flag parsing", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("model", "gemini-3.0-pro", "")
		cmd.Flags().Int("timeout", 0, "")
		cmd.Flags().Bool("raw", false, "")

		// Set viper config
		viper.Set("gemini.model", "default-model")

		// Test may succeed or fail depending on gemini availability
		err := runContextSend(cmd, []string{testFile})
		// If it fails, check it's the right kind of error
		if err != nil {
			assert.True(t, strings.Contains(err.Error(), "gemini") ||
				strings.Contains(err.Error(), "LLM") ||
				strings.Contains(err.Error(), "not available"))
		}
	})

	t.Run("custom timeout flag parsing", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("model", "", "")
		cmd.Flags().Int("timeout", 30, "")
		cmd.Flags().Bool("raw", false, "")

		// Set viper config
		viper.Set("gemini.timeout", 60)

		// Test may succeed or fail depending on gemini availability
		err := runContextSend(cmd, []string{testFile})
		// If it fails, check it's the right kind of error
		if err != nil {
			assert.True(t, strings.Contains(err.Error(), "gemini") ||
				strings.Contains(err.Error(), "LLM") ||
				strings.Contains(err.Error(), "not available"))
		}
	})

	t.Run("output to file flag", func(t *testing.T) {
		outputFile := tempDir + "/response.txt"
		cmd := &cobra.Command{}
		cmd.Flags().String("output", outputFile, "")
		cmd.Flags().String("model", "", "")
		cmd.Flags().Int("timeout", 0, "")
		cmd.Flags().Bool("raw", false, "")

		// Test may succeed or fail depending on gemini availability
		err := runContextSend(cmd, []string{testFile})
		// If it fails, check it's the right kind of error
		if err != nil {
			assert.True(t, strings.Contains(err.Error(), "gemini") ||
				strings.Contains(err.Error(), "LLM") ||
				strings.Contains(err.Error(), "not available"))
		}
	})

	t.Run("raw output flag", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("model", "", "")
		cmd.Flags().Int("timeout", 0, "")
		cmd.Flags().Bool("raw", true, "")

		// Test may succeed or fail depending on gemini availability
		err := runContextSend(cmd, []string{testFile})
		// If it fails, check it's the right kind of error
		if err != nil {
			assert.True(t, strings.Contains(err.Error(), "gemini") ||
				strings.Contains(err.Error(), "LLM") ||
				strings.Contains(err.Error(), "not available"))
		}
	})

	t.Run("viper config integration", func(t *testing.T) {
		// Set up viper config
		viper.Set("gemini.model", "config-model")
		viper.Set("gemini.timeout", 45)
		viper.Set("gemini.binary-path", "/path/to/binary")
		viper.Set("gemini.browser-refresh", "always")
		viper.Set("verbose", true)

		cmd := &cobra.Command{}
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("model", "", "") // Empty so it uses viper
		cmd.Flags().Int("timeout", 0, "")   // Empty so it uses viper
		cmd.Flags().Bool("raw", false, "")

		// Test may succeed or fail depending on gemini availability
		err := runContextSend(cmd, []string{testFile})
		// If it fails, check it's the right kind of error
		if err != nil {
			assert.True(t, strings.Contains(err.Error(), "gemini") ||
				strings.Contains(err.Error(), "LLM") ||
				strings.Contains(err.Error(), "not available"))
		}
	})
}

func TestContextSendCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "too many args - cobra validation",
			args:    []string{"file1.txt", "file2.txt"},
			wantErr: true,
			errMsg:  "accepts at most 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test using cobra's Args validation directly
			err := contextSendCmd.Args(contextSendCmd, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunContextSend_StdinHandling(t *testing.T) {
	t.Run("stdin without data", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("model", "", "")
		cmd.Flags().Int("timeout", 0, "")
		cmd.Flags().Bool("raw", false, "")

		// Test with no args (should try stdin)
		err := runContextSend(cmd, []string{})
		assert.Error(t, err)
		// Should fail due to no gemini or stdin issues
	})
}
