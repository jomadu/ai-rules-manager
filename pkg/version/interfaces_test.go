package version

import (
	"testing"

	"github.com/jomadu/ai-rules-manager/pkg/registry"
)

func TestSemVerResolver_ResolveVersion(t *testing.T) {
	resolver := NewSemVerResolver()

	tests := []struct {
		name        string
		constraint  string
		available   []registry.VersionRef
		expected    string
		expectError bool
	}{
		{
			name:       "exact version match",
			constraint: "=1.0.0",
			available: []registry.VersionRef{
				{ID: "1.0.0", Type: registry.Tag},
				{ID: "1.1.0", Type: registry.Tag},
			},
			expected:    "1.0.0",
			expectError: false,
		},
		{
			name:       "caret constraint - latest compatible",
			constraint: "^1.0.0",
			available: []registry.VersionRef{
				{ID: "1.0.0", Type: registry.Tag},
				{ID: "1.1.0", Type: registry.Tag},
				{ID: "2.0.0", Type: registry.Tag},
			},
			expected:    "1.1.0",
			expectError: false,
		},
		{
			name:       "tilde constraint - patch updates only",
			constraint: "~1.0.0",
			available: []registry.VersionRef{
				{ID: "1.0.0", Type: registry.Tag},
				{ID: "1.0.1", Type: registry.Tag},
				{ID: "1.1.0", Type: registry.Tag},
			},
			expected:    "1.0.1",
			expectError: false,
		},
		{
			name:       "branch reference",
			constraint: "main",
			available: []registry.VersionRef{
				{ID: "main", Type: registry.Branch},
				{ID: "1.0.0", Type: registry.Tag},
			},
			expected:    "main",
			expectError: false,
		},
		{
			name:        "no matching version",
			constraint:  "=2.0.0",
			available:   []registry.VersionRef{{ID: "1.0.0", Type: registry.Tag}},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.ResolveVersion(tt.constraint, tt.available)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.ID != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.ID)
			}
		})
	}
}

func TestGitContentResolver_ResolveContent(t *testing.T) {
	resolver := NewGitContentResolver()

	tests := []struct {
		name      string
		selector  registry.ContentSelector
		available []registry.File
		expected  []string // file paths
	}{
		{
			name: "pattern matching",
			selector: registry.GitContentSelector{
				Patterns: []string{"rules/amazonq/*.md"},
			},
			available: []registry.File{
				{Path: "rules/amazonq/test.md", Content: []byte("content")},
				{Path: "rules/cursor/test.mdc", Content: []byte("content")},
				{Path: "README.md", Content: []byte("content")},
			},
			expected: []string{"rules/amazonq/test.md"},
		},
		{
			name: "multiple patterns",
			selector: registry.GitContentSelector{
				Patterns: []string{"rules/amazonq/*.md", "rules/cursor/*.mdc"},
			},
			available: []registry.File{
				{Path: "rules/amazonq/test.md", Content: []byte("content")},
				{Path: "rules/cursor/test.mdc", Content: []byte("content")},
				{Path: "README.md", Content: []byte("content")},
			},
			expected: []string{"rules/amazonq/test.md", "rules/cursor/test.mdc"},
		},
		{
			name: "exclusion patterns",
			selector: registry.GitContentSelector{
				Patterns: []string{"rules/**/*.md"},
				Excludes: []string{"**/README.md"},
			},
			available: []registry.File{
				{Path: "rules/amazonq/test.md", Content: []byte("content")},
				{Path: "rules/amazonq/README.md", Content: []byte("content")},
			},
			expected: []string{"rules/amazonq/test.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.ResolveContent(tt.selector, tt.available)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d files, got %d", len(tt.expected), len(result))
				return
			}

			resultPaths := make([]string, len(result))
			for i, file := range result {
				resultPaths[i] = file.Path
			}

			for _, expectedPath := range tt.expected {
				found := false
				for _, resultPath := range resultPaths {
					if resultPath == expectedPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected file %s not found in results", expectedPath)
				}
			}
		})
	}
}
