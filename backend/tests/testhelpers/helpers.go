package testhelpers

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/developertyrone/notimulti/internal/providers"
)

// MockProvider is a simple mock implementation for testing
type MockProvider struct {
	IDFunc              func() string
	TypeFunc            func() string
	SendFunc            func(context.Context, *providers.Notification) error
	StatusFunc          func() *providers.ProviderStatus
	CloseFunc           func() error
	GetTestRecipientFunc func() (string, error)
	TestFunc            func(context.Context) error
}

func (m *MockProvider) Send(ctx context.Context, notification *providers.Notification) error {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, notification)
	}
	return nil
}

func (m *MockProvider) GetStatus() *providers.ProviderStatus {
	if m.StatusFunc != nil {
		return m.StatusFunc()
	}
	return &providers.ProviderStatus{Status: providers.StatusActive}
}

func (m *MockProvider) GetID() string {
	if m.IDFunc != nil {
		return m.IDFunc()
	}
	return ""
}

func (m *MockProvider) GetType() string {
	if m.TypeFunc != nil {
		return m.TypeFunc()
	}
	return ""
}

func (m *MockProvider) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockProvider) GetTestRecipient() (string, error) {
	if m.GetTestRecipientFunc != nil {
		return m.GetTestRecipientFunc()
	}
	return "test@example.com", nil
}

func (m *MockProvider) Test(ctx context.Context) error {
	if m.TestFunc != nil {
		return m.TestFunc(ctx)
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
