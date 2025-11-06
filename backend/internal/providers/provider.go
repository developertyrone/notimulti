package providers

import "context"

// Provider defines the interface that all notification providers must implement
type Provider interface {
	// Send sends a notification and returns an error if the operation fails
	Send(ctx context.Context, notification *Notification) error

	// GetStatus returns the current status of the provider
	GetStatus() *ProviderStatus

	// GetID returns the unique identifier of this provider instance
	GetID() string

	// GetType returns the type of provider (e.g., "telegram", "email")
	GetType() string

	// Close performs cleanup operations and releases resources
	Close() error
}
