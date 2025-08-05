package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitRegistry(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		apiType string
		wantErr bool
	}{
		{
			name:    "valid github repository",
			url:     "https://github.com/owner/repo",
			apiType: "github",
			wantErr: false,
		},
		{
			name:    "valid gitlab repository",
			url:     "https://gitlab.com/owner/repo",
			apiType: "gitlab",
			wantErr: false,
		},
		{
			name:    "generic git repository",
			url:     "https://git.example.com/owner/repo",
			apiType: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, err := NewGitRegistry("test", tt.url, "", tt.apiType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, registry)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, registry)
				assert.Equal(t, tt.url, registry.URL)
				assert.Equal(t, tt.apiType, registry.APIType)
			}
		})
	}
}

func TestParseReference(t *testing.T) {
	registry, err := NewGitRegistry("test", "https://github.com/owner/repo", "", "github")
	require.NoError(t, err)

	tests := []struct {
		name     string
		ref      string
		expected ReferenceType
		value    string
	}{
		{
			name:     "empty reference defaults to default",
			ref:      "",
			expected: RefTypeDefault,
			value:    "",
		},
		{
			name:     "branch reference",
			ref:      "main",
			expected: RefTypeBranch,
			value:    "main",
		},
		{
			name:     "semver tag",
			ref:      "v1.0.0",
			expected: RefTypeTag,
			value:    "v1.0.0",
		},
		{
			name:     "semver tag without v prefix",
			ref:      "1.0.0",
			expected: RefTypeTag,
			value:    "1.0.0",
		},
		{
			name:     "commit SHA (full)",
			ref:      "abc123def456789012345678901234567890abcd",
			expected: RefTypeCommit,
			value:    "abc123def456789012345678901234567890abcd",
		},
		{
			name:     "commit SHA (short)",
			ref:      "abc1234",
			expected: RefTypeCommit,
			value:    "abc1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitRef := registry.parseReference(tt.ref)
			assert.Equal(t, tt.expected, gitRef.Type)
			assert.Equal(t, tt.value, gitRef.Value)
		})
	}
}

func TestApplyPatterns(t *testing.T) {
	registry, err := NewGitRegistry("test", "https://github.com/owner/repo", "", "github")
	require.NoError(t, err)

	files := []string{
		"rules/typescript.md",
		"rules/react.md",
		"docs/readme.txt",
		"docs/guide.md",
		"src/main.go",
		"test/example.js",
	}

	tests := []struct {
		name     string
		patterns []string
		expected []string
	}{
		{
			name:     "single pattern matching markdown files",
			patterns: []string{"*.md"},
			expected: []string{"rules/typescript.md", "rules/react.md", "docs/guide.md"},
		},
		{
			name:     "pattern matching rules directory",
			patterns: []string{"rules/*"},
			expected: []string{"rules/typescript.md", "rules/react.md"},
		},
		{
			name:     "multiple patterns",
			patterns: []string{"rules/*.md", "docs/*.txt"},
			expected: []string{"rules/typescript.md", "rules/react.md", "docs/readme.txt"},
		},
		{
			name:     "no matches",
			patterns: []string{"*.py"},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.applyPatterns(files, tt.patterns)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"v1.0.0", true},
		{"1.2.3-alpha", true},
		{"1.2.3+build", true},
		{"main", false},
		{"develop", false},
		{"invalid", false},
		{"1.0", true},
		{"1", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := isValidSemver(tt.version)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsCommitSHA(t *testing.T) {
	tests := []struct {
		ref   string
		valid bool
	}{
		{"abc123def456789012345678901234567890abcd", true}, // 40 chars
		{"abc1234", true}, // 7 chars
		{"abc123", false}, // 6 chars (too short)
		{"abc123def456789012345678901234567890abcde", true}, // 41 chars (7+ hex chars)
		{"abc123g", false}, // invalid hex char
		{"ABC1234", true},  // uppercase hex
		{"main", false},    // not hex
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			result := isCommitSHA(tt.ref)
			assert.Equal(t, tt.valid, result)
		})
	}
}
