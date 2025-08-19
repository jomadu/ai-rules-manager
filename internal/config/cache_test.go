package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheConfigFromINI(t *testing.T) {
	// Test cache configuration from INI settings
	cacheSettings := map[string]string{
		"maxSize":         "2GB",
		"ttl":             "12h",
		"cleanupInterval": "3h",
	}

	cfg := LoadCacheConfig(cacheSettings)

	// Verify values
	if cfg.MaxSize != 2*1024*1024*1024 {
		t.Errorf("Expected maxSize 2GB (2147483648 bytes), got %d", cfg.MaxSize)
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

func TestParseSizeFunction(t *testing.T) {
	tests := []struct {
		input     string
		expected  int64
		shouldErr bool
	}{
		{"1GB", 1024 * 1024 * 1024, false},
		{"500MB", 500 * 1024 * 1024, false},
		{"2048KB", 2048 * 1024, false},
		{"1073741824", 1073741824, false},
		{"0", 0, false},
		{"", 0, false},
		{"invalid", 0, true},
		{"1XB", 0, true},
	}

	for _, test := range tests {
		result, err := parseSize(test.input)
		if test.shouldErr {
			if err == nil {
				t.Errorf("Expected error for input %s, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("For input %s, expected %d, got %d", test.input, test.expected, result)
			}
		}
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

func TestCacheConfigWithNilSettings(t *testing.T) {
	cfg := LoadCacheConfig(nil)

	// Should return default configuration
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
