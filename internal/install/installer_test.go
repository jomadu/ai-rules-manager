package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/max-dunn/ai-rules-manager/internal/config"
)

func TestInstaller_Install(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-install-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test configuration
	cfg := &config.Config{
		Channels: map[string]config.ChannelConfig{
			"cursor": {
				Directories: []string{filepath.Join(tempDir, ".cursor", "rules")},
			},
			"q": {
				Directories: []string{filepath.Join(tempDir, ".aws", "amazonq", "rules")},
			},
		},
	}

	installer := New(cfg)

	// Create test source files that simulate what would come from a real download
	// The installer expects files with full paths that include temp directory names
	sourceDir, err := os.MkdirTemp("", "arm-install-")
	if err != nil {
		t.Fatalf("Failed to create source temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(sourceDir) }()

	testFile1 := filepath.Join(sourceDir, "rule1.md")
	testFile2 := filepath.Join(sourceDir, "rule2.mdc")

	if err := os.WriteFile(testFile1, []byte("# Test Rule 1"), 0o644); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("# Test Rule 2"), 0o644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	// Test installation
	req := &InstallRequest{
		Registry:    "test-registry",
		Ruleset:     "test-ruleset",
		Version:     "1.0.0",
		SourceFiles: []string{testFile1, testFile2},
		Channels:    []string{"cursor"},
	}

	result, err := installer.Install(req)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify result
	if result.Registry != "test-registry" {
		t.Errorf("Expected registry 'test-registry', got '%s'", result.Registry)
	}
	if result.Ruleset != "test-ruleset" {
		t.Errorf("Expected ruleset 'test-ruleset', got '%s'", result.Ruleset)
	}
	if result.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", result.Version)
	}
	if result.FilesCount != 2 {
		t.Errorf("Expected 2 files, got %d", result.FilesCount)
	}

	// Verify files were installed
	expectedPath := filepath.Join(tempDir, ".cursor", "rules", "arm", "test-registry", "test-ruleset", "1.0.0")
	if _, err := os.Stat(filepath.Join(expectedPath, "rule1.md")); err != nil {
		t.Errorf("rule1.md not found in expected location: %v", err)
	}
	if _, err := os.Stat(filepath.Join(expectedPath, "rule2.mdc")); err != nil {
		t.Errorf("rule2.mdc not found in expected location: %v", err)
	}
}

func TestInstaller_Uninstall(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-uninstall-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test configuration
	cfg := &config.Config{
		Channels: map[string]config.ChannelConfig{
			"cursor": {
				Directories: []string{filepath.Join(tempDir, ".cursor", "rules")},
			},
		},
	}

	installer := New(cfg)

	// Create test installation
	rulesetPath := filepath.Join(tempDir, ".cursor", "rules", "arm", "test-registry", "test-ruleset")
	if err := os.MkdirAll(filepath.Join(rulesetPath, "1.0.0"), 0o755); err != nil {
		t.Fatalf("Failed to create test installation: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesetPath, "1.0.0", "test.md"), []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test uninstall
	err = installer.Uninstall("test-registry", "test-ruleset", []string{"cursor"})
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	// Verify ruleset was removed
	if _, err := os.Stat(rulesetPath); !os.IsNotExist(err) {
		t.Errorf("Ruleset directory still exists after uninstall")
	}
}

func TestInstaller_ListInstalled(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-list-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test configuration
	cfg := &config.Config{
		Channels: map[string]config.ChannelConfig{
			"cursor": {
				Directories: []string{filepath.Join(tempDir, ".cursor", "rules")},
			},
		},
	}

	installer := New(cfg)

	// Create test installations
	armDir := filepath.Join(tempDir, ".cursor", "rules", "arm")

	// Registry 1 with 2 rulesets
	reg1Path := filepath.Join(armDir, "registry1")
	if err := os.MkdirAll(filepath.Join(reg1Path, "ruleset1", "1.0.0"), 0o755); err != nil {
		t.Fatalf("Failed to create test installation: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(reg1Path, "ruleset2", "2.0.0"), 0o755); err != nil {
		t.Fatalf("Failed to create test installation: %v", err)
	}

	// Registry 2 with 1 ruleset
	reg2Path := filepath.Join(armDir, "registry2")
	if err := os.MkdirAll(filepath.Join(reg2Path, "ruleset3", "1.5.0"), 0o755); err != nil {
		t.Fatalf("Failed to create test installation: %v", err)
	}

	// Test list installed
	result, err := installer.ListInstalled([]string{"cursor"})
	if err != nil {
		t.Fatalf("ListInstalled failed: %v", err)
	}

	// Verify results
	cursorRulesets, exists := result["cursor"]
	if !exists {
		t.Fatalf("cursor channel not found in results")
	}

	reg1Rulesets, exists := cursorRulesets["registry1"]
	if !exists {
		t.Fatalf("registry1 not found in cursor channel")
	}
	if len(reg1Rulesets) != 2 {
		t.Errorf("Expected 2 rulesets in registry1, got %d", len(reg1Rulesets))
	}

	reg2Rulesets, exists := cursorRulesets["registry2"]
	if !exists {
		t.Fatalf("registry2 not found in cursor channel")
	}
	if len(reg2Rulesets) != 1 {
		t.Errorf("Expected 1 ruleset in registry2, got %d", len(reg2Rulesets))
	}
}

func TestExpandPath(t *testing.T) {
	// Test tilde expansion
	homeDir, _ := os.UserHomeDir()
	result := expandPath("~/test/path")
	expected := filepath.Join(homeDir, "test", "path")
	if result != expected {
		t.Errorf("Tilde expansion failed: expected %s, got %s", expected, result)
	}

	// Test environment variable expansion
	_ = os.Setenv("TEST_VAR", "test_value")
	result = expandPath("$TEST_VAR/path")
	expected = "test_value/path"
	if result != expected {
		t.Errorf("Environment variable expansion failed: expected %s, got %s", expected, result)
	}
}

func TestInstaller_CleanupPreviousVersion(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-cleanup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cfg := &config.Config{}
	installer := New(cfg)

	// Create ruleset directory with multiple versions
	rulesetDir := filepath.Join(tempDir, "test-ruleset")
	if err := os.MkdirAll(filepath.Join(rulesetDir, "1.0.0"), 0o755); err != nil {
		t.Fatalf("Failed to create version 1.0.0: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(rulesetDir, "1.1.0"), 0o755); err != nil {
		t.Fatalf("Failed to create version 1.1.0: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(rulesetDir, "2.0.0"), 0o755); err != nil {
		t.Fatalf("Failed to create version 2.0.0: %v", err)
	}

	// Add test files to each version
	for _, version := range []string{"1.0.0", "1.1.0", "2.0.0"} {
		testFile := filepath.Join(rulesetDir, version, "test.md")
		if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
			t.Fatalf("Failed to create test file for version %s: %v", version, err)
		}
	}

	// Cleanup previous versions, keeping only 2.0.0
	installer.cleanupPreviousVersion(rulesetDir, "2.0.0")

	// Verify only 2.0.0 remains
	entries, err := os.ReadDir(rulesetDir)
	if err != nil {
		t.Fatalf("Failed to read ruleset directory: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 version directory, got %d", len(entries))
	}

	if entries[0].Name() != "2.0.0" {
		t.Errorf("Expected version 2.0.0 to remain, got %s", entries[0].Name())
	}
}

func TestInstaller_LockFileManagement(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-lock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory for lock file operations
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)

	// Create test configuration
	cfg := &config.Config{
		Registries: map[string]string{
			"test-registry": "https://github.com/test/repo",
		},
		RegistryConfigs: map[string]map[string]string{
			"test-registry": {
				"type":   "git",
				"region": "us-east-1",
			},
		},
	}

	installer := New(cfg)

	// Test updating lock file
	req := &InstallRequest{
		Registry: "test-registry",
		Ruleset:  "test-ruleset",
		Version:  "1.0.0",
	}
	err = installer.updateLockFile(req, "abc123def")
	if err != nil {
		t.Fatalf("Failed to update lock file: %v", err)
	}

	// Verify lock file was created and contains entry
	lockFile, err := installer.GetLockFile()
	if err != nil {
		t.Fatalf("Failed to load lock file: %v", err)
	}

	entry, exists := lockFile.Rulesets["test-registry"]["test-ruleset"]
	if !exists {
		t.Fatalf("Lock file entry not found")
	}

	if entry.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", entry.Version)
	}
	if entry.Type != "git" {
		t.Errorf("Expected type git, got %s", entry.Type)
	}
	if entry.Registry != "https://github.com/test/repo" {
		t.Errorf("Expected registry URL, got %s", entry.Registry)
	}

	// Test removing lock entry
	err = installer.removeLockEntry("test-registry", "test-ruleset")
	if err != nil {
		t.Fatalf("Failed to remove lock entry: %v", err)
	}

	// Verify entry was removed
	lockFile, err = installer.GetLockFile()
	if err != nil {
		t.Fatalf("Failed to load lock file: %v", err)
	}

	if len(lockFile.Rulesets) != 0 {
		t.Errorf("Expected empty lock file, got %d registries", len(lockFile.Rulesets))
	}
}

func TestInstaller_SyncLockFile(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "arm-sync-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory for lock file operations
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)

	// Create test configuration with rulesets
	cfg := &config.Config{
		Registries: map[string]string{
			"registry1": "https://github.com/test/repo1",
			"registry2": "s3://test-bucket",
		},
		RegistryConfigs: map[string]map[string]string{
			"registry1": {"type": "git"},
			"registry2": {"type": "s3", "region": "us-west-2"},
		},
		Rulesets: map[string]map[string]config.RulesetSpec{
			"registry1": {
				"ruleset1": {Version: "^1.0.0"},
				"ruleset2": {Version: "~2.1.0"},
			},
			"registry2": {
				"ruleset3": {Version: "latest"},
			},
		},
	}

	installer := New(cfg)

	// Test syncing lock file
	err = installer.SyncLockFile()
	if err != nil {
		t.Fatalf("Failed to sync lock file: %v", err)
	}

	// Verify lock file contains all rulesets
	lockFile, err := installer.GetLockFile()
	if err != nil {
		t.Fatalf("Failed to load lock file: %v", err)
	}

	if len(lockFile.Rulesets) != 2 {
		t.Errorf("Expected 2 registries in lock file, got %d", len(lockFile.Rulesets))
	}

	// Check registry1 entries
	reg1, exists := lockFile.Rulesets["registry1"]
	if !exists {
		t.Fatalf("registry1 not found in lock file")
	}
	if len(reg1) != 2 {
		t.Errorf("Expected 2 rulesets in registry1, got %d", len(reg1))
	}

	// Check registry2 entries
	reg2, exists := lockFile.Rulesets["registry2"]
	if !exists {
		t.Fatalf("registry2 not found in lock file")
	}
	if len(reg2) != 1 {
		t.Errorf("Expected 1 ruleset in registry2, got %d", len(reg2))
	}

	// Verify specific entry details
	entry := reg2["ruleset3"]
	if entry.Version != "latest" {
		t.Errorf("Expected version 'latest', got %s", entry.Version)
	}
	if entry.Type != "s3" {
		t.Errorf("Expected type 's3', got %s", entry.Type)
	}
	if entry.Region != "us-west-2" {
		t.Errorf("Expected region 'us-west-2', got %s", entry.Region)
	}
}
