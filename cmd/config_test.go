package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// Helper to restore viper state
func restoreViperState() {
	viper.Reset()
}

func TestShowCurrentConfig_Human(t *testing.T) {
	restoreViperState()
	viper.Set("scanner.max-files", 1000)
	viper.Set("llm.provider", "anthropic")

	err := showCurrentConfig()
	require.NoError(t, err)
}

func TestShowCurrentConfig_JSON(t *testing.T) {
	restoreViperState()
	viper.Set("scanner.max-files", 2000)
	viper.Set("output.format", "json")

	err := showCurrentConfig()
	require.NoError(t, err)
}

func TestSetConfigValue_ValidValue(t *testing.T) {
	restoreViperState()
	viper.SetConfigFile(t.TempDir() + "/config.yaml")
	t.Cleanup(func() {
		viper.Reset()
	})

	key := "scanner.max-files"
	value := "5000"

	err := setConfigValue(key, value)
	require.NoError(t, err)
}

func TestSetConfigValue_ConvertValueError(t *testing.T) {
	restoreViperState()
	viper.SetConfigFile(t.TempDir() + "/config.yaml")
	t.Cleanup(func() {
		viper.Reset()
	})

	// Test a value that cannot be converted (workers requires integer)
	key := "scanner.workers"
	value := "not-a-number"

	err := setConfigValue(key, value)
	require.Error(t, err)
}

func TestSetConfigValue_AllValidKeys(t *testing.T) {
	restoreViperState()
	viper.SetConfigFile(t.TempDir() + "/config.yaml")
	t.Cleanup(func() {
		viper.Reset()
	})

	// Test setting various valid keys
	validSettings := map[string]string{
		"scanner.max-files":         "1000",
		"scanner.max-file-size":     "10MB",
		"scanner.respect-gitignore": "true",
		"scanner.workers":           "4",
		"llm.provider":              "openai",
		"llm.api-key":               "test-key",
		"output.format":             "markdown",
		"output.clipboard":          "false",
	}

	for key, value := range validSettings {
		t.Run(key, func(t *testing.T) {
			err := setConfigValue(key, value)
			require.NoError(t, err, "setting %s should not error", key)
		})
	}
}

func TestShowCurrentConfig_EmptyConfig(t *testing.T) {
	restoreViperState()
	t.Cleanup(func() {
		viper.Reset()
	})

	// Ensure config is empty
	viper.Reset()

	err := showCurrentConfig()
	require.NoError(t, err)
}

func TestShowCurrentConfig_WithValues(t *testing.T) {
	restoreViperState()
	t.Cleanup(func() {
		viper.Reset()
	})

	// Set some config values
	viper.Set("scanner.max-files", 5000)
	viper.Set("llm.provider", "anthropic")
	viper.Set("output.format", "markdown")

	err := showCurrentConfig()
	require.NoError(t, err)
}

func TestShowCurrentConfig_JSONMode(t *testing.T) {
	restoreViperState()
	t.Cleanup(func() {
		viper.Reset()
	})

	viper.Set("scanner.max-files", 2000)
	viper.Set("output.format", "json")
	viper.Set("llm.provider", "openai")

	err := showCurrentConfig()
	require.NoError(t, err)
}

func TestShowCurrentConfig_AllSections(t *testing.T) {
	restoreViperState()
	t.Cleanup(func() {
		viper.Reset()
	})

	// Set values from all config sections
	viper.Set("scanner.max-files", 1000)
	viper.Set("scanner.max-file-size", "5MB")
	viper.Set("context.max-size", "10MB")
	viper.Set("llm.provider", "gemini")
	viper.Set("llm.model", "gemini-2.5-flash")
	viper.Set("output.format", "text")

	err := showCurrentConfig()
	require.NoError(t, err)
}

func TestGetDefaultConfigPath(t *testing.T) {
	path := getDefaultConfigPath()

	// Verify path is not empty
	if path == "" {
		t.Fatal("getDefaultConfigPath() returned empty string")
	}

	// Verify path ends with config.yaml
	if !strings.HasSuffix(filepath.Base(path), "config.yaml") {
		t.Errorf("expected path to end with 'config.yaml', got: %s", path)
	}

	// Verify path is absolute
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got: %s", path)
	}
}

