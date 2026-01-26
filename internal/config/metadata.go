// Package config provides centralized configuration metadata.
package config

// ConfigType represents the type of a configuration value.
type ConfigType int

const (
	// TypeString represents a free-form string value.
	TypeString ConfigType = iota
	// TypeInt represents an integer value.
	TypeInt
	// TypeBool represents a boolean value.
	TypeBool
	// TypeSize represents a size format value (e.g., '10MB', '500KB').
	TypeSize
	// TypePath represents a file system path.
	TypePath
	// TypeURL represents a URL value.
	TypeURL
	// TypeEnum represents a value from a predefined set.
	TypeEnum
	// TypeTimeout represents an integer timeout in seconds with range validation.
	TypeTimeout
)

// String returns the string representation of ConfigType.
func (t ConfigType) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeInt:
		return "int"
	case TypeBool:
		return "bool"
	case TypeSize:
		return "size"
	case TypePath:
		return "path"
	case TypeURL:
		return "url"
	case TypeEnum:
		return "enum"
	case TypeTimeout:
		return "timeout"
	default:
		return "unknown"
	}
}

// ConfigCategory represents a logical grouping of configuration keys.
type ConfigCategory string

const (
	// CategoryScanner groups file scanning configuration.
	CategoryScanner ConfigCategory = "Scanner"
	// CategoryContext groups context generation configuration.
	CategoryContext ConfigCategory = "Context"
	// CategoryTemplate groups template configuration.
	CategoryTemplate ConfigCategory = "Template"
	// CategoryOutput groups output configuration.
	CategoryOutput ConfigCategory = "Output"
	// CategoryLLM groups LLM provider configuration.
	CategoryLLM ConfigCategory = "LLM Provider"
)

// ConfigMetadata describes a single configuration key.
type ConfigMetadata struct {
	// Key is the configuration key (e.g., "scanner.max-files").
	Key string
	// Category is the logical grouping for this key.
	Category ConfigCategory
	// Type is the value type for validation and UI rendering.
	Type ConfigType
	// Description is a human-readable description of the key.
	Description string
	// DefaultValue is the default value for this key.
	DefaultValue interface{}
	// EnumOptions lists valid values for TypeEnum keys.
	EnumOptions []string
	// MinValue is the minimum value for TypeInt/TypeTimeout keys.
	MinValue int
	// MaxValue is the maximum value for TypeInt/TypeTimeout keys.
	MaxValue int
	// Required indicates if this key must have a non-empty value.
	Required bool
}

// allMetadata holds all configuration metadata, built once at init.
var allMetadata []ConfigMetadata

func init() {
	allMetadata = buildAllMetadata()
}

// AllConfigMetadata returns all configuration metadata.
func AllConfigMetadata() []ConfigMetadata {
	return allMetadata
}

// GetMetadata returns metadata for a specific key.
func GetMetadata(key string) (ConfigMetadata, bool) {
	for _, m := range allMetadata {
		if m.Key == key {
			return m, true
		}
	}
	return ConfigMetadata{}, false
}

// GetByCategory returns all metadata for a specific category.
func GetByCategory(category ConfigCategory) []ConfigMetadata {
	var result []ConfigMetadata
	for _, m := range allMetadata {
		if m.Category == category {
			result = append(result, m)
		}
	}
	return result
}

// AllCategories returns all categories in display order.
func AllCategories() []ConfigCategory {
	return []ConfigCategory{
		CategoryScanner,
		CategoryContext,
		CategoryTemplate,
		CategoryOutput,
		CategoryLLM,
	}
}

