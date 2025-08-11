package cache

import (
	"strings"
	"testing"
)

func TestURLNormalizer_NormalizeGitURL(t *testing.T) {
	normalizer := NewURLNormalizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "GitHub HTTPS URL",
			input:    "https://github.com/user/repo",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "GitHub HTTPS URL with .git",
			input:    "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "GitHub SSH URL",
			input:    "git@github.com:user/repo",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "GitHub SSH URL with .git",
			input:    "git@github.com:user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "GitLab SSH URL",
			input:    "git@gitlab.com:user/repo",
			expected: "https://gitlab.com/user/repo",
		},
		{
			name:     "Bitbucket SSH URL",
			input:    "git@bitbucket.org:user/repo",
			expected: "https://bitbucket.org/user/repo",
		},
		{
			name:     "Generic SSH URL",
			input:    "git@example.com:user/repo",
			expected: "https://example.com/user/repo",
		},
		{
			name:     "URL with trailing slash",
			input:    "https://github.com/user/repo/",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "Mixed case URL",
			input:    "HTTPS://GitHub.com/User/Repo",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "Domain without protocol",
			input:    "github.com/user/repo",
			expected: "https://github.com/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizer.normalizeGitURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeGitURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestURLNormalizer_NormalizeGitLabURL(t *testing.T) {
	normalizer := NewURLNormalizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "GitLab project URL",
			input:    "https://gitlab.example.com/projects/123",
			expected: "https://gitlab.example.com/projects/123",
		},
		{
			name:     "GitLab group URL",
			input:    "https://gitlab.example.com/groups/456",
			expected: "https://gitlab.example.com/groups/456",
		},
		{
			name:     "URL with trailing slash",
			input:    "https://gitlab.example.com/projects/789/",
			expected: "https://gitlab.example.com/projects/789",
		},
		{
			name:     "Mixed case URL",
			input:    "HTTPS://GitLab.Example.com/Projects/123",
			expected: "https://gitlab.example.com/projects/123",
		},
		{
			name:     "URL without protocol",
			input:    "gitlab.example.com/projects/123",
			expected: "https://gitlab.example.com/projects/123",
		},
		{
			name:     "URL with multiple slashes",
			input:    "https://gitlab.example.com//projects//123",
			expected: "https://gitlab.example.com/projects/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizer.normalizeGitLabURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeGitLabURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestURLNormalizer_NormalizeS3URL(t *testing.T) {
	normalizer := NewURLNormalizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple bucket name",
			input:    "my-bucket",
			expected: "my-bucket",
		},
		{
			name:     "Bucket with prefix",
			input:    "my-bucket/rules",
			expected: "my-bucket/rules",
		},
		{
			name:     "Bucket with nested prefix",
			input:    "my-bucket/path/to/rules",
			expected: "my-bucket/path/to/rules",
		},
		{
			name:     "S3 URL with protocol",
			input:    "s3://my-bucket/rules",
			expected: "my-bucket/rules",
		},
		{
			name:     "URL with trailing slash",
			input:    "my-bucket/rules/",
			expected: "my-bucket/rules",
		},
		{
			name:     "URL with backslashes",
			input:    "my-bucket\\rules\\subdir",
			expected: "my-bucket/rules/subdir",
		},
		{
			name:     "URL with multiple slashes",
			input:    "my-bucket//rules//subdir",
			expected: "my-bucket/rules/subdir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizer.normalizeS3URL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeS3URL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestURLNormalizer_NormalizeHTTPSURL(t *testing.T) {
	normalizer := NewURLNormalizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple HTTPS URL",
			input:    "https://example.com/registry",
			expected: "https://example.com/registry",
		},
		{
			name:     "URL with trailing slash",
			input:    "https://example.com/registry/",
			expected: "https://example.com/registry",
		},
		{
			name:     "Mixed case URL",
			input:    "HTTPS://Example.COM/Registry",
			expected: "https://example.com/Registry",
		},
		{
			name:     "URL without protocol",
			input:    "example.com/registry",
			expected: "https://example.com/registry",
		},
		{
			name:     "HTTP URL",
			input:    "http://example.com/registry",
			expected: "http://example.com/registry",
		},
		{
			name:     "URL with query parameters",
			input:    "https://example.com/registry?param=value",
			expected: "https://example.com/registry?param=value",
		},
		{
			name:     "URL with fragment",
			input:    "https://example.com/registry#fragment",
			expected: "https://example.com/registry",
		},
		{
			name:     "Root path",
			input:    "https://example.com/",
			expected: "https://example.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizer.normalizeHTTPSURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeHTTPSURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestURLNormalizer_NormalizeLocalURL(t *testing.T) {
	normalizer := NewURLNormalizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Absolute path",
			input:    "/path/to/registry",
			expected: "/path/to/registry",
		},
		{
			name:     "Relative path",
			input:    "./registry",
			expected: "./registry", // Will be cleaned by filepath.Clean
		},
		{
			name:     "File URL",
			input:    "file:///path/to/registry",
			expected: "/path/to/registry",
		},
		{
			name:     "Path with redundant elements",
			input:    "/path/./to/../to/registry",
			expected: "/path/to/registry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizer.normalizeLocalURL(tt.input)
			// For relative paths, we need to handle the fact that filepath.Abs will resolve them
			if tt.input == "./registry" {
				// Just check that it's been cleaned and uses forward slashes
				if !strings.HasSuffix(result, "/registry") {
					t.Errorf("normalizeLocalURL(%q) = %q, expected to end with '/registry'", tt.input, result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("normalizeLocalURL(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestURLNormalizer_NormalizeURL(t *testing.T) {
	normalizer := NewURLNormalizer()

	tests := []struct {
		name         string
		registryType string
		input        string
		expected     string
	}{
		{
			name:         "Git registry",
			registryType: "git",
			input:        "git@github.com:user/repo.git",
			expected:     "https://github.com/user/repo",
		},
		{
			name:         "GitLab registry",
			registryType: "gitlab",
			input:        "https://gitlab.example.com/projects/123/",
			expected:     "https://gitlab.example.com/projects/123",
		},
		{
			name:         "S3 registry",
			registryType: "s3",
			input:        "s3://my-bucket/rules/",
			expected:     "my-bucket/rules",
		},
		{
			name:         "HTTPS registry",
			registryType: "https",
			input:        "HTTPS://Example.com/Registry/",
			expected:     "https://example.com/Registry",
		},
		{
			name:         "Local registry",
			registryType: "local",
			input:        "/path/./to/../to/registry",
			expected:     "/path/to/registry",
		},
		{
			name:         "Unknown registry type",
			registryType: "unknown",
			input:        "SOME://Example.com//Path/",
			expected:     "some://example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizer.NormalizeURL(tt.registryType, tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeURL(%q, %q) = %q, want %q", tt.registryType, tt.input, result, tt.expected)
			}
		})
	}
}
