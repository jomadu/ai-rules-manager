package cache

import (
	"testing"
)

func TestGenerateRegistryKey(t *testing.T) {
	tests := []struct {
		name         string
		registryType string
		registryURL  string
		expected     string
	}{
		{
			name:         "git registry",
			registryType: "git",
			registryURL:  "https://github.com/user/repo",
			expected:     "a8b9c1d2e3f4567890abcdef1234567890abcdef1234567890abcdef12345678", // This will be the actual hash
		},
		{
			name:         "s3 registry",
			registryType: "s3",
			registryURL:  "my-bucket/rules",
			expected:     "b9c1d2e3f4567890abcdef1234567890abcdef1234567890abcdef123456789a", // This will be the actual hash
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRegistryKey(tt.registryType, tt.registryURL)

			// Check that we get a 64-character hex string (SHA-256)
			if len(result) != 64 {
				t.Errorf("GenerateRegistryKey() returned hash of length %d, expected 64", len(result))
			}

			// Check that the result is consistent
			result2 := GenerateRegistryKey(tt.registryType, tt.registryURL)
			if result != result2 {
				t.Errorf("GenerateRegistryKey() is not consistent: %s != %s", result, result2)
			}
		})
	}
}

func TestGeneratePatternsKey(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		expected string
	}{
		{
			name:     "empty patterns",
			patterns: []string{},
			expected: GenerateStringKey("__EMPTY__"),
		},
		{
			name:     "single pattern",
			patterns: []string{"*.md"},
			expected: GenerateStringKey("*.md"),
		},
		{
			name:     "multiple patterns",
			patterns: []string{"rules/**", "*.md"},
			expected: GenerateStringKey("*.md,rules/**"), // Should be sorted
		},
		{
			name:     "patterns with whitespace",
			patterns: []string{" *.md ", "rules/** "},
			expected: GenerateStringKey("*.md,rules/**"), // Should be trimmed and sorted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GeneratePatternsKey(tt.patterns)

			// Check that we get a 64-character hex string (SHA-256)
			if len(result) != 64 {
				t.Errorf("GeneratePatternsKey() returned hash of length %d, expected 64", len(result))
			}

			// Check that the result is consistent
			result2 := GeneratePatternsKey(tt.patterns)
			if result != result2 {
				t.Errorf("GeneratePatternsKey() is not consistent: %s != %s", result, result2)
			}

			// For specific test cases, check expected values
			if tt.expected != "" && result != tt.expected {
				t.Errorf("GeneratePatternsKey() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestGenerateRulesetKey(t *testing.T) {
	tests := []struct {
		name        string
		rulesetName string
	}{
		{
			name:        "simple ruleset name",
			rulesetName: "power-up-rules",
		},
		{
			name:        "ruleset name with whitespace",
			rulesetName: " power-up-rules ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRulesetKey(tt.rulesetName)

			// Check that we get a 64-character hex string (SHA-256)
			if len(result) != 64 {
				t.Errorf("GenerateRulesetKey() returned hash of length %d, expected 64", len(result))
			}

			// Check that the result is consistent
			result2 := GenerateRulesetKey(tt.rulesetName)
			if result != result2 {
				t.Errorf("GenerateRulesetKey() is not consistent: %s != %s", result, result2)
			}
		})
	}
}

func TestGenerateStringKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple string",
			input: "test",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "complex string",
			input: "git:https://github.com/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateStringKey(tt.input)

			// Check that we get a 64-character hex string (SHA-256)
			if len(result) != 64 {
				t.Errorf("GenerateStringKey() returned hash of length %d, expected 64", len(result))
			}

			// Check that the result is consistent
			result2 := GenerateStringKey(tt.input)
			if result != result2 {
				t.Errorf("GenerateStringKey() is not consistent: %s != %s", result, result2)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "url with trailing slash",
			input:    "https://github.com/user/repo/",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "url with whitespace",
			input:    " https://github.com/user/repo ",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "url without trailing slash",
			input:    "https://github.com/user/repo",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "s3 bucket",
			input:    "my-bucket/rules/",
			expected: "my-bucket/rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeURL() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestFactoryFunctions(t *testing.T) {
	cacheRoot := "/tmp/test-cache"

	t.Run("NewGitRegistryCacheManager", func(t *testing.T) {
		manager := NewGitRegistryCacheManager(cacheRoot)
		// Should now return a valid implementation
		if manager == nil {
			t.Errorf("NewGitRegistryCacheManager() should return a valid implementation")
		}
	})

	t.Run("NewS3RegistryCacheManager", func(t *testing.T) {
		manager := NewS3RegistryCacheManager(cacheRoot)
		// Should now return a valid implementation
		if manager == nil {
			t.Errorf("NewS3RegistryCacheManager() should return a valid implementation")
		}
	})

	t.Run("NewHTTPSRegistryCacheManager", func(t *testing.T) {
		manager := NewHTTPSRegistryCacheManager(cacheRoot)
		// Should now return a valid implementation
		if manager == nil {
			t.Errorf("NewHTTPSRegistryCacheManager() should return a valid implementation")
		}
	})

	t.Run("NewLocalRegistryCacheManager", func(t *testing.T) {
		manager := NewLocalRegistryCacheManager(cacheRoot)
		// Should now return a valid implementation
		if manager == nil {
			t.Errorf("NewLocalRegistryCacheManager() should return a valid implementation")
		}
	})
}
