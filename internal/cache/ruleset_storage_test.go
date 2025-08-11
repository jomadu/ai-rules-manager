package cache

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRulesetStorage_StoreAndRetrieveFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "arm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cacheManager := NewManager(tempDir)
	rulesetStorage := NewRulesetStorage(cacheManager)

	registryType := "git"
	registryURL := "https://github.com/test/repo"
	rulesetName := "test-rules"
	version := "abc123"

	// Test files
	files := map[string][]byte{
		"rule1.md":        []byte("# Rule 1\nContent of rule 1"),
		"subdir/rule2.md": []byte("# Rule 2\nContent of rule 2"),
		"config.json":     []byte(`{"setting": "value"}`),
	}

	// Store files
	err = rulesetStorage.StoreRulesetFiles(registryType, registryURL, rulesetName, version, files, nil)
	if err != nil {
		t.Fatalf("Failed to store ruleset files: %v", err)
	}

	// Verify files were stored
	rulesetPath, err := rulesetStorage.GetRulesetVersionPath(registryType, registryURL, rulesetName, version)
	if err != nil {
		t.Fatalf("Failed to get ruleset path: %v", err)
	}

	for filename := range files {
		filePath := filepath.Join(rulesetPath, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File %s was not stored", filename)
		}
	}

	// Retrieve files
	retrievedFiles, err := rulesetStorage.GetRulesetFiles(registryType, registryURL, rulesetName, version, nil)
	if err != nil {
		t.Fatalf("Failed to retrieve ruleset files: %v", err)
	}

	// Verify retrieved files match stored files
	if len(retrievedFiles) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(retrievedFiles))
	}

	for filename, expectedContent := range files {
		if actualContent, exists := retrievedFiles[filename]; !exists {
			t.Errorf("File %s was not retrieved", filename)
		} else if !bytes.Equal(actualContent, expectedContent) {
			t.Errorf("File %s content mismatch. Expected: %s, Got: %s", filename, string(expectedContent), string(actualContent))
		}
	}
}

func TestRulesetStorage_ListVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "arm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cacheManager := NewManager(tempDir)
	rulesetStorage := NewRulesetStorage(cacheManager)

	registryType := "git"
	registryURL := "https://github.com/test/repo"
	rulesetName := "test-rules"

	versions := []string{"abc123", "def456", "ghi789"}
	testFile := map[string][]byte{"test.md": []byte("test content")}

	// Store multiple versions
	for _, version := range versions {
		err = rulesetStorage.StoreRulesetFiles(registryType, registryURL, rulesetName, version, testFile, nil)
		if err != nil {
			t.Fatalf("Failed to store version %s: %v", version, err)
		}
	}

	// List versions
	listedVersions, err := rulesetStorage.ListRulesetVersions(registryType, registryURL, rulesetName, nil)
	if err != nil {
		t.Fatalf("Failed to list versions: %v", err)
	}

	// Verify all versions are listed
	if len(listedVersions) != len(versions) {
		t.Errorf("Expected %d versions, got %d", len(versions), len(listedVersions))
	}

	versionMap := make(map[string]bool)
	for _, version := range listedVersions {
		versionMap[version] = true
	}

	for _, expectedVersion := range versions {
		if !versionMap[expectedVersion] {
			t.Errorf("Version %s was not listed", expectedVersion)
		}
	}
}

func TestRulesetStorage_GetRulesetStats(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "arm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cacheManager := NewManager(tempDir)
	rulesetStorage := NewRulesetStorage(cacheManager)

	registryType := "git"
	registryURL := "https://github.com/test/repo"
	rulesetName := "test-rules"
	version := "abc123"

	// Test files with known sizes
	files := map[string][]byte{
		"rule1.md": []byte("12345"),      // 5 bytes
		"rule2.md": []byte("1234567890"), // 10 bytes
		"rule3.md": []byte("123"),        // 3 bytes
	}

	// Store files
	err = rulesetStorage.StoreRulesetFiles(registryType, registryURL, rulesetName, version, files, nil)
	if err != nil {
		t.Fatalf("Failed to store ruleset files: %v", err)
	}

	// Get stats
	fileCount, totalSize, err := rulesetStorage.GetRulesetStats(registryType, registryURL, rulesetName, version, nil)
	if err != nil {
		t.Fatalf("Failed to get ruleset stats: %v", err)
	}

	// Verify stats
	expectedFileCount := len(files)
	expectedTotalSize := int64(5 + 10 + 3) // Sum of file sizes

	if fileCount != expectedFileCount {
		t.Errorf("Expected file count %d, got %d", expectedFileCount, fileCount)
	}

	if totalSize != expectedTotalSize {
		t.Errorf("Expected total size %d, got %d", expectedTotalSize, totalSize)
	}
}

