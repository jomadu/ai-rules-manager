package config

import (
	"testing"
)

func TestFileConfigManager_Operations(t *testing.T) {
	manager := NewFileConfigManager(".armrc.json", "arm.json", "arm.lock")

	// Test LoadInfraConfig behavior
	infraConfig, err := manager.LoadInfraConfig()
	// Should handle missing file gracefully or return empty config
	if err != nil {
		// File not found is acceptable for first run
		if infraConfig != nil {
			t.Errorf("if error occurs, config should be nil")
		}
	} else if infraConfig == nil {
		// If no error, should return valid config structure
		t.Errorf("successful load should return non-nil config")
	}

	// Test SaveInfraConfig behavior
	testConfig := &InfraConfig{
		Registries: map[string]*RegistryConfig{
			"test": {URL: "https://github.com/user/repo", Type: "git"},
		},
		Sinks: map[string]*SinkConfig{
			"cursor": {Directories: []string{".cursor/rules"}},
		},
	}
	err = manager.SaveInfraConfig(testConfig)
	if err != nil {
		t.Errorf("save should not error for valid config: %v", err)
	}

	// Test LoadManifest behavior
	manifest, err := manager.LoadManifest()
	if err != nil {
		// File not found is acceptable
		if manifest != nil {
			t.Errorf("if error occurs, manifest should be nil")
		}
	} else if manifest == nil {
		t.Errorf("successful load should return non-nil manifest")
	}

	// Test SaveManifest behavior
	testManifest := &Manifest{
		Rulesets: map[string]map[string]*ManifestEntry{
			"test": {
				"ruleset": {Version: "^1.0.0", Patterns: []string{"*.md"}},
			},
		},
	}
	err = manager.SaveManifest(testManifest)
	if err != nil {
		t.Errorf("save should not error for valid manifest: %v", err)
	}

	// Test LoadLockFile behavior
	lockFile, err := manager.LoadLockFile()
	if err != nil {
		// File not found is acceptable
		if lockFile != nil {
			t.Errorf("if error occurs, lockFile should be nil")
		}
	} else if lockFile == nil {
		t.Errorf("successful load should return non-nil lockFile")
	}

	// Test SaveLockFile behavior
	testLockFile := &LockFile{
		Rulesets: map[string]map[string]*LockEntry{
			"test": {
				"ruleset": {
					URL:        "https://github.com/user/repo",
					Type:       "git",
					Constraint: "^1.0.0",
					Resolved:   "1.2.3",
					Patterns:   []string{"*.md"},
				},
			},
		},
	}
	err = manager.SaveLockFile(testLockFile)
	if err != nil {
		t.Errorf("save should not error for valid lockFile: %v", err)
	}
}

func TestConfigStructures(t *testing.T) {
	// Test that config structures can be created and have expected fields
	infraConfig := &InfraConfig{
		Registries: map[string]*RegistryConfig{
			"test": {
				URL:  "https://github.com/user/repo",
				Type: "git",
			},
		},
		Sinks: map[string]*SinkConfig{
			"cursor": {
				Directories: []string{".cursor/rules"},
				Rulesets:    []string{"test/ruleset"},
			},
		},
	}

	if infraConfig.Registries["test"].URL != "https://github.com/user/repo" {
		t.Errorf("registry URL not set correctly")
	}

	manifest := &Manifest{
		Rulesets: map[string]map[string]*ManifestEntry{
			"test": {
				"ruleset": {
					Version:  "^1.0.0",
					Patterns: []string{"rules/*.md"},
				},
			},
		},
	}

	if manifest.Rulesets["test"]["ruleset"].Version != "^1.0.0" {
		t.Errorf("manifest version not set correctly")
	}

	lockFile := &LockFile{
		Rulesets: map[string]map[string]*LockEntry{
			"test": {
				"ruleset": {
					URL:        "https://github.com/user/repo",
					Type:       "git",
					Constraint: "^1.0.0",
					Resolved:   "1.2.3",
					Patterns:   []string{"rules/*.md"},
				},
			},
		},
	}

	if lockFile.Rulesets["test"]["ruleset"].Resolved != "1.2.3" {
		t.Errorf("lock file resolved version not set correctly")
	}
}
