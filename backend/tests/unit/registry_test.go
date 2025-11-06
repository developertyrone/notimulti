package unit

import (
	"context"
	"sync"
	"testing"

	"github.com/developertyrone/notimulti/internal/providers"
	"github.com/developertyrone/notimulti/tests/testhelpers"
)

func TestRegistryRegister(t *testing.T) {
	registry := providers.NewRegistry()

	// Create a mock provider
	mockProvider := &testhelpers.MockProvider{
		IDFunc:    func() string { return "test-provider" },
		TypeFunc:  func() string { return "mock" },
		CloseFunc: func() error { return nil },
	}

	// Test successful registration
	err := registry.Register(mockProvider)
	if err != nil {
		t.Errorf("Register() failed: %v", err)
	}

	// Verify provider was registered
	if registry.Count() != 1 {
		t.Errorf("Expected 1 provider, got %d", registry.Count())
	}
}

func TestRegistryRegisterNil(t *testing.T) {
	registry := providers.NewRegistry()

	err := registry.Register(nil)
	if err == nil {
		t.Error("Expected error when registering nil provider")
	}
}

func TestRegistryRegisterEmptyID(t *testing.T) {
	registry := providers.NewRegistry()

	mockProvider := &testhelpers.MockProvider{
		IDFunc: func() string { return "" },
	}

	err := registry.Register(mockProvider)
	if err == nil {
		t.Error("Expected error when registering provider with empty ID")
	}
}

func TestRegistryGet(t *testing.T) {
	registry := providers.NewRegistry()

	mockProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-provider" },
		TypeFunc: func() string { return "mock" },
	}

	registry.Register(mockProvider)

	// Test successful retrieval
	provider, err := registry.Get("test-provider")
	if err != nil {
		t.Errorf("Get() failed: %v", err)
	}
	if provider == nil {
		t.Error("Get() returned nil provider")
	}
	if provider.GetID() != "test-provider" {
		t.Errorf("Get() returned wrong provider: got %s, want test-provider", provider.GetID())
	}

	// Test retrieval of non-existent provider
	_, err = registry.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent provider")
	}
}

func TestRegistryList(t *testing.T) {
	registry := providers.NewRegistry()

	// Register multiple providers
	for i := 1; i <= 3; i++ {
		id := string(rune('a' + i - 1))
		mockProvider := &testhelpers.MockProvider{
			IDFunc:   func() string { return "provider-" + id },
			TypeFunc: func() string { return "mock" },
		}
		registry.Register(mockProvider)
	}

	providers := registry.List()
	if len(providers) != 3 {
		t.Errorf("List() returned %d providers, expected 3", len(providers))
	}
}

func TestRegistryRemove(t *testing.T) {
	registry := providers.NewRegistry()

	closeCalled := false
	mockProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-provider" },
		TypeFunc: func() string { return "mock" },
		CloseFunc: func() error {
			closeCalled = true
			return nil
		},
	}

	registry.Register(mockProvider)

	// Test successful removal
	err := registry.Remove("test-provider")
	if err != nil {
		t.Errorf("Remove() failed: %v", err)
	}

	if !closeCalled {
		t.Error("Close() was not called on removed provider")
	}

	if registry.Count() != 0 {
		t.Errorf("Expected 0 providers after removal, got %d", registry.Count())
	}

	// Test removal of non-existent provider
	err = registry.Remove("non-existent")
	if err == nil {
		t.Error("Expected error when removing non-existent provider")
	}
}

func TestRegistryClear(t *testing.T) {
	registry := providers.NewRegistry()

	closeCount := 0
	for i := 1; i <= 3; i++ {
		id := string(rune('a' + i - 1))
		mockProvider := &testhelpers.MockProvider{
			IDFunc:   func() string { return "provider-" + id },
			TypeFunc: func() string { return "mock" },
			CloseFunc: func() error {
				closeCount++
				return nil
			},
		}
		registry.Register(mockProvider)
	}

	err := registry.Clear()
	if err != nil {
		t.Errorf("Clear() failed: %v", err)
	}

	if closeCount != 3 {
		t.Errorf("Expected Close() to be called 3 times, got %d", closeCount)
	}

	if registry.Count() != 0 {
		t.Errorf("Expected 0 providers after clear, got %d", registry.Count())
	}
}

