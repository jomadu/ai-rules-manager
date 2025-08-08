package version

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Build information - will be injected by ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// GetVersion returns the current ARM version
func GetVersion() string {
	return Version
}

// Resolver interface for version resolution
type Resolver interface {
	Resolve(versionSpec string, availableVersions []string) (string, error)
	Validate(versionSpec string) error
}

// ResolverType represents different version resolution strategies
type ResolverType int

const (
	SemverResolver ResolverType = iota
	GitResolver
	ExactResolver
)

// NewResolver creates a resolver based on the version specification
func NewResolver(versionSpec string) Resolver {
	if isGitVersion(versionSpec) {
		return &gitResolver{}
	}
	if isExactVersion(versionSpec) {
		return &exactResolver{}
	}
	return &semverResolver{}
}

// semverResolver handles semantic version ranges
type semverResolver struct{}

func (r *semverResolver) Resolve(versionSpec string, availableVersions []string) (string, error) {
	constraint, err := semver.NewConstraint(versionSpec)
	if err != nil {
		return "", fmt.Errorf("invalid version constraint '%s': %w", versionSpec, err)
	}

	// Filter and sort available versions
	var validVersions []*semver.Version
	for _, v := range availableVersions {
		// Try parsing with and without 'v' prefix
		cleanV := strings.TrimPrefix(v, "v")
		if version, err := semver.NewVersion(cleanV); err == nil {
			if constraint.Check(version) {
				validVersions = append(validVersions, version)
			}
		}
	}

	if len(validVersions) == 0 {
		return "", fmt.Errorf("no versions satisfy constraint '%s'", versionSpec)
	}

	// Sort and return highest version
	sort.Sort(semver.Collection(validVersions))
	highest := validVersions[len(validVersions)-1]
	return highest.String(), nil
}

func (r *semverResolver) Validate(versionSpec string) error {
	_, err := semver.NewConstraint(versionSpec)
	return err
}

// gitResolver handles Git-specific versions (branches, commits, latest)
type gitResolver struct{}

func (r *gitResolver) Resolve(versionSpec string, availableVersions []string) (string, error) {
	switch versionSpec {
	case "latest":
		// For Git registries, "latest" should resolve to HEAD of default branch
		// This would be handled by the registry implementation
		return "latest", nil
	default:
		// Branch names or commit hashes - validate they exist
		for _, v := range availableVersions {
			if v == versionSpec {
				return versionSpec, nil
			}
		}
		return "", fmt.Errorf("version '%s' not found", versionSpec)
	}
}

func (r *gitResolver) Validate(versionSpec string) error {
	if versionSpec == "latest" {
		return nil
	}
	// Validate branch name or commit hash format
	if len(versionSpec) == 0 {
		return fmt.Errorf("empty version specification")
	}
	// Basic validation - more specific validation would be registry-dependent
	return nil
}

// exactResolver handles exact version matches
type exactResolver struct{}

func (r *exactResolver) Resolve(versionSpec string, availableVersions []string) (string, error) {
	// Remove '=' prefix if present
	targetVersion := strings.TrimPrefix(versionSpec, "=")
	
	for _, v := range availableVersions {
		if v == targetVersion || strings.TrimPrefix(v, "v") == targetVersion {
			return v, nil
		}
	}
	return "", fmt.Errorf("exact version '%s' not found", targetVersion)
}

func (r *exactResolver) Validate(versionSpec string) error {
	targetVersion := strings.TrimPrefix(versionSpec, "=")
	if len(targetVersion) == 0 {
		return fmt.Errorf("empty version specification")
	}
	return nil
}

// Helper functions

func isGitVersion(versionSpec string) bool {
	// Git-specific versions: latest, branch names, commit hashes
	if versionSpec == "latest" {
		return true
	}
	// Check if it looks like a commit hash (hex string, 7-40 chars)
	if matched, _ := regexp.MatchString(`^[a-f0-9]{7,40}$`, versionSpec); matched {
		return true
	}
	// Check if it's a branch name (not a semver pattern)
	if !strings.ContainsAny(versionSpec, "^~>=<") && !regexp.MustCompile(`^\d+\.\d+\.\d+`).MatchString(versionSpec) {
		return true
	}
	return false
}

func isExactVersion(versionSpec string) bool {
	return strings.HasPrefix(versionSpec, "=")
}

// ResolveVersion is a convenience function for version resolution
func ResolveVersion(versionSpec string, availableVersions []string) (string, error) {
	resolver := NewResolver(versionSpec)
	return resolver.Resolve(versionSpec, availableVersions)
}

// ValidateVersionSpec validates a version specification
func ValidateVersionSpec(versionSpec string) error {
	resolver := NewResolver(versionSpec)
	return resolver.Validate(versionSpec)
}
