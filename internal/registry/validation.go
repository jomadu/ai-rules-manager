package registry

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/jomadu/arm/internal/config"
)

// ValidateSource validates registry source configuration
func ValidateSource(name string, source config.Source) error {
	if source.Type == "" {
		source.Type = "generic" // Default type
	}

	regType := RegistryType(source.Type)
	switch regType {
	case RegistryTypeGeneric:
		return validateGenericSource(source)
	case RegistryTypeGitLab:
		return validateGitLabSource(source)
	case RegistryTypeS3:
		return validateS3Source(source)
	case RegistryTypeFilesystem:
		return validateFilesystemSource(source)
	default:
		return fmt.Errorf("unsupported registry type: %s", source.Type)
	}
}

func validateGenericSource(source config.Source) error {
	if source.URL == "" {
		return fmt.Errorf("URL is required for generic registry")
	}

	if _, err := url.Parse(source.URL); err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	return nil
}

func validateGitLabSource(source config.Source) error {
	if source.URL == "" {
		return fmt.Errorf("URL is required for GitLab registry")
	}

	if _, err := url.Parse(source.URL); err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if source.ProjectID == "" && source.GroupID == "" {
		return fmt.Errorf("either projectID or groupID is required for GitLab registry")
	}

	// Validate numeric IDs
	if source.ProjectID != "" {
		if !isNumeric(source.ProjectID) {
			return fmt.Errorf("projectID must be numeric")
		}
	}

	if source.GroupID != "" {
		if !isNumeric(source.GroupID) {
			return fmt.Errorf("groupID must be numeric")
		}
	}

	return nil
}

func validateS3Source(source config.Source) error {
	if source.Bucket == "" {
		return fmt.Errorf("bucket is required for S3 registry")
	}

	if source.Region == "" {
		return fmt.Errorf("region is required for S3 registry")
	}

	// Validate bucket name format
	if !isValidS3BucketName(source.Bucket) {
		return fmt.Errorf("invalid S3 bucket name format")
	}

	// Validate AWS region format
	if !isValidAWSRegion(source.Region) {
		return fmt.Errorf("invalid AWS region format")
	}

	return nil
}

func validateFilesystemSource(source config.Source) error {
	if source.Path == "" {
		return fmt.Errorf("path is required for filesystem registry")
	}

	// Check if path exists and is accessible
	if _, err := os.Stat(source.Path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("filesystem path does not exist: %s", source.Path)
		}
		return fmt.Errorf("cannot access filesystem path: %w", err)
	}

	return nil
}

// Helper functions
func isNumeric(s string) bool {
	matched, _ := regexp.MatchString(`^\d+$`, s)
	return matched
}

func isValidS3BucketName(bucket string) bool {
	// Basic S3 bucket name validation
	if len(bucket) < 3 || len(bucket) > 63 {
		return false
	}

	// Must start and end with lowercase letter or number
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9.-]*[a-z0-9]$`, bucket)
	if !matched {
		return false
	}

	// Cannot contain consecutive periods or period-dash combinations
	if strings.Contains(bucket, "..") || strings.Contains(bucket, ".-") || strings.Contains(bucket, "-.") {
		return false
	}

	return true
}

func isValidAWSRegion(region string) bool {
	// Basic AWS region format validation
	matched, _ := regexp.MatchString(`^[a-z]{2}-[a-z]+-\d+$`, region)
	return matched
}