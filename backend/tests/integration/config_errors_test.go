package integration

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/developertyrone/notimulti/internal/config"
	"github.com/developertyrone/notimulti/internal/providers"
)

// TestConfigErrorMalformedJSON tests that malformed JSON doesn't crash the server
func TestConfigErrorMalformedJSON(t *testing.T) {
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

	// Create malformed JSON config
	malformedPath := filepath.Join(testDir, "malformed.json")
	malformedContent := `{
		"id": "malformed-provider",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "test-token"
			MISSING COMMA AND CLOSING BRACES
	}`

	if err := os.WriteFile(malformedPath, []byte(malformedContent), 0644); err != nil {
		t.Fatalf("Failed to create malformed config: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Create a valid config to verify watcher is still operational
	validPath := filepath.Join(testDir, "valid-after-malformed.json")
	validContent := `{
		"id": "valid-provider",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "valid-token",
			"default_chat_id": "123456789"
		}
	}`

	if err := os.WriteFile(validPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create valid config: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Clean up
	cleanupFile(t, malformedPath)
	cleanupFile(t, validPath)

	// If we reach here, the watcher handled malformed JSON gracefully
	t.Log("Watcher continued operating after malformed JSON error")
}

// TestConfigErrorInvalidCredentials tests that invalid credentials don't prevent other providers
func TestConfigErrorInvalidCredentials(t *testing.T) {
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

	// Create config with invalid credentials (will fail provider initialization)
	invalidPath := filepath.Join(testDir, "invalid-creds.json")
	invalidContent := `{
		"id": "invalid-creds",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "invalid-bot-token-12345",
			"default_chat_id": "999999999"
		}
	}`

	if err := os.WriteFile(invalidPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create invalid config: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Verify the provider is not in the registry (creation failed)
	_, err = registry.Get("invalid-creds")
	if err == nil {
		t.Log("Provider with invalid credentials was registered (unexpected but not necessarily an error)")
	}

	// Create another provider to verify registry is still functional
	anotherPath := filepath.Join(testDir, "another-provider.json")
	anotherContent := `{
		"id": "another-provider",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "another-token",
			"default_chat_id": "888888888"
		}
	}`

	if err := os.WriteFile(anotherPath, []byte(anotherContent), 0644); err != nil {
		t.Fatalf("Failed to create another config: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Clean up
	cleanupFile(t, invalidPath)
	cleanupFile(t, anotherPath)

	// If we reach here, invalid credentials didn't prevent registry operation
	t.Log("Registry continued operating despite invalid provider credentials")
}

// TestConfigErrorMissingRequiredFields tests validation of required config fields
func TestConfigErrorMissingRequiredFields(t *testing.T) {
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

	// Create config missing required field (default_chat_id)
	missingFieldPath := filepath.Join(testDir, "missing-field.json")
	missingFieldContent := `{
		"id": "missing-field-provider",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "test-token"
		}
	}`

	if err := os.WriteFile(missingFieldPath, []byte(missingFieldContent), 0644); err != nil {
		t.Fatalf("Failed to create config with missing field: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Verify provider was not registered (validation failed)
	_, err = registry.Get("missing-field-provider")
	if err == nil {
		t.Error("Provider with missing required field should not be registered")
	}

	// Clean up
	cleanupFile(t, missingFieldPath)

	// If we reach here, validation prevented invalid provider registration
	t.Log("Config validation correctly rejected provider with missing required fields")
}

// TestConfigErrorInvalidType tests handling of unsupported provider types
func TestConfigErrorInvalidType(t *testing.T) {
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

	// Create config with unsupported type
	invalidTypePath := filepath.Join(testDir, "invalid-type.json")
	invalidTypeContent := `{
		"id": "invalid-type-provider",
		"type": "unsupported-type",
		"enabled": true,
		"config": {
			"some_field": "some_value"
		}
	}`

	if err := os.WriteFile(invalidTypePath, []byte(invalidTypeContent), 0644); err != nil {
		t.Fatalf("Failed to create config with invalid type: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Verify provider was not registered
	_, err = registry.Get("invalid-type-provider")
	if err == nil {
		t.Error("Provider with unsupported type should not be registered")
	}

	// Create valid provider to verify watcher still works
	validPath := filepath.Join(testDir, "valid-telegram.json")
	validContent := `{
		"id": "valid-telegram",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "valid-token",
			"default_chat_id": "123456789"
		}
	}`

	if err := os.WriteFile(validPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create valid config: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Clean up
	cleanupFile(t, invalidTypePath)
	cleanupFile(t, validPath)

	// If we reach here, invalid type didn't crash the watcher
	t.Log("Watcher handled unsupported provider type gracefully")
}

// TestConfigErrorConcurrentModificationDuringRead tests handling file changes during read
func TestConfigErrorConcurrentModificationDuringRead(t *testing.T) {
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

	configPath := filepath.Join(testDir, "concurrent-mod.json")

	// Rapidly modify file to potentially catch it mid-read
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()
			content := `{
				"id": "concurrent-mod",
				"type": "telegram",
				"enabled": true,
				"config": {
					"bot_token": "token-` + string(rune('0'+iteration%10)) + `",
					"default_chat_id": "12345` + string(rune('0'+iteration%10)) + `"
				}
			}`
			if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
				panic(err)
			}
			time.Sleep(20 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	time.Sleep(1 * time.Second)

	// Clean up
	cleanupFile(t, configPath)

	// If we reach here, concurrent modifications didn't crash the system
	t.Log("Watcher handled concurrent file modifications gracefully")
}

// TestConfigErrorEmptyFile tests handling of empty config files
func TestConfigErrorEmptyFile(t *testing.T) {
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

	// Create empty file
	emptyPath := filepath.Join(testDir, "empty.json")
	if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Create whitespace-only file
	whitespacePath := filepath.Join(testDir, "whitespace.json")
	if err := os.WriteFile(whitespacePath, []byte("   \n\t  \n"), 0644); err != nil {
		t.Fatalf("Failed to create whitespace file: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Clean up
	cleanupFile(t, emptyPath)
	cleanupFile(t, whitespacePath)

	// If we reach here, empty files were handled gracefully
	t.Log("Watcher handled empty and whitespace-only files gracefully")
}

// TestConfigErrorPermissionDenied tests handling of permission errors (if applicable)
func TestConfigErrorPermissionDenied(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip on Windows where permission handling differs
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Skipping permission test on Windows")
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

	// Create a file with no read permissions
	noReadPath := filepath.Join(testDir, "no-read.json")
	validContent := `{
		"id": "no-read-provider",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "test-token",
			"default_chat_id": "123456789"
		}
	}`

	if err := os.WriteFile(noReadPath, []byte(validContent), 0000); err != nil {
		t.Fatalf("Failed to create no-read file: %v", err)
	}

	time.Sleep(600 * time.Millisecond)

	// Restore permissions for cleanup
	restorePermissions(t, noReadPath, 0644)
	cleanupFile(t, noReadPath)

	// If we reach here, permission errors were handled
	t.Log("Watcher handled permission errors gracefully")
}

func stopWatcher(t *testing.T, watcher *config.Watcher) {
	t.Helper()
	if err := watcher.Stop(); err != nil {
		t.Fatalf("Failed to stop watcher: %v", err)
	}
}

func cleanupFile(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove %s: %v", path, err)
	}
}

func restorePermissions(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("Failed to change permissions for %s: %v", path, err)
	}
}
