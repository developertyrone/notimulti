package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/developertyrone/notimulti/internal/providers"
	"github.com/fsnotify/fsnotify"
)

// Watcher watches configuration directory for changes
type Watcher struct {
	configDir string
	registry  *providers.Registry
	loader    *Loader
	factory   *providers.Factory
	watcher   *fsnotify.Watcher
	logger    *slog.Logger
	debouncer *debouncer
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// debouncer handles debouncing of rapid file writes
type debouncer struct {
	mu      sync.Mutex
	timers  map[string]*time.Timer
	delay   time.Duration
	handler func(string)
}

// NewWatcher creates a new configuration file watcher
func NewWatcher(configDir string, registry *providers.Registry, logger *slog.Logger) (*Watcher, error) {
	loader := NewLoader(configDir)
	factory := providers.NewFactory()

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	w := &Watcher{
		configDir: configDir,
		registry:  registry,
		loader:    loader,
		factory:   factory,
		watcher:   fsWatcher,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize debouncer with 300ms delay
	w.debouncer = &debouncer{
		timers:  make(map[string]*time.Timer),
		delay:   300 * time.Millisecond,
		handler: w.handleFileChange,
	}

	// Add config directory to watch list
	if err := fsWatcher.Add(configDir); err != nil {
		cancel()
		if closeErr := fsWatcher.Close(); closeErr != nil {
			logger.Warn("Failed to close fsnotify watcher after error", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to watch directory %s: %w", configDir, err)
	}

	w.logger.Info("Configuration watcher initialized", "directory", configDir)

	return w, nil
}

// Start begins watching for configuration changes
func (w *Watcher) Start() {
	w.wg.Add(1)
	go w.run()
}

// run is the main event loop for the watcher
func (w *Watcher) run() {
	defer w.wg.Done()

	w.logger.Info("Configuration watcher started")

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("Configuration watcher stopped")
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only process JSON files
			if !strings.HasSuffix(event.Name, ".json") {
				continue
			}

			// Only handle CREATE, WRITE, and REMOVE events
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) {
				w.logger.Debug("File event detected",
					"file", filepath.Base(event.Name),
					"operation", event.Op.String())

				// Debounce the event
				w.debouncer.debounce(event.Name)
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.logger.Error("Watcher error", "error", err)
		}
	}
}

// debounce handles debouncing for rapid file writes
func (d *debouncer) debounce(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Cancel existing timer if any
	if timer, exists := d.timers[path]; exists {
		timer.Stop()
	}

	// Create new timer
	d.timers[path] = time.AfterFunc(d.delay, func() {
		d.handler(path)
		d.mu.Lock()
		delete(d.timers, path)
		d.mu.Unlock()
	})
}

// handleFileChange processes a file change after debouncing
func (w *Watcher) handleFileChange(path string) {
	filename := filepath.Base(path)
	configID := strings.TrimSuffix(filename, ".json")

	w.logger.Info("Processing configuration change",
		"file", filename,
		"config_id", configID)

	// Check if file still exists (handles REMOVE events)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		w.handleRemove(configID)
		return
	}

	// File exists, so it's either CREATE or WRITE
	w.handleCreateOrWrite(path, configID)
}

// handleCreateOrWrite processes CREATE and WRITE events
func (w *Watcher) handleCreateOrWrite(path, configID string) {
	// Load the configuration
	config, err := w.loader.LoadFile(path)
	if err != nil {
		w.logger.Error("Failed to load configuration",
			"file", filepath.Base(path),
			"error", err)
		return
	}

	// Skip disabled providers
	if !config.Enabled {
		w.logger.Info("Skipping disabled provider", "id", config.ID, "type", config.Type)
		// If provider was previously registered, remove it
		if _, err := w.registry.Get(configID); err == nil {
			w.handleRemove(configID)
		}
		return
	}

	// Check if provider already exists
	existingProvider, err := w.registry.Get(configID)

	if err != nil {
		// Provider doesn't exist - this is a CREATE event
		w.handleCreate(config)
	} else {
		// Provider exists - check if config actually changed
		existingStatus := existingProvider.GetStatus()
		if existingStatus != nil && existingStatus.ConfigChecksum == config.Checksum {
			w.logger.Debug("Configuration unchanged (same checksum), skipping reload",
				"config_id", configID,
				"checksum", config.Checksum)
			return
		}

		// Config changed - this is a WRITE event
		w.handleWrite(config)
	}
}

// handleCreate processes CREATE events (new provider)
func (w *Watcher) handleCreate(config *ProviderConfig) {
	w.logger.Info("Creating new provider",
		"id", config.ID,
		"type", config.Type,
		"checksum", config.Checksum)

	// Convert config.Config to providers.ProviderConfig
	providerConfig := &providers.ProviderConfig{
		ID:   config.ID,
		Type: config.Type,
	}

	// Set type-specific config
	var parseErr error
	switch config.Type {
	case "telegram":
		if tgConfig, err := parseTelegramConfig(config.Config); err == nil {
			providerConfig.Telegram = tgConfig
		} else {
			parseErr = err
			w.logger.Error("Failed to parse Telegram config",
				"id", config.ID,
				"error", err)
		}
	case "email":
		if emailConfig, err := parseEmailConfig(config.Config); err == nil {
			providerConfig.Email = emailConfig
		} else {
			parseErr = err
			w.logger.Error("Failed to parse Email config",
				"id", config.ID,
				"error", err)
		}
	default:
		parseErr = fmt.Errorf("unknown provider type: %s", config.Type)
		w.logger.Error("Unknown provider type",
			"id", config.ID,
			"type", config.Type)
	}

	// Create provider (or failed provider if parsing failed)
	var provider providers.Provider
	if parseErr != nil {
		// Create failed provider so it shows in UI with error
		provider = providers.NewFailedProvider(config.ID, config.Type, parseErr)
		w.logger.Warn("Registering provider with parse error",
			"id", config.ID,
			"error", parseErr)
	} else {
		// Try to create the actual provider
		var err error
		provider, err = w.factory.NewProvider(providerConfig)
		if err != nil {
			// Provider initialization failed - create failed provider
			provider = providers.NewFailedProvider(config.ID, config.Type, err)
			w.logger.Warn("Registering provider with initialization error",
				"id", config.ID,
				"error", err)
		}
	}

	// Update provider status with checksum
	if status := provider.GetStatus(); status != nil {
		status.ConfigChecksum = config.Checksum
	}

	// Register the provider (even if failed)
	if err := w.registry.Register(provider); err != nil {
		w.logger.Error("Failed to register provider",
			"id", config.ID,
			"error", err)
		if closeErr := provider.Close(); closeErr != nil {
			w.logger.Error("Failed to close provider after registration failure",
				"id", config.ID,
				"error", closeErr)
		}
		return
	}

	w.logger.Info("Provider created successfully",
		"id", config.ID,
		"type", config.Type)
}

// handleWrite processes WRITE events (provider update)
func (w *Watcher) handleWrite(config *ProviderConfig) {
	w.logger.Info("Updating existing provider",
		"id", config.ID,
		"type", config.Type,
		"checksum", config.Checksum)

	// Convert config.Config to providers.ProviderConfig
	providerConfig := &providers.ProviderConfig{
		ID:   config.ID,
		Type: config.Type,
	}

	// Set type-specific config
	var parseErr error
	switch config.Type {
	case "telegram":
		if tgConfig, err := parseTelegramConfig(config.Config); err == nil {
			providerConfig.Telegram = tgConfig
		} else {
			parseErr = err
			w.logger.Error("Failed to parse Telegram config",
				"id", config.ID,
				"error", err)
		}
	case "email":
		if emailConfig, err := parseEmailConfig(config.Config); err == nil {
			providerConfig.Email = emailConfig
		} else {
			parseErr = err
			w.logger.Error("Failed to parse Email config",
				"id", config.ID,
				"error", err)
		}
	default:
		parseErr = fmt.Errorf("unknown provider type: %s", config.Type)
		w.logger.Error("Unknown provider type",
			"id", config.ID,
			"type", config.Type)
	}

	// Create new provider (or failed provider if parsing failed)
	var newProvider providers.Provider
	if parseErr != nil {
		// Create failed provider so it shows in UI with error
		newProvider = providers.NewFailedProvider(config.ID, config.Type, parseErr)
		w.logger.Warn("Replacing with provider in error state (parse error)",
			"id", config.ID,
			"error", parseErr)
	} else {
		// Try to create the actual provider
		var err error
		newProvider, err = w.factory.NewProvider(providerConfig)
		if err != nil {
			// Provider initialization failed - create failed provider
			newProvider = providers.NewFailedProvider(config.ID, config.Type, err)
			w.logger.Warn("Replacing with provider in error state (initialization error)",
				"id", config.ID,
				"error", err)
		}
	}

	// Update provider status with checksum
	if status := newProvider.GetStatus(); status != nil {
		status.ConfigChecksum = config.Checksum
	}

	// Replace the provider atomically (even if failed)
	if err := w.registry.Replace(config.ID, newProvider); err != nil {
		w.logger.Error("Failed to replace provider",
			"id", config.ID,
			"error", err)
		if closeErr := newProvider.Close(); closeErr != nil {
			w.logger.Error("Failed to close provider after replace failure",
				"id", config.ID,
				"error", closeErr)
		}
		return
	}

	w.logger.Info("Provider updated successfully",
		"id", config.ID,
		"type", config.Type)
}

// handleRemove processes REMOVE events (provider deletion)
func (w *Watcher) handleRemove(configID string) {
	w.logger.Info("Removing provider",
		"id", configID)

	if err := w.registry.Remove(configID); err != nil {
		w.logger.Error("Failed to remove provider",
			"id", configID,
			"error", err)
		return
	}

	w.logger.Info("Provider removed successfully",
		"id", configID)
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

	// Accept both legacy keys (host/port/from) and newer smtp_* keys.
	if host, ok := config["smtp_host"].(string); ok && host != "" {
		emailConfig.Host = host
	} else if host, ok := config["host"].(string); ok && host != "" {
		emailConfig.Host = host
	} else {
		return nil, fmt.Errorf("missing or invalid smtp_host/host")
	}

	if port, ok := config["smtp_port"].(float64); ok && port > 0 {
		emailConfig.Port = int(port)
	} else if port, ok := config["port"].(float64); ok && port > 0 {
		emailConfig.Port = int(port)
	} else {
		return nil, fmt.Errorf("missing or invalid smtp_port/port")
	}

	if username, ok := config["username"].(string); ok {
		emailConfig.Username = username
	}

	if password, ok := config["password"].(string); ok {
		emailConfig.Password = password
	}

	if from, ok := config["from_address"].(string); ok && from != "" {
		emailConfig.From = from
	} else if from, ok := config["from"].(string); ok && from != "" {
		emailConfig.From = from
	} else {
		return nil, fmt.Errorf("missing or invalid from_address/from")
	}

	if useTLS, ok := config["use_tls"].(bool); ok {
		emailConfig.UseTLS = useTLS
	}

	if timeout, ok := config["timeout_seconds"].(float64); ok {
		emailConfig.TimeoutSeconds = int(timeout)
	}

	return emailConfig, nil
}

// Stop stops the watcher and waits for cleanup
func (w *Watcher) Stop() error {
	w.logger.Info("Stopping configuration watcher")

	// Cancel context to stop event loop
	w.cancel()

	// Close the file watcher
	if err := w.watcher.Close(); err != nil {
		w.logger.Error("Error closing watcher", "error", err)
	}

	// Wait for goroutine to finish
	w.wg.Wait()

	// Stop all pending debounce timers
	w.debouncer.mu.Lock()
	for _, timer := range w.debouncer.timers {
		timer.Stop()
	}
	w.debouncer.mu.Unlock()

	w.logger.Info("Configuration watcher stopped")
	return nil
}
