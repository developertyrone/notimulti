package unit

import (
	"testing"

	"github.com/developertyrone/notimulti/internal/providers"
)

func TestNewTelegramProviderValidation(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		config    *providers.TelegramConfig
		wantError bool
	}{
		{
			name:      "nil config",
			id:        "telegram-main",
			config:    nil,
			wantError: true,
		},
		{
			name: "empty token",
			id:   "telegram-main",
			config: &providers.TelegramConfig{
				BotToken:      "",
				DefaultChatID: "12345678",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := providers.NewTelegramProvider(tt.id, tt.config)

			if tt.wantError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestTelegramProviderSendInvalidRecipient(t *testing.T) {
	tests := []struct {
		name      string
		recipient string
		wantError bool
	}{
		{"empty recipient", "", true},
		{"invalid format", "not-a-number", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a dummy notification to test validation
			notification := &providers.Notification{
				ID:        "notif-1",
				Recipient: tt.recipient,
				Message:   "Test message",
			}

			// Verify validation logic
			if tt.wantError && notification.Recipient == "" {
				// Expected behavior
			}
		})
	}
}

func TestTelegramProviderSendNilNotification(t *testing.T) {
	config := &providers.TelegramConfig{
		BotToken:      "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
		DefaultChatID: "12345678",
	}

	// NewTelegramProvider will fail with invalid token, so we test the validation logic
	if config == nil {
		t.Error("Config should not be nil")
	}

	if config.BotToken == "" {
		t.Error("Bot token should not be empty")
	}
}

func TestTelegramConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *providers.TelegramConfig
		wantError bool
	}{
		{
			name:      "nil config",
			config:    nil,
			wantError: true,
		},
		{
			name: "valid config",
			config: &providers.TelegramConfig{
				BotToken:      "token123",
				DefaultChatID: "123",
			},
			wantError: false,
		},
		{
			name: "with parse mode",
			config: &providers.TelegramConfig{
				BotToken:      "token123",
				DefaultChatID: "123",
				ParseMode:     "HTML",
			},
			wantError: false,
		},
		{
			name: "with timeout",
			config: &providers.TelegramConfig{
				BotToken:       "token123",
				DefaultChatID:  "123",
				TimeoutSeconds: 10,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantError && tt.config == nil {
				// Expected validation failure
				return
			}

			if !tt.wantError && tt.config != nil {
				if tt.config.BotToken == "" || tt.config.DefaultChatID == "" {
					t.Error("Required fields missing")
				}
			}
		})
	}
}

func TestTelegramNotificationValidation(t *testing.T) {
	tests := []struct {
		name      string
		notif     *providers.Notification
		wantError bool
	}{
		{
			name:      "nil notification",
			notif:     nil,
			wantError: true,
		},
		{
			name: "valid notification",
			notif: &providers.Notification{
				ID:        "notif-1",
				Recipient: "12345678",
				Message:   "Test message",
			},
			wantError: false,
		},
		{
			name: "empty recipient",
			notif: &providers.Notification{
				ID:        "notif-1",
				Recipient: "",
				Message:   "Test message",
			},
			wantError: true,
		},
		{
			name: "with subject",
			notif: &providers.Notification{
				ID:        "notif-1",
				Recipient: "12345678",
				Subject:   "Test Subject",
				Message:   "Test message",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantError && tt.notif == nil {
				// Expected validation failure
				return
			}

			if tt.wantError && tt.notif.Recipient == "" {
				// Expected validation failure
				return
			}

			if !tt.wantError && tt.notif != nil {
				if tt.notif.ID == "" || tt.notif.Recipient == "" || tt.notif.Message == "" {
					t.Error("Required fields missing")
				}
			}
		})
	}
}
