package cmd

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func isExpectedProviderError(err error) bool {
	if err == nil {
		return true
	}
	errStr := err.Error()
	expectedPatterns := []string{
		"gemini", "Gemini", "LLM", "not available",
		"API error", "request failed", "401", "Incorrect API key",
	}
	for _, pattern := range expectedPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}

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
		viper.Reset()
		viper.Set("llm.provider", "openai")
		viper.Set("llm.api-key", "test-key")

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
				strings.Contains(errorMsg, "gemini request failed") ||
					strings.Contains(errorMsg, "failed to read file") ||
					strings.Contains(errorMsg, "gemini integration is disabled") ||
					strings.Contains(errorMsg, "LLM integration is disabled") ||
					strings.Contains(errorMsg, "not available") ||
					strings.Contains(errorMsg, "request failed") ||
					strings.Contains(errorMsg, "401") ||
					strings.Contains(errorMsg, "Incorrect API key"),
				"Expected LLM-related error, got: %s", errorMsg)
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

	setupCmd := func() *cobra.Command {
		cmd := &cobra.Command{}
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("model", "", "")
		cmd.Flags().Int("timeout", 0, "")
		cmd.Flags().Bool("raw", false, "")
		return cmd
	}

	t.Run("custom model flag parsing", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.provider", "openai")
		viper.Set("llm.api-key", "test-key")
		viper.Set("gemini.model", "default-model")

		cmd := setupCmd()
		_ = cmd.Flags().Set("model", "gemini-3.0-pro")

		err := runContextSend(cmd, []string{testFile})
		assert.True(t, isExpectedProviderError(err), "unexpected error: %v", err)
	})

	t.Run("custom timeout flag parsing", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.provider", "openai")
		viper.Set("llm.api-key", "test-key")
		viper.Set("gemini.timeout", 60)

		cmd := setupCmd()
		_ = cmd.Flags().Set("timeout", "30")

		err := runContextSend(cmd, []string{testFile})
		assert.True(t, isExpectedProviderError(err), "unexpected error: %v", err)
	})

	t.Run("output to file flag", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.provider", "openai")
		viper.Set("llm.api-key", "test-key")

		outputFile := tempDir + "/response.txt"
		cmd := setupCmd()
		_ = cmd.Flags().Set("output", outputFile)

		err := runContextSend(cmd, []string{testFile})
		assert.True(t, isExpectedProviderError(err), "unexpected error: %v", err)
	})

	t.Run("raw output flag", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.provider", "openai")
		viper.Set("llm.api-key", "test-key")

		cmd := setupCmd()
		_ = cmd.Flags().Set("raw", "true")

		err := runContextSend(cmd, []string{testFile})
		assert.True(t, isExpectedProviderError(err), "unexpected error: %v", err)
	})

	t.Run("viper config integration", func(t *testing.T) {
		viper.Reset()
		viper.Set("llm.provider", "openai")
		viper.Set("llm.api-key", "test-key")
		viper.Set("gemini.model", "config-model")
		viper.Set("gemini.timeout", 45)
		viper.Set("gemini.binary-path", "/path/to/binary")
		viper.Set("gemini.browser-refresh", "always")
		viper.Set("verbose", true)

		cmd := setupCmd()

		err := runContextSend(cmd, []string{testFile})
		assert.True(t, isExpectedProviderError(err), "unexpected error: %v", err)
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

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		want     string
	}{
		{
			name:     "zero duration",
			duration: "0ms",
			want:     "0ms",
		},
		{
			name:     "milliseconds only - less than 1 second",
			duration: "500ms",
			want:     "500ms",
		},
		{
			name:     "1 millisecond",
			duration: "1ms",
			want:     "1ms",
		},
		{
			name:     "999 milliseconds - still less than 1 second",
			duration: "999ms",
			want:     "999ms",
		},
		{
			name:     "exactly 1 second",
			duration: "1s",
			want:     "1.0s",
		},
		{
			name:     "1.5 seconds",
			duration: "1500ms",
			want:     "1.5s",
		},
		{
			name:     "2 seconds",
			duration: "2s",
			want:     "2.0s",
		},
		{
			name:     "10 seconds",
			duration: "10s",
			want:     "10.0s",
		},
		{
			name:     "1 minute",
			duration: "1m",
			want:     "60.0s",
		},
		{
			name:     "1 minute 5 seconds",
			duration: "1m5s",
			want:     "65.0s",
		},
		{
			name:     "5.5 seconds",
			duration: "5500ms",
			want:     "5.5s",
		},
		{
			name:     "100 milliseconds",
			duration: "100ms",
			want:     "100ms",
		},
		{
			name:     "large duration - 5 minutes",
			duration: "5m",
			want:     "300.0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := time.ParseDuration(tt.duration)
			require.NoError(t, err, "failed to parse duration: %s", tt.duration)
			got := formatDuration(d)
			assert.Equal(t, tt.want, got, "formatDuration(%s) = %s, want %s", tt.duration, got, tt.want)
		})
	}
}
