package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	cacheRoot := "/tmp/test-cache"
	manager := NewManager(cacheRoot)

	if manager.cacheRoot != cacheRoot {
		t.Errorf("Expected cache root %s, got %s", cacheRoot, manager.cacheRoot)
	}
}

func TestNormalizeURL(t *testing.T) {
	manager := NewManager("/tmp/cache")

	// Test the deprecated generic normalization method
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/user/repo", "https://github.com/user/repo"},
		{"https://github.com/user/repo/", "https://github.com/user/repo"},
		{"HTTPS://GITHUB.COM/User/Repo", "https://github.com/user/repo"},
		// Note: Generic normalization doesn't handle SSH URLs
		{"git@github.com:user/repo.git", "git@github.com:user/repo.git"},
		{"git@github.com:user/repo", "git@github.com:user/repo"},
	}

	for _, test := range tests {
		result := manager.NormalizeURL(test.input)
		if result != test.expected {
			t.Errorf("NormalizeURL(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestGetCacheKey(t *testing.T) {
	manager := NewManager("/tmp/cache")

	// Test that same registry configs produce same keys
	registryType := "git"
	url := "https://github.com/user/repo"
	key1, err1 := manager.GetCacheKey(registryType, url)
	key2, err2 := manager.GetCacheKey(registryType, url)

	if err1 != nil || err2 != nil {
		t.Fatalf("GetCacheKey failed: %v, %v", err1, err2)
	}

	if key1 != key2 {
		t.Errorf("Same registry config should produce same key: %s != %s", key1, key2)
	}

	// Test that key is 64 characters (SHA-256 hex)
	if len(key1) != 64 {
		t.Errorf("Cache key should be 64 characters, got %d", len(key1))
	}

	// Test that different URLs produce different keys
	key3, err3 := manager.GetCacheKey(registryType, "https://github.com/other/repo")
	if err3 != nil {
		t.Fatalf("GetCacheKey failed: %v", err3)
	}

	if key1 == key3 {
		t.Errorf("Different URLs should produce different keys")
	}

	// Test that different registry types produce different keys for same URL
	key4, err4 := manager.GetCacheKey("https", url)
	if err4 != nil {
		t.Fatalf("GetCacheKey failed: %v", err4)
	}

	if key1 == key4 {
		t.Errorf("Different registry types should produce different keys for same URL")
	}
}

func TestGetCachePath(t *testing.T) {
	cacheRoot := "/tmp/test-cache"
	manager := NewManager(cacheRoot)

	registryType := "git"
	url := "https://github.com/user/repo"
	path, err := manager.GetCachePath(registryType, url)

	if err != nil {
		t.Fatalf("GetCachePath failed: %v", err)
	}

	// Should be under cache root/registries/hash
	expectedPrefix := filepath.Join(cacheRoot, "registries")
	if !strings.HasPrefix(path, expectedPrefix) {
		t.Errorf("Cache path should start with %s, got %s", expectedPrefix, path)
	}

	// Should end with 64-character hash
	hash := filepath.Base(path)
	if len(hash) != 64 {
		t.Errorf("Cache path should end with 64-character hash, got %s", hash)
	}
}

func TestEnsureCacheDir(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager := NewManager(tempDir)
	registryType := "git"
	url := "https://github.com/user/repo"

	// Ensure cache directory is created
	err = manager.EnsureCacheDir(registryType, url)
	if err != nil {
		t.Fatalf("EnsureCacheDir failed: %v", err)
	}

	// Verify directory exists
	cachePath, err := manager.GetCachePath(registryType, url)
	if err != nil {
		t.Fatalf("GetCachePath failed: %v", err)
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Errorf("Cache directory was not created: %s", cachePath)
	}

	// Test idempotency - calling again should not fail
	err = manager.EnsureCacheDir(registryType, url)
	if err != nil {
		t.Errorf("EnsureCacheDir should be idempotent: %v", err)
	}
}

func TestUpdateCacheInfo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager := NewManager(tempDir)
	registryType := "git"
	registryURL := "https://github.com/user/repo"
	version := "v1.0.0"

	// Update cache info
	err = manager.UpdateCacheInfo(registryType, registryURL, version)
	if err != nil {
		t.Fatalf("UpdateCacheInfo failed: %v", err)
	}

	// Verify cache-info.json was created
	cachePath, err := manager.GetCachePath(registryType, registryURL)
	if err != nil {
		t.Fatalf("GetCachePath failed: %v", err)
	}

	infoPath := filepath.Join(cachePath, "cache-info.json")
	if _, err := os.Stat(infoPath); os.IsNotExist(err) {
		t.Error("cache-info.json should be created")
	}

	// Load and verify cache info
	cacheInfo, err := manager.loadCacheInfo(cachePath)
	if err != nil {
		t.Fatalf("Failed to load cache info: %v", err)
	}

	if cacheInfo.RegistryType != registryType {
		t.Errorf("Expected registry type %s, got %s", registryType, cacheInfo.RegistryType)
	}
	if cacheInfo.RegistryURL != registryURL {
		t.Errorf("Expected registry URL %s, got %s", registryURL, cacheInfo.RegistryURL)
	}
	if cacheInfo.Version != version {
		t.Errorf("Expected version %s, got %s", version, cacheInfo.Version)
	}
}
func TestIsCacheValid(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Test non-existent cache
	valid, err := manager.IsCacheValid("git", "https://github.com/user/repo", time.Hour)
	if err != nil || valid {
		t.Error("Non-existent cache should be invalid")
	}

	// Create cache entry
	_ = manager.UpdateCacheInfo("git", "https://github.com/user/repo", "v1.0.0")

	// Test valid cache (recent)
	valid, err = manager.IsCacheValid("git", "https://github.com/user/repo", time.Hour)
	if err != nil || !valid {
		t.Error("Recent cache should be valid")
	}

	// Test no TTL (always valid)
	valid, err = manager.IsCacheValid("git", "https://github.com/user/repo", 0)
	if err != nil || !valid {
		t.Error("Cache with no TTL should always be valid")
	}
}

func TestGetCacheSize(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Empty cache should have size 0
	size, err := manager.GetCacheSize()
	if err != nil || size != 0 {
		t.Errorf("Empty cache size should be 0, got %d", size)
	}

	// Create cache entry
	_ = manager.UpdateCacheInfo("git", "https://github.com/user/repo", "v1.0.0")

	// Size should be > 0 after creating cache info
	size, err = manager.GetCacheSize()
	if err != nil || size <= 0 {
		t.Errorf("Cache with entries should have size > 0, got %d", size)
	}
}

func TestCleanupExpired(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create cache entry
	_ = manager.UpdateCacheInfo("git", "https://github.com/user/repo", "v1.0.0")

	// Cleanup with no TTL should do nothing
	err := manager.CleanupExpired(0)
	if err != nil {
		t.Errorf("CleanupExpired with no TTL failed: %v", err)
	}

	// Cleanup with long TTL should keep entries
	err = manager.CleanupExpired(time.Hour)
	if err != nil {
		t.Errorf("CleanupExpired failed: %v", err)
	}

	// Verify entry still exists
	valid, _ := manager.IsCacheValid("git", "https://github.com/user/repo", time.Hour)
	if !valid {
		t.Error("Cache entry should still exist after cleanup with long TTL")
	}
}

func TestCleanupOversized(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create cache entry
	_ = manager.UpdateCacheInfo("git", "https://github.com/user/repo", "v1.0.0")

	// Cleanup with no size limit should do nothing
	err := manager.CleanupOversized(0)
	if err != nil {
		t.Errorf("CleanupOversized with no limit failed: %v", err)
	}

	// Cleanup with large size limit should keep entries
	err = manager.CleanupOversized(1024 * 1024) // 1MB
	if err != nil {
		t.Errorf("CleanupOversized failed: %v", err)
	}
}
