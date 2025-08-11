package registry

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestMatchesAnyPattern(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		patterns []string
		expected bool
	}{
		{
			name:     "empty_patterns_match_all",
			filePath: "any/file.txt",
			patterns: []string{},
			expected: true,
		},
		{
			name:     "exact_match",
			filePath: "rules/test.md",
			patterns: []string{"rules/test.md"},
			expected: true,
		},
		{
			name:     "wildcard_extension",
			filePath: "rules/test.md",
			patterns: []string{"*.md"},
			expected: true,
		},
		{
			name:     "glob_pattern",
			filePath: "rules/subfolder/test.md",
			patterns: []string{"rules/**/*.md"},
			expected: true,
		},
		{
			name:     "no_match",
			filePath: "rules/test.txt",
			patterns: []string{"*.md"},
			expected: false,
		},
		{
			name:     "multiple_patterns_first_match",
			filePath: "rules/test.md",
			patterns: []string{"*.md", "*.txt"},
			expected: true,
		},
		{
			name:     "multiple_patterns_second_match",
			filePath: "rules/test.txt",
			patterns: []string{"*.md", "*.txt"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesAnyPattern(tt.filePath, tt.patterns)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for file %s with patterns %v",
					tt.expected, result, tt.filePath, tt.patterns)
			}
		})
	}
}

func TestMatchesGlobPattern(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		pattern  string
		expected bool
	}{
		{
			name:     "double_star_recursive",
			filePath: "rules/subfolder/test.md",
			pattern:  "rules/**/*.md",
			expected: true,
		},
		{
			name:     "single_star_filename",
			filePath: "rules/test.md",
			pattern:  "rules/*.md",
			expected: true,
		},
		{
			name:     "single_star_no_slash",
			filePath: "rules/subfolder/test.md",
			pattern:  "rules/*.md",
			expected: false,
		},
		{
			name:     "double_star_deep_nesting",
			filePath: "rules/a/b/c/test.md",
			pattern:  "rules/**/*.md",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesGlobPattern(tt.filePath, tt.pattern)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for file %s with pattern %s",
					tt.expected, result, tt.filePath, tt.pattern)
			}
		})
	}
}

func TestFindMatchingFiles(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "find-matching-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test files
	testFiles := []string{
		"rules/test1.md",
		"rules/test2.txt",
		"rules/subfolder/test3.md",
		"docs/readme.md",
		".git/config",
		".hidden/file.txt",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("test content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name     string
		patterns []string
		expected []string
	}{
		{
			name:     "all_md_files",
			patterns: []string{"*.md"},
			expected: []string{"docs/readme.md", "rules/test1.md", "rules/subfolder/test3.md"},
		},
		{
			name:     "rules_directory_only",
			patterns: []string{"rules/*"},
			expected: []string{"rules/test1.md", "rules/test2.txt"},
		},
		{
			name:     "recursive_rules",
			patterns: []string{"rules/**"},
			expected: []string{"rules/test1.md", "rules/test2.txt", "rules/subfolder/test3.md"},
		},
		{
			name:     "no_hidden_files",
			patterns: []string{"**/*"},
			expected: []string{"docs/readme.md", "rules/test1.md", "rules/test2.txt", "rules/subfolder/test3.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FindMatchingFiles(tempDir, tt.patterns)
			if err != nil {
				t.Fatalf("FindMatchingFiles failed: %v", err)
			}

			// Sort both slices for consistent comparison
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"abc123", true},
		{"ABC123", true},
		{"0123456789abcdef", true},
		{"xyz123", false},
		{"123g", false},
		{"", true}, // Empty string is technically all hex
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsHexString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v for input %s, got %v", tt.expected, tt.input, result)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"valid/path/file.txt", true},
		{"../invalid/path", false},
		{"valid/../invalid", false},
		{"./valid/path", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := ValidatePath(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v for path %s, got %v", tt.expected, tt.path, result)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "copy-file-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create source file
	srcPath := filepath.Join(tempDir, "source.txt")
	srcContent := []byte("test content")
	if err := os.WriteFile(srcPath, srcContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// Copy file
	dstPath := filepath.Join(tempDir, "destination.txt")
	if err := CopyFile(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if !reflect.DeepEqual(srcContent, dstContent) {
		t.Errorf("File content mismatch. Expected %s, got %s", srcContent, dstContent)
	}
}
