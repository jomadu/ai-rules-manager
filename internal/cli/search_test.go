package cli

import (
	"strings"
	"testing"
)

func TestSearchPatternMatching(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		content  string
		expected bool
	}{
		{
			name:     "exact match",
			query:    "test-ruleset",
			content:  "test-ruleset.md",
			expected: true,
		},
		{
			name:     "partial match",
			query:    "ghost",
			content:  "ghost-detection.md",
			expected: true,
		},
		{
			name:     "case insensitive",
			query:    "GHOST",
			content:  "ghost-detection.md",
			expected: true,
		},
		{
			name:     "no match",
			query:    "nonexistent",
			content:  "test-ruleset.md",
			expected: false,
		},
		{
			name:     "directory match",
			query:    "ai-assistants",
			content:  "ai-assistants/q-developer.md",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesSearchQuery(tt.query, tt.content)
			if result != tt.expected {
				t.Errorf("matchesSearchQuery(%q, %q) = %v, expected %v", tt.query, tt.content, result, tt.expected)
			}
		})
	}
}

func TestSearchRegistryFiltering(t *testing.T) {
	allRegistries := map[string]string{
		"default":   "https://github.com/user/repo",
		"my-git":    "https://github.com/other/repo",
		"my-s3":     "my-bucket",
		"test-repo": "https://test.com/repo",
	}

	tests := []struct {
		name     string
		filter   string
		expected []string
	}{
		{
			name:     "no filter",
			filter:   "",
			expected: []string{"default", "my-git", "my-s3", "test-repo"},
		},
		{
			name:     "exact match",
			filter:   "default",
			expected: []string{"default"},
		},
		{
			name:     "glob pattern",
			filter:   "my-*",
			expected: []string{"my-git", "my-s3"},
		},
		{
			name:     "suffix pattern",
			filter:   "*-repo",
			expected: []string{"test-repo"},
		},
		{
			name:     "no matches",
			filter:   "nonexistent",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTargetRegistries(allRegistries, tt.filter)
			if len(result) != len(tt.expected) {
				t.Errorf("getTargetRegistries(%q) returned %d items, expected %d", tt.filter, len(result), len(tt.expected))
				return
			}

			resultMap := make(map[string]bool)
			for _, r := range result {
				resultMap[r] = true
			}

			for _, expected := range tt.expected {
				if !resultMap[expected] {
					t.Errorf("getTargetRegistries(%q) missing expected result: %s", tt.filter, expected)
				}
			}
		})
	}
}

// matchesSearchQuery is a helper function that would be used in the actual search implementation
func matchesSearchQuery(query, content string) bool {
	// Simple case-insensitive substring matching for testing
	// In real implementation, this would be more sophisticated
	query = strings.ToLower(query)
	content = strings.ToLower(content)
	return strings.Contains(content, query)
}
