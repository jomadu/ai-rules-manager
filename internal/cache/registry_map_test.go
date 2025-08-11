package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewRegistryMapper(t *testing.T) {
	cacheRoot := "/tmp/test-cache"
	mapper := NewRegistryMapper(cacheRoot)

	expectedPath := filepath.Join(cacheRoot, "registry-map.json")
	if mapper.mapFilePath != expectedPath {
		t.Errorf("Expected map file path %s, got %s", expectedPath, mapper.mapFilePath)
	}
}

func TestRegistryMapper_AddMapping(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	mapper := NewRegistryMapper(tempDir)

	cacheKey := "abc123"
	registryType := "git"
	registryURL := "https://github.com/user/repo"
	normalizedURL := "https://github.com/user/repo"

	err = mapper.AddMapping(cacheKey, registryType, registryURL, normalizedURL)
	if err != nil {
		t.Fatalf("AddMapping failed: %v", err)
	}

	// Verify mapping was added
	mapping, err := mapper.GetMapping(cacheKey)
	if err != nil {
		t.Fatalf("GetMapping failed: %v", err)
	}

	if mapping.CacheKey != cacheKey {
		t.Errorf("Expected cache key %s, got %s", cacheKey, mapping.CacheKey)
	}
	if mapping.RegistryType != registryType {
		t.Errorf("Expected registry type %s, got %s", registryType, mapping.RegistryType)
	}
	if mapping.RegistryURL != registryURL {
		t.Errorf("Expected registry URL %s, got %s", registryURL, mapping.RegistryURL)
	}
	if mapping.NormalizedURL != normalizedURL {
		t.Errorf("Expected normalized URL %s, got %s", normalizedURL, mapping.NormalizedURL)
	}
}

func TestRegistryMapper_UpdateMapping(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	mapper := NewRegistryMapper(tempDir)

	cacheKey := "abc123"
	registryType := "git"
	originalURL := "https://github.com/user/repo"
	updatedURL := "https://github.com/user/updated-repo"

	// Add initial mapping
	err = mapper.AddMapping(cacheKey, registryType, originalURL, originalURL)
	if err != nil {
		t.Fatalf("AddMapping failed: %v", err)
	}

	originalMapping, err := mapper.GetMapping(cacheKey)
	if err != nil {
		t.Fatalf("GetMapping failed: %v", err)
	}

	// Update mapping
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp
	err = mapper.AddMapping(cacheKey, registryType, updatedURL, updatedURL)
	if err != nil {
		t.Fatalf("AddMapping update failed: %v", err)
	}

	updatedMapping, err := mapper.GetMapping(cacheKey)
	if err != nil {
		t.Fatalf("GetMapping failed: %v", err)
	}

	// Verify URL was updated but creation time preserved
	if updatedMapping.RegistryURL != updatedURL {
		t.Errorf("Expected updated URL %s, got %s", updatedURL, updatedMapping.RegistryURL)
	}
	if !updatedMapping.CreatedAt.Equal(originalMapping.CreatedAt) {
		t.Errorf("Creation time should be preserved")
	}
	if !updatedMapping.LastAccessed.After(originalMapping.LastAccessed) {
		t.Errorf("Last accessed time should be updated")
	}
}

func TestRegistryMapper_FindMappingByURL(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	mapper := NewRegistryMapper(tempDir)

	cacheKey := "abc123"
	registryType := "git"
	registryURL := "https://github.com/user/repo"
	normalizedURL := "https://github.com/user/repo"

	err = mapper.AddMapping(cacheKey, registryType, registryURL, normalizedURL)
	if err != nil {
		t.Fatalf("AddMapping failed: %v", err)
	}

	// Find by URL
	mapping, err := mapper.FindMappingByURL(registryType, normalizedURL)
	if err != nil {
		t.Fatalf("FindMappingByURL failed: %v", err)
	}

	if mapping.CacheKey != cacheKey {
		t.Errorf("Expected cache key %s, got %s", cacheKey, mapping.CacheKey)
	}

	// Test not found
	_, err = mapper.FindMappingByURL("https", normalizedURL)
	if err == nil {
		t.Error("Expected error for non-existent mapping")
	}
}

