package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		configType ConfigType
		expected   string
	}{
		{TypeString, "string"},
		{TypeInt, "int"},
		{TypeBool, "bool"},
		{TypeSize, "size"},
		{TypePath, "path"},
		{TypeURL, "url"},
		{TypeEnum, "enum"},
		{TypeTimeout, "timeout"},
		{ConfigType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.configType.String())
		})
	}
}

func TestAllConfigMetadata_ReturnsAllKeys(t *testing.T) {
	t.Parallel()

	metadata := AllConfigMetadata()

	assert.NotEmpty(t, metadata)
	assert.Len(t, metadata, 28, "should have 28 configuration keys")
}

func TestAllConfigMetadata_MatchesValidKeys(t *testing.T) {
	t.Parallel()

	metadata := AllConfigMetadata()
	validKeys := ValidKeys()

	metadataKeys := make(map[string]bool)
	for _, m := range metadata {
		metadataKeys[m.Key] = true
	}

	for _, key := range validKeys {
		assert.True(t, metadataKeys[key], "metadata missing key: %s", key)
	}

	validKeysMap := make(map[string]bool)
	for _, key := range validKeys {
		validKeysMap[key] = true
	}

	for _, m := range metadata {
		assert.True(t, validKeysMap[m.Key], "metadata has unknown key: %s", m.Key)
	}
}

func TestAllConfigMetadata_NoDuplicateKeys(t *testing.T) {
	t.Parallel()

	metadata := AllConfigMetadata()
	seen := make(map[string]bool)

	for _, m := range metadata {
		assert.False(t, seen[m.Key], "duplicate key found: %s", m.Key)
		seen[m.Key] = true
	}
}

func TestAllConfigMetadata_AllHaveCategories(t *testing.T) {
	t.Parallel()

	metadata := AllConfigMetadata()

	for _, m := range metadata {
		assert.NotEmpty(t, m.Category, "key %s has no category", m.Key)
	}
}

func TestAllConfigMetadata_AllHaveDescriptions(t *testing.T) {
	t.Parallel()

	metadata := AllConfigMetadata()

	for _, m := range metadata {
		assert.NotEmpty(t, m.Description, "key %s has no description", m.Key)
	}
}

func TestAllConfigMetadata_AllHaveTypes(t *testing.T) {
	t.Parallel()

	metadata := AllConfigMetadata()
	validTypes := []ConfigType{
		TypeString, TypeInt, TypeBool, TypeSize,
		TypePath, TypeURL, TypeEnum, TypeTimeout,
	}

	for _, m := range metadata {
		found := false
		for _, vt := range validTypes {
			if m.Type == vt {
				found = true
				break
			}
		}
		assert.True(t, found, "key %s has invalid type: %v", m.Key, m.Type)
	}
}

func TestAllConfigMetadata_EnumTypesHaveOptions(t *testing.T) {
	t.Parallel()

	metadata := AllConfigMetadata()

	for _, m := range metadata {
		if m.Type == TypeEnum {
			assert.NotEmpty(t, m.EnumOptions, "enum key %s has no options", m.Key)
		}
	}
}

func TestAllConfigMetadata_IntTypesHaveRanges(t *testing.T) {
	t.Parallel()

	metadata := AllConfigMetadata()

	for _, m := range metadata {
		if m.Type == TypeInt || m.Type == TypeTimeout {
			if m.MinValue != 0 || m.MaxValue != 0 {
				assert.Less(t, m.MinValue, m.MaxValue,
					"key %s has invalid range: min=%d, max=%d", m.Key, m.MinValue, m.MaxValue)
			}
		}
	}
}

func TestGetMetadata_ExistingKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key      string
		category ConfigCategory
		cfgType  ConfigType
	}{
		{KeyScannerMaxFiles, CategoryScanner, TypeInt},
		{KeyScannerMaxFileSize, CategoryScanner, TypeSize},
		{KeyScannerWorkers, CategoryScanner, TypeInt},
		{KeyContextIncludeTree, CategoryContext, TypeBool},
		{KeyTemplateCustomPath, CategoryTemplate, TypePath},
		{KeyOutputFormat, CategoryOutput, TypeEnum},
		{KeyLLMProvider, CategoryLLM, TypeEnum},
		{KeyLLMTimeout, CategoryLLM, TypeTimeout},
		{KeyGeminiEnabled, CategoryGemini, TypeBool},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			m, found := GetMetadata(tt.key)
			require.True(t, found, "key not found: %s", tt.key)
			assert.Equal(t, tt.category, m.Category)
			assert.Equal(t, tt.cfgType, m.Type)
		})
	}
}

func TestGetMetadata_NonExistingKey(t *testing.T) {
	t.Parallel()

	_, found := GetMetadata("nonexistent.key")
	assert.False(t, found)
}

