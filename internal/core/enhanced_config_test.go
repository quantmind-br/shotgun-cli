package core

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhancedConfigDefaults(t *testing.T) {
	config := DefaultEnhancedConfig()

	// Test OpenAI defaults
	assert.Equal(t, "https://api.openai.com/v1", config.OpenAI.BaseURL)
	assert.Equal(t, "gpt-4o", config.OpenAI.Model)
	assert.Equal(t, 300, config.OpenAI.Timeout)
	assert.Equal(t, 4096, config.OpenAI.MaxTokens)
	assert.Equal(t, float64(0.7), config.OpenAI.Temperature)
	assert.Equal(t, 3, config.OpenAI.MaxRetries)
	assert.Equal(t, 2, config.OpenAI.RetryDelay)

	// Test Translation defaults
	assert.False(t, config.Translation.Enabled)
	assert.Equal(t, "en", config.Translation.TargetLanguage)
	assert.True(t, config.Translation.CacheEnabled)
	assert.Equal(t, 1000, config.Translation.CacheSize)
	assert.Equal(t, 3600, config.Translation.CacheTTL)

	// Test App defaults
	assert.Equal(t, "auto", config.App.Theme)
	assert.True(t, config.App.AutoSave)
	assert.True(t, config.App.ShowLineNumbers)
	assert.Equal(t, "dev", config.App.DefaultTemplate)
	assert.Equal(t, int64(10485760), config.App.MaxFileSize) // 10MB
	assert.Equal(t, 10, config.App.MaxDirectoryDepth)
	assert.Equal(t, 10, config.App.WorkerPoolSize)
	assert.Equal(t, 1000, config.App.RefreshInterval)
	assert.True(t, config.App.EnableHotReload)
}

func TestEnhancedConfigValidation(t *testing.T) {
	validator := SetupEnhancedValidator()

	tests := []struct {
		name        string
		config      *EnhancedConfig
		expectValid bool
		expectError string
	}{
		{
			name:        "valid default config",
			config:      DefaultEnhancedConfig(),
			expectValid: true,
		},
		{
			name: "invalid OpenAI timeout",
			config: func() *EnhancedConfig {
				c := DefaultEnhancedConfig()
				c.OpenAI.Timeout = -1
				return c
			}(),
			expectValid: false,
			expectError: "timeout",
		},
		{
			name: "invalid temperature range",
			config: func() *EnhancedConfig {
				c := DefaultEnhancedConfig()
				c.OpenAI.Temperature = 5.0 // Outside 0-2 range
				return c
			}(),
			expectValid: false,
			expectError: "temperature",
		},
		{
			name: "invalid max tokens",
			config: func() *EnhancedConfig {
				c := DefaultEnhancedConfig()
				c.OpenAI.MaxTokens = 0
				return c
			}(),
			expectValid: false,
			expectError: "maxTokens",
		},
		{
			name: "invalid cache size",
			config: func() *EnhancedConfig {
				c := DefaultEnhancedConfig()
				c.Translation.CacheSize = -1
				return c
			}(),
			expectValid: false,
			expectError: "cacheSize",
		},
		{
			name: "invalid worker pool size",
			config: func() *EnhancedConfig {
				c := DefaultEnhancedConfig()
				c.App.WorkerPoolSize = 0
				return c
			}(),
			expectValid: false,
			expectError: "workerPoolSize",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate(validator)

			if tt.expectValid {
				assert.True(t, result.Valid, "Expected config to be valid")
				assert.Empty(t, result.Errors, "Expected no validation errors")
			} else {
				assert.False(t, result.Valid, "Expected config to be invalid")
				assert.NotEmpty(t, result.Errors, "Expected validation errors")

				// Check that the expected error field is mentioned in the error message
				found := false
				for _, err := range result.Errors {
					if strings.Contains(strings.ToLower(err.Message), strings.ToLower(tt.expectError)) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error for field %s not found in error messages. Got: %s", tt.expectError, result.Errors[0].Message)
			}
		})
	}
}

func TestEnhancedConfigClone(t *testing.T) {
	original := DefaultEnhancedConfig()
	original.OpenAI.APIKeyAlias = "test-key"
	original.Translation.Enabled = true
	original.Translation.TargetLanguage = "es"

	cloned := original.Clone()

	// Verify values are copied
	assert.Equal(t, original.OpenAI.APIKeyAlias, cloned.OpenAI.APIKeyAlias)
	assert.Equal(t, original.Translation.Enabled, cloned.Translation.Enabled)
	assert.Equal(t, original.Translation.TargetLanguage, cloned.Translation.TargetLanguage)

	// Verify they are independent
	cloned.OpenAI.APIKeyAlias = "different-key"
	assert.NotEqual(t, original.OpenAI.APIKeyAlias, cloned.OpenAI.APIKeyAlias)
}

