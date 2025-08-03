package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jomadu/arm/pkg/types"
)

// GitLabRegistry implements Registry interface for GitLab package registries
type GitLabRegistry struct {
	*GenericHTTPRegistry
	projectID string
	groupID   string
}

// NewGitLab creates a new GitLab registry client
func NewGitLab(baseURL, authToken, projectID, groupID string) *GitLabRegistry {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var auth AuthProvider
	if authToken != "" {
		auth = &BearerAuth{Token: authToken}
	}

	generic := &GenericHTTPRegistry{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  client,
		auth:    auth,
	}

	return &GitLabRegistry{
		GenericHTTPRegistry: generic,
		projectID:           projectID,
		groupID:             groupID,
	}
}

func (r *GitLabRegistry) buildDownloadURL(name, version string) string {
	_, pkg := types.ParseRulesetName(name)

	if r.groupID != "" {
		// Group-level package registry
		return fmt.Sprintf("%s/api/v4/groups/%s/packages/generic/%s/%s/%s-%s.tar.gz",
			r.baseURL, r.groupID, pkg, version, pkg, version)
	}

	// Project-level package registry
	return fmt.Sprintf("%s/api/v4/projects/%s/packages/generic/%s/%s/%s-%s.tar.gz",
		r.baseURL, r.projectID, pkg, version, pkg, version)
}

// GetMetadata retrieves metadata from GitLab's package API
func (r *GitLabRegistry) GetMetadata(name string) (*Metadata, error) {
	_, pkg := types.ParseRulesetName(name)
	
	var url string
	if r.groupID != "" {
		url = fmt.Sprintf("%s/api/v4/groups/%s/packages?package_name=%s", r.baseURL, r.groupID, pkg)
	} else {
		url = fmt.Sprintf("%s/api/v4/projects/%s/packages?package_name=%s", r.baseURL, r.projectID, pkg)
	}
	
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if r.auth != nil {
		r.auth.SetAuth(req)
	}
	
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitLab packages: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API returned status %d", resp.StatusCode)
	}
	
	// Parse GitLab packages API response
	var packages []GitLabPackage
	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, fmt.Errorf("failed to decode GitLab response: %w", err)
	}
	
	if len(packages) == 0 {
		return nil, fmt.Errorf("package %s not found", name)
	}
	
	// Convert GitLab package data to ARM metadata
	return r.convertGitLabToMetadata(packages[0], name), nil
}

// HealthCheck verifies GitLab registry connectivity and authentication
func (r *GitLabRegistry) HealthCheck() error {
	var url string
	if r.groupID != "" {
		url = fmt.Sprintf("%s/api/v4/groups/%s", r.baseURL, r.groupID)
	} else {
		url = fmt.Sprintf("%s/api/v4/projects/%s", r.baseURL, r.projectID)
	}
	
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}
	
	if r.auth != nil {
		r.auth.SetAuth(req)
	}
	
	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("GitLab registry unreachable: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	
	switch resp.StatusCode {
	case 200:
		return nil
	case 401:
		return fmt.Errorf("GitLab authentication failed - check token")
	case 403:
		return fmt.Errorf("GitLab access forbidden - insufficient permissions")
	case 404:
		if r.groupID != "" {
			return fmt.Errorf("GitLab group %s not found", r.groupID)
		}
		return fmt.Errorf("GitLab project %s not found", r.projectID)
	default:
		return fmt.Errorf("GitLab registry returned status %d", resp.StatusCode)
	}
}

// GitLabPackage represents GitLab's package API response
type GitLabPackage struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	PackageType string                 `json:"package_type"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	Links       map[string]interface{} `json:"_links"`
	PackageFiles []GitLabPackageFile   `json:"package_files"`
}

type GitLabPackageFile struct {
	ID       int    `json:"id"`
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}

// convertGitLabToMetadata converts GitLab package data to ARM metadata
func (r *GitLabRegistry) convertGitLabToMetadata(pkg GitLabPackage, name string) *Metadata {
	versions := make([]Version, len(pkg.PackageFiles))
	for i, file := range pkg.PackageFiles {
		versions[i] = Version{
			Version:   pkg.Version,
			Published: file.CreatedAt,
			Size:      file.Size,
		}
	}
	
	return &Metadata{
		Name:         name,
		Description:  fmt.Sprintf("GitLab package: %s", pkg.Name),
		Versions:     versions,
		LastModified: pkg.UpdatedAt,
		Extra: map[string]string{
			"gitlab_id":   fmt.Sprintf("%d", pkg.ID),
			"package_type": pkg.PackageType,
		},
	}
}
