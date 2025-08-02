package uninstaller

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUninstaller_cleanupEmptyDirs(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "arm", "@org", "package")

	if err := os.MkdirAll(testPath, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	u := New()

	// Test cleanup of empty directories
	if err := u.cleanupEmptyDirs(testPath); err != nil {
		t.Errorf("cleanupEmptyDirs failed: %v", err)
	}

	// Verify @org directory was removed
	orgPath := filepath.Join(tmpDir, "arm", "@org")
	if _, err := os.Stat(orgPath); !os.IsNotExist(err) {
		t.Error("Expected @org directory to be removed")
	}

	// Verify arm directory still exists
	armPath := filepath.Join(tmpDir, "arm")
	if _, err := os.Stat(armPath); os.IsNotExist(err) {
		t.Error("Expected arm directory to remain")
	}
}

func TestUninstaller_removeRuleset(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	rulesetPath := filepath.Join(tmpDir, "arm", "test-package", "1.0.0")

	if err := os.MkdirAll(rulesetPath, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(rulesetPath, "rule.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	u := New()

	// Test removal
	if err := u.removeRuleset(rulesetPath); err != nil {
		t.Errorf("removeRuleset failed: %v", err)
	}

	// Verify ruleset directory was removed
	if _, err := os.Stat(rulesetPath); !os.IsNotExist(err) {
		t.Error("Expected ruleset directory to be removed")
	}
}
