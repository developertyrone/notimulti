package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/developertyrone/notimulti/internal/api"
	"github.com/developertyrone/notimulti/internal/providers"
	"github.com/developertyrone/notimulti/tests/testhelpers"
	"github.com/gin-gonic/gin"
)

func TestGetProvidersContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	registry := providers.NewRegistry()

	// Register test providers
	mockProvider1 := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-telegram" },
		TypeFunc: func() string { return "telegram" },
		StatusFunc: func() *providers.ProviderStatus {
			return &providers.ProviderStatus{
				Status: providers.StatusActive,
			}
		},
		CloseFunc: func() error { return nil },
	}

	mockProvider2 := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-email" },
		TypeFunc: func() string { return "email" },
		StatusFunc: func() *providers.ProviderStatus {
			return &providers.ProviderStatus{
				Status:       providers.StatusError,
				ErrorMessage: "connection failed",
			}
		},
		CloseFunc: func() error { return nil },
	}

	registry.Register(mockProvider1)
	registry.Register(mockProvider2)

	// Create router
	router := gin.New()
	router.GET("/api/v1/providers", api.HandleGetProviders(registry))

	// Make request
	req, _ := http.NewRequest("GET", "/api/v1/providers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check response structure
	if _, ok := response["providers"]; !ok {
		t.Error("Response should contain 'providers' key")
	}

	if _, ok := response["count"]; !ok {
		t.Error("Response should contain 'count' key")
	}

	// Check providers array
	providersList, ok := response["providers"].([]interface{})
	if !ok {
		t.Fatal("'providers' should be an array")
	}

	if len(providersList) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providersList))
	}

	// Check first provider structure
	if len(providersList) > 0 {
		provider := providersList[0].(map[string]interface{})

		requiredFields := []string{"id", "type", "status", "last_updated"}
		for _, field := range requiredFields {
			if _, ok := provider[field]; !ok {
				t.Errorf("Provider should have '%s' field", field)
			}
		}

		// Verify status is one of valid values
		status, _ := provider["status"].(string)
		validStatuses := []string{providers.StatusActive, providers.StatusError, providers.StatusDisabled, "initializing"}
		validStatus := false
		for _, vs := range validStatuses {
			if status == vs {
				validStatus = true
				break
			}
		}
		if !validStatus {
			t.Errorf("Status '%s' is not valid", status)
		}
	}

	// Check that error provider includes error_message
	foundError := false
	for _, p := range providersList {
		provider := p.(map[string]interface{})
		if status, _ := provider["status"].(string); status == providers.StatusError {
			if _, ok := provider["error_message"]; ok {
				foundError = true
				break
			}
		}
	}
	if !foundError {
		t.Error("Provider with error status should include error_message")
	}
}

func TestGetProviderByIDContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	registry := providers.NewRegistry()

	mockProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-provider" },
		TypeFunc: func() string { return "telegram" },
		StatusFunc: func() *providers.ProviderStatus {
			return &providers.ProviderStatus{
				Status:         providers.StatusActive,
				ConfigChecksum: "abc123",
			}
		},
		CloseFunc: func() error { return nil },
	}

	registry.Register(mockProvider)

	// Create router
	router := gin.New()
	router.GET("/api/v1/providers/:id", api.HandleGetProvider(registry))

	// Test valid provider
	t.Run("Valid provider ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/providers/test-provider", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Check required fields
		requiredFields := []string{"id", "type", "status", "last_updated", "config_checksum"}
		for _, field := range requiredFields {
			if _, ok := response[field]; !ok {
				t.Errorf("Response should have '%s' field", field)
			}
		}

		// Verify values
		if response["id"] != "test-provider" {
			t.Errorf("Expected id 'test-provider', got %v", response["id"])
		}

		if response["type"] != "telegram" {
			t.Errorf("Expected type 'telegram', got %v", response["type"])
		}

		if response["config_checksum"] != "abc123" {
			t.Errorf("Expected config_checksum 'abc123', got %v", response["config_checksum"])
		}
	})

	// Test invalid provider ID
	t.Run("Invalid provider ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/providers/non-existent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := response["error"]; !ok {
			t.Error("Error response should contain 'error' field")
		}
	})
}

func TestGetProviderWithErrorStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	registry := providers.NewRegistry()

	mockProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "error-provider" },
		TypeFunc: func() string { return "email" },
		StatusFunc: func() *providers.ProviderStatus {
			return &providers.ProviderStatus{
				Status:       providers.StatusError,
				ErrorMessage: "SMTP connection failed",
			}
		},
		CloseFunc: func() error { return nil },
	}

	registry.Register(mockProvider)

	// Create router
	router := gin.New()
	router.GET("/api/v1/providers/:id", api.HandleGetProvider(registry))

	// Make request
	req, _ := http.NewRequest("GET", "/api/v1/providers/error-provider", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 even for error provider, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check error message is included
	if _, ok := response["error_message"]; !ok {
		t.Error("Provider with error status should include error_message field")
	}

	if response["error_message"] != "SMTP connection failed" {
		t.Errorf("Expected error_message 'SMTP connection failed', got %v", response["error_message"])
	}
}
