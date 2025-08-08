package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	koanfJson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// EnhancedConfigManager provides advanced configuration management with Koanf
type EnhancedConfigManager struct {
	koanf      *koanf.Koanf
	validator  *validator.Validate
	keyring    *SecureKeyManager
	watcher    *fsnotify.Watcher
	config     *EnhancedConfig
	configPath string
	mu         sync.RWMutex
	eventChan  chan ConfigurationEvent
	stopChan   chan struct{}
	isWatching bool
}

// NewEnhancedConfigManager creates a new enhanced configuration manager
func NewEnhancedConfigManager() (*EnhancedConfigManager, error) {
	configPath := filepath.Join(xdg.ConfigHome, "shotgun-cli", "config.json")

	keyring, err := NewSecureKeyManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create secure key manager: %w", err)
	}

	return &EnhancedConfigManager{
		koanf:      koanf.New("."),
		validator:  SetupEnhancedValidator(),
		keyring:    keyring,
		configPath: configPath,
		eventChan:  make(chan ConfigurationEvent, 10),
		stopChan:   make(chan struct{}),
	}, nil
}

// Initialize sets up the configuration manager and loads initial configuration
func (cm *EnhancedConfigManager) Initialize() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Ensure config directory exists
	if err := cm.ensureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load configuration from multiple sources
	if err := cm.loadConfiguration(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	return nil
}

// Load implements ConfigManagerInterface for backward compatibility
func (cm *EnhancedConfigManager) Load() error {
	return cm.Initialize()
}

// loadConfiguration loads configuration from multiple sources with priorities
func (cm *EnhancedConfigManager) loadConfiguration() error {
	// Clear any existing configuration
	cm.koanf = koanf.New(".")

	// 1. Load defaults first (lowest priority)
	defaultConfig := cm.getDefaultConfigMap()
	if err := cm.koanf.Load(confmap.Provider(defaultConfig, "."), nil); err != nil {
		return fmt.Errorf("failed to load default configuration: %w", err)
	}

	// 2. Load from config file if it exists (medium priority)
	if _, err := os.Stat(cm.configPath); err == nil {
		if err := cm.koanf.Load(file.Provider(cm.configPath), koanfJson.Parser()); err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// 3. Override with environment variables (highest priority)
	if err := cm.koanf.Load(env.Provider("SHOTGUN_", ".", cm.envTransform), nil); err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// 4. Parse into enhanced config structure
	enhanced := &EnhancedConfig{}
	if err := cm.koanf.Unmarshal("", enhanced); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// 5. Load secure values from keyring
	if err := cm.loadSecureValues(enhanced); err != nil {
		return fmt.Errorf("failed to load secure values: %w", err)
	}

	// 6. Validate complete configuration
	if validationResult := enhanced.Validate(cm.validator); !validationResult.Valid {
		return fmt.Errorf("configuration validation failed: %v", validationResult.Errors)
	}

	// 7. Store validated configuration
	cm.config = enhanced

	// 8. Emit configuration loaded event
	cm.emitEvent("loaded", "initialization", nil, "")

	return nil
}

// Save persists the current configuration to file
func (cm *EnhancedConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	// Store secure values in keyring before saving
	if err := cm.storeSecureValues(cm.config); err != nil {
		return fmt.Errorf("failed to store secure values: %w", err)
	}

	// Create a copy for file storage
	configForFile := cm.config.Clone()
	configForFile.LastUpdated = time.Now()

	// Serialize to JSON with proper formatting
	data, err := json.MarshalIndent(configForFile, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write to file atomically
	tempFile := cm.configPath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if err := os.Rename(tempFile, cm.configPath); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to save config file: %w", err)
	}

	// Update internal state
	cm.config.LastUpdated = configForFile.LastUpdated

	// Emit configuration saved event
	cm.emitEvent("saved", "user", nil, "")

	return nil
}

// Get returns the current configuration (legacy compatibility)
func (cm *EnhancedConfigManager) Get() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return DefaultEnhancedConfig().ToLegacyConfig()
	}

	return cm.config.ToLegacyConfig()
}

