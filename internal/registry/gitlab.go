package registry

import (
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

func (r *GitLabRegistry) buildMetadataURL(name string) string {
	_, pkg := types.ParseRulesetName(name)

	if r.groupID != "" {
		// Group-level package registry
		return fmt.Sprintf("%s/api/v4/groups/%s/packages/generic/%s/metadata.json",
			r.baseURL, r.groupID, pkg)
	}

	// Project-level package registry
	return fmt.Sprintf("%s/api/v4/projects/%s/packages/generic/%s/metadata.json",
		r.baseURL, r.projectID, pkg)
}
