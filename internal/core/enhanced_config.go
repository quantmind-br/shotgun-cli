package core

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// EnhancedConfig represents the enhanced application configuration with validation
type EnhancedConfig struct {
	OpenAI      EnhancedOpenAIConfig      `koanf:"openai" validate:"required"`
	Translation EnhancedTranslationConfig `koanf:"translation" validate:"required"`
	App         EnhancedAppConfig         `koanf:"app" validate:"required"`
	Version     string                    `koanf:"version"`
	LastUpdated time.Time                 `koanf:"lastUpdated"`
}

// EnhancedOpenAIConfig contains enhanced OpenAI API configuration with validation
type EnhancedOpenAIConfig struct {
	// API Key (stored in keyring, this holds reference/alias)
	APIKeyAlias string `koanf:"apiKeyAlias" validate:"omitempty,min=1"`

	// API Configuration
	BaseURL string `koanf:"baseUrl" validate:"required,url"`
	Model   string `koanf:"model" validate:"required,oneof=gpt-4o gpt-4o-mini gpt-4-turbo gpt-4 gpt-3.5-turbo gpt-3.5-turbo-16k"`

	// Request Settings
	Timeout     int     `koanf:"timeout" validate:"min=1,max=3600"`     // Seconds, Range: 1-3600 (1 hour)
	MaxTokens   int     `koanf:"maxTokens" validate:"min=1,max=128000"` // Range: 1-128000
	Temperature float64 `koanf:"temperature" validate:"min=0,max=2"`    // Range: 0.0-2.0

	// Retry Settings
	MaxRetries int `koanf:"maxRetries" validate:"min=0,max=10"` // Range: 0-10 retries
	RetryDelay int `koanf:"retryDelay" validate:"min=0,max=60"` // Seconds, Range: 0-60
}

// EnhancedTranslationConfig contains enhanced translation settings with validation
type EnhancedTranslationConfig struct {
	Enabled        bool   `koanf:"enabled"`                                                                // Enable/disable translation
	TargetLanguage string `koanf:"targetLanguage" validate:"required,oneof=en es fr de it pt ru ja ko zh"` // Supported language codes
	ContextPrompt  string `koanf:"contextPrompt" validate:"omitempty,max=1000"`                            // Max 1000 chars for context
	CacheEnabled   bool   `koanf:"cacheEnabled"`                                                           // Enable translation caching
	CacheSize      int    `koanf:"cacheSize" validate:"min=10,max=10000"`                                  // Cache size: 10-10000 entries
	CacheTTL       int    `koanf:"cacheTTL" validate:"min=300,max=86400"`                                  // Cache TTL: 5min-24h in seconds
}

// EnhancedAppConfig contains enhanced application preferences with validation
type EnhancedAppConfig struct {
	Theme           string `koanf:"theme" validate:"required,oneof=auto dark light"`                               // UI theme
	AutoSave        bool   `koanf:"autoSave"`                                                                      // Auto-save configuration
	ShowLineNumbers bool   `koanf:"showLineNumbers"`                                                               // Show line numbers in UI
	DefaultTemplate string `koanf:"defaultTemplate" validate:"required,oneof=dev architect debug project-manager"` // Default prompt template

	// Performance Settings
	MaxFileSize       int64 `koanf:"maxFileSize" validate:"min=1024,max=104857600"` // 1KB - 100MB
	MaxDirectoryDepth int   `koanf:"maxDirectoryDepth" validate:"min=1,max=20"`     // 1-20 levels
	WorkerPoolSize    int   `koanf:"workerPoolSize" validate:"min=1,max=50"`        // 1-50 workers

	// UI Settings
	RefreshInterval int  `koanf:"refreshInterval" validate:"min=100,max=10000"` // 100ms - 10s
	EnableHotReload bool `koanf:"enableHotReload"`                              // Enable hot config reload
}

// ValidationError represents a configuration validation error with field context
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ConfigValidationResult contains validation results with detailed error information
type ConfigValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// DefaultEnhancedConfig returns the default enhanced configuration
func DefaultEnhancedConfig() *EnhancedConfig {
	return &EnhancedConfig{
		OpenAI:      DefaultEnhancedOpenAIConfig(),
		Translation: DefaultEnhancedTranslationConfig(),
		App:         DefaultEnhancedAppConfig(),
		Version:     "1.0.0",
		LastUpdated: time.Now(),
	}
}

