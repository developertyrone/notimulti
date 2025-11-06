package api

import "strings"

// MaskConfig masks sensitive fields in provider configuration based on provider type
func MaskConfig(providerType string, config map[string]interface{}) map[string]interface{} {
	if config == nil {
		return config
	}

	// Create a copy to avoid modifying the original
	masked := make(map[string]interface{})
	for k, v := range config {
		masked[k] = v
	}

	switch strings.ToLower(providerType) {
	case "telegram":
		// Mask bot_token - show only last 4 characters
		if token, ok := masked["bot_token"].(string); ok && len(token) > 4 {
			masked["bot_token"] = "***" + token[len(token)-4:]
		}
	case "email":
		// Mask password - show ****masked****
		if _, ok := masked["password"].(string); ok {
			masked["password"] = "****masked****"
		}
	}

	return masked
}

// MaskSensitiveString masks a string showing only the last 4 characters
func MaskSensitiveString(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return "***" + value[len(value)-4:]
}