// GetEnhanced returns the current enhanced configuration
func (cm *EnhancedConfigManager) GetEnhanced() *EnhancedConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return DefaultEnhancedConfig()
	}

	return cm.config.Clone()
}

// Update updates the configuration (legacy compatibility)
func (cm *EnhancedConfigManager) Update(config *Config) error {
	enhanced := FromLegacyConfig(config)
	return cm.UpdateEnhanced(enhanced)
}

// UpdateEnhanced updates the enhanced configuration
func (cm *EnhancedConfigManager) UpdateEnhanced(config *EnhancedConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate the new configuration
	if validationResult := config.Validate(cm.validator); !validationResult.Valid {
		return fmt.Errorf("configuration validation failed: %v", validationResult.Errors)
	}

	// Store the old configuration for comparison
	oldConfig := cm.config
	cm.config = config.Clone()
	cm.config.LastUpdated = time.Now()

	// Determine what changed
	changes := cm.detectChanges(oldConfig, cm.config)

	// Emit configuration changed event
	cm.emitEvent("changed", "api", changes, "")

	return nil
}

// Reset resets configuration to defaults
func (cm *EnhancedConfigManager) Reset() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Clear keyring entries
	if err := cm.keyring.Clear(); err != nil {
		return fmt.Errorf("failed to clear keyring: %w", err)
	}

	// Remove config file
	if err := os.Remove(cm.configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	// Reset to defaults
	cm.config = DefaultEnhancedConfig()

	// Emit configuration reset event
	cm.emitEvent("reset", "user", nil, "")

	return nil
}

// GetConfigPath returns the configuration file path
func (cm *EnhancedConfigManager) GetConfigPath() string {
	return cm.configPath
}

// StartWatching begins watching for configuration file changes
func (cm *EnhancedConfigManager) StartWatching(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isWatching {
		return nil // Already watching
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	cm.watcher = watcher
	cm.isWatching = true

	// Watch the config directory (not just the file, in case it gets recreated)
	configDir := filepath.Dir(cm.configPath)
	if err := cm.watcher.Add(configDir); err != nil {
		cm.watcher.Close()
		cm.isWatching = false
		return fmt.Errorf("failed to watch config directory: %w", err)
	}

	// Start watching goroutine
	go cm.watchConfigFile(ctx)

	return nil
}

// StopWatching stops watching for configuration file changes
func (cm *EnhancedConfigManager) StopWatching() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.isWatching {
		return
	}

	close(cm.stopChan)
	if cm.watcher != nil {
		cm.watcher.Close()
	}
	cm.isWatching = false
	cm.stopChan = make(chan struct{}) // Reset for next use
}

// IsConfigValid returns whether the current configuration is valid
func (cm *EnhancedConfigManager) IsConfigValid() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return false
	}

	validationResult := cm.config.Validate(cm.validator)
	return validationResult.Valid
}

// GetValidationErrors returns current configuration validation errors
func (cm *EnhancedConfigManager) GetValidationErrors() map[string]string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return map[string]string{"config": "no configuration loaded"}
	}

	validationResult := cm.config.Validate(cm.validator)
	if validationResult.Valid {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range validationResult.Errors {
		errors[err.Field] = err.Message
	}

	return errors
}

// GetEventChannel returns the configuration event channel
func (cm *EnhancedConfigManager) GetEventChannel() <-chan ConfigurationEvent {
	return cm.eventChan
}

// TestConnection tests the API connection with current configuration
func (cm *EnhancedConfigManager) TestConnection() error {
	config := cm.GetEnhanced()
	if config == nil {
		return fmt.Errorf("no configuration available")
	}

	// Use API key directly from config
	apiKey := config.OpenAI.APIKey
	if apiKey == "" {
		return fmt.Errorf("no API key configured")
	}

	// Test the connection (simplified - would use actual OpenAI client)
	// This is a placeholder for the actual implementation
	return nil
}

