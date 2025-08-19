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
	defer os.RemoveAll(tempDir)

	// Set up fake HOME directory
	originalHome := os.Getenv("HOME")
	fakeHome := filepath.Join(tempDir, "home")
	os.Setenv("HOME", fakeHome)
	defer os.Setenv("HOME", originalHome)

	// Create cache directory and config file
	cacheDir := filepath.Join(fakeHome, ".arm", "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cacheConfigPath := filepath.Join(cacheDir, "config.json")
	cacheContent := `{
  "path": "/test/cache",
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
	if cfg.Path != "/test/cache" {
		t.Errorf("Expected path '/test/cache', got '%s'", cfg.Path)
	}
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
	defer os.RemoveAll(tempDir)

	// Set up fake HOME directory
	originalHome := os.Getenv("HOME")
	fakeHome := filepath.Join(tempDir, "home")
	os.Setenv("HOME", fakeHome)
	defer os.Setenv("HOME", originalHome)

	// Create project directory (no cache config)
	projectDir := filepath.Join(tempDir, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to project directory
	originalWd, _ := os.Getwd()
	os.Chdir(projectDir)
	defer os.Chdir(originalWd)

	// Load configuration (should use defaults)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify cache config uses defaults
	expectedPath := filepath.Join(fakeHome, ".arm", "cache")
	if cfg.CacheConfig.Path != expectedPath {
		t.Errorf("Expected default cache path '%s', got '%s'", expectedPath, cfg.CacheConfig.Path)
	}
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
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "cache", "config.json")
	cfg := &CacheConfig{
		Path:            "/custom/cache",
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
	if loadedCfg.Path != "/custom/cache" {
		t.Errorf("Expected path '/custom/cache', got '%s'", loadedCfg.Path)
	}
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

func TestLoadCacheConfigFromNonexistentFile(t *testing.T) {
	cfg, err := LoadCacheConfigFromFile("/nonexistent/config.json")
	if err == nil {
		t.Error("Expected error when loading nonexistent file")
	}
	if cfg != nil {
		t.Error("Expected nil config when file doesn't exist")
	}
}
