package cli

import (
	"fmt"
	"strings"
	"testing"
)

func TestParsePatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single pattern",
			input:    "*.md",
			expected: []string{"*.md"},
		},
		{
			name:     "multiple patterns comma separated",
			input:    "*.md,*.txt",
			expected: []string{"*.md", "*.txt"},
		},
		{
			name:     "patterns with spaces",
			input:    "*.md, *.txt, rules/*.json",
			expected: []string{"*.md", "*.txt", "rules/*.json"},
		},
		{
			name:     "directory patterns",
			input:    "ai-assistants/*.md,tools/*.md",
			expected: []string{"ai-assistants/*.md", "tools/*.md"},
		},
		{
			name:     "empty pattern",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single comma",
			input:    ",",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePatterns(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parsePatterns(%q) returned %d patterns, expected %d", tt.input, len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("parsePatterns(%q)[%d] = %q, expected %q", tt.input, i, result[i], expected)
				}
			}
		})
	}
}

func TestValidatePatterns(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		expectError bool
	}{
		{
			name:        "valid patterns",
			patterns:    []string{"*.md", "rules/*.json"},
			expectError: false,
		},
		{
			name:        "empty patterns",
			patterns:    []string{},
			expectError: false,
		},
		{
			name:        "invalid pattern with absolute path",
			patterns:    []string{"/absolute/path/*.md"},
			expectError: true,
		},
		{
			name:        "pattern with parent directory reference",
			patterns:    []string{"../parent/*.md"},
			expectError: true,
		},
		{
			name:        "valid relative patterns",
			patterns:    []string{"subdir/*.md", "**/*.txt"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePatterns(tt.patterns)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// parsePatterns splits a comma-separated pattern string into individual patterns
func parsePatterns(input string) []string {
	if input == "" {
		return []string{}
	}

	parts := strings.Split(input, ",")
	var patterns []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			patterns = append(patterns, trimmed)
		}
	}

	return patterns
}

// validatePatterns checks if patterns are valid (no absolute paths, no parent references)
func validatePatterns(patterns []string) error {
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "/") {
			return fmt.Errorf("pattern cannot be absolute path: %s", pattern)
		}
		if strings.Contains(pattern, "..") {
			return fmt.Errorf("pattern cannot contain parent directory references: %s", pattern)
		}
	}
	return nil
}
