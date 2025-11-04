package providers

import "time"

// Notification represents a notification to be sent through a provider
type Notification struct {
	ID         string                 `json:"id"`
	ProviderID string                 `json:"provider_id"`
	Recipient  string                 `json:"recipient"`
	Message    string                 `json:"message"`
	Subject    string                 `json:"subject,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Priority   string                 `json:"priority,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// ProviderStatus represents the current status of a provider
type ProviderStatus struct {
	Status         string    `json:"status"` // "active", "inactive", "error", "disabled"
	LastUpdated    time.Time `json:"last_updated"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	ConfigChecksum string    `json:"config_checksum,omitempty"`
}

// Priority constants for notifications
const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
)

// Status constants for providers
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusError    = "error"
	StatusDisabled = "disabled"
)
