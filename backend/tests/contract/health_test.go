package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/developertyrone/notimulti/internal/api"
	"github.com/developertyrone/notimulti/internal/providers"
)

func TestHealthCheckReturns200(t *testing.T) {
	registry := providers.NewRegistry()
	router := api.SetupRouter(registry, nil)
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHealthCheckJSONStructure(t *testing.T) {
	registry := providers.NewRegistry()
	router := api.SetupRouter(registry, nil)
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Verify required fields exist
	requiredFields := []string{"status", "version", "timestamp"}
	for _, field := range requiredFields {
		if _, ok := response[field]; !ok {
			t.Errorf("Response missing required field: %s", field)
		}
	}

	// Verify status is "ok"
	if status, ok := response["status"].(string); !ok || status != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}

	// Verify version exists and is a string
	if _, ok := response["version"].(string); !ok {
		t.Errorf("Expected version to be a string, got %T", response["version"])
	}

	// Verify timestamp exists and is a string
	if _, ok := response["timestamp"].(string); !ok {
		t.Errorf("Expected timestamp to be a string, got %T", response["timestamp"])
	}
}
