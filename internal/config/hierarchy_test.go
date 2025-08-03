package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigHierarchy(t *testing.T) {
	// Save original working directory
	originalWd, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	_ = os.Chdir(tmpDir)

	// Create user config
	userConfigContent := `[sources]
default = https://user.registry.com/
user-only = https://user-only.com/

[cache]
location = ~/.user/cache
`
	userConfigPath := filepath.Join(tmpDir, ".armrc-user")
	if err := os.WriteFile(userConfigPath, []byte(userConfigContent), 0o644); err != nil {
		t.Fatalf("Failed to write user config: %v", err)
	}

	// Create project config
	projectConfigContent := `[sources]
default = https://project.registry.com/
project-only = https://project-only.com/

[sources.project-only]
authToken = project_token

[cache]
maxSize = 2GB
`
	projectConfigPath := ".armrc"
	if err := os.WriteFile(projectConfigPath, []byte(projectConfigContent), 0o644); err != nil {
		t.Fatalf("Failed to write project config: %v", err)
	}

	// Test merging by manually loading configs
	userConfig, err := ParseFile(userConfigPath)
	if err != nil {
		t.Fatalf("Failed to parse user config: %v", err)
	}

	projectConfig, err := ParseFile(projectConfigPath)
	if err != nil {
		t.Fatalf("Failed to parse project config: %v", err)
	}

	// Merge configs (project overrides user)
	mergedConfig := &ARMConfig{Sources: make(map[string]Source)}
	mergeConfigs(mergedConfig, userConfig)
	mergeConfigs(mergedConfig, projectConfig)

	// Verify merged results
	if len(mergedConfig.Sources) != 3 {
		t.Errorf("Expected 3 sources, got %d", len(mergedConfig.Sources))
	}

	// Project config should override user config for 'default'
	defaultSource := mergedConfig.Sources["default"]
	if defaultSource.URL != "https://project.registry.com/" {
		t.Errorf("Expected default URL to be project URL, got %s", defaultSource.URL)
	}

	// User-only source should be preserved
	userOnlySource := mergedConfig.Sources["user-only"]
	if userOnlySource.URL != "https://user-only.com/" {
		t.Errorf("Expected user-only URL to be preserved, got %s", userOnlySource.URL)
	}

	// Project-only source should be present
	projectOnlySource := mergedConfig.Sources["project-only"]
	if projectOnlySource.URL != "https://project-only.com/" {
		t.Errorf("Expected project-only URL, got %s", projectOnlySource.URL)
	}
	if projectOnlySource.AuthToken != "project_token" {
		t.Errorf("Expected project-only auth token, got %s", projectOnlySource.AuthToken)
	}

	// Cache config should be merged (project overrides user where present)
	if mergedConfig.Cache.Location != "~/.user/cache" {
		t.Errorf("Expected cache location from user config, got %s", mergedConfig.Cache.Location)
	}
	if mergedConfig.Cache.MaxSize != "2GB" {
		t.Errorf("Expected cache max size from project config, got %s", mergedConfig.Cache.MaxSize)
	}
}
