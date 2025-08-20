package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/max-dunn/ai-rules-manager/internal/config"
)

// Registry defines the interface for all registry implementations
type Registry interface {
	// GetRulesets returns available rulesets matching the given patterns
	GetRulesets(ctx context.Context, patterns []string) ([]RulesetInfo, error)

	// GetRuleset returns detailed information about a specific ruleset
	GetRuleset(ctx context.Context, name, version string) (*RulesetInfo, error)

	// DownloadRuleset downloads a ruleset to the specified directory
	DownloadRuleset(ctx context.Context, name, version, destDir string) error

	// DownloadRulesetWithPatterns downloads a ruleset with pattern matching (for Git registries)
	DownloadRulesetWithPatterns(ctx context.Context, name, version, destDir string, patterns []string) error

	// GetVersions returns available versions for a ruleset
	GetVersions(ctx context.Context, name string) ([]string, error)

	// GetType returns the registry type
	GetType() string

	// GetName returns the registry name
	GetName() string

	// Close cleans up any resources
	Close() error
}

// Searcher defines the optional search interface for registries
type Searcher interface {
	// Search returns rulesets matching the query
	Search(ctx context.Context, query string) ([]SearchResult, error)
}

// SearchResult contains minimal search result information
type SearchResult struct {
	RulesetName  string `json:"ruleset_name"`
	RegistryName string `json:"registry_name"`
	Match        string `json:"match"`
}

// RulesetInfo contains metadata about a ruleset
type RulesetInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Tags        []string          `json:"tags"`
	Patterns    []string          `json:"patterns"`
	Metadata    map[string]string `json:"metadata"`
	Registry    string            `json:"registry"`
	Type        string            `json:"type"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// RegistryConfig contains registry configuration
type RegistryConfig struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`

	Concurrency  int                    `json:"concurrency"`
	RateLimit    string                 `json:"rate_limit"`
	Timeout      time.Duration          `json:"timeout"`
	RetryConfig  *RetryConfig           `json:"retry_config,omitempty"`
	CustomConfig map[string]interface{} `json:"custom_config,omitempty"`
}

// ResolvePath resolves the registry path using the config package
func (c *RegistryConfig) ResolvePath(path string) (string, error) {
	return config.ResolvePath(path)
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	MaxAttempts       int           `json:"max_attempts"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	MaxBackoff        time.Duration `json:"max_backoff"`
	RetryableErrors   []string      `json:"retryable_errors"`
}

// ValidateRegistryConfig validates a registry configuration
func ValidateRegistryConfig(config *RegistryConfig) error {
	if config.Name == "" {
		return fmt.Errorf("registry name cannot be empty")
	}
	if config.Type == "" {
		return fmt.Errorf("registry type cannot be empty")
	}

	validTypes := []string{"git"}
	if !contains(validTypes, config.Type) {
		return fmt.Errorf("unsupported registry type: %s", config.Type)
	}

	// Type-specific validation
	if config.Type == "git" {
		if config.URL == "" {
			return fmt.Errorf("%s registry requires URL", config.Type)
		}
		// Git registries support HTTPS, SSH, and local file paths
		if !strings.HasPrefix(config.URL, "https://") &&
			!strings.HasPrefix(config.URL, "git@") &&
			!strings.HasPrefix(config.URL, "ssh://") &&
			!strings.HasPrefix(config.URL, "file://") &&
			!strings.HasPrefix(config.URL, "/") {
			return fmt.Errorf("%s registry URL must use HTTPS, SSH, or local path", config.Type)
		}
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
