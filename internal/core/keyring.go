package core

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/99designs/keyring"
	"github.com/adrg/xdg"
)

// SecureKeyManager handles secure storage and retrieval of API keys
type SecureKeyManager struct {
	ring keyring.Keyring
	mu   sync.RWMutex
}

// NewSecureKeyManager creates a new secure key manager
func NewSecureKeyManager() (*SecureKeyManager, error) {
	// Configure keyring with multiple backends for cross-platform support
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "shotgun-cli",

		// macOS Keychain configuration
		KeychainTrustApplication: true,
		KeychainName:             "shotgun-cli",

		// Linux Secret Service configuration
		LibSecretCollectionName: "shotgun-cli",

		// Windows Credential Manager uses default configuration

		// File backend as fallback
		FileDir:          filepath.Join(xdg.ConfigHome, "shotgun-cli", "keyring"),
		FilePasswordFunc: keyring.FixedStringPrompt("Enter password to encrypt API keys:"),

		// Allow all backends for maximum compatibility
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,      // macOS
			keyring.SecretServiceBackend, // Linux
			keyring.WinCredBackend,       // Windows
			keyring.FileBackend,          // Cross-platform fallback
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize secure keyring: %w", err)
	}

	return &SecureKeyManager{
		ring: ring,
	}, nil
}

// StoreAPIKey stores an API key with the given alias
func (skm *SecureKeyManager) StoreAPIKey(alias, apiKey string) error {
	skm.mu.Lock()
	defer skm.mu.Unlock()

	if alias == "" {
		return fmt.Errorf("API key alias cannot be empty")
	}

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	item := keyring.Item{
		Key:         skm.makeKeyName(alias),
		Data:        []byte(apiKey),
		Label:       fmt.Sprintf("shotgun-cli API key: %s", alias),
		Description: fmt.Sprintf("OpenAI API key for shotgun-cli service (%s)", alias),
	}

	if err := skm.ring.Set(item); err != nil {
		return fmt.Errorf("failed to store API key '%s': %w", alias, err)
	}

	return nil
}

// GetAPIKey retrieves an API key by alias
func (skm *SecureKeyManager) GetAPIKey(alias string) (string, error) {
	skm.mu.RLock()
	defer skm.mu.RUnlock()

	if alias == "" {
		return "", fmt.Errorf("API key alias cannot be empty")
	}

	item, err := skm.ring.Get(skm.makeKeyName(alias))
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", fmt.Errorf("API key '%s' not found", alias)
		}
		return "", fmt.Errorf("failed to retrieve API key '%s': %w", alias, err)
	}

	return string(item.Data), nil
}

// DeleteAPIKey removes an API key by alias
func (skm *SecureKeyManager) DeleteAPIKey(alias string) error {
	skm.mu.Lock()
	defer skm.mu.Unlock()

	if alias == "" {
		return fmt.Errorf("API key alias cannot be empty")
	}

	if err := skm.ring.Remove(skm.makeKeyName(alias)); err != nil {
		if err == keyring.ErrKeyNotFound {
			return fmt.Errorf("API key '%s' not found", alias)
		}
		return fmt.Errorf("failed to delete API key '%s': %w", alias, err)
	}

	return nil
}

// ListAPIKeys returns all stored API key aliases
func (skm *SecureKeyManager) ListAPIKeys() ([]string, error) {
	skm.mu.RLock()
	defer skm.mu.RUnlock()

	keys, err := skm.ring.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	var aliases []string
	prefix := skm.makeKeyName("")

	for _, key := range keys {
		// Only include keys that belong to shotgun-cli
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			alias := key[len(prefix):]
			aliases = append(aliases, alias)
		}
	}

	return aliases, nil
}

// HasAPIKey checks if an API key exists for the given alias
func (skm *SecureKeyManager) HasAPIKey(alias string) bool {
	skm.mu.RLock()
	defer skm.mu.RUnlock()

	if alias == "" {
		return false
	}

	_, err := skm.ring.Get(skm.makeKeyName(alias))
	return err == nil
}

// UpdateAPIKey updates an existing API key or creates it if it doesn't exist
func (skm *SecureKeyManager) UpdateAPIKey(alias, apiKey string) error {
	// UpdateAPIKey is the same as StoreAPIKey since keyring.Set() upserts
	return skm.StoreAPIKey(alias, apiKey)
}

// TestAPIKey validates that an API key can be stored and retrieved
func (skm *SecureKeyManager) TestAPIKey(alias, apiKey string) error {
	testAlias := fmt.Sprintf("test_%s", alias)

	// Store test key
	if err := skm.StoreAPIKey(testAlias, apiKey); err != nil {
		return fmt.Errorf("failed to store test API key: %w", err)
	}

	// Retrieve test key
	retrievedKey, err := skm.GetAPIKey(testAlias)
	if err != nil {
		// Clean up on failure
		skm.DeleteAPIKey(testAlias)
		return fmt.Errorf("failed to retrieve test API key: %w", err)
	}

	// Verify keys match
	if retrievedKey != apiKey {
		// Clean up on failure
		skm.DeleteAPIKey(testAlias)
		return fmt.Errorf("test API key mismatch: stored and retrieved keys do not match")
	}

	// Clean up test key
	if err := skm.DeleteAPIKey(testAlias); err != nil {
		return fmt.Errorf("failed to clean up test API key: %w", err)
	}

	return nil
}

// GetBackendInfo returns information about the active keyring backend
func (skm *SecureKeyManager) GetBackendInfo() string {
	// This is a simple implementation; the keyring library doesn't expose
	// the active backend directly, but we can infer from the type
	switch skm.ring.(type) {
	default:
		return "Unknown backend"
	}
}

// Clear removes all shotgun-cli API keys from the keyring
func (skm *SecureKeyManager) Clear() error {
	skm.mu.Lock()
	defer skm.mu.Unlock()

	aliases, err := skm.ListAPIKeys()
	if err != nil {
		return fmt.Errorf("failed to list API keys for clearing: %w", err)
	}

	var errors []error
	for _, alias := range aliases {
		if err := skm.ring.Remove(skm.makeKeyName(alias)); err != nil {
			errors = append(errors, fmt.Errorf("failed to remove key '%s': %w", alias, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to clear some API keys: %v", errors)
	}

	return nil
}

// makeKeyName creates a prefixed key name for the keyring
func (skm *SecureKeyManager) makeKeyName(alias string) string {
	if alias == "" {
		return "shotgun-cli-apikey-"
	}
	return fmt.Sprintf("shotgun-cli-apikey-%s", alias)
}

// IsAvailable checks if the keyring is available and functional
func (skm *SecureKeyManager) IsAvailable() bool {
	// Test basic functionality with a temporary key
	testAlias := "availability_test"
	testKey := "test_key_12345"

	// Try to store and retrieve a test key
	if err := skm.StoreAPIKey(testAlias, testKey); err != nil {
		return false
	}

	retrievedKey, err := skm.GetAPIKey(testAlias)
	if err != nil || retrievedKey != testKey {
		return false
	}

	// Clean up
	skm.DeleteAPIKey(testAlias)

	return true
}
