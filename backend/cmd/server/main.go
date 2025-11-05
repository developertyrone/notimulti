package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/developertyrone/notimulti/internal/api"
	"github.com/developertyrone/notimulti/internal/config"
	"github.com/developertyrone/notimulti/internal/logging"
	"github.com/developertyrone/notimulti/internal/providers"
	"github.com/developertyrone/notimulti/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file (if exists)
	_ = godotenv.Load()

	// Initialize logger
	logger := logging.InitLogger()
	logger.Info("Starting notification server", "version", "1.0.0")

	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/notifications.db"
	}

	dbWrapper, err := storage.InitDB(dbPath)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer dbWrapper.Close()
	logger.Info("Database initialized", "path", dbPath)

	// Initialize notification logger
	notifLogger, err := storage.NewNotificationLogger(dbWrapper.GetConn())
	if err != nil {
		logger.Error("Failed to initialize notification logger", "error", err)
		os.Exit(1)
	}
	defer notifLogger.Close()

	// Load provider configurations
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "./configs"
	}

	loader := config.NewLoader(configDir)
	_, err = loader.LoadAll()
	if err != nil {
		logger.Warn("Failed to load configurations", "error", err)
	}

	// Create provider registry
	registry := providers.NewRegistry()

	// Note: Provider loading from config files will be implemented in Phase 4 (T048-T050)
	// For now, providers must be registered programmatically for testing
	logger.Info("Provider registry initialized", "count", registry.Count())

	// Setup router with registry and logger
	router := api.SetupRouter(registry, notifLogger)

	// Get server port
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Graceful shutdown with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	// Close all providers
	for _, provider := range registry.List() {
		if err := provider.Close(); err != nil {
			logger.Error("Failed to close provider", "id", provider.GetID(), "error", err)
		}
	}

	logger.Info("Server stopped")
}
