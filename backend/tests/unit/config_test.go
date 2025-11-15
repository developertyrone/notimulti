package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/developertyrone/notimulti/internal/config"
)

func TestLoadAll(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create valid telegram config
	telegramConfig := `{
		"id": "telegram-main",
		"type": "telegram",
		"enabled": true,
		"config": {
			"bot_token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
			"default_chat_id": "12345678"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "telegram.json"), []byte(telegramConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Create valid email config
	emailConfig := `{
		"id": "email-main",
		"type": "email",
		"enabled": true,
		"config": {
			"host": "smtp.gmail.com",
			"port": 587,
			"username": "test@example.com",
			"password": "secret123",
			"from": "test@example.com"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "email.json"), []byte(emailConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Create invalid config (should be skipped)
	invalidConfig := `{"invalid": "json"`
	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.json"), []byte(invalidConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Test LoadAll
	loader := config.NewLoader(tmpDir)
	configs, err := loader.LoadAll()

	// Should succeed with 2 valid configs despite 1 invalid
	if err != nil {
		t.Errorf("LoadAll failed: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	// Verify telegram config
	var foundTelegram, foundEmail bool
	for _, cfg := range configs {
		if cfg.ID == "telegram-main" {
			foundTelegram = true
			if cfg.Type != "telegram" {
				t.Errorf("Expected type telegram, got %s", cfg.Type)
			}
		}
		if cfg.ID == "email-main" {
			foundEmail = true
			if cfg.Type != "email" {
				t.Errorf("Expected type email, got %s", cfg.Type)
			}
		}
	}

	if !foundTelegram {
		t.Error("Telegram config not found")
	}
	if !foundEmail {
		t.Error("Email config not found")
	}
}

func TestLoadAllNonexistentDir(t *testing.T) {
	loader := config.NewLoader("/nonexistent/path")
	configs, err := loader.LoadAll()

	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}

	if configs != nil {
		t.Error("Expected nil configs for nonexistent directory")
	}
}

func TestLoadFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "valid telegram config",
			content: `{
				"id": "test-telegram",
				"type": "telegram",
				"enabled": true,
				"config": {
					"bot_token": "123456:ABC",
					"default_chat_id": "12345"
				}
			}`,
			wantError: false,
		},
		{
			name: "valid email config",
			content: `{
				"id": "test-email",
				"type": "email",
				"enabled": false,
				"config": {
					"host": "smtp.example.com",
					"port": 587,
					"username": "user",
					"password": "pass",
					"from": "test@example.com"
				}
			}`,
			wantError: false,
		},
		{
			name:      "invalid json",
			content:   `{"invalid": json}`,
			wantError: true,
		},
		{
			name: "missing id",
			content: `{
				"type": "telegram",
				"enabled": true,
				"config": {"bot_token": "123", "default_chat_id": "456"}
			}`,
			wantError: true,
		},
		{
			name: "missing type",
			content: `{
				"id": "test",
				"enabled": true,
				"config": {"bot_token": "123", "default_chat_id": "456"}
			}`,
			wantError: true,
		},
		{
			name: "invalid id pattern",
			content: `{
				"id": "Test_Invalid",
				"type": "telegram",
				"enabled": true,
				"config": {"bot_token": "123", "default_chat_id": "456"}
			}`,
			wantError: true,
		},
		{
			name: "unsupported type",
			content: `{
				"id": "test",
				"type": "sms",
				"enabled": true,
				"config": {}
			}`,
			wantError: true,
		},
		{
			name: "telegram missing bot_token",
			content: `{
				"id": "test",
				"type": "telegram",
				"enabled": true,
				"config": {"default_chat_id": "456"}
			}`,
			wantError: true,
		},
		{
			name: "telegram missing chat_id",
			content: `{
				"id": "test",
				"type": "telegram",
				"enabled": true,
				"config": {"bot_token": "123"}
			}`,
			wantError: true,
		},
		{
			name: "email missing host",
			content: `{
				"id": "test",
				"type": "email",
				"enabled": true,
				"config": {
					"port": 587,
					"username": "user",
					"password": "pass",
					"from": "test@example.com"
				}
			}`,
			wantError: true,
		},
		{
			name: "email invalid port",
			content: `{
				"id": "test",
				"type": "email",
				"enabled": true,
				"config": {
					"host": "smtp.example.com",
					"port": 99999,
					"username": "user",
					"password": "pass",
					"from": "test@example.com"
				}
			}`,
			wantError: true,
		},
		{
			name: "email invalid from address",
			content: `{
				"id": "test",
				"type": "email",
				"enabled": true,
				"config": {
					"host": "smtp.example.com",
					"port": 587,
					"username": "user",
					"password": "pass",
					"from": "not-an-email"
				}
			}`,
			wantError: true,
		},
	}

	loader := config.NewLoader(tmpDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(tmpDir, "test.json")
			if err := os.WriteFile(filename, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
					t.Fatalf("failed to remove file %s: %v", filename, err)
				}
			})

			config, err := loader.LoadFile(filename)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if config == nil {
					t.Error("Expected config but got nil")
				}
			}
		})
	}
}

func TestValidateConfigs(t *testing.T) {
	tests := []struct {
		name      string
		configs   []*config.ProviderConfig
		wantError bool
	}{
		{
			name: "valid configs",
			configs: []*config.ProviderConfig{
				{
					ID:      "telegram-1",
					Type:    "telegram",
					Enabled: true,
					Config:  map[string]interface{}{"bot_token": "123", "default_chat_id": "456"},
				},
				{
					ID:      "email-1",
					Type:    "email",
					Enabled: true,
					Config: map[string]interface{}{
						"host":     "smtp.example.com",
						"port":     587.0,
						"username": "user",
						"password": "pass",
						"from":     "test@example.com",
					},
				},
			},
			wantError: false,
		},
		{
			name: "duplicate IDs",
			configs: []*config.ProviderConfig{
				{
					ID:      "duplicate",
					Type:    "telegram",
					Enabled: true,
					Config:  map[string]interface{}{"bot_token": "123", "default_chat_id": "456"},
				},
				{
					ID:      "duplicate",
					Type:    "email",
					Enabled: true,
					Config: map[string]interface{}{
						"host":     "smtp.example.com",
						"port":     587.0,
						"username": "user",
						"password": "pass",
						"from":     "test@example.com",
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateConfigs(tt.configs)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateConfigTelegramNumericChatID(t *testing.T) {
	cfg := &config.ProviderConfig{
		ID:      "telegram-numeric",
		Type:    "telegram",
		Enabled: true,
		Config: map[string]interface{}{
			"bot_token":       "123:ABC",
			"default_chat_id": float64(987654),
		},
	}

	if err := config.ValidateConfig(cfg); err != nil {
		t.Fatalf("ValidateConfig failed for numeric chat id: %v", err)
	}
}

func TestValidateConfigEmailStringPort(t *testing.T) {
	cfg := &config.ProviderConfig{
		ID:      "email-string-port",
		Type:    "email",
		Enabled: true,
		Config: map[string]interface{}{
			"host":     "smtp.example.com",
			"port":     "587",
			"username": "user@example.com",
			"password": "secret",
			"from":     "user@example.com",
		},
	}

	if err := config.ValidateConfig(cfg); err != nil {
		t.Fatalf("ValidateConfig failed for string port: %v", err)
	}
}

func TestGetConfigPath(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "configs")
	loader := config.NewLoader(baseDir)

	got := loader.GetConfigPath("email.json")
	want := filepath.Join(baseDir, "email.json")

	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
