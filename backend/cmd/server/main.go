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
	// T070: Use DB_PATH environment variable with default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/app/data/notifications.db" // Default for container
	}

	dbWrapper, err := storage.InitDB(dbPath)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := dbWrapper.Close(); err != nil {
			logger.Error("Failed to close database", "error", err)
		}
	}()
	logger.Info("Database initialized", "path", dbPath)

	// Initialize notification logger
	notifLogger, err := storage.NewNotificationLogger(dbWrapper.GetConn())
	if err != nil {
		logger.Error("Failed to initialize notification logger", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := notifLogger.Close(); err != nil {
			logger.Error("Failed to close notification logger", "error", err)
		}
	}()

	// Initialize repository for notification history queries
	repo := storage.NewRepository(dbWrapper.GetConn())
	logger.Info("Repository initialized")

	// Load provider configurations
	// T070: Use CONFIG_DIR environment variable with default
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "/app/configs" // Default for container
	}

	loader := config.NewLoader(configDir)
	configs, err := loader.LoadAll()
	if err != nil {
		logger.Warn("Failed to load configurations", "error", err)
	}

	// Create provider registry
	registry := providers.NewRegistry()
	registry.SetLogger(logger)

	// Load initial providers from existing config files
	factory := providers.NewFactory()
	for _, cfg := range configs {
		// Skip disabled providers
		if !cfg.Enabled {
			logger.Info("Skipping disabled provider", "id", cfg.ID, "type", cfg.Type)
			continue
		}

		// Convert config.Config to providers.ProviderConfig
		providerConfig := &providers.ProviderConfig{
			ID:   cfg.ID,
			Type: cfg.Type,
		}

		// Set type-specific config
		var parseErr error
		switch cfg.Type {
		case "telegram":
			if tgConfig, err := parseTelegramConfig(cfg.Config); err == nil {
				providerConfig.Telegram = tgConfig
			} else {
				parseErr = err
				logger.Error("Failed to parse Telegram config", "id", cfg.ID, "error", err)
			}
		case "email":
			if emailConfig, err := parseEmailConfig(cfg.Config); err == nil {
				providerConfig.Email = emailConfig
			} else {
				parseErr = err
				logger.Error("Failed to parse Email config", "id", cfg.ID, "error", err)
			}
		default:
			parseErr = fmt.Errorf("unknown provider type: %s", cfg.Type)
			logger.Error("Unknown provider type", "id", cfg.ID, "type", cfg.Type)
		}

		// Create provider (or failed provider if parsing failed)
		var provider providers.Provider
		if parseErr != nil {
			// Register as failed provider so it shows in UI with error
			provider = providers.NewFailedProvider(cfg.ID, cfg.Type, parseErr)
			logger.Warn("Registering provider with parse error", "id", cfg.ID, "error", parseErr)
		} else {
			// Try to create the actual provider
			var err error
			provider, err = factory.NewProvider(providerConfig)
			if err != nil {
				// Provider initialization failed - register as failed provider
				provider = providers.NewFailedProvider(cfg.ID, cfg.Type, err)
				logger.Warn("Registering provider with initialization error", "id", cfg.ID, "error", err)
			} else {
				logger.Info("Provider loaded successfully", "id", cfg.ID, "type", cfg.Type)
			}
		}

		// Always register the provider (even if failed)
		if err := registry.Register(provider); err != nil {
			logger.Error("Failed to register provider", "id", cfg.ID, "error", err)
			if closeErr := provider.Close(); closeErr != nil {
				logger.Error("Failed to close provider after registration failure", "id", cfg.ID, "error", closeErr)
			}
			continue
		}
	}

	logger.Info("Provider registry initialized", "count", registry.Count())

	// Start configuration file watcher
	watcher, err := config.NewWatcher(configDir, registry, logger)
	if err != nil {
		logger.Error("Failed to initialize configuration watcher", "error", err)
		os.Exit(1)
	}
	watcher.Start()
	logger.Info("Configuration watcher started", "directory", configDir)

	// Setup API router
	router := api.SetupRouter(registry, notifLogger, repo)

	// Get server port
	// T070: Use PORT environment variable (standard for containers)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
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

	// Stop configuration watcher
	if err := watcher.Stop(); err != nil {
		logger.Error("Error stopping watcher", "error", err)
	}

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

// Helper functions to parse type-specific configs
func parseTelegramConfig(config map[string]interface{}) (*providers.TelegramConfig, error) {
	tgConfig := &providers.TelegramConfig{}

	if botToken, ok := config["bot_token"].(string); ok {
		tgConfig.BotToken = botToken
	} else {
		return nil, fmt.Errorf("missing or invalid bot_token")
	}

	if chatID, ok := config["default_chat_id"].(string); ok {
		tgConfig.DefaultChatID = chatID
	}

	if parseMode, ok := config["parse_mode"].(string); ok {
		tgConfig.ParseMode = parseMode
	}

	if timeout, ok := config["timeout_seconds"].(float64); ok {
		tgConfig.TimeoutSeconds = int(timeout)
	}

	return tgConfig, nil
}

func parseEmailConfig(config map[string]interface{}) (*providers.EmailConfig, error) {
	emailConfig := &providers.EmailConfig{}

	if host, ok := config["host"].(string); ok {
		emailConfig.Host = host
	} else {
		return nil, fmt.Errorf("missing or invalid host")
	}

	if port, ok := config["port"].(float64); ok {
		emailConfig.Port = int(port)
	} else {
		return nil, fmt.Errorf("missing or invalid port")
	}

	if username, ok := config["username"].(string); ok {
		emailConfig.Username = username
	}

	if password, ok := config["password"].(string); ok {
		emailConfig.Password = password
	}

	if from, ok := config["from"].(string); ok {
		emailConfig.From = from
	} else {
		return nil, fmt.Errorf("missing or invalid from")
	}

	if useTLS, ok := config["use_tls"].(bool); ok {
		emailConfig.UseTLS = useTLS
	}

	if timeout, ok := config["timeout_seconds"].(float64); ok {
		emailConfig.TimeoutSeconds = int(timeout)
	}

	return emailConfig, nil
}
