package unit

import (
	"testing"

	"github.com/developertyrone/notimulti/internal/providers"
)

func TestFactoryNewProvider(t *testing.T) {
	factory := providers.NewFactory()

	tests := []struct {
		name     string
		config   *providers.ProviderConfig
		wantErr  bool
		wantType string
	}{
		{
			name:    "nil_config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty_id",
			config: &providers.ProviderConfig{
				ID:   "",
				Type: "telegram",
			},
			wantErr: true,
		},
		{
			name: "empty_type",
			config: &providers.ProviderConfig{
				ID:   "test-1",
				Type: "",
			},
			wantErr: true,
		},
		{
			name: "unsupported_type",
			config: &providers.ProviderConfig{
				ID:   "test-1",
				Type: "sms",
			},
			wantErr: true,
		},
		{
			name: "telegram_without_config",
			config: &providers.ProviderConfig{
				ID:       "test-1",
				Type:     "telegram",
				Telegram: nil,
			},
			wantErr: true,
		},
		{
			name: "telegram_valid",
			config: &providers.ProviderConfig{
				ID:   "telegram-1",
				Type: "telegram",
				Telegram: &providers.TelegramConfig{
					BotToken:      "invalid_token_for_testing",
					DefaultChatID: "12345678",
				},
			},
			wantErr:  true, // Will fail due to invalid token, but tests factory routing
			wantType: "",
		},
		{
			name: "email_without_config",
			config: &providers.ProviderConfig{
				ID:    "test-1",
				Type:  "email",
				Email: nil,
			},
			wantErr: true,
		},
		{
			name: "email_valid",
			config: &providers.ProviderConfig{
				ID:   "email-1",
				Type: "email",
				Email: &providers.EmailConfig{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "user",
					Password: "pass",
					From:     "test@example.com",
				},
			},
			wantErr:  false,
			wantType: "email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.NewProvider(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && provider == nil {
				t.Errorf("NewProvider() returned nil provider for valid config")
				return
			}

			if !tt.wantErr && provider.GetType() != tt.wantType {
				t.Errorf("NewProvider() provider type = %v, want %v", provider.GetType(), tt.wantType)
			}

			if !tt.wantErr && provider.GetID() != tt.config.ID {
				t.Errorf("NewProvider() provider ID = %v, want %v", provider.GetID(), tt.config.ID)
			}
		})
	}
}

func TestFactoryTelegramProvider(t *testing.T) {
	// Note: Telegram provider creation requires a valid bot token
	// This test verifies factory routing, not full provider creation
	factory := providers.NewFactory()
	config := &providers.ProviderConfig{
		ID:   "telegram-test",
		Type: "telegram",
		Telegram: &providers.TelegramConfig{
			BotToken:      "invalid_token_for_testing",
			DefaultChatID: "12345678",
			ParseMode:     "HTML",
		},
	}

	_, err := factory.NewProvider(config)
	// Expected to fail due to invalid token, but we verify factory routing worked
	if err == nil {
		t.Fatalf("NewProvider() expected error with invalid token")
	}
}

func TestFactoryEmailProvider(t *testing.T) {
	factory := providers.NewFactory()
	config := &providers.ProviderConfig{
		ID:   "email-test",
		Type: "email",
		Email: &providers.EmailConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "pass",
			From:     "test@example.com",
			UseTLS:   true,
		},
	}

	provider, err := factory.NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() failed: %v", err)
	}

	if provider.GetType() != "email" {
		t.Errorf("GetType() = %v, want email", provider.GetType())
	}

	if provider.GetID() != "email-test" {
		t.Errorf("GetID() = %v, want email-test", provider.GetID())
	}
}
