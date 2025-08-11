package registry

import (
	"testing"
	"time"
)

func TestNewGitRegistry(t *testing.T) {
	config := &RegistryConfig{
		Name:    "test-git",
		Type:    "git",
		URL:     "https://github.com/user/repo",
		Timeout: 30 * time.Second,
	}
	auth := &AuthConfig{
		Token: "test-token",
	}

	registry, err := NewGitRegistry(config, auth)
	if err != nil {
		t.Fatalf("Failed to create Git registry: %v", err)
	}

	if registry.GetType() != "git" {
		t.Errorf("Expected type 'git', got %q", registry.GetType())
	}
	if registry.GetName() != "test-git" {
		t.Errorf("Expected name 'test-git', got %q", registry.GetName())
	}
}

func TestNewGitRegistryInvalidConfig(t *testing.T) {
	config := &RegistryConfig{
		Name: "test-git",
		Type: "git",
		// Missing URL
	}
	auth := &AuthConfig{}

	_, err := NewGitRegistry(config, auth)
	if err == nil {
		t.Error("Expected error for invalid config")
	}
}

// TestParseGitHubURL moved to remote_git_operations_test.go
// since parseGitHubURL is now part of RemoteGitOperations

// TestMatchesPatterns moved to git_common_test.go
// since pattern matching is now in shared utilities

func TestGetRulesetsAPI(t *testing.T) {
	// Skip API tests for now - they require complex mocking
	// Will be tested in integration tests
	t.Skip("API tests require complex GitHub API mocking")
}

func TestGetVersionsAPI(t *testing.T) {
	// Skip API tests for now - they require complex mocking
	// Will be tested in integration tests
	t.Skip("API tests require complex GitHub API mocking")
}

// TestScanDirectory functionality moved to git_common_test.go
// since directory scanning is now handled by shared utilities

func TestGitRegistryClose(t *testing.T) {
	config := &RegistryConfig{
		Name: "test-git",
		Type: "git",
		URL:  "https://github.com/owner/repo",
	}
	auth := &AuthConfig{}

	registry, err := NewGitRegistry(config, auth)
	if err != nil {
		t.Fatalf("Failed to create Git registry: %v", err)
	}

	err = registry.Close()
	if err != nil {
		t.Errorf("Expected no error from Close(), got: %v", err)
	}
}