// Private helper methods

func (cm *EnhancedConfigManager) ensureConfigDir() error {
	configDir := filepath.Dir(cm.configPath)
	return os.MkdirAll(configDir, 0700)
}

func (cm *EnhancedConfigManager) getDefaultConfigMap() map[string]interface{} {
	defaultConfig := DefaultEnhancedConfig()

	// Convert to map for Koanf
	configMap := make(map[string]interface{})
	data, _ := json.Marshal(defaultConfig)
	json.Unmarshal(data, &configMap)

	return configMap
}

func (cm *EnhancedConfigManager) envTransform(s string) string {
	// Transform SHOTGUN_OPENAI_BASE_URL to openai.baseUrl
	s = strings.ToLower(strings.TrimPrefix(s, "SHOTGUN_"))
	s = strings.ReplaceAll(s, "_", ".")
	return s
}

func (cm *EnhancedConfigManager) loadSecureValues(config *EnhancedConfig) error {
	// API keys are stored directly in config now
	return nil
}

func (cm *EnhancedConfigManager) storeSecureValues(config *EnhancedConfig) error {
	// API keys are stored directly in config now
	return nil
}

func (cm *EnhancedConfigManager) watchConfigFile(ctx context.Context) {
	defer cm.watcher.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopChan:
			return
		case event, ok := <-cm.watcher.Events:
			if !ok {
				return
			}

			// Only react to changes to our config file
			if event.Name == cm.configPath && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				// Debounce rapid file changes
				time.Sleep(100 * time.Millisecond)

				if err := cm.loadConfiguration(); err != nil {
					cm.emitEvent("error", "file", nil, err.Error())
				} else {
					cm.emitEvent("reloaded", "file", nil, "")
				}
			}
		case err, ok := <-cm.watcher.Errors:
			if !ok {
				return
			}
			cm.emitEvent("error", "watcher", nil, err.Error())
		}
	}
}

func (cm *EnhancedConfigManager) detectChanges(old, new *EnhancedConfig) map[string]interface{} {
	changes := make(map[string]interface{})

	if old == nil {
		changes["initial"] = true
		return changes
	}

	// Compare key fields and track changes
	if old.OpenAI != new.OpenAI {
		changes["openai"] = true
	}
	if old.Translation != new.Translation {
		changes["translation"] = true
	}
	if !isAppConfigEqual(old.App, new.App) {
		changes["app"] = true
	}

	return changes
}

// Helper function to compare EnhancedAppConfig structs that contain slices
func isAppConfigEqual(a, b EnhancedAppConfig) bool {
	// Compare all non-slice fields first
	if a.AutoSave != b.AutoSave ||
		a.ShowLineNumbers != b.ShowLineNumbers ||
		a.MaxFileSize != b.MaxFileSize ||
		a.MaxDirectoryDepth != b.MaxDirectoryDepth ||
		a.WorkerPoolSize != b.WorkerPoolSize ||
		a.RefreshInterval != b.RefreshInterval ||
		a.EnableHotReload != b.EnableHotReload ||
		a.PatternValidationEnabled != b.PatternValidationEnabled {
		return false
	}

	// Compare slice fields
	if len(a.CustomIgnorePatterns) != len(b.CustomIgnorePatterns) {
		return false
	}
	for i, v := range a.CustomIgnorePatterns {
		if v != b.CustomIgnorePatterns[i] {
			return false
		}
	}

	if len(a.ForceIncludePatterns) != len(b.ForceIncludePatterns) {
		return false
	}
	for i, v := range a.ForceIncludePatterns {
		if v != b.ForceIncludePatterns[i] {
			return false
		}
	}

	return true
}

func (cm *EnhancedConfigManager) emitEvent(eventType, source string, changes map[string]interface{}, errorMsg string) {
	event := ConfigurationEvent{
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now(),
		Changes:   changes,
		Error:     errorMsg,
	}

	select {
	case cm.eventChan <- event:
	default:
		// Channel full, drop the event to avoid blocking
	}
}
