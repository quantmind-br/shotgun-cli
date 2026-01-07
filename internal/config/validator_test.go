package config

import (
	"os"
	"testing"
)

func TestIsValidKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  string
		want bool
	}{
		{KeyScannerMaxFiles, true},
		{KeyLLMProvider, true},
		{KeyGeminiEnabled, true},
		{KeyTemplateCustomPath, true},
		{"invalid.key", false},
		{"", false},
		{"scanner", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			if got := IsValidKey(tt.key); got != tt.want {
				t.Errorf("IsValidKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestValidKeys(t *testing.T) {
	t.Parallel()

	keys := ValidKeys()
	if len(keys) == 0 {
		t.Error("ValidKeys() returned empty slice")
	}

	seen := make(map[string]bool)
	for _, key := range keys {
		if seen[key] {
			t.Errorf("ValidKeys() contains duplicate: %s", key)
		}
		seen[key] = true
	}
}

func TestValidateValue_Workers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"1", false},
		{"4", false},
		{"32", false},
		{"0", true},
		{"33", true},
		{"-1", true},
		{"abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyScannerWorkers, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(workers, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue_MaxFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"100", false},
		{"10000", false},
		{"0", true},
		{"-1", true},
		{"10MB", true},
		{"1KB", true},
		{"abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyScannerMaxFiles, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(max-files, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue_SizeFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"1MB", false},
		{"500KB", false},
		{"1GB", false},
		{"100", false},
		{"abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyScannerMaxFileSize, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(max-file-size, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue_Boolean(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"true", false},
		{"false", false},
		{"TRUE", false},
		{"FALSE", false},
		{"yes", true},
		{"no", true},
		{"1", true},
		{"0", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyGeminiEnabled, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(gemini.enabled, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue_LLMProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"openai", false},
		{"anthropic", false},
		{"gemini", false},
		{"geminiweb", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyLLMProvider, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(llm.provider, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue_OutputFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"markdown", false},
		{"text", false},
		{"json", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyOutputFormat, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(output.format, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue_Timeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"60", false},
		{"300", false},
		{"3600", false},
		{"0", true},
		{"-1", true},
		{"3601", true},
		{"abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyLLMTimeout, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(llm.timeout, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue_URL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"", false},
		{"https://api.openai.com/v1", false},
		{"http://localhost:8080", false},
		{"ftp://invalid", true},
		{"not-a-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyLLMBaseURL, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(llm.base-url, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestConvertValue_Integer(t *testing.T) {
	t.Parallel()

	val, err := ConvertValue(KeyScannerMaxFiles, "1000")
	if err != nil {
		t.Fatalf("ConvertValue failed: %v", err)
	}
	if val != 1000 {
		t.Errorf("ConvertValue(max-files, \"1000\") = %v, want 1000", val)
	}
}

func TestConvertValue_Boolean(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		want  bool
	}{
		{"true", true},
		{"TRUE", true},
		{"false", false},
		{"FALSE", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			val, err := ConvertValue(KeyGeminiEnabled, tt.value)
			if err != nil {
				t.Fatalf("ConvertValue failed: %v", err)
			}
			if val != tt.want {
				t.Errorf("ConvertValue(gemini.enabled, %q) = %v, want %v", tt.value, val, tt.want)
			}
		})
	}
}

func TestConvertValue_String(t *testing.T) {
	t.Parallel()

	val, err := ConvertValue(KeyLLMProvider, "openai")
	if err != nil {
		t.Fatalf("ConvertValue failed: %v", err)
	}
	if val != "openai" {
		t.Errorf("ConvertValue(llm.provider, \"openai\") = %v, want \"openai\"", val)
	}
}

func TestValidateValue_BrowserRefresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		// Valid values
		{"", false},
		{"auto", false},
		{"chrome", false},
		{"firefox", false},
		{"edge", false},
		{"chromium", false},
		{"opera", false},
		// Invalid values
		{"safari", true},
		{"brave", true},
		{"invalid", true},
		{"Chrome", true},  // case sensitive
		{"FIREFOX", true}, // case sensitive
		{"AUTO", true},    // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyGeminiBrowserRefresh, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(gemini.browser-refresh, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateBrowserRefresh_Direct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		// Valid values
		{"", false},
		{"auto", false},
		{"chrome", false},
		{"firefox", false},
		{"edge", false},
		{"chromium", false},
		{"opera", false},
		// Invalid values
		{"safari", true},
		{"brave", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := validateBrowserRefresh(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBrowserRefresh(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Empty value should be valid
		{"empty string", "", false},
		// Valid paths
		{"tmp", "/tmp", false},
		{"home", "/home/user", false},
		{"documents", "~/Documents", false},
		// Non-existent parent directory is ok (will be created)
		{"nonexistent", "/nonexistent/dir/config.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validatePath(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePath_ExistingFile(t *testing.T) {
	t.Parallel()

	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	// Parent directory exists, so no error
	err = validatePath(tmpFile.Name())
	if err != nil {
		t.Errorf("validatePath(%q) should not error, got: %v", tmpFile.Name(), err)
	}
}

func TestValidateGeminiModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		// Valid models
		{"gemini-2.5-flash", false},
		{"gemini-2.5-pro", false},
		{"gemini-3.0-pro", false},
		// Invalid models
		{"gemini-1.5-pro", true},
		{"gemini-1.5-flash", true},
		{"gpt-4", true},
		{"claude-3", true},
		{"", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := validateGeminiModel(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGeminiModel(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue_Path(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"", false},
		{"/tmp", false},
		{"/home/user/config.yaml", false},
		{"~/Documents", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyTemplateCustomPath, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(template-custom-path, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue_GeminiModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"gemini-2.5-flash", false},
		{"gemini-2.5-pro", false},
		{"gemini-3.0-pro", false},
		{"gemini-1.5-pro", true},
		{"gpt-4", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			err := ValidateValue(KeyGeminiModel, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(gemini.model, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}
