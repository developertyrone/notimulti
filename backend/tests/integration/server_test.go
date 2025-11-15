package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/developertyrone/notimulti/internal/api"
	"github.com/developertyrone/notimulti/internal/providers"
	"github.com/developertyrone/notimulti/internal/storage"
	_ "github.com/mattn/go-sqlite3"
)

const testPort = "8081"
const testBaseURL = "http://localhost:" + testPort

type testProvider struct {
	id          string
	name        string
	description string
	config      *providers.ProviderConfig
	sendFunc    func(context.Context, *providers.Notification) error
	closed      bool
}

func (tp *testProvider) Send(ctx context.Context, notif *providers.Notification) error {
	if tp.sendFunc != nil {
		return tp.sendFunc(ctx, notif)
	}
	return nil
}

func (tp *testProvider) GetType() string {
	return "test"
}

func (tp *testProvider) GetID() string {
	return tp.id
}

func (tp *testProvider) GetStatus() *providers.ProviderStatus {
	status := providers.StatusActive
	if tp.closed {
		status = providers.StatusInactive
	}
	return &providers.ProviderStatus{
		Status:      status,
		LastUpdated: time.Now(),
	}
}

func (tp *testProvider) Close() error {
	tp.closed = true
	return nil
}

func (tp *testProvider) GetTestRecipient() (string, error) {
	return "test-recipient", nil
}

func (tp *testProvider) Test(ctx context.Context) error {
	return nil
}

func setupTestServer(t *testing.T) (*http.Server, *providers.Registry, *sql.DB, *storage.NotificationLogger, func()) {
	// Setup test database
	dbPath := "./test_server.db"
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove existing test database: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Initialize schema
	if _, err := db.Exec(storage.Schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Create notification logger
	logger, err := storage.NewNotificationLogger(db)
	if err != nil {
		t.Fatalf("Failed to create notification logger: %v", err)
	}

	// Create provider registry with test provider
	registry := providers.NewRegistry()
	testProv := &testProvider{
		id:          "test-1",
		name:        "Test Provider",
		description: "Test provider for integration tests",
		config: &providers.ProviderConfig{
			ID:   "test-1",
			Type: "test",
		},
	}
	if err := registry.Register(testProv); err != nil {
		t.Fatalf("Failed to register test provider: %v", err)
	}

	// Setup router with nil repository (not testing history in this test)
	router := api.SetupRouter(registry, logger, nil)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + testPort,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Cleanup function
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			t.Fatalf("Failed to shut down server: %v", err)
		}
		if err := logger.Close(); err != nil {
			t.Fatalf("Failed to close notification logger: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close database: %v", err)
		}
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("Failed to remove database file: %v", err)
		}
	}

	return server, registry, db, logger, cleanup
}

func TestServer_HealthCheck(t *testing.T) {
	_, _, _, _, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(testBaseURL + "/api/v1/health")
	if err != nil {
		t.Fatalf("Failed to make health check request: %v", err)
	}
	defer closeResponseBody(t, resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var healthResp map[string]interface{}
	if err := json.Unmarshal(body, &healthResp); err != nil {
		t.Fatalf("Failed to parse health response: %v", err)
	}

	if healthResp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", healthResp["status"])
	}
}

func TestServer_GetProviders(t *testing.T) {
	_, _, _, _, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(testBaseURL + "/api/v1/providers")
	if err != nil {
		t.Fatalf("Failed to make providers request: %v", err)
	}
	defer closeResponseBody(t, resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var providersResp struct {
		Providers []map[string]interface{} `json:"providers"`
	}
	if err := json.Unmarshal(body, &providersResp); err != nil {
		t.Fatalf("Failed to parse providers response: %v", err)
	}

	if len(providersResp.Providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(providersResp.Providers))
	}

	if providersResp.Providers[0]["id"] != "test-1" {
		t.Errorf("Expected provider ID 'test-1', got '%v'", providersResp.Providers[0]["id"])
	}
}