func TestGetByCategory_ReturnsCorrectKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		category      ConfigCategory
		expectedCount int
		expectedKeys  []string
	}{
		{CategoryScanner, 9, []string{KeyScannerMaxFiles, KeyScannerWorkers}},
		{CategoryContext, 3, []string{KeyContextIncludeTree, KeyContextMaxSize}},
		{CategoryTemplate, 1, []string{KeyTemplateCustomPath}},
		{CategoryOutput, 2, []string{KeyOutputFormat, KeyOutputClipboard}},
		{CategoryLLM, 6, []string{KeyLLMProvider, KeyLLMAPIKey}},
		{CategoryGemini, 7, []string{KeyGeminiEnabled, KeyGeminiModel}},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			t.Parallel()
			result := GetByCategory(tt.category)
			assert.Len(t, result, tt.expectedCount)

			keys := make(map[string]bool)
			for _, m := range result {
				keys[m.Key] = true
				assert.Equal(t, tt.category, m.Category)
			}

			for _, expectedKey := range tt.expectedKeys {
				assert.True(t, keys[expectedKey], "missing expected key: %s", expectedKey)
			}
		})
	}
}

func TestGetByCategory_NonExistingCategory(t *testing.T) {
	t.Parallel()

	result := GetByCategory("NonExisting")
	assert.Empty(t, result)
}

func TestAllCategories_ReturnsAllCategories(t *testing.T) {
	t.Parallel()

	categories := AllCategories()

	assert.Len(t, categories, 6)
	assert.Equal(t, CategoryScanner, categories[0])
	assert.Equal(t, CategoryContext, categories[1])
	assert.Equal(t, CategoryTemplate, categories[2])
	assert.Equal(t, CategoryOutput, categories[3])
	assert.Equal(t, CategoryLLM, categories[4])
	assert.Equal(t, CategoryGemini, categories[5])
}

func TestAllCategories_CoversAllMetadata(t *testing.T) {
	t.Parallel()

	categories := AllCategories()
	metadata := AllConfigMetadata()

	categorySet := make(map[ConfigCategory]bool)
	for _, c := range categories {
		categorySet[c] = true
	}

	for _, m := range metadata {
		assert.True(t, categorySet[m.Category],
			"category %s not in AllCategories()", m.Category)
	}
}

func TestMetadataDefaults_MatchRootDefaults(t *testing.T) {
	t.Parallel()

	expectedDefaults := map[string]interface{}{
		KeyScannerMaxFiles:             10000,
		KeyScannerMaxFileSize:          "1MB",
		KeyScannerMaxMemory:            "500MB",
		KeyScannerSkipBinary:           true,
		KeyScannerIncludeHidden:        false,
		KeyScannerIncludeIgnored:       false,
		KeyScannerWorkers:              1,
		KeyScannerRespectGitignore:     true,
		KeyScannerRespectShotgunignore: true,
		KeyContextIncludeTree:          true,
		KeyContextIncludeSummary:       true,
		KeyContextMaxSize:              "10MB",
		KeyTemplateCustomPath:          "",
		KeyOutputFormat:                "markdown",
		KeyOutputClipboard:             true,
		KeyLLMProvider:                 "geminiweb",
		KeyLLMAPIKey:                   "",
		KeyLLMBaseURL:                  "",
		KeyLLMModel:                    "",
		KeyLLMTimeout:                  300,
		KeyLLMSaveResponse:             false,
		KeyGeminiEnabled:               false,
		KeyGeminiModel:                 "gemini-2.5-flash",
		KeyGeminiTimeout:               300,
		KeyGeminiBinaryPath:            "",
		KeyGeminiBrowserRefresh:        "auto",
		KeyGeminiAutoSend:              false,
		KeyGeminiSaveResponse:          true,
	}

	for key, expectedDefault := range expectedDefaults {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			m, found := GetMetadata(key)
			require.True(t, found, "key not found: %s", key)
			assert.Equal(t, expectedDefault, m.DefaultValue,
				"default mismatch for key %s", key)
		})
	}
}

func TestMetadataEnumOptions_MatchValidators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key     string
		options []string
	}{
		{KeyOutputFormat, []string{"markdown", "text"}},
		{KeyLLMProvider, []string{"openai", "anthropic", "gemini", "geminiweb"}},
		{KeyGeminiBrowserRefresh, []string{"", "auto", "chrome", "firefox", "edge", "chromium", "opera"}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			m, found := GetMetadata(tt.key)
			require.True(t, found)
			assert.Equal(t, tt.options, m.EnumOptions)
		})
	}
}

func TestMetadataRanges_MatchValidators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key string
		min int
		max int
	}{
		{KeyScannerWorkers, 1, 32},
		{KeyScannerMaxFiles, 1, 1000000},
		{KeyLLMTimeout, 1, 3600},
		{KeyGeminiTimeout, 1, 3600},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			m, found := GetMetadata(tt.key)
			require.True(t, found)
			assert.Equal(t, tt.min, m.MinValue)
			assert.Equal(t, tt.max, m.MaxValue)
		})
	}
}
