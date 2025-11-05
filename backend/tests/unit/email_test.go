package unit

import (
	"context"
	"testing"

	"github.com/developertyrone/notimulti/internal/providers"
)

func TestNewEmailProviderValidation(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		config  *providers.EmailConfig
		wantErr bool
	}{
		{
			name:    "nil_config",
			id:      "email-1",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty_host",
			id:   "email-1",
			config: &providers.EmailConfig{
				Host:     "",
				Port:     587,
				Username: "user",
				Password: "pass",
				From:     "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "empty_port",
			id:   "email-1",
			config: &providers.EmailConfig{
				Host:     "smtp.example.com",
				Port:     0,
				Username: "user",
				Password: "pass",
				From:     "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "empty_from",
			id:   "email-1",
			config: &providers.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user",
				Password: "pass",
				From:     "",
			},
			wantErr: true,
		},
		{
			name: "valid_config",
			id:   "email-1",
			config: &providers.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "user@gmail.com",
				Password: "password",
				From:     "sender@example.com",
				UseTLS:   true,
			},
			wantErr: false,
		},
		{
			name: "valid_config_with_timeout",
			id:   "email-1",
			config: &providers.EmailConfig{
				Host:           "smtp.gmail.com",
				Port:           587,
				Username:       "user@gmail.com",
				Password:       "password",
				From:           "sender@example.com",
				UseTLS:         true,
				TimeoutSeconds: 60,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := providers.NewEmailProvider(tt.id, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmailProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *providers.EmailConfig
		wantErr bool
	}{
		{
			name:    "nil",
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid",
			config: &providers.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user",
				Password: "pass",
				From:     "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "with_timeout",
			config: &providers.EmailConfig{
				Host:           "smtp.example.com",
				Port:           587,
				Username:       "user",
				Password:       "pass",
				From:           "test@example.com",
				TimeoutSeconds: 45,
			},
			wantErr: false,
		},
		{
			name: "with_tls",
			config: &providers.EmailConfig{
				Host:     "smtp.example.com",
				Port:     465,
				Username: "user",
				Password: "pass",
				From:     "test@example.com",
				UseTLS:   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := providers.NewEmailProvider("email-test", tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailConfig validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailNotificationValidation(t *testing.T) {
	config := &providers.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "test@example.com",
	}

	provider, err := providers.NewEmailProvider("email-1", config)
	if err != nil {
		t.Fatalf("Failed to create email provider: %v", err)
	}

	tests := []struct {
		name         string
		notification *providers.Notification
		wantErr      bool
	}{
		{
			name:         "nil_notification",
			notification: nil,
			wantErr:      true,
		},
		{
			name: "empty_recipient",
			notification: &providers.Notification{
				ID:         "notif-1",
				ProviderID: "email-1",
				Recipient:  "",
				Message:    "Test message",
			},
			wantErr: true,
		},
		{
			name: "invalid_email_format",
			notification: &providers.Notification{
				ID:         "notif-1",
				ProviderID: "email-1",
				Recipient:  "invalid-email",
				Message:    "Test message",
			},
			wantErr: true,
		},
		{
			name: "valid_email",
			notification: &providers.Notification{
				ID:         "notif-1",
				ProviderID: "email-1",
				Recipient:  "user@example.com",
				Message:    "Test message",
			},
			wantErr: true, // Will fail due to connection, but validates format
		},
		{
			name: "valid_with_subject",
			notification: &providers.Notification{
				ID:         "notif-1",
				ProviderID: "email-1",
				Recipient:  "user@example.com",
				Message:    "Test message",
				Subject:    "Test Subject",
			},
			wantErr: true, // Will fail due to connection, but validates format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := provider.Send(ctx, tt.notification)

			// For valid formats, we expect connection errors (not format errors)
			if tt.name == "valid_email" || tt.name == "valid_with_subject" {
				if err == nil {
					t.Errorf("Send() expected error for test environment, got nil")
				}
				// Check that it's not a format validation error
				if err.Error() == "invalid email format: user@example.com" {
					t.Errorf("Send() returned format error for valid email: %v", err)
				}
			} else if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailProviderGetStatus(t *testing.T) {
	config := &providers.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "test@example.com",
	}

	provider, err := providers.NewEmailProvider("email-1", config)
	if err != nil {
		t.Fatalf("Failed to create email provider: %v", err)
	}

	status := provider.GetStatus()
	if status == nil {
		t.Errorf("GetStatus() returned nil")
	}
	// Status will show error due to connectivity in test environment
	if status.Status == "" {
		t.Errorf("GetStatus() returned empty status")
	}
}

func TestEmailProviderGetID(t *testing.T) {
	config := &providers.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "test@example.com",
	}

	provider, err := providers.NewEmailProvider("email-test-id", config)
	if err != nil {
		t.Fatalf("Failed to create email provider: %v", err)
	}

	if got := provider.GetID(); got != "email-test-id" {
		t.Errorf("GetID() = %v, want email-test-id", got)
	}
}

func TestEmailProviderGetType(t *testing.T) {
	config := &providers.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "test@example.com",
	}

	provider, err := providers.NewEmailProvider("email-1", config)
	if err != nil {
		t.Fatalf("Failed to create email provider: %v", err)
	}

	if got := provider.GetType(); got != "email" {
		t.Errorf("GetType() = %v, want email", got)
	}
}
