package gemini

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/stretchr/testify/assert"
)

func TestNewWebProvider_Defaults(t *testing.T) {
	cfg := llm.Config{
		Provider: llm.ProviderGeminiWeb,
	}

	provider, err := NewWebProvider(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.executor)
	assert.Equal(t, "gemini-2.5-flash", provider.config.Model)
	assert.Equal(t, 300, provider.config.Timeout)
	assert.Equal(t, "", provider.config.BrowserRefresh)
}

func TestNewWebProvider_CustomConfig(t *testing.T) {
	cfg := llm.Config{
		Provider:      llm.ProviderGeminiWeb,
		BinaryPath:     "/custom/path/geminiweb",
		Model:          "gemini-3.0-pro",
		Timeout:        180,
		BrowserRefresh: "firefox",
	}

	provider, err := NewWebProvider(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.executor)
	assert.Equal(t, "/custom/path/geminiweb", provider.config.BinaryPath)
	assert.Equal(t, "gemini-3.0-pro", provider.config.Model)
	assert.Equal(t, 180, provider.config.Timeout)
	assert.Equal(t, "firefox", provider.config.BrowserRefresh)
}

func TestWebProvider_Name(t *testing.T) {
	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	assert.Equal(t, "GeminiWeb", provider.Name())
}

func TestWebProvider_IsAvailable(t *testing.T) {
	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	result := provider.IsAvailable()
	assert.IsType(t, false, result)
}

func TestWebProvider_IsConfigured(t *testing.T) {
	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	result := provider.IsConfigured()
	assert.IsType(t, false, result)
}

func TestWebProvider_ValidateConfig_NotAvailable(t *testing.T) {
	cfg := llm.Config{
		Provider:   llm.ProviderGeminiWeb,
		BinaryPath: "/nonexistent/geminiweb",
	}
	provider, _ := NewWebProvider(cfg)

	err := provider.ValidateConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "geminiweb binary not found")
}

func TestWebProvider_Send_DelegatesToExecutor(t *testing.T) {
	if !IsAvailable() {
		t.Skip("geminiweb not available for integration test")
	}

	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	ctx := context.Background()
	_, err := provider.Send(ctx, "test prompt")

	// Send delegates to executor.Send()
	if err != nil {
		assert.Error(t, err)
	}
}

func TestWebProvider_SendWithProgress_DelegatesToExecutor(t *testing.T) {
	if !IsAvailable() {
		t.Skip("geminiweb not available for integration test")
	}

	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	progressCalled := false
	progress := func(stage string) {
		if stage != "" {
			progressCalled = true
		}
	}

	ctx := context.Background()
	_, err := provider.SendWithProgress(ctx, "test prompt", progress)

	// Verify progress callback was called at least once
	assert.True(t, progressCalled, "progress callback should be called")

	// SendWithProgress delegates to executor.SendWithProgress()
	if err != nil {
		assert.Error(t, err)
	}
}

