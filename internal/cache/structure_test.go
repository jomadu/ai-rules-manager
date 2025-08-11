package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestThreeLevelCacheStructure(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	registryType := "git"
	registryURL := "https://github.com/user/repo"
	ruleset := "test-ruleset"
	version := "v1.0.0"
	patterns := []string{"*.md"}

	// Create cache structure
	err := manager.CreateRulesetCache(registryType, registryURL, ruleset, version, patterns)
	if err != nil {
		t.Fatalf("CreateRulesetCache failed: %v", err)
	}

	// Validate three-level hierarchy
	registriesDir := filepath.Join(tempDir, "registries")
	if _, err := os.Stat(registriesDir); os.IsNotExist(err) {
		t.Error("registries directory should exist")
	}

	// Find registry hash directory
	entries, err := os.ReadDir(registriesDir)
	if err != nil || len(entries) == 0 {
		t.Fatal("registry hash directory should exist")
	}

	registryHashDir := filepath.Join(registriesDir, entries[0].Name())

	// Validate repository directory exists at registry level
	repositoryDir := filepath.Join(registryHashDir, "repository")
	if _, err := os.Stat(repositoryDir); os.IsNotExist(err) {
		t.Error("repository directory should exist at registry level")
	}

	// Validate rulesets directory exists at registry level
	rulesetsDir := filepath.Join(registryHashDir, "rulesets")
	if _, err := os.Stat(rulesetsDir); os.IsNotExist(err) {
		t.Error("rulesets directory should exist at registry level")
	}

	// Validate ruleset hash directories exist
	rulesetEntries, err := os.ReadDir(rulesetsDir)
	if err != nil || len(rulesetEntries) == 0 {
		t.Fatal("ruleset hash directory should exist")
	}

	rulesetHashDir := filepath.Join(rulesetsDir, rulesetEntries[0].Name())

	// Validate version directories exist in rulesets
	versionEntries, err := os.ReadDir(rulesetHashDir)
	if err != nil || len(versionEntries) == 0 {
		t.Fatal("version directory should exist in ruleset")
	}
}

func TestRulesetMapping(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	registryType := "git"
	registryURL := "https://github.com/user/repo"
	ruleset1 := "test-ruleset"
	ruleset2 := "test-ruleset" // Same ruleset name
	version := "v1.0.0"
	patterns1 := []string{"*.md"}
	patterns2 := []string{"ghost-*.md"} // Different patterns

	// Create cache for same ruleset with different patterns
	err := manager.CreateRulesetCache(registryType, registryURL, ruleset1, version, patterns1)
	if err != nil {
		t.Fatalf("CreateRulesetCache failed: %v", err)
	}

	err = manager.CreateRulesetCache(registryType, registryURL, ruleset2, version, patterns2)
	if err != nil {
		t.Fatalf("CreateRulesetCache failed: %v", err)
	}

	// Validate ruleset-map.json exists
	mapPath := filepath.Join(tempDir, "ruleset-map.json")
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		t.Error("ruleset-map.json should exist")
	}

	// Load and validate mapping
	rulesetMap, err := manager.LoadRulesetMap()
	if err != nil {
		t.Fatalf("LoadRulesetMap failed: %v", err)
	}

	if len(rulesetMap.Mappings) < 2 {
		t.Errorf("Expected at least 2 mappings for different patterns, got %d", len(rulesetMap.Mappings))
	}

	// Verify different patterns create different cache entries
	registriesDir := filepath.Join(tempDir, "registries")
	entries, _ := os.ReadDir(registriesDir)
	registryHashDir := filepath.Join(registriesDir, entries[0].Name())
	rulesetsDir := filepath.Join(registryHashDir, "rulesets")

	rulesetEntries, err := os.ReadDir(rulesetsDir)
	if err != nil {
		t.Fatalf("Failed to read rulesets directory: %v", err)
	}

	if len(rulesetEntries) < 2 {
		t.Errorf("Expected at least 2 ruleset hash directories for different patterns, got %d", len(rulesetEntries))
	}
}

func TestCachePathGeneration(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	tests := []struct {
		name         string
		registryType string
		registryURL  string
		ruleset      string
		patterns     []string
	}{
		{
			name:         "git registry with md files",
			registryType: "git",
			registryURL:  "https://github.com/user/repo",
			ruleset:      "test-ruleset",
			patterns:     []string{"*.md"},
		},
		{
			name:         "s3 registry with json files",
			registryType: "s3",
			registryURL:  "my-bucket",
			ruleset:      "config-rules",
			patterns:     []string{"*.json", "*.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get registry cache path
			registryPath, err := manager.GetCachePath(tt.registryType, tt.registryURL)
			if err != nil {
				t.Fatalf("GetCachePath failed: %v", err)
			}

			// Get ruleset cache path
			rulesetPath, err := manager.GetRulesetCachePath(tt.registryType, tt.registryURL, tt.ruleset, tt.patterns)
			if err != nil {
				t.Fatalf("GetRulesetCachePath failed: %v", err)
			}

			// Verify ruleset path is under registry path
			if !strings.HasPrefix(rulesetPath, registryPath) {
				t.Errorf("Ruleset path %s should be under registry path %s", rulesetPath, registryPath)
			}

			// Verify path structure
			expectedPrefix := filepath.Join(registryPath, "rulesets")
			if !strings.HasPrefix(rulesetPath, expectedPrefix) {
				t.Errorf("Ruleset path should start with %s, got %s", expectedPrefix, rulesetPath)
			}
		})
	}
}
