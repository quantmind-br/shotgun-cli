package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestIsValidConfigKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		// Valid keys
		{"scanner.max-files", "scanner.max-files", true},
		{"scanner.max-file-size", "scanner.max-file-size", true},
		{"scanner.respect-gitignore", "scanner.respect-gitignore", true},
		{"scanner.skip-binary", "scanner.skip-binary", true},
		{"context.max-size", "context.max-size", true},
		{"context.include-tree", "context.include-tree", true},
		{"context.include-summary", "context.include-summary", true},
		{"template.custom-path", "template.custom-path", true},
		{"output.format", "output.format", true},
		{"output.clipboard", "output.clipboard", true},

		// Invalid keys
		{"empty key", "", false},
		{"unknown key", "unknown.key", false},
		{"partial key", "scanner", false},
		{"typo in key", "scannner.max-files", false},
		{"case sensitive", "Scanner.max-files", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.IsValidKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateMaxFiles(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"valid positive integer", "1000", false},
		{"valid single digit", "1", false},
		{"valid large number", "999999", false},
		{"zero", "0", true},
		{"negative number", "-5", true},
		{"not a number", "abc", true},
		{"float number truncated", "10.5", false}, // Sscanf with %d truncates to 10
		{"size format rejected", "10MB", true},    // Size formats should be rejected
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateValue(config.KeyScannerMaxFiles, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSizeFormat(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"megabytes", "10MB", false},
		{"kilobytes", "500KB", false},
		{"gigabytes", "1GB", false},
		{"bytes", "1024B", false},
		{"lowercase", "10mb", false},
		{"mixed case", "10Mb", false},
		{"with decimal", "1.5MB", false},
		{"plain number", "1024", false},
		{"invalid unit", "10TB", true},
		{"no number", "MB", true},
		{"invalid format", "abc", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateValue(config.KeyScannerMaxFileSize, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBooleanValue(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"true lowercase", "true", false},
		{"false lowercase", "false", false},
		{"TRUE uppercase", "TRUE", false},
		{"FALSE uppercase", "FALSE", false},
		{"True mixed", "True", false},
		{"False mixed", "False", false},
		{"yes", "yes", true},
		{"no", "no", true},
		{"1", "1", true},
		{"0", "0", true},
		{"empty", "", true},
		{"random", "maybe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateValue(config.KeyGeminiEnabled, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOutputFormat(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"markdown", "markdown", false},
		{"text", "text", false},
		{"json", "json", true},
		{"html", "html", true},
		{"empty", "", true},
		{"uppercase MARKDOWN", "MARKDOWN", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateValue(config.KeyOutputFormat, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTemplatePath(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()

	// Create a file (not directory) for testing
	filePath := filepath.Join(tmpDir, "afile.txt")
	_ = os.WriteFile(filePath, []byte("test"), 0o600)

	tests := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"empty path", "", false},
		{"absolute path", "/some/path/to/templates", false},
		{"path with tilde", "~/templates", false},
		{"relative path", "templates", false},
		{"existing directory as parent", tmpDir + "/new-templates", false},
		{"parent is file not directory", filePath + "/templates", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateValue(config.KeyTemplateCustomPath, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfigValue(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		expectErr bool
	}{
		{"max-files valid", "scanner.max-files", "1000", false},
		{"max-files invalid", "scanner.max-files", "abc", true},
		{"max-file-size valid", "scanner.max-file-size", "10MB", false},
		{"max-file-size invalid", "scanner.max-file-size", "abc", true},
		{"respect-gitignore valid", "scanner.respect-gitignore", "true", false},
		{"respect-gitignore invalid", "scanner.respect-gitignore", "yes", true},
		{"skip-binary valid", "scanner.skip-binary", "false", false},
		{"context.max-size valid", "context.max-size", "5MB", false},
		{"context.include-tree valid", "context.include-tree", "true", false},
		{"context.include-summary valid", "context.include-summary", "false", false},
		{"output.format valid", "output.format", "markdown", false},
		{"output.format invalid", "output.format", "xml", true},
		{"output.clipboard valid", "output.clipboard", "true", false},
		{"template.custom-path empty", "template.custom-path", "", false},
		{"unknown key", "unknown.key", "value", false}, // unknown keys pass through
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateValue(tt.key, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConvertConfigValue(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		expectedValue interface{}
		expectErr     bool
	}{
		{"max-files integer", "scanner.max-files", "1000", 1000, false},
		{"max-files invalid", "scanner.max-files", "abc", nil, true},
		{"respect-gitignore true", "scanner.respect-gitignore", "true", true, false},
		{"respect-gitignore false", "scanner.respect-gitignore", "false", false, false},
		{"skip-binary TRUE", "scanner.skip-binary", "TRUE", true, false},
		{"include-tree", "context.include-tree", "true", true, false},
		{"include-summary", "context.include-summary", "false", false, false},
		{"clipboard", "output.clipboard", "true", true, false},
		{"string value", "output.format", "markdown", "markdown", false},
		{"string path", "template.custom-path", "/path/to/templates", "/path/to/templates", false},
		{"size format preserved", "scanner.max-file-size", "10MB", "10MB", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := config.ConvertValue(tt.key, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, result)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil value", nil, "<nil>"},
		{"empty string", "", `""`},
		{"non-empty string", "hello", `"hello"`},
		{"true bool", true, "true"},
		{"false bool", false, "false"},
		{"integer", 42, "42"},
		{"float", 3.14, "3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetConfigSource(t *testing.T) {
	// This test is limited because getConfigSource depends on viper state
	// We test the basic behavior when viper is not configured

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"unknown key returns default", "unknown.key", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getConfigSource(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