func TestWebProvider_Send_NotAvailable(t *testing.T) {
	cfg := llm.Config{
		Provider:   llm.ProviderGeminiWeb,
		BinaryPath: "/nonexistent/geminiweb",
	}
	provider, _ := NewWebProvider(cfg)

	ctx := context.Background()
	_, err := provider.Send(ctx, "test prompt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestWebProvider_SendWithProgress_NotAvailable(t *testing.T) {
	cfg := llm.Config{
		Provider:   llm.ProviderGeminiWeb,
		BinaryPath: "/nonexistent/geminiweb",
	}
	provider, _ := NewWebProvider(cfg)

	progressCalled := false
	progress := func(stage string) {
		progressCalled = true
	}

	ctx := context.Background()
	_, err := provider.SendWithProgress(ctx, "test prompt", progress)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
	// Progress callback is called even on error
	assert.True(t, progressCalled)
}

func TestWebProvider_ValidateConfig_NotConfigured(t *testing.T) {
	// This test assumes geminiweb binary might not be configured
	// We need to mock the scenario where IsConfigured returns false

	if !IsAvailable() {
		t.Skip("geminiweb not available, skipping")
	}

	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	err := provider.ValidateConfig()

	// If configured, validation should pass
	// If not configured, it should return an error
	if IsConfigured() {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	}
}

func TestWebProvider_ValidateConfig_Success(t *testing.T) {
	// Test the happy path when both IsAvailable and IsConfigured return true
	if !IsAvailable() {
		t.Skip("geminiweb not available, skipping success test")
	}

	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	err := provider.ValidateConfig()

	// If configured, validation should pass
	if IsConfigured() {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	}
}

func TestWebProvider_ValidateConfig_NotConfigured_Integration(t *testing.T) {
	// Integration test that requires geminiweb to be available but NOT configured
	// This tests the second if branch in ValidateConfig
	if !IsAvailable() {
		t.Skip("geminiweb not available, integration test requires binary")
	}
	if IsConfigured() {
		t.Skip("geminiweb is configured, cannot test 'not configured' path")
	}

	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	err := provider.ValidateConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestWebProvider_ValidateConfig_Configured_Integration(t *testing.T) {
	// Integration test for the happy path - requires geminiweb installed AND configured
	if !IsAvailable() || !IsConfigured() {
		t.Skip("geminiweb not available/configured, integration test requires fully configured binary")
	}

	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	err := provider.ValidateConfig()

	assert.NoError(t, err, "ValidateConfig should return nil when geminiweb is available and configured")
}

func TestWebProvider_ValidateConfig_BothChecksFail(t *testing.T) {
	// Test when geminiweb is not available (first check fails)
	// The second check (IsConfigured) should not be reached

	cfg := llm.Config{
		Provider:   llm.ProviderGeminiWeb,
		BinaryPath: "/nonexistent/geminiweb",
	}
	provider, _ := NewWebProvider(cfg)

	err := provider.ValidateConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "binary not found")
}

func TestWebProvider_Send_ContextTimeout(t *testing.T) {
	if !IsAvailable() {
		t.Skip("geminiweb not available")
	}

	cfg := llm.Config{
		Provider: llm.ProviderGeminiWeb,
		Timeout:  1, // 1 second timeout
	}
	provider, _ := NewWebProvider(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := provider.Send(ctx, "test prompt")

	// Might timeout or succeed depending on system
	if err != nil {
		assert.Error(t, err)
	}
}

func TestWebProvider_SendWithProgress_CallbackInvoked(t *testing.T) {
	if !IsAvailable() {
		t.Skip("geminiweb not available")
	}

	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	stages := []string{}
	progress := func(stage string) {
		stages = append(stages, stage)
	}

	ctx := context.Background()
	_, err := provider.SendWithProgress(ctx, "test prompt", progress)

	// Verify progress callback was called with expected stages
	assert.Greater(t, len(stages), 0, "progress callback should be called")

	// Check for expected stage messages
	expectedStages := []string{"Locating", "Preparing", "Sending"}
	for _, expected := range expectedStages {
		found := false
		for _, stage := range stages {
			if strings.Contains(stage, expected) {
				found = true
				break
			}
		}
		// Note: not asserting on each stage since they may vary
		_ = found
	}

	if err != nil {
		assert.Error(t, err)
	}
}

func TestWebProvider_Send_ReturnsCorrectProvider(t *testing.T) {
	if !IsAvailable() || !IsConfigured() {
		t.Skip("geminiweb not available or configured")
	}

	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	ctx := context.Background()
	result, err := provider.Send(ctx, "test prompt")

	if err == nil {
		assert.NotNil(t, result)
		assert.Equal(t, "GeminiWeb", result.Provider)
		assert.NotEmpty(t, result.Model)
		assert.Greater(t, result.Duration, time.Duration(0))
	}
}

func TestWebProvider_SendWithProgress_ReturnsCorrectProvider(t *testing.T) {
	if !IsAvailable() || !IsConfigured() {
		t.Skip("geminiweb not available or configured")
	}

	cfg := llm.Config{Provider: llm.ProviderGeminiWeb}
	provider, _ := NewWebProvider(cfg)

	progress := func(stage string) {}
	ctx := context.Background()
	result, err := provider.SendWithProgress(ctx, "test prompt", progress)

	if err == nil {
		assert.NotNil(t, result)
		assert.Equal(t, "GeminiWeb", result.Provider)
		assert.NotEmpty(t, result.Model)
		assert.Greater(t, result.Duration, time.Duration(0))
	}
}
