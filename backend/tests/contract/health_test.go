package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/developertyrone/notimulti/internal/api"
	"github.com/developertyrone/notimulti/internal/providers"
)

func TestHealthEndpoint(t *testing.T) {
	registry := providers.NewRegistry()
	router := api.SetupRouter(registry, nil, nil)
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
	router := api.SetupRouter(registry, nil, nil)
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

// T065: Contract tests for GET /ready endpoint
func TestReadyEndpoint_AllHealthy(t *testing.T) {
	t.Skip("TODO: T065 - Implement once /ready handler with database check is implemented")
	
	// This test will verify:
	// 1. Setup registry with at least one provider
	// 2. Setup repository with healthy database connection
	// 3. Send GET /api/v1/ready request
	// 4. Verify response status is 200
	// 5. Verify response has status="ready"
	// 6. Verify checks.database="ok"
	// 7. Verify checks.providers="ok"
}

func TestReadyEndpoint_NoProviders(t *testing.T) {
	t.Skip("TODO: T065 - Implement once /ready handler is implemented")
	
	// This test will verify:
	// 1. Setup empty registry (no providers)
	// 2. Setup repository with healthy database
	// 3. Send GET /api/v1/ready request
	// 4. Verify response status is 503 (Service Unavailable)
	// 5. Verify response has status="not_ready"
	// 6. Verify checks.providers indicates no providers loaded
}

func TestReadyEndpoint_DatabaseUnhealthy(t *testing.T) {
	t.Skip("TODO: T065 - Implement once /ready handler with database check is implemented")
	
	// This test will verify:
	// 1. Setup registry with providers
	// 2. Setup repository with broken/nil database connection
	// 3. Send GET /api/v1/ready request
	// 4. Verify response status is 503
	// 5. Verify response has status="not_ready"
	// 6. Verify checks.database indicates failure
	// 
	// This ensures Kubernetes readiness probe can detect database issues
}

func TestReadyEndpoint_JSONStructure(t *testing.T) {
	t.Skip("TODO: T065 - Implement once /ready handler is implemented")
	
	// This test verifies the response structure matches the OpenAPI spec:
	// {
	//   "status": "ready" | "not_ready",
	//   "checks": {
	//     "database": "ok" | "error: ...",
	//     "providers": "ok" | "no providers loaded"
	//   }
	// }
}
