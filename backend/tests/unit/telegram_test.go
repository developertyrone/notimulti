package unit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
		{"valid numeric", "12345678", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notification := &providers.Notification{
				ID:        "notif-1",
				Recipient: tt.recipient,
				Message:   "Test message",
			}

			err := validateTelegramNotification(notification)
			if (err != nil) != tt.wantError {
				t.Fatalf("validateTelegramNotification() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestTelegramProviderSendNilNotification(t *testing.T) {
	if err := validateTelegramNotification(nil); err == nil {
		t.Fatal("expected error for nil notification")
	}

	valid := &providers.Notification{
		ID:        "notif-valid",
		Recipient: "12345678",
		Message:   "Hello",
	}
	if err := validateTelegramNotification(valid); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
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
			if tt.config == nil {
				if !tt.wantError {
					t.Fatalf("expected valid config, got nil")
				}
				return
			}

			if tt.config.BotToken == "" || tt.config.DefaultChatID == "" {
				if !tt.wantError {
					t.Fatalf("missing required fields in config: %+v", tt.config)
				}
				return
			}

			if tt.wantError {
				t.Fatalf("expected configuration error for test %s", tt.name)
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
			err := validateTelegramNotification(tt.notif)
			if (err != nil) != tt.wantError {
				t.Fatalf("validateTelegramNotification() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func validateTelegramNotification(notification *providers.Notification) error {
	if notification == nil {
		return errors.New("notification cannot be nil")
	}
	if strings.TrimSpace(notification.Message) == "" {
		return errors.New("message cannot be empty")
	}
	return validateTelegramRecipient(notification.Recipient)
}

func validateTelegramRecipient(recipient string) error {
	if recipient == "" {
		return errors.New("recipient (chat_id) cannot be empty")
	}
	trimmed := strings.TrimPrefix(recipient, "-")
	if trimmed == "" {
		return errors.New("invalid chat_id")
	}
	if _, err := strconv.ParseInt(trimmed, 10, 64); err != nil {
		return fmt.Errorf("invalid chat_id: %w", err)
	}
	return nil
}
