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

type integrationMockProvider struct {
	name           string
	available      bool
	validateErr    error
	sendResult     *llm.Result
	sendErr        error
	progressStages []string
	callCount      int
}

func (m *integrationMockProvider) Name() string      { return m.name }
func (m *integrationMockProvider) IsAvailable() bool { return m.available }
func (m *integrationMockProvider) IsConfigured() bool {
	return true
}
func (m *integrationMockProvider) ValidateConfig() error { return m.validateErr }
func (m *integrationMockProvider) Send(ctx context.Context, content string) (*llm.Result, error) {
	m.callCount++
	return m.sendResult, m.sendErr
}
func (m *integrationMockProvider) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
	m.callCount++
	if progress != nil {
		for _, stage := range m.progressStages {
			progress(stage)
		}
	}
	return m.sendResult, m.sendErr
}

func TestLLMFlow_EndToEnd(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "response.md")

	provider := &integrationMockProvider{
		name:           "MockProvider",
		available:      true,
		progressStages: []string{"connecting", "sending", "receiving"},
		sendResult: &llm.Result{
			Response:    "Generated response content",
			RawResponse: "Raw response",
			Model:       "test-model",
			Provider:    "MockProvider",
			Duration:    150 * time.Millisecond,
		},
	}

	registry := llm.NewRegistry()
	registry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
		return provider, nil
	})

	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider:     llm.ProviderOpenAI,
		APIKey:       "test-key",
		Model:        "test-model",
		SaveResponse: true,
		OutputPath:   outputPath,
	}

	var stages []string
	progress := func(stage string) {
		stages = append(stages, stage)
	}

	result, err := svc.SendToLLMWithProgress(context.Background(), "test prompt content", cfg, progress)

	require.NoError(t, err)
	assert.Equal(t, "Generated response content", result.Response)
	assert.Equal(t, "test-model", result.Model)
	assert.Equal(t, 150*time.Millisecond, result.Duration)

	assert.Equal(t, []string{"connecting", "sending", "receiving"}, stages)

	savedContent, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "Generated response content", string(savedContent))

	assert.Equal(t, 1, provider.callCount)
}

func TestLLMFlow_MultipleProviders(t *testing.T) {
	t.Parallel()

	providers := []llm.ProviderType{
		llm.ProviderOpenAI,
		llm.ProviderAnthropic,
		llm.ProviderGemini,
		llm.ProviderGeminiWeb,
	}

	for _, providerType := range providers {
		providerType := providerType
		t.Run(string(providerType), func(t *testing.T) {
			t.Parallel()

			provider := &integrationMockProvider{
				name:      string(providerType),
				available: true,
				sendResult: &llm.Result{
					Response: "response from " + string(providerType),
					Provider: string(providerType),
				},
			}

			registry := llm.NewRegistry()
			registry.Register(providerType, func(cfg llm.Config) (llm.Provider, error) {
				return provider, nil
			})

			svc := NewContextService(WithRegistry(registry))

			cfg := LLMSendConfig{
				Provider: providerType,
			}

			result, err := svc.SendToLLMWithProgress(context.Background(), "test content", cfg, nil)

			require.NoError(t, err)
			assert.Equal(t, "response from "+string(providerType), result.Response)
		})
	}
}

func TestLLMFlow_ErrorRecovery(t *testing.T) {
	t.Parallel()

	callCount := 0
	provider := &integrationMockProvider{
		name:      "MockProvider",
		available: true,
	}

	registry := llm.NewRegistry()
	registry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
		callCount++
		if callCount == 1 {
			provider.sendErr = errors.New("temporary failure")
			provider.sendResult = nil
		} else {
			provider.sendErr = nil
			provider.sendResult = &llm.Result{Response: "success on retry"}
		}
		return provider, nil
	})

	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider: llm.ProviderOpenAI,
	}

	_, err := svc.SendToLLMWithProgress(context.Background(), "test", cfg, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "temporary failure")

	result, err := svc.SendToLLMWithProgress(context.Background(), "test", cfg, nil)
	require.NoError(t, err)
	assert.Equal(t, "success on retry", result.Response)
}

func TestLLMFlow_ProgressReporting(t *testing.T) {
	t.Parallel()

	provider := &integrationMockProvider{
		name:           "MockProvider",
		available:      true,
		progressStages: []string{"Initializing", "Connecting", "Sending", "Waiting", "Receiving", "Complete"},
		sendResult:     &llm.Result{Response: "ok"},
	}

	registry := llm.NewRegistry()
	registry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
		return provider, nil
	})

	svc := NewContextService(WithRegistry(registry))

	var reportedStages []string
	progress := func(stage string) {
		reportedStages = append(reportedStages, stage)
	}

	cfg := LLMSendConfig{
		Provider: llm.ProviderOpenAI,
	}

	_, err := svc.SendToLLMWithProgress(context.Background(), "test", cfg, progress)

	require.NoError(t, err)
	assert.Contains(t, reportedStages, "Connecting")
	assert.Contains(t, reportedStages, "Complete")
	assert.Len(t, reportedStages, 6)
}