func TestEnhancedConfigLegacyConversion(t *testing.T) {
	enhanced := DefaultEnhancedConfig()
	enhanced.OpenAI.APIKeyAlias = "test-key"
	enhanced.Translation.Enabled = true
	enhanced.Translation.TargetLanguage = "fr"

	// Convert to legacy
	legacy := enhanced.ToLegacyConfig()

	// Verify conversion
	assert.Equal(t, enhanced.OpenAI.BaseURL, legacy.OpenAI.BaseURL)
	assert.Equal(t, enhanced.OpenAI.Model, legacy.OpenAI.Model)
	assert.Equal(t, enhanced.Translation.Enabled, legacy.Translation.Enabled)
	assert.Equal(t, enhanced.Translation.TargetLanguage, legacy.Translation.TargetLanguage)

	// Convert back from legacy
	backConverted := FromLegacyConfig(legacy)

	// Verify round-trip conversion
	assert.Equal(t, enhanced.OpenAI.BaseURL, backConverted.OpenAI.BaseURL)
	assert.Equal(t, enhanced.OpenAI.Model, backConverted.OpenAI.Model)
	assert.Equal(t, enhanced.Translation.Enabled, backConverted.Translation.Enabled)
	assert.Equal(t, enhanced.Translation.TargetLanguage, backConverted.Translation.TargetLanguage)
}

func TestEnhancedConfigManagerCreation(t *testing.T) {
	// Create config manager
	manager, err := NewEnhancedConfigManager()
	require.NoError(t, err)
	assert.NotNil(t, manager)

	// Test initial state
	config := manager.GetEnhanced()
	assert.NotNil(t, config)

	// Test that it has default values
	assert.Equal(t, "gpt-4o", config.OpenAI.Model)
	assert.Equal(t, "auto", config.App.Theme)
}

func TestEnhancedConfigManagerSaveLoad(t *testing.T) {
	// Create config manager
	manager, err := NewEnhancedConfigManager()
	require.NoError(t, err)

	// Modify configuration
	config := manager.GetEnhanced()
	config.OpenAI.Model = "gpt-4"
	config.Translation.Enabled = true
	config.Translation.TargetLanguage = "es"
	config.App.Theme = "dark"

	// Update and save
	err = manager.UpdateEnhanced(config)
	require.NoError(t, err)

	err = manager.Save()
	require.NoError(t, err)

	// Save should succeed (file will be saved to default location)

	// Create new manager to test loading
	manager2, err := NewEnhancedConfigManager()
	require.NoError(t, err)

	err = manager2.Load()
	require.NoError(t, err)

	// Verify loaded configuration
	loadedConfig := manager2.GetEnhanced()
	assert.Equal(t, "gpt-4", loadedConfig.OpenAI.Model)
	assert.True(t, loadedConfig.Translation.Enabled)
	assert.Equal(t, "es", loadedConfig.Translation.TargetLanguage)
	assert.Equal(t, "dark", loadedConfig.App.Theme)
}

func TestEnhancedConfigManagerEnvironmentOverrides(t *testing.T) {
	// Set environment variables with valid values that meet validation requirements
	os.Setenv("SHOTGUN_OPENAI_MODEL", "gpt-3.5-turbo")
	os.Setenv("SHOTGUN_OPENAI_BASEURL", "https://api.openai.com/v1")
	os.Setenv("SHOTGUN_OPENAI_TIMEOUT", "300")
	os.Setenv("SHOTGUN_OPENAI_MAXTOKENS", "4096")
	os.Setenv("SHOTGUN_TRANSLATION_ENABLED", "true")
	os.Setenv("SHOTGUN_TRANSLATION_TARGETLANGUAGE", "en")
	os.Setenv("SHOTGUN_TRANSLATION_CACHESIZE", "1000")
	os.Setenv("SHOTGUN_TRANSLATION_CACHETTL", "3600")
	os.Setenv("SHOTGUN_APP_THEME", "light")
	os.Setenv("SHOTGUN_APP_DEFAULTTEMPLATE", "dev")
	os.Setenv("SHOTGUN_APP_MAXFILESIZE", "10485760")
	os.Setenv("SHOTGUN_APP_MAXDIRECTORYDEPTH", "10")
	os.Setenv("SHOTGUN_APP_WORKERPOOLSIZE", "10")
	os.Setenv("SHOTGUN_APP_REFRESHINTERVAL", "1000")
	defer func() {
		os.Unsetenv("SHOTGUN_OPENAI_MODEL")
		os.Unsetenv("SHOTGUN_OPENAI_BASEURL")
		os.Unsetenv("SHOTGUN_OPENAI_TIMEOUT")
		os.Unsetenv("SHOTGUN_OPENAI_MAXTOKENS")
		os.Unsetenv("SHOTGUN_TRANSLATION_ENABLED")
		os.Unsetenv("SHOTGUN_TRANSLATION_TARGETLANGUAGE")
		os.Unsetenv("SHOTGUN_TRANSLATION_CACHESIZE")
		os.Unsetenv("SHOTGUN_TRANSLATION_CACHETTL")
		os.Unsetenv("SHOTGUN_APP_THEME")
		os.Unsetenv("SHOTGUN_APP_DEFAULTTEMPLATE")
		os.Unsetenv("SHOTGUN_APP_MAXFILESIZE")
		os.Unsetenv("SHOTGUN_APP_MAXDIRECTORYDEPTH")
		os.Unsetenv("SHOTGUN_APP_WORKERPOOLSIZE")
		os.Unsetenv("SHOTGUN_APP_REFRESHINTERVAL")
	}()

	// Create config manager
	manager, err := NewEnhancedConfigManager()
	require.NoError(t, err)

	// Load configuration (should pick up environment variables)
	err = manager.Load()
	require.NoError(t, err)

	config := manager.GetEnhanced()

	// Verify environment overrides
	assert.Equal(t, "gpt-3.5-turbo", config.OpenAI.Model)
	assert.True(t, config.Translation.Enabled)
	assert.Equal(t, "light", config.App.Theme)
}

