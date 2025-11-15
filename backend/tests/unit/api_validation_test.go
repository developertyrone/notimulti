package unit

import (
	"fmt"
	"strings"
	"testing"

	"github.com/developertyrone/notimulti/internal/api"
)

func TestValidateNotificationRequest(t *testing.T) {
	longMessage := strings.Repeat("a", 4100)
	longSubject := strings.Repeat("b", 205)
	longValue := strings.Repeat("x", 205)

	metadata := map[string]interface{}{}
	for i := 0; i < 11; i++ {
		metadata[fmt.Sprintf("key-%d", i)] = i
	}

	cases := []struct {
		name string
		req  *api.NotificationRequest
		want int
	}{
		{
			name: "missing_required_fields",
			req:  &api.NotificationRequest{},
			want: 3,
		},
		{
			name: "message_subject_metadata_priority",
			req: &api.NotificationRequest{
				ProviderID: "email-1",
				Recipient:  "user@example.com",
				Message:    longMessage,
				Subject:    longSubject,
				Metadata: map[string]interface{}{
					strings.Repeat("k", 55): "value",
					"short":                 longValue,
				},
				Priority: "urgent",
			},
			want: 5,
		},
		{
			name: "metadata_size_limit",
			req: &api.NotificationRequest{
				ProviderID: "email-1",
				Recipient:  "user@example.com",
				Message:    "hello",
				Metadata:   metadata,
			},
			want: 1,
		},
		{
			name: "valid_request",
			req: &api.NotificationRequest{
				ProviderID: "email-1",
				Recipient:  "user@example.com",
				Message:    "Hello world",
				Metadata: map[string]interface{}{
					"source": "unit-test",
				},
				Priority: "high",
			},
			want: 0,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			errs := api.ValidateNotificationRequest(tt.req)
			if len(errs) != tt.want {
				t.Fatalf("expected %d errors, got %d (%v)", tt.want, len(errs), errs)
			}
		})
	}
}

func TestValidateEmailAndTelegram(t *testing.T) {
	if api.ValidateEmailAddress("invalid") {
		t.Fatal("expected invalid email to return false")
	}
	if !api.ValidateEmailAddress("user@example.com") {
		t.Fatal("expected valid email to return true")
	}

	if !api.ValidateTelegramChatID("-123456") {
		t.Fatal("expected numeric telegram id to be valid")
	}
	if !api.ValidateTelegramChatID("@user_name") {
		t.Fatal("expected username telegram id to be valid")
	}
	if api.ValidateTelegramChatID("bad id") {
		t.Fatal("expected invalid telegram id to be false")
	}
}

func TestValidateProviderForRecipient(t *testing.T) {
	if err := api.ValidateProviderForRecipient("email", "bad"); err == nil {
		t.Fatal("expected invalid email error")
	}
	if err := api.ValidateProviderForRecipient("telegram", "not-valid"); err == nil {
		t.Fatal("expected invalid telegram error")
	}
	if err := api.ValidateProviderForRecipient("custom", "whatever"); err != nil {
		t.Fatalf("expected nil for unknown provider, got %v", err)
	}
}

func TestValidateHistoryQueryParams(t *testing.T) {
	errs := api.ValidateHistoryQueryParams("sms", "done", "bad", "also-bad", 200)
	if len(errs) != 5 {
		t.Fatalf("expected 5 validation errors, got %d", len(errs))
	}

	errs = api.ValidateHistoryQueryParams("email", "sent", "2025-11-01T00:00:00Z", "2025-11-02T00:00:00Z", 10)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for valid params, got %d", len(errs))
	}
}

func TestValidateTestRequest(t *testing.T) {
	if err := api.ValidateTestRequest("", ""); err == nil {
		t.Fatal("expected error when provider_id missing")
	}
	if err := api.ValidateTestRequest("provider-1", ""); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestMaskingHelpers(t *testing.T) {
	cfg := map[string]interface{}{
		"bot_token": "1234567890",
		"password":  "secret",
	}

	maskedTelegram := api.MaskConfig("telegram", cfg)
	if val := maskedTelegram["bot_token"].(string); val != "***7890" {
		t.Fatalf("unexpected telegram mask: %v", val)
	}

	maskedEmail := api.MaskConfig("email", cfg)
	if val := maskedEmail["password"].(string); val != "****masked****" {
		t.Fatalf("unexpected email mask: %v", val)
	}

	if masked := api.MaskSensitiveString("abcd"); masked != "****" {
		t.Fatalf("expected short mask ****, got %s", masked)
	}
	if masked := api.MaskSensitiveString("12345678"); masked != "***5678" {
		t.Fatalf("unexpected mask value %s", masked)
	}
}
