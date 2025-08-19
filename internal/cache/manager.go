package cache

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"
)

// RegistryCacheManager defines the base interface for all registry cache managers
type RegistryCacheManager interface {
	// Store stores files for a registry with given identifier and version
	Store(registryURL string, identifier []string, version string, files map[string][]byte) error

	// Get retrieves files for a registry with given identifier and version
	Get(registryURL string, identifier []string, version string) (map[string][]byte, error)

	// GetPath returns the filesystem path for a registry with given identifier and version
	GetPath(registryURL string, identifier []string, version string) (string, error)

	// IsValid checks if cached content is still valid based on TTL
	IsValid(registryURL string, ttl time.Duration) (bool, error)

	// Cleanup removes expired cache entries based on TTL and size limits
	Cleanup(ttl time.Duration, maxSize int64) error
}

// GitRegistryCacheManager defines the interface for Git-specific registry cache managers
type GitRegistryCacheManager interface {
	RegistryCacheManager

	// StoreRuleset stores ruleset files for a Git registry with patterns and commit hash
	StoreRuleset(registryURL string, patterns []string, commitHash string, files map[string][]byte) error

	// GetRuleset retrieves ruleset files for a Git registry with patterns and commit hash
	GetRuleset(registryURL string, patterns []string, commitHash string) (map[string][]byte, error)

	// GetRepositoryPath returns the path to the Git repository clone
	GetRepositoryPath(registryURL string) (string, error)
}

// RulesetRegistryCacheManager defines the interface for non-Git registry cache managers
type RulesetRegistryCacheManager interface {
	RegistryCacheManager

	// StoreRuleset stores ruleset files for a non-Git registry with ruleset name and version
	StoreRuleset(registryURL, rulesetName, version string, files map[string][]byte) error

	// GetRuleset retrieves ruleset files for a non-Git registry with ruleset name and version
	GetRuleset(registryURL, rulesetName, version string) (map[string][]byte, error)
}

// Factory functions for creating registry cache managers

// NewGitRegistryCacheManager creates a new Git registry cache manager
func NewGitRegistryCacheManager(cacheRoot string) GitRegistryCacheManager {
	// Implementation will be added in git_cache.go
	return nil
}

// NewS3RegistryCacheManager creates a new S3 registry cache manager
func NewS3RegistryCacheManager(cacheRoot string) RulesetRegistryCacheManager {
	// Implementation will be added in ruleset_cache.go
	return nil
}

// NewHTTPSRegistryCacheManager creates a new HTTPS registry cache manager
func NewHTTPSRegistryCacheManager(cacheRoot string) RulesetRegistryCacheManager {
	// Implementation will be added in ruleset_cache.go
	return nil
}

// NewLocalRegistryCacheManager creates a new Local registry cache manager
func NewLocalRegistryCacheManager(cacheRoot string) RulesetRegistryCacheManager {
	// Implementation will be added in ruleset_cache.go
	return nil
}

// Utility functions for cache key generation

// GenerateRegistryKey generates a SHA-256 hash key for a registry
func GenerateRegistryKey(registryType, registryURL string) string {
	normalizedURL := normalizeURL(registryURL)
	cacheInput := fmt.Sprintf("%s:%s", registryType, normalizedURL)
	hash := sha256.Sum256([]byte(cacheInput))
	return fmt.Sprintf("%x", hash)
}

// GeneratePatternsKey generates a SHA-256 hash key for patterns (Git registries)
func GeneratePatternsKey(patterns []string) string {
	if len(patterns) == 0 {
		return GenerateStringKey("__EMPTY__")
	}

	// Normalize patterns: sort and trim whitespace for consistency
	normalizedPatterns := make([]string, len(patterns))
	for i, pattern := range patterns {
		normalizedPatterns[i] = strings.TrimSpace(pattern)
	}
	sort.Strings(normalizedPatterns)

	patternsStr := strings.Join(normalizedPatterns, ",")
	return GenerateStringKey(patternsStr)
}

// GenerateRulesetKey generates a SHA-256 hash key for a ruleset name (non-Git registries)
func GenerateRulesetKey(rulesetName string) string {
	return GenerateStringKey(strings.TrimSpace(rulesetName))
}

// GenerateStringKey generates a SHA-256 hash for any string
func GenerateStringKey(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}

// normalizeURL performs minimal URL normalization
func normalizeURL(url string) string {
	// Minimal normalization: trim whitespace and ensure consistent protocol
	normalized := strings.TrimSpace(url)

	// Remove trailing slash for consistency
	if strings.HasSuffix(normalized, "/") {
		normalized = strings.TrimSuffix(normalized, "/")
	}

	return normalized
}
