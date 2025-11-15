package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/developertyrone/notimulti/internal/providers"
)

func TestFailedProviderBehavior(t *testing.T) {
	initErr := errors.New("boom")
	fp := providers.NewFailedProvider("broken", "email", initErr)

	if fp.GetID() != "broken" {
		t.Fatalf("expected id broken, got %s", fp.GetID())
	}
	if fp.GetType() != "email" {
		t.Fatalf("expected type email, got %s", fp.GetType())
	}

	if status := fp.GetStatus(); status == nil || status.Status != providers.StatusError {
		t.Fatalf("expected error status, got %+v", status)
	}

	if err := fp.Send(context.Background(), &providers.Notification{}); err == nil {
		t.Fatalf("expected send error")
	}

	if _, err := fp.GetTestRecipient(); err == nil {
		t.Fatalf("expected test recipient error")
	}

	if err := fp.Test(context.Background()); err == nil {
		t.Fatalf("expected test error")
	}

	if err := fp.Close(); err != nil {
		t.Fatalf("expected close to succeed, got %v", err)
	}
}
