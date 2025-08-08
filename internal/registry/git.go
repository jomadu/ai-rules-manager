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
		return g.downloadRulesetAPI(ctx, name, version, destDir)
	}
	return g.downloadRulesetClone(ctx, name, version, destDir)
}

// GetVersions returns available versions for a ruleset
func (g *GitRegistry) GetVersions(ctx context.Context, name string) ([]string, error) {
	if g.auth.APIType == "github" {
		return g.getVersionsAPI(ctx, name)
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
	defer resp.Body.Close()

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
	// Create temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", "arm-git-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	// Clone repository
	cloneOptions := &git.CloneOptions{
		URL:      g.config.URL,
		Progress: nil, // Silent clone
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

	_, err = git.PlainCloneContext(ctx, tmpDir, false, cloneOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Scan for rulesets
	return g.scanDirectory(tmpDir, patterns)
}

// getRulesetAPI gets a specific ruleset using GitHub API
func (g *GitRegistry) getRulesetAPI(ctx context.Context, name, version string) (*RulesetInfo, error) {
	rulesets, err := g.getRulesetsAPI(ctx, []string{name + "*"})
	if err != nil {
		return nil, err
	}

	for _, ruleset := range rulesets {
		if ruleset.Name == name {
			ruleset.Version = version
			return &ruleset, nil
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

	for _, ruleset := range rulesets {
		if ruleset.Name == name {
			ruleset.Version = version
			return &ruleset, nil
		}
	}

	return nil, fmt.Errorf("ruleset %s not found", name)
}

// downloadRulesetAPI downloads a ruleset using GitHub API
func (g *GitRegistry) downloadRulesetAPI(ctx context.Context, name, version, destDir string) error {
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
	defer resp.Body.Close()

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
	defer fileResp.Body.Close()

	// Create destination file
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, content.Name)
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, fileResp.Body)
	return err
}

// downloadRulesetClone downloads a ruleset using git clone
func (g *GitRegistry) downloadRulesetClone(ctx context.Context, name, version, destDir string) error {
	// Create temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", "arm-git-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Clone repository
	cloneOptions := &git.CloneOptions{
		URL:      g.config.URL,
		Progress: nil,
	}

	if g.auth.Token != "" {
		cloneOptions.Auth = &githttp.BasicAuth{
			Username: "token",
			Password: g.auth.Token,
		}
	}

	// Checkout specific version if not "latest"
	if version != "latest" {
		cloneOptions.ReferenceName = plumbing.ReferenceName("refs/tags/" + version)
	}

	_, err = git.PlainCloneContext(ctx, tmpDir, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Copy files matching the ruleset name
	return g.copyRulesetFiles(tmpDir, name, destDir)
}

// getVersionsAPI gets versions using GitHub API
func (g *GitRegistry) getVersionsAPI(ctx context.Context, name string) ([]string, error) {
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
	defer resp.Body.Close()

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
	if err := os.MkdirAll(destDir, 0755); err != nil {
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
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			// Copy file
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			destFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

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

// GitHubTag represents GitHub API tag response
type GitHubTag struct {
	Name string `json:"name"`
}
