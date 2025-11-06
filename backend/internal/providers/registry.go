package providers

import (
	"fmt"
	"log/slog"
	"sync"
)

// Registry manages a thread-safe collection of notification providers
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	logger    *slog.Logger
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// SetLogger sets the logger for the registry
func (r *Registry) SetLogger(logger *slog.Logger) {
	r.logger = logger
}

func (r *Registry) log(level slog.Level, msg string, args ...any) {
	if r.logger != nil {
		r.logger.Log(nil, level, msg, args...)
	}
}

// Register adds a provider to the registry
// If a provider with the same ID already exists, it will be replaced
func (r *Registry) Register(provider Provider) error {
	if provider == nil {
		return fmt.Errorf("cannot register nil provider")
	}

	id := provider.GetID()
	if id == "" {
		return fmt.Errorf("provider ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Close existing provider if being replaced
	if existing, exists := r.providers[id]; exists {
		if err := existing.Close(); err != nil {
			// Log error but continue with replacement
			fmt.Printf("Warning: error closing existing provider %s: %v\n", id, err)
		}
	}

	r.providers[id] = provider
	return nil
}

// Get retrieves a provider by ID
func (r *Registry) Get(id string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[id]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", id)
	}

	return provider, nil
}

// List returns all registered providers
func (r *Registry) List() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}

	return providers
}

// Remove removes a provider from the registry and closes it
func (r *Registry) Remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	provider, exists := r.providers[id]
	if !exists {
		return fmt.Errorf("provider not found: %s", id)
	}

	// Close the provider
	if err := provider.Close(); err != nil {
		return fmt.Errorf("error closing provider %s: %w", id, err)
	}

	delete(r.providers, id)
	r.log(slog.LevelInfo, "Provider removed", "id", id, "type", provider.GetType())
	return nil
}

// Replace atomically replaces a provider with a new one
// The old provider is closed after the swap
func (r *Registry) Replace(id string, newProvider Provider) error {
	if newProvider == nil {
		return fmt.Errorf("cannot replace with nil provider")
	}

	if newProvider.GetID() != id {
		return fmt.Errorf("provider ID mismatch: expected %s, got %s", id, newProvider.GetID())
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldProvider, exists := r.providers[id]

	// Get checksums for logging
	oldChecksum := ""
	newChecksum := ""
	if exists {
		if status := oldProvider.GetStatus(); status != nil {
			oldChecksum = status.ConfigChecksum
		}
	}
	if status := newProvider.GetStatus(); status != nil {
		newChecksum = status.ConfigChecksum
	}

	// Atomic swap - this is the critical section
	// Once we update the map, new requests will use the new provider
	r.providers[id] = newProvider

	r.log(slog.LevelInfo, "Provider replaced",
		"id", id,
		"type", newProvider.GetType(),
		"old_checksum", oldChecksum,
		"new_checksum", newChecksum)

	// Close old provider after swap (outside critical section would be ideal,
	// but we're already in the lock, so do it here)
	if exists {
		if err := oldProvider.Close(); err != nil {
			// Log error but don't fail the replacement
			r.log(slog.LevelWarn, "Error closing old provider after replacement",
				"id", id,
				"error", err)
		}
	}

	return nil
}

// Count returns the number of registered providers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.providers)
}

// Clear removes all providers from the registry
func (r *Registry) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errors []error
	for id, provider := range r.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("error closing provider %s: %w", id, err))
		}
	}

	r.providers = make(map[string]Provider)

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while clearing registry: %v", errors)
	}

	return nil
}
