package registry

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestDefaultAuthProvider(t *testing.T) {
	provider := NewDefaultAuthProvider()

	// Test empty credentials
	auth, err := provider.GetCredentials("nonexistent")
	if err != nil {
		t.Errorf("Expected no error for nonexistent registry, got: %v", err)
	}
	if auth.Token != "" {
		t.Error("Expected empty token for nonexistent registry")
	}

	// Set auth config
	expectedAuth := &AuthConfig{
		Token:    "test-token",
		Username: "test-user",
		Region:   "us-east-1",
	}
	provider.SetAuth("test-registry", expectedAuth)

	// Get auth config
	auth, err = provider.GetCredentials("test-registry")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if auth.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got %q", auth.Token)
	}
	if auth.Username != "test-user" {
		t.Errorf("Expected username 'test-user', got %q", auth.Username)
	}
	if auth.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got %q", auth.Region)
	}
}

func TestAuthProviderEnvironmentVariables(t *testing.T) {
	provider := NewDefaultAuthProvider()

	// Set environment variables
	_ = os.Setenv("TEST_TOKEN", "env-token")
	_ = os.Setenv("TEST_REGION", "us-west-2")
	defer func() { _ = os.Unsetenv("TEST_TOKEN") }()
	defer func() { _ = os.Unsetenv("TEST_REGION") }()

	// Set auth config with environment variables
	authConfig := &AuthConfig{
		Token:  "$TEST_TOKEN",
		Region: "${TEST_REGION}",
	}
	provider.SetAuth("env-registry", authConfig)

	// Get expanded credentials
	auth, err := provider.GetCredentials("env-registry")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if auth.Token != "env-token" {
		t.Errorf("Expected expanded token 'env-token', got %q", auth.Token)
	}
	if auth.Region != "us-west-2" {
		t.Errorf("Expected expanded region 'us-west-2', got %q", auth.Region)
	}
}

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "no env var",
			input:    "plain-text",
			expected: "plain-text",
		},
		{
			name:     "$VAR format",
			input:    "$TEST_VAR",
			envVar:   "TEST_VAR",
			envValue: "test-value",
			expected: "test-value",
		},
		{
			name:     "${VAR} format",
			input:    "${TEST_VAR}",
			envVar:   "TEST_VAR",
			envValue: "test-value",
			expected: "test-value",
		},
		{
			name:     "missing env var",
			input:    "$MISSING_VAR",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				_ = os.Setenv(tt.envVar, tt.envValue)
				defer func() { _ = os.Unsetenv(tt.envVar) }()
			}

			result := expandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateRegistryConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        *RegistryConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "valid git registry",
			config: &RegistryConfig{
				Name: "my-git",
				Type: "git",
				URL:  "https://github.com/user/repo",
			},
			expectError: false,
		},
		{
			name: "valid s3 registry",
			config: &RegistryConfig{
				Name: "my-s3",
				Type: "s3",
				URL:  "my-bucket",
				Auth: &AuthConfig{Region: "us-east-1"},
			},
			expectError: false,
		},
		{
			name: "valid local registry",
			config: &RegistryConfig{
				Name: "my-local",
				Type: "local",
				URL:  "/path/to/registry",
			},
			expectError: false,
		},
		{
			name: "empty name",
			config: &RegistryConfig{
				Type: "git",
				URL:  "https://github.com/user/repo",
			},
			expectError:   true,
			errorContains: "name cannot be empty",
		},
		{
			name: "empty type",
			config: &RegistryConfig{
				Name: "test",
				URL:  "https://github.com/user/repo",
			},
			expectError:   true,
			errorContains: "type cannot be empty",
		},
		{
			name: "invalid type",
			config: &RegistryConfig{
				Name: "test",
				Type: "ftp",
				URL:  "ftp://example.com",
			},
			expectError:   true,
			errorContains: "unsupported registry type",
		},
		{
			name: "git without URL",
			config: &RegistryConfig{
				Name: "test",
				Type: "git",
			},
			expectError:   true,
			errorContains: "requires URL",
		},
		{
			name: "git with non-HTTPS URL",
			config: &RegistryConfig{
				Name: "test",
				Type: "git",
				URL:  "http://github.com/user/repo",
			},
			expectError:   true,
			errorContains: "must use HTTPS",
		},
		{
			name: "s3 without region",
			config: &RegistryConfig{
				Name: "test",
				Type: "s3",
				URL:  "my-bucket",
			},
			expectError:   true,
			errorContains: "requires region",
		},
		{
			name: "local without path",
			config: &RegistryConfig{
				Name: "test",
				Type: "local",
			},
			expectError:   true,
			errorContains: "requires path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryConfig(tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestRulesetInfo(t *testing.T) {
	// Test RulesetInfo structure
	ruleset := &RulesetInfo{
		Name:        "test-ruleset",
		Version:     "1.0.0",
		Description: "Test ruleset",
		Author:      "Test Author",
		Tags:        []string{"test", "example"},
		Patterns:    []string{"*.md", "**/*.txt"},
		Metadata:    map[string]string{"key": "value"},
		Registry:    "test-registry",
		Type:        "git",
		UpdatedAt:   time.Now(),
	}

	if ruleset.Name != "test-ruleset" {
		t.Errorf("Expected name 'test-ruleset', got %q", ruleset.Name)
	}
	if len(ruleset.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(ruleset.Tags))
	}
	if len(ruleset.Patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(ruleset.Patterns))
	}
	if ruleset.Metadata["key"] != "value" {
		t.Errorf("Expected metadata value 'value', got %q", ruleset.Metadata["key"])
	}
}

func TestAuthConfigStructure(t *testing.T) {
	// Test AuthConfig structure
	auth := &AuthConfig{
		Token:      "test-token",
		Username:   "test-user",
		Password:   "test-pass",
		Profile:    "test-profile",
		Region:     "us-east-1",
		APIType:    "github",
		APIVersion: "2022-11-28",
	}

	if auth.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got %q", auth.Token)
	}
	if auth.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got %q", auth.Region)
	}
	if auth.APIType != "github" {
		t.Errorf("Expected API type 'github', got %q", auth.APIType)
	}
}
