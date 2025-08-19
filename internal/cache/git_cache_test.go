package cache

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewGitCacheManager(t *testing.T) {
	cacheRoot := "/tmp/test-cache"
	manager := NewGitCacheManager(cacheRoot)

	if manager == nil {
		t.Fatal("NewGitCacheManager() returned nil")
	}

	if manager.cacheRoot != cacheRoot {
		t.Errorf("Expected cache root %s, got %s", cacheRoot, manager.cacheRoot)
	}
}

func TestGitCacheManager_GetPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "git-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewGitCacheManager(tempDir)
	registryURL := "https://github.com/user/repo"
	patterns := []string{"*.md", "rules/**"}
	commitHash := "abc123def456"

	path, err := manager.GetPath(registryURL, patterns, commitHash)
	if err != nil {
		t.Fatalf("GetPath() failed: %v", err)
	}

	registryKey := GenerateRegistryKey("git", registryURL)
	patternsKey := GeneratePatternsKey(patterns)
	expectedPath := filepath.Join(tempDir, "registries", registryKey, "rulesets", patternsKey, commitHash)

	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestGitCacheManager_StoreAndGetRuleset(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "git-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewGitCacheManager(tempDir)
	registryURL := "https://github.com/user/repo"
	patterns := []string{"*.md"}
	commitHash := "abc123def456"

	files := map[string][]byte{
		"README.md":     []byte("# Test README"),
		"rules/test.md": []byte("# Test Rule"),
	}

	// Test StoreRuleset
	err = manager.StoreRuleset(registryURL, patterns, commitHash, files)
	if err != nil {
		t.Fatalf("StoreRuleset() failed: %v", err)
	}

	// Verify files were stored
	path, _ := manager.GetPath(registryURL, patterns, commitHash)
	if _, err := os.Stat(filepath.Join(path, "README.md")); os.IsNotExist(err) {
		t.Errorf("README.md was not stored")
	}
	if _, err := os.Stat(filepath.Join(path, "rules", "test.md")); os.IsNotExist(err) {
		t.Errorf("rules/test.md was not stored")
	}

	// Test GetRuleset
	retrievedFiles, err := manager.GetRuleset(registryURL, patterns, commitHash)
	if err != nil {
		t.Fatalf("GetRuleset() failed: %v", err)
	}

	if len(retrievedFiles) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(retrievedFiles))
	}

	for filename, expectedContent := range files {
		if retrievedContent, exists := retrievedFiles[filename]; !exists {
			t.Errorf("File %s not found in retrieved files", filename)
		} else if !bytes.Equal(retrievedContent, expectedContent) {
			t.Errorf("File %s content mismatch: expected %s, got %s", filename, expectedContent, retrievedContent)
		}
	}
}

func TestGitCacheManager_GetRepositoryPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "git-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewGitCacheManager(tempDir)
	registryURL := "https://github.com/user/repo"

	path, err := manager.GetRepositoryPath(registryURL)
	if err != nil {
		t.Fatalf("GetRepositoryPath() failed: %v", err)
	}

	registryKey := GenerateRegistryKey("git", registryURL)
	expectedPath := filepath.Join(tempDir, "registries", registryKey, "repository")

	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestGitCacheManager_IsValid(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "git-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewGitCacheManager(tempDir)
	registryURL := "https://github.com/user/repo"

	// Test with no TTL (always valid)
	valid, err := manager.IsValid(registryURL, 0)
	if err != nil {
		t.Fatalf("IsValid() failed: %v", err)
	}
	if !valid {
		t.Errorf("Expected cache to be valid with no TTL")
	}

	// Test with non-existent cache
	valid, err = manager.IsValid(registryURL, time.Hour)
	if err != nil {
		t.Fatalf("IsValid() failed: %v", err)
	}
	if valid {
		t.Errorf("Expected cache to be invalid when it doesn't exist")
	}

	// Create a cache entry
	patterns := []string{"*.md"}
	commitHash := "abc123"
	files := map[string][]byte{"test.md": []byte("test")}

	err = manager.StoreRuleset(registryURL, patterns, commitHash, files)
	if err != nil {
		t.Fatalf("StoreRuleset() failed: %v", err)
	}

	// Test with valid TTL
	valid, err = manager.IsValid(registryURL, time.Hour)
	if err != nil {
		t.Fatalf("IsValid() failed: %v", err)
	}
	if !valid {
		t.Errorf("Expected cache to be valid within TTL")
	}
}

func TestGitCacheManager_Store(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "git-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewGitCacheManager(tempDir)
	registryURL := "https://github.com/user/repo"
	patterns := []string{"*.md"}
	commitHash := "abc123def456"

	files := map[string][]byte{
		"test.md": []byte("# Test"),
	}

	// Test Store (should delegate to StoreRuleset)
	err = manager.Store(registryURL, patterns, commitHash, files)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Verify it was stored
	retrievedFiles, err := manager.Get(registryURL, patterns, commitHash)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if len(retrievedFiles) != 1 {
		t.Errorf("Expected 1 file, got %d", len(retrievedFiles))
	}
}

func TestGitCacheManager_EmptyPatterns(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "git-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewGitCacheManager(tempDir)
	registryURL := "https://github.com/user/repo"
	patterns := []string{} // Empty patterns
	commitHash := "abc123def456"

	files := map[string][]byte{
		"test.md": []byte("# Test"),
	}

	// Test with empty patterns (should use __EMPTY__ key)
	err = manager.StoreRuleset(registryURL, patterns, commitHash, files)
	if err != nil {
		t.Fatalf("StoreRuleset() with empty patterns failed: %v", err)
	}

	retrievedFiles, err := manager.GetRuleset(registryURL, patterns, commitHash)
	if err != nil {
		t.Fatalf("GetRuleset() with empty patterns failed: %v", err)
	}

	if len(retrievedFiles) != 1 {
		t.Errorf("Expected 1 file, got %d", len(retrievedFiles))
	}
}

func TestGitCacheManager_GetNonExistentRuleset(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "git-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewGitCacheManager(tempDir)
	registryURL := "https://github.com/user/repo"
	patterns := []string{"*.md"}
	commitHash := "nonexistent"

	_, err = manager.GetRuleset(registryURL, patterns, commitHash)
	if err == nil {
		t.Errorf("Expected error when getting non-existent ruleset")
	}
}
