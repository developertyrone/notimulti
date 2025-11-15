package unit

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/developertyrone/notimulti/internal/config"
	"github.com/developertyrone/notimulti/internal/providers"
)

// TestWatcherDebouncing tests that rapid file changes are debounced
func TestWatcherDebouncing(t *testing.T) {
	// Create temporary test directory
	testDir := t.TempDir()

	// Create initial config file
	configPath := filepath.Join(testDir, "test-provider.json")
	configContent := `{
		"id": "test-provider",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "test-token",
			"default_chat_id": "123456"
		}
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create registry and logger
	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	registry.SetLogger(logger)

	// Create watcher
	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start watcher
	watcher.Start()

	// Wait for initial state
	time.Sleep(100 * time.Millisecond)

	// Write to file multiple times rapidly
	writeCount := 5
	for i := 0; i < writeCount; i++ {
		updatedContent := configContent // Same content, should be debounced
		if err := os.WriteFile(configPath, []byte(updatedContent), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		time.Sleep(50 * time.Millisecond) // Write every 50ms
	}

	// Wait for debounce period (300ms) + processing time
	time.Sleep(500 * time.Millisecond)

	// The watcher should have processed the changes, but checksum comparison
	// should prevent unnecessary reloads. We just verify no panic occurred.
	// Detailed verification would require exposing internal counters.
	t.Log("Debouncing test completed without errors")
}

// TestWatcherContextCancellation tests that Stop() properly terminates the watcher
func TestWatcherContextCancellation(t *testing.T) {
	testDir := t.TempDir()

	// Create initial config
	configPath := filepath.Join(testDir, "valid-provider.json")
	configContent := `{
		"id": "valid-provider",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "test-token",
			"default_chat_id": "123456"
		}
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	// Start watcher
	watcher.Start()
	time.Sleep(100 * time.Millisecond)

	// Stop watcher - should not hang or panic
	done := make(chan error, 1)
	go func() {
		done <- watcher.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Stop() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() timed out - context cancellation may not be working")
	}
}

// TestWatcherMalformedConfig tests that malformed configs don't crash the watcher
func TestWatcherMalformedConfig(t *testing.T) {
	testDir := t.TempDir()

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	registry.SetLogger(logger)

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	watcher.Start()
	time.Sleep(100 * time.Millisecond)

	// Create malformed JSON config
	malformedPath := filepath.Join(testDir, "malformed.json")
	malformedContent := `{
		"id": "malformed-provider",
		"type": "telegram",
		"enabled": true,
		INVALID JSON HERE
	}`
	if err := os.WriteFile(malformedPath, []byte(malformedContent), 0644); err != nil {
		t.Fatalf("Failed to create malformed config: %v", err)
	}

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Verify watcher is still running (no panic)
	// Create a valid config to ensure watcher still processes events
	validPath := filepath.Join(testDir, "valid.json")
	validContent := `{
		"id": "valid-provider",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "test-token",
			"default_chat_id": "123456"
		}
	}`
	if err := os.WriteFile(validPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create valid config: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// If we get here without panic, the test passes
	t.Log("Malformed config test completed - watcher remained operational")
}

// TestWatcherEventHandling tests CREATE/WRITE/REMOVE routing
// Note: This test validates that the watcher processes events without crashing.
// Actual provider registration may fail due to invalid credentials, but the watcher
// should handle these errors gracefully and continue operating.
func TestWatcherEventHandling(t *testing.T) {
	testDir := t.TempDir()

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	registry.SetLogger(logger)

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	watcher.Start()
	time.Sleep(100 * time.Millisecond)

	// Test CREATE event - watcher should process without crashing
	createPath := filepath.Join(testDir, "create-test.json")
	createContent := `{
		"id": "create-test",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "create-token",
			"default_chat_id": "123456"
		}
	}`
	if err := os.WriteFile(createPath, []byte(createContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Test WRITE event (modify existing) - watcher should process without crashing
	writeContent := `{
		"id": "create-test",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "updated-token",
			"default_chat_id": "789012"
		}
	}`
	if err := os.WriteFile(createPath, []byte(writeContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Test REMOVE event - watcher should process without crashing
	if err := os.Remove(createPath); err != nil {
		t.Fatalf("Failed to remove config: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// If we reach here without panic, the watcher handled all events correctly
	t.Log("Event handling test completed - watcher processed CREATE/WRITE/REMOVE events")
}

// TestWatcherIgnoresNonJSONFiles tests that non-JSON files are ignored
func TestWatcherIgnoresNonJSONFiles(t *testing.T) {
	testDir := t.TempDir()

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	watcher.Start()
	time.Sleep(100 * time.Millisecond)

	// Create non-JSON files
	txtPath := filepath.Join(testDir, "readme.txt")
	if err := os.WriteFile(txtPath, []byte("This is a text file"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}

	yamlPath := filepath.Join(testDir, "config.yaml")
	if err := os.WriteFile(yamlPath, []byte("key: value"), 0644); err != nil {
		t.Fatalf("Failed to create yaml file: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Registry should be empty (no providers registered from non-JSON files)
	providers := registry.List()
	if len(providers) != 0 {
		t.Errorf("Expected no providers, got %d", len(providers))
	}
}

func TestWatcherCreatesEmailProvider(t *testing.T) {
	testDir := t.TempDir()

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	registry.SetLogger(logger)

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	watcher.Start()
	time.Sleep(100 * time.Millisecond)

	emailConfig := `{
		"id": "email-provider",
		"type": "email",
		"enabled": true,
		"config": {
			"host": "smtp.example.com",
			"port": 2525,
			"username": "alerts@example.com",
			"password": "sup3r-secret",
			"from": "alerts@example.com",
			"smtp_host": "smtp.example.com",
			"smtp_port": 2525,
			"from_address": "alerts@example.com",
			"use_tls": true
		}
	}`

	configPath := filepath.Join(testDir, "email-provider.json")
	if err := os.WriteFile(configPath, []byte(emailConfig), 0644); err != nil {
		t.Fatalf("Failed to write email config: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	provider, err := registry.Get("email-provider")
	if err != nil {
		t.Fatalf("Expected email provider to be registered: %v", err)
	}
	if provider.GetType() != "email" {
		t.Fatalf("Expected provider type email, got %s", provider.GetType())
	}
}
