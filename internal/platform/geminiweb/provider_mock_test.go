package geminiweb

import (
	"context"
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebProvider_IsAvailable_BinaryExists tests IsAvailable when binary exists.
func TestWebProvider_IsAvailable_BinaryExists(t *testing.T) {
	mock := NewMockRunnerAvailable()
	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, err := NewWebProviderWithRunner(cfg, mock)
	require.NoError(t, err)

	assert.True(t, provider.IsAvailable())
	assert.True(t, mock.WasLookPathCalled("geminiweb"))
}

// TestWebProvider_IsAvailable_BinaryNotFound tests IsAvailable when binary doesn't exist.
func TestWebProvider_IsAvailable_BinaryNotFound(t *testing.T) {
	mock := NewMockRunner()
	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, err := NewWebProviderWithRunner(cfg, mock)
	require.NoError(t, err)

	assert.False(t, provider.IsAvailable())
	assert.True(t, mock.WasLookPathCalled("geminiweb"))
}

// TestWebProvider_IsConfigured_WithBinary tests IsConfigured when binary is available.
func TestWebProvider_IsConfigured_WithBinary(t *testing.T) {
	mock := NewMockRunnerAvailable()
	cfg := llm.Config{
		Provider: llm.ProviderGeminiWeb,
		Model:    "gemini-2.5-flash",
	}
	provider, err := NewWebProviderWithRunner(cfg, mock)
	require.NoError(t, err)

	// IsConfigured still depends on actual file system check for cookies
	// This test verifies the method can be called without errors
	_ = provider.IsConfigured()
}

// TestWebProvider_IsConfigured_NoBinary tests IsConfigured when binary is not available.
func TestWebProvider_IsConfigured_NoBinary(t *testing.T) {
	mock := NewMockRunner()
	cfg := llm.Config{
		Provider: llm.ProviderGeminiWeb,
		Model:    "gemini-2.5-flash",
	}
	provider, err := NewWebProviderWithRunner(cfg, mock)
	require.NoError(t, err)

	// IsConfigured still depends on actual file system check for cookies
	_ = provider.IsConfigured()
}

// TestWebProvider_Send_BinaryNotFound tests Send when binary doesn't exist.
func TestWebProvider_Send_BinaryNotFound(t *testing.T) {
	mock := NewMockRunner()
	cfg := llm.Config{
		Provider:   llm.ProviderGeminiWeb,
		BinaryPath: "/nonexistent/geminiweb",
	}
	provider, err := NewWebProviderWithRunner(cfg, mock)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = provider.Send(ctx, "test prompt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

// TestWebProvider_SendWithProgress_NotAvailable_Mock tests SendWithProgress when binary not available.
func TestWebProvider_SendWithProgress_NotAvailable_Mock(t *testing.T) {
	mock := NewMockRunner()
	cfg := llm.Config{
		Provider:   llm.ProviderGeminiWeb,
		BinaryPath: "/nonexistent/geminiweb",
	}
	provider, err := NewWebProviderWithRunner(cfg, mock)
	require.NoError(t, err)

	progressCalled := false
	progress := func(stage string) {
		progressCalled = true
	}

	ctx := context.Background()
	_, err = provider.SendWithProgress(ctx, "test prompt", progress)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
	assert.True(t, progressCalled)
}

// TestWebProvider_ValidateConfig_BinaryNotFound tests ValidateConfig when binary not found.
func TestWebProvider_ValidateConfig_BinaryNotFound(t *testing.T) {
	mock := NewMockRunner()
	cfg := llm.Config{
		Provider:   llm.ProviderGeminiWeb,
		BinaryPath: "/nonexistent/geminiweb",
	}
	provider, err := NewWebProviderWithRunner(cfg, mock)
	require.NoError(t, err)

	err = provider.ValidateConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "binary not found")
}

// TestWebProvider_ValidateConfig_BinaryExists tests ValidateConfig when binary exists.
func TestWebProvider_ValidateConfig_BinaryExists(t *testing.T) {
	mock := NewMockRunnerAvailable()
	cfg := llm.Config{
		Provider: llm.ProviderGeminiWeb,
	}
	provider, err := NewWebProviderWithRunner(cfg, mock)
	require.NoError(t, err)

	err = provider.ValidateConfig()

	// If cookies exist (IsConfigured returns true), validation passes
	// Otherwise, it should return configuration error
	if IsConfigured() {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	}
}

// TestWebProvider_Name_Mock tests the Name method with mock runner.
func TestWebProvider_Name_Mock(t *testing.T) {
	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, err := NewWebProvider(cfg)
	require.NoError(t, err)

	assert.Equal(t, "GeminiWeb", provider.Name())
}

// TestWebProvider_Send_ContextTimeout_Mock tests context cancellation with mock.
func TestWebProvider_Send_ContextTimeout_Mock(t *testing.T) {
	mock := NewMockRunnerAvailable()
	cfg := llm.Config{
		Provider: llm.ProviderGeminiWeb,
		Timeout:  1,
	}
	provider, err := NewWebProviderWithRunner(cfg, mock)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = provider.Send(ctx, "test prompt")

	// Should get context canceled error
	assert.Error(t, err)
}

// TestWebProvider_SendWithProgress_CallbackOrder_Mock tests that progress callbacks are called in order.
func TestWebProvider_SendWithProgress_CallbackOrder_Mock(t *testing.T) {
	mock := NewMockRunnerAvailable()
	cfg := llm.Config{
		Provider: llm.ProviderGeminiWeb,
	}
	provider, err := NewWebProviderWithRunner(cfg, mock)
	require.NoError(t, err)

	stages := []string{}
	progress := func(stage string) {
		stages = append(stages, stage)
	}

	ctx := context.Background()
	_, err = provider.SendWithProgress(ctx, "test prompt", progress)

	// Progress is called even when binary fails to execute
	assert.Greater(t, len(stages), 0, "progress callback should be called")

	// Should have "Locating" as first stage
	if len(stages) > 0 {
		assert.Contains(t, stages[0], "Locating")
	}

	// Should get an error since mock doesn't actually execute
	assert.Error(t, err)
}

// TestNewWebProviderWithRunner tests creating provider with custom runner.
func TestNewWebProviderWithRunner(t *testing.T) {
	mock := NewMockRunnerAvailable()
	cfg := llm.Config{
		Provider: llm.ProviderGeminiWeb,
		Model:    "gemini-3.0-pro",
	}

	provider, err := NewWebProviderWithRunner(cfg, mock)

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.executor)
	assert.Equal(t, mock, provider.runner)
	assert.Equal(t, "gemini-3.0-pro", provider.config.Model)
}

// TestMockCommandRunner_LookPath tests the mock LookPath implementation.
func TestMockCommandRunner_LookPath(t *testing.T) {
	t.Run("BinaryExists", func(t *testing.T) {
		mock := NewMockRunnerAvailable()
		path, err := mock.LookPath("geminiweb")
		assert.NoError(t, err)
		assert.Equal(t, "/usr/bin/geminiweb", path)
	})

	t.Run("BinaryNotFound", func(t *testing.T) {
		mock := NewMockRunner()
		_, err := mock.LookPath("geminiweb")
		assert.Error(t, err)
	})

	t.Run("CustomPath", func(t *testing.T) {
		mock := NewMockRunner()
		mock.SetLookPathFunc(func(file string) (string, error) {
			return "/custom/path/" + file, nil
		})
		path, err := mock.LookPath("geminiweb")
		assert.NoError(t, err)
		assert.Equal(t, "/custom/path/geminiweb", path)
	})
}

// TestMockCommandRunner_CommandContext tests the mock CommandContext implementation.
func TestMockCommandRunner_CommandContext(t *testing.T) {
	mock := NewMockRunnerAvailable()
	ctx := context.Background()

	cmd := mock.CommandContext(ctx, "geminiweb", "-m", "test-model")

	assert.NotNil(t, cmd)

	calls := mock.GetCommandCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "geminiweb", calls[0].Name)
	assert.Equal(t, []string{"-m", "test-model"}, calls[0].Args)
}

// TestMockCommandRunner_VerificationMethods tests mock verification methods.
func TestMockCommandRunner_VerificationMethods(t *testing.T) {
	mock := NewMockRunnerAvailable()

	// Test LookPath calls
	_, _ = mock.LookPath("geminiweb")
	assert.True(t, mock.WasLookPathCalled("geminiweb"))
	assert.False(t, mock.WasLookPathCalled("other"))

	// Test CommandContext calls
	ctx := context.Background()
	_ = mock.CommandContext(ctx, "geminiweb", "-arg")
	assert.True(t, mock.WasCommandCalled("geminiweb"))
	assert.False(t, mock.WasCommandCalled("other"))

	// Test reset
	mock.Reset()
	assert.False(t, mock.WasLookPathCalled("geminiweb"))
	assert.False(t, mock.WasCommandCalled("geminiweb"))
}

// TestMockCommandRunner_ConcurrentCalls tests that mock is thread-safe.
func TestMockCommandRunner_ConcurrentCalls(t *testing.T) {
	mock := NewMockRunnerAvailable()

	// Make concurrent calls
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = mock.LookPath("geminiweb")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	calls := mock.GetLookPathCalls()
	assert.Len(t, calls, 10)
}
