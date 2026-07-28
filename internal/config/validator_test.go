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

// TestValidKeys_MatchesRegistry pins CI-009's contract: ValidKeys() is derived
// from the metadata registry, so registering a key is the only edit needed to
// make it settable. Before this, the two lists were maintained by hand and a key
// had to appear in both to work at all.
func TestValidKeys_MatchesRegistry(t *testing.T) {
	t.Parallel()

	metadata := AllConfigMetadata()
	want := make([]string, 0, len(metadata))
	for _, m := range metadata {
		want = append(want, m.Key)
	}

	got := ValidKeys()
	if len(got) != len(want) {
		t.Fatalf("ValidKeys() has %d keys, registry has %d", len(got), len(want))
	}

	inRegistry := make(map[string]bool, len(want))
	for _, k := range want {
		inRegistry[k] = true
	}
	for _, k := range got {
		if !inRegistry[k] {
			t.Errorf("ValidKeys() returned %q, which is not in the metadata registry", k)
		}
	}
	for _, k := range want {
		if !IsValidKey(k) {
			t.Errorf("registry key %q is not reported valid by IsValidKey", k)
		}
	}
}

func TestDeprecationMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key            string
		wantDeprecated bool
	}{
		{"scanner.workers", true},
		{"scanner.max-memory", true},
		{KeyScannerMaxFiles, false},
		{"nonsense.key", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			msg, deprecated := DeprecationMessage(tt.key)
			if deprecated != tt.wantDeprecated {
				t.Fatalf("DeprecationMessage(%q) deprecated = %v, want %v", tt.key, deprecated, tt.wantDeprecated)
			}
			if deprecated && msg == "" {
				t.Errorf("DeprecationMessage(%q) reported deprecated with an empty message", tt.key)
			}
		})
	}
}

// TestDeprecatedKeysAreNotValid guards the invariant that a retired key is gone
// from the key set, so `config set` reaches the deprecation branch rather than
// silently writing a key nothing reads.
func TestDeprecatedKeysAreNotValid(t *testing.T) {
	t.Parallel()

	for key := range deprecatedKeys {
		if IsValidKey(key) {
			t.Errorf("deprecated key %q is still reported as valid", key)
		}
		if _, found := GetMetadata(key); found {
			t.Errorf("deprecated key %q is still in the metadata registry", key)
		}
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
			err := ValidateValue(KeyScannerSkipBinary, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue(scanner.skip-binary, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
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
			val, err := ConvertValue(KeyScannerSkipBinary, tt.value)
			if err != nil {
				t.Fatalf("ConvertValue failed: %v", err)
			}
			if val != tt.want {
				t.Errorf("ConvertValue(scanner.skip-binary, %q) = %v, want %v", tt.value, val, tt.want)
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
