package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// InitLogger initializes the structured logger based on environment variables
func InitLogger() *slog.Logger {
	logLevel := getLogLevel()
	logFormat := getLogFormat()

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	if logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// getLogLevel reads LOG_LEVEL env var and returns appropriate slog.Level
func getLogLevel() slog.Level {
	levelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if levelStr == "" {
		levelStr = "INFO"
	}

	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// getLogFormat reads LOG_FORMAT env var
func getLogFormat() string {
	format := strings.ToLower(os.Getenv("LOG_FORMAT"))
	if format == "" {
		return "json"
	}
	return format
}

// WithRequestID adds request ID to context for correlation
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// RedactSensitive redacts sensitive data from log attributes
func RedactSensitive(key string, value string) string {
	sensitiveKeys := []string{"token", "password", "secret", "key", "auth"}
	lowerKey := strings.ToLower(key)

	for _, sensitive := range sensitiveKeys {
		if strings.Contains(lowerKey, sensitive) {
			if len(value) > 4 {
				return "****" + value[len(value)-4:]
			}
			return "****"
		}
	}
	return value
}

// LogWithContext creates a logger with request ID from context
func LogWithContext(ctx context.Context) *slog.Logger {
	logger := slog.Default()
	if reqID := GetRequestID(ctx); reqID != "" {
		logger = logger.With("request_id", reqID)
	}
	return logger
}
