package cleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanTargets(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Set up test directory structure
	targets := []string{
		filepath.Join(tmpDir, ".cursorrules"),
		filepath.Join(tmpDir, ".amazonq", "rules"),
	}

	// Create test rulesets
	createTestRuleset(t, targets[0], "used-package", "1.0.0")
	createTestRuleset(t, targets[0], "unused-package", "1.0.0")
	createTestRuleset(t, targets[1], "@company", "used-org-package", "2.0.0")
	createTestRuleset(t, targets[1], "@unused-org", "unused-org-package", "1.5.0")

	// Define used rulesets
	usedRulesets := map[string]bool{
		"used-package":             true,
		"company@used-org-package": true,
	}

	cleaner := New()

	// Test dry run
	err := cleaner.CleanTargets(targets, usedRulesets, true)
	if err != nil {
		t.Fatalf("CleanTargets dry run failed: %v", err)
	}

	// Verify nothing was actually removed in dry run
	if !pathExists(filepath.Join(targets[0], "arm", "unused-package")) {
		t.Error("Dry run should not remove files")
	}

	// Test actual cleanup
	err = cleaner.CleanTargets(targets, usedRulesets, false)
	if err != nil {
		t.Fatalf("CleanTargets failed: %v", err)
	}

	// Verify used packages still exist
	if !pathExists(filepath.Join(targets[0], "arm", "used-package")) {
		t.Error("Used package should not be removed")
	}
	if !pathExists(filepath.Join(targets[1], "arm", "@company", "used-org-package")) {
		t.Error("Used org package should not be removed")
	}

	// Verify unused packages were removed
	if pathExists(filepath.Join(targets[0], "arm", "unused-package")) {
		t.Error("Unused package should be removed")
	}
	if pathExists(filepath.Join(targets[1], "arm", "@unused-org", "unused-org-package")) {
		t.Error("Unused org package should be removed")
	}
}

func TestCleanTargetsNoArmDir(t *testing.T) {
	tmpDir := t.TempDir()
	targets := []string{filepath.Join(tmpDir, ".cursorrules")}

	cleaner := New()
	err := cleaner.CleanTargets(targets, nil, false)
	if err != nil {
		t.Fatalf("CleanTargets should handle missing arm directory: %v", err)
	}
}

func TestExtractRulesetName(t *testing.T) {
	cleaner := New()
	armDir := "/test/arm"

	tests := []struct {
		path     string
		expected string
	}{
		{"/test/arm/simple-package", "simple-package"},
		{"/test/arm/@company/package", "company@package"},
		{"/test/arm/@org/my-rules", "org@my-rules"},
		{"/test/arm", ""},
	}

	for _, test := range tests {
		result := cleaner.extractRulesetName(armDir, test.path)
		if result != test.expected {
			t.Errorf("extractRulesetName(%q) = %q, want %q", test.path, result, test.expected)
		}
	}
}

func TestCleanEmptyDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested empty directories
	emptyDir := filepath.Join(tmpDir, "empty", "nested")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create directory with content
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.MkdirAll(contentDir, 0o755); err != nil {
		t.Fatalf("Failed to create content directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(contentDir, "file.txt"), []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cleaner := New()
	cleaner.cleanEmptyDirs(tmpDir)

	// Empty directories should be removed
	if pathExists(emptyDir) {
		t.Error("Empty directory should be removed")
	}

	// Directory with content should remain
	if !pathExists(contentDir) {
		t.Error("Directory with content should not be removed")
	}
}

// Helper functions
func createTestRuleset(t *testing.T, target string, parts ...string) {
	path := filepath.Join(append([]string{target, "arm"}, parts...)...)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("Failed to create test ruleset directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(path, "rule.md"), []byte("test rule"), 0o644); err != nil {
		t.Fatalf("Failed to create test rule file: %v", err)
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
