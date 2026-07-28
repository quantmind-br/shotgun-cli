// Package config provides centralized configuration validation.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/quantmind-br/shotgun-cli/internal/utils"
)

// ValidKeys returns all valid configuration keys, derived from the metadata
// registry so the two cannot drift apart. Registering a key in
// buildAllMetadata() is the single edit needed to make it usable.
func ValidKeys() []string {
	metadata := AllConfigMetadata()
	keys := make([]string, 0, len(metadata))
	for _, m := range metadata {
		keys = append(keys, m.Key)
	}

	return keys
}

// deprecatedKeys maps retired configuration keys to the message shown when one is
// still used. Raw string literals on purpose: the Key* constants are gone, and
// the whole point is to keep recognising what users may still have on disk.
var deprecatedKeys = map[string]string{
	"scanner.workers":    "scanner.workers was removed; it had no effect",
	"scanner.max-memory": "scanner.max-memory was removed; it had no effect",
}

// DeprecationMessage reports whether key was retired and, if so, the message
// explaining it. Callers should consult this before IsValidKey so a retired key
// produces a specific explanation rather than a generic "invalid key" error.
func DeprecationMessage(key string) (string, bool) {
	msg, ok := deprecatedKeys[key]

	return msg, ok
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

// ValidateValue validates a configuration value for the given key, dispatching
// on the type declared in the metadata registry. Registering a key in
// buildAllMetadata() is the single edit needed to make it validate.
func ValidateValue(key, value string) error {
	meta, ok := GetMetadata(key)
	if !ok {
		// Unknown key: IsValidKey is the gate, and the previous per-key switch
		// fell through to nil here. Preserve that.
		return nil
	}

	switch meta.Type {
	case TypeBool:
		return validateBooleanValue(value)
	case TypeSize:
		return validateSizeFormat(value)
	case TypeEnum:
		return validateEnumValue(value, meta.EnumOptions)
	case TypeInt:
		return validateIntValue(value, meta)
	case TypeTimeout:
		return validateTimeoutValue(value, meta)
	case TypePath:
		return validatePath(value)
	case TypeURL:
		return validateURL(value)
	case TypeString:
		return nil
	}

	return nil
}

// ConvertValue converts a string configuration value to the appropriate type,
// dispatching on the type declared in the metadata registry.
func ConvertValue(key, value string) (interface{}, error) {
	meta, ok := GetMetadata(key)
	if !ok {
		return value, nil
	}

	switch meta.Type {
	case TypeInt, TypeTimeout:
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err != nil {
			return nil, fmt.Errorf("failed to parse integer value: %w", err)
		}

		return intVal, nil
	case TypeBool:
		return strings.ToLower(value) == "true", nil
	default:
		return value, nil
	}
}

// isSizeFormat reports whether value looks like a size literal (e.g. "10MB",
// "1KB", "512B") rather than a plain integer.
func isSizeFormat(value string) bool {
	upper := strings.ToUpper(strings.TrimSpace(value))
	if strings.HasSuffix(upper, "GB") || strings.HasSuffix(upper, "MB") || strings.HasSuffix(upper, "KB") {
		return true
	}
	if strings.HasSuffix(upper, "B") && len(upper) > 1 {
		return upper[len(upper)-2] >= '0' && upper[len(upper)-2] <= '9'
	}

	return false
}

// validateIntValue validates a TypeInt value against the bounds declared in its
// metadata. The size-format pre-check carries a distinct message and is the
// reason TypeInt does not simply reuse a generic parser.
func validateIntValue(value string, meta ConfigMetadata) error {
	if isSizeFormat(value) {
		return fmt.Errorf("expected a number, got size format")
	}

	var n int
	if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
		return fmt.Errorf("expected a positive integer")
	}
	if meta.MinValue > 0 && n < meta.MinValue {
		if meta.MinValue == 1 {
			return fmt.Errorf("must be positive, got %d", n)
		}

		return fmt.Errorf("must be at least %d, got %d", meta.MinValue, n)
	}
	if meta.MaxValue > 0 && n > meta.MaxValue {
		return fmt.Errorf("too large (max %d), got %d", meta.MaxValue, n)
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

// validateEnumValue validates a TypeEnum value against the options declared in
// its metadata.
func validateEnumValue(value string, options []string) error {
	for _, opt := range options {
		if value == opt {
			return nil
		}
	}

	return fmt.Errorf("expected one of: %s, got '%s'", strings.Join(options, ", "), value)
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

// validateTimeoutValue validates a TypeTimeout value against the bounds declared
// in its metadata.
func validateTimeoutValue(value string, meta ConfigMetadata) error {
	var timeout int
	if _, err := fmt.Sscanf(value, "%d", &timeout); err != nil {
		return fmt.Errorf("expected a positive integer (seconds)")
	}
	if meta.MinValue > 0 && timeout < meta.MinValue {
		return fmt.Errorf("timeout must be positive, got %d", timeout)
	}
	if meta.MaxValue > 0 && timeout > meta.MaxValue {
		return fmt.Errorf("timeout too large (max %d seconds), got %d", meta.MaxValue, timeout)
	}

	return nil
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