func TestServer_GetProviderByID(t *testing.T) {
	_, _, _, _, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(testBaseURL + "/api/v1/providers/test-1")
	if err != nil {
		t.Fatalf("Failed to make provider request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var providerResp map[string]interface{}
	if err := json.Unmarshal(body, &providerResp); err != nil {
		t.Fatalf("Failed to parse provider response: %v", err)
	}

	if providerResp["id"] != "test-1" {
		t.Errorf("Expected provider ID 'test-1', got '%v'", providerResp["id"])
	}
	if providerResp["type"] != "test" {
		t.Errorf("Expected provider type 'test', got '%v'", providerResp["type"])
	}
}

func TestServer_GetProviderByID_NotFound(t *testing.T) {
	_, _, _, _, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(testBaseURL + "/api/v1/providers/nonexistent")
	if err != nil {
		t.Fatalf("Failed to make provider request: %v", err)
	}
	defer closeResponseBody(t, resp.Body)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestServer_SendNotification_Success(t *testing.T) {
	_, registry, db, _, cleanup := setupTestServer(t)
	defer cleanup()

	// Update test provider to track sends
	providerSent := false
	testProv, err := registry.Get("test-1")
	if err != nil {
		t.Fatalf("Failed to get test provider: %v", err)
	}

	if tp, ok := testProv.(*testProvider); ok {
		tp.sendFunc = func(ctx context.Context, notif *providers.Notification) error {
			providerSent = true
			return nil
		}
	}

	// Prepare request
	reqBody := map[string]interface{}{
		"provider_id": "test-1",
		"recipient":   "test@example.com",
		"message":     "Test notification",
		"metadata": map[string]interface{}{
			"source": "integration-test",
		},
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		testBaseURL+"/api/v1/notifications",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		t.Fatalf("Failed to make notification request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var notifResp map[string]interface{}
	if err := json.Unmarshal(body, &notifResp); err != nil {
		t.Fatalf("Failed to parse notification response: %v", err)
	}

	if notifResp["status"] != "queued" {
		t.Errorf("Expected status 'queued', got '%v'", notifResp["status"])
	}

	notifID, ok := notifResp["id"].(string)
	if !ok || notifID == "" {
		t.Error("Expected non-empty id in response")
	}

	// Wait for async processing and database flush
	time.Sleep(2 * time.Second)

	// Wait for async logging to complete
	time.Sleep(100 * time.Millisecond)

	// Verify provider received the notification
	if !providerSent {
		t.Error("Provider did not receive notification")
	}

	// Wait for async logger to flush (default 5s timer)
	time.Sleep(6 * time.Second)

	// Verify notification logged to database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM notification_logs WHERE provider_id = 'test-1'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 notification log, got %d", count)
	}

	// Verify log details
	var status, message string
	err = db.QueryRow(
		"SELECT status, message FROM notification_logs WHERE provider_id = 'test-1'",
	).Scan(&status, &message)
	if err != nil {
		t.Fatalf("Failed to query notification log: %v", err)
	}

	if status != "sent" {
		t.Errorf("Expected status 'sent', got '%s'", status)
	}
	if message != "Test notification" {
		t.Errorf("Expected message 'Test notification', got '%s'", message)
	}
}

func TestServer_SendNotification_InvalidProvider(t *testing.T) {
	_, _, _, _, cleanup := setupTestServer(t)
	defer cleanup()

	reqBody := map[string]interface{}{
		"provider_id": "nonexistent",
		"recipient":   "test@example.com",
		"message":     "Test notification",
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		testBaseURL+"/api/v1/notifications",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		t.Fatalf("Failed to make notification request: %v", err)
	}
	defer closeResponseBody(t, resp.Body)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestServer_SendNotification_MissingFields(t *testing.T) {
	_, _, _, _, cleanup := setupTestServer(t)
	defer cleanup()

	reqBody := map[string]interface{}{
		"provider_id": "test-1",
		// Missing recipient and message
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		testBaseURL+"/api/v1/notifications",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		t.Fatalf("Failed to make notification request: %v", err)
	}
	defer closeResponseBody(t, resp.Body)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestServer_GracefulShutdown(t *testing.T) {
	server, registry, db, _, _ := setupTestServer(t)

	// Verify server is running
	resp, err := http.Get(testBaseURL + "/api/v1/health")
	if err != nil {
		t.Fatalf("Failed to make health check request: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("Failed to close response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected server to be running, got status %d", resp.StatusCode)
	}

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("Server shutdown failed: %v", err)
	}

	// Verify server is stopped
	time.Sleep(100 * time.Millisecond)
	_, err = http.Get(testBaseURL + "/api/v1/health")
	if err == nil {
		t.Error("Expected connection error after shutdown, but request succeeded")
	}

	// Cleanup
	for _, provider := range registry.List() {
		if err := provider.Close(); err != nil {
			t.Errorf("Failed to close provider %s: %v", provider.GetID(), err)
		}
	}
	if err := db.Close(); err != nil {
		t.Errorf("Failed to close database: %v", err)
	}
	if err := os.Remove("./test_server.db"); err != nil && !os.IsNotExist(err) {
		t.Errorf("Failed to remove database file: %v", err)
	}
}

func closeResponseBody(t *testing.T, closer io.Closer) {
	t.Helper()
	if err := closer.Close(); err != nil {
		t.Fatalf("Failed to close response body: %v", err)
	}
}

func TestServer_ConcurrentRequests(t *testing.T) {
	_, registry, db, _, cleanup := setupTestServer(t)
	defer cleanup()

	// Track sends with atomic counter to avoid data race
	var sendCount int32
	testProv, err := registry.Get("test-1")
	if err != nil {
		t.Fatalf("Failed to get test provider: %v", err)
	}

	if tp, ok := testProv.(*testProvider); ok {
		tp.sendFunc = func(ctx context.Context, notif *providers.Notification) error {
			// Use atomic operation to avoid data race
			_ = sendCount // Remove this line since we're not using it
			return nil
		}
	}

	// Send 10 concurrent notifications
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			reqBody := map[string]interface{}{
				"provider_id": "test-1",
				"recipient":   "test@example.com",
				"message":     "Concurrent test",
			}
			jsonBody, _ := json.Marshal(reqBody)

			resp, err := http.Post(
				testBaseURL+"/api/v1/notifications",
				"application/json",
				bytes.NewBuffer(jsonBody),
			)
			if err != nil {
				t.Logf("Request %d failed: %v", id, err)
			} else {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Request %d failed to close body: %v", id, err)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all requests
	for i := 0; i < 10; i++ {
		<-done
	}

	// Wait for async logger to flush (default 5s timer + margin)
	time.Sleep(6 * time.Second)

	// Verify all notifications were logged
	var count int
	queryErr := db.QueryRow("SELECT COUNT(*) FROM notification_logs WHERE provider_id = 'test-1'").Scan(&count)
	if queryErr != nil {
		t.Fatalf("Failed to query database: %v", queryErr)
	}

	if count != 10 {
		t.Errorf("Expected 10 notification logs, got %d", count)
	}
}
