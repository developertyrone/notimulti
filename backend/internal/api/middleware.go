package api

import (
	"time"

	"github.com/developertyrone/notimulti/internal/logging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggingMiddleware logs all HTTP requests with structured logging
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Add request ID to context
		ctx := logging.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request
		duration := time.Since(start)
		logger := logging.LogWithContext(ctx)

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
