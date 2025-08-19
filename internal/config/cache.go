package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// CacheConfig represents cache configuration settings
type CacheConfig struct {
	// MaxSize is the maximum cache size in bytes (0 = unlimited)
	MaxSize int64 `json:"maxSize"`

	// TTL is the time-to-live for cache entries (0 = no expiration)
	TTL Duration `json:"ttl"`

	// CleanupInterval is how often to run cache cleanup
	CleanupInterval Duration `json:"cleanupInterval"`
}

// Duration wraps time.Duration to provide JSON marshaling
type Duration time.Duration

// MarshalJSON implements json.Marshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// String returns the duration as a string
func (d Duration) String() string {
	return time.Duration(d).String()
}

// GetCachePath returns the cache directory path
func GetCachePath() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".arm", "cache")
}

// DefaultCacheConfig returns the default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize:         0,                        // Unlimited by default
		TTL:             Duration(24 * time.Hour), // 24 hours default
		CleanupInterval: Duration(6 * time.Hour),  // Cleanup every 6 hours
	}
}

// LoadCacheConfig loads cache configuration from dedicated cache/config.json file
func LoadCacheConfig() *CacheConfig {
	cfg := DefaultCacheConfig()

	// Load from dedicated cache config file
	cacheConfigPath := filepath.Join(GetCachePath(), "config.json")
	if loadedCfg, err := LoadCacheConfigFromFile(cacheConfigPath); err == nil {
		return loadedCfg
	}

	return cfg
}

// LoadCacheConfigFromFile loads cache configuration from specified JSON file
func LoadCacheConfigFromFile(path string) (*CacheConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err // Config file doesn't exist
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expand environment variables in JSON content
	expandedData := os.ExpandEnv(string(data))

	var cfg CacheConfig
	if err := json.Unmarshal([]byte(expandedData), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveCacheConfigToFile saves cache configuration to specified JSON file
func SaveCacheConfigToFile(path string, cfg *CacheConfig) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}
