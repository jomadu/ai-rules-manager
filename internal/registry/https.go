package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HTTPSRegistry implements the Registry interface for HTTPS registries with manifest.json discovery
type HTTPSRegistry struct {
	config           *RegistryConfig
	auth             *AuthConfig
	client           *http.Client
	baseURL          string
	manifest         *HTTPSManifest
	manifestCachedAt time.Time
}

// HTTPSManifest represents the manifest.json structure
type HTTPSManifest struct {
	Rulesets map[string][]string `json:"rulesets"`
}

// NewHTTPSRegistry creates a new HTTPS registry instance
func NewHTTPSRegistry(config *RegistryConfig, auth *AuthConfig) (*HTTPSRegistry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	baseURL := strings.TrimSuffix(config.URL, "/")

	return &HTTPSRegistry{
		config:  config,
		auth:    auth,
		client:  &http.Client{Timeout: config.Timeout},
		baseURL: baseURL,
	}, nil
}

// GetRulesets returns available rulesets matching the given patterns
func (h *HTTPSRegistry) GetRulesets(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	manifest, err := h.getManifest(ctx)
	if err != nil {
		return nil, err
	}

	var rulesets []RulesetInfo
	for name, versions := range manifest.Rulesets {
		if len(versions) == 0 {
			continue
		}

		// Use the latest version (last in array)
		latestVersion := versions[len(versions)-1]

		ruleset := RulesetInfo{
			Name:     name,
			Version:  latestVersion,
			Registry: h.config.Name,
			Type:     "https",
			Metadata: map[string]string{
				"base_url": h.baseURL,
			},
		}
		rulesets = append(rulesets, ruleset)
	}

	return rulesets, nil
}

// GetRuleset returns detailed information about a specific ruleset
func (h *HTTPSRegistry) GetRuleset(ctx context.Context, name, version string) (*RulesetInfo, error) {
	manifest, err := h.getManifest(ctx)
	if err != nil {
		return nil, err
	}

	versions, exists := manifest.Rulesets[name]
	if !exists {
		return nil, fmt.Errorf("ruleset %s not found", name)
	}

	// If version is "latest", use the last version in the array
	if version == "latest" {
		if len(versions) == 0 {
			return nil, fmt.Errorf("no versions available for ruleset %s", name)
		}
		version = versions[len(versions)-1]
	} else {
		// Validate that the requested version exists
		found := false
		for _, v := range versions {
			if v == version {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("version %s not found for ruleset %s", version, name)
		}
	}

	return &RulesetInfo{
		Name:     name,
		Version:  version,
		Registry: h.config.Name,
		Type:     "https",
		Metadata: map[string]string{
			"base_url": h.baseURL,
		},
	}, nil
}

// DownloadRuleset downloads a ruleset to the specified directory
func (h *HTTPSRegistry) DownloadRuleset(ctx context.Context, name, version, destDir string) error {
	// Construct download URL: baseURL/ruleset/version/ruleset.tar.gz
	url := fmt.Sprintf("%s/%s/%s/ruleset.tar.gz", h.baseURL, name, version)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return err
	}

	// Add authentication if configured
	if h.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+h.auth.Token)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTPS registry error: %s", resp.Status)
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Create destination file
	destPath := filepath.Join(destDir, "ruleset.tar.gz")
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy content
	_, err = io.Copy(destFile, resp.Body)
	return err
}

// GetVersions returns available versions for a ruleset
func (h *HTTPSRegistry) GetVersions(ctx context.Context, name string) ([]string, error) {
	manifest, err := h.getManifest(ctx)
	if err != nil {
		return nil, err
	}

	versions, exists := manifest.Rulesets[name]
	if !exists {
		return nil, fmt.Errorf("ruleset %s not found", name)
	}

	if len(versions) == 0 {
		return []string{"latest"}, nil
	}

	return versions, nil
}

// GetType returns the registry type
func (h *HTTPSRegistry) GetType() string {
	return "https"
}

// GetName returns the registry name
func (h *HTTPSRegistry) GetName() string {
	return h.config.Name
}

// Close cleans up any resources
func (h *HTTPSRegistry) Close() error {
	return nil
}

// getManifest fetches and caches the manifest.json file
func (h *HTTPSRegistry) getManifest(ctx context.Context) (*HTTPSManifest, error) {
	// Check if we have a cached manifest that's still valid (1 hour TTL)
	if h.manifest != nil && time.Since(h.manifestCachedAt) < time.Hour {
		return h.manifest, nil
	}

	// Fetch fresh manifest
	url := fmt.Sprintf("%s/manifest.json", h.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	// Add authentication if configured
	if h.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+h.auth.Token)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest fetch error: %s", resp.Status)
	}

	var manifest HTTPSManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest.json format: %w", err)
	}

	// Validate manifest has required fields
	if manifest.Rulesets == nil {
		return nil, fmt.Errorf("manifest.json missing required 'rulesets' field")
	}

	// Cache the manifest
	h.manifest = &manifest
	h.manifestCachedAt = time.Now()

	return &manifest, nil
}
