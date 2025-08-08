package registry

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Registry implements the Registry interface for S3 buckets
type S3Registry struct {
	config *RegistryConfig
	auth   *AuthConfig
	client *s3.Client
	bucket string
	prefix string
}

// NewS3Registry creates a new S3 registry instance
func NewS3Registry(config *RegistryConfig, auth *AuthConfig) (*S3Registry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Load AWS configuration
	awsConfig, err := loadAWSConfig(context.Background(), auth)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsConfig)

	// Parse bucket and prefix from URL
	bucket, prefix := parseBucketURL(config.URL)

	return &S3Registry{
		config: config,
		auth:   auth,
		client: client,
		bucket: bucket,
		prefix: prefix,
	}, nil
}

// GetRulesets returns available rulesets matching the given patterns
func (s *S3Registry) GetRulesets(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	// List objects in the bucket to discover rulesets
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Delimiter: aws.String("/"),
	}

	if s.prefix != "" {
		input.Prefix = aws.String(s.prefix)
	}

	rulesetMap := make(map[string]*RulesetInfo)
	paginator := s3.NewListObjectsV2Paginator(s.client, input)

	// First, get all ruleset directories
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		// Process common prefixes (directories)
		for _, prefix := range page.CommonPrefixes {
			prefixStr := aws.ToString(prefix.Prefix)
			rulesetName := s.extractRulesetName(prefixStr)
			if rulesetName != "" {
				rulesetMap[rulesetName] = &RulesetInfo{
					Name:     rulesetName,
					Registry: s.config.Name,
					Type:     "s3",
					Metadata: map[string]string{
						"bucket": s.bucket,
						"prefix": prefixStr,
					},
				}
			}
		}
	}

	// Now get versions for each ruleset to find the latest
	for name, ruleset := range rulesetMap {
		versions, err := s.GetVersions(ctx, name)
		if err != nil {
			continue // Skip rulesets we can't get versions for
		}
		if len(versions) > 0 {
			ruleset.Version = versions[0] // First version is latest
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
func (s *S3Registry) GetRuleset(ctx context.Context, name, version string) (*RulesetInfo, error) {
	rulesets, err := s.GetRulesets(ctx, []string{name + "*"})
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

// DownloadRuleset downloads a ruleset to the specified directory
func (s *S3Registry) DownloadRuleset(ctx context.Context, name, version, destDir string) error {
	// Construct the S3 key for the ruleset tarball
	key := s.prefix + name + "/" + version + "/ruleset.tar.gz"

	// Download the tarball
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to download S3 object %s: %w", key, err)
	}
	defer result.Body.Close()

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Create destination file for the tarball
	tarballPath := filepath.Join(destDir, "ruleset.tar.gz")
	destFile, err := os.Create(tarballPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy tarball content
	_, err = io.Copy(destFile, result.Body)
	return err
}

// GetVersions returns available versions for a ruleset
func (s *S3Registry) GetVersions(ctx context.Context, name string) ([]string, error) {
	// List version directories for the ruleset
	rulesetPrefix := s.prefix + name + "/"
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(rulesetPrefix),
		Delimiter: aws.String("/"),
	}

	var versions []string
	paginator := s3.NewListObjectsV2Paginator(s.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 versions: %w", err)
		}

		// Process common prefixes (version directories)
		for _, prefix := range page.CommonPrefixes {
			prefixStr := aws.ToString(prefix.Prefix)
			version := s.extractVersionFromPrefix(prefixStr, rulesetPrefix)
			if version != "" {
				versions = append(versions, version)
			}
		}
	}

	// Sort versions (latest first) - simple string sort for now
	// TODO: Implement proper semver sorting
	if len(versions) == 0 {
		return []string{"latest"}, nil
	}

	return versions, nil
}

// GetType returns the registry type
func (s *S3Registry) GetType() string {
	return "s3"
}

// GetName returns the registry name
func (s *S3Registry) GetName() string {
	return s.config.Name
}

// Close cleans up any resources
func (s *S3Registry) Close() error {
	return nil
}

// loadAWSConfig loads AWS configuration with credential chain
func loadAWSConfig(ctx context.Context, auth *AuthConfig) (aws.Config, error) {
	// Start with default config loading (uses credential chain)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(auth.Region))
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Override with profile if specified
	if auth.Profile != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(auth.Region),
			config.WithSharedConfigProfile(auth.Profile),
		)
		if err != nil {
			return aws.Config{}, fmt.Errorf("failed to load AWS config with profile %s: %w", auth.Profile, err)
		}
	}

	return cfg, nil
}

// parseBucketURL parses bucket name and prefix from URL
func parseBucketURL(url string) (bucket, prefix string) {
	// Handle formats like:
	// "my-bucket" -> bucket="my-bucket", prefix=""
	// "my-bucket/path/to/rules" -> bucket="my-bucket", prefix="path/to/rules"
	parts := strings.SplitN(url, "/", 2)
	bucket = parts[0]
	if len(parts) > 1 {
		prefix = parts[1]
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
	}
	return bucket, prefix
}

// extractRulesetName extracts ruleset name from S3 prefix
func (s *S3Registry) extractRulesetName(prefix string) string {
	// Remove the configured prefix
	if s.prefix != "" && strings.HasPrefix(prefix, s.prefix) {
		prefix = strings.TrimPrefix(prefix, s.prefix)
	}
	// Remove trailing slash and return the ruleset name
	return strings.TrimSuffix(prefix, "/")
}

// extractVersionFromPrefix extracts version from S3 prefix
func (s *S3Registry) extractVersionFromPrefix(prefix, rulesetPrefix string) string {
	// Remove the ruleset prefix to get version/
	version := strings.TrimPrefix(prefix, rulesetPrefix)
	// Remove trailing slash
	return strings.TrimSuffix(version, "/")
}
