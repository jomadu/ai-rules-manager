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
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

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

// DownloadRuleset downloads a ruleset to the specified directory
func (g *GitRegistry) DownloadRuleset(ctx context.Context, name, version, destDir string) error {
	if g.auth.APIType == "github" {
		return g.downloadRulesetAPI(ctx, name, destDir)
	}
	return g.downloadRulesetClone(ctx, name, version, destDir)
}

// GetVersions returns available versions for a ruleset
func (g *GitRegistry) GetVersions(ctx context.Context, name string) ([]string, error) {
	if g.auth.APIType == "github" {
		return g.getVersionsAPI(ctx)
	}
	return g.getVersionsClone(ctx, name)
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
func (g *GitRegistry) downloadRulesetAPI(ctx context.Context, name, destDir string) error {
	owner, repo, err := g.parseGitHubURL()
	if err != nil {
		return err
	}

	// Find the file
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, name)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return err
	}

	if g.auth.Token != "" {
		req.Header.Set("Authorization", "token "+g.auth.Token)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var content GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return err
	}

	// Download the file content
	fileReq, err := http.NewRequestWithContext(ctx, "GET", content.DownloadURL, http.NoBody)
	if err != nil {
		return err
	}

	fileResp, err := g.client.Do(fileReq)
	if err != nil {
		return err
	}
	defer func() { _ = fileResp.Body.Close() }()

	// Create destination file
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, content.Name)
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, fileResp.Body)
	return err
}

// downloadRulesetClone downloads a ruleset using git clone
func (g *GitRegistry) downloadRulesetClone(ctx context.Context, name, version, destDir string) error {
	// Get or create cached repository
	repoDir, err := g.getCachedRepository(ctx)
	if err != nil {
		return err
	}

	// Checkout specific version if not "latest"
	if version != "latest" {
		repo, err := git.PlainOpen(repoDir)
		if err != nil {
			return fmt.Errorf("failed to open repository: %w", err)
		}
		w, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree: %w", err)
		}
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(version),
		})
		if err != nil {
			return fmt.Errorf("failed to checkout version %s: %w", version, err)
		}
	}

	// Copy files matching the ruleset name
	return g.copyRulesetFiles(repoDir, name, destDir)
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
func (g *GitRegistry) getVersionsClone(ctx context.Context, name string) ([]string, error) {
	// For clone mode, we'll just return "latest" for now
	// Full implementation would require cloning and listing tags
	return []string{"latest"}, nil
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

// matchesPatterns checks if a filename matches any of the given patterns
func (g *GitRegistry) matchesPatterns(filename string, patterns []string) bool {
	if len(patterns) == 0 {
		return true // No patterns means match all
	}

	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, filename)
		if err == nil && matched {
			return true
		}
		// Also try glob-style matching
		if strings.Contains(pattern, "*") {
			re := regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, ".*"))
			if re.MatchString(filename) {
				return true
			}
		}
	}
	return false
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

// copyRulesetFiles copies files matching the ruleset name
func (g *GitRegistry) copyRulesetFiles(srcDir, name, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches the ruleset name
		filename := info.Name()
		if strings.Contains(filename, name) || g.matchesPatterns(filename, []string{name + "*"}) {
			relPath, _ := filepath.Rel(srcDir, path)
			destPath := filepath.Join(destDir, relPath)

			// Create destination directory
			if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
				return err
			}

			// Copy file
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() { _ = srcFile.Close() }()

			destFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer func() { _ = destFile.Close() }()

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				return err
			}
		}

		return nil
	})
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
	cloneOptions := &git.CloneOptions{
		URL:      g.config.URL,
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
