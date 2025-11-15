package unit

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/developertyrone/notimulti/internal/api"
	"github.com/developertyrone/notimulti/internal/providers"
	"github.com/developertyrone/notimulti/internal/storage"
	"github.com/developertyrone/notimulti/tests/testhelpers"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

//go:embed dist/*
var testFrontendFS embed.FS

func TestHandleGetNotificationHistoryResponses(t *testing.T) {
	repo, db := setupTestRepository(t)
	defer closeSQLDB(t, db)

	now := time.Now().Add(-time.Minute)
	insertLog(t, db, "email-1", "email", storage.StatusSent, now, false)
	insertLog(t, db, "email-1", "email", storage.StatusFailed, now.Add(-time.Minute), false)
	insertLog(t, db, "email-1", "email", storage.StatusPending, now.Add(-2*time.Minute), true)

	handler := api.HandleGetNotificationHistory(repo)

	t.Run("invalid page size", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/history?page_size=200", nil)
		c.Request = req

		handler(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("success with pagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/history?page_size=1&provider_id=email-1&include_tests=false", nil)
		c.Request = req

		handler(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var body struct {
			Notifications []storage.NotificationLogEntry `json:"notifications"`
			Pagination    struct {
				PageSize   int  `json:"page_size"`
				HasMore    bool `json:"has_more"`
				NextCursor *int `json:"next_cursor"`
			} `json:"pagination"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(body.Notifications) != 1 {
			t.Fatalf("expected 1 notification, got %d", len(body.Notifications))
		}
		if !body.Pagination.HasMore || body.Pagination.NextCursor == nil {
			t.Fatalf("expected pagination to indicate more data: %+v", body.Pagination)
		}
	})
}

func TestHandleGetNotificationDetail(t *testing.T) {
	repo, db := setupTestRepository(t)
	defer closeSQLDB(t, db)

	id := insertLog(t, db, "email-2", "email", storage.StatusSent, time.Now(), false)

	handler := api.HandleGetNotificationDetail(repo)

	t.Run("invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/notifications/not-an-int", nil)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}

		handler(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/notifications/999", nil)
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		handler(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/notifications/", nil)
		c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", id)}}

		handler(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var entry storage.NotificationLogEntry
		if err := json.Unmarshal(w.Body.Bytes(), &entry); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if entry.ID != id {
			t.Fatalf("expected id %d, got %d", id, entry.ID)
		}
	})
}

func TestHandleReadinessCheck(t *testing.T) {
	registry := providers.NewRegistry()
	repo, db := setupTestRepository(t)

	handler := api.HandleReadinessCheck(registry, repo)

	t.Run("not ready when db closed", func(t *testing.T) {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)

		handler(c)

		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", w.Code)
		}

		var body map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if body["status"] != "not_ready" {
			t.Fatalf("expected status not_ready, got %v", body["status"])
		}
	})

	t.Run("ready when checks pass", func(t *testing.T) {
		// Recreate resources for this subtest
		repoReady, dbReady := setupTestRepository(t)
		defer func() {
			if err := dbReady.Close(); err != nil {
				t.Fatalf("failed to close db: %v", err)
			}
		}()

		mock := &testhelpers.MockProvider{
			IDFunc:   func() string { return "ok" },
			TypeFunc: func() string { return "email" },
			StatusFunc: func() *providers.ProviderStatus {
				return &providers.ProviderStatus{Status: providers.StatusActive}
			},
		}
		registryReady := providers.NewRegistry()
		if err := registryReady.Register(mock); err != nil {
			t.Fatalf("failed to register provider: %v", err)
		}

		handlerReady := api.HandleReadinessCheck(registryReady, repoReady)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)

		handlerReady(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var body map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if body["status"] != "ready" {
			t.Fatalf("expected ready status, got %v", body["status"])
		}
	})
}

func TestHandleTestProvider(t *testing.T) {
	t.Run("provider not found", func(t *testing.T) {
		registry := providers.NewRegistry()
		handler := api.HandleTestProvider(registry, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/providers/missing/test", nil)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}

		handler(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("rate limited when tested recently", func(t *testing.T) {
		registry := providers.NewRegistry()
		now := time.Now()
		mock := &testhelpers.MockProvider{
			IDFunc:   func() string { return "limited" },
			TypeFunc: func() string { return "email" },
			StatusFunc: func() *providers.ProviderStatus {
				return &providers.ProviderStatus{LastTestAt: &now}
			},
		}
		if err := registry.Register(mock); err != nil {
			t.Fatalf("failed to register provider: %v", err)
		}

		handler := api.HandleTestProvider(registry, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/providers/limited/test", nil)
		c.Params = gin.Params{{Key: "id", Value: "limited"}}

		handler(c)

		if w.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429, got %d", w.Code)
		}
	})

	t.Run("successful test execution", func(t *testing.T) {
		registry := providers.NewRegistry()
		var called bool
		mock := &testhelpers.MockProvider{
			IDFunc:   func() string { return "ok" },
			TypeFunc: func() string { return "telegram" },
			StatusFunc: func() *providers.ProviderStatus {
				return &providers.ProviderStatus{Status: providers.StatusActive}
			},
			TestFunc: func(ctx context.Context) error {
				called = true
				return nil
			},
		}
		if err := registry.Register(mock); err != nil {
			t.Fatalf("failed to register provider: %v", err)
		}

		handler := api.HandleTestProvider(registry, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/providers/ok/test", nil)
		c.Params = gin.Params{{Key: "id", Value: "ok"}}

		handler(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !called {
			t.Fatalf("expected provider.Test to be called")
		}
	})
}

func TestServeFrontend(t *testing.T) {
	router := gin.New()
	api.ServeFrontend(router, testFrontendFS)

	t.Run("serves static assets", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("serves index for spa routes", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Test Frontend") {
			t.Fatalf("expected index html content")
		}
	})

	t.Run("preserves api 404s", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}
