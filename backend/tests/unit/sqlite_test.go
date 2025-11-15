package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/developertyrone/notimulti/internal/storage"
)

func TestInitDBCreatesSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "notimulti.db")

	db, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("failed to remove db file: %v", err)
		}
	})

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	conn := db.GetConn()
	if conn == nil {
		t.Fatalf("expected underlying sql.DB to be non-nil")
	}

	// Ensure schema exists by counting rows in notification_logs
	var count int
	if err := conn.QueryRow("SELECT COUNT(*) FROM notification_logs").Scan(&count); err != nil {
		t.Fatalf("failed to query schema: %v", err)
	}
}

func TestDBCloseHandlesNilConnection(t *testing.T) {
	var db storage.DB // zero value has nil connection
	if err := db.Close(); err != nil {
		t.Fatalf("expected nil close error, got %v", err)
	}
}
