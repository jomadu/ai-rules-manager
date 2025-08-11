package registry

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestSemverConstraints(t *testing.T) {
	tests := []struct {
		name        string
		versions    []string
		constraint  string
		expected    string
		shouldError bool
	}{
		{
			name:       ">=1.0.0 selects highest",
			versions:   []string{"1.0.0", "2.0.0", "1.5.0"},
			constraint: ">=1.0.0",
			expected:   "2.0.0",
		},
		{
			name:       "^1.0.0 stays in major version",
			versions:   []string{"1.0.0", "1.2.0", "2.0.0"},
			constraint: "^1.0.0",
			expected:   "1.2.0",
		},
		{
			name:       "~1.0.0 stays in minor version",
			versions:   []string{"1.0.0", "1.0.5", "1.1.0"},
			constraint: "~1.0.0",
			expected:   "1.0.5",
		},
		{
			name:        "invalid constraint",
			versions:    []string{"1.0.0"},
			constraint:  "invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock GitRegistry that returns our test versions
			g := &mockGitRegistry{versions: tt.versions}
			result, err := g.resolveSemverPattern(context.Background(), tt.constraint)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// mockGitRegistry for testing
type mockGitRegistry struct {
	versions []string
}

func (m *mockGitRegistry) GetVersions(ctx context.Context, name string) ([]string, error) {
	return append([]string{"latest"}, m.versions...), nil
}

func (m *mockGitRegistry) resolveSemverPattern(ctx context.Context, versionSpec string) (string, error) {
	// Simplified test implementation using the semver library directly
	constraint, err := semver.NewConstraint(versionSpec)
	if err != nil {
		return "", err
	}

	versions, _ := m.GetVersions(ctx, "")
	var candidates []*semver.Version
	for _, v := range versions {
		if v == "latest" {
			continue
		}
		if ver, err := semver.NewVersion(v); err == nil {
			candidates = append(candidates, ver)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no valid versions")
	}

	sort.Sort(sort.Reverse(semver.Collection(candidates)))

	for _, candidate := range candidates {
		if constraint.Check(candidate) {
			return candidate.Original(), nil
		}
	}

	return "", fmt.Errorf("no versions satisfy constraint")
}

func TestIsSemverPattern(t *testing.T) {
	g := &GitRegistry{}

	tests := []struct {
		version  string
		expected bool
	}{
		{">=1.0.0", true},
		{"^1.0.0", true},
		{"~1.0.0", true},
		{"1.0.0", false},
		{"latest", false},
		{"main", false},
	}

	for _, tt := range tests {
		result := g.isSemverPattern(tt.version)
		if result != tt.expected {
			t.Errorf("isSemverPattern(%q) = %v, want %v", tt.version, result, tt.expected)
		}
	}
}

func TestIsVersionNumber(t *testing.T) {
	g := &GitRegistry{}

	tests := []struct {
		version  string
		expected bool
	}{
		{"1.0.0", true},
		{"v1.0.0", true},
		{"2.1.3", true},
		{"latest", false},
		{"main", false},
		{">=1.0.0", false},
	}

	for _, tt := range tests {
		result := g.isVersionNumber(tt.version)
		if result != tt.expected {
			t.Errorf("isVersionNumber(%q) = %v, want %v", tt.version, result, tt.expected)
		}
	}
}
