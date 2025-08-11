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
	cacheManager cache.Manager
}

// NewGitRegistry creates a new Git registry instance
func NewGitRegistry(config *RegistryConfig, auth *AuthConfig) (*GitRegistry, error) {
	return NewGitRegistryWithCache(config, auth, nil)
}

// NewGitRegistryWithCache creates a new Git registry instance with cache manager
func NewGitRegistryWithCache(config *RegistryConfig, auth *AuthConfig, cacheManager cache.Manager) (*GitRegistry, error) {
	base, err := NewBaseGitRegistry(config, auth)
	if err != nil {
		return nil, err
	}

	operations := NewRemoteGitOperations(config, auth)

	return &GitRegistry{
		BaseGitRegistry: base,
		operations:      operations,
		cacheManager:    cacheManager,
	}, nil
}

// GetRulesets returns available rulesets matching the given patterns
func (g *GitRegistry) GetRulesets(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	if g.GetAuth().APIType == "github" {
		return g.getRulesetsAPI(ctx, patterns)
	}
	return g.getRulesetsClone(ctx, patterns)
}

// GetRuleset returns detailed information about a specific ruleset
func (g *GitRegistry) GetRuleset(ctx context.Context, name, version string) (*RulesetInfo, error) {
	if g.GetAuth().APIType == "github" {
		return g.getRulesetAPI(ctx, name, version)
	}
	return g.getRulesetClone(ctx, name, version)
}

// DownloadRuleset downloads a ruleset to the specified directory (legacy method)
func (g *GitRegistry) DownloadRuleset(ctx context.Context, name, version, destDir string) error {
	// For backward compatibility, use empty patterns
	return g.DownloadRulesetWithPatterns(ctx, name, version, destDir, []string{})
}

