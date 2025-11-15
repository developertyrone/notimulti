package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/developertyrone/notimulti/internal/logging"
	"github.com/developertyrone/notimulti/internal/providers"
	"github.com/developertyrone/notimulti/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NotificationRequest represents the incoming notification request
type NotificationRequest struct {
	ProviderID string                 `json:"provider_id" binding:"required"`
	Recipient  string                 `json:"recipient" binding:"required"`
	Message    string                 `json:"message" binding:"required"`
	Subject    string                 `json:"subject,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Priority   string                 `json:"priority,omitempty"`
}

// NotificationResponse represents the response after sending a notification
type NotificationResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// HandleSendNotification handles POST /api/v1/notifications
func HandleSendNotification(registry *providers.Registry, logger *storage.NotificationLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req NotificationRequest

		// Bind JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid request body",
				"details": err.Error(),
			})
			return
		}

		// Validate request
		validationErrors := ValidateNotificationRequest(&req)
		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation failed",
				"details": validationErrors,
			})
			return
		}

		// Get provider from registry
		provider, err := registry.Get(req.ProviderID)
		if err != nil || provider == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("provider not found: %s", req.ProviderID),
			})
			return
		}

		// Generate notification ID
		notificationID := uuid.New().String()
		timestamp := time.Now()

		// Create notification object
		notification := &providers.Notification{
			ID:         notificationID,
			ProviderID: req.ProviderID,
			Recipient:  req.Recipient,
			Message:    req.Message,
			Subject:    req.Subject,
			Metadata:   req.Metadata,
			Priority:   req.Priority,
			Timestamp:  timestamp,
		}

		// Send notification asynchronously
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Track attempts for logging
			attempts := 1
			err := provider.Send(ctx, notification)

			// Log to database
			if logger != nil {
				status := storage.StatusSent
				errorMsg := ""
				deliveredAt := time.Now().Format(time.RFC3339)
				if err != nil {
					status = storage.StatusFailed
					errorMsg = err.Error()
					deliveredAt = ""
				}

				logEntry := storage.LogEntry{
					Notification: notification,
					Status:       status,
					ErrorMessage: errorMsg,
					ProviderType: provider.GetType(),
					Attempts:     attempts,
					DeliveredAt:  deliveredAt,
					IsTest:       false,
				}
				logger.Log(logEntry)
			}

			// Log to console for debugging
			if err != nil {
				fmt.Printf("Error sending notification %s: %v\n", notificationID, err)
			} else {
				fmt.Printf("Notification %s sent successfully\n", notificationID)
			}
		}()

		// Return 201 with notification ID
		c.JSON(http.StatusCreated, NotificationResponse{
			ID:        notificationID,
			Status:    "queued",
			Timestamp: timestamp,
			Message:   "notification queued for delivery",
		})
	}
}

// HandleHealthCheck handles GET /api/v1/health
func HandleHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{
			Status:    "ok",
			Version:   "1.0.0",
			Timestamp: time.Now(),
		})
	}
}

// HandleGetProviders handles GET /api/v1/providers
func HandleGetProviders(registry *providers.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		providersList := registry.List()
		response := make([]gin.H, len(providersList))

		for i, p := range providersList {
			status := p.GetStatus()
			summary := gin.H{
				"id":           p.GetID(),
				"type":         p.GetType(),
				"status":       status.Status,
				"last_updated": status.LastUpdated,
			}

			// Include error message if present
			if status.ErrorMessage != "" {
				summary["error_message"] = status.ErrorMessage
			}

			// T056: Include last test metadata
			if status.LastTestAt != nil {
				summary["last_test_at"] = status.LastTestAt.Format(time.RFC3339)
			}
			if status.LastTestStatus != "" {
				summary["last_test_status"] = status.LastTestStatus
			}

			response[i] = summary
		}

		c.JSON(http.StatusOK, gin.H{
			"providers": response,
			"count":     len(response),
		})
	}
}

// HandleGetProvider handles GET /api/v1/providers/:id
func HandleGetProvider(registry *providers.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		provider, err := registry.Get(id)

		if err != nil || provider == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("provider not found: %s", id),
			})
			return
		}

		status := provider.GetStatus()

		response := gin.H{
			"id":              provider.GetID(),
			"type":            provider.GetType(),
			"status":          status.Status,
			"last_updated":    status.LastUpdated,
			"config_checksum": status.ConfigChecksum,
		}

		// Include error message if present
		if status.ErrorMessage != "" {
			response["error_message"] = status.ErrorMessage
		}

		c.JSON(http.StatusOK, response)
	}
}

// HandleGetNotificationHistory handles GET /api/v1/notifications/history
func HandleGetNotificationHistory(repo *storage.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		filters := storage.HistoryFilters{
			ProviderID:   c.Query("provider_id"),
			ProviderType: c.Query("provider_type"),
			Status:       c.Query("status"),
			DateFrom:     c.Query("date_from"),
			DateTo:       c.Query("date_to"),
			IncludeTests: c.Query("include_tests") != "false", // Default true
			Cursor:       0,
			PageSize:     50, // Default page size
			SortOrder:    "DESC",
		}

		// Parse cursor if provided
		if cursorStr := c.Query("cursor"); cursorStr != "" {
			if _, err := fmt.Sscanf(cursorStr, "%d", &filters.Cursor); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "cursor must be a valid integer",
				})
				return
			}
		}

		// Parse page size if provided
		if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
			if _, err := fmt.Sscanf(pageSizeStr, "%d", &filters.PageSize); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "page_size must be a valid integer",
				})
				return
			}
			if filters.PageSize < 1 || filters.PageSize > 100 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "page_size must be between 1 and 100",
				})
				return
			}
		}

		// Get notification history
		notifications, nextCursor, err := repo.GetNotificationHistory(filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to retrieve notification history",
			})
			return
		}

		// Build response
		response := gin.H{
			"notifications": notifications,
			"pagination": gin.H{
				"page_size": filters.PageSize,
				"has_more":  nextCursor != nil,
			},
		}

		if nextCursor != nil {
			response["pagination"].(gin.H)["next_cursor"] = *nextCursor
		}

		c.JSON(http.StatusOK, response)
	}
}

// HandleGetNotificationDetail handles GET /api/v1/notifications/:id
func HandleGetNotificationDetail(repo *storage.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse ID from path
		var id int
		if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid notification ID",
			})
			return
		}

		// Get notification by ID
		notification, err := repo.GetNotificationByID(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to retrieve notification",
			})
			return
		}

		if notification == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "notification not found",
			})
			return
		}

		c.JSON(http.StatusOK, notification)
	}
}

// HandleReadinessCheck handles GET /api/v1/ready (T071)
func HandleReadinessCheck(registry *providers.Registry, repo *storage.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		checks := gin.H{}

		// T071: Check database connection with "SELECT 1" query
		dbOk := true
		if err := repo.Ping(); err != nil {
			checks["database"] = fmt.Sprintf("error: %v", err)
			dbOk = false
		} else {
			checks["database"] = "ok"
		}

		// Check providers loaded
		providers := registry.List()
		if len(providers) > 0 {
			checks["providers"] = "ok"
		} else {
			checks["providers"] = "no providers loaded"
		}

		// T071: Determine overall status - return 503 if any check fails
		allOk := dbOk && len(providers) > 0

		if allOk {
			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
				"checks": checks,
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"checks": checks,
			})
		}
	}
}

// HandleTestProvider handles POST /api/v1/providers/:id/test (T053)
func HandleTestProvider(registry *providers.Registry, logger *storage.NotificationLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Get provider from registry
		provider, err := registry.Get(id)
		if err != nil || provider == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "PROVIDER_NOT_FOUND",
				"message": fmt.Sprintf("Provider with ID '%s' not found", id),
			})
			return
		}

		// T054: Check rate limit (10 seconds between tests)
		status := provider.GetStatus()
		if status.LastTestAt != nil {
			timeSinceLastTest := time.Since(*status.LastTestAt)
			if timeSinceLastTest < 10*time.Second {
				retryAfter := int((10*time.Second - timeSinceLastTest).Seconds())
				c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code":    "RATE_LIMITED",
					"message": "Provider test already in progress or tested recently. Please wait.",
				})
				return
			}
		}

		// T055: Structured logging for test operation
		logging.LogWithContext(c.Request.Context()).Info("Provider test initiated",
			"test_initiator", "UI",
			"provider_id", id,
			"provider_type", provider.GetType(),
		)

		// T053: Call provider Test() method
		testCtx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		testErr := provider.Test(testCtx)
		testedAt := time.Now()

		// Prepare response
		response := gin.H{
			"tested_at": testedAt.Format(time.RFC3339),
		}

		if testErr != nil {
			// Test failed
			response["result"] = "failed"
			response["message"] = "Test notification failed"
			response["error_details"] = testErr.Error()

			// T055: Log test failure
			logging.LogWithContext(c.Request.Context()).Error("Provider test failed",
				"provider_id", id,
				"result", "failed",
				"error", testErr.Error(),
				"tested_at", testedAt.Format(time.RFC3339),
			)

			c.JSON(http.StatusOK, response) // Return 200 even on test failure
		} else {
			// Test succeeded
			response["result"] = "success"
			response["message"] = fmt.Sprintf("Test notification sent successfully to provider %s", id)

			// T055: Log test success
			logging.LogWithContext(c.Request.Context()).Info("Provider test succeeded",
				"provider_id", id,
				"result", "success",
				"tested_at", testedAt.Format(time.RFC3339),
			)

			c.JSON(http.StatusOK, response)
		}
	}
}
