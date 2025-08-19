package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// LoadCacheConfig loads cache configuration from INI configuration
func LoadCacheConfig(cacheSettings map[string]string) *CacheConfig {
	cfg := DefaultCacheConfig()

	// Apply settings from INI [cache] section
	if cacheSettings != nil {
		if maxSizeStr, exists := cacheSettings["maxSize"]; exists {
			if maxSize, err := parseSize(maxSizeStr); err == nil {
				cfg.MaxSize = maxSize
			}
		}

		if ttlStr, exists := cacheSettings["ttl"]; exists {
			if ttl, err := time.ParseDuration(ttlStr); err == nil {
				cfg.TTL = Duration(ttl)
			}
		}

		if cleanupStr, exists := cacheSettings["cleanupInterval"]; exists {
			if cleanup, err := time.ParseDuration(cleanupStr); err == nil {
				cfg.CleanupInterval = Duration(cleanup)
			}
		}
	}

	return cfg
}

// parseSize parses size strings like "1GB", "500MB", "1073741824" (bytes)
func parseSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil
	}

	// Try parsing as plain number (bytes)
	if size, err := json.Number(sizeStr).Int64(); err == nil {
		return size, nil
	}

	// Parse size with units (GB, MB, KB)
	var size int64
	var unit string
	if _, err := fmt.Sscanf(strings.ToUpper(sizeStr), "%d%s", &size, &unit); err != nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	switch unit {
	case "GB":
		return size * 1024 * 1024 * 1024, nil
	case "MB":
		return size * 1024 * 1024, nil
	case "KB":
		return size * 1024, nil
	case "B", "":
		return size, nil
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
}
