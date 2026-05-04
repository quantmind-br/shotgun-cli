package llm

import (
	"fmt"
	"sync"
)

// ProviderCreator is a function that creates a Provider from Config.
type ProviderCreator func(cfg Config) (Provider, error)

// Registry manages provider registration.
type Registry struct {
	mu       sync.RWMutex
	creators map[ProviderType]ProviderCreator
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		creators: make(map[ProviderType]ProviderCreator),
	}
}

// Register registers a creator for a provider type.
func (r *Registry) Register(providerType ProviderType, creator ProviderCreator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.creators[providerType] = creator
}

// Create creates a provider from the configuration.
func (r *Registry) Create(cfg Config) (Provider, error) {
	r.mu.RLock()
	creator, ok := r.creators[cfg.Provider]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}

	return creator(cfg)
}

// SupportedProviders returns registered providers.
func (r *Registry) SupportedProviders() []ProviderType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ProviderType, 0, len(r.creators))
	for pt := range r.creators {
		result = append(result, pt)
	}
	return result
}

// IsRegistered checks if a provider is registered.
func (r *Registry) IsRegistered(providerType ProviderType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.creators[providerType]
	return ok
}
