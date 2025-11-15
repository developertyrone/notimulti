package unit

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/developertyrone/notimulti/internal/providers"
)

func TestTelegramProviderSendAndTest(t *testing.T) {
	var sendAttempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "getMe"):
			writeTelegramResponse(t, w, `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Unit","username":"unit_bot"}}`)
		case strings.Contains(r.URL.Path, "sendMessage"):
			attempt := atomic.AddInt32(&sendAttempts, 1)
			if attempt == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				writeTelegramResponse(t, w, `{"ok":false,"error_code":429,"parameters":{"retry_after":1},"description":"Too Many Requests"}`)
				return
			}
			writeTelegramResponse(t, w, `{"ok":true,"result":{"message_id":42}}`)
		default:
			writeTelegramResponse(t, w, `{"ok":true}`)
		}
	}))
	defer server.Close()

	cfg := &providers.TelegramConfig{
		BotToken:       "token",
		DefaultChatID:  "5551234",
		ParseMode:      "HTML",
		TimeoutSeconds: 5,
		APIEndpoint:    server.URL + "/bot%s/%s",
	}

	provider, err := providers.NewTelegramProvider("telegram-unit", cfg)
	if err != nil {
		t.Fatalf("failed to create telegram provider: %v", err)
	}
	defer closeTelegramProvider(t, provider)

	notification := &providers.Notification{
		ID:         "notif-1",
		ProviderID: "telegram-unit",
		Recipient:  "-1234567",
		Message:    "hello world",
		Subject:    "greetings",
		Timestamp:  time.Now(),
	}

	if err := provider.Send(context.Background(), notification); err != nil {
		t.Fatalf("expected send to succeed, got %v", err)
	}
	if atomic.LoadInt32(&sendAttempts) < 2 {
		t.Fatalf("expected retry logic to run at least once")
	}

	status := provider.GetStatus()
	if status == nil || status.Status != providers.StatusActive {
		t.Fatalf("expected active status, got %+v", status)
	}

	if provider.GetID() != "telegram-unit" {
		t.Fatalf("unexpected provider id %s", provider.GetID())
	}
	if provider.GetType() != "telegram" {
		t.Fatalf("unexpected provider type %s", provider.GetType())
	}

	recipient, err := provider.GetTestRecipient()
	if err != nil || recipient != cfg.DefaultChatID {
		t.Fatalf("expected test recipient %s, got %s (err=%v)", cfg.DefaultChatID, recipient, err)
	}

	if err := provider.Test(context.Background()); err != nil {
		t.Fatalf("expected test notification to succeed, got %v", err)
	}
}

func writeTelegramResponse(t *testing.T, w http.ResponseWriter, payload string) {
	t.Helper()
	if _, err := fmt.Fprint(w, payload); err != nil {
		t.Fatalf("failed to write telegram response: %v", err)
	}
}

func closeTelegramProvider(t *testing.T, provider providers.Provider) {
	t.Helper()
	if provider == nil {
		return
	}
	if err := provider.Close(); err != nil {
		t.Fatalf("failed to close telegram provider: %v", err)
	}
}
