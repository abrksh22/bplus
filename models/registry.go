package models

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages all registered providers.
type Registry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register registers a provider.
func (r *Registry) Register(provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := provider.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = provider
	return nil
}

// Get retrieves a provider by name.
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// List returns all registered provider names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ListAll returns all registered providers.
func (r *Registry) ListAll() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}

// TestAll tests connectivity for all providers.
func (r *Registry) TestAll(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	for name, provider := range r.providers {
		results[name] = provider.TestConnection(ctx)
	}
	return results
}

// Global registry instance
var globalRegistry = NewRegistry()

// Register registers a provider globally.
func Register(provider Provider) error {
	return globalRegistry.Register(provider)
}

// Get retrieves a provider by name from the global registry.
func Get(name string) (Provider, error) {
	return globalRegistry.Get(name)
}

// List returns all registered provider names from the global registry.
func List() []string {
	return globalRegistry.List()
}

// ListAll returns all registered providers from the global registry.
func ListAll() []Provider {
	return globalRegistry.ListAll()
}

// TestAll tests connectivity for all providers in the global registry.
func TestAll(ctx context.Context) map[string]error {
	return globalRegistry.TestAll(ctx)
}
