package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_PerformanceConfig(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".armrc")

	configContent := `[sources]
default = https://registry.armjs.org/
company = https://gitlab.company.com
company-2 = https://gitlab2.company.com

[sources.company]
type = gitlab
concurrency = 2

[sources.company-2]
type = gitlab

[performance]
defaultConcurrency = 5

[performance.gitlab]
concurrency = 3

[performance.s3]
concurrency = 8
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Parse the config
	config, err := ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	// Test performance config
	if config.Performance.DefaultConcurrency != 5 {
		t.Errorf("Expected defaultConcurrency = 5, got %d", config.Performance.DefaultConcurrency)
	}

	// Test registry type configs
	if gitlabConfig, exists := config.Performance.RegistryTypes["gitlab"]; !exists {
		t.Error("Expected gitlab performance config to exist")
	} else if gitlabConfig.Concurrency != 3 {
		t.Errorf("Expected gitlab concurrency = 3, got %d", gitlabConfig.Concurrency)
	}

	if s3Config, exists := config.Performance.RegistryTypes["s3"]; !exists {
		t.Error("Expected s3 performance config to exist")
	} else if s3Config.Concurrency != 8 {
		t.Errorf("Expected s3 concurrency = 8, got %d", s3Config.Concurrency)
	}

	// Test source-specific concurrency
	if companySource, exists := config.Sources["company"]; !exists {
		t.Error("Expected company source to exist")
	} else if companySource.Concurrency != 2 {
		t.Errorf("Expected company concurrency = 2, got %d", companySource.Concurrency)
	}

	if company2Source, exists := config.Sources["company-2"]; !exists {
		t.Error("Expected company-2 source to exist")
	} else if company2Source.Concurrency != 0 {
		t.Errorf("Expected company-2 concurrency = 0 (not set), got %d", company2Source.Concurrency)
	}
}

func TestParseFile_DefaultPerformanceConfig(t *testing.T) {
	// Create temporary config file with minimal content
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".armrc")

	configContent := `[sources]
default = https://registry.armjs.org/
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Parse the config
	config, err := ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	// Test default performance config
	if config.Performance.DefaultConcurrency != 3 {
		t.Errorf("Expected default defaultConcurrency = 3, got %d", config.Performance.DefaultConcurrency)
	}

	if config.Performance.RegistryTypes == nil {
		t.Error("Expected RegistryTypes map to be initialized")
	}
}
