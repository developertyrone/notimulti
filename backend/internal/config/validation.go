package config

import (
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

var idPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// ValidateConfig validates a provider configuration
func ValidateConfig(config *ProviderConfig) error {
	// Validate required fields
	if config.ID == "" {
		return &ValidationError{Field: "id", Message: "ID is required"}
	}

	if config.Type == "" {
		return &ValidationError{Field: "type", Message: "Type is required"}
	}

	// Validate ID pattern
	if !idPattern.MatchString(config.ID) {
		return &ValidationError{
			Field:   "id",
			Message: "ID must contain only lowercase letters, numbers, and hyphens",
		}
	}

	// Validate type-specific configuration
	switch config.Type {
	case "telegram":
		return validateTelegramConfig(config.Config)
	case "email":
		return validateEmailConfig(config.Config)
	default:
		return &ValidationError{
			Field:   "type",
			Message: fmt.Sprintf("unsupported provider type: %s", config.Type),
		}
	}
}

// ValidateConfigs validates multiple configurations and checks for duplicates
func ValidateConfigs(configs []*ProviderConfig) error {
	ids := make(map[string]bool)

	for _, config := range configs {
		// Check for duplicate IDs
		if ids[config.ID] {
			return &ValidationError{
				Field:   "id",
				Message: fmt.Sprintf("duplicate provider ID: %s", config.ID),
			}
		}
		ids[config.ID] = true

		// Validate individual config
		if err := ValidateConfig(config); err != nil {
			return fmt.Errorf("provider %s: %w", config.ID, err)
		}
	}

	return nil
}

func validateTelegramConfig(config map[string]interface{}) error {
	// Validate bot_token
	token, ok := config["bot_token"].(string)
	if !ok || token == "" {
		return &ValidationError{Field: "bot_token", Message: "bot_token is required"}
	}

	// Validate default_chat_id
	chatID, ok := config["default_chat_id"]
	if !ok {
		return &ValidationError{Field: "default_chat_id", Message: "default_chat_id is required"}
	}

	// Check if chat_id is a valid number or string
	switch v := chatID.(type) {
	case string:
		if v == "" {
			return &ValidationError{Field: "default_chat_id", Message: "default_chat_id cannot be empty"}
		}
		if _, err := strconv.ParseInt(v, 10, 64); err != nil {
			return &ValidationError{Field: "default_chat_id", Message: "default_chat_id must be a valid integer"}
		}
	case float64:
		// JSON numbers are float64 by default
		if v == 0 {
			return &ValidationError{Field: "default_chat_id", Message: "default_chat_id cannot be zero"}
		}
	default:
		return &ValidationError{Field: "default_chat_id", Message: "default_chat_id must be a string or number"}
	}

	return nil
}

func validateEmailConfig(config map[string]interface{}) error {
	// Host can be provided as host or smtp_host
	host := ""
	if h, ok := config["host"].(string); ok && h != "" {
		host = h
	} else if h, ok := config["smtp_host"].(string); ok && h != "" {
		host = h
	}
	if host == "" {
		return &ValidationError{Field: "host", Message: "host/smtp_host is required"}
	}

	// Port can be provided as port or smtp_port
	portVal, hasPort := config["port"]
	if !hasPort {
		portVal, hasPort = config["smtp_port"]
	}
	if !hasPort {
		return &ValidationError{Field: "port", Message: "port/smtp_port is required"}
	}

	// Check if port is a valid number
	switch v := portVal.(type) {
	case float64:
		if v <= 0 || v > 65535 {
			return &ValidationError{Field: "port", Message: "port must be between 1 and 65535"}
		}
	case string:
		p, err := strconv.Atoi(v)
		if err != nil || p <= 0 || p > 65535 {
			return &ValidationError{Field: "port", Message: "port must be between 1 and 65535"}
		}
	default:
		return &ValidationError{Field: "port", Message: "port must be a number"}
	}

	// Validate username
	username, ok := config["username"].(string)
	if !ok || username == "" {
		return &ValidationError{Field: "username", Message: "username is required"}
	}

	// Validate password
	password, ok := config["password"].(string)
	if !ok || password == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}

	// from/from_address
	from := ""
	if f, ok := config["from"].(string); ok && f != "" {
		from = f
	} else if f, ok := config["from_address"].(string); ok && f != "" {
		from = f
	}
	if from == "" {
		return &ValidationError{Field: "from", Message: "from/from_address is required"}
	}

	// Validate email format
	if _, err := mail.ParseAddress(from); err != nil {
		return &ValidationError{Field: "from", Message: "from must be a valid email address"}
	}

	return nil
}
