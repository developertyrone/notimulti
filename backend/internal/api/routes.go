package api

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/developertyrone/notimulti/internal/providers"
	"github.com/developertyrone/notimulti/internal/storage"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter(registry *providers.Registry, logger *storage.NotificationLogger, repo *storage.Repository) *gin.Engine {
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
		v1.GET("/notifications/history", HandleGetNotificationHistory(repo))
		v1.GET("/notifications/:id", HandleGetNotificationDetail(repo))

		// Health check
		v1.GET("/health", HandleHealthCheck())
		v1.GET("/ready", HandleReadinessCheck(registry, repo))

		// Provider endpoints
		v1.GET("/providers", HandleGetProviders(registry))
		v1.GET("/providers/:id", HandleGetProvider(registry))
		v1.POST("/providers/:id/test", HandleTestProvider(registry, logger))
	}

	return router
}

// ServeFrontend serves embedded frontend static files (T069)
func ServeFrontend(router *gin.Engine, frontendFS embed.FS) {
	// Extract the dist subdirectory from the embedded filesystem
	distFS, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		// If dist doesn't exist (dev mode), just return
		return
	}

	// Serve static files from /assets/*
	router.StaticFS("/assets", http.FS(distFS))

	// Serve index.html for all non-API routes (SPA routing)
	router.NoRoute(func(c *gin.Context) {
		// Don't serve frontend for API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// Serve index.html for all other routes
		data, err := distFS.Open("index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Frontend not available")
			return
		}
		defer func() {
			if closeErr := data.Close(); closeErr != nil {
				_ = c.Error(closeErr)
			}
		}()

		c.DataFromReader(http.StatusOK, -1, "text/html", data, nil)
	})
}

// ServeFrontendFromDisk serves frontend files from a directory on disk (Docker runtime path).
// This is used when the dist assets are copied into the image instead of embedded.
func ServeFrontendFromDisk(router *gin.Engine, distPath string) {
	// Ensure the directory exists; if not, skip registration
	info, err := os.Stat(distPath)
	if err != nil || !info.IsDir() {
		return
	}

	// Serve static assets (Vite outputs /assets/*)
	router.StaticFS("/assets", gin.Dir(filepath.Join(distPath, "assets"), false))

	// SPA fallback for all non-API routes
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		c.File(filepath.Join(distPath, "index.html"))
	})
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowOrigin := "*"
		if origin != "" {
			allowOrigin = origin
		}

		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", allowOrigin)
		h.Set("Vary", "Origin")
		h.Set("Access-Control-Allow-Credentials", "true")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		h.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
