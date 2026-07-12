// Package config loads and stores gust's persistent default settings in a
// small JSON file under the user's configuration directory.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config holds user-overridable defaults for a gust run.
type Config struct {
	SizeMB  int  `json:"size_mb"`
	Streams int  `json:"streams"`
	Pings   int  `json:"pings"`
	NoColor bool `json:"no_color"`
}

// Default returns the built-in defaults used when no config file exists.
func Default() Config {
	return Config{SizeMB: 25, Streams: 4, Pings: 6}
}

// Path returns the config file location, honouring XDG_CONFIG_HOME.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gust", "config.json"), nil
}

// Load reads the config file, returning Default() when it does not exist.
func Load() (Config, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Save writes the config file, creating parent directories as needed.
func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
