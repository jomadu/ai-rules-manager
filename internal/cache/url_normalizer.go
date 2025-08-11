package cache

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// URLNormalizer provides comprehensive URL normalization for different registry types
type URLNormalizer struct{}

// NewURLNormalizer creates a new URL normalizer
func NewURLNormalizer() *URLNormalizer {
	return &URLNormalizer{}
}

// NormalizeURL normalizes a registry URL for consistent hashing across all registry types
func (n *URLNormalizer) NormalizeURL(registryType, rawURL string) string {
	switch registryType {
	case "git":
		return n.normalizeGitURL(rawURL)
	case "gitlab":
		return n.normalizeGitLabURL(rawURL)
	case "s3":
		return n.normalizeS3URL(rawURL)
	case "https":
		return n.normalizeHTTPSURL(rawURL)
	case "local":
		return n.normalizeLocalURL(rawURL)
	default:
		return n.normalizeGenericURL(rawURL)
	}
}

// normalizeGitURL normalizes Git repository URLs
func (n *URLNormalizer) normalizeGitURL(rawURL string) string {
	normalized := strings.ToLower(strings.TrimSpace(rawURL))

	// Remove trailing slashes
	normalized = strings.TrimSuffix(normalized, "/")

	// Convert SSH URLs to HTTPS format
	switch {
	case strings.HasPrefix(normalized, "git@github.com:"):
		normalized = strings.Replace(normalized, "git@github.com:", "https://github.com/", 1)
	case strings.HasPrefix(normalized, "git@gitlab.com:"):
		normalized = strings.Replace(normalized, "git@gitlab.com:", "https://gitlab.com/", 1)
	case strings.HasPrefix(normalized, "git@bitbucket.org:"):
		normalized = strings.Replace(normalized, "git@bitbucket.org:", "https://bitbucket.org/", 1)
	}

	// Handle generic SSH format: git@host:path
	sshPattern := regexp.MustCompile(`^git@([^:]+):(.+)$`)
	if matches := sshPattern.FindStringSubmatch(normalized); len(matches) == 3 {
		normalized = "https://" + matches[1] + "/" + matches[2]
	}

	// Remove .git suffix
	normalized = strings.TrimSuffix(normalized, ".git")

	// Ensure https:// prefix for URLs that look like domains
	if !strings.HasPrefix(normalized, "http://") && !strings.HasPrefix(normalized, "https://") {
		if strings.Contains(normalized, ".") && !strings.HasPrefix(normalized, "/") {
			normalized = "https://" + normalized
		}
	}

	return normalized
}

// normalizeGitLabURL normalizes GitLab registry URLs
func (n *URLNormalizer) normalizeGitLabURL(rawURL string) string {
	normalized := strings.ToLower(strings.TrimSpace(rawURL))

	// Remove trailing slashes
	normalized = strings.TrimSuffix(normalized, "/")

	// Ensure https:// prefix
	if !strings.HasPrefix(normalized, "http://") && !strings.HasPrefix(normalized, "https://") {
		normalized = "https://" + normalized
	}

	// Normalize multiple slashes in path only (not in protocol)
	if strings.Contains(normalized, "://") {
		parts := strings.SplitN(normalized, "://", 2)
		if len(parts) == 2 {
			protocol := parts[0]
			rest := parts[1]
			// Only normalize slashes in the path part
			rest = regexp.MustCompile(`/+`).ReplaceAllString(rest, "/")
			normalized = protocol + "://" + rest
		}
	}

	return normalized
}

// normalizeS3URL normalizes S3 bucket URLs
func (n *URLNormalizer) normalizeS3URL(rawURL string) string {
	normalized := strings.TrimSpace(rawURL)

	// Remove s3:// prefix if present
	normalized = strings.TrimPrefix(normalized, "s3://")

	// Remove trailing slashes
	normalized = strings.TrimSuffix(normalized, "/")

	// Normalize path separators
	normalized = strings.ReplaceAll(normalized, "\\", "/")

	// Remove duplicate slashes
	normalized = regexp.MustCompile(`/+`).ReplaceAllString(normalized, "/")

	return normalized
}

// normalizeHTTPSURL normalizes HTTPS registry URLs
func (n *URLNormalizer) normalizeHTTPSURL(rawURL string) string {
	normalized := strings.TrimSpace(rawURL)

	// Parse URL to normalize components
	if parsedURL, err := url.Parse(normalized); err == nil {
		// Ensure scheme is present
		if parsedURL.Scheme == "" {
			parsedURL.Scheme = "https"
		}

		// Normalize scheme to lowercase
		parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)

		// Normalize host to lowercase
		parsedURL.Host = strings.ToLower(parsedURL.Host)

		// Clean path but preserve case
		if parsedURL.Path != "" {
			parsedURL.Path = filepath.Clean(parsedURL.Path)
			// Remove trailing slash unless it's root
			if parsedURL.Path != "/" {
				parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/")
			}
		}

		// Remove fragment
		parsedURL.Fragment = ""

		normalized = parsedURL.String()
	} else {
		// Fallback for malformed URLs
		normalized = strings.ToLower(normalized)
		normalized = strings.TrimSuffix(normalized, "/")
		if !strings.HasPrefix(normalized, "http://") && !strings.HasPrefix(normalized, "https://") {
			normalized = "https://" + normalized
		}
	}

	return normalized
}

// normalizeLocalURL normalizes local file system paths
func (n *URLNormalizer) normalizeLocalURL(rawURL string) string {
	normalized := strings.TrimSpace(rawURL)

	// Remove file:// prefix if present
	normalized = strings.TrimPrefix(normalized, "file://")

	// Convert to absolute path and clean
	if absPath, err := filepath.Abs(normalized); err == nil {
		normalized = absPath
	} else {
		// Fallback to cleaning the path
		normalized = filepath.Clean(normalized)
	}

	// Ensure consistent path separators (use forward slashes)
	normalized = filepath.ToSlash(normalized)

	return normalized
}

// normalizeGenericURL provides basic normalization for unknown registry types
func (n *URLNormalizer) normalizeGenericURL(rawURL string) string {
	normalized := strings.ToLower(strings.TrimSpace(rawURL))

	// Remove trailing slashes
	normalized = strings.TrimSuffix(normalized, "/")

	// Normalize multiple slashes but preserve protocol separators
	if strings.Contains(normalized, "://") {
		parts := strings.SplitN(normalized, "://", 2)
		if len(parts) == 2 {
			protocol := parts[0]
			rest := parts[1]
			// Only normalize slashes in the path part
			rest = regexp.MustCompile(`/+`).ReplaceAllString(rest, "/")
			normalized = protocol + "://" + rest
		}
	} else {
		// No protocol, safe to normalize all slashes
		normalized = regexp.MustCompile(`/+`).ReplaceAllString(normalized, "/")
	}

	return normalized
}
