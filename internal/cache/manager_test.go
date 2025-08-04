package cache

import (
	"os"
	"strings"
	"testing"
)

func TestManagerFallback(t *testing.T) {
	// Test manager creation when home directory is not accessible
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", "/nonexistent")
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Manager creation should not fail: %v", err)
	}

	// Manager should work in fallback mode
	if manager.storage != nil {
		t.Error("Expected nil storage in fallback mode")
	}

	// Operations should gracefully handle missing cache
	if _, found := manager.GetPackage("test", "test", "1.0.0"); found {
		t.Error("Expected cache miss in fallback mode")
	}

	if versions, found := manager.GetVersions("test"); found || versions != nil {
		t.Error("Expected no versions in fallback mode")
	}

	// Store operations should not error
	if err := manager.StoreVersions("test", []string{"1.0.0"}); err != nil {
		t.Errorf("StoreVersions should not error in fallback mode: %v", err)
	}
}

func TestManagerIntegration(t *testing.T) {
	// Set up temporary home directory
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpHome)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	registryURL := "https://registry.armjs.org/"
	ruleset := "typescript-rules"
	version := "1.0.0"

	// Test package operations
	testData := strings.NewReader("test package data")
	path, err := manager.StorePackage(registryURL, ruleset, version, testData)
	if err != nil {
		t.Fatalf("Failed to store package: %v", err)
	}

	cachedPath, found := manager.GetPackage(registryURL, ruleset, version)
	if !found {
		t.Error("Expected to find cached package")
	}
	if cachedPath != path {
		t.Errorf("Cached path %q != stored path %q", cachedPath, path)
	}

	// Test metadata operations
	versions := []string{"1.0.0", "1.1.0"}
	if err := manager.StoreVersions(registryURL, versions); err != nil {
		t.Fatalf("Failed to store versions: %v", err)
	}

	cachedVersions, found := manager.GetVersions(registryURL)
	if !found {
		t.Error("Expected to find cached versions")
	}
	if len(cachedVersions) != len(versions) {
		t.Errorf("Cached versions length %d != stored length %d", len(cachedVersions), len(versions))
	}

	metadata := map[string]interface{}{"test": "data"}
	if err := manager.StoreMetadata(registryURL, metadata); err != nil {
		t.Fatalf("Failed to store metadata: %v", err)
	}

	cachedMetadata, found := manager.GetMetadata(registryURL)
	if !found {
		t.Error("Expected to find cached metadata")
	}
	if cachedMetadata["test"] != metadata["test"] {
		t.Errorf("Cached metadata %v != stored %v", cachedMetadata, metadata)
	}
}
