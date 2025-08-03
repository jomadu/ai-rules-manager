package registry

import (
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
}

// NewS3 creates a new S3 registry client
func NewS3(authToken, bucket, region string) *S3Registry {
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
	}
}

func (r *S3Registry) buildDownloadURL(name, version string) string {
	org, pkg := types.ParseRulesetName(name)
	if org == "" {
		return fmt.Sprintf("%s/packages/%s/%s/%s-%s.tar.gz",
			r.baseURL, pkg, version, pkg, version)
	}
	return fmt.Sprintf("%s/packages/%s/%s/%s/%s-%s.tar.gz",
		r.baseURL, org, pkg, version, pkg, version)
}

func (r *S3Registry) buildMetadataURL(name string) string {
	org, pkg := types.ParseRulesetName(name)
	if org == "" {
		return fmt.Sprintf("%s/packages/%s/metadata.json", r.baseURL, pkg)
	}
	return fmt.Sprintf("%s/packages/%s/%s/metadata.json", r.baseURL, org, pkg)
}