func TestRulesetStorage_RemoveRulesetVersion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "arm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cacheManager := NewManager(tempDir)
	rulesetStorage := NewRulesetStorage(cacheManager)

	registryType := "git"
	registryURL := "https://github.com/test/repo"
	rulesetName := "test-rules"
	version := "abc123"

	testFile := map[string][]byte{"test.md": []byte("test content")}

	// Store version
	err = rulesetStorage.StoreRulesetFiles(registryType, registryURL, rulesetName, version, testFile, nil)
	if err != nil {
		t.Fatalf("Failed to store ruleset files: %v", err)
	}

	// Verify version exists
	versions, err := rulesetStorage.ListRulesetVersions(registryType, registryURL, rulesetName, nil)
	if err != nil {
		t.Fatalf("Failed to list versions: %v", err)
	}
	if len(versions) != 1 || versions[0] != version {
		t.Fatalf("Expected version %s to exist", version)
	}

	// Remove version
	err = rulesetStorage.RemoveRulesetVersion(registryType, registryURL, rulesetName, version, nil)
	if err != nil {
		t.Fatalf("Failed to remove version: %v", err)
	}

	// Verify version was removed
	versions, err = rulesetStorage.ListRulesetVersions(registryType, registryURL, rulesetName, nil)
	if err != nil {
		t.Fatalf("Failed to list versions after removal: %v", err)
	}
	if len(versions) != 0 {
		t.Errorf("Expected no versions after removal, got %d", len(versions))
	}
}

func TestRulesetStorage_CleanupUnreferencedVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "arm-cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cacheManager := NewManager(tempDir)
	rulesetStorage := NewRulesetStorage(cacheManager)

	registryType := "git"
	registryURL := "https://github.com/test/repo"
	rulesetName := "test-rules"

	allVersions := []string{"abc123", "def456", "ghi789"}
	referencedVersions := []string{"abc123", "ghi789"} // def456 should be removed
	testFile := map[string][]byte{"test.md": []byte("test content")}

	// Store all versions
	for _, version := range allVersions {
		err = rulesetStorage.StoreRulesetFiles(registryType, registryURL, rulesetName, version, testFile, nil)
		if err != nil {
			t.Fatalf("Failed to store version %s: %v", version, err)
		}
	}

	// Cleanup unreferenced versions
	err = rulesetStorage.CleanupUnreferencedVersions(registryType, registryURL, rulesetName, referencedVersions, nil)
	if err != nil {
		t.Fatalf("Failed to cleanup unreferenced versions: %v", err)
	}

	// Verify only referenced versions remain
	remainingVersions, err := rulesetStorage.ListRulesetVersions(registryType, registryURL, rulesetName, nil)
	if err != nil {
		t.Fatalf("Failed to list remaining versions: %v", err)
	}

	if len(remainingVersions) != len(referencedVersions) {
		t.Errorf("Expected %d remaining versions, got %d", len(referencedVersions), len(remainingVersions))
	}

	remainingMap := make(map[string]bool)
	for _, version := range remainingVersions {
		remainingMap[version] = true
	}

	for _, referencedVersion := range referencedVersions {
		if !remainingMap[referencedVersion] {
			t.Errorf("Referenced version %s was removed", referencedVersion)
		}
	}

	// Verify unreferenced version was removed
	if remainingMap["def456"] {
		t.Errorf("Unreferenced version def456 was not removed")
	}
}
