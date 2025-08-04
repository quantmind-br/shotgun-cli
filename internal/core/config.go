package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"
)

// ConfigManager handles application configuration with thread-safe operations
type ConfigManager struct {
	config     *Config
	configPath string
	mu         sync.RWMutex
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() (*ConfigManager, error) {
	configPath := filepath.Join(xdg.ConfigHome, "shotgun-cli", "config.json")

	cm := &ConfigManager{
		configPath: configPath,
	}

	// Ensure config directory exists
	if err := cm.ensureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config or create default
	if err := cm.Load(); err != nil {
		// If load fails, create default config
		cm.config = DefaultConfig()
		if saveErr := cm.Save(); saveErr != nil {
			return nil, fmt.Errorf("failed to create default config: %w", saveErr)
		}
	}

	return cm, nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		OpenAI: OpenAIConfig{
			BaseURL:     "https://api.openai.com/v1",
			Model:       "gpt-4o",
			Timeout:     300,
			MaxTokens:   4096,
			Temperature: 0.7,
			MaxRetries:  3,
			RetryDelay:  2,
		},
		Translation: TranslationConfig{
			Enabled:        false,
			TargetLanguage: "en",
			ContextPrompt:  "Translate the following text to English, preserving technical terms and maintaining the original meaning:",
		},
		App: AppConfig{
			Theme:           "auto",
			AutoSave:        true,
			ShowLineNumbers: true,
			DefaultTemplate: "dev",
		},
		Version:     "1.0.0",
		LastUpdated: time.Now(),
	}
}

// Load reads configuration from file
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use default config
			cm.config = DefaultConfig()
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and apply defaults for missing fields
	cm.config = cm.validateAndFillDefaults(&config)

	return nil
}

// Save writes configuration to file
func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	// Update timestamp
	cm.config.LastUpdated = time.Now()

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Get returns a copy of the current configuration
func (cm *ConfigManager) Get() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return DefaultConfig()
	}

	// Return a deep copy to prevent external modifications
	return cm.copyConfig(cm.config)
}

// Update updates the configuration with new values
func (cm *ConfigManager) Update(config *Config) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate configuration
	if err := cm.validateConfig(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	cm.config = cm.copyConfig(config)
	return nil
}

// GetOpenAI returns the OpenAI configuration
func (cm *ConfigManager) GetOpenAI() OpenAIConfig {
	config := cm.Get()
	return config.OpenAI
}

// UpdateOpenAI updates only the OpenAI configuration
func (cm *ConfigManager) UpdateOpenAI(openaiConfig OpenAIConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		cm.config = DefaultConfig()
	}

	cm.config.OpenAI = openaiConfig
	return nil
}

// GetTranslation returns the translation configuration
func (cm *ConfigManager) GetTranslation() TranslationConfig {
	config := cm.Get()
	return config.Translation
}

// UpdateTranslation updates only the translation configuration
func (cm *ConfigManager) UpdateTranslation(translationConfig TranslationConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		cm.config = DefaultConfig()
	}

	cm.config.Translation = translationConfig
	return nil
}

// GetApp returns the application configuration
func (cm *ConfigManager) GetApp() AppConfig {
	config := cm.Get()
	return config.App
}

// UpdateApp updates only the application configuration
func (cm *ConfigManager) UpdateApp(appConfig AppConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		cm.config = DefaultConfig()
	}

	cm.config.App = appConfig
	return nil
}

// GetConfigPath returns the path to the configuration file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}

// Reset resets configuration to defaults
func (cm *ConfigManager) Reset() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.config = DefaultConfig()
	return nil
}

// ensureConfigDir creates the configuration directory if it doesn't exist
func (cm *ConfigManager) ensureConfigDir() error {
	configDir := filepath.Dir(cm.configPath)
	return os.MkdirAll(configDir, 0755)
}

