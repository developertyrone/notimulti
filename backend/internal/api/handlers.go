package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
				status := "delivered"
				errorMsg := ""
				if err != nil {
					status = "failed"
					errorMsg = err.Error()
				}

				if logErr := logger.LogNotification(notification, status, errorMsg, provider.GetType(), attempts); logErr != nil {
					fmt.Printf("Failed to log notification %s: %v\n", notificationID, logErr)
				}
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
