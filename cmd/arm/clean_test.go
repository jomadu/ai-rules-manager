package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jomadu/arm/pkg/types"
)

func TestCleanProjectTargets(t *testing.T) {
	// Create temporary directory and change to it
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create rules.json
	manifest := &types.RulesManifest{
		Targets:      []string{".cursorrules", ".amazonq/rules"},
		Dependencies: map[string]string{"used-package": "^1.0.0"},
	}
	if err := manifest.SaveManifest("rules.json"); err != nil {
		t.Fatalf("Failed to create rules.json: %v", err)
	}

	// Create rules.lock
	lock := &types.RulesLock{
		Version: "1",
		Dependencies: map[string]types.LockedDependency{
			"used-package": {
				Version:  "1.0.0",
				Source:   "https://registry.armjs.org/",
				Checksum: "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234",
			},
		},
	}
	if err := lock.SaveLockFile("rules.lock"); err != nil {
		t.Fatalf("Failed to create rules.lock: %v", err)
	}

	// Create test rulesets
	createTestRuleset(t, ".cursorrules", "used-package", "1.0.0")
	createTestRuleset(t, ".cursorrules", "unused-package", "1.0.0")

	// Test cleanup
	err := cleanProjectTargets()
	if err != nil {
		t.Fatalf("cleanProjectTargets failed: %v", err)
	}

	// Verify results
	if !pathExists(".cursorrules/arm/used-package") {
		t.Error("Used package should not be removed")
	}
	if pathExists(".cursorrules/arm/unused-package") {
		t.Error("Unused package should be removed")
	}
}

func TestCleanProjectTargetsNoManifest(t *testing.T) {
	// Create temporary directory and change to it
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create test rulesets in default targets
	createTestRuleset(t, ".cursorrules", "some-package", "1.0.0")
	createTestRuleset(t, ".amazonq/rules", "@org", "another-package", "2.0.0")

	// Test cleanup without manifest (should remove all packages)
	err := cleanProjectTargets()
	if err != nil {
		t.Fatalf("cleanProjectTargets failed: %v", err)
	}

	// All packages should be removed since no rules.lock exists
	if pathExists(".cursorrules/arm/some-package") {
		t.Error("Package should be removed when no rules.lock exists")
	}
	if pathExists(".amazonq/rules/arm/@org") {
		t.Error("Org package should be removed when no rules.lock exists")
	}
}

func TestCleanGlobalCache(t *testing.T) {
	// Set up temporary home directory
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpHome)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Create cache directory with test content
	cacheDir := filepath.Join(tmpHome, ".arm", "cache")
	if err := os.MkdirAll(filepath.Join(cacheDir, "packages"), 0o755); err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "test.txt"), []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test cache file: %v", err)
	}

	// Set dry run flag
	cleanDryRun = true
	defer func() { cleanDryRun = false }()

	// Test dry run
	err := cleanGlobalCache()
	if err != nil {
		t.Fatalf("cleanGlobalCache dry run failed: %v", err)
	}

	// Cache should still exist after dry run
	if !pathExists(cacheDir) {
		t.Error("Cache should exist after dry run")
	}
}

func TestCleanGlobalCacheNonExistent(t *testing.T) {
	// Set up temporary home directory without cache
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpHome)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Set dry run to avoid stdin interaction
	cleanDryRun = true
	defer func() { cleanDryRun = false }()

	// Test cleanup of non-existent cache
	err := cleanGlobalCache()
	if err != nil {
		t.Fatalf("cleanGlobalCache should handle non-existent cache: %v", err)
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
