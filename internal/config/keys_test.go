package config

import (
	"strings"
	"testing"
)

func TestNoDuplicateKeyValues(t *testing.T) {
	keys := getAllKeyValues()
	seen := make(map[string]string)

	for name, value := range keys {
		if existing, ok := seen[value]; ok {
			t.Errorf("duplicate key value %q used by both %s and %s", value, existing, name)
		}
		seen[value] = name
	}
}

func TestKeyNamingConvention(t *testing.T) {
	keys := getAllKeyValues()

	for name := range keys {
		if !strings.HasPrefix(name, "Key") {
			t.Errorf("constant %s should start with 'Key' prefix", name)
		}
	}
}

func TestKeyValueFormat(t *testing.T) {
	keys := getAllKeyValues()

	for name, value := range keys {
		if value != "verbose" && value != "quiet" && !strings.Contains(value, ".") {
			t.Errorf("key %s has value %q which doesn't contain a dot separator", name, value)
		}
	}
}

func TestAllKeysDocumented(t *testing.T) {
	keys := getAllKeyValues()

	if len(keys) < 25 {
		t.Errorf("expected at least 25 configuration keys, got %d", len(keys))
	}
}

func TestScannerKeysExist(t *testing.T) {
	expected := []string{
		KeyScannerMaxFiles,
		KeyScannerMaxFileSize,
		KeyScannerMaxMemory,
		KeyScannerSkipBinary,
		KeyScannerIncludeHidden,
		KeyScannerWorkers,
		KeyScannerRespectGitignore,
		KeyScannerRespectShotgunignore,
	}

	for _, key := range expected {
		if key == "" {
			t.Error("scanner key constant is empty")
		}
		if !strings.HasPrefix(key, "scanner.") {
			t.Errorf("scanner key %q should start with 'scanner.'", key)
		}
	}
}

func TestLLMKeysExist(t *testing.T) {
	expected := []string{
		KeyLLMProvider,
		KeyLLMAPIKey,
		KeyLLMBaseURL,
		KeyLLMModel,
		KeyLLMTimeout,
	}

	for _, key := range expected {
		if key == "" {
			t.Error("LLM key constant is empty")
		}
		if !strings.HasPrefix(key, "llm.") {
			t.Errorf("LLM key %q should start with 'llm.'", key)
		}
	}
}

func TestGeminiKeysExist(t *testing.T) {
	expected := []string{
		KeyGeminiEnabled,
		KeyGeminiModel,
		KeyGeminiTimeout,
		KeyGeminiBinaryPath,
		KeyGeminiBrowserRefresh,
		KeyGeminiAutoSend,
		KeyGeminiSaveResponse,
	}

	for _, key := range expected {
		if key == "" {
			t.Error("gemini key constant is empty")
		}
		if !strings.HasPrefix(key, "gemini.") {
			t.Errorf("gemini key %q should start with 'gemini.'", key)
		}
	}
}

func getAllKeyValues() map[string]string {
	return map[string]string{
		"KeyScannerMaxFiles":             KeyScannerMaxFiles,
		"KeyScannerMaxFileSize":          KeyScannerMaxFileSize,
		"KeyScannerMaxMemory":            KeyScannerMaxMemory,
		"KeyScannerSkipBinary":           KeyScannerSkipBinary,
		"KeyScannerIncludeHidden":        KeyScannerIncludeHidden,
		"KeyScannerIncludeIgnored":       KeyScannerIncludeIgnored,
		"KeyScannerWorkers":              KeyScannerWorkers,
		"KeyScannerRespectGitignore":     KeyScannerRespectGitignore,
		"KeyScannerRespectShotgunignore": KeyScannerRespectShotgunignore,
		"KeyLLMProvider":                 KeyLLMProvider,
		"KeyLLMAPIKey":                   KeyLLMAPIKey,
		"KeyLLMBaseURL":                  KeyLLMBaseURL,
		"KeyLLMModel":                    KeyLLMModel,
		"KeyLLMTimeout":                  KeyLLMTimeout,
		"KeyGeminiEnabled":               KeyGeminiEnabled,
		"KeyGeminiModel":                 KeyGeminiModel,
		"KeyGeminiTimeout":               KeyGeminiTimeout,
		"KeyGeminiBinaryPath":            KeyGeminiBinaryPath,
		"KeyGeminiBrowserRefresh":        KeyGeminiBrowserRefresh,
		"KeyGeminiAutoSend":              KeyGeminiAutoSend,
		"KeyGeminiSaveResponse":          KeyGeminiSaveResponse,
		"KeyContextIncludeTree":          KeyContextIncludeTree,
		"KeyContextIncludeSummary":       KeyContextIncludeSummary,
		"KeyContextMaxSize":              KeyContextMaxSize,
		"KeyTemplateCustomPath":          KeyTemplateCustomPath,
		"KeyOutputFormat":                KeyOutputFormat,
		"KeyOutputClipboard":             KeyOutputClipboard,
		"KeyVerbose":                     KeyVerbose,
		"KeyQuiet":                       KeyQuiet,
	}
}
