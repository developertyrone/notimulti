package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// ProviderConfig represents a provider configuration loaded from file
type ProviderConfig struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Enabled  bool                   `json:"enabled"`
	Config   map[string]interface{} `json:"config"`
	Checksum string                 `json:"-"` // SHA256 hash of config content
}

// ComputeChecksum calculates SHA256 hash of the configuration
func (pc *ProviderConfig) ComputeChecksum() error {
	// Serialize config to JSON for consistent hashing
	data, err := json.Marshal(map[string]interface{}{
		"id":      pc.ID,
		"type":    pc.Type,
		"enabled": pc.Enabled,
		"config":  pc.Config,
	})
	if err != nil {
		return err
	}

	hash := sha256.Sum256(data)
	pc.Checksum = hex.EncodeToString(hash[:])
	return nil
}

// TelegramConfig represents Telegram-specific configuration
type TelegramConfig struct {
	BotToken       string `json:"bot_token"`
	DefaultChatID  string `json:"default_chat_id"`
	ParseMode      string `json:"parse_mode"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// EmailConfig represents Email/SMTP-specific configuration
type EmailConfig struct {
	SMTPHost       string `json:"smtp_host"`
	SMTPPort       int    `json:"smtp_port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	FromAddress    string `json:"from_address"`
	FromName       string `json:"from_name"`
	UseTLS         bool   `json:"use_tls"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}
