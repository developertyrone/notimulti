package providers

import (
	"context"
	"fmt"
	"time"
)

// FailedProvider represents a provider that failed to initialize
// It implements the Provider interface but always returns errors for Send operations
type FailedProvider struct {
	id           string
	providerType string
	errorMessage string
	status       *ProviderStatus
}

// NewFailedProvider creates a new failed provider instance
func NewFailedProvider(id, providerType string, initError error) *FailedProvider {
	return &FailedProvider{
		id:           id,
		providerType: providerType,
		errorMessage: initError.Error(),
		status: &ProviderStatus{
			Status:       "error",
			LastUpdated:  time.Now(),
			ErrorMessage: fmt.Sprintf("Initialization failed: %v", initError),
		},
	}
}

// Send always returns an error for failed providers
func (fp *FailedProvider) Send(ctx context.Context, notification *Notification) error {
	return fmt.Errorf("provider is in error state: %s", fp.errorMessage)
}

// GetStatus returns the error status
func (fp *FailedProvider) GetStatus() *ProviderStatus {
	return fp.status
}

// GetID returns the provider ID
func (fp *FailedProvider) GetID() string {
	return fp.id
}

// GetType returns the provider type
func (fp *FailedProvider) GetType() string {
	return fp.providerType
}

// GetTestRecipient returns error for failed providers (T050)
func (fp *FailedProvider) GetTestRecipient() (string, error) {
	return "", fmt.Errorf("provider is in error state: %s", fp.errorMessage)
}

// Test returns error for failed providers (T051)
func (fp *FailedProvider) Test(ctx context.Context) error {
	return fmt.Errorf("cannot test provider in error state: %s", fp.errorMessage)
}

// Close does nothing for failed providers
func (fp *FailedProvider) Close() error {
	return nil
}
