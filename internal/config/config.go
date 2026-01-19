package config

import (
	"encoding/json"
	"os"
)

const (
	DefaultAPIUrl = "https://kiosk-four.vercel.app"
	EnvAPIUrl     = "KIOSK_API_URL"
)

// Config holds the kiosk CLI configuration
type Config struct {
	APIUrl string `json:"apiUrl"`
}

// Default returns a Config with default values
func Default() *Config {
	return &Config{
		APIUrl: DefaultAPIUrl,
	}
}

// Load reads the config from disk and applies env var overrides
func Load() (*Config, error) {
	cfg := Default()

	// Try to read from config file
	data, err := os.ReadFile(ConfigPath())
	if err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// Env var overrides
	if envURL := os.Getenv(EnvAPIUrl); envURL != "" {
		cfg.APIUrl = envURL
	}

	return cfg, nil
}

// Save writes the config to disk
func Save(cfg *Config) error {
	// Ensure directories exist
	if err := os.MkdirAll(KioskDir(), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), data, 0644)
}

// EnsureInitialized creates the kiosk directory structure if it doesn't exist
func EnsureInitialized() error {
	dirs := []string{
		KioskDir(),
		AppsDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create default config if it doesn't exist
	if _, err := os.Stat(ConfigPath()); os.IsNotExist(err) {
		data, err := json.MarshalIndent(Default(), "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(ConfigPath(), data, 0644); err != nil {
			return err
		}
	}

	return nil
}