func TestRegistryReplaceProvider(t *testing.T) {
	registry := providers.NewRegistry()

	// Register initial provider
	closeCount := 0
	mockProvider1 := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-provider" },
		TypeFunc: func() string { return "mock-v1" },
		CloseFunc: func() error {
			closeCount++
			return nil
		},
	}
	registry.Register(mockProvider1)

	// Replace with new provider with same ID
	mockProvider2 := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-provider" },
		TypeFunc: func() string { return "mock-v2" },
		CloseFunc: func() error {
			closeCount++
			return nil
		},
	}
	registry.Register(mockProvider2)

	// Verify old provider was closed
	if closeCount != 1 {
		t.Errorf("Expected old provider to be closed once, got %d", closeCount)
	}

	// Verify only one provider exists
	if registry.Count() != 1 {
		t.Errorf("Expected 1 provider after replacement, got %d", registry.Count())
	}

	// Verify new provider type
	provider, _ := registry.Get("test-provider")
	if provider.GetType() != "mock-v2" {
		t.Errorf("Expected provider type mock-v2, got %s", provider.GetType())
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	registry := providers.NewRegistry()
	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			mockProvider := &testhelpers.MockProvider{
				IDFunc:    func() string { return string(rune('a' + index)) },
				TypeFunc:  func() string { return "mock" },
				CloseFunc: func() error { return nil },
			}
			registry.Register(mockProvider)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			registry.Get(string(rune('a' + index)))
		}(i)
	}

	// Concurrent lists
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.List()
		}()
	}

	wg.Wait()

	// Verify final state
	if registry.Count() != 10 {
		t.Errorf("Expected 10 providers after concurrent operations, got %d", registry.Count())
	}
}

func TestRegistryConcurrentRemoval(t *testing.T) {
	registry := providers.NewRegistry()

	// Register providers
	for i := 0; i < 10; i++ {
		mockProvider := &testhelpers.MockProvider{
			IDFunc:    func() string { return string(rune('a' + i)) },
			TypeFunc:  func() string { return "mock" },
			CloseFunc: func() error { return nil },
		}
		registry.Register(mockProvider)
	}

	// Concurrent removals
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			registry.Remove(string(rune('a' + index)))
		}(i)
	}

	wg.Wait()

	// Verify all providers removed
	if registry.Count() != 0 {
		t.Errorf("Expected 0 providers after concurrent removal, got %d", registry.Count())
	}
}

func TestMockProviderSend(t *testing.T) {
	mockProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test" },
		TypeFunc: func() string { return "mock" },
		SendFunc: func(ctx context.Context, notification *providers.Notification) error {
			return nil
		},
	}

	notification := &providers.Notification{
		ID:        "notif-1",
		Recipient: "test@example.com",
		Message:   "Test message",
	}

	err := mockProvider.Send(context.Background(), notification)
	if err != nil {
		t.Errorf("Send() failed: %v", err)
	}
}

func TestMockProviderStatus(t *testing.T) {
	mockProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test" },
		TypeFunc: func() string { return "mock" },
		StatusFunc: func() *providers.ProviderStatus {
			return &providers.ProviderStatus{
				Status: providers.StatusActive,
			}
		},
	}

	status := mockProvider.GetStatus()
	if status == nil {
		t.Error("GetStatus() returned nil")
	}
	if status.Status != providers.StatusActive {
		t.Errorf("Expected status active, got %s", status.Status)
	}
}

// TestRegistryReplaceMethod tests the Replace() method for atomic swap
func TestRegistryReplaceMethod(t *testing.T) {
	registry := providers.NewRegistry()

	// Register initial provider
	closeCount := 0
	oldProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-provider" },
		TypeFunc: func() string { return "mock-v1" },
		StatusFunc: func() *providers.ProviderStatus {
			return &providers.ProviderStatus{
				ConfigChecksum: "old-checksum",
			}
		},
		CloseFunc: func() error {
			closeCount++
			return nil
		},
	}
	if err := registry.Register(oldProvider); err != nil {
		t.Fatalf("Failed to register initial provider: %v", err)
	}

	// Create new provider
	newProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-provider" },
		TypeFunc: func() string { return "mock-v2" },
		StatusFunc: func() *providers.ProviderStatus {
			return &providers.ProviderStatus{
				ConfigChecksum: "new-checksum",
			}
		},
		CloseFunc: func() error {
			closeCount++
			return nil
		},
	}

	// Replace using Replace() method
	err := registry.Replace("test-provider", newProvider)
	if err != nil {
		t.Errorf("Replace() failed: %v", err)
	}

	// Verify old provider was closed
	if closeCount != 1 {
		t.Errorf("Expected old provider to be closed once, got %d", closeCount)
	}

	// Verify new provider is in registry
	provider, err := registry.Get("test-provider")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}
	if provider.GetType() != "mock-v2" {
		t.Errorf("Expected provider type mock-v2, got %s", provider.GetType())
	}

	// Verify only one provider exists
	if registry.Count() != 1 {
		t.Errorf("Expected 1 provider after replacement, got %d", registry.Count())
	}
}

