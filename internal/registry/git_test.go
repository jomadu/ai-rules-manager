package registry

import (
	"os"
	"path/filepath"
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

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedOwner string
		expectedRepo  string
		expectError bool
	}{
		{
			name:         "standard GitHub URL",
			url:          "https://github.com/owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:  false,
		},
		{
			name:         "GitHub URL with .git",
			url:          "https://github.com/owner/repo.git",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:  false,
		},
		{
			name:         "GitHub URL with trailing slash",
			url:          "https://github.com/owner/repo/",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:  false,
		},
		{
			name:        "invalid URL",
			url:         "https://gitlab.com/owner/repo",
			expectError: true,
		},
		{
			name:        "malformed URL",
			url:         "not-a-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RegistryConfig{
				Name: "test",
				Type: "git",
				URL:  tt.url,
			}
			registry := &GitRegistry{config: config}

			owner, repo, err := registry.parseGitHubURL()
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if owner != tt.expectedOwner {
					t.Errorf("Expected owner %q, got %q", tt.expectedOwner, owner)
				}
				if repo != tt.expectedRepo {
					t.Errorf("Expected repo %q, got %q", tt.expectedRepo, repo)
				}
			}
		})
	}
}

func TestMatchesPatterns(t *testing.T) {
	registry := &GitRegistry{}

	tests := []struct {
		name     string
		filename string
		patterns []string
		expected bool
	}{
		{
			name:     "no patterns matches all",
			filename: "test.md",
			patterns: []string{},
			expected: true,
		},
		{
			name:     "exact match",
			filename: "test.md",
			patterns: []string{"test.md"},
			expected: true,
		},
		{
			name:     "wildcard match",
			filename: "test.md",
			patterns: []string{"*.md"},
			expected: true,
		},
		{
			name:     "prefix wildcard",
			filename: "test-rules.md",
			patterns: []string{"test*"},
			expected: true,
		},
		{
			name:     "no match",
			filename: "test.md",
			patterns: []string{"*.txt"},
			expected: false,
		},
		{
			name:     "multiple patterns with match",
			filename: "test.md",
			patterns: []string{"*.txt", "*.md"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.matchesPatterns(tt.filename, tt.patterns)
			if result != tt.expected {
				t.Errorf("matchesPatterns(%q, %v) = %v, want %v", tt.filename, tt.patterns, result, tt.expected)
			}
		})
	}
}

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

func TestScanDirectory(t *testing.T) {
	// Create temporary directory with test files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"python-rules.md",
		"javascript-rules.md",
		"README.md",
		"config.json",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	config := &RegistryConfig{
		Name: "test-git",
		Type: "git",
		URL:  "https://github.com/owner/repo",
	}
	registry := &GitRegistry{config: config}

	// Test scanning with pattern
	rulesets, err := registry.scanDirectory(tmpDir, []string{"*-rules.md"})
	if err != nil {
		t.Fatalf("Failed to scan directory: %v", err)
	}

	if len(rulesets) != 2 {
		t.Errorf("Expected 2 rulesets, got %d", len(rulesets))
	}

	// Check that we got the right files
	foundNames := make(map[string]bool)
	for _, ruleset := range rulesets {
		foundNames[ruleset.Name] = true
	}

	expectedNames := []string{"python-rules", "javascript-rules"}
	for _, expected := range expectedNames {
		if !foundNames[expected] {
			t.Errorf("Expected to find ruleset %q", expected)
		}
	}
}

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
