package registry

import (
	"testing"
)

func TestGitContentSelector_String(t *testing.T) {
	tests := []struct {
		name     string
		selector GitContentSelector
		expected string
	}{
		{
			name: "patterns only",
			selector: GitContentSelector{
				Patterns: []string{"rules/*.md", "docs/*.txt"},
			},
			expected: "patterns:[rules/*.md docs/*.txt],excludes:[]",
		},
		{
			name: "patterns with excludes",
			selector: GitContentSelector{
				Patterns: []string{"rules/*.md"},
				Excludes: []string{"**/test.md"},
			},
			expected: "patterns:[rules/*.md],excludes:[**/test.md]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.selector.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGitContentSelector_Validate(t *testing.T) {
	tests := []struct {
		name        string
		selector    GitContentSelector
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid patterns",
			selector: GitContentSelector{
				Patterns: []string{"rules/*.md", "docs/*.txt"},
			},
			expectError: false,
		},
		{
			name: "empty patterns",
			selector: GitContentSelector{
				Patterns: []string{},
			},
			expectError: true,
			errorMsg:    "at least one pattern is required",
		},
		{
			name: "nil patterns",
			selector: GitContentSelector{
				Patterns: nil,
			},
			expectError: true,
			errorMsg:    "at least one pattern is required",
		},
		{
			name: "valid with excludes",
			selector: GitContentSelector{
				Patterns: []string{"rules/*.md"},
				Excludes: []string{"**/test.md"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.selector.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGitRegistry_GetMetadata(t *testing.T) {
	registry := NewGitRegistry("https://github.com/user/repo")

	metadata := registry.GetMetadata()

	if metadata.URL != "https://github.com/user/repo" {
		t.Errorf("expected URL https://github.com/user/repo, got %s", metadata.URL)
	}

	if metadata.Type != "git" {
		t.Errorf("expected type git, got %s", metadata.Type)
	}
}

func TestGitRegistry_ListVersions(t *testing.T) {
	registry := NewGitRegistry("https://github.com/user/repo")

	// Test expected behavior: should return list of version references
	versions, err := registry.ListVersions()

	// Should not error for valid repository
	if err != nil {
		t.Errorf("ListVersions should not error for valid repo, got: %v", err)
	}

	// Should return a slice (even if empty for stub)
	if versions == nil {
		t.Errorf("ListVersions should return non-nil slice")
	}

	// TODO: Once implemented, verify version references have proper fields
	for _, version := range versions {
		if version.ID == "" {
			t.Errorf("Version ID should not be empty")
		}
		if version.Type < Tag || version.Type > Label {
			t.Errorf("Version type should be valid VersionRefType")
		}
	}
}

func TestGitRegistry_GetContent(t *testing.T) {
	registry := NewGitRegistry("https://github.com/user/repo")

	versionRef := VersionRef{ID: "1.0.0", Type: Tag}
	selector := GitContentSelector{Patterns: []string{"*.md"}}

	// Test expected behavior: should return matching files
	files, err := registry.GetContent(versionRef, selector)

	// Should not error for valid inputs
	if err != nil {
		t.Errorf("GetContent should not error for valid inputs, got: %v", err)
	}

	// Should return a slice (even if empty for stub)
	if files == nil {
		t.Errorf("GetContent should return non-nil slice")
	}

	// TODO: Once implemented, verify returned files have proper structure
	for _, file := range files {
		if file.Path == "" {
			t.Errorf("File path should not be empty")
		}
		if file.Content == nil {
			t.Errorf("File content should not be nil")
		}
		if file.Size < 0 {
			t.Errorf("File size should not be negative")
		}
	}
}