// TestRegistryReplaceNilProvider tests that Replace() rejects nil providers
func TestRegistryReplaceNilProvider(t *testing.T) {
	registry := providers.NewRegistry()

	// Register initial provider
	oldProvider := &testhelpers.MockProvider{
		IDFunc:    func() string { return "test-provider" },
		TypeFunc:  func() string { return "mock" },
		CloseFunc: func() error { return nil },
	}
	registry.Register(oldProvider)

	// Attempt to replace with nil
	err := registry.Replace("test-provider", nil)
	if err == nil {
		t.Error("Expected error when replacing with nil provider")
	}

	// Verify original provider still exists
	provider, err := registry.Get("test-provider")
	if err != nil {
		t.Error("Original provider should still exist")
	}
	if provider.GetType() != "mock" {
		t.Error("Original provider should not have changed")
	}
}

// TestRegistryReplaceIDMismatch tests that Replace() rejects ID mismatch
func TestRegistryReplaceIDMismatch(t *testing.T) {
	registry := providers.NewRegistry()

	// Register initial provider
	oldProvider := &testhelpers.MockProvider{
		IDFunc:    func() string { return "test-provider" },
		TypeFunc:  func() string { return "mock" },
		CloseFunc: func() error { return nil },
	}
	registry.Register(oldProvider)

	// Attempt to replace with different ID
	newProvider := &testhelpers.MockProvider{
		IDFunc:    func() string { return "different-id" },
		TypeFunc:  func() string { return "mock-v2" },
		CloseFunc: func() error { return nil },
	}

	err := registry.Replace("test-provider", newProvider)
	if err == nil {
		t.Error("Expected error when provider ID doesn't match")
	}

	// Verify original provider still exists
	provider, _ := registry.Get("test-provider")
	if provider.GetType() != "mock" {
		t.Error("Original provider should not have changed")
	}
}

// TestRegistryReplaceConcurrentSend tests atomic swap during concurrent Send() calls
func TestRegistryReplaceConcurrentSend(t *testing.T) {
	registry := providers.NewRegistry()

	sendCount := 0
	var mu sync.Mutex

	// Register initial provider
	oldProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-provider" },
		TypeFunc: func() string { return "mock-v1" },
		SendFunc: func(ctx context.Context, notification *providers.Notification) error {
			mu.Lock()
			sendCount++
			mu.Unlock()
			return nil
		},
		CloseFunc: func() error { return nil },
	}
	registry.Register(oldProvider)

	// Start concurrent Send() operations
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			provider, err := registry.Get("test-provider")
			if err == nil && provider != nil {
				notification := &providers.Notification{
					ID:        string(rune('a' + index)),
					Recipient: "test",
					Message:   "test",
				}
				provider.Send(context.Background(), notification)
			}
		}(i)
	}

	// Perform replacement while Sends are happening
	newProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-provider" },
		TypeFunc: func() string { return "mock-v2" },
		SendFunc: func(ctx context.Context, notification *providers.Notification) error {
			mu.Lock()
			sendCount++
			mu.Unlock()
			return nil
		},
		CloseFunc: func() error { return nil },
	}

	err := registry.Replace("test-provider", newProvider)
	if err != nil {
		t.Errorf("Replace() failed during concurrent operations: %v", err)
	}

	wg.Wait()

	// Verify some sends completed (no deadlock occurred)
	mu.Lock()
	finalSendCount := sendCount
	mu.Unlock()

	if finalSendCount == 0 {
		t.Error("No sends completed - possible deadlock or race condition")
	}

	t.Logf("Completed %d sends during replacement", finalSendCount)
}

// TestRegistryReplaceCloseError tests that Replace() continues even if Close() fails
func TestRegistryReplaceCloseError(t *testing.T) {
	registry := providers.NewRegistry()

	// Register provider that fails on Close()
	oldProvider := &testhelpers.MockProvider{
		IDFunc:   func() string { return "test-provider" },
		TypeFunc: func() string { return "mock-v1" },
		CloseFunc: func() error {
			return context.Canceled // Simulate close failure
		},
	}
	registry.Register(oldProvider)

	// New provider
	newProvider := &testhelpers.MockProvider{
		IDFunc:    func() string { return "test-provider" },
		TypeFunc:  func() string { return "mock-v2" },
		CloseFunc: func() error { return nil },
	}

	// Replace should succeed despite old provider Close() error
	err := registry.Replace("test-provider", newProvider)
	if err != nil {
		t.Errorf("Replace() should succeed even if Close() fails: %v", err)
	}

	// Verify new provider is in place
	provider, err := registry.Get("test-provider")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}
	if provider.GetType() != "mock-v2" {
		t.Errorf("Expected new provider (mock-v2), got %s", provider.GetType())
	}
}
