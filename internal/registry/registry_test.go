package registry

import (
	"strings"
	"testing"
	"time"
)

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
