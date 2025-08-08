package registry

import (
	"context"
	"strings"
	"testing"
)

func TestNewS3Registry(t *testing.T) {
	config := &RegistryConfig{
		Name: "test-s3",
		Type: "s3",
		URL:  "my-test-bucket",
		Auth: &AuthConfig{
			Region: "us-east-1",
		},
	}
	auth := &AuthConfig{
		Region: "us-east-1",
	}

	// Skip actual AWS connection for unit tests
	t.Skip("S3 registry requires AWS credentials for testing")

	registry, err := NewS3Registry(config, auth)
	if err != nil {
		t.Fatalf("Failed to create S3 registry: %v", err)
	}

	if registry.GetType() != "s3" {
		t.Errorf("Expected type 's3', got %q", registry.GetType())
	}
	if registry.GetName() != "test-s3" {
		t.Errorf("Expected name 'test-s3', got %q", registry.GetName())
	}
}

func TestNewS3RegistryInvalidConfig(t *testing.T) {
	config := &RegistryConfig{
		Name: "test-s3",
		Type: "s3",
		URL:  "my-bucket",
		// Missing Auth with Region
	}
	auth := &AuthConfig{}

	_, err := NewS3Registry(config, auth)
	if err == nil {
		t.Error("Expected error for missing region")
	}
}

func TestParseBucketURL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedBucket string
		expectedPrefix string
	}{
		{
			name:           "bucket only",
			url:            "my-bucket",
			expectedBucket: "my-bucket",
			expectedPrefix: "",
		},
		{
			name:           "bucket with prefix",
			url:            "my-bucket/rules",
			expectedBucket: "my-bucket",
			expectedPrefix: "rules/",
		},
		{
			name:           "bucket with nested prefix",
			url:            "my-bucket/path/to/rules",
			expectedBucket: "my-bucket",
			expectedPrefix: "path/to/rules/",
		},
		{
			name:           "bucket with trailing slash",
			url:            "my-bucket/rules/",
			expectedBucket: "my-bucket",
			expectedPrefix: "rules/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, prefix := parseBucketURL(tt.url)
			if bucket != tt.expectedBucket {
				t.Errorf("Expected bucket %q, got %q", tt.expectedBucket, bucket)
			}
			if prefix != tt.expectedPrefix {
				t.Errorf("Expected prefix %q, got %q", tt.expectedPrefix, prefix)
			}
		})
	}
}

func TestExtractRulesetName(t *testing.T) {
	tests := []struct {
		name     string
		registry *S3Registry
		prefix   string
		expected string
	}{
		{
			name:     "simple ruleset name",
			registry: &S3Registry{prefix: ""},
			prefix:   "my-rules/",
			expected: "my-rules",
		},
		{
			name:     "with bucket prefix",
			registry: &S3Registry{prefix: "registries/"},
			prefix:   "registries/python-rules/",
			expected: "python-rules",
		},
		{
			name:     "nested prefix",
			registry: &S3Registry{prefix: "arm/rules/"},
			prefix:   "arm/rules/typescript-rules/",
			expected: "typescript-rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.registry.extractRulesetName(tt.prefix)
			if result != tt.expected {
				t.Errorf("extractRulesetName(%q) = %q, want %q", tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestExtractVersionFromPrefix(t *testing.T) {
	tests := []struct {
		name           string
		prefix         string
		rulesetPrefix  string
		expected       string
	}{
		{
			name:          "simple version",
			prefix:        "my-rules/1.0.0/",
			rulesetPrefix: "my-rules/",
			expected:      "1.0.0",
		},
		{
			name:          "semver version",
			prefix:        "python-rules/2.1.3/",
			rulesetPrefix: "python-rules/",
			expected:      "2.1.3",
		},
		{
			name:          "with bucket prefix",
			prefix:        "registries/my-rules/1.5.0/",
			rulesetPrefix: "registries/my-rules/",
			expected:      "1.5.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &S3Registry{}
			result := registry.extractVersionFromPrefix(tt.prefix, tt.rulesetPrefix)
			if result != tt.expected {
				t.Errorf("extractVersionFromPrefix(%q, %q) = %q, want %q", tt.prefix, tt.rulesetPrefix, result, tt.expected)
			}
		})
	}
}

func TestS3GetVersions(t *testing.T) {
	// Skip test that requires actual S3 client
	t.Skip("S3 GetVersions requires AWS credentials and S3 client for testing")
}

func TestS3RegistryClose(t *testing.T) {
	registry := &S3Registry{
		config: &RegistryConfig{
			Name: "test-s3",
			Type: "s3",
		},
	}

	err := registry.Close()
	if err != nil {
		t.Errorf("Expected no error from Close(), got: %v", err)
	}
}

// TestS3RegistryIntegration tests would require actual AWS credentials
// and S3 bucket setup, so they are skipped in unit tests
func TestS3RegistryIntegration(t *testing.T) {
	t.Skip("Integration tests require AWS credentials and S3 bucket")

	// This would test:
	// - Actual S3 connection
	// - Listing objects in bucket
	// - Downloading objects
	// - AWS credential chain
	// - Different AWS profiles
	// - Cross-region access
}

func TestLoadAWSConfigValidation(t *testing.T) {
	// Test that the function signature is correct
	// Actual testing would require AWS credentials
	ctx := context.Background()
	auth := &AuthConfig{
		Region: "us-east-1",
	}

	// This will fail without AWS credentials, but validates the function exists
	_, err := loadAWSConfig(ctx, auth)
	// We expect an error in test environment without AWS setup
	if err == nil {
		t.Skip("AWS credentials available, skipping validation test")
	}

	// Validate error message contains expected AWS-related content
	if !strings.Contains(err.Error(), "AWS") && !strings.Contains(err.Error(), "credential") {
		t.Errorf("Expected AWS-related error, got: %v", err)
	}
}