// DefaultEnhancedOpenAIConfig returns default enhanced OpenAI configuration
func DefaultEnhancedOpenAIConfig() EnhancedOpenAIConfig {
	return EnhancedOpenAIConfig{
		APIKeyAlias: "",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4o",
		Timeout:     300,
		MaxTokens:   4096,
		Temperature: 0.7,
		MaxRetries:  3,
		RetryDelay:  2,
	}
}

// DefaultEnhancedTranslationConfig returns default enhanced translation configuration
func DefaultEnhancedTranslationConfig() EnhancedTranslationConfig {
	return EnhancedTranslationConfig{
		Enabled:        false,
		TargetLanguage: "en",
		ContextPrompt:  "Translate the following text to English, preserving technical terms and maintaining the original meaning:",
		CacheEnabled:   true,
		CacheSize:      1000,
		CacheTTL:       3600, // 1 hour
	}
}

// DefaultEnhancedAppConfig returns default enhanced application configuration
func DefaultEnhancedAppConfig() EnhancedAppConfig {
	return EnhancedAppConfig{
		Theme:             "auto",
		AutoSave:          true,
		ShowLineNumbers:   true,
		DefaultTemplate:   "dev",
		MaxFileSize:       10485760, // 10MB
		MaxDirectoryDepth: 10,
		WorkerPoolSize:    10,
		RefreshInterval:   1000, // 1 second
		EnableHotReload:   true,
	}
}

// ToLegacyConfig converts enhanced config to legacy config format for backward compatibility
func (ec *EnhancedConfig) ToLegacyConfig() *Config {
	return &Config{
		OpenAI: OpenAIConfig{
			APIKeyAlias: ec.OpenAI.APIKeyAlias,
			BaseURL:     ec.OpenAI.BaseURL,
			Model:       ec.OpenAI.Model,
			Timeout:     ec.OpenAI.Timeout,
			MaxTokens:   ec.OpenAI.MaxTokens,
			Temperature: ec.OpenAI.Temperature,
			MaxRetries:  ec.OpenAI.MaxRetries,
			RetryDelay:  ec.OpenAI.RetryDelay,
		},
		Translation: TranslationConfig{
			Enabled:        ec.Translation.Enabled,
			TargetLanguage: ec.Translation.TargetLanguage,
			ContextPrompt:  ec.Translation.ContextPrompt,
		},
		App: AppConfig{
			Theme:           ec.App.Theme,
			AutoSave:        ec.App.AutoSave,
			ShowLineNumbers: ec.App.ShowLineNumbers,
			DefaultTemplate: ec.App.DefaultTemplate,
		},
		Version:     ec.Version,
		LastUpdated: ec.LastUpdated,
	}
}

// FromLegacyConfig creates enhanced config from legacy config
func FromLegacyConfig(legacy *Config) *EnhancedConfig {
	enhanced := DefaultEnhancedConfig()

	if legacy != nil {
		// Convert OpenAI config
		enhanced.OpenAI = EnhancedOpenAIConfig{
			APIKeyAlias: legacy.OpenAI.APIKeyAlias,
			BaseURL:     legacy.OpenAI.BaseURL,
			Model:       legacy.OpenAI.Model,
			Timeout:     legacy.OpenAI.Timeout,
			MaxTokens:   legacy.OpenAI.MaxTokens,
			Temperature: legacy.OpenAI.Temperature,
			MaxRetries:  legacy.OpenAI.MaxRetries,
			RetryDelay:  legacy.OpenAI.RetryDelay,
		}

		// Convert Translation config
		enhanced.Translation = EnhancedTranslationConfig{
			Enabled:        legacy.Translation.Enabled,
			TargetLanguage: legacy.Translation.TargetLanguage,
			ContextPrompt:  legacy.Translation.ContextPrompt,
			CacheEnabled:   true, // Default enhanced value
			CacheSize:      1000, // Default enhanced value
			CacheTTL:       3600, // Default enhanced value (1 hour)
		}

		// Convert App config
		enhanced.App = EnhancedAppConfig{
			Theme:             legacy.App.Theme,
			AutoSave:          legacy.App.AutoSave,
			ShowLineNumbers:   legacy.App.ShowLineNumbers,
			DefaultTemplate:   legacy.App.DefaultTemplate,
			MaxFileSize:       10485760, // Default enhanced value (10MB)
			MaxDirectoryDepth: 10,       // Default enhanced value
			WorkerPoolSize:    10,       // Default enhanced value
			RefreshInterval:   1000,     // Default enhanced value (1 second)
			EnableHotReload:   true,     // Default enhanced value
		}

		enhanced.Version = legacy.Version
		enhanced.LastUpdated = legacy.LastUpdated
	}

	return enhanced
}

