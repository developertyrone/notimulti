package unit

import (
	"database/sql"
	"testing"
	"time"

	"github.com/developertyrone/notimulti/internal/storage"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestRepository creates a sqlite-backed repository for unit tests.
func setupTestRepository(t *testing.T) (*storage.Repository, *sql.DB) {
	t.Helper()

	dbPath := t.TempDir() + "/repo.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	if _, err := db.Exec(storage.Schema); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}

	return storage.NewRepository(db), db
}

// insertLog creates a minimal notification log entry returning its ID.
func insertLog(t *testing.T, db *sql.DB, providerID, providerType, status string, created time.Time, isTest bool) int {
	t.Helper()

	var deliveredAt interface{}
	if status == storage.StatusSent {
		deliveredAt = created.Format("2006-01-02 15:04:05")
	}
	createdValue := created.Format("2006-01-02 15:04:05")

	res, err := db.Exec(`INSERT INTO notification_logs (
        provider_id, provider_type, recipient, message, subject, metadata,
        priority, status, error_message, attempts, created_at, delivered_at, is_test
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		providerID,
		providerType,
		"recipient@example.com",
		"hello",
		nil,
		"{}",
		"normal",
		status,
		"",
		1,
		createdValue,
		deliveredAt,
		boolToInt(isTest),
	)
	if err != nil {
		t.Fatalf("failed to insert log: %v", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to fetch last insert id: %v", err)
	}
	return int(lastID)
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