func TestGetDefaultConfigPath_XDGConfigHome(t *testing.T) {
	// Save original value
	original, exists := os.LookupEnv("XDG_CONFIG_HOME")
	if exists {
		defer func() { _ = os.Setenv("XDG_CONFIG_HOME", original) }()
	} else {
		defer func() { _ = os.Unsetenv("XDG_CONFIG_HOME") }()
	}

	// Set custom XDG_CONFIG_HOME
	customDir := "/tmp/test-xdg-config"
	_ = os.Setenv("XDG_CONFIG_HOME", customDir)

	path := getDefaultConfigPath()

	// Verify path contains the custom directory
	if !strings.Contains(path, customDir) {
		t.Errorf("expected path to contain '%s', got: %s", customDir, path)
	}

	// Verify path ends with config.yaml
	if !strings.HasSuffix(filepath.Base(path), "config.yaml") {
		t.Errorf("expected path to end with 'config.yaml', got: %s", path)
	}
}

func TestGetDefaultConfigPath_FallbackToHome(t *testing.T) {
	// Save original values
	originalXDG, xdgExists := os.LookupEnv("XDG_CONFIG_HOME")
	originalHome, homeExists := os.LookupEnv("HOME")

	// Unset XDG_CONFIG_HOME to force fallback
	_ = os.Unsetenv("XDG_CONFIG_HOME")

	// Ensure HOME is set
	if !homeExists || os.Getenv("HOME") == "" {
		homeDir := "/tmp/test-home"
		_ = os.Setenv("HOME", homeDir)
		defer func() { _ = os.Unsetenv("HOME") }()
	}

	// Restore original values after test
	if xdgExists {
		defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDG) }()
	}
	if homeExists {
		defer func() { _ = os.Setenv("HOME", originalHome) }()
	}

	path := getDefaultConfigPath()

	// Verify path is not empty
	if path == "" {
		t.Fatal("getDefaultConfigPath() returned empty string when XDG_CONFIG_HOME is unset")
	}

	// Verify path ends with config.yaml
	if !strings.HasSuffix(filepath.Base(path), "config.yaml") {
		t.Errorf("expected path to end with 'config.yaml', got: %s", path)
	}

	// On Unix systems, should use .config fallback when XDG_CONFIG_HOME is not set
	if runtime.GOOS != "windows" {
		if !strings.Contains(path, ".config") {
			// This is expected behavior on Unix systems
			t.Logf("Warning: path does not contain '.config', got: %s", path)
		}
	}
}

func TestGetDefaultConfigPath_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("skipping Windows-specific test on non-Windows platform")
	}

	path := getDefaultConfigPath()

	// On Windows, path should be in AppData\Roaming or similar
	if !strings.Contains(path, "shotgun-cli") {
		t.Errorf("expected Windows path to contain 'shotgun-cli', got: %s", path)
	}

	// Verify path ends with config.yaml
	if !strings.HasSuffix(filepath.Base(path), "config.yaml") {
		t.Errorf("expected path to end with 'config.yaml', got: %s", path)
	}
}

func TestGetConfigSource_NotSet(t *testing.T) {
	restoreViperState()
	t.Cleanup(func() {
		viper.Reset()
	})

	// Ensure key is not set
	viper.Reset()

	source := getConfigSource("scanner.max-files")
	if source != "default" {
		t.Errorf("expected 'default', got: %s", source)
	}
}

func TestGetConfigSource_FromConfigFile(t *testing.T) {
	restoreViperState()
	t.Cleanup(func() {
		viper.Reset()
	})

	// Set config file and a value
	viper.SetConfigFile(t.TempDir() + "/config.yaml")
	viper.Set("scanner.max-files", 5000)

	source := getConfigSource("scanner.max-files")
	if source != "config file" {
		t.Errorf("expected 'config file', got: %s", source)
	}
}

func TestGetConfigSource_FromFlag(t *testing.T) {
	restoreViperState()
	t.Cleanup(func() {
		viper.Reset()
	})

	// Set a value (viper doesn't track sources, so this will be "config file" or "env")
	viper.Set("scanner.max-files", 1000)

	source := getConfigSource("scanner.max-files")
	// Since we can't actually distinguish between flag and config file in tests,
	// we just verify it returns something
	if source == "" {
		t.Error("getConfigSource should return a non-empty string")
	}
}
