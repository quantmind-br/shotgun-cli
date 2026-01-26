// Package config provides centralized configuration validation.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/quantmind-br/shotgun-cli/internal/utils"
)

// ValidKeys returns all valid configuration keys.
func ValidKeys() []string {
	return []string{
		// Scanner keys
		KeyScannerMaxFiles,
		KeyScannerMaxFileSize,
		KeyScannerRespectGitignore,
		KeyScannerSkipBinary,
		KeyScannerWorkers,
		KeyScannerIncludeHidden,
		KeyScannerIncludeIgnored,
		KeyScannerRespectShotgunignore,
		KeyScannerMaxMemory,
		// Context keys
		KeyContextMaxSize,
		KeyContextIncludeTree,
		KeyContextIncludeSummary,
		// Template keys
		KeyTemplateCustomPath,
		// Output keys
		KeyOutputFormat,
		KeyOutputClipboard,
		// LLM Provider keys
		KeyLLMProvider,
		KeyLLMAPIKey,
		KeyLLMBaseURL,
		KeyLLMModel,
		KeyLLMTimeout,
		// LLM save response key
		KeyLLMSaveResponse,
	}
}

// IsValidKey checks if the given key is a valid configuration key.
func IsValidKey(key string) bool {
	for _, validKey := range ValidKeys() {
		if key == validKey {
			return true
		}
	}
	return false
}

// ValidateValue validates a configuration value for the given key.
func ValidateValue(key, value string) error {
	switch key {
	case KeyScannerMaxFiles:
		return validateMaxFiles(value)
	case KeyScannerMaxFileSize, KeyContextMaxSize, KeyScannerMaxMemory:
		return validateSizeFormat(value)
	case KeyScannerRespectGitignore, KeyScannerSkipBinary,
		KeyScannerIncludeHidden, KeyScannerIncludeIgnored, KeyScannerRespectShotgunignore,
		KeyContextIncludeTree, KeyContextIncludeSummary, KeyOutputClipboard,
		KeyLLMSaveResponse:
		return validateBooleanValue(value)
	case KeyScannerWorkers:
		return validateWorkers(value)
	case KeyOutputFormat:
		return validateOutputFormat(value)
	case KeyTemplateCustomPath:
		return validatePath(value)
	case KeyLLMTimeout:
		return validateTimeout(value)
	case KeyLLMProvider:
		return validateLLMProvider(value)
	case KeyLLMAPIKey:
		return nil // API key can be any string
	case KeyLLMBaseURL:
		return validateURL(value)
	case KeyLLMModel:
		return nil // Model can be any string, validation is provider-specific
	}

	return nil
}

// ConvertValue converts a string configuration value to the appropriate type.
func ConvertValue(key, value string) (interface{}, error) {
	switch key {
	case KeyScannerMaxFiles, KeyScannerWorkers, KeyLLMTimeout:
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err != nil {
			return nil, fmt.Errorf("failed to parse integer value: %w", err)
		}
		return intVal, nil

	case KeyScannerRespectGitignore, KeyScannerSkipBinary,
		KeyScannerIncludeHidden, KeyScannerIncludeIgnored, KeyScannerRespectShotgunignore,
		KeyContextIncludeTree, KeyContextIncludeSummary, KeyOutputClipboard,
		KeyLLMSaveResponse:
		return strings.ToLower(value) == "true", nil

	default:
		// String values
		return value, nil
	}
}

// validateWorkers validates the workers configuration value.
func validateWorkers(value string) error {
	var workers int
	if _, err := fmt.Sscanf(value, "%d", &workers); err != nil {
		return fmt.Errorf("expected a positive integer")
	}
	if workers < 1 || workers > 32 {
		return fmt.Errorf("must be between 1 and 32, got %d", workers)
	}
	return nil
}

// validateMaxFiles validates the max-files configuration value.
func validateMaxFiles(value string) error {
	// Reject size formats (e.g., "10MB", "1KB")
	upper := strings.ToUpper(strings.TrimSpace(value))
	isSizeFormat := strings.HasSuffix(upper, "GB") || strings.HasSuffix(upper, "MB") ||
		strings.HasSuffix(upper, "KB")
	if !isSizeFormat && strings.HasSuffix(upper, "B") && len(upper) > 1 {
		isSizeFormat = upper[len(upper)-2] >= '0' && upper[len(upper)-2] <= '9'
	}
	if isSizeFormat {
		return fmt.Errorf("expected a number, got size format")
	}

	var dummy int
	if _, err := fmt.Sscanf(value, "%d", &dummy); err != nil {
		return fmt.Errorf("expected a positive integer")
	}
	if dummy <= 0 {
		return fmt.Errorf("must be positive, got %d", dummy)
	}
	return nil
}

// validateSizeFormat validates size format values (e.g., "1MB", "500KB").
func validateSizeFormat(value string) error {
	if _, err := utils.ParseSize(value); err != nil {
		return fmt.Errorf("expected size format (e.g., 1MB, 500KB): %w", err)
	}
	return nil
}

// validateBooleanValue validates boolean configuration values.
func validateBooleanValue(value string) error {
	lower := strings.ToLower(value)
	if lower != "true" && lower != "false" {
		return fmt.Errorf("expected 'true' or 'false', got '%s'", value)
	}
	return nil
}

// validateOutputFormat validates output format configuration values.
func validateOutputFormat(value string) error {
	if value != "markdown" && value != "text" {
		return fmt.Errorf("expected 'markdown' or 'text', got '%s'", value)
	}
	return nil
}

// validatePath validates file/directory path configuration values.
func validatePath(value string) error {
	if value == "" {
		return nil
	}

	expandedValue := value
	if strings.HasPrefix(value, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to expand home directory: %w", err)
		}
		expandedValue = filepath.Join(home, value[2:])
	}

	parentDir := filepath.Dir(expandedValue)
	if parentDir != "." && parentDir != "/" {
		if info, err := os.Stat(parentDir); err == nil {
			if !info.IsDir() {
				return fmt.Errorf("parent path exists but is not a directory: %s", parentDir)
			}
		}
	}
	return nil
}

// validateTimeout validates timeout configuration values.
func validateTimeout(value string) error {
	var timeout int
	if _, err := fmt.Sscanf(value, "%d", &timeout); err != nil {
		return fmt.Errorf("expected a positive integer (seconds)")
	}
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %d", timeout)
	}
	if timeout > 3600 {
		return fmt.Errorf("timeout too large (max 3600 seconds), got %d", timeout)
	}
	return nil
}

// validateLLMProvider validates LLM provider configuration values.
func validateLLMProvider(value string) error {
	validProviders := []string{"openai", "anthropic", "gemini"}
	for _, provider := range validProviders {
		if value == provider {
			return nil
		}
	}
	return fmt.Errorf("expected one of: %s", strings.Join(validProviders, ", "))
}

// validateURL validates URL configuration values.
func validateURL(value string) error {
	if value == "" {
		return nil
	}
	// Basic URL validation
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	return nil
}