// DownloadRulesetWithPatterns downloads a ruleset with pattern matching
func (g *GitRegistry) DownloadRulesetWithPatterns(ctx context.Context, name, version, destDir string, patterns []string) error {
	// Use cache if cache manager is available (enabled check already done in factory)
	if g.cacheManager != nil {
		repositoryPath, err := g.getRepositoryPath()
		if err == nil {
			if err := g.cacheManager.EnsureCacheDir(g.GetType(), g.GetConfig().URL); err == nil {
				return g.BaseGitRegistry.DownloadRulesetWithPatterns(ctx, g.operations, version, repositoryPath, patterns)
			}
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
	// Try to get versions from cache first
	if g.cacheManager != nil {
		metadataManager := g.cacheManager.GetMetadataManager()
		if versions, _, err := metadataManager.GetVersions(g.GetType(), g.GetConfig().URL, name); err == nil {
			return versions, nil
		}
	}

	// Fallback to git operations
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

// getRulesetsAPI gets rulesets using GitHub API
func (g *GitRegistry) getRulesetsAPI(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	files, err := g.operations.GetFiles(ctx, "latest", patterns)
	if err != nil {
		return nil, err
	}
	return g.ConvertFilesToRulesets(files, "git"), nil
}

// getRulesetsClone gets rulesets using git clone
func (g *GitRegistry) getRulesetsClone(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	files, err := g.operations.GetFiles(ctx, "latest", patterns)
	if err != nil {
		return nil, err
	}
	return g.ConvertFilesToRulesets(files, "git"), nil
}

// getRulesetAPI gets a specific ruleset using GitHub API
func (g *GitRegistry) getRulesetAPI(ctx context.Context, name, version string) (*RulesetInfo, error) {
	rulesets, err := g.getRulesetsAPI(ctx, []string{name + "*"})
	if err != nil {
		return nil, err
	}
	return g.FindRulesetByName(rulesets, name, version)
}

// getRulesetClone gets a specific ruleset using git clone
func (g *GitRegistry) getRulesetClone(ctx context.Context, name, version string) (*RulesetInfo, error) {
	rulesets, err := g.getRulesetsClone(ctx, []string{name + "*"})
	if err != nil {
		return nil, err
	}
	return g.FindRulesetByName(rulesets, name, version)
}

// getCachePath returns the content-based cache path for this registry
// getCachePath is deprecated, use cache manager directly
/*
func (g *GitRegistry) getCachePath() (string, error) {
	if g.cacheManager == nil {
		return "", fmt.Errorf("cache manager not available")
	}
	return g.cacheManager.GetCachePath(g.GetType(), g.GetConfig().URL)
}
*/

// getRepositoryPath returns the repository subdirectory path for git clones
func (g *GitRegistry) getRepositoryPath() (string, error) {
	if g.cacheManager == nil {
		return "", fmt.Errorf("cache manager not available")
	}
	return g.cacheManager.GetRepositoryPath(g.GetType(), g.GetConfig().URL)
}

// downloadWithCache downloads a ruleset using the enhanced cache structure
func (g *GitRegistry) downloadWithCache(ctx context.Context, rulesetName, versionSpec, destDir string, patterns []string) (*DownloadResult, error) {
	// Ensure cache directory structure exists
	if err := g.cacheManager.EnsureCacheDir(g.GetType(), g.GetConfig().URL); err != nil {
		return nil, fmt.Errorf("failed to ensure cache directory: %w", err)
	}

	// Create version resolver
	versionResolver := cache.NewVersionResolver(g.cacheManager)

	// Resolve git version to commit hash
	resolvedCommit, mappings, err := versionResolver.ResolveGitVersion(ctx, g.operations, versionSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve git version: %w", err)
	}

	// Get ruleset storage manager
	rulesetStorage := g.cacheManager.GetRulesetStorage()

	// Check if this commit is already cached
	cachedFiles, err := rulesetStorage.GetRulesetFiles(g.GetType(), g.GetConfig().URL, rulesetName, resolvedCommit, patterns)
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
	repositoryPath, err := g.getRepositoryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository path: %w", err)
	}

	// Download files using git operations
	files, err := g.downloadFilesFromGit(ctx, repositoryPath, resolvedCommit, patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to download files from git: %w", err)
	}

	// Store files in cache with ruleset/version structure
	if err := rulesetStorage.StoreRulesetFiles(g.GetType(), g.GetConfig().URL, rulesetName, resolvedCommit, files, patterns); err != nil {
		return nil, fmt.Errorf("failed to store files in cache: %w", err)
	}

	// Update metadata
	if err := g.updateCacheMetadata(rulesetName, resolvedCommit, mappings, files); err != nil {
		// Log error but don't fail the download
		_ = err
	}

	// Write files to destination if needed
	if !strings.Contains(destDir, "/repository") {
		if err := g.BaseGitRegistry.WriteFilesToDest(files, destDir); err != nil {
			return nil, fmt.Errorf("failed to write files to destination: %w", err)
		}
	}

	// Update cache info
	_ = g.cacheManager.UpdateCacheInfo(g.GetType(), g.GetConfig().URL, resolvedCommit)

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

// updateCacheMetadata updates versions.json and metadata.json
func (g *GitRegistry) updateCacheMetadata(rulesetName, resolvedCommit string, mappings map[string]string, files map[string][]byte) error {
	metadataManager := g.cacheManager.GetMetadataManager()

	// Calculate file stats
	fileCount := len(files)
	var totalSize int64
	for _, content := range files {
		totalSize += int64(len(content))
	}

	// Get existing versions or create new list
	existingVersions, existingMappings, err := metadataManager.GetVersions(g.GetType(), g.GetConfig().URL, rulesetName)
	if err != nil {
		// First time caching this ruleset
		existingVersions = []string{}
		existingMappings = make(map[string]string)
	}

	// Add resolved commit if not already present
	commitExists := false
	for _, version := range existingVersions {
		if version == resolvedCommit {
			commitExists = true
			break
		}
	}
	if !commitExists {
		existingVersions = append(existingVersions, resolvedCommit)
	}

	// Merge mappings
	for versionSpec, commit := range mappings {
		existingMappings[versionSpec] = commit
	}

	// Update versions.json
	if err := metadataManager.UpdateVersions(g.GetType(), g.GetConfig().URL, rulesetName, existingVersions, existingMappings); err != nil {
		return fmt.Errorf("failed to update versions cache: %w", err)
	}

	// Update metadata.json
	if err := metadataManager.UpdateMetadata(g.GetType(), g.GetConfig().URL, rulesetName, resolvedCommit, fileCount, totalSize); err != nil {
		return fmt.Errorf("failed to update metadata cache: %w", err)
	}

	return nil
}
