package cache

import (
	"strings"
	"testing"
	"time"
)

func TestPackageCache(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &Storage{basePath: tmpDir}
	config := CacheConfig{PackageTTL: 0} // Never expire
	pc := NewPackageCache(storage, config)

	registryURL := "https://registry.armjs.org/"
	ruleset := "typescript-rules"
	version := "1.0.0"

	// Test cache miss
	if _, found := pc.Get(registryURL, ruleset, version); found {
		t.Error("Expected cache miss for non-existent package")
	}

	// Test store
	testData := strings.NewReader("test package data")
	path, err := pc.Store(registryURL, ruleset, version, testData)
	if err != nil {
		t.Fatalf("Failed to store package: %v", err)
	}

	// Test cache hit
	cachedPath, found := pc.Get(registryURL, ruleset, version)
	if !found {
		t.Error("Expected cache hit after storing package")
	}
	if cachedPath != path {
		t.Errorf("Cached path %q != stored path %q", cachedPath, path)
	}
}

func TestPackageCacheExpiration(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &Storage{basePath: tmpDir}
	config := CacheConfig{PackageTTL: time.Millisecond} // Very short TTL
	pc := NewPackageCache(storage, config)

	registryURL := "https://registry.armjs.org/"
	ruleset := "typescript-rules"
	version := "1.0.0"

	// Store package
	testData := strings.NewReader("test package data")
	_, err := pc.Store(registryURL, ruleset, version, testData)
	if err != nil {
		t.Fatalf("Failed to store package: %v", err)
	}

	// Wait for expiration
	time.Sleep(2 * time.Millisecond)

	// Test cache miss due to expiration
	if _, found := pc.Get(registryURL, ruleset, version); found {
		t.Error("Expected cache miss due to expiration")
	}
}
