package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantmind-br/shotgun-cli/internal/app"
	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

// mockLLMProvider is the reason these tests never reach the network.
type mockLLMProvider struct {
	available bool
	result    *llm.Result
	sendErr   error
}

func (m *mockLLMProvider) Name() string          { return "mock" }
func (m *mockLLMProvider) IsAvailable() bool     { return m.available }
func (m *mockLLMProvider) IsConfigured() bool    { return true }
func (m *mockLLMProvider) ValidateConfig() error { return nil }

func (m *mockLLMProvider) Send(ctx context.Context, content string) (*llm.Result, error) {
	return m.result, m.sendErr
}

func (m *mockLLMProvider) SendWithProgress(
	ctx context.Context, content string, progress func(stage string),
) (*llm.Result, error) {
	return m.result, m.sendErr
}

// newMockRegistry occupies the openai slot, so Registry.Create never builds a real client.
func newMockRegistry(p llm.Provider) *llm.Registry {
	r := llm.NewRegistry()
	r.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) { return p, nil })
	return r
}

// installSendService swaps the package-level seam and restores it on cleanup.
func installSendService(t *testing.T, registry *llm.Registry) {
	t.Helper()
	prev := newSendService
	t.Cleanup(func() { newSendService = prev })
	newSendService = func() app.ContextService {
		return app.NewContextService(app.WithRegistry(registry))
	}
}

// setViper restores the prior value on cleanup, so sibling tests keep their state.
func setViper(t *testing.T, key string, value interface{}) {
	t.Helper()
	prev := viper.Get(key)
	t.Cleanup(func() { viper.Set(key, prev) })
	viper.Set(key, value)
}

func newSendCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("model", "", "")
	cmd.Flags().Int("timeout", 0, "")
	cmd.Flags().Bool("raw", false, "")
	return cmd
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
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-prompt.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("This is a test prompt"), 0o600))

	newMock := func() *mockLLMProvider {
		return &mockLLMProvider{
			available: true,
			result:    &llm.Result{Response: "processed", RawResponse: `{"raw":"payload"}`},
		}
	}
	arrange := func(t *testing.T, mock *mockLLMProvider) {
		t.Helper()
		installSendService(t, newMockRegistry(mock))
		setViper(t, config.KeyLLMProvider, "openai")
		setViper(t, config.KeyLLMAPIKey, "test-key")
		setViper(t, config.KeyLLMSaveResponse, false)
	}

	t.Run("writes the response to --output", func(t *testing.T) {
		arrange(t, newMock())
		outputFile := filepath.Join(t.TempDir(), "out.md")
		cmd := newSendCmd(t)
		require.NoError(t, cmd.Flags().Set("output", outputFile))

		require.NoError(t, runContextSend(cmd, []string{testFile}))

		written, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Equal(t, "processed", string(written))

		info, err := os.Stat(outputFile)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("--raw writes RawResponse instead of Response", func(t *testing.T) {
		arrange(t, newMock())
		outputFile := filepath.Join(t.TempDir(), "raw.md")
		cmd := newSendCmd(t)
		require.NoError(t, cmd.Flags().Set("output", outputFile))
		require.NoError(t, cmd.Flags().Set("raw", "true"))

		require.NoError(t, runContextSend(cmd, []string{testFile}))

		written, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Equal(t, `{"raw":"payload"}`, string(written))
		assert.NotContains(t, string(written), "processed")

		info, err := os.Stat(outputFile)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("generates a timestamped filename when save-response is on", func(t *testing.T) {
		t.Chdir(t.TempDir())
		arrange(t, newMock())
		setViper(t, config.KeyLLMSaveResponse, true)

		require.NoError(t, runContextSend(newSendCmd(t), []string{testFile}))

		matches, err := filepath.Glob("llm-response-*.md")
		require.NoError(t, err)
		require.Len(t, matches, 1)
		assert.Regexp(t, `^llm-response-\d{8}-\d{6}\.md$`, matches[0])

		written, readErr := os.ReadFile(matches[0])
		require.NoError(t, readErr)
		assert.Equal(t, "processed", string(written))
	})

	t.Run("prints to stdout when there is nothing to save", func(t *testing.T) {
		t.Chdir(t.TempDir())
		arrange(t, newMock())

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runContextSend(newSendCmd(t), []string{testFile})

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "processed")

		matches, globErr := filepath.Glob("llm-response-*.md")
		require.NoError(t, globErr)
		assert.Empty(t, matches)
	})

	t.Run("fails when the provider is unavailable", func(t *testing.T) {
		mock := newMock()
		mock.available = false
		arrange(t, mock)

		err := runContextSend(newSendCmd(t), []string{testFile})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("fails when the send fails", func(t *testing.T) {
		mock := newMock()
		mock.result = nil
		mock.sendErr = errors.New("connection reset")
		arrange(t, mock)

		err := runContextSend(newSendCmd(t), []string{testFile})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "LLM request failed")
		assert.Contains(t, err.Error(), "connection reset")
	})

	t.Run("rejects an empty file", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty.txt")
		require.NoError(t, os.WriteFile(emptyFile, []byte(""), 0o600))

		err := runContextSend(newSendCmd(t), []string{emptyFile})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no content to send")
	})

	t.Run("rejects a whitespace-only file", func(t *testing.T) {
		whitespaceFile := filepath.Join(tempDir, "whitespace.txt")
		require.NoError(t, os.WriteFile(whitespaceFile, []byte("   \n\t  \n   "), 0o600))

		err := runContextSend(newSendCmd(t), []string{whitespaceFile})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no content to send")
	})

	t.Run("reports an unreadable file", func(t *testing.T) {
		err := runContextSend(newSendCmd(t), []string{"/non/existent/file.txt"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})
}

func TestRunContextSend_Flags(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-prompt.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("Test content"), 0o600))

	// The returned Config is empty until runContextSend reaches the registry.
	arrange := func(t *testing.T) *llm.Config {
		t.Helper()
		got := &llm.Config{}
		mock := &mockLLMProvider{
			available: true,
			result:    &llm.Result{Response: "processed", RawResponse: `{"raw":"payload"}`},
		}
		registry := llm.NewRegistry()
		registry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
			*got = cfg
			return mock, nil
		})
		installSendService(t, registry)
		setViper(t, config.KeyLLMProvider, "openai")
		setViper(t, config.KeyLLMAPIKey, "test-key")
		setViper(t, config.KeyLLMSaveResponse, false)
		return got
	}

	t.Run("model flag overrides the configured model", func(t *testing.T) {
		got := arrange(t)
		setViper(t, config.KeyLLMModel, "config-model")

		cmd := newSendCmd(t)
		require.NoError(t, cmd.Flags().Set("model", "flag-model"))
		require.NoError(t, cmd.Flags().Set("output", filepath.Join(t.TempDir(), "out.md")))

		require.NoError(t, runContextSend(cmd, []string{testFile}))

		assert.Equal(t, "flag-model", got.Model)
	})

	t.Run("timeout flag overrides the configured timeout", func(t *testing.T) {
		got := arrange(t)
		setViper(t, config.KeyLLMTimeout, 60)

		cmd := newSendCmd(t)
		require.NoError(t, cmd.Flags().Set("timeout", "30"))
		require.NoError(t, cmd.Flags().Set("output", filepath.Join(t.TempDir(), "out.md")))

		require.NoError(t, runContextSend(cmd, []string{testFile}))

		assert.Equal(t, 30, got.Timeout)
	})

	t.Run("output flag selects the destination path", func(t *testing.T) {
		arrange(t)
		outputFile := filepath.Join(t.TempDir(), "response.txt")

		cmd := newSendCmd(t)
		require.NoError(t, cmd.Flags().Set("output", outputFile))

		require.NoError(t, runContextSend(cmd, []string{testFile}))

		written, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Equal(t, "processed", string(written))
	})

	t.Run("raw flag selects the raw payload", func(t *testing.T) {
		arrange(t)
		outputFile := filepath.Join(t.TempDir(), "response.txt")

		cmd := newSendCmd(t)
		require.NoError(t, cmd.Flags().Set("output", outputFile))
		require.NoError(t, cmd.Flags().Set("raw", "true"))

		require.NoError(t, runContextSend(cmd, []string{testFile}))

		written, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Equal(t, `{"raw":"payload"}`, string(written))
	})

	t.Run("config values feed through when no flags are set", func(t *testing.T) {
		got := arrange(t)
		setViper(t, config.KeyLLMModel, "config-model")
		setViper(t, config.KeyLLMTimeout, 45)

		cmd := newSendCmd(t)
		require.NoError(t, cmd.Flags().Set("output", filepath.Join(t.TempDir(), "out.md")))

		require.NoError(t, runContextSend(cmd, []string{testFile}))

		assert.Equal(t, "config-model", got.Model)
		assert.Equal(t, 45, got.Timeout)
		assert.Equal(t, llm.ProviderOpenAI, got.Provider)
		assert.Equal(t, "test-key", got.APIKey)
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
