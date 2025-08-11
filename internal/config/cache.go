package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// CacheConfig represents cache configuration settings
type CacheConfig struct {
	// Path is the root directory for cache storage
	Path string `json:"path"`

	// MaxSize is the maximum cache size in bytes (0 = unlimited)
	MaxSize int64 `json:"max_size"`

	// TTL is the time-to-live for cache entries (0 = no expiration)
	TTL time.Duration `json:"ttl"`

	// CleanupInterval is how often to run cache cleanup
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// DefaultCacheConfig returns the default cache configuration
func DefaultCacheConfig() *CacheConfig {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "."
	}

	return &CacheConfig{
		Path:            filepath.Join(homeDir, ".arm", "cache"),
		MaxSize:         0,              // Unlimited by default
		TTL:             24 * time.Hour, // 24 hours default
		CleanupInterval: 6 * time.Hour,  // Cleanup every 6 hours
	}
}

// LoadCacheConfig loads cache configuration from the config sections
func (c *Config) LoadCacheConfig() *CacheConfig {
	cfg := DefaultCacheConfig()

	// Load from [cache] section in INI file
	if cacheSection, exists := c.TypeDefaults["cache"]; exists {
		if path, ok := cacheSection["path"]; ok && path != "" {
			cfg.Path = expandEnvVars(path)
		}

		if maxSizeStr, ok := cacheSection["maxSize"]; ok && maxSizeStr != "" {
			if maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
				cfg.MaxSize = maxSize
			}
		}

		if ttlStr, ok := cacheSection["ttl"]; ok && ttlStr != "" {
			if ttl, err := time.ParseDuration(ttlStr); err == nil {
				cfg.TTL = ttl
			}
		}

		if cleanupStr, ok := cacheSection["cleanupInterval"]; ok && cleanupStr != "" {
			if cleanup, err := time.ParseDuration(cleanupStr); err == nil {
				cfg.CleanupInterval = cleanup
			}
		}
	}

	return cfg
}
