package registry

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// RemoteGitOperations implements GitOperations for remote Git repositories
type RemoteGitOperations struct {
	config *RegistryConfig
	client *http.Client
}

// NewRemoteGitOperations creates a new remote Git operations instance
func NewRemoteGitOperations(config *RegistryConfig) *RemoteGitOperations {
	return &RemoteGitOperations{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// ResolveVersion resolves a version spec to a concrete commit hash
func (r *RemoteGitOperations) ResolveVersion(ctx context.Context, constraint string) (string, error) {
	if constraint == "latest" {
		return r.resolveLatestClone(ctx)
	}

	// Check if it's a commit hash (40 hex characters)
	if len(constraint) == 40 && IsHexString(constraint) {
		return constraint, nil
	}

	// Check if it's a semver pattern
	if IsSemverPattern(constraint) {
		return r.resolveSemverPattern(ctx, constraint)
	}

	// Check if it looks like a version number - for git registries, resolve to commit hash
	if IsVersionNumber(constraint) {
		return r.resolveTagToCommitClone(ctx, constraint)
	}

	// Assume it's a branch name and resolve to commit hash
	return r.resolveBranchClone(ctx, constraint)
}

// ListVersions returns available versions for the repository
func (r *RemoteGitOperations) ListVersions(ctx context.Context) ([]string, error) {
	return r.getVersionsClone(ctx)
}

// GetFiles retrieves files matching patterns for a specific version
func (r *RemoteGitOperations) GetFiles(ctx context.Context, version string, patterns []string) (map[string][]byte, error) {
	return r.getFilesClone(ctx, version, patterns)
}

// Clone-based implementations

func (r *RemoteGitOperations) resolveLatestClone(ctx context.Context) (string, error) {
	// For "latest", resolve to HEAD of default branch, not latest tag
	return r.resolveDefaultBranchClone(ctx)
}

func (r *RemoteGitOperations) resolveBranchClone(ctx context.Context, branch string) (string, error) {
	repoDir, err := r.getCachedRepository(ctx)
	if err != nil {
		return "", &GitError{Operation: "resolve_branch", Repo: r.config.URL, Version: branch, Cause: err}
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", &GitError{Operation: "resolve_branch", Repo: r.config.URL, Version: branch, Cause: err}
	}

	// Try to get branch reference
	branchRef, err := repo.Reference(plumbing.ReferenceName("refs/heads/"+branch), true)
	if err != nil {
		// Branch not found, try as remote branch
		branchRef, err = repo.Reference(plumbing.ReferenceName("refs/remotes/origin/"+branch), true)
		if err != nil {
			return "", &GitError{Operation: "resolve_branch", Repo: r.config.URL, Version: branch,
				Cause: fmt.Errorf("branch '%s' not found: %w", branch, err)}
		}
	}

	return branchRef.Hash().String(), nil
}

func (r *RemoteGitOperations) getVersionsClone(ctx context.Context) ([]string, error) {
	repoDir, err := r.getCachedRepository(ctx)
	if err != nil {
		return nil, &GitError{Operation: "list_versions", Repo: r.config.URL, Cause: err}
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return nil, &GitError{Operation: "list_versions", Repo: r.config.URL, Cause: err}
	}

	// Get all tags
	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, &GitError{Operation: "list_versions", Repo: r.config.URL, Cause: err}
	}

	versions := []string{"latest"}
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		versions = append(versions, tagName)
		return nil
	})

	if err != nil {
		return nil, &GitError{Operation: "list_versions", Repo: r.config.URL, Cause: err}
	}

	return versions, nil
}

