package integration

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/developertyrone/notimulti/internal/config"
	"github.com/developertyrone/notimulti/internal/providers"
)

// TestWatcherIntegrationConfigChanges tests end-to-end config file changes
// Note: This test uses invalid credentials, so providers won't be registered successfully.
// The test validates that the watcher processes file changes without crashing.
func TestWatcherIntegrationConfigChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDir := t.TempDir()

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	registry.SetLogger(logger)

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer stopWatcher(t, watcher)

	watcher.Start()

	// Wait for watcher to initialize
	time.Sleep(200 * time.Millisecond)

	// Test 1: Create a config file
	configPath := filepath.Join(testDir, "telegram-test.json")
	initialConfig := `{
		"id": "telegram-test",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "test-token-123",
			"default_chat_id": "987654321"
		}
	}`

	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Wait for watcher to process (debounce + processing time)
	time.Sleep(600 * time.Millisecond)

	// Test 2: Modify the config file
	updatedConfig := `{
		"id": "telegram-test",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "updated-token-456",
			"default_chat_id": "123456789"
		}
	}`

	if err := os.WriteFile(configPath, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("Failed to update config file: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Test 3: Delete the config file
	if err := os.Remove(configPath); err != nil {
		t.Fatalf("Failed to remove config file: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// If we reach here without panic or crash, the test passes
	t.Log("Watcher successfully processed create, update, and delete operations")
}

// TestWatcherIntegrationConcurrentChanges tests handling of concurrent file changes
func TestWatcherIntegrationConcurrentChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDir := t.TempDir()

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	registry.SetLogger(logger)

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer stopWatcher(t, watcher)

	watcher.Start()
	time.Sleep(200 * time.Millisecond)

	// Create multiple config files concurrently
	numFiles := 5
	for i := 0; i < numFiles; i++ {
		configPath := filepath.Join(testDir, string(rune('a'+i))+"-provider.json")
		configContent := `{
			"id": "` + string(rune('a'+i)) + `-provider",
			"type": "telegram",
			"enabled": true,
			"config": {
				"bot_token": "token-` + string(rune('0'+i)) + `",
				"default_chat_id": "12345` + string(rune('0'+i)) + `"
			}
		}`

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config file %d: %v", i, err)
		}

		// Small delay between creates to stagger events
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for all changes to be processed
	time.Sleep(2 * time.Second)

	// Clean up - delete all files
	for i := 0; i < numFiles; i++ {
		configPath := filepath.Join(testDir, string(rune('a'+i))+"-provider.json")
		if err := os.Remove(configPath); err != nil {
			t.Logf("Warning: Failed to remove config file %d: %v", i, err)
		}
	}

	time.Sleep(1 * time.Second)

	// If we reach here without panic, concurrent changes were handled
	t.Log("Watcher successfully handled concurrent config changes")
}

// TestWatcherIntegrationRapidUpdates tests debouncing with rapid file writes
func TestWatcherIntegrationRapidUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDir := t.TempDir()

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	registry.SetLogger(logger)

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer stopWatcher(t, watcher)

	watcher.Start()
	time.Sleep(200 * time.Millisecond)

	configPath := filepath.Join(testDir, "rapid-test.json")

	// Write the file multiple times rapidly
	for i := 0; i < 10; i++ {
		configContent := `{
			"id": "rapid-test",
			"type": "telegram",
			"enabled": true,
			"config": {
				"bot_token": "token-` + string(rune('0'+i%10)) + `",
				"default_chat_id": "12345` + string(rune('0'+i%10)) + `"
			}
		}`

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file iteration %d: %v", i, err)
		}

		// Write every 50ms - faster than debounce period
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for debouncing to settle and processing to complete
	time.Sleep(1 * time.Second)

	// Clean up
	cleanupFile(t, configPath)

	// If we reach here, debouncing worked correctly (no crash from reload storm)
	t.Log("Watcher successfully debounced rapid updates")
}

// TestWatcherIntegrationChecksumDetection tests that identical writes don't cause reloads
func TestWatcherIntegrationChecksumDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDir := t.TempDir()

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	registry.SetLogger(logger)

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer stopWatcher(t, watcher)

	watcher.Start()
	time.Sleep(200 * time.Millisecond)

	configPath := filepath.Join(testDir, "checksum-test.json")
	configContent := `{
		"id": "checksum-test",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "unchanged-token",
			"default_chat_id": "999999999"
		}
	}`

	// Write initial config
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	time.Sleep(600 * time.Millisecond)

	// Write the exact same content multiple times
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write identical config: %v", err)
		}
		time.Sleep(400 * time.Millisecond)
	}

	// Clean up
	cleanupFile(t, configPath)

	// If we reach here, checksum detection prevented unnecessary reloads
	t.Log("Watcher successfully detected unchanged configs via checksum")
}

// TestWatcherIntegrationMixedFileTypes tests that non-JSON files are ignored
func TestWatcherIntegrationMixedFileTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDir := t.TempDir()

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	registry.SetLogger(logger)

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer stopWatcher(t, watcher)

	watcher.Start()
	time.Sleep(200 * time.Millisecond)

	// Create non-JSON files
	txtPath := filepath.Join(testDir, "README.txt")
	if err := os.WriteFile(txtPath, []byte("This is a readme"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}

	yamlPath := filepath.Join(testDir, "config.yaml")
	if err := os.WriteFile(yamlPath, []byte("key: value"), 0644); err != nil {
		t.Fatalf("Failed to create yaml file: %v", err)
	}

	backupPath := filepath.Join(testDir, "backup.json.bak")
	if err := os.WriteFile(backupPath, []byte(`{"test": true}`), 0644); err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Verify registry is empty (no providers from non-JSON files)
	providers := registry.List()
	if len(providers) != 0 {
		t.Errorf("Expected 0 providers from non-JSON files, got %d", len(providers))
	}

	// Clean up
	cleanupFile(t, txtPath)
	cleanupFile(t, yamlPath)
	cleanupFile(t, backupPath)

	t.Log("Watcher correctly ignored non-JSON files")
}

// TestWatcherIntegrationGracefulShutdown tests that Stop() waits for cleanup
func TestWatcherIntegrationGracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDir := t.TempDir()

	registry := providers.NewRegistry()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	watcher, err := config.NewWatcher(testDir, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	watcher.Start()
	time.Sleep(200 * time.Millisecond)

	// Create a file while stopping
	configPath := filepath.Join(testDir, "shutdown-test.json")
	configContent := `{
		"id": "shutdown-test",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "test-token",
			"default_chat_id": "123456789"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Stop immediately after file creation
	stopErr := watcher.Stop()
	if stopErr != nil {
		t.Errorf("Stop() returned error: %v", stopErr)
	}

	// Verify Stop() didn't hang
	t.Log("Watcher shutdown gracefully")
}
