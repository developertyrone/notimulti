package api

import (
	"github.com/developertyrone/notimulti/internal/providers"
	"github.com/developertyrone/notimulti/internal/storage"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter(registry *providers.Registry, logger *storage.NotificationLogger) *gin.Engine {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add recovery middleware to handle panics
	router.Use(gin.Recovery())

	// Add logging middleware
	router.Use(LoggingMiddleware())

	// Add CORS middleware
	router.Use(CORSMiddleware())

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Notification endpoints
		v1.POST("/notifications", HandleSendNotification(registry, logger))

		// Health check
		v1.GET("/health", HandleHealthCheck())

		// Provider endpoints
		v1.GET("/providers", HandleGetProviders(registry))
		v1.GET("/providers/:id", HandleGetProvider(registry))
	}

	return router
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
