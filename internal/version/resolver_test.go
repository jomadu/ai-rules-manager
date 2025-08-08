package version

import (
	"fmt"
	"testing"
)

func TestSemverResolver(t *testing.T) {
	resolver := &semverResolver{}
	availableVersions := []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0", "2.1.0", "v1.0.5", "v1.1.5"}

	tests := []struct {
		name        string
		versionSpec string
		expected    string
		shouldError bool
	}{
		{"caret range", "^1.0.0", "1.2.0", false},
		{"tilde range", "~1.1.0", "1.1.5", false},
		{"greater than equal", ">=1.1.0", "2.1.0", false},
		{"less than", "<2.0.0", "1.2.0", false},
		{"greater than", ">1.0.0", "2.1.0", false},
		{"less than equal", "<=1.1.0", "1.1.0", false},
		{"exact match", "1.2.0", "1.2.0", false},
		{"no match", "^3.0.0", "", true},
		{"invalid constraint", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.versionSpec, availableVersions)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for %s, got result: %s", tt.versionSpec, result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %s: %v", tt.versionSpec, err)
				}
				if result != tt.expected {
					t.Errorf("expected %s, got %s for spec %s", tt.expected, result, tt.versionSpec)
				}
			}
		})
	}
}

func TestGitResolver(t *testing.T) {
	resolver := &gitResolver{}
	availableVersions := []string{"main", "develop", "feature-branch", "abc123def", "1a2b3c4d5e6f7890"}

	tests := []struct {
		name        string
		versionSpec string
		expected    string
		shouldError bool
	}{
		{"latest", "latest", "latest", false},
		{"main branch", "main", "main", false},
		{"develop branch", "develop", "develop", false},
		{"feature branch", "feature-branch", "feature-branch", false},
		{"short commit hash", "abc123def", "abc123def", false},
		{"long commit hash", "1a2b3c4d5e6f7890", "1a2b3c4d5e6f7890", false},
		{"nonexistent branch", "nonexistent", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.versionSpec, availableVersions)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for %s, got result: %s", tt.versionSpec, result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %s: %v", tt.versionSpec, err)
				}
				if result != tt.expected {
					t.Errorf("expected %s, got %s for spec %s", tt.expected, result, tt.versionSpec)
				}
			}
		})
	}
}

func TestExactResolver(t *testing.T) {
	resolver := &exactResolver{}
	availableVersions := []string{"1.0.0", "1.1.0", "v1.2.0", "2.0.0"}

	tests := []struct {
		name        string
		versionSpec string
		expected    string
		shouldError bool
	}{
		{"exact match", "=1.0.0", "1.0.0", false},
		{"exact match with v prefix", "=1.2.0", "v1.2.0", false},
		{"exact match no equals", "1.1.0", "1.1.0", false},
		{"no match", "=1.5.0", "", true},
		{"empty spec", "=", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.versionSpec, availableVersions)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for %s, got result: %s", tt.versionSpec, result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %s: %v", tt.versionSpec, err)
				}
				if result != tt.expected {
					t.Errorf("expected %s, got %s for spec %s", tt.expected, result, tt.versionSpec)
				}
			}
		})
	}
}

func TestNewResolver(t *testing.T) {
	tests := []struct {
		name        string
		versionSpec string
		expected    string
	}{
		{"semver caret", "^1.0.0", "*version.semverResolver"},
		{"semver tilde", "~1.0.0", "*version.semverResolver"},
		{"semver comparison", ">=1.0.0", "*version.semverResolver"},
		{"semver exact", "1.0.0", "*version.semverResolver"},
		{"git latest", "latest", "*version.gitResolver"},
		{"git branch", "main", "*version.gitResolver"},
		{"git commit", "abc123def", "*version.gitResolver"},
		{"exact with equals", "=1.0.0", "*version.exactResolver"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(tt.versionSpec)
			resolverType := fmt.Sprintf("%T", resolver)
			if resolverType != tt.expected {
				t.Errorf("expected %s, got %s for spec %s", tt.expected, resolverType, tt.versionSpec)
			}
		})
	}
}

func TestIsGitVersion(t *testing.T) {
	tests := []struct {
		name        string
		versionSpec string
		expected    bool
	}{
		{"latest", "latest", true},
		{"branch name", "main", true},
		{"feature branch", "feature-branch", true},
		{"short commit", "abc123d", true},
		{"long commit", "abc123def0123456789abcdef01234567", true},
		{"semver", "1.0.0", false},
		{"semver caret", "^1.0.0", false},
		{"semver tilde", "~1.0.0", false},
		{"comparison", ">=1.0.0", false},
		{"exact", "=1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGitVersion(tt.versionSpec)
			if result != tt.expected {
				t.Errorf("expected %t, got %t for spec %s", tt.expected, result, tt.versionSpec)
			}
		})
	}
}

func TestIsExactVersion(t *testing.T) {
	tests := []struct {
		name        string
		versionSpec string
		expected    bool
	}{
		{"exact with equals", "=1.0.0", true},
		{"exact empty", "=", true},
		{"semver", "1.0.0", false},
		{"caret", "^1.0.0", false},
		{"latest", "latest", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExactVersion(tt.versionSpec)
			if result != tt.expected {
				t.Errorf("expected %t, got %t for spec %s", tt.expected, result, tt.versionSpec)
			}
		})
	}
}

func TestResolveVersion(t *testing.T) {
	availableVersions := []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0", "main", "develop"}

	tests := []struct {
		name        string
		versionSpec string
		expected    string
		shouldError bool
	}{
		{"semver resolution", "^1.0.0", "1.2.0", false},
		{"git resolution", "main", "main", false},
		{"exact resolution", "=1.1.0", "1.1.0", false},
		{"invalid git", "invalid", "", true},
		{"nonexistent git", "nonexistent", "", true},
		{"nonexistent exact", "=3.0.0", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveVersion(tt.versionSpec, availableVersions)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for %s, got result: %s", tt.versionSpec, result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %s: %v", tt.versionSpec, err)
				}
				if result != tt.expected {
					t.Errorf("expected %s, got %s for spec %s", tt.expected, result, tt.versionSpec)
				}
			}
		})
	}
}

func TestValidateVersionSpec(t *testing.T) {
	tests := []struct {
		name        string
		versionSpec string
		shouldError bool
	}{
		{"valid semver", "^1.0.0", false},
		{"valid git", "main", false},
		{"valid exact", "=1.0.0", false},
		{"valid latest", "latest", false},
		{"invalid git", "invalid", false},
		{"empty exact", "=", true},
		{"empty git", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersionSpec(tt.versionSpec)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for %s", tt.versionSpec)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %s: %v", tt.versionSpec, err)
				}
			}
		})
	}
}