package testhelpers

import (
	"context"
	"testing"

	"github.com/developertyrone/notimulti/internal/providers"
	_ "github.com/mattn/go-sqlite3"
)

func TestSetupTestDB(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	t.Cleanup(cleanup)

	if _, err := db.Exec("SELECT 1"); err != nil {
		t.Fatalf("expected SELECT 1 to succeed, got %v", err)
	}
}

func TestMockProviderDefaults(t *testing.T) {
	mock := &MockProvider{}

	if err := mock.Send(context.Background(), nil); err != nil {
		t.Fatalf("expected Send to default to nil error, got %v", err)
	}

	status := mock.GetStatus()
	if status == nil || status.Status != providers.StatusActive {
		t.Fatalf("expected default active status, got %+v", status)
	}

	if mock.GetID() != "" {
		t.Fatalf("expected empty default ID")
	}
	if mock.GetType() != "" {
		t.Fatalf("expected empty default type")
	}

	if err := mock.Close(); err != nil {
		t.Fatalf("expected Close to succeed, got %v", err)
	}

	recipient, err := mock.GetTestRecipient()
	if err != nil || recipient == "" {
		t.Fatalf("expected default test recipient, got %s (err=%v)", recipient, err)
	}

	if err := mock.Test(context.Background()); err != nil {
		t.Fatalf("expected Test to succeed by default, got %v", err)
	}
}
