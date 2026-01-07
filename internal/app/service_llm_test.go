package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLLMProvider struct {
	name            string
	available       bool
	configured      bool
	validateErr     error
	sendResult      *llm.Result
	sendErr         error
	progressStages  []string
	sendContentSeen string
}

func (m *mockLLMProvider) Name() string      { return m.name }
func (m *mockLLMProvider) IsAvailable() bool { return m.available }
func (m *mockLLMProvider) IsConfigured() bool {
	return m.configured
}
func (m *mockLLMProvider) ValidateConfig() error { return m.validateErr }
func (m *mockLLMProvider) Send(ctx context.Context, content string) (*llm.Result, error) {
	m.sendContentSeen = content
	return m.sendResult, m.sendErr
}
func (m *mockLLMProvider) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	m.sendContentSeen = content
	if progress != nil {
		for _, stage := range m.progressStages {
			progress(stage)
		}
	}
	return m.sendResult, m.sendErr
}

func newMockRegistry(provider *mockLLMProvider, providerType llm.ProviderType) *llm.Registry {
	registry := llm.NewRegistry()
	registry.Register(providerType, func(cfg llm.Config) (llm.Provider, error) {
		return provider, nil
	})
	return registry
}

func newMockRegistryWithError(err error, providerType llm.ProviderType) *llm.Registry {
	registry := llm.NewRegistry()
	registry.Register(providerType, func(cfg llm.Config) (llm.Provider, error) {
		return nil, err
	})
	return registry
}

func TestSendToLLMWithProgress_Success(t *testing.T) {
	t.Parallel()

	provider := &mockLLMProvider{
		name:      "TestProvider",
		available: true,
		sendResult: &llm.Result{
			Response: "test response",
			Duration: 100 * time.Millisecond,
		},
	}
	registry := newMockRegistry(provider, llm.ProviderOpenAI)
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider: llm.ProviderOpenAI,
		APIKey:   "test-key",
	}

	result, err := svc.SendToLLMWithProgress(context.Background(), "test content", cfg, nil)

	require.NoError(t, err)
	assert.Equal(t, "test response", result.Response)
	assert.Equal(t, "test content", provider.sendContentSeen)
}

func TestSendToLLMWithProgress_WithProgressCallback(t *testing.T) {
	t.Parallel()

	provider := &mockLLMProvider{
		name:           "TestProvider",
		available:      true,
		progressStages: []string{"connecting", "sending", "receiving"},
		sendResult: &llm.Result{
			Response: "response",
		},
	}
	registry := newMockRegistry(provider, llm.ProviderAnthropic)
	svc := NewContextService(WithRegistry(registry))

	var receivedStages []string
	progress := func(stage string) {
		receivedStages = append(receivedStages, stage)
	}

	cfg := LLMSendConfig{
		Provider: llm.ProviderAnthropic,
		APIKey:   "test-key",
	}

	result, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, progress)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, []string{"connecting", "sending", "receiving"}, receivedStages)
}

func TestSendToLLMWithProgress_NilProgressCallback(t *testing.T) {
	t.Parallel()

	provider := &mockLLMProvider{
		name:      "TestProvider",
		available: true,
		sendResult: &llm.Result{
			Response: "response",
		},
	}
	registry := newMockRegistry(provider, llm.ProviderGemini)
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider: llm.ProviderGemini,
	}

	result, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSendToLLMWithProgress_WithSaveResponse(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "response.md")

	provider := &mockLLMProvider{
		name:      "TestProvider",
		available: true,
		sendResult: &llm.Result{
			Response: "saved response content",
		},
	}
	registry := newMockRegistry(provider, llm.ProviderOpenAI)
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider:     llm.ProviderOpenAI,
		SaveResponse: true,
		OutputPath:   outputPath,
	}

	result, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	require.NoError(t, err)
	assert.Equal(t, "saved response content", result.Response)

	savedContent, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "saved response content", string(savedContent))
}

func TestSendToLLMWithProgress_WithoutSaveResponse(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "response.md")

	provider := &mockLLMProvider{
		name:      "TestProvider",
		available: true,
		sendResult: &llm.Result{
			Response: "response",
		},
	}
	registry := newMockRegistry(provider, llm.ProviderOpenAI)
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider:     llm.ProviderOpenAI,
		SaveResponse: false,
		OutputPath:   outputPath,
	}

	result, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)

	_, err = os.Stat(outputPath)
	assert.True(t, os.IsNotExist(err), "file should not be created when SaveResponse is false")
}

func TestSendToLLMWithProgress_SaveResponseEmptyPath(t *testing.T) {
	t.Parallel()

	provider := &mockLLMProvider{
		name:      "TestProvider",
		available: true,
		sendResult: &llm.Result{
			Response: "response",
		},
	}
	registry := newMockRegistry(provider, llm.ProviderOpenAI)
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider:     llm.ProviderOpenAI,
		SaveResponse: true,
		OutputPath:   "",
	}

	result, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSendToLLMWithProgress_ProviderCreationFails(t *testing.T) {
	t.Parallel()

	registry := newMockRegistryWithError(errors.New("creation failed"), llm.ProviderOpenAI)
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider: llm.ProviderOpenAI,
	}

	_, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM provider")
}

