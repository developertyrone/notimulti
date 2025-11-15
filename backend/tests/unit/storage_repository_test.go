package unit

import (
	"testing"
	"time"

	"github.com/developertyrone/notimulti/internal/storage"
)

func TestRepositoryHistoryAndCursor(t *testing.T) {
	repo, db := setupTestRepository(t)
	defer closeSQLDB(t, db)

	base := time.Now().Add(-2 * time.Hour)
	firstID := insertLog(t, db, "email-1", "email", storage.StatusSent, base, false)
	insertLog(t, db, "email-1", "email", storage.StatusFailed, base.Add(10*time.Minute), false)
	insertLog(t, db, "email-1", "email", storage.StatusPending, base.Add(20*time.Minute), true)

	filters := storage.HistoryFilters{
		ProviderID:   "email-1",
		IncludeTests: false,
		PageSize:     1,
		SortOrder:    "DESC",
	}

	entries, cursor, err := repo.GetNotificationHistory(filters)
	if err != nil {
		t.Fatalf("GetNotificationHistory failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID == firstID {
		t.Fatalf("expected newest entry first")
	}
	if cursor == nil {
		t.Fatalf("expected next cursor when more records exist")
	}
}

func TestRepositoryGetNotificationByIDAndCleanup(t *testing.T) {
	repo, db := setupTestRepository(t)
	defer closeSQLDB(t, db)

	created := time.Now().Add(-48 * time.Hour)
	id := insertLog(t, db, "email-2", "email", storage.StatusFailed, created, true)

	entry, err := repo.GetNotificationByID(id)
	if err != nil {
		t.Fatalf("GetNotificationByID failed: %v", err)
	}
	if entry == nil || entry.ID != id {
		t.Fatalf("expected entry with id %d", id)
	}
	if !entry.IsTest {
		t.Fatalf("expected IsTest flag to persist")
	}

	if err := repo.CleanupOldLogs(1); err != nil {
		t.Fatalf("CleanupOldLogs failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM notification_logs").Scan(&count); err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected logs to be deleted, remaining=%d", count)
	}

	entry, err = repo.GetNotificationByID(id)
	if err != nil {
		t.Fatalf("GetNotificationByID after cleanup failed: %v", err)
	}
	if entry != nil {
		t.Fatalf("expected nil entry after cleanup")
	}
}

func TestRepositoryHistoryFiltersAdvanced(t *testing.T) {
	repo, db := setupTestRepository(t)
	defer closeSQLDB(t, db)

	oldest := time.Now().Add(-3 * time.Hour)
	first := insertLog(t, db, "email-advanced", "email", storage.StatusSent, oldest, false)
	second := insertLog(t, db, "email-advanced", "email", storage.StatusFailed, oldest.Add(30*time.Minute), false)
	insertLog(t, db, "telegram-advanced", "telegram", storage.StatusSent, oldest.Add(time.Hour), true)

	filters := storage.HistoryFilters{
		ProviderID:   "email-advanced",
		ProviderType: "email",
		Status:       storage.StatusFailed,
		DateFrom:     oldest.Add(-5 * time.Minute).Format("2006-01-02 15:04:05"),
		DateTo:       time.Now().Add(5 * time.Minute).Format("2006-01-02 15:04:05"),
		IncludeTests: false,
		Cursor:       second + 1,
		PageSize:     5,
		SortOrder:    "DESC",
	}

	entries, cursor, err := repo.GetNotificationHistory(filters)
	if err != nil {
		t.Fatalf("GetNotificationHistory with advanced filters failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry matching filters, got %d", len(entries))
	}
	if entries[0].ID != second {
		t.Fatalf("expected entry id %d, got %d", second, entries[0].ID)
	}
	if cursor != nil {
		t.Fatalf("expected no next cursor when results under page size")
	}

	// Now use cursor to fetch older records
	filters.Status = ""
	filters.Cursor = second
	filters.PageSize = 1
	entries, cursor, err = repo.GetNotificationHistory(filters)
	if err != nil {
		t.Fatalf("GetNotificationHistory with cursor failed: %v", err)
	}
	if len(entries) != 1 || entries[0].ID != first {
		t.Fatalf("expected to page to first entry, got %+v", entries)
	}
	if cursor != nil {
		t.Fatalf("expected nil cursor when no more data")
	}
}

func TestRepositoryPing(t *testing.T) {
	repo, db := setupTestRepository(t)
	defer closeSQLDB(t, db)

	if err := repo.Ping(); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}