// buildAllMetadata constructs the complete metadata list.
func buildAllMetadata() []ConfigMetadata {
	return []ConfigMetadata{
		// Scanner (9 keys)
		{
			Key:          KeyScannerMaxFiles,
			Category:     CategoryScanner,
			Type:         TypeInt,
			Description:  "Maximum number of files to scan",
			DefaultValue: 10000,
			MinValue:     1,
			MaxValue:     1000000,
		},
		{
			Key:          KeyScannerMaxFileSize,
			Category:     CategoryScanner,
			Type:         TypeSize,
			Description:  "Maximum size per file (e.g., 10MB, 500KB)",
			DefaultValue: "1MB",
		},
		{
			Key:          KeyScannerMaxMemory,
			Category:     CategoryScanner,
			Type:         TypeSize,
			Description:  "Maximum memory usage for scanning",
			DefaultValue: "500MB",
		},
		{
			Key:          KeyScannerSkipBinary,
			Category:     CategoryScanner,
			Type:         TypeBool,
			Description:  "Skip binary files during scanning",
			DefaultValue: true,
		},
		{
			Key:          KeyScannerIncludeHidden,
			Category:     CategoryScanner,
			Type:         TypeBool,
			Description:  "Include hidden files (starting with .)",
			DefaultValue: false,
		},
		{
			Key:          KeyScannerIncludeIgnored,
			Category:     CategoryScanner,
			Type:         TypeBool,
			Description:  "Include git-ignored files",
			DefaultValue: false,
		},
		{
			Key:          KeyScannerWorkers,
			Category:     CategoryScanner,
			Type:         TypeInt,
			Description:  "Number of parallel scanner workers",
			DefaultValue: 1,
			MinValue:     1,
			MaxValue:     32,
		},
		{
			Key:          KeyScannerRespectGitignore,
			Category:     CategoryScanner,
			Type:         TypeBool,
			Description:  "Respect .gitignore files during scanning",
			DefaultValue: true,
		},
		{
			Key:          KeyScannerRespectShotgunignore,
			Category:     CategoryScanner,
			Type:         TypeBool,
			Description:  "Respect .shotgunignore files during scanning",
			DefaultValue: true,
		},

		// Context (3 keys)
		{
			Key:          KeyContextIncludeTree,
			Category:     CategoryContext,
			Type:         TypeBool,
			Description:  "Include file tree in generated context",
			DefaultValue: true,
		},
		{
			Key:          KeyContextIncludeSummary,
			Category:     CategoryContext,
			Type:         TypeBool,
			Description:  "Include file summary in generated context",
			DefaultValue: true,
		},
		{
			Key:          KeyContextMaxSize,
			Category:     CategoryContext,
			Type:         TypeSize,
			Description:  "Maximum size of generated context",
			DefaultValue: "10MB",
		},

		// Template (1 key)
		{
			Key:          KeyTemplateCustomPath,
			Category:     CategoryTemplate,
			Type:         TypePath,
			Description:  "Custom path to template directory",
			DefaultValue: "",
		},

		// Output (2 keys)
		{
			Key:          KeyOutputFormat,
			Category:     CategoryOutput,
			Type:         TypeEnum,
			Description:  "Output format for generated context",
			DefaultValue: "markdown",
			EnumOptions:  []string{"markdown", "text"},
		},
		{
			Key:          KeyOutputClipboard,
			Category:     CategoryOutput,
			Type:         TypeBool,
			Description:  "Copy generated context to clipboard",
			DefaultValue: true,
		},

		// LLM Provider (6 keys)
		{
			Key:          KeyLLMProvider,
			Category:     CategoryLLM,
			Type:         TypeEnum,
			Description:  "LLM provider to use",
			DefaultValue: "gemini",
			EnumOptions:  []string{"openai", "anthropic", "gemini"},
		},
		{
			Key:          KeyLLMAPIKey,
			Category:     CategoryLLM,
			Type:         TypeString,
			Description:  "API key for the LLM provider",
			DefaultValue: "",
		},
		{
			Key:          KeyLLMBaseURL,
			Category:     CategoryLLM,
			Type:         TypeURL,
			Description:  "Custom base URL for API requests",
			DefaultValue: "",
		},
		{
			Key:          KeyLLMModel,
			Category:     CategoryLLM,
			Type:         TypeString,
			Description:  "Model name to use",
			DefaultValue: "",
		},
		{
			Key:          KeyLLMTimeout,
			Category:     CategoryLLM,
			Type:         TypeTimeout,
			Description:  "Request timeout in seconds",
			DefaultValue: 300,
			MinValue:     1,
			MaxValue:     3600,
		},
		{
			Key:          KeyLLMSaveResponse,
			Category:     CategoryLLM,
			Type:         TypeBool,
			Description:  "Save LLM response to file",
			DefaultValue: false,
		},
	}
}
