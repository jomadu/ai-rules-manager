package registry

import (
	"testing"
)

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
		name          string
		prefix        string
		rulesetPrefix string
		expected      string
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
