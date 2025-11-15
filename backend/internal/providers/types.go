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

// ProviderConfig represents the configuration for a provider
type ProviderConfig struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"` // "telegram" or "email"
	Telegram *TelegramConfig `json:"telegram,omitempty"`
	Email    *EmailConfig    `json:"email,omitempty"`
}

// ProviderStatus represents the current status of a provider
type ProviderStatus struct {
	Status         string     `json:"status"` // "active", "inactive", "error", "disabled"
	LastUpdated    time.Time  `json:"last_updated"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	ConfigChecksum string     `json:"config_checksum,omitempty"`
	LastTestAt     *time.Time `json:"last_test_at,omitempty"`     // T049: When provider was last tested
	LastTestStatus string     `json:"last_test_status,omitempty"` // T049: "success" or "failed"
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

// TelegramConfig contains Telegram-specific configuration
type TelegramConfig struct {
	BotToken       string `json:"bot_token"`
	DefaultChatID  string `json:"default_chat_id"`
	ParseMode      string `json:"parse_mode,omitempty"` // HTML or Markdown
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
	APIEndpoint    string `json:"api_endpoint,omitempty"`
}

// EmailConfig contains Email-specific configuration
type EmailConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	From           string `json:"from"`
	UseTLS         bool   `json:"use_tls,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
	TestRecipient  string `json:"test_recipient,omitempty"` // T050: Email address for test notifications
}
