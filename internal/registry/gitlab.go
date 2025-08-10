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
	"time"
)

// GitLabRegistry implements the Registry interface for GitLab package registries
type GitLabRegistry struct {
	config    *RegistryConfig
	auth      *AuthConfig
	client    *http.Client
	baseURL   string
	projectID string
}

// NewGitLabRegistry creates a new GitLab registry instance
func NewGitLabRegistry(config *RegistryConfig, auth *AuthConfig) (*GitLabRegistry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Parse GitLab URL to extract base URL and project ID
	baseURL, projectID, err := parseGitLabURL(config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid GitLab URL: %w", err)
	}

	return &GitLabRegistry{
		config:    config,
		auth:      auth,
		client:    &http.Client{Timeout: config.Timeout},
		baseURL:   baseURL,
		projectID: projectID,
	}, nil
}

// GetRulesets returns available rulesets matching the given patterns
func (g *GitLabRegistry) GetRulesets(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	// Get all packages from GitLab Generic Packages API
	url := fmt.Sprintf("%s/api/v4/projects/%s/packages", g.baseURL, g.projectID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if g.auth.Token != "" {
		req.Header.Set("PRIVATE-TOKEN", g.auth.Token)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var packages []GitLabPackage
	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, err
	}

	// Convert packages to rulesets
	rulesetMap := make(map[string]*RulesetInfo)
	for _, pkg := range packages {
		if pkg.PackageType == "generic" {
			ruleset, exists := rulesetMap[pkg.Name]
			if !exists {
				ruleset = &RulesetInfo{
					Name:      pkg.Name,
					Version:   pkg.Version,
					Registry:  g.config.Name,
					Type:      "gitlab",
					UpdatedAt: pkg.UpdatedAt,
					Metadata: map[string]string{
						"project_id": g.projectID,
						"base_url":   g.baseURL,
					},
				}
				rulesetMap[pkg.Name] = ruleset
			} else if pkg.UpdatedAt.After(ruleset.UpdatedAt) {
				// Update to latest version if this package is newer
				ruleset.Version = pkg.Version
				ruleset.UpdatedAt = pkg.UpdatedAt
			}
		}
	}

	// Convert map to slice
	var rulesets []RulesetInfo
	for _, ruleset := range rulesetMap {
		rulesets = append(rulesets, *ruleset)
	}

	return rulesets, nil
}

// GetRuleset returns detailed information about a specific ruleset
func (g *GitLabRegistry) GetRuleset(ctx context.Context, name, version string) (*RulesetInfo, error) {
	rulesets, err := g.GetRulesets(ctx, nil)
	if err != nil {
		return nil, err
	}

	for i := range rulesets {
		if rulesets[i].Name == name {
			if version != "latest" {
				rulesets[i].Version = version
			}
			return &rulesets[i], nil
		}
	}

	return nil, fmt.Errorf("ruleset %s not found", name)
}

// DownloadRuleset downloads a ruleset to the specified directory
func (g *GitLabRegistry) DownloadRuleset(ctx context.Context, name, version, destDir string) error {
	return g.DownloadRulesetWithPatterns(ctx, name, version, destDir, nil)
}

// DownloadRulesetWithPatterns downloads a ruleset (patterns ignored for GitLab)
func (g *GitLabRegistry) DownloadRulesetWithPatterns(ctx context.Context, name, version, destDir string, patterns []string) error {
	// GitLab registries use pre-packaged tar.gz files, so patterns are ignored
	_ = patterns

	// Construct GitLab Generic Packages API URL
	url := fmt.Sprintf("%s/api/v4/projects/%s/packages/generic/%s/%s/ruleset.tar.gz",
		g.baseURL, g.projectID, name, version)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return err
	}

	if g.auth.Token != "" {
		req.Header.Set("PRIVATE-TOKEN", g.auth.Token)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	// Create destination file
	destPath := filepath.Join(destDir, "ruleset.tar.gz")
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	// Copy content
	_, err = io.Copy(destFile, resp.Body)
	return err
}

// GetVersions returns available versions for a ruleset
func (g *GitLabRegistry) GetVersions(ctx context.Context, name string) ([]string, error) {
	// Get packages filtered by name
	url := fmt.Sprintf("%s/api/v4/projects/%s/packages?package_name=%s", g.baseURL, g.projectID, name)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if g.auth.Token != "" {
		req.Header.Set("PRIVATE-TOKEN", g.auth.Token)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %s", resp.Status)
	}

	var packages []GitLabPackage
	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, err
	}

	// Extract versions
	var versions []string
	for _, pkg := range packages {
		if pkg.PackageType == "generic" && pkg.Name == name {
			versions = append(versions, pkg.Version)
		}
	}

	if len(versions) == 0 {
		return []string{"latest"}, nil
	}

	return versions, nil
}

// GetType returns the registry type
func (g *GitLabRegistry) GetType() string {
	return "gitlab"
}

// GetName returns the registry name
func (g *GitLabRegistry) GetName() string {
	return g.config.Name
}

// Close cleans up any resources
func (g *GitLabRegistry) Close() error {
	return nil
}

// parseGitLabURL parses GitLab URL to extract base URL and project ID
func parseGitLabURL(url string) (baseURL, projectID string, err error) {
	// Handle formats like:
	// https://gitlab.example.com/projects/123
	// https://gitlab.example.com/groups/456
	re := regexp.MustCompile(`^(https://[^/]+)/(?:projects|groups)/(\d+)/?$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid GitLab URL format: %s", url)
	}
	return matches[1], matches[2], nil
}

// GitLabPackage represents a GitLab package from the API
type GitLabPackage struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	PackageType string    `json:"package_type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
