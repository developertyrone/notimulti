package unit

import (
	"testing"

	"github.com/developertyrone/notimulti/internal/providers"
)

func newTestEmailProvider(t *testing.T, cfg *providers.EmailConfig) *providers.EmailProvider {
	t.Helper()

	provider, err := providers.NewEmailProvider("email-test", cfg)
	if err != nil {
		t.Fatalf("failed to create email provider: %v", err)
	}
	return provider
}

func TestEmailProviderGetTestRecipient(t *testing.T) {
	cfg := &providers.EmailConfig{
		Host:          "smtp.example.com",
		Port:          587,
		Username:      "user@example.com",
		Password:      "sup3r-secret",
		From:          "from@example.com",
		TestRecipient: "test-recipient@example.com",
	}

	provider := newTestEmailProvider(t, cfg)

	recipient, err := provider.GetTestRecipient()
	if err != nil {
		t.Fatalf("expected recipient, got error %v", err)
	}
	if recipient != cfg.TestRecipient {
		t.Fatalf("expected %s, got %s", cfg.TestRecipient, recipient)
	}
}

func TestEmailProviderGetTestRecipientFallback(t *testing.T) {
	cfg := &providers.EmailConfig{
		Host:     "smtp.example.com",
		Port:     2525,
		Username: "user@example.com",
		Password: "sup3r-secret",
		From:     "alerts@example.com",
	}

	provider := newTestEmailProvider(t, cfg)

	recipient, err := provider.GetTestRecipient()
	if err != nil {
		t.Fatalf("expected fallback recipient, got error %v", err)
	}
	if recipient != cfg.From {
		t.Fatalf("expected fallback to %s, got %s", cfg.From, recipient)
	}
}

func TestEmailProviderClose(t *testing.T) {
	cfg := &providers.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user@example.com",
		Password: "secret",
		From:     "from@example.com",
	}

	provider := newTestEmailProvider(t, cfg)

	if err := provider.Close(); err != nil {
		t.Fatalf("expected Close to succeed, got %v", err)
	}
}
