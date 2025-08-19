package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheConfigJSON(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-cache-json-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// Test saving and loading
	originalCfg := &CacheConfig{
		Path:            "/test/cache/path",
		MaxSize:         1000000000,
		TTL:             Duration(12 * time.Hour),
		CleanupInterval: Duration(3 * time.Hour),
	}

	// Save config
	if err := SaveCacheConfigToFile(configPath, originalCfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedCfg, err := LoadCacheConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if loadedCfg.Path != originalCfg.Path {
		t.Errorf("Path mismatch: expected %s, got %s", originalCfg.Path, loadedCfg.Path)
	}
	if loadedCfg.MaxSize != originalCfg.MaxSize {
		t.Errorf("MaxSize mismatch: expected %d, got %d", originalCfg.MaxSize, loadedCfg.MaxSize)
	}
	if time.Duration(loadedCfg.TTL) != time.Duration(originalCfg.TTL) {
		t.Errorf("TTL mismatch: expected %v, got %v", originalCfg.TTL, loadedCfg.TTL)
	}
	if time.Duration(loadedCfg.CleanupInterval) != time.Duration(originalCfg.CleanupInterval) {
		t.Errorf("CleanupInterval mismatch: expected %v, got %v", originalCfg.CleanupInterval, loadedCfg.CleanupInterval)
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

func TestDefaultCacheConfig(t *testing.T) {
	// Set up fake HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	os.Setenv("HOME", "/test/home")

	cfg := DefaultCacheConfig()

	expectedPath := "/test/home/.arm/cache"
	if cfg.Path != expectedPath {
		t.Errorf("Expected default path %s, got %s", expectedPath, cfg.Path)
	}

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
