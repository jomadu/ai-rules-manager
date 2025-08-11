package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMetadataManager_UpdateVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "arm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cacheManager := NewManager(tempDir)
	metadataManager := NewMetadataManager(cacheManager)

	registryType := "git"
	registryURL := "https://github.com/test/repo"
	rulesetName := "test-rules"
	versions := []string{"abc123", "def456"}
	mappings := map[string]string{
		"main":   "abc123",
		"v1.0.0": "def456",
	}

	// Update versions
	err = metadataManager.UpdateVersions(registryType, registryURL, rulesetName, versions, mappings)
	if err != nil {
		t.Fatalf("Failed to update versions: %v", err)
	}

	// Verify versions file was created
	cachePath, _ := cacheManager.GetCachePath(registryType, registryURL)
	versionsPath := filepath.Join(cachePath, "versions.json")
	if _, err := os.Stat(versionsPath); os.IsNotExist(err) {
		t.Fatalf("Versions file was not created")
	}

	// Retrieve versions
	retrievedVersions, retrievedMappings, err := metadataManager.GetVersions(registryType, registryURL, rulesetName)
	if err != nil {
		t.Fatalf("Failed to get versions: %v", err)
	}

	// Verify versions
	if len(retrievedVersions) != len(versions) {
		t.Errorf("Expected %d versions, got %d", len(versions), len(retrievedVersions))
	}

	// Verify mappings
	if len(retrievedMappings) != len(mappings) {
		t.Errorf("Expected %d mappings, got %d", len(mappings), len(retrievedMappings))
	}

	for key, expectedValue := range mappings {
		if actualValue, exists := retrievedMappings[key]; !exists || actualValue != expectedValue {
			t.Errorf("Expected mapping %s -> %s, got %s -> %s", key, expectedValue, key, actualValue)
		}
	}
}

func TestMetadataManager_UpdateMetadata(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "arm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cacheManager := NewManager(tempDir)
	metadataManager := NewMetadataManager(cacheManager)

	registryType := "git"
	registryURL := "https://github.com/test/repo"
	rulesetName := "test-rules"
	latestVersion := "abc123"
	fileCount := 5
	totalSize := int64(1024)

	// Update metadata
	err = metadataManager.UpdateMetadata(registryType, registryURL, rulesetName, latestVersion, fileCount, totalSize)
	if err != nil {
		t.Fatalf("Failed to update metadata: %v", err)
	}

	// Verify metadata file was created
	cachePath, _ := cacheManager.GetCachePath(registryType, registryURL)
	metadataPath := filepath.Join(cachePath, "metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Fatalf("Metadata file was not created")
	}

	// Retrieve metadata
	retrievedMetadata, err := metadataManager.GetMetadata(registryType, registryURL, rulesetName)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}

	// Verify metadata
	if retrievedMetadata.LatestVersion != latestVersion {
		t.Errorf("Expected latest version %s, got %s", latestVersion, retrievedMetadata.LatestVersion)
	}
	if retrievedMetadata.FileCount != fileCount {
		t.Errorf("Expected file count %d, got %d", fileCount, retrievedMetadata.FileCount)
	}
	if retrievedMetadata.TotalSizeBytes != totalSize {
		t.Errorf("Expected total size %d, got %d", totalSize, retrievedMetadata.TotalSizeBytes)
	}
}

func TestMetadataManager_IsVersionsCacheValid(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "arm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cacheManager := NewManager(tempDir)
	metadataManager := NewMetadataManager(cacheManager)

	registryType := "git"
	registryURL := "https://github.com/test/repo"
	rulesetName := "test-rules"
	versions := []string{"abc123"}

	// Update versions
	err = metadataManager.UpdateVersions(registryType, registryURL, rulesetName, versions, nil)
	if err != nil {
		t.Fatalf("Failed to update versions: %v", err)
	}

	// Check if cache is valid (should be valid immediately)
	valid, err := metadataManager.IsVersionsCacheValid(registryType, registryURL, time.Hour)
	if err != nil {
		t.Fatalf("Failed to check cache validity: %v", err)
	}
	if !valid {
		t.Errorf("Expected cache to be valid")
	}

	// Check if cache is invalid with very short TTL
	valid, err = metadataManager.IsVersionsCacheValid(registryType, registryURL, time.Nanosecond)
	if err != nil {
		t.Fatalf("Failed to check cache validity: %v", err)
	}
	if valid {
		t.Errorf("Expected cache to be invalid with short TTL")
	}
}

func TestMetadataManager_NonGitRegistry(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "arm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cacheManager := NewManager(tempDir)
	metadataManager := NewMetadataManager(cacheManager)

	registryType := "s3"
	registryURL := "my-bucket"
	rulesetName := "test-rules"
	versions := []string{"1.0.0", "1.1.0", "2.0.0"}

	// Update versions for non-git registry (no mappings)
	err = metadataManager.UpdateVersions(registryType, registryURL, rulesetName, versions, nil)
	if err != nil {
		t.Fatalf("Failed to update versions: %v", err)
	}

	// Retrieve versions
	retrievedVersions, retrievedMappings, err := metadataManager.GetVersions(registryType, registryURL, rulesetName)
	if err != nil {
		t.Fatalf("Failed to get versions: %v", err)
	}

	// Verify versions
	if len(retrievedVersions) != len(versions) {
		t.Errorf("Expected %d versions, got %d", len(versions), len(retrievedVersions))
	}

	// Verify no mappings for non-git registry
	if len(retrievedMappings) > 0 {
		t.Errorf("Expected no mappings for non-git registry, got %d", len(retrievedMappings))
	}
}
