package cache

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRulesetCacheManager_StoreAndGetRuleset(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewRulesetCacheManager(tempDir)

	registryURL := "https://example.com/registry"
	rulesetName := "test-ruleset"
	version := "1.0.0"
	files := map[string][]byte{
		"rule1.md": []byte("# Rule 1\nContent of rule 1"),
		"rule2.md": []byte("# Rule 2\nContent of rule 2"),
	}

	// Test StoreRuleset
	err := manager.StoreRuleset(registryURL, rulesetName, version, files)
	if err != nil {
		t.Fatalf("StoreRuleset failed: %v", err)
	}

	// Test GetRuleset
	retrievedFiles, err := manager.GetRuleset(registryURL, rulesetName, version)
	if err != nil {
		t.Fatalf("GetRuleset failed: %v", err)
	}

	// Verify files match
	if len(retrievedFiles) != len(files) {
		t.Fatalf("Expected %d files, got %d", len(files), len(retrievedFiles))
	}

	for filename, expectedContent := range files {
		actualContent, exists := retrievedFiles[filename]
		if !exists {
			t.Fatalf("File %s not found in retrieved files", filename)
		}
		if !bytes.Equal(actualContent, expectedContent) {
			t.Fatalf("File %s content mismatch. Expected: %s, Got: %s", filename, expectedContent, actualContent)
		}
	}
}

func TestRulesetCacheManager_Store(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewRulesetCacheManager(tempDir)

	registryURL := "https://example.com/registry"
	identifier := []string{"test-ruleset"}
	version := "1.0.0"
	files := map[string][]byte{
		"rule.md": []byte("# Test Rule"),
	}

	err := manager.Store(registryURL, identifier, version, files)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Verify using Get
	retrievedFiles, err := manager.Get(registryURL, identifier, version)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(retrievedFiles) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(retrievedFiles))
	}
}

func TestRulesetCacheManager_StoreEmptyIdentifier(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewRulesetCacheManager(tempDir)

	err := manager.Store("https://example.com", []string{}, "1.0.0", map[string][]byte{})
	if err == nil {
		t.Fatal("Expected error for empty identifier, got nil")
	}
}

func TestRulesetCacheManager_GetPath(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewRulesetCacheManager(tempDir)

	registryURL := "https://example.com/registry"
	identifier := []string{"test-ruleset"}
	version := "1.0.0"

	path, err := manager.GetPath(registryURL, identifier, version)
	if err != nil {
		t.Fatalf("GetPath failed: %v", err)
	}

	// Verify path structure: no repository subdirectory
	expectedPattern := filepath.Join(tempDir, "registries")
	if !filepath.HasPrefix(path, expectedPattern) {
		t.Fatalf("Path should start with %s, got %s", expectedPattern, path)
	}

	// Verify it ends with version
	if !strings.HasSuffix(path, version) {
		t.Fatalf("Path should end with version %s, got %s", version, path)
	}
}

func TestRulesetCacheManager_IsValid(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewRulesetCacheManager(tempDir)

	registryURL := "https://example.com/registry"
	rulesetName := "test-ruleset"
	version := "1.0.0"
	files := map[string][]byte{"rule.md": []byte("content")}

	// Store a ruleset to create registry index
	err := manager.StoreRuleset(registryURL, rulesetName, version, files)
	if err != nil {
		t.Fatalf("StoreRuleset failed: %v", err)
	}

	// Test with zero TTL (always valid)
	valid, err := manager.IsValid(registryURL, 0)
	if err != nil {
		t.Fatalf("IsValid failed: %v", err)
	}
	if !valid {
		t.Fatal("Expected valid=true for zero TTL")
	}

	// Test with long TTL (should be valid)
	valid, err = manager.IsValid(registryURL, 24*time.Hour)
	if err != nil {
		t.Fatalf("IsValid failed: %v", err)
	}
	if !valid {
		t.Fatal("Expected valid=true for long TTL")
	}
}

func TestRulesetCacheManager_GetNonExistentRuleset(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewRulesetCacheManager(tempDir)

	_, err := manager.GetRuleset("https://example.com", "nonexistent", "1.0.0")
	if err == nil {
		t.Fatal("Expected error for non-existent ruleset, got nil")
	}
}

func TestRulesetCacheManager_RegistryIndexCreation(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewRulesetCacheManager(tempDir)

	registryURL := "https://example.com/registry"
	rulesetName := "test-ruleset"
	version := "1.0.0"
	files := map[string][]byte{"rule.md": []byte("content")}

	err := manager.StoreRuleset(registryURL, rulesetName, version, files)
	if err != nil {
		t.Fatalf("StoreRuleset failed: %v", err)
	}

	// Verify registry index was created
	registryKey := GenerateRegistryKey("ruleset", registryURL)
	indexPath := filepath.Join(tempDir, "registries", registryKey, "index.json")

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatal("Registry index file was not created")
	}

	// Load and verify index content
	index, err := LoadRegistryIndex(filepath.Dir(indexPath))
	if err != nil {
		t.Fatalf("Failed to load registry index: %v", err)
	}

	if index.NormalizedRegistryType != "ruleset" {
		t.Fatalf("Expected registry type 'ruleset', got '%s'", index.NormalizedRegistryType)
	}

	rulesetKey := GenerateRulesetKey(rulesetName)
	rulesetCache, exists := index.Rulesets[rulesetKey]
	if !exists {
		t.Fatal("Ruleset not found in index")
	}

	if rulesetCache.NormalizedRulesetName != rulesetName {
		t.Fatalf("Expected ruleset name '%s', got '%s'", rulesetName, rulesetCache.NormalizedRulesetName)
	}

	if _, exists := rulesetCache.Versions[version]; !exists {
		t.Fatalf("Version '%s' not found in ruleset cache", version)
	}
}
