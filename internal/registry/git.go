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
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// DownloadResult contains the results of a Git registry download
type DownloadResult struct {
	VersionSpec     string   // Original version spec (e.g., "latest")
	ResolvedVersion string   // Actual commit hash
	Files           []string // Downloaded file paths
}

// GitRegistry implements the Registry interface for Git repositories
type GitRegistry struct {
	config *RegistryConfig
	auth   *AuthConfig
	client *http.Client
}

// NewGitRegistry creates a new Git registry instance
func NewGitRegistry(config *RegistryConfig, auth *AuthConfig) (*GitRegistry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &GitRegistry{
		config: config,
		auth:   auth,
		client: &http.Client{Timeout: config.Timeout},
	}, nil
}

// GetRulesets returns available rulesets matching the given patterns
func (g *GitRegistry) GetRulesets(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	// Use API mode if configured, otherwise use clone mode
	if g.auth.APIType == "github" {
		return g.getRulesetsAPI(ctx, patterns)
	}
	return g.getRulesetsClone(ctx, patterns)
}

// GetRuleset returns detailed information about a specific ruleset
func (g *GitRegistry) GetRuleset(ctx context.Context, name, version string) (*RulesetInfo, error) {
	if g.auth.APIType == "github" {
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
	if len(patterns) == 0 {
		// Default pattern if none provided
		patterns = []string{"**/*"}
	}

	if g.auth.APIType == "github" {
		return g.downloadRulesetAPI(ctx, destDir, patterns)
	}
	return g.downloadRulesetClone(ctx, version, destDir, patterns)
}

// DownloadRulesetWithResult downloads a ruleset and returns structured results
func (g *GitRegistry) DownloadRulesetWithResult(ctx context.Context, name, version, destDir string, patterns []string) (*DownloadResult, error) {
	if len(patterns) == 0 {
		patterns = []string{"**/*"}
	}

	// Resolve version to get actual commit hash
	resolvedVersion, err := g.ResolveVersion(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve version: %w", err)
	}

	// Download the files
	var files []string
	if g.auth.APIType == "github" {
		if err := g.downloadRulesetAPI(ctx, destDir, patterns); err != nil {
			return nil, err
		}
	} else {
		if err := g.downloadRulesetClone(ctx, resolvedVersion, destDir, patterns); err != nil {
			return nil, err
		}
	}

	// Find downloaded files
	files, err = g.findDownloadedFiles(destDir)
	if err != nil {
		return nil, err
	}

	return &DownloadResult{
		VersionSpec:     version,
		ResolvedVersion: resolvedVersion,
		Files:           files,
	}, nil
}

// findDownloadedFiles finds all files in the download directory
func (g *GitRegistry) findDownloadedFiles(destDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// GetVersions returns available versions for a ruleset
func (g *GitRegistry) GetVersions(ctx context.Context, name string) ([]string, error) {
	if g.auth.APIType == "github" {
		return g.getVersionsAPI(ctx)
	}
	return g.getVersionsClone(ctx, name)
}

// ResolveVersion resolves a version spec to a concrete commit hash
func (g *GitRegistry) ResolveVersion(ctx context.Context, version string) (string, error) {
	if version == "latest" {
		if g.auth.APIType == "github" {
			return g.resolveLatestAPI(ctx)
		}
		return g.resolveLatestClone(ctx)
	}

	// Check if it's a commit hash (40 hex characters)
	if len(version) == 40 && isHexString(version) {
		return version, nil // Already a commit hash
	}

	// Check if it's a semver pattern
	if g.isSemverPattern(version) {
		return g.resolveSemverPattern(ctx, version)
	}

	// Check if it looks like a version number (e.g., "1.0.0", "v1.0.0")
	if g.isVersionNumber(version) {
		return version, nil // Already a concrete version
	}

	// Assume it's a branch name and resolve to commit hash
	if g.auth.APIType == "github" {
		return g.resolveBranchAPI(ctx, version)
	}
	return g.resolveBranchClone(ctx, version)
}

// GetType returns the registry type
func (g *GitRegistry) GetType() string {
	return "git"
}

// GetName returns the registry name
func (g *GitRegistry) GetName() string {
	return g.config.Name
}

// Close cleans up any resources
func (g *GitRegistry) Close() error {
	return nil
}

// getRulesetsAPI gets rulesets using GitHub API
func (g *GitRegistry) getRulesetsAPI(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	// Extract owner/repo from URL
	owner, repo, err := g.parseGitHubURL()
	if err != nil {
		return nil, err
	}

	// Get repository contents
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if g.auth.Token != "" {
		req.Header.Set("Authorization", "token "+g.auth.Token)
	}
	if g.auth.APIVersion != "" {
		req.Header.Set("X-GitHub-Api-Version", g.auth.APIVersion)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var contents []GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
	}

	// Filter and convert to RulesetInfo
	var rulesets []RulesetInfo
	for _, content := range contents {
		if content.Type == "file" && g.matchesPatterns(content.Name, patterns) {
			ruleset := RulesetInfo{
				Name:      strings.TrimSuffix(content.Name, filepath.Ext(content.Name)),
				Version:   "latest",
				Registry:  g.config.Name,
				Type:      "git",
				Patterns:  patterns,
				UpdatedAt: time.Now(),
			}
			rulesets = append(rulesets, ruleset)
		}
	}

	return rulesets, nil
}

// getRulesetsClone gets rulesets using git clone
func (g *GitRegistry) getRulesetsClone(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	// Get or create cached repository
	repoDir, err := g.getCachedRepository(ctx)
	if err != nil {
		return nil, err
	}

	// Scan for rulesets
	return g.scanDirectory(repoDir, patterns)
}

// getRulesetAPI gets a specific ruleset using GitHub API
func (g *GitRegistry) getRulesetAPI(ctx context.Context, name, version string) (*RulesetInfo, error) {
	rulesets, err := g.getRulesetsAPI(ctx, []string{name + "*"})
	if err != nil {
		return nil, err
	}

	for i := range rulesets {
		if rulesets[i].Name == name {
			rulesets[i].Version = version
			return &rulesets[i], nil
		}
	}

	return nil, fmt.Errorf("ruleset %s not found", name)
}

// getRulesetClone gets a specific ruleset using git clone
func (g *GitRegistry) getRulesetClone(ctx context.Context, name, version string) (*RulesetInfo, error) {
	rulesets, err := g.getRulesetsClone(ctx, []string{name + "*"})
	if err != nil {
		return nil, err
	}

	for i := range rulesets {
		if rulesets[i].Name == name {
			rulesets[i].Version = version
			return &rulesets[i], nil
		}
	}

	return nil, fmt.Errorf("ruleset %s not found", name)
}

// downloadRulesetAPI downloads a ruleset using GitHub API
func (g *GitRegistry) downloadRulesetAPI(ctx context.Context, destDir string, patterns []string) error {
	owner, repo, err := g.parseGitHubURL()
	if err != nil {
		return err
	}

	// Get repository file tree
	fileTree, err := g.getFileTreeAPI(ctx, owner, repo)
	if err != nil {
		return err
	}

	// Apply patterns to find matching files
	matchingFiles := g.applyPatternsToFileTree(fileTree, patterns)
	if len(matchingFiles) == 0 {
		return fmt.Errorf("no files match patterns: %v", patterns)
	}

	// Download each matching file
	for _, file := range matchingFiles {
		if err := g.downloadFileAPI(ctx, owner, repo, file, destDir); err != nil {
			return fmt.Errorf("failed to download %s: %w", file, err)
		}
	}

	return nil
}

// downloadRulesetClone downloads a ruleset using git clone
func (g *GitRegistry) downloadRulesetClone(ctx context.Context, version, destDir string, patterns []string) error {
	// Get or create cached repository
	repoDir, err := g.getCachedRepository(ctx)
	if err != nil {
		return err
	}

	// Resolve and checkout specific version
	if err := g.checkoutVersion(repoDir, version); err != nil {
		return fmt.Errorf("failed to checkout version %s: %w", version, err)
	}

	// Apply patterns to find matching files
	matchingFiles, err := g.findMatchingFiles(repoDir, patterns)
	if err != nil {
		return err
	}

	if len(matchingFiles) == 0 {
		return fmt.Errorf("no files match patterns: %v", patterns)
	}

	// Copy matching files to destination
	return g.copyMatchingFiles(repoDir, matchingFiles, destDir)
}

// getVersionsAPI gets versions using GitHub API
func (g *GitRegistry) getVersionsAPI(ctx context.Context) ([]string, error) {
	owner, repo, err := g.parseGitHubURL()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if g.auth.Token != "" {
		req.Header.Set("Authorization", "token "+g.auth.Token)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var tags []GitHubTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	versions := []string{"latest"}
	for _, tag := range tags {
		versions = append(versions, tag.Name)
	}

	return versions, nil
}

// getVersionsClone gets versions using git clone
func (g *GitRegistry) getVersionsClone(ctx context.Context, _ string) ([]string, error) {
	// Get or create cached repository
	repoDir, err := g.getCachedRepository(ctx)
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get all tags
	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	versions := []string{"latest"}
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		versions = append(versions, tagName)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return versions, nil
}

// parseGitHubURL extracts owner and repo from GitHub URL
func (g *GitRegistry) parseGitHubURL() (owner, repo string, err error) {
	// Handle both https://github.com/owner/repo and https://github.com/owner/repo.git
	re := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+?)(?:\.git)?/?$`)
	matches := re.FindStringSubmatch(g.config.URL)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid GitHub URL format: %s", g.config.URL)
	}
	return matches[1], matches[2], nil
}

// matchesPatterns checks if a filename matches any of the given patterns (legacy function)
func (g *GitRegistry) matchesPatterns(filename string, patterns []string) bool {
	return g.matchesAnyPattern(filename, patterns)
}

// scanDirectory scans a directory for rulesets matching patterns
func (g *GitRegistry) scanDirectory(dir string, patterns []string) ([]RulesetInfo, error) {
	var rulesets []RulesetInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		if g.matchesPatterns(info.Name(), patterns) {
			ruleset := RulesetInfo{
				Name:      strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
				Version:   "latest",
				Registry:  g.config.Name,
				Type:      "git",
				Patterns:  []string{relPath},
				UpdatedAt: info.ModTime(),
			}
			rulesets = append(rulesets, ruleset)
		}

		return nil
	})

	return rulesets, err
}

// checkoutVersion resolves and checks out a specific version
func (g *GitRegistry) checkoutVersion(repoDir, version string) error {
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

	// Try to resolve version as branch, tag, or commit
	if err := g.resolveAndCheckout(w, repo, version); err != nil {
		return err
	}

	return nil
}

// resolveAndCheckout resolves version spec and checks out the appropriate ref
func (g *GitRegistry) resolveAndCheckout(w *git.Worktree, repo *git.Repository, version string) error {
	// Check if it's a semver pattern
	if g.isSemverPattern(version) {
		return g.checkoutSemverTag(w, repo, version)
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

	// Try as commit hash
	if err := w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(version),
	}); err == nil {
		return nil
	}

	return fmt.Errorf("unable to resolve version: %s", version)
}

// isSemverPattern checks if version is a semver pattern
func (g *GitRegistry) isSemverPattern(version string) bool {
	return strings.HasPrefix(version, "^") || strings.HasPrefix(version, "~") ||
		strings.HasPrefix(version, ">=") || strings.HasPrefix(version, "<=") ||
		strings.HasPrefix(version, ">") || strings.HasPrefix(version, "<")
}

// checkoutSemverTag finds and checks out the highest matching semver tag
func (g *GitRegistry) checkoutSemverTag(w *git.Worktree, repo *git.Repository, versionSpec string) error {
	// Get all tags
	tagRefs, err := repo.Tags()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	// Find matching tags (simplified - would use proper semver library)
	var matchingTag string
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		// Simple pattern matching - would use proper semver resolution
		if g.matchesSemverPattern(tagName, versionSpec) {
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

// matchesSemverPattern checks if a tag matches a semver pattern (simplified)
func (g *GitRegistry) matchesSemverPattern(tag, pattern string) bool {
	// Remove 'v' prefix if present
	tag = strings.TrimPrefix(tag, "v")
	// Simplified matching - would use proper semver library
	if strings.HasPrefix(pattern, "^") {
		return strings.HasPrefix(tag, pattern[1:])
	}
	return tag == pattern
}

// findMatchingFiles finds files in repository that match the given patterns
func (g *GitRegistry) findMatchingFiles(repoDir string, patterns []string) ([]string, error) {
	var matchingFiles []string

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory and other hidden directories
		if info.IsDir() && (info.Name() == ".git" || strings.HasPrefix(info.Name(), ".")) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path from repo root
		relPath, err := filepath.Rel(repoDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Validate path to prevent traversal
		if strings.Contains(relPath, "..") {
			return nil // Skip potentially malicious paths
		}

		// Check if file matches any pattern
		if g.matchesAnyPattern(relPath, patterns) {
			matchingFiles = append(matchingFiles, relPath)
		}

		return nil
	})

	return matchingFiles, err
}

// matchesAnyPattern checks if a file path matches any of the given patterns
func (g *GitRegistry) matchesAnyPattern(filePath string, patterns []string) bool {
	if len(patterns) == 0 {
		return true // Empty patterns match everything
	}

	for _, pattern := range patterns {
		// Direct filepath.Match for exact patterns
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return true
		}

		// Check if pattern matches just the filename
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}

		// Handle glob patterns with ** and *
		if g.matchesGlobPattern(filePath, pattern) {
			return true
		}
	}
	return false
}

// matchesGlobPattern handles glob patterns including **
func (g *GitRegistry) matchesGlobPattern(filePath, pattern string) bool {
	// Handle ** patterns
	if strings.Contains(pattern, "**") {
		// Convert glob pattern to regex
		regexPattern := regexp.QuoteMeta(pattern)
		// Replace ** with .* (matches any characters including /)
		regexPattern = strings.ReplaceAll(regexPattern, `\*\*`, ".*")
		// Replace single * with [^/]* (matches any characters except /)
		regexPattern = strings.ReplaceAll(regexPattern, `\*`, "[^/]*")
		regexPattern = "^" + regexPattern + "$"

		if matched, _ := regexp.MatchString(regexPattern, filePath); matched {
			return true
		}
	}

	// Handle simple * patterns
	if strings.Contains(pattern, "*") && !strings.Contains(pattern, "**") {
		// For patterns like *.md, check if it matches the full path or just the filename
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
		// Also check the full path
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return true
		}
	}

	return false
}

// copyMatchingFiles copies the matching files to destination directory preserving structure
func (g *GitRegistry) copyMatchingFiles(repoDir string, matchingFiles []string, destDir string) error {
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return err
	}

	for _, relPath := range matchingFiles {
		// Validate path to prevent traversal
		if strings.Contains(relPath, "..") {
			continue // Skip potentially malicious paths
		}

		srcPath := filepath.Join(repoDir, relPath)
		// Preserve directory structure in destination
		destPath := filepath.Join(destDir, relPath)

		// Create destination directory
		destFileDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destFileDir, 0o700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destFileDir, err)
		}

		// Copy file
		if err := g.copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", relPath, err)
		}
	}

	return nil
}

// copyFile copies a single file
func (g *GitRegistry) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// GitHubContent represents GitHub API content response
type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

// getCachedRepository gets or creates a cached Git repository
func (g *GitRegistry) getCachedRepository(ctx context.Context) (string, error) {
	// Create cache directory path
	cacheDir, err := g.getCacheDir()
	if err != nil {
		return "", err
	}
	repoDir := filepath.Join(cacheDir, "repository")

	// Check if repository already exists
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
		// Repository exists, update it
		return g.updateRepository(ctx, repoDir)
	}

	// Repository doesn't exist, clone it
	return g.cloneRepository(ctx, repoDir)
}

// getCacheDir returns the cache directory for this registry
func (g *GitRegistry) getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	cacheDir := filepath.Join(homeDir, ".arm", "cache", "registries", g.config.Name)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	return cacheDir, nil
}

// cloneRepository clones the repository to the cache directory
func (g *GitRegistry) cloneRepository(ctx context.Context, repoDir string) (string, error) {
	// Convert file:// URLs back to local paths for git operations
	cloneURL := strings.TrimPrefix(g.config.URL, "file://")

	cloneOptions := &git.CloneOptions{
		URL:      cloneURL,
		Progress: nil,
	}

	if g.auth.Token != "" {
		cloneOptions.Auth = &githttp.BasicAuth{
			Username: "token",
			Password: g.auth.Token,
		}
	} else if g.auth.Username != "" && g.auth.Password != "" {
		cloneOptions.Auth = &githttp.BasicAuth{
			Username: g.auth.Username,
			Password: g.auth.Password,
		}
	}

	_, err := git.PlainCloneContext(ctx, repoDir, false, cloneOptions)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return repoDir, nil
}

// updateRepository updates an existing cached repository
func (g *GitRegistry) updateRepository(ctx context.Context, repoDir string) (string, error) {
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	pullOptions := &git.PullOptions{
		RemoteName: "origin",
	}

	if g.auth.Token != "" {
		pullOptions.Auth = &githttp.BasicAuth{
			Username: "token",
			Password: g.auth.Token,
		}
	} else if g.auth.Username != "" && g.auth.Password != "" {
		pullOptions.Auth = &githttp.BasicAuth{
			Username: g.auth.Username,
			Password: g.auth.Password,
		}
	}

	err = w.PullContext(ctx, pullOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", fmt.Errorf("failed to pull repository: %w", err)
	}

	return repoDir, nil
}

// GitHubTag represents GitHub API tag response
type GitHubTag struct {
	Name string `json:"name"`
}

// getFileTreeAPI gets the complete file tree from GitHub API
func (g *GitRegistry) getFileTreeAPI(ctx context.Context, owner, repo string) ([]string, error) {
	// Get repository tree recursively using correct GitHub API
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/HEAD?recursive=1", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use Bearer token format for GitHub API
	if g.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.auth.Token)
	}
	if g.auth.APIVersion != "" {
		req.Header.Set("X-GitHub-Api-Version", g.auth.APIVersion)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
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

	// Extract file paths (not directories)
	var filePaths []string
	for _, item := range treeResp.Tree {
		if item.Type == "blob" { // blob = file, tree = directory
			// Validate path to prevent traversal
			if !strings.Contains(item.Path, "..") {
				filePaths = append(filePaths, item.Path)
			}
		}
	}

	return filePaths, nil
}

// applyPatternsToFileTree filters file paths using glob patterns
func (g *GitRegistry) applyPatternsToFileTree(filePaths, patterns []string) []string {
	var matchingFiles []string
	for _, filePath := range filePaths {
		if g.matchesAnyPattern(filePath, patterns) {
			matchingFiles = append(matchingFiles, filePath)
		}
	}
	return matchingFiles
}

// resolveLatestAPI resolves "latest" to HEAD commit hash using GitHub API
func (g *GitRegistry) resolveLatestAPI(ctx context.Context) (string, error) {
	owner, repo, err := g.parseGitHubURL()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/HEAD", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", err
	}

	if g.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.auth.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, resp.Status)
	}

	var commit struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return "", err
	}

	return commit.SHA, nil
}

// resolveLatestClone resolves "latest" to HEAD commit hash using git clone
func (g *GitRegistry) resolveLatestClone(ctx context.Context) (string, error) {
	repoDir, err := g.getCachedRepository(ctx)
	if err != nil {
		return "", err
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return head.Hash().String(), nil
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// isVersionNumber checks if a string looks like a version number
func (g *GitRegistry) isVersionNumber(version string) bool {
	// Match patterns like "1.0.0", "v1.0.0", "2.1.3", etc.
	matched, _ := regexp.MatchString(`^v?\d+\.\d+\.\d+`, version)
	return matched
}

// resolveBranchAPI resolves a branch name to commit hash using GitHub API
func (g *GitRegistry) resolveBranchAPI(ctx context.Context, branch string) (string, error) {
	owner, repo, err := g.parseGitHubURL()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s", owner, repo, branch)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", err
	}

	if g.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.auth.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, resp.Status)
	}

	var branchInfo struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&branchInfo); err != nil {
		return "", err
	}

	return branchInfo.Commit.SHA, nil
}

// resolveBranchClone resolves a branch name to commit hash using git clone
func (g *GitRegistry) resolveBranchClone(ctx context.Context, branch string) (string, error) {
	repoDir, err := g.getCachedRepository(ctx)
	if err != nil {
		return "", err
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Try to get branch reference
	branchRef, err := repo.Reference(plumbing.ReferenceName("refs/heads/"+branch), true)
	if err != nil {
		// Branch not found, try as remote branch
		branchRef, err = repo.Reference(plumbing.ReferenceName("refs/remotes/origin/"+branch), true)
		if err != nil {
			return "", fmt.Errorf("branch '%s' not found: %w", branch, err)
		}
	}

	return branchRef.Hash().String(), nil
}

// resolveSemverPattern resolves a semver pattern to the highest matching version
func (g *GitRegistry) resolveSemverPattern(ctx context.Context, versionSpec string) (string, error) {
	constraint, err := semver.NewConstraint(versionSpec)
	if err != nil {
		return "", fmt.Errorf("invalid semver constraint: %w", err)
	}

	// Get all available versions
	versions, err := g.GetVersions(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get versions: %w", err)
	}

	// Parse and filter valid semantic versions
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
		return "", fmt.Errorf("no valid semantic versions found")
	}

	// Sort versions (highest first)
	sort.Sort(sort.Reverse(semver.Collection(candidates)))

	// Find the highest version that satisfies the constraint
	for _, candidate := range candidates {
		if constraint.Check(candidate) {
			return candidate.Original(), nil
		}
	}

	return "", fmt.Errorf("no versions satisfy constraint: %s", versionSpec)
}

// downloadFileAPI downloads a single file using GitHub API preserving directory structure
func (g *GitRegistry) downloadFileAPI(ctx context.Context, owner, repo, filePath, destDir string) error {
	// Validate path to prevent traversal
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// Get file content from GitHub API
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, filePath)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Use Bearer token format for GitHub API
	if g.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.auth.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API error %d for %s: %s", resp.StatusCode, filePath, resp.Status)
	}

	var content GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return fmt.Errorf("failed to decode GitHub API response: %w", err)
	}

	// Download the file content
	fileReq, err := http.NewRequestWithContext(ctx, "GET", content.DownloadURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	fileResp, err := g.client.Do(fileReq)
	if err != nil {
		return fmt.Errorf("file download failed: %w", err)
	}
	defer func() { _ = fileResp.Body.Close() }()

	// Preserve directory structure in destination
	destPath := filepath.Join(destDir, filePath)
	destFileDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destFileDir, 0o700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destFileDir, err)
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, fileResp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}
