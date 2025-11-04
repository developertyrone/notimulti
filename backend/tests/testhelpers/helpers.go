package testhelpers

import (
	"database/sql"
	"os"
	"testing"
)

// MockProvider is a simple mock implementation for testing
type MockProvider struct {
	ID           string
	Type         string
	SendFunc     func() error
	StatusFunc   func() string
	CloseFunc    func() error
}

func (m *MockProvider) Send() error {
	if m.SendFunc != nil {
		return m.SendFunc()
	}
	return nil
}

func (m *MockProvider) GetStatus() string {
	if m.StatusFunc != nil {
		return m.StatusFunc()
	}
	return "active"
}

func (m *MockProvider) GetID() string {
	return m.ID
}

func (m *MockProvider) GetType() string {
	return m.Type
}

func (m *MockProvider) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// SetupTestDB creates a temporary test database
func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	
	dbPath := t.TempDir() + "/test.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}
