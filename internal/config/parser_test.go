package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".armrc")

	configContent := `[sources]
default = https://registry.armjs.org/
company = https://internal.company.local/

[sources.company]
authToken = secret123
timeout = 30s

[cache]
location = ~/.arm/cache
maxSize = 1GB
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Parse the config
	config, err := ParseFile(configPath)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Verify sources
	if len(config.Sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(config.Sources))
	}

	defaultSource, exists := config.Sources["default"]
	if !exists {
		t.Error("Expected default source to exist")
	}
	if defaultSource.URL != "https://registry.armjs.org/" {
		t.Errorf("Expected default URL to be https://registry.armjs.org/, got %s", defaultSource.URL)
	}

	companySource, exists := config.Sources["company"]
	if !exists {
		t.Error("Expected company source to exist")
	}
	if companySource.URL != "https://internal.company.local/" {
		t.Errorf("Expected company URL to be https://internal.company.local/, got %s", companySource.URL)
	}
	if companySource.AuthToken != "secret123" {
		t.Errorf("Expected company auth token to be secret123, got %s", companySource.AuthToken)
	}
	if companySource.Timeout != "30s" {
		t.Errorf("Expected company timeout to be 30s, got %s", companySource.Timeout)
	}

	// Verify cache config
	if config.Cache.Location != "~/.arm/cache" {
		t.Errorf("Expected cache location to be ~/.arm/cache, got %s", config.Cache.Location)
	}
	if config.Cache.MaxSize != "1GB" {
		t.Errorf("Expected cache max size to be 1GB, got %s", config.Cache.MaxSize)
	}
}

func TestEnvironmentVariableSubstitution(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("TEST_TOKEN", "env_token_123")
	_ = os.Setenv("TEST_URL", "https://env.example.com")
	defer func() {
		_ = os.Unsetenv("TEST_TOKEN")
		_ = os.Unsetenv("TEST_URL")
	}()

	// Create a temporary config file with env vars
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".armrc")

	configContent := `[sources]
default = $TEST_URL
company = ${TEST_URL}/company

[sources.company]
authToken = $TEST_TOKEN
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Parse the config
	config, err := ParseFile(configPath)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Verify environment variable substitution
	defaultSource := config.Sources["default"]
	if defaultSource.URL != "https://env.example.com" {
		t.Errorf("Expected default URL to be https://env.example.com, got %s", defaultSource.URL)
	}

	companySource := config.Sources["company"]
	if companySource.URL != "https://env.example.com/company" {
		t.Errorf("Expected company URL to be https://env.example.com/company, got %s", companySource.URL)
	}
	if companySource.AuthToken != "env_token_123" {
		t.Errorf("Expected company auth token to be env_token_123, got %s", companySource.AuthToken)
	}
}