// validateConfig validates the configuration values
func (cm *ConfigManager) validateConfig(config *Config) error {
	// Validate OpenAI config
	if config.OpenAI.BaseURL == "" {
		return fmt.Errorf("OpenAI base URL cannot be empty")
	}
	if config.OpenAI.Model == "" {
		return fmt.Errorf("OpenAI model cannot be empty")
	}
	if config.OpenAI.Timeout <= 0 {
		return fmt.Errorf("OpenAI timeout must be positive")
	}
	if config.OpenAI.MaxTokens <= 0 {
		return fmt.Errorf("OpenAI max tokens must be positive")
	}
	if config.OpenAI.Temperature < 0 || config.OpenAI.Temperature > 2 {
		return fmt.Errorf("OpenAI temperature must be between 0 and 2")
	}
	if config.OpenAI.MaxRetries < 0 {
		return fmt.Errorf("OpenAI max retries cannot be negative")
	}
	if config.OpenAI.RetryDelay < 0 {
		return fmt.Errorf("OpenAI retry delay cannot be negative")
	}

	// Validate Translation config
	if config.Translation.TargetLanguage == "" {
		return fmt.Errorf("translation target language cannot be empty")
	}

	// Validate App config
	validThemes := map[string]bool{"auto": true, "dark": true, "light": true}
	if !validThemes[config.App.Theme] {
		return fmt.Errorf("invalid theme: %s (must be auto, dark, or light)", config.App.Theme)
	}

	validTemplates := map[string]bool{"dev": true, "architect": true, "debug": true, "project-manager": true}
	if !validTemplates[config.App.DefaultTemplate] {
		return fmt.Errorf("invalid default template: %s", config.App.DefaultTemplate)
	}

	return nil
}

// validateAndFillDefaults validates config and fills missing fields with defaults
func (cm *ConfigManager) validateAndFillDefaults(config *Config) *Config {
	defaults := DefaultConfig()

	// Fill missing OpenAI fields
	if config.OpenAI.BaseURL == "" {
		config.OpenAI.BaseURL = defaults.OpenAI.BaseURL
	}
	if config.OpenAI.Model == "" {
		config.OpenAI.Model = defaults.OpenAI.Model
	}
	if config.OpenAI.Timeout <= 0 {
		config.OpenAI.Timeout = defaults.OpenAI.Timeout
	}
	if config.OpenAI.MaxTokens <= 0 {
		config.OpenAI.MaxTokens = defaults.OpenAI.MaxTokens
	}
	if config.OpenAI.Temperature < 0 || config.OpenAI.Temperature > 2 {
		config.OpenAI.Temperature = defaults.OpenAI.Temperature
	}
	if config.OpenAI.MaxRetries < 0 {
		config.OpenAI.MaxRetries = defaults.OpenAI.MaxRetries
	}
	if config.OpenAI.RetryDelay < 0 {
		config.OpenAI.RetryDelay = defaults.OpenAI.RetryDelay
	}

	// Fill missing Translation fields
	if config.Translation.TargetLanguage == "" {
		config.Translation.TargetLanguage = defaults.Translation.TargetLanguage
	}
	if config.Translation.ContextPrompt == "" {
		config.Translation.ContextPrompt = defaults.Translation.ContextPrompt
	}

	// Fill missing App fields
	validThemes := map[string]bool{"auto": true, "dark": true, "light": true}
	if !validThemes[config.App.Theme] {
		config.App.Theme = defaults.App.Theme
	}
	validTemplates := map[string]bool{"dev": true, "architect": true, "debug": true, "project-manager": true}
	if !validTemplates[config.App.DefaultTemplate] {
		config.App.DefaultTemplate = defaults.App.DefaultTemplate
	}

	// Fill missing metadata
	if config.Version == "" {
		config.Version = defaults.Version
	}

	return config
}

// copyConfig creates a deep copy of the configuration
func (cm *ConfigManager) copyConfig(config *Config) *Config {
	return &Config{
		OpenAI:      config.OpenAI,
		Translation: config.Translation,
		App:         config.App,
		Version:     config.Version,
		LastUpdated: config.LastUpdated,
	}
}
