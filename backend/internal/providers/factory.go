package providers

import (
	"fmt"
)

// Factory creates provider instances based on configuration
type Factory struct{}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{}
}

// NewProvider creates a new provider instance based on the configuration type
func (f *Factory) NewProvider(config *ProviderConfig) (Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("provider config cannot be nil")
	}

	if config.ID == "" {
		return nil, fmt.Errorf("provider ID is required")
	}

	if config.Type == "" {
		return nil, fmt.Errorf("provider type is required")
	}

	switch config.Type {
	case "telegram":
		if config.Telegram == nil {
			return nil, fmt.Errorf("telegram config is required for telegram provider")
		}
		return NewTelegramProvider(config.ID, config.Telegram)

	case "email":
		if config.Email == nil {
			return nil, fmt.Errorf("email config is required for email provider")
		}
		return NewEmailProvider(config.ID, config.Email)

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}
}
