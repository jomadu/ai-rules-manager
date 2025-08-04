package cache

import (
	"testing"
	"time"
)

func TestMetadataCache(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &Storage{basePath: tmpDir}
	config := CacheConfig{
		MetadataTTL: time.Hour,
		VersionTTL:  15 * time.Minute,
	}
	mc := NewMetadataCache(storage, config)

	registryURL := "https://registry.armjs.org/"

	// Test versions cache miss
	if _, found := mc.GetVersions(registryURL); found {
		t.Error("Expected cache miss for non-existent versions")
	}

	// Test store and retrieve versions
	versions := []string{"1.0.0", "1.1.0", "2.0.0"}
	if err := mc.StoreVersions(registryURL, versions); err != nil {
		t.Fatalf("Failed to store versions: %v", err)
	}

	cachedVersions, found := mc.GetVersions(registryURL)
	if !found {
		t.Error("Expected cache hit after storing versions")
	}
	if len(cachedVersions) != len(versions) {
		t.Errorf("Cached versions length %d != stored length %d", len(cachedVersions), len(versions))
	}
	for i, v := range versions {
		if cachedVersions[i] != v {
			t.Errorf("Cached version[%d] %q != stored %q", i, cachedVersions[i], v)
		}
	}

	// Test metadata cache miss
	if _, found := mc.GetMetadata(registryURL); found {
		t.Error("Expected cache miss for non-existent metadata")
	}

	// Test store and retrieve metadata
	metadata := map[string]interface{}{
		"name":        "test-registry",
		"description": "Test registry",
		"version":     "1.0.0",
	}
	if err := mc.StoreMetadata(registryURL, metadata); err != nil {
		t.Fatalf("Failed to store metadata: %v", err)
	}

	cachedMetadata, found := mc.GetMetadata(registryURL)
	if !found {
		t.Error("Expected cache hit after storing metadata")
	}
	if cachedMetadata["name"] != metadata["name"] {
		t.Errorf("Cached metadata name %q != stored %q", cachedMetadata["name"], metadata["name"])
	}
}

func TestMetadataCacheExpiration(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &Storage{basePath: tmpDir}
	config := CacheConfig{
		MetadataTTL: time.Millisecond,
		VersionTTL:  time.Millisecond,
	}
	mc := NewMetadataCache(storage, config)

	registryURL := "https://registry.armjs.org/"

	// Store versions and metadata
	versions := []string{"1.0.0"}
	metadata := map[string]interface{}{"test": "data"}

	if err := mc.StoreVersions(registryURL, versions); err != nil {
		t.Fatalf("Failed to store versions: %v", err)
	}
	if err := mc.StoreMetadata(registryURL, metadata); err != nil {
		t.Fatalf("Failed to store metadata: %v", err)
	}

	// Wait for expiration
	time.Sleep(2 * time.Millisecond)

	// Test cache miss due to expiration
	if _, found := mc.GetVersions(registryURL); found {
		t.Error("Expected cache miss for expired versions")
	}
	if _, found := mc.GetMetadata(registryURL); found {
		t.Error("Expected cache miss for expired metadata")
	}
}