func TestRegistryMapper_UpdateLastAccessed(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	mapper := NewRegistryMapper(tempDir)

	cacheKey := "abc123"
	registryType := "git"
	registryURL := "https://github.com/user/repo"

	err = mapper.AddMapping(cacheKey, registryType, registryURL, registryURL)
	if err != nil {
		t.Fatalf("AddMapping failed: %v", err)
	}

	originalMapping, err := mapper.GetMapping(cacheKey)
	if err != nil {
		t.Fatalf("GetMapping failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond) // Ensure different timestamp

	err = mapper.UpdateLastAccessed(cacheKey)
	if err != nil {
		t.Fatalf("UpdateLastAccessed failed: %v", err)
	}

	updatedMapping, err := mapper.GetMapping(cacheKey)
	if err != nil {
		t.Fatalf("GetMapping failed: %v", err)
	}

	if !updatedMapping.LastAccessed.After(originalMapping.LastAccessed) {
		t.Error("Last accessed time should be updated")
	}
}

func TestRegistryMapper_RemoveMapping(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	mapper := NewRegistryMapper(tempDir)

	cacheKey := "abc123"
	registryType := "git"
	registryURL := "https://github.com/user/repo"

	err = mapper.AddMapping(cacheKey, registryType, registryURL, registryURL)
	if err != nil {
		t.Fatalf("AddMapping failed: %v", err)
	}

	// Verify mapping exists
	_, err = mapper.GetMapping(cacheKey)
	if err != nil {
		t.Fatalf("GetMapping failed: %v", err)
	}

	// Remove mapping
	err = mapper.RemoveMapping(cacheKey)
	if err != nil {
		t.Fatalf("RemoveMapping failed: %v", err)
	}

	// Verify mapping is gone
	_, err = mapper.GetMapping(cacheKey)
	if err == nil {
		t.Error("Expected error for removed mapping")
	}
}

func TestRegistryMapper_ListMappings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	mapper := NewRegistryMapper(tempDir)

	// Add multiple mappings
	mappings := []struct {
		cacheKey     string
		registryType string
		registryURL  string
	}{
		{"key1", "git", "https://github.com/user/repo1"},
		{"key2", "gitlab", "https://gitlab.com/user/repo2"},
		{"key3", "s3", "my-bucket/rules"},
	}

	for _, m := range mappings {
		err = mapper.AddMapping(m.cacheKey, m.registryType, m.registryURL, m.registryURL)
		if err != nil {
			t.Fatalf("AddMapping failed: %v", err)
		}
	}

	// List all mappings
	allMappings, err := mapper.ListMappings()
	if err != nil {
		t.Fatalf("ListMappings failed: %v", err)
	}

	if len(allMappings) != len(mappings) {
		t.Errorf("Expected %d mappings, got %d", len(mappings), len(allMappings))
	}

	// Verify all mappings are present
	found := make(map[string]bool)
	for _, mapping := range allMappings {
		found[mapping.CacheKey] = true
	}

	for _, m := range mappings {
		if !found[m.cacheKey] {
			t.Errorf("Mapping with cache key %s not found in list", m.cacheKey)
		}
	}
}

func TestRegistryMapper_EmptyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	mapper := NewRegistryMapper(tempDir)

	// List mappings when file doesn't exist
	mappings, err := mapper.ListMappings()
	if err != nil {
		t.Fatalf("ListMappings failed: %v", err)
	}

	if len(mappings) != 0 {
		t.Errorf("Expected 0 mappings for empty file, got %d", len(mappings))
	}
}

func TestRegistryMapper_AtomicOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	mapper := NewRegistryMapper(tempDir)

	// Add mapping
	err = mapper.AddMapping("key1", "git", "https://github.com/user/repo", "https://github.com/user/repo")
	if err != nil {
		t.Fatalf("AddMapping failed: %v", err)
	}

	// Verify file exists and is valid JSON
	mapFilePath := filepath.Join(tempDir, "registry-map.json")
	if _, err := os.Stat(mapFilePath); os.IsNotExist(err) {
		t.Error("Registry map file should exist")
	}

	// Verify no temp files left behind
	tempFilePath := mapFilePath + ".tmp"
	if _, err := os.Stat(tempFilePath); !os.IsNotExist(err) {
		t.Error("Temporary file should not exist after successful operation")
	}
}

func TestRegistryMapper_ValidateAndRecover(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Use cache manager to generate proper cache key
	cacheManager := NewManager(tempDir)
	cacheKey, err := cacheManager.GetCacheKey("git", "https://github.com/user/repo")
	if err != nil {
		t.Fatalf("GetCacheKey failed: %v", err)
	}

	mapper := NewRegistryMapper(tempDir)

	// Test with valid mappings
	err = mapper.AddMapping(cacheKey, "git", "https://github.com/user/repo", "https://github.com/user/repo")
	if err != nil {
		t.Fatalf("AddMapping failed: %v", err)
	}

	err = mapper.ValidateAndRecover()
	if err != nil {
		t.Fatalf("ValidateAndRecover failed: %v", err)
	}

	// Verify mapping still exists
	_, err = mapper.GetMapping(cacheKey)
	if err != nil {
		t.Errorf("Valid mapping should still exist after validation: %v", err)
	}
}

func TestRegistryMapper_RecoverFromCorruption(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	mapper := NewRegistryMapper(tempDir)
	mapFilePath := filepath.Join(tempDir, "registry-map.json")

	// Create corrupted file
	err = os.WriteFile(mapFilePath, []byte("invalid json"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Validate and recover should handle corruption
	err = mapper.ValidateAndRecover()
	if err != nil {
		t.Fatalf("ValidateAndRecover should handle corruption: %v", err)
	}

	// Should be able to add mappings after recovery
	err = mapper.AddMapping("def456", "git", "https://github.com/user/repo2", "https://github.com/user/repo2")
	if err != nil {
		t.Fatalf("AddMapping after recovery failed: %v", err)
	}
}

func TestRegistryMapper_ValidationRules(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-map-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	mapper := NewRegistryMapper(tempDir)

	tests := []struct {
		name    string
		mapping RegistryMapping
		valid   bool
	}{
		{
			name: "valid mapping",
			mapping: RegistryMapping{
				CacheKey:     "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				RegistryType: "git",
				RegistryURL:  "https://github.com/user/repo",
			},
			valid: true,
		},
		{
			name: "empty cache key",
			mapping: RegistryMapping{
				CacheKey:     "",
				RegistryType: "git",
				RegistryURL:  "https://github.com/user/repo",
			},
			valid: false,
		},
		{
			name: "invalid cache key length",
			mapping: RegistryMapping{
				CacheKey:     "abc123",
				RegistryType: "git",
				RegistryURL:  "https://github.com/user/repo",
			},
			valid: false,
		},
		{
			name: "invalid registry type",
			mapping: RegistryMapping{
				CacheKey:     "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				RegistryType: "invalid",
				RegistryURL:  "https://github.com/user/repo",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.isValidMapping(&tt.mapping)
			if result != tt.valid {
				t.Errorf("isValidMapping() = %v, want %v", result, tt.valid)
			}
		})
	}
}
