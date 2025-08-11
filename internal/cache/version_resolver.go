package cache

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// VersionResolver handles version resolution for different registry types
type VersionResolver struct {
	cacheManager Manager
}

// NewVersionResolver creates a new version resolver
func NewVersionResolver(cacheManager Manager) *VersionResolver {
	return &VersionResolver{
		cacheManager: cacheManager,
	}
}

// GitOperations interface for git operations needed by version resolver
type GitOperations interface {
	ResolveVersion(ctx context.Context, version string) (string, error)
	ListVersions(ctx context.Context) ([]string, error)
}

// ResolveGitVersion resolves a git version spec to a commit hash
func (vr *VersionResolver) ResolveGitVersion(ctx context.Context, gitOps GitOperations, versionSpec string) (resolvedVersion string, mappings map[string]string, err error) {
	// For git registries, always resolve to commit hash
	resolvedCommit, err := gitOps.ResolveVersion(ctx, versionSpec)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve git version %s: %w", versionSpec, err)
	}

	// Create mapping from version spec to commit
	versionMappings := map[string]string{
		versionSpec: resolvedCommit,
	}

	// If this is a branch or tag, also check if other refs point to the same commit
	if versionSpec != resolvedCommit {
		allVersions, err := gitOps.ListVersions(ctx)
		if err == nil {
			for _, version := range allVersions {
				if version != versionSpec {
					if commit, err := gitOps.ResolveVersion(ctx, version); err == nil && commit == resolvedCommit {
						versionMappings[version] = resolvedCommit
					}
				}
			}
		}
	}

	return resolvedCommit, versionMappings, nil
}

// ResolveSemverVersion resolves a semver constraint to a specific version
func (vr *VersionResolver) ResolveSemverVersion(availableVersions []string, constraint string) (string, error) {
	if constraint == "latest" {
		return vr.getLatestSemver(availableVersions)
	}

	// Parse constraint
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		// If not a valid constraint, try exact match
		for _, version := range availableVersions {
			if version == constraint {
				return version, nil
			}
		}
		return "", fmt.Errorf("invalid version constraint: %s", constraint)
	}

	// Find matching versions
	var matchingVersions []*semver.Version
	for _, versionStr := range availableVersions {
		version, err := semver.NewVersion(versionStr)
		if err != nil {
			continue // Skip invalid semver versions
		}
		if c.Check(version) {
			matchingVersions = append(matchingVersions, version)
		}
	}

	if len(matchingVersions) == 0 {
		return "", fmt.Errorf("no versions match constraint: %s", constraint)
	}

	// Sort and return highest matching version
	sort.Sort(semver.Collection(matchingVersions))
	return matchingVersions[len(matchingVersions)-1].String(), nil
}

// getLatestSemver returns the latest semantic version from a list
func (vr *VersionResolver) getLatestSemver(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions available")
	}

	var semverVersions []*semver.Version
	for _, versionStr := range versions {
		version, err := semver.NewVersion(versionStr)
		if err != nil {
			continue // Skip invalid semver versions
		}
		semverVersions = append(semverVersions, version)
	}

	if len(semverVersions) == 0 {
		// If no valid semver versions, return the last one lexicographically
		sort.Strings(versions)
		return versions[len(versions)-1], nil
	}

	// Sort and return highest version
	sort.Sort(semver.Collection(semverVersions))
	return semverVersions[len(semverVersions)-1].String(), nil
}

// IsGitCommitHash checks if a string looks like a git commit hash
func (vr *VersionResolver) IsGitCommitHash(version string) bool {
	// Git commit hashes are 40 characters of hexadecimal (full hash)
	// or 7+ characters of hexadecimal (abbreviated hash)
	if len(version) < 7 || len(version) > 40 {
		return false
	}

	matched, _ := regexp.MatchString("^[a-f0-9]+$", version)
	return matched
}

// IsGitBranchOrTag checks if a version spec looks like a branch or tag name
func (vr *VersionResolver) IsGitBranchOrTag(version string) bool {
	// Common branch/tag patterns
	if version == "latest" || version == "main" || version == "master" || version == "develop" {
		return true
	}

	// Version tags (v1.0.0, 1.0.0, etc.)
	if strings.HasPrefix(version, "v") && vr.isSemverLike(strings.TrimPrefix(version, "v")) {
		return true
	}

	if vr.isSemverLike(version) {
		return true
	}

	// Other branch-like names
	if matched, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9._/-]*$", version); matched {
		return true
	}

	return false
}

// isSemverLike checks if a string looks like a semantic version
func (vr *VersionResolver) isSemverLike(version string) bool {
	_, err := semver.NewVersion(version)
	return err == nil
}

// NormalizeVersionForCache normalizes a version for cache storage
func (vr *VersionResolver) NormalizeVersionForCache(registryType, version string) string {
	if registryType == "git" {
		// For git, always use commit hashes as normalized versions
		// This should be called after resolution
		return version
	}

	// For other registry types, use the version as-is
	return version
}

// GetVersionDisplayName returns a user-friendly display name for a version
func (vr *VersionResolver) GetVersionDisplayName(registryType, version string, mappings map[string]string) string {
	if registryType != "git" || mappings == nil {
		return version
	}

	// For git registries, show commit hash with original references
	var aliases []string
	for originalRef, commitHash := range mappings {
		if commitHash == version && originalRef != version {
			aliases = append(aliases, originalRef)
		}
	}

	if len(aliases) == 0 {
		return version
	}

	sort.Strings(aliases)
	return fmt.Sprintf("%s (%s)", version, strings.Join(aliases, ", "))
}