func (r *RemoteGitOperations) resolveTagToCommitClone(ctx context.Context, tag string) (string, error) {
	repoDir, err := r.getCachedRepository(ctx)
	if err != nil {
		return "", &GitError{Operation: "resolve_tag", Repo: r.config.URL, Version: tag, Cause: err}
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", &GitError{Operation: "resolve_tag", Repo: r.config.URL, Version: tag, Cause: err}
	}

	// Try to get tag reference
	tagRef, err := repo.Reference(plumbing.ReferenceName("refs/tags/"+tag), true)
	if err != nil {
		// Try with 'v' prefix
		tagRef, err = repo.Reference(plumbing.ReferenceName("refs/tags/v"+tag), true)
		if err != nil {
			return "", &GitError{Operation: "resolve_tag", Repo: r.config.URL, Version: tag,
				Cause: fmt.Errorf("tag '%s' not found: %w", tag, err)}
		}
	}

	return tagRef.Hash().String(), nil
}

func (r *RemoteGitOperations) resolveDefaultBranchClone(ctx context.Context) (string, error) {
	repoDir, err := r.getCachedRepository(ctx)
	if err != nil {
		return "", &GitError{Operation: "resolve_default_branch", Repo: r.config.URL, Cause: err}
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", &GitError{Operation: "resolve_default_branch", Repo: r.config.URL, Cause: err}
	}

	// Get HEAD reference (points to default branch)
	headRef, err := repo.Head()
	if err != nil {
		return "", &GitError{Operation: "resolve_default_branch", Repo: r.config.URL,
			Cause: fmt.Errorf("failed to get HEAD reference: %w", err)}
	}

	return headRef.Hash().String(), nil
}

func (r *RemoteGitOperations) getFilesClone(ctx context.Context, version string, patterns []string) (map[string][]byte, error) {
	repoDir, err := r.getCachedRepository(ctx)
	if err != nil {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
	}

	// Checkout specific version
	if err := r.checkoutVersion(repoDir, version); err != nil {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
	}

	// Find matching files
	matchingFiles, err := FindMatchingFiles(repoDir, patterns)
	if err != nil {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
	}

	if len(matchingFiles) == 0 {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version,
			Cause: fmt.Errorf("no files match patterns: %v", patterns)}
	}

	// Read file contents
	files := make(map[string][]byte)
	for _, relPath := range matchingFiles {
		fullPath := filepath.Join(repoDir, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
		}
		files[relPath] = content
	}

	return files, nil
}

// Helper methods

func (r *RemoteGitOperations) getCachedRepository(ctx context.Context) (string, error) {
	// Create temporary directory for repository operations
	tempDir, err := os.MkdirTemp("", "arm-git-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	repoDir := filepath.Join(tempDir, "repository")

	// Always clone fresh repository (no caching)
	return r.cloneRepository(ctx, repoDir)
}

// getCachedRepositoryAt clones repository to specific directory
func (r *RemoteGitOperations) getCachedRepositoryAt(ctx context.Context, repoDir string) (string, error) {
	// Check if repository already exists and is valid
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
		return repoDir, nil
	}

	// Clone repository to specified directory
	return r.cloneRepository(ctx, repoDir)
}

// getFilesCloneAt retrieves files from repository at specific path
func (r *RemoteGitOperations) getFilesCloneAt(ctx context.Context, repoDir, version string, patterns []string) (map[string][]byte, error) {
	// Ensure repository exists at specified path
	if _, err := r.getCachedRepositoryAt(ctx, repoDir); err != nil {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
	}

	// Checkout specific version
	if err := r.checkoutVersion(repoDir, version); err != nil {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
	}

	// Find matching files
	matchingFiles, err := FindMatchingFiles(repoDir, patterns)
	if err != nil {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
	}

	if len(matchingFiles) == 0 {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version,
			Cause: fmt.Errorf("no files match patterns: %v", patterns)}
	}

	// Read file contents
	files := make(map[string][]byte)
	for _, relPath := range matchingFiles {
		fullPath := filepath.Join(repoDir, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
		}
		files[relPath] = content
	}

	return files, nil
}

// getFilesCloneAt retrieves files from repository at specific path (exported for git registry)
func (r *RemoteGitOperations) GetFilesCloneAt(ctx context.Context, repoDir, version string, patterns []string) (map[string][]byte, error) {
	return r.getFilesCloneAt(ctx, repoDir, version, patterns)
}

