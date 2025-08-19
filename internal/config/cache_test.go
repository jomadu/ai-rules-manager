package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheConfigFromFile(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cacheConfigPath := filepath.Join(tempDir, "config.json")
	cacheContent := `{
  "maxSize": 2000000000,
  "ttl": "12h",
  "cleanupInterval": "3h"
}`
	if err := os.WriteFile(cacheConfigPath, []byte(cacheContent), 0o600); err != nil {
		t.Fatal(err)
	}

	// Load configuration from file
	cfg, err := LoadCacheConfigFromFile(cacheConfigPath)
	if err != nil {
		t.Fatalf("Failed to load cache config: %v", err)
	}

	// Verify values
	if cfg.MaxSize != 2000000000 {
		t.Errorf("Expected maxSize 2000000000, got %d", cfg.MaxSize)
	}
	if time.Duration(cfg.TTL) != 12*time.Hour {
		t.Errorf("Expected TTL 12h, got %v", cfg.TTL)
	}
	if time.Duration(cfg.CleanupInterval) != 3*time.Hour {
		t.Errorf("Expected cleanup interval 3h, got %v", cfg.CleanupInterval)
	}
}

func TestCacheConfigDefaults(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-cache-defaults-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create project directory (no cache config)
	projectDir := filepath.Join(tempDir, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to project directory
	originalWd, _ := os.Getwd()
	_ = os.Chdir(projectDir)
	defer func() { _ = os.Chdir(originalWd) }()

	// Load configuration (should use defaults)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify cache config uses defaults
	if cfg.CacheConfig.MaxSize != 0 {
		t.Errorf("Expected default cache maxSize 0, got %d", cfg.CacheConfig.MaxSize)
	}
	if time.Duration(cfg.CacheConfig.TTL) != 24*time.Hour {
		t.Errorf("Expected default cache TTL 24h, got %v", cfg.CacheConfig.TTL)
	}
	if time.Duration(cfg.CacheConfig.CleanupInterval) != 6*time.Hour {
		t.Errorf("Expected default cache cleanup interval 6h, got %v", cfg.CacheConfig.CleanupInterval)
	}
}

func TestSaveCacheConfigToFile(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-save-cache-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	configPath := filepath.Join(tempDir, "cache", "config.json")
	cfg := &CacheConfig{
		MaxSize:         5000000000,
		TTL:             Duration(8 * time.Hour),
		CleanupInterval: Duration(2 * time.Hour),
	}

	// Save config
	if err := SaveCacheConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to save cache config: %v", err)
	}

	// Load it back
	loadedCfg, err := LoadCacheConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	// Verify values
	if loadedCfg.MaxSize != 5000000000 {
		t.Errorf("Expected maxSize 5000000000, got %d", loadedCfg.MaxSize)
	}
	if time.Duration(loadedCfg.TTL) != 8*time.Hour {
		t.Errorf("Expected TTL 8h, got %v", loadedCfg.TTL)
	}
	if time.Duration(loadedCfg.CleanupInterval) != 2*time.Hour {
		t.Errorf("Expected cleanup interval 2h, got %v", loadedCfg.CleanupInterval)
	}
}

func TestDurationJSONMarshaling(t *testing.T) {
	d := Duration(2 * time.Hour)

	// Test marshaling
	data, err := d.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal duration: %v", err)
	}

	expected := `"2h0m0s"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}

	// Test unmarshaling
	var d2 Duration
	if err := d2.UnmarshalJSON(data); err != nil {
		t.Fatalf("Failed to unmarshal duration: %v", err)
	}

	if time.Duration(d2) != time.Duration(d) {
		t.Errorf("Duration mismatch after round-trip: expected %v, got %v", d, d2)
	}
}

func TestDefaultCacheConfigIsolated(t *testing.T) {
	cfg := DefaultCacheConfig()

	if cfg.MaxSize != 0 {
		t.Errorf("Expected default MaxSize 0, got %d", cfg.MaxSize)
	}

	if time.Duration(cfg.TTL) != 24*time.Hour {
		t.Errorf("Expected default TTL 24h, got %v", cfg.TTL)
	}

	if time.Duration(cfg.CleanupInterval) != 6*time.Hour {
		t.Errorf("Expected default CleanupInterval 6h, got %v", cfg.CleanupInterval)
	}
}

func TestGetCachePath(t *testing.T) {
	// Set up fake HOME
	originalHome := os.Getenv("HOME")
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	_ = os.Setenv("HOME", "/test/home")

	path := GetCachePath()
	expectedPath := "/test/home/.arm/cache"
	if path != expectedPath {
		t.Errorf("Expected cache path %s, got %s", expectedPath, path)
	}
}

func TestLoadCacheConfigFromNonexistentFile(t *testing.T) {
	cfg, err := LoadCacheConfigFromFile("/nonexistent/config.json")
	if err == nil {
		t.Error("Expected error when loading nonexistent file")
	}
	if cfg != nil {
		t.Error("Expected nil config when file doesn't exist")
	}
}
