package registry

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jomadu/arm/pkg/types"
)

// S3Registry implements Registry interface for AWS S3 buckets
type S3Registry struct {
	*GenericHTTPRegistry
	bucket string
	region string
	prefix string
}

// NewS3 creates a new S3 registry client
func NewS3(authToken, bucket, region, prefix string) *S3Registry {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var auth AuthProvider
	if authToken != "" {
		auth = &HeaderAuth{Header: "Authorization", Value: authToken}
	}

	baseURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region)

	generic := &GenericHTTPRegistry{
		baseURL: baseURL,
		client:  client,
		auth:    auth,
	}

	return &S3Registry{
		GenericHTTPRegistry: generic,
		bucket:              bucket,
		region:              region,
		prefix:              prefix,
	}
}

// GetMetadata retrieves metadata for a ruleset
// S3 registries provide minimal metadata with discovered versions
func (r *S3Registry) GetMetadata(name string) (*Metadata, error) {
	versions, err := r.ListVersions(name)
	if err != nil {
		// Return minimal metadata if version discovery fails
		return &Metadata{
			Name:        name,
			Description: fmt.Sprintf("S3 registry: %s - version discovery failed", r.bucket),
			Versions:    []Version{},
			Repository:  r.baseURL,
			Extra: map[string]string{
				"bucket": r.bucket,
				"region": r.region,
				"prefix": r.prefix,
			},
		}, nil
	}

	versionList := make([]Version, len(versions))
	for i, v := range versions {
		versionList[i] = Version{Version: v}
	}

	return &Metadata{
		Name:        name,
		Description: fmt.Sprintf("S3 registry: %s with version discovery", r.bucket),
		Versions:    versionList,
		Repository:  r.baseURL,
		Extra: map[string]string{
			"bucket": r.bucket,
			"region": r.region,
			"prefix": r.prefix,
		},
	}, nil
}

// ListVersions returns all available versions for a ruleset
// Uses S3 prefix listing to discover available versions
func (r *S3Registry) ListVersions(name string) ([]string, error) {
	org, pkg := types.ParseRulesetName(name)
	basePath := "packages"
	if r.prefix != "" {
		basePath = r.prefix + "/packages"
	}

	var listPrefix string
	if org == "" {
		listPrefix = fmt.Sprintf("%s/%s/", basePath, pkg)
	} else {
		listPrefix = fmt.Sprintf("%s/%s/%s/", basePath, org, pkg)
	}

	// Use S3 list-objects-v2 API to get prefixes
	listURL := fmt.Sprintf("%s?list-type=2&prefix=%s&delimiter=/", r.baseURL, listPrefix)

	req, err := http.NewRequestWithContext(context.Background(), "GET", listURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %w", err)
	}

	if r.auth != nil {
		r.auth.SetAuth(req)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 prefixes: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("S3 list request failed with status %d", resp.StatusCode)
	}

	// Parse S3 XML response for CommonPrefixes
	versions, err := r.parseS3ListResponse(resp.Body, listPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to parse S3 response: %w", err)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found in S3 bucket")
	}

	return versions, nil
}

// HealthCheck verifies S3 registry connectivity
func (r *S3Registry) HealthCheck() error {
	req, err := http.NewRequestWithContext(context.Background(), "HEAD", r.baseURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	if r.auth != nil {
		r.auth.SetAuth(req)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("S3 registry unreachable: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case 200, 403: // 403 is OK for bucket existence check
		return nil
	case 404:
		return fmt.Errorf("S3 bucket %s not found in region %s", r.bucket, r.region)
	default:
		return fmt.Errorf("S3 registry returned status %d", resp.StatusCode)
	}
}

// S3ListResult represents S3 list-objects-v2 XML response
type S3ListResult struct {
	CommonPrefixes []S3CommonPrefix `xml:"CommonPrefixes"`
}

type S3CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// parseS3ListResponse parses S3 XML response to extract version directories
func (r *S3Registry) parseS3ListResponse(body io.Reader, listPrefix string) ([]string, error) {
	var result S3ListResult
	if err := xml.NewDecoder(body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode S3 XML: %w", err)
	}

	var versions []string
	for _, prefix := range result.CommonPrefixes {
		// Extract version from prefix (remove base path and trailing slash)
		version := strings.TrimPrefix(prefix.Prefix, listPrefix)
		version = strings.TrimSuffix(version, "/")
		if version != "" {
			versions = append(versions, version)
		}
	}

	return versions, nil
}
