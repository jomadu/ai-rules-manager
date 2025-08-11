package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// RemoteGitOperations implements GitOperations for remote Git repositories
type RemoteGitOperations struct {
	config *RegistryConfig
	auth   *AuthConfig
	client *http.Client
}

// NewRemoteGitOperations creates a new remote Git operations instance
func NewRemoteGitOperations(config *RegistryConfig, auth *AuthConfig) *RemoteGitOperations {
	return &RemoteGitOperations{
		config: config,
		auth:   auth,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// ResolveVersion resolves a version spec to a concrete commit hash
func (r *RemoteGitOperations) ResolveVersion(ctx context.Context, constraint string) (string, error) {
	if constraint == "latest" {
		if r.auth.APIType == "github" {
			return r.resolveLatestAPI(ctx)
		}
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
		if r.auth.APIType == "github" {
			return r.resolveTagToCommitAPI(ctx, constraint)
		}
		return r.resolveTagToCommitClone(ctx, constraint)
	}

	// Assume it's a branch name and resolve to commit hash
	if r.auth.APIType == "github" {
		return r.resolveBranchAPI(ctx, constraint)
	}
	return r.resolveBranchClone(ctx, constraint)
}

// ListVersions returns available versions for the repository
func (r *RemoteGitOperations) ListVersions(ctx context.Context) ([]string, error) {
	if r.auth.APIType == "github" {
		return r.getVersionsAPI(ctx)
	}
	return r.getVersionsClone(ctx)
}

// GetFiles retrieves files matching patterns for a specific version
func (r *RemoteGitOperations) GetFiles(ctx context.Context, version string, patterns []string) (map[string][]byte, error) {
	if r.auth.APIType == "github" {
		return r.getFilesAPI(ctx, version, patterns)
	}
	return r.getFilesClone(ctx, version, patterns)
}

// API-based implementations

func (r *RemoteGitOperations) resolveLatestAPI(ctx context.Context) (string, error) {
	// For "latest", resolve to HEAD of default branch, not latest tag
	return r.resolveDefaultBranchAPI(ctx)
}

func (r *RemoteGitOperations) resolveBranchAPI(ctx context.Context, branch string) (string, error) {
	owner, repo, err := r.parseGitHubURL()
	if err != nil {
		return "", &GitError{Operation: "resolve_branch", Repo: r.config.URL, Version: branch, Cause: err}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s", owner, repo, branch)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", &GitError{Operation: "resolve_branch", Repo: r.config.URL, Version: branch, Cause: err}
	}

	if r.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.auth.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", &GitError{Operation: "resolve_branch", Repo: r.config.URL, Version: branch, Cause: err}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", &GitError{Operation: "resolve_branch", Repo: r.config.URL, Version: branch,
			Cause: fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, resp.Status)}
	}

	var branchInfo struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&branchInfo); err != nil {
		return "", &GitError{Operation: "resolve_branch", Repo: r.config.URL, Version: branch, Cause: err}
	}

	return branchInfo.Commit.SHA, nil
}

