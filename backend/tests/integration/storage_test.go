package integration

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/developertyrone/notimulti/internal/storage"
)

func TestDatabaseInitialization(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize database
	db, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Verify connection is alive
	if err := db.Ping(); err != nil {
		t.Errorf("Ping() failed: %v", err)
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestSchemaCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Verify notification_logs table exists
	var tableName string
	err = db.GetConn().QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='notification_logs'",
	).Scan(&tableName)

	if err != nil {
		t.Fatalf("Table 'notification_logs' not found: %v", err)
	}

	if tableName != "notification_logs" {
		t.Errorf("Expected table name 'notification_logs', got '%s'", tableName)
	}
}

func TestTableStructure(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Get table info
	rows, err := db.GetConn().Query("PRAGMA table_info(notification_logs)")
	if err != nil {
		t.Fatalf("Failed to get table info: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]bool{
		"id":            false,
		"provider_id":   false,
		"provider_type": false,
		"recipient":     false,
		"message":       false,
		"subject":       false,
		"metadata":      false,
		"priority":      false,
		"status":        false,
		"error_message": false,
		"attempts":      false,
		"created_at":    false,
		"delivered_at":  false,
	}

	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, dfltValue, pk sql.NullString

		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}

		if _, exists := expectedColumns[name]; exists {
			expectedColumns[name] = true
		}
	}

	// Verify all expected columns exist
	for col, found := range expectedColumns {
		if !found {
			t.Errorf("Expected column '%s' not found in table", col)
		}
	}
}

func TestIndexCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Get indexes
	rows, err := db.GetConn().Query(
		"SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='notification_logs'",
	)
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}
	defer rows.Close()

	expectedIndexes := map[string]bool{
		"idx_provider_created": false,
		"idx_status_created":   false,
		"idx_type_created":     false,
		"idx_created_id":       false,
	}

	foundCount := 0
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("Failed to scan index name: %v", err)
		}

		if _, exists := expectedIndexes[name]; exists {
			expectedIndexes[name] = true
			foundCount++
		}
	}

	// Verify all expected indexes exist
	for idx, found := range expectedIndexes {
		if !found {
			t.Errorf("Expected index '%s' not found", idx)
		}
	}

	if foundCount < len(expectedIndexes) {
		t.Errorf("Expected %d indexes, found %d", len(expectedIndexes), foundCount)
	}
}

func TestWALMode(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Check journal mode
	var mode string
	err = db.GetConn().QueryRow("PRAGMA journal_mode").Scan(&mode)
	if err != nil {
		t.Fatalf("Failed to get journal_mode: %v", err)
	}

	if mode != "wal" {
		t.Errorf("Expected journal_mode 'wal', got '%s'", mode)
	}
}

func TestBusyTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Check busy timeout
	var timeout int
	err = db.GetConn().QueryRow("PRAGMA busy_timeout").Scan(&timeout)
	if err != nil {
		t.Fatalf("Failed to get busy_timeout: %v", err)
	}

	if timeout != 5000 {
		t.Errorf("Expected busy_timeout 5000, got %d", timeout)
	}
}

func TestConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Perform concurrent writes
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			_, err := db.GetConn().Exec(`
				INSERT INTO notification_logs 
				(provider_id, provider_type, recipient, message, status) 
				VALUES (?, ?, ?, ?, ?)`,
				"test-provider",
				"test",
				"recipient@example.com",
				"Test message",
				storage.StatusPending,
			)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent write failed: %v", err)
	}

	// Verify all records were inserted
	var count int
	err = db.GetConn().QueryRow("SELECT COUNT(*) FROM notification_logs").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	if count != 10 {
		t.Errorf("Expected 10 records, got %d", count)
	}
}

func TestMultipleInitializations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// First initialization
	db1, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("First InitDB() failed: %v", err)
	}
	db1.Close()

	// Second initialization (should not fail due to IF NOT EXISTS)
	db2, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Second InitDB() failed: %v", err)
	}
	defer db2.Close()

	// Verify schema is still intact
	var tableName string
	err = db2.GetConn().QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='notification_logs'",
	).Scan(&tableName)

	if err != nil {
		t.Errorf("Table verification failed after second init: %v", err)
	}
}

func TestInvalidDatabasePath(t *testing.T) {
	// Try to create database in non-existent directory without proper path
	dbPath := "/nonexistent/path/test.db"

	_, err := storage.InitDB(dbPath)
	if err == nil {
		t.Error("Expected error for invalid database path, got nil")
	}
}
