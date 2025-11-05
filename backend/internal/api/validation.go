package api

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateNotificationRequest validates a notification request
func ValidateNotificationRequest(req *NotificationRequest) []ValidationError {
	var errors []ValidationError

	// Validate ProviderID
	if req.ProviderID == "" {
		errors = append(errors, ValidationError{
			Field:   "provider_id",
			Message: "provider_id is required",
		})
	}

	// Validate Recipient
	if req.Recipient == "" {
		errors = append(errors, ValidationError{
			Field:   "recipient",
			Message: "recipient is required",
		})
	}

	// Validate Message
	if req.Message == "" {
		errors = append(errors, ValidationError{
			Field:   "message",
			Message: "message is required",
		})
	} else if len(req.Message) > 4096 {
		errors = append(errors, ValidationError{
			Field:   "message",
			Message: fmt.Sprintf("message must be ≤4096 characters (got %d)", len(req.Message)),
		})
	}

	// Validate Subject (optional but bounded)
	if len(req.Subject) > 200 {
		errors = append(errors, ValidationError{
			Field:   "subject",
			Message: fmt.Sprintf("subject must be ≤200 characters (got %d)", len(req.Subject)),
		})
	}

	// Validate Metadata
	if req.Metadata != nil {
		if len(req.Metadata) > 10 {
			errors = append(errors, ValidationError{
				Field:   "metadata",
				Message: fmt.Sprintf("metadata must have ≤10 key-value pairs (got %d)", len(req.Metadata)),
			})
		}

		for key, value := range req.Metadata {
			// Validate key length
			if len(key) > 50 {
				errors = append(errors, ValidationError{
					Field:   "metadata",
					Message: fmt.Sprintf("metadata key '%s' must be ≤50 characters (got %d)", key, len(key)),
				})
			}

			// Validate value length (convert to string for measurement)
			valueStr := fmt.Sprintf("%v", value)
			if len(valueStr) > 200 {
				errors = append(errors, ValidationError{
					Field:   "metadata",
					Message: fmt.Sprintf("metadata value for key '%s' must be ≤200 characters (got %d)", key, len(valueStr)),
				})
			}
		}
	}

	// Validate Priority (if provided)
	if req.Priority != "" {
		if !isValidPriority(req.Priority) {
			errors = append(errors, ValidationError{
				Field:   "priority",
				Message: fmt.Sprintf("priority must be one of: low, normal, high (got '%s')", req.Priority),
			})
		}
	}

	return errors
}

// ValidateEmailAddress validates email format
func ValidateEmailAddress(email string) bool {
	if email == "" {
		return false
	}

	// Simple email regex pattern
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// ValidateTelegramChatID validates Telegram chat ID format
func ValidateTelegramChatID(chatID string) bool {
	if chatID == "" {
		return false
	}

	// Telegram chat IDs can be negative numbers or positive numbers (as strings)
	// They can also include @ for usernames
	chatID = strings.TrimSpace(chatID)

	// Check if it's a number (positive or negative)
	if isNumeric(chatID) {
		return true
	}

	// Check if it's a username (starts with @ followed by alphanumeric and underscore)
	if strings.HasPrefix(chatID, "@") && len(chatID) > 1 {
		username := chatID[1:]
		pattern := `^[a-zA-Z0-9_]{5,32}$` // Telegram username rules
		matched, _ := regexp.MatchString(pattern, username)
		return matched
	}

	return false
}

// ValidateProviderForRecipient validates recipient format based on provider type
func ValidateProviderForRecipient(providerType string, recipient string) error {
	switch providerType {
	case "telegram":
		if !ValidateTelegramChatID(recipient) {
			return fmt.Errorf("invalid Telegram chat ID format: '%s' (must be numeric ID or @username)", recipient)
		}

	case "email":
		if !ValidateEmailAddress(recipient) {
			return fmt.Errorf("invalid email address format: '%s'", recipient)
		}

	default:
		// Unknown provider type - skip validation
		return nil
	}

	return nil
}

// Helper functions

// isValidPriority checks if priority is one of the allowed values
func isValidPriority(priority string) bool {
	priority = strings.ToLower(priority)
	return priority == "low" || priority == "normal" || priority == "high"
}

// isNumeric checks if a string is a valid number (including negative)
func isNumeric(s string) bool {
	if s == "" {
		return false
	}

	s = strings.TrimSpace(s)
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
	}

	if s == "" {
		return false
	}

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
