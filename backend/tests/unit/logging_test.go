package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/developertyrone/notimulti/internal/logging"
)

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected slog.Level
	}{
		{"default (no env)", "", slog.LevelInfo},
		{"INFO", "INFO", slog.LevelInfo},
		{"DEBUG", "DEBUG", slog.LevelDebug},
		{"WARN", "WARN", slog.LevelWarn},
		{"ERROR", "ERROR", slog.LevelError},
		{"invalid", "INVALID", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				if err := os.Setenv("LOG_LEVEL", tt.envValue); err != nil {
					t.Fatalf("Failed to set LOG_LEVEL: %v", err)
				}
			} else {
				if err := os.Unsetenv("LOG_LEVEL"); err != nil {
					t.Fatalf("Failed to unset LOG_LEVEL: %v", err)
				}
			}
			defer func() {
				if err := os.Unsetenv("LOG_LEVEL"); err != nil {
					t.Fatalf("Failed to unset LOG_LEVEL: %v", err)
				}
			}()

			// Test indirectly through InitLogger
			logger := logging.InitLogger()
			if logger == nil {
				t.Error("InitLogger returned nil")
			}
		})
	}
}

func TestGetLogFormat(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{"default", "", "json"},
		{"json", "json", "json"},
		{"text", "text", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Setenv("LOG_FORMAT", tt.envValue); err != nil {
				t.Fatalf("Failed to set LOG_FORMAT: %v", err)
			}
			defer func() {
				if err := os.Unsetenv("LOG_FORMAT"); err != nil {
					t.Fatalf("Failed to unset LOG_FORMAT: %v", err)
				}
			}()

			// Test indirectly through InitLogger
			logger := logging.InitLogger()
			if logger == nil {
				t.Error("InitLogger returned nil")
			}
		})
	}
}

func TestRedactSensitive(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  string
	}{
		{"token", "bot_token", "1234567890abcdef", "****cdef"},
		{"password", "smtp_password", "secret123", "****t123"},
		{"short_secret", "api_key", "abc", "****"},
		{"non_sensitive", "chat_id", "1234567890", "1234567890"},
		{"mixed_case", "API_TOKEN", "mytoken123", "****n123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := logging.RedactSensitive(tt.key, tt.value)
			if got != tt.want {
				t.Errorf("RedactSensitive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-123"

	ctx = logging.WithRequestID(ctx, requestID)
	got := logging.GetRequestID(ctx)

	if got != requestID {
		t.Errorf("GetRequestID() = %v, want %v", got, requestID)
	}
}

func TestLogWithContext(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-456"
	ctx = logging.WithRequestID(ctx, requestID)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(logger)

	logging.LogWithContext(ctx).Info("test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	if logEntry["request_id"] != requestID {
		t.Errorf("Log entry request_id = %v, want %v", logEntry["request_id"], requestID)
	}

	if logEntry["msg"] != "test message" {
		t.Errorf("Log entry msg = %v, want 'test message'", logEntry["msg"])
	}
}

func TestInitLogger_JSONFormat(t *testing.T) {
	if err := os.Setenv("LOG_FORMAT", "json"); err != nil {
		t.Fatalf("Failed to set LOG_FORMAT: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "DEBUG"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("LOG_FORMAT"); err != nil {
			t.Fatalf("Failed to unset LOG_FORMAT: %v", err)
		}
	}()
	defer func() {
		if err := os.Unsetenv("LOG_LEVEL"); err != nil {
			t.Fatalf("Failed to unset LOG_LEVEL: %v", err)
		}
	}()

	logger := logging.InitLogger()
	if logger == nil {
		t.Error("InitLogger() returned nil")
	}
}

func TestInitLogger_TextFormat(t *testing.T) {
	if err := os.Setenv("LOG_FORMAT", "text"); err != nil {
		t.Fatalf("Failed to set LOG_FORMAT: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "INFO"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("LOG_FORMAT"); err != nil {
			t.Fatalf("Failed to unset LOG_FORMAT: %v", err)
		}
	}()
	defer func() {
		if err := os.Unsetenv("LOG_LEVEL"); err != nil {
			t.Fatalf("Failed to unset LOG_LEVEL: %v", err)
		}
	}()

	logger := logging.InitLogger()
	if logger == nil {
		t.Error("InitLogger() returned nil")
	}
}

func TestSensitiveDataRedaction(t *testing.T) {
	// Test that multiple sensitive keywords are detected
	sensitiveKeys := []string{"token", "password", "secret", "key", "auth"}

	for _, key := range sensitiveKeys {
		t.Run(key, func(t *testing.T) {
			value := "sensitive_value_12345"
			result := logging.RedactSensitive(key, value)

			if !strings.Contains(result, "****") {
				t.Errorf("Expected redaction for key %s, got %s", key, result)
			}
		})
	}
}
