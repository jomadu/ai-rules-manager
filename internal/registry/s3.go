package registry

import (
	"context"
	"fmt"
	"net/http"
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

func (r *S3Registry) buildDownloadURL(name, version string) string {
	org, pkg := types.ParseRulesetName(name)
	basePath := "packages"
	if r.prefix != "" {
		basePath = r.prefix + "/packages"
	}
	
	if org == "" {
		return fmt.Sprintf("%s/%s/%s/%s/%s-%s.tar.gz",
			r.baseURL, basePath, pkg, version, pkg, version)
	}
	return fmt.Sprintf("%s/%s/%s/%s/%s/%s-%s.tar.gz",
		r.baseURL, basePath, org, pkg, version, pkg, version)
}

// GetMetadata retrieves metadata for a ruleset
// S3 registries provide minimal metadata
func (r *S3Registry) GetMetadata(name string) (*Metadata, error) {
	return &Metadata{
		Name:        name,
		Description: fmt.Sprintf("S3 registry: %s", r.bucket),
		Versions:    []Version{},
		Repository:  r.baseURL,
		Extra: map[string]string{
			"bucket": r.bucket,
			"region": r.region,
			"prefix": r.prefix,
		},
	}, nil
}

// ListVersions returns all available versions for a ruleset
// S3 registries don't support version discovery without additional tooling
func (r *S3Registry) ListVersions(name string) ([]string, error) {
	return nil, fmt.Errorf("version listing not supported by S3 registry - specify exact version")
}

// HealthCheck verifies S3 registry connectivity
func (r *S3Registry) HealthCheck() error {
	req, err := http.NewRequestWithContext(context.Background(), "HEAD", r.baseURL, nil)
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
