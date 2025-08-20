package registry

import (
	"context"
	"fmt"
	"strings"

	"github.com/max-dunn/ai-rules-manager/internal/cache"
)

// DownloadResult contains the results of a Git registry download
type DownloadResult struct {
	VersionSpec     string   // Original version spec (e.g., "latest")
	ResolvedVersion string   // Actual commit hash
	Files           []string // Downloaded file paths
}

// GitRegistry implements the Registry interface for Git repositories
type GitRegistry struct {
	*BaseGitRegistry
	operations   GitOperations
	cacheManager cache.GitRegistryCacheManager
}

// NewGitRegistry creates a new Git registry instance
func NewGitRegistry(config *RegistryConfig) (*GitRegistry, error) {
	return NewGitRegistryWithCache(config, nil)
}

// NewGitRegistryWithCache creates a new Git registry instance with cache manager
func NewGitRegistryWithCache(config *RegistryConfig, cacheManager cache.GitRegistryCacheManager) (*GitRegistry, error) {
	base, err := NewBaseGitRegistry(config)
	if err != nil {
		return nil, err
	}

	operations := NewRemoteGitOperations(config)

	return &GitRegistry{
		BaseGitRegistry: base,
		operations:      operations,
		cacheManager:    cacheManager,
	}, nil
}

// GetRulesets returns available rulesets matching the given patterns
func (g *GitRegistry) GetRulesets(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	return g.getRulesetsClone(ctx, patterns)
}

// GetRuleset returns detailed information about a specific ruleset
func (g *GitRegistry) GetRuleset(ctx context.Context, name, version string) (*RulesetInfo, error) {
	return g.getRulesetClone(ctx, name, version)
}

// DownloadRuleset downloads a ruleset to the specified directory (legacy method)
func (g *GitRegistry) DownloadRuleset(ctx context.Context, name, version, destDir string) error {
	// For backward compatibility, use empty patterns
	return g.DownloadRulesetWithPatterns(ctx, name, version, destDir, []string{})
}

// DownloadRulesetWithPatterns downloads a ruleset with pattern matching
func (g *GitRegistry) DownloadRulesetWithPatterns(ctx context.Context, name, version, destDir string, patterns []string) error {
	// Use cache if cache manager is available
	if g.cacheManager != nil {
		repositoryPath, err := g.cacheManager.GetRepositoryPath(g.GetConfig().URL)
		if err == nil {
			return g.BaseGitRegistry.DownloadRulesetWithPatterns(ctx, g.operations, version, repositoryPath, patterns)
		}
	}
	return g.BaseGitRegistry.DownloadRulesetWithPatterns(ctx, g.operations, version, destDir, patterns)
}

// DownloadRulesetWithResult downloads a ruleset and returns structured results
func (g *GitRegistry) DownloadRulesetWithResult(ctx context.Context, name, version, destDir string, patterns []string) (*DownloadResult, error) {
	// Use cache if cache manager is available
	if g.cacheManager != nil {
		return g.downloadWithCache(ctx, name, version, destDir, patterns)
	}
	return g.BaseGitRegistry.DownloadRulesetWithResult(ctx, g.operations, version, destDir, patterns)
}

// GetVersions returns available versions for a ruleset
func (g *GitRegistry) GetVersions(ctx context.Context, name string) ([]string, error) {
	// For Git registries, versions are commit hashes, so we get them from git operations
	return g.operations.ListVersions(ctx)
}

// ResolveVersion resolves a version spec to a concrete commit hash
func (g *GitRegistry) ResolveVersion(ctx context.Context, version string) (string, error) {
	return g.operations.ResolveVersion(ctx, version)
}

// GetType returns the registry type
func (g *GitRegistry) GetType() string {
	return "git"
}

// Search implements the Searcher interface (not supported for git registries)
func (g *GitRegistry) Search(ctx context.Context, query string) ([]SearchResult, error) {
	return nil, fmt.Errorf("search not supported for git registries")
}

// Close cleans up any resources
func (g *GitRegistry) Close() error {
	return nil
}

// getRulesetsClone gets rulesets using git clone
func (g *GitRegistry) getRulesetsClone(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	files, err := g.operations.GetFiles(ctx, "latest", patterns)
	if err != nil {
		return nil, err
	}
	return g.ConvertFilesToRulesets(files, "git"), nil
}

// getRulesetClone gets a specific ruleset using git clone
func (g *GitRegistry) getRulesetClone(ctx context.Context, name, version string) (*RulesetInfo, error) {
	rulesets, err := g.getRulesetsClone(ctx, []string{name + "*"})
	if err != nil {
		return nil, err
	}
	return g.FindRulesetByName(rulesets, name, version)
}

// downloadWithCache downloads a ruleset using the new cache structure
func (g *GitRegistry) downloadWithCache(ctx context.Context, _, versionSpec, destDir string, patterns []string) (*DownloadResult, error) {
	// Resolve git version to commit hash
	resolvedCommit, err := g.operations.ResolveVersion(ctx, versionSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve git version: %w", err)
	}

	// Check if this commit is already cached
	cachedFiles, err := g.cacheManager.GetRuleset(g.GetConfig().URL, patterns, resolvedCommit)
	if err == nil && len(cachedFiles) > 0 {
		// Files are cached, copy to destination if needed
		if !strings.Contains(destDir, "/repository") {
			if err := g.BaseGitRegistry.WriteFilesToDest(cachedFiles, destDir); err != nil {
				return nil, fmt.Errorf("failed to write cached files to destination: %w", err)
			}
		}
		return g.BaseGitRegistry.CreateDownloadResult(versionSpec, resolvedCommit, cachedFiles, destDir), nil
	}

	// Files not cached, download from repository
	repositoryPath, err := g.cacheManager.GetRepositoryPath(g.GetConfig().URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository path: %w", err)
	}

	// Download files using git operations
	files, err := g.downloadFilesFromGit(ctx, repositoryPath, resolvedCommit, patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to download files from git: %w", err)
	}

	// Store files in cache
	if err := g.cacheManager.StoreRuleset(g.GetConfig().URL, patterns, resolvedCommit, files); err != nil {
		return nil, fmt.Errorf("failed to store files in cache: %w", err)
	}

	// Write files to destination if needed
	if !strings.Contains(destDir, "/repository") {
		if err := g.BaseGitRegistry.WriteFilesToDest(files, destDir); err != nil {
			return nil, fmt.Errorf("failed to write files to destination: %w", err)
		}
	}

	return g.BaseGitRegistry.CreateDownloadResult(versionSpec, resolvedCommit, files, destDir), nil
}

// downloadFilesFromGit downloads files from git repository
func (g *GitRegistry) downloadFilesFromGit(ctx context.Context, repositoryPath, version string, patterns []string) (map[string][]byte, error) {
	if len(patterns) == 0 {
		patterns = []string{"**/*"}
	}

	// Use remote git operations to get files at specific path
	if remoteOps, ok := g.operations.(*RemoteGitOperations); ok {
		return remoteOps.GetFilesCloneAt(ctx, repositoryPath, version, patterns)
	}

	// Fallback to regular operations
	return g.operations.GetFiles(ctx, version, patterns)
}
