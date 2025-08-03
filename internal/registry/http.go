package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jomadu/arm/pkg/types"
)

// GenericHTTPRegistry implements Registry interface for generic HTTP file servers
type GenericHTTPRegistry struct {
	baseURL string
	client  *http.Client
	auth    AuthProvider
}

// NewGenericHTTP creates a new generic HTTP registry client
func NewGenericHTTP(baseURL, authToken string) *GenericHTTPRegistry {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var auth AuthProvider
	if authToken != "" {
		auth = &BearerAuth{Token: authToken}
	}

	return &GenericHTTPRegistry{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  client,
		auth:    auth,
	}
}

// GetRuleset retrieves a specific ruleset version
func (r *GenericHTTPRegistry) GetRuleset(name, version string) (*types.Ruleset, error) {
	return &types.Ruleset{
		Name:     name,
		Version:  version,
		Source:   r.baseURL,
		Files:    []string{}, // TODO: Fetch actual files from registry
		Checksum: "",         // TODO: Fetch actual checksum from registry
	}, nil
}

// ListVersions returns all available versions for a ruleset
// Generic HTTP registries don't support version discovery without directory listing
func (r *GenericHTTPRegistry) ListVersions(name string) ([]string, error) {
	return nil, fmt.Errorf("version listing not supported by generic HTTP registry - specify exact version")
}

// Download downloads a ruleset archive
func (r *GenericHTTPRegistry) Download(name, version string) (io.ReadCloser, error) {
	url := r.buildDownloadURL(name, version)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if r.auth != nil {
		r.auth.SetAuth(req)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download from %s: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d from %s", resp.StatusCode, url)
	}

	return resp.Body, nil
}

// GetMetadata retrieves metadata for a ruleset
// Generic HTTP registries provide minimal metadata only
func (r *GenericHTTPRegistry) GetMetadata(name string) (*Metadata, error) {
	return &Metadata{
		Name:        name,
		Description: "Generic HTTP registry - no version discovery",
		Versions:    []Version{},
		Repository:  r.baseURL,
	}, nil
}

func (r *GenericHTTPRegistry) buildDownloadURL(name, version string) string {
	org, pkg := types.ParseRulesetName(name)
	if org == "" {
		return fmt.Sprintf("%s/%s/%s.tar.gz", r.baseURL, pkg, version)
	}
	return fmt.Sprintf("%s/%s/%s/%s.tar.gz", r.baseURL, org, pkg, version)
}

// HealthCheck verifies registry connectivity
func (r *GenericHTTPRegistry) HealthCheck() error {
	req, err := http.NewRequestWithContext(context.Background(), "HEAD", r.baseURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	if r.auth != nil {
		r.auth.SetAuth(req)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("registry unreachable: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	return nil
}
