package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Loader handles loading provider configurations from files
type Loader struct {
	configDir string
}

// NewLoader creates a new configuration loader
func NewLoader(configDir string) *Loader {
	return &Loader{
		configDir: configDir,
	}
}

// LoadAll loads all provider configurations from the config directory
func (l *Loader) LoadAll() ([]*ProviderConfig, error) {
	// Check if directory exists
	if _, err := os.Stat(l.configDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("config directory does not exist: %s", l.configDir)
	}

	// Read all JSON files
	files, err := filepath.Glob(filepath.Join(l.configDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list config files: %w", err)
	}

	var configs []*ProviderConfig
	var errors []error

	for _, file := range files {
		config, err := l.LoadFile(file)
		if err != nil {
			errors = append(errors, fmt.Errorf("file %s: %w", filepath.Base(file), err))
			continue
		}
		configs = append(configs, config)
	}

	// Log errors but continue with valid configs
	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d config files failed to load\n", len(errors))
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "  - %v\n", err)
		}
	}

	return configs, nil
}

// LoadFile loads a single provider configuration from a file
func (l *Loader) LoadFile(filename string) (*ProviderConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var config ProviderConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate the config
	if err := ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &config, nil
}

// GetConfigPath returns the absolute path for a config file
func (l *Loader) GetConfigPath(filename string) string {
	return filepath.Join(l.configDir, filename)
}
