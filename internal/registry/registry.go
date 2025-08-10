package registry

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
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

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Token      string `json:"token"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Profile    string `json:"profile"`     // For AWS profiles
	Region     string `json:"region"`      // For AWS regions
	APIType    string `json:"api_type"`    // For API-specific auth
	APIVersion string `json:"api_version"` // For API versioning
}

// RegistryConfig contains registry configuration
type RegistryConfig struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	URL          string                 `json:"url"`
	Auth         *AuthConfig            `json:"auth,omitempty"`
	Concurrency  int                    `json:"concurrency"`
	RateLimit    string                 `json:"rate_limit"`
	Timeout      time.Duration          `json:"timeout"`
	RetryConfig  *RetryConfig           `json:"retry_config,omitempty"`
	CustomConfig map[string]interface{} `json:"custom_config,omitempty"`
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	MaxAttempts       int           `json:"max_attempts"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	MaxBackoff        time.Duration `json:"max_backoff"`
	RetryableErrors   []string      `json:"retryable_errors"`
}

// AuthProvider handles authentication for registries
type AuthProvider interface {
	// GetCredentials returns credentials for the given registry
	GetCredentials(registryName string) (*AuthConfig, error)

	// RefreshCredentials refreshes expired credentials
	RefreshCredentials(registryName string) (*AuthConfig, error)
}

// DefaultAuthProvider implements common authentication patterns
type DefaultAuthProvider struct {
	configs map[string]*AuthConfig
}

// NewDefaultAuthProvider creates a new default auth provider
func NewDefaultAuthProvider() *DefaultAuthProvider {
	return &DefaultAuthProvider{
		configs: make(map[string]*AuthConfig),
	}
}

// SetAuth sets authentication configuration for a registry
func (p *DefaultAuthProvider) SetAuth(registryName string, auth *AuthConfig) {
	p.configs[registryName] = auth
}

// GetCredentials returns credentials for the given registry
func (p *DefaultAuthProvider) GetCredentials(registryName string) (*AuthConfig, error) {
	auth, exists := p.configs[registryName]
	if !exists {
		return &AuthConfig{}, nil // Return empty auth if not configured
	}

	// Expand environment variables in auth config
	expandedAuth := &AuthConfig{
		Token:      expandEnvVars(auth.Token),
		Username:   expandEnvVars(auth.Username),
		Password:   expandEnvVars(auth.Password),
		Profile:    expandEnvVars(auth.Profile),
		Region:     expandEnvVars(auth.Region),
		APIType:    expandEnvVars(auth.APIType),
		APIVersion: expandEnvVars(auth.APIVersion),
	}

	return expandedAuth, nil
}

// RefreshCredentials refreshes expired credentials
func (p *DefaultAuthProvider) RefreshCredentials(registryName string) (*AuthConfig, error) {
	// For now, just return the current credentials
	// This can be extended for OAuth flows, AWS credential refresh, etc.
	return p.GetCredentials(registryName)
}

// expandEnvVars expands environment variables in strings
func expandEnvVars(s string) string {
	if s == "" {
		return s
	}

	// Handle $VAR and ${VAR} patterns
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		varName := s[2 : len(s)-1]
		return os.Getenv(varName)
	}
	if strings.HasPrefix(s, "$") {
		varName := s[1:]
		return os.Getenv(varName)
	}

	return s
}

// ValidateRegistryConfig validates a registry configuration
func ValidateRegistryConfig(config *RegistryConfig) error {
	if config.Name == "" {
		return fmt.Errorf("registry name cannot be empty")
	}
	if config.Type == "" {
		return fmt.Errorf("registry type cannot be empty")
	}

	validTypes := []string{"git", "https", "s3", "gitlab", "local"}
	if !contains(validTypes, config.Type) {
		return fmt.Errorf("unsupported registry type: %s", config.Type)
	}

	// Type-specific validation
	switch config.Type {
	case "git", "https", "gitlab":
		if config.URL == "" {
			return fmt.Errorf("%s registry requires URL", config.Type)
		}
		if !strings.HasPrefix(config.URL, "https://") {
			return fmt.Errorf("%s registry URL must use HTTPS", config.Type)
		}
	case "s3":
		if config.Auth == nil || config.Auth.Region == "" {
			return fmt.Errorf("s3 registry requires region in auth config")
		}
	case "local":
		if config.URL == "" {
			return fmt.Errorf("local registry requires path in URL field")
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