func TestLLMFlow_FileSave(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		saveResponse bool
		outputPath   string
		shouldSave   bool
	}{
		{
			name:         "saves when enabled with path",
			saveResponse: true,
			outputPath:   filepath.Join(tmpDir, "response1.md"),
			shouldSave:   true,
		},
		{
			name:         "does not save when disabled",
			saveResponse: false,
			outputPath:   filepath.Join(tmpDir, "response2.md"),
			shouldSave:   false,
		},
		{
			name:         "does not save when path empty",
			saveResponse: true,
			outputPath:   "",
			shouldSave:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider := &integrationMockProvider{
				name:       "MockProvider",
				available:  true,
				sendResult: &llm.Result{Response: "test response content"},
			}

			registry := llm.NewRegistry()
			registry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
				return provider, nil
			})

			svc := NewContextService(WithRegistry(registry))

			cfg := LLMSendConfig{
				Provider:     llm.ProviderOpenAI,
				SaveResponse: tt.saveResponse,
				OutputPath:   tt.outputPath,
			}

			result, err := svc.SendToLLMWithProgress(context.Background(), "test", cfg, nil)

			require.NoError(t, err)
			assert.Equal(t, "test response content", result.Response)

			if tt.shouldSave {
				content, err := os.ReadFile(tt.outputPath)
				require.NoError(t, err)
				assert.Equal(t, "test response content", string(content))
			} else if tt.outputPath != "" {
				_, err := os.Stat(tt.outputPath)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestLLMFlow_ConfigPropagation(t *testing.T) {
	t.Parallel()

	var receivedCfg llm.Config
	registry := llm.NewRegistry()
	registry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
		receivedCfg = cfg
		return &integrationMockProvider{
			name:       "MockProvider",
			available:  true,
			sendResult: &llm.Result{Response: "ok"},
		}, nil
	})

	svc := NewContextService(WithRegistry(registry))

	cfg := LLMSendConfig{
		Provider:       llm.ProviderOpenAI,
		APIKey:         "secret-api-key",
		BaseURL:        "https://custom.api.endpoint.com",
		Model:          "custom-model-v2",
		Timeout:        120,
		BinaryPath:     "/custom/binary/path",
		BrowserRefresh: "chrome",
	}

	_, err := svc.SendToLLMWithProgress(context.Background(), "test", cfg, nil)

	require.NoError(t, err)
	assert.Equal(t, llm.ProviderOpenAI, receivedCfg.Provider)
	assert.Equal(t, "secret-api-key", receivedCfg.APIKey)
	assert.Equal(t, "https://custom.api.endpoint.com", receivedCfg.BaseURL)
	assert.Equal(t, "custom-model-v2", receivedCfg.Model)
	assert.Equal(t, 120, receivedCfg.Timeout)
	assert.Equal(t, "/custom/binary/path", receivedCfg.BinaryPath)
	assert.Equal(t, "chrome", receivedCfg.BrowserRefresh)
}

func TestLLMFlow_ContextCancellation(t *testing.T) {
	t.Parallel()

	provider := &integrationMockProvider{
		name:      "MockProvider",
		available: true,
		sendErr:   context.Canceled,
	}

	registry := llm.NewRegistry()
	registry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
		return provider, nil
	})

	svc := NewContextService(WithRegistry(registry))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := LLMSendConfig{
		Provider: llm.ProviderOpenAI,
	}

	_, err := svc.SendToLLMWithProgress(ctx, "test", cfg, nil)

	require.Error(t, err)
}

func TestLLMFlow_ServiceLayerIsolation(t *testing.T) {
	t.Parallel()

	provider1 := &integrationMockProvider{
		name:       "Provider1",
		available:  true,
		sendResult: &llm.Result{Response: "response1"},
	}
	provider2 := &integrationMockProvider{
		name:       "Provider2",
		available:  true,
		sendResult: &llm.Result{Response: "response2"},
	}

	registry1 := llm.NewRegistry()
	registry1.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
		return provider1, nil
	})

	registry2 := llm.NewRegistry()
	registry2.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
		return provider2, nil
	})

	svc1 := NewContextService(WithRegistry(registry1))
	svc2 := NewContextService(WithRegistry(registry2))

	cfg := LLMSendConfig{Provider: llm.ProviderOpenAI}

	result1, err1 := svc1.SendToLLMWithProgress(context.Background(), "test", cfg, nil)
	result2, err2 := svc2.SendToLLMWithProgress(context.Background(), "test", cfg, nil)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, "response1", result1.Response)
	assert.Equal(t, "response2", result2.Response)
}