func TestSendToLLMWithProgress_UnsupportedProvider(t *testing.T) {
	t.Parallel()

	registry := llm.NewRegistry()
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider: llm.ProviderOpenAI,
	}

	_, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM provider")
}

func TestSendToLLMWithProgress_ProviderNotAvailable(t *testing.T) {
	t.Parallel()

	provider := &mockLLMProvider{
		name:      "TestProvider",
		available: false,
	}
	registry := newMockRegistry(provider, llm.ProviderGeminiWeb)
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider: llm.ProviderGeminiWeb,
	}

	_, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestSendToLLMWithProgress_ValidateConfigFails(t *testing.T) {
	t.Parallel()

	provider := &mockLLMProvider{
		name:        "TestProvider",
		available:   true,
		validateErr: errors.New("API key required"),
	}
	registry := newMockRegistry(provider, llm.ProviderOpenAI)
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider: llm.ProviderOpenAI,
	}

	_, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid provider config")
}

func TestSendToLLMWithProgress_SendFails(t *testing.T) {
	t.Parallel()

	provider := &mockLLMProvider{
		name:      "TestProvider",
		available: true,
		sendErr:   errors.New("connection timeout"),
	}
	registry := newMockRegistry(provider, llm.ProviderAnthropic)
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider: llm.ProviderAnthropic,
	}

	_, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "LLM request failed")
}

func TestSendToLLMWithProgress_SaveFails_ReturnsResultWithError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	provider := &mockLLMProvider{
		name:      "TestProvider",
		available: true,
		sendResult: &llm.Result{
			Response: "response",
		},
	}
	registry := newMockRegistry(provider, llm.ProviderOpenAI)
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider:     llm.ProviderOpenAI,
		SaveResponse: true,
		OutputPath:   tmpDir,
	}

	result, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	assert.NotNil(t, result, "result should be returned even when save fails")
	assert.Equal(t, "response", result.Response)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save response")
}

func TestSendToLLMWithProgress_ConfigPassedToProvider(t *testing.T) {
	t.Parallel()

	var receivedCfg llm.Config
	registry := llm.NewRegistry()
	registry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
		receivedCfg = cfg
		return &mockLLMProvider{
			name:       "TestProvider",
			available:  true,
			sendResult: &llm.Result{Response: "ok"},
		}, nil
	})
	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider:       llm.ProviderOpenAI,
		APIKey:         "my-api-key",
		BaseURL:        "https://custom.api.com",
		Model:          "gpt-4o",
		Timeout:        60,
		BinaryPath:     "/path/to/binary",
		BrowserRefresh: "chrome",
	}

	_, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

	require.NoError(t, err)
	assert.Equal(t, llm.ProviderOpenAI, receivedCfg.Provider)
	assert.Equal(t, "my-api-key", receivedCfg.APIKey)
	assert.Equal(t, "https://custom.api.com", receivedCfg.BaseURL)
	assert.Equal(t, "gpt-4o", receivedCfg.Model)
	assert.Equal(t, 60, receivedCfg.Timeout)
	assert.Equal(t, "/path/to/binary", receivedCfg.BinaryPath)
	assert.Equal(t, "chrome", receivedCfg.BrowserRefresh)
}

func TestSendToLLMWithProgress_AllProviderTypes(t *testing.T) {
	t.Parallel()

	providerTypes := []llm.ProviderType{
		llm.ProviderOpenAI,
		llm.ProviderAnthropic,
		llm.ProviderGemini,
		llm.ProviderGeminiWeb,
	}

	for _, pt := range providerTypes {
		pt := pt
		t.Run(string(pt), func(t *testing.T) {
			t.Parallel()

			provider := &mockLLMProvider{
				name:       string(pt),
				available:  true,
				sendResult: &llm.Result{Response: "response from " + string(pt)},
			}
			registry := newMockRegistry(provider, pt)
			svc := NewContextService(WithRegistry(registry))

			cfg := LLMSendConfig{
				Provider: pt,
			}

			result, err := svc.SendToLLMWithProgress(context.Background(), "content", cfg, nil)

			require.NoError(t, err)
			assert.Equal(t, "response from "+string(pt), result.Response)
		})
	}
}

func TestSendToLLMWithProgress_ContextCancellation(t *testing.T) {
	t.Parallel()

	provider := &mockLLMProvider{
		name:      "TestProvider",
		available: true,
		sendErr:   context.Canceled,
	}
	registry := newMockRegistry(provider, llm.ProviderOpenAI)
	svc := NewContextService(WithRegistry(registry))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := LLMSendConfig{
		Provider: llm.ProviderOpenAI,
	}

	_, err := svc.SendToLLMWithProgress(ctx, "content", cfg, nil)

	require.Error(t, err)
}

func TestWithRegistry_Option(t *testing.T) {
	t.Parallel()

	customRegistry := llm.NewRegistry()
	svc := NewContextService(WithRegistry(customRegistry))

	assert.Equal(t, customRegistry, svc.registry)
}

func TestNewContextService_DefaultRegistry(t *testing.T) {
	t.Parallel()

	svc := NewContextService()

	assert.NotNil(t, svc.registry)
	assert.Equal(t, DefaultProviderRegistry, svc.registry)
}