// getFilesCloneAtWithRulesetCache retrieves files and caches them in rulesets directory
func (r *RemoteGitOperations) getFilesCloneAtWithRulesetCache(ctx context.Context, repoDir, rulesetsDir, version string, patterns []string) (map[string][]byte, error) {
	// Get files from repository
	files, err := r.getFilesCloneAt(ctx, repoDir, version, patterns)
	if err != nil {
		return nil, err
	}

	// Create rulesets directory if it doesn't exist
	if err := os.MkdirAll(rulesetsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create rulesets directory: %w", err)
	}

	// Cache extracted files in rulesets directory
	for relPath, content := range files {
		rulesetsPath := filepath.Join(rulesetsDir, relPath)
		rulesetsFileDir := filepath.Dir(rulesetsPath)
		if err := os.MkdirAll(rulesetsFileDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create rulesets subdirectory %s: %w", rulesetsFileDir, err)
		}
		if err := os.WriteFile(rulesetsPath, content, 0o644); err != nil {
			return nil, fmt.Errorf("failed to cache ruleset file %s: %w", rulesetsPath, err)
		}
	}

	return files, nil
}

func (r *RemoteGitOperations) cloneRepository(ctx context.Context, repoDir string) (string, error) {
	cloneURL := strings.TrimPrefix(r.config.URL, "file://")

	cloneOptions := &git.CloneOptions{
		URL:      cloneURL,
		Progress: nil,
	}

	// Git authentication is handled by git itself (SSH keys, credential helpers, etc.)
	// No explicit authentication configuration needed

	_, err := git.PlainCloneContext(ctx, repoDir, false, cloneOptions)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return repoDir, nil
}

func (r *RemoteGitOperations) checkoutVersion(repoDir, version string) error {
	if version == "latest" {
		return nil // Already on latest after clone/pull
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	return r.resolveAndCheckout(w, repo, version)
}

func (r *RemoteGitOperations) resolveAndCheckout(w *git.Worktree, repo *git.Repository, version string) error {
	// Check if it's a semver pattern
	if IsSemverPattern(version) {
		return r.checkoutSemverTag(w, repo, version)
	}

	// Try as branch name first
	if err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + version),
	}); err == nil {
		return nil
	}

	// Try as tag name
	if err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/tags/" + version),
	}); err == nil {
		return nil
	}

	// Try as tag name with 'v' prefix
	if err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/tags/v" + version),
	}); err == nil {
		return nil
	}

	// Try as commit hash
	if err := w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(version),
	}); err == nil {
		return nil
	}

	return fmt.Errorf("unable to resolve version: %s", version)
}

func (r *RemoteGitOperations) checkoutSemverTag(w *git.Worktree, repo *git.Repository, versionSpec string) error {
	// Get all tags
	tagRefs, err := repo.Tags()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	// Find matching tags
	var matchingTag string
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		if r.matchesSemverPattern(tagName, versionSpec) {
			matchingTag = tagName
		}
		return nil
	})

	if err != nil {
		return err
	}

	if matchingTag == "" {
		return fmt.Errorf("no tags match version spec: %s", versionSpec)
	}

	// Checkout the matching tag
	return w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/tags/" + matchingTag),
	})
}

func (r *RemoteGitOperations) matchesSemverPattern(tag, pattern string) bool {
	// Remove 'v' prefix if present
	tag = strings.TrimPrefix(tag, "v")
	// Simplified matching - would use proper semver library
	if strings.HasPrefix(pattern, "^") {
		return strings.HasPrefix(tag, pattern[1:])
	}
	return tag == pattern
}

func (r *RemoteGitOperations) resolveSemverPattern(ctx context.Context, versionSpec string) (string, error) {
	versions, err := r.ListVersions(ctx)
	if err != nil {
		return "", &GitError{Operation: "resolve_semver", Repo: r.config.URL, Version: versionSpec, Cause: err}
	}
	resolved, err := ResolveSemverPattern(versionSpec, versions)
	if err != nil {
		return "", &GitError{Operation: "resolve_semver", Repo: r.config.URL, Version: versionSpec, Cause: err}
	}
	return resolved, nil
}