func TestEnhancedConfigManagerValidation(t *testing.T) {
	// Create config manager
	manager, err := NewEnhancedConfigManager()
	require.NoError(t, err)

	// Create invalid configuration
	invalidConfig := DefaultEnhancedConfig()
	invalidConfig.OpenAI.Timeout = -1      // Invalid timeout
	invalidConfig.OpenAI.Temperature = 5.0 // Invalid temperature

	// Try to update with invalid config
	err = manager.UpdateEnhanced(invalidConfig)
	assert.Error(t, err, "Should reject invalid configuration")

	// Verify original config is unchanged
	config := manager.GetEnhanced()
	assert.Equal(t, 300, config.OpenAI.Timeout)              // Should still be default
	assert.Equal(t, float64(0.7), config.OpenAI.Temperature) // Should still be default
}

func TestEnhancedConfigManagerReset(t *testing.T) {
	// Create config manager
	manager, err := NewEnhancedConfigManager()
	require.NoError(t, err)

	// Modify configuration
	config := manager.GetEnhanced()
	config.OpenAI.Model = "gpt-4"
	config.Translation.Enabled = true

	err = manager.UpdateEnhanced(config)
	require.NoError(t, err)

	// Reset configuration
	err = manager.Reset()
	require.NoError(t, err)

	// Verify configuration is back to defaults
	resetConfig := manager.GetEnhanced()
	assert.Equal(t, "gpt-4o", resetConfig.OpenAI.Model)
	assert.False(t, resetConfig.Translation.Enabled)
}

func TestEnhancedConfigManagerConcurrency(t *testing.T) {
	// Create config manager
	manager, err := NewEnhancedConfigManager()
	require.NoError(t, err)

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			config := manager.GetEnhanced()
			assert.NotNil(t, config)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent reads")
		}
	}

	// Test concurrent writes
	for i := 0; i < 5; i++ {
		go func(index int) {
			defer func() { done <- true }()
			config := manager.GetEnhanced()
			config.App.RefreshInterval = 100 + index
			err := manager.UpdateEnhanced(config)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all writes to complete
	for i := 0; i < 5; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent writes")
		}
	}

	// Verify final state is consistent
	finalConfig := manager.GetEnhanced()
	assert.True(t, finalConfig.App.RefreshInterval >= 100 && finalConfig.App.RefreshInterval <= 104)
}

func TestEnhancedConfigManagerInvalidFilePath(t *testing.T) {
	// Create config manager
	manager, err := NewEnhancedConfigManager()
	require.NoError(t, err) // Should create manager but not fail yet

	// Loading from invalid path should work (creates defaults)
	err = manager.Load()
	assert.NoError(t, err, "Should handle missing file gracefully")

	// But saving to invalid path should fail
	err = manager.Save()
	assert.Error(t, err, "Should fail to save to invalid path")
}

// Benchmark tests for performance validation
func BenchmarkEnhancedConfigValidation(b *testing.B) {
	validator := SetupEnhancedValidator()
	config := DefaultEnhancedConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := config.Validate(validator)
		if !result.Valid {
			b.Fatal("Config should be valid")
		}
	}
}

func BenchmarkEnhancedConfigClone(b *testing.B) {
	config := DefaultEnhancedConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cloned := config.Clone()
		if cloned == nil {
			b.Fatal("Clone should not be nil")
		}
	}
}

func BenchmarkEnhancedConfigManagerGetConfig(b *testing.B) {
	// Create config manager
	manager, err := NewEnhancedConfigManager()
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := manager.GetEnhanced()
		if config == nil {
			b.Fatal("Config should not be nil")
		}
	}
}
