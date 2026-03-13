package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration.
type Config struct {
	Devices         map[string]DeviceConfig `json:"devices"` // key = hex BT address
	PollIntervalSec int                     `json:"poll_interval_sec"`
}

// DeviceConfig holds per-device settings.
type DeviceConfig struct {
	AutoDisableHFP bool `json:"auto_disable_hfp"`
}

// DefaultPath returns the default config file path: %APPDATA%\NoHandsfree\config.json
func DefaultPath() (string, error) {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		return "", errors.New("APPDATA environment variable not set")
	}
	return filepath.Join(appdata, "NoHandsfree", "config.json"), nil
}

// Load reads config from path. Returns default config if the file doesn't exist.
func Load(path string) (*Config, error) {
	cfg := &Config{
		Devices:         make(map[string]DeviceConfig),
		PollIntervalSec: 5,
	}

	data, err := os.ReadFile(path) //nolint:gosec // path is derived from %APPDATA%, not user input
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Devices == nil {
		cfg.Devices = make(map[string]DeviceConfig)
	}
	if cfg.PollIntervalSec <= 0 {
		cfg.PollIntervalSec = 5
	}
	return cfg, nil
}

// Save writes config to path, creating directories as needed.
func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