func (r *RemoteGitOperations) resolveTagToCommitAPI(ctx context.Context, tag string) (string, error) {
	owner, repo, err := r.parseGitHubURL()
	if err != nil {
		return "", &GitError{Operation: "resolve_tag", Repo: r.config.URL, Version: tag, Cause: err}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/tags/%s", owner, repo, tag)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", &GitError{Operation: "resolve_tag", Repo: r.config.URL, Version: tag, Cause: err}
	}

	if r.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.auth.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", &GitError{Operation: "resolve_tag", Repo: r.config.URL, Version: tag, Cause: err}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", &GitError{Operation: "resolve_tag", Repo: r.config.URL, Version: tag,
			Cause: fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, resp.Status)}
	}

	var tagInfo struct {
		Object struct {
			SHA  string `json:"sha"`
			Type string `json:"type"`
		} `json:"object"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tagInfo); err != nil {
		return "", &GitError{Operation: "resolve_tag", Repo: r.config.URL, Version: tag, Cause: err}
	}

	// If the tag points to a commit directly, return the SHA
	if tagInfo.Object.Type == "commit" {
		return tagInfo.Object.SHA, nil
	}

	// If the tag points to a tag object (annotated tag), we need to resolve it further
	if tagInfo.Object.Type == "tag" {
		return r.resolveAnnotatedTagAPI(ctx, tagInfo.Object.SHA)
	}

	return "", &GitError{Operation: "resolve_tag", Repo: r.config.URL, Version: tag,
		Cause: fmt.Errorf("unexpected tag object type: %s", tagInfo.Object.Type)}
}

func (r *RemoteGitOperations) resolveAnnotatedTagAPI(ctx context.Context, tagSHA string) (string, error) {
	owner, repo, err := r.parseGitHubURL()
	if err != nil {
		return "", &GitError{Operation: "resolve_annotated_tag", Repo: r.config.URL, Cause: err}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/tags/%s", owner, repo, tagSHA)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", &GitError{Operation: "resolve_annotated_tag", Repo: r.config.URL, Cause: err}
	}

	if r.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.auth.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", &GitError{Operation: "resolve_annotated_tag", Repo: r.config.URL, Cause: err}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", &GitError{Operation: "resolve_annotated_tag", Repo: r.config.URL,
			Cause: fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, resp.Status)}
	}

	var annotatedTag struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&annotatedTag); err != nil {
		return "", &GitError{Operation: "resolve_annotated_tag", Repo: r.config.URL, Cause: err}
	}

	return annotatedTag.Object.SHA, nil
}

func (r *RemoteGitOperations) resolveDefaultBranchAPI(ctx context.Context) (string, error) {
	owner, repo, err := r.parseGitHubURL()
	if err != nil {
		return "", &GitError{Operation: "resolve_default_branch", Repo: r.config.URL, Cause: err}
	}

	// Get repository info to find default branch
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", &GitError{Operation: "resolve_default_branch", Repo: r.config.URL, Cause: err}
	}

	if r.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.auth.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", &GitError{Operation: "resolve_default_branch", Repo: r.config.URL, Cause: err}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", &GitError{Operation: "resolve_default_branch", Repo: r.config.URL,
			Cause: fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, resp.Status)}
	}

	var repoInfo struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return "", &GitError{Operation: "resolve_default_branch", Repo: r.config.URL, Cause: err}
	}

	// Now get the commit hash for the default branch
	return r.resolveBranchAPI(ctx, repoInfo.DefaultBranch)
}

func (r *RemoteGitOperations) getVersionsAPI(ctx context.Context) ([]string, error) {
	owner, repo, err := r.parseGitHubURL()
	if err != nil {
		return nil, &GitError{Operation: "list_versions", Repo: r.config.URL, Cause: err}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, &GitError{Operation: "list_versions", Repo: r.config.URL, Cause: err}
	}

	if r.auth.Token != "" {
		req.Header.Set("Authorization", "token "+r.auth.Token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, &GitError{Operation: "list_versions", Repo: r.config.URL, Cause: err}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, &GitError{Operation: "list_versions", Repo: r.config.URL,
			Cause: fmt.Errorf("GitHub API error: %s", resp.Status)}
	}

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, &GitError{Operation: "list_versions", Repo: r.config.URL, Cause: err}
	}

	versions := []string{"latest"}
	for _, tag := range tags {
		versions = append(versions, tag.Name)
	}

	return versions, nil
}

func (r *RemoteGitOperations) getFilesAPI(ctx context.Context, version string, patterns []string) (map[string][]byte, error) {
	owner, repo, err := r.parseGitHubURL()
	if err != nil {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
	}

	// Get file tree
	fileTree, err := r.getFileTreeAPI(ctx, owner, repo)
	if err != nil {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
	}

	// Apply patterns
	matchingFiles := r.applyPatternsToFileTree(fileTree, patterns)
	if len(matchingFiles) == 0 {
		return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version,
			Cause: fmt.Errorf("no files match patterns: %v", patterns)}
	}

	// Download files
	files := make(map[string][]byte)
	for _, filePath := range matchingFiles {
		content, err := r.downloadFileContentAPI(ctx, owner, repo, filePath)
		if err != nil {
			return nil, &GitError{Operation: "get_files", Repo: r.config.URL, Version: version, Cause: err}
		}
		files[filePath] = content
	}

	return files, nil
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

func (r *RemoteGitOperations) parseGitHubURL() (owner, repo string, err error) {
	re := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+?)(?:\.git)?/?$`)
	matches := re.FindStringSubmatch(r.config.URL)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid GitHub URL format: %s", r.config.URL)
	}
	return matches[1], matches[2], nil
}

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

	if r.auth.Token != "" {
		cloneOptions.Auth = &githttp.BasicAuth{
			Username: "token",
			Password: r.auth.Token,
		}
	} else if r.auth.Username != "" && r.auth.Password != "" {
		cloneOptions.Auth = &githttp.BasicAuth{
			Username: r.auth.Username,
			Password: r.auth.Password,
		}
	}

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

func (r *RemoteGitOperations) getFileTreeAPI(ctx context.Context, owner, repo string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/HEAD?recursive=1", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if r.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.auth.Token)
	}
	if r.auth.APIVersion != "" {
		req.Header.Set("X-GitHub-Api-Version", r.auth.APIVersion)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, resp.Status)
	}

	var treeResp struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&treeResp); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub API response: %w", err)
	}

	var filePaths []string
	for _, item := range treeResp.Tree {
		if item.Type == "blob" && ValidatePath(item.Path) {
			filePaths = append(filePaths, item.Path)
		}
	}

	return filePaths, nil
}

func (r *RemoteGitOperations) applyPatternsToFileTree(filePaths, patterns []string) []string {
	var matchingFiles []string
	for _, filePath := range filePaths {
		if MatchesAnyPattern(filePath, patterns) {
			matchingFiles = append(matchingFiles, filePath)
		}
	}
	return matchingFiles
}

func (r *RemoteGitOperations) downloadFileContentAPI(ctx context.Context, owner, repo, filePath string) ([]byte, error) {
	if !ValidatePath(filePath) {
		return nil, fmt.Errorf("invalid file path: %s", filePath)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, filePath)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if r.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.auth.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error %d for %s: %s", resp.StatusCode, filePath, resp.Status)
	}

	var content struct {
		DownloadURL string `json:"download_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub API response: %w", err)
	}

	// Download the file content
	fileReq, err := http.NewRequestWithContext(ctx, "GET", content.DownloadURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	fileResp, err := r.client.Do(fileReq)
	if err != nil {
		return nil, fmt.Errorf("file download failed: %w", err)
	}
	defer func() { _ = fileResp.Body.Close() }()

	return io.ReadAll(fileResp.Body)
}
