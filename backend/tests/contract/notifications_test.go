package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/developertyrone/notimulti/internal/api"
	"github.com/developertyrone/notimulti/internal/providers"
)

// setupTestRouter creates a test router with a mock registry
func setupTestRouter() *httptest.Server {
	registry := providers.NewRegistry()

	// Note: We can only add Email provider in tests since Telegram requires valid token
	// Add mock Email provider
	emailConfig := &providers.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "test@example.com",
	}
	emailProvider, _ := providers.NewEmailProvider("email-test", emailConfig)
	if emailProvider != nil {
		mustRegisterProvider(registry, emailProvider)
	}

	router := api.SetupRouter(registry, nil, nil)
	return httptest.NewServer(router)
}

func closeBody(t *testing.T, closer io.Closer) {
	t.Helper()
	if err := closer.Close(); err != nil {
		t.Fatalf("Failed to close response body: %v", err)
	}
}

func mustRegisterProvider(registry *providers.Registry, provider providers.Provider) {
	if err := registry.Register(provider); err != nil {
		panic(fmt.Sprintf("failed to register provider %s: %v", provider.GetID(), err))
	}
}

func TestPostNotificationValidTelegram(t *testing.T) {
	t.Skip("Skipping Telegram test - requires valid bot token from environment")
	// This test would require TELEGRAM_BOT_TOKEN env var to be set
	// It's covered by integration tests instead
}

func TestPostNotificationValidEmail(t *testing.T) {
	ts := setupTestRouter()
	defer ts.Close()

	payload := map[string]interface{}{
		"provider_id": "email-test",
		"recipient":   "user@example.com",
		"message":     "Test email message",
		"subject":     "Test Subject",
		"priority":    "normal",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/api/v1/notifications", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer closeBody(t, resp.Body)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["id"]; !ok {
		t.Errorf("Response missing 'id' field")
	}
}

func TestPostNotificationInvalidProvider(t *testing.T) {
	ts := setupTestRouter()
	defer ts.Close()

	payload := map[string]interface{}{
		"provider_id": "nonexistent-provider",
		"recipient":   "12345678",
		"message":     "Test message",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/api/v1/notifications", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer closeBody(t, resp.Body)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Errorf("Response should contain 'error' field for 404")
	}
}

func TestPostNotificationMissingRequiredFields(t *testing.T) {
	ts := setupTestRouter()
	defer ts.Close()

	tests := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "missing_provider_id",
			payload: map[string]interface{}{
				"recipient": "12345678",
				"message":   "Test message",
			},
		},
		{
			name: "missing_recipient",
			payload: map[string]interface{}{
				"provider_id": "telegram-test",
				"message":     "Test message",
			},
		},
		{
			name: "missing_message",
			payload: map[string]interface{}{
				"provider_id": "telegram-test",
				"recipient":   "12345678",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			resp, err := http.Post(ts.URL+"/api/v1/notifications", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer closeBody(t, resp.Body)

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", resp.StatusCode)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if _, ok := response["error"]; !ok {
				t.Errorf("Response should contain 'error' field for 400")
			}
		})
	}
}

func TestPostNotificationMessageExceeds4096Chars(t *testing.T) {
	ts := setupTestRouter()
	defer ts.Close()

	// Create a message longer than 4096 characters
	longMessage := string(make([]byte, 4097))
	for i := range longMessage {
		longMessage = longMessage[:i] + "a" + longMessage[i+1:]
	}

	payload := map[string]interface{}{
		"provider_id": "telegram-test",
		"recipient":   "12345678",
		"message":     longMessage,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/api/v1/notifications", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer closeBody(t, resp.Body)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for message exceeding limit, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["details"]; !ok {
		t.Errorf("Response should contain 'details' field with limit information")
	}
}

func TestPostNotificationMetadataExceedsLimits(t *testing.T) {
	ts := setupTestRouter()
	defer ts.Close()

	tests := []struct {
		name     string
		metadata map[string]interface{}
	}{
		{
			name: "too_many_keys",
			metadata: map[string]interface{}{
				"key1":  "value1",
				"key2":  "value2",
				"key3":  "value3",
				"key4":  "value4",
				"key5":  "value5",
				"key6":  "value6",
				"key7":  "value7",
				"key8":  "value8",
				"key9":  "value9",
				"key10": "value10",
				"key11": "value11", // Exceeds 10 pairs
			},
		},
		{
			name: "key_too_long",
			metadata: map[string]interface{}{
				string(make([]byte, 51)): "value", // 51 chars, exceeds 50
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize long key properly
			if tt.name == "key_too_long" {
				longKey := ""
				for i := 0; i < 51; i++ {
					longKey += "a"
				}
				tt.metadata = map[string]interface{}{
					longKey: "value",
				}
			}

			payload := map[string]interface{}{
				"provider_id": "telegram-test",
				"recipient":   "12345678",
				"message":     "Test message",
				"metadata":    tt.metadata,
			}

			body, _ := json.Marshal(payload)
			resp, err := http.Post(ts.URL+"/api/v1/notifications", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer closeBody(t, resp.Body)

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected status 400 for metadata exceeding limits, got %d", resp.StatusCode)
			}
		})
	}
}
