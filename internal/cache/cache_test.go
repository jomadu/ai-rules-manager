package cache

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	maxSize := int64(1024 * 1024) // 1MB
	ttl := time.Hour

	cache := New(tempDir, maxSize, ttl)

	if cache.basePath != tempDir {
		t.Errorf("Expected basePath %s, got %s", tempDir, cache.basePath)
	}
	if cache.maxSize != maxSize {
		t.Errorf("Expected maxSize %d, got %d", maxSize, cache.maxSize)
	}
	if cache.ttl != ttl {
		t.Errorf("Expected ttl %v, got %v", ttl, cache.ttl)
	}
}

func TestVersionCache(t *testing.T) {
	tempDir := t.TempDir()
	cache := New(tempDir, 1024*1024, time.Hour)

	rulesets := map[string][]string{
		"my-rules":     {"1.0.0", "1.1.0"},
		"python-rules": {"2.0.0"},
	}

	// Test SetVersions
	err := cache.SetVersions("test-registry", rulesets, 3600)
	if err != nil {
		t.Fatalf("SetVersions failed: %v", err)
	}

	// Test GetVersions
	cachedVersions, err := cache.GetVersions("test-registry")
	if err != nil {
		t.Fatalf("GetVersions failed: %v", err)
	}

	if len(cachedVersions.Rulesets) != 2 {
		t.Errorf("Expected 2 rulesets, got %d", len(cachedVersions.Rulesets))
	}

	if len(cachedVersions.Rulesets["my-rules"]) != 2 {
		t.Errorf("Expected 2 versions for my-rules, got %d", len(cachedVersions.Rulesets["my-rules"]))
	}
}

func TestMetadataCache(t *testing.T) {
	tempDir := t.TempDir()
	cache := New(tempDir, 1024*1024, time.Hour)

	rulesets := map[string]map[string]string{
		"my-rules": {
			"description":    "Python coding rules",
			"latest_version": "1.2.0",
		},
	}

	// Test SetMetadata
	err := cache.SetMetadata("test-registry", rulesets, 3600)
	if err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}

	// Test GetMetadata
	cachedMetadata, err := cache.GetMetadata("test-registry")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	if cachedMetadata.Rulesets["my-rules"]["description"] != "Python coding rules" {
		t.Errorf("Expected description 'Python coding rules', got %s", cachedMetadata.Rulesets["my-rules"]["description"])
	}
}

func TestRulesetCache(t *testing.T) {
	tempDir := t.TempDir()
	cache := New(tempDir, 1024*1024, time.Hour)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.tar.gz")
	testData := []byte("test ruleset data")
	err := os.WriteFile(testFile, testData, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test SetRuleset
	err = cache.SetRuleset("test-registry", "my-rules", "1.0.0", testFile)
	if err != nil {
		t.Fatalf("SetRuleset failed: %v", err)
	}

	// Test GetRuleset
	cachedPath, err := cache.GetRuleset("test-registry", "my-rules", "1.0.0")
	if err != nil {
		t.Fatalf("GetRuleset failed: %v", err)
	}

	// Verify file exists and has correct content
	cachedData, err := os.ReadFile(cachedPath)
	if err != nil {
		t.Fatalf("Failed to read cached file: %v", err)
	}

	if !bytes.Equal(cachedData, testData) {
		t.Errorf("Expected cached data %s, got %s", string(testData), string(cachedData))
	}
}

func TestTTLExpiration(t *testing.T) {
	tempDir := t.TempDir()
	cache := New(tempDir, 1024*1024, time.Hour)

	rulesets := map[string][]string{
		"my-rules": {"1.0.0"},
	}

	// Set with very short TTL (1 second)
	err := cache.SetVersions("test-registry", rulesets, 1)
	if err != nil {
		t.Fatalf("SetVersions failed: %v", err)
	}

	// Should work immediately
	_, err = cache.GetVersions("test-registry")
	if err != nil {
		t.Fatalf("GetVersions should work immediately: %v", err)
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Should be expired now
	_, err = cache.GetVersions("test-registry")
	if err == nil {
		t.Error("GetVersions should fail after TTL expiration")
	}
}

func TestLRUEviction(t *testing.T) {
	tempDir := t.TempDir()
	// Set very small cache size to trigger eviction
	cache := New(tempDir, 100, time.Hour)

	// Create test files that exceed cache size
	testFile1 := filepath.Join(tempDir, "test1.tar.gz")
	testFile2 := filepath.Join(tempDir, "test2.tar.gz")

	testData1 := make([]byte, 60) // 60 bytes
	testData2 := make([]byte, 60) // 60 bytes (total 120 > 100)

	err := os.WriteFile(testFile1, testData1, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}

	err = os.WriteFile(testFile2, testData2, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	// Add first ruleset
	err = cache.SetRuleset("test-registry", "rules1", "1.0.0", testFile1)
	if err != nil {
		t.Fatalf("SetRuleset 1 failed: %v", err)
	}

	// Add second ruleset (should trigger eviction of first)
	err = cache.SetRuleset("test-registry", "rules2", "1.0.0", testFile2)
	if err != nil {
		t.Fatalf("SetRuleset 2 failed: %v", err)
	}

	// First ruleset should be evicted
	_, err = cache.GetRuleset("test-registry", "rules1", "1.0.0")
	if err == nil {
		t.Error("First ruleset should have been evicted")
	}

	// Second ruleset should still be available
	_, err = cache.GetRuleset("test-registry", "rules2", "1.0.0")
	if err != nil {
		t.Errorf("Second ruleset should still be available: %v", err)
	}
}

func TestClean(t *testing.T) {
	tempDir := t.TempDir()
	cache := New(tempDir, 1024*1024, time.Hour)

	rulesets := map[string][]string{
		"my-rules": {"1.0.0"},
	}

	// Add some data
	err := cache.SetVersions("test-registry", rulesets, 3600)
	if err != nil {
		t.Fatalf("SetVersions failed: %v", err)
	}

	// Verify data exists
	_, err = cache.GetVersions("test-registry")
	if err != nil {
		t.Fatalf("GetVersions should work: %v", err)
	}

	// Clean cache
	err = cache.Clean()
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	// Verify data is gone
	_, err = cache.GetVersions("test-registry")
	if err == nil {
		t.Error("GetVersions should fail after clean")
	}

	// Verify cache size is 0
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clean, got %d", cache.Size())
	}
}