// Validate performs comprehensive validation of the enhanced configuration
func (ec *EnhancedConfig) Validate(validator *validator.Validate) *ConfigValidationResult {
	err := validator.Struct(ec)
	if err == nil {
		return &ConfigValidationResult{Valid: true}
	}

	var validationErrors []ValidationError

	// Check if it's validation errors using string checking as fallback
	errStr := err.Error()
	if validationErrs, ok := err.(interface{ Error() string }); ok && validationErrs != nil {
		// Try to parse the error for field information
		validationErrors = append(validationErrors, ValidationError{
			Field:   "Configuration",
			Tag:     "validation",
			Value:   "",
			Message: errStr,
		})
	}

	return &ConfigValidationResult{
		Valid:  false,
		Errors: validationErrors,
	}
}

// generateErrorMessage creates user-friendly error messages from validation tags
func generateErrorMessage(tag, field, param string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "url":
		return field + " must be a valid URL"
	case "min":
		return field + " must be at least " + param
	case "max":
		return field + " must be at most " + param
	case "oneof":
		return field + " must be one of: " + param
	case "email":
		return field + " must be a valid email address"
	default:
		return field + " validation failed: " + tag
	}
}

// SetupEnhancedValidator creates and configures a validator with custom rules
func SetupEnhancedValidator() *validator.Validate {
	validate := validator.New()

	// Register custom validation for OpenAI models
	validate.RegisterValidation("openai_model", func(fl validator.FieldLevel) bool {
		validModels := map[string]bool{
			"gpt-4o":            true,
			"gpt-4o-mini":       true,
			"gpt-4-turbo":       true,
			"gpt-4":             true,
			"gpt-3.5-turbo":     true,
			"gpt-3.5-turbo-16k": true,
		}
		return validModels[fl.Field().String()]
	})

	// Register custom validation for API URLs (allows localhost and custom ports)
	validate.RegisterValidation("api_url", func(fl validator.FieldLevel) bool {
		url := fl.Field().String()
		// Allow HTTP and HTTPS protocols
		return len(url) > 0 && (url[:8] == "https://" || url[:7] == "http://")
	})

	// Register custom validation for language codes
	validate.RegisterValidation("language_code", func(fl validator.FieldLevel) bool {
		validLanguages := map[string]bool{
			"en": true, "es": true, "fr": true, "de": true, "it": true,
			"pt": true, "ru": true, "ja": true, "ko": true, "zh": true,
		}
		return validLanguages[fl.Field().String()]
	})

	// Register custom validation for theme values
	validate.RegisterValidation("theme_value", func(fl validator.FieldLevel) bool {
		validThemes := map[string]bool{
			"auto": true, "dark": true, "light": true,
		}
		return validThemes[fl.Field().String()]
	})

	return validate
}

// IsEquivalent checks if two enhanced configurations are functionally equivalent
func (ec *EnhancedConfig) IsEquivalent(other *EnhancedConfig) bool {
	if ec == nil || other == nil {
		return ec == other
	}

	// Compare core configuration fields (excluding LastUpdated)
	return ec.OpenAI == other.OpenAI &&
		ec.Translation == other.Translation &&
		ec.App == other.App &&
		ec.Version == other.Version
}

// Clone creates a deep copy of the enhanced configuration
func (ec *EnhancedConfig) Clone() *EnhancedConfig {
	if ec == nil {
		return nil
	}

	return &EnhancedConfig{
		OpenAI:      ec.OpenAI,
		Translation: ec.Translation,
		App:         ec.App,
		Version:     ec.Version,
		LastUpdated: ec.LastUpdated,
	}
}
