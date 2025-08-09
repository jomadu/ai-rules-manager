package registry

import (
	"fmt"
)

// CreateRegistry creates a registry instance based on configuration
func CreateRegistry(config *RegistryConfig, auth *AuthConfig) (Registry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	switch config.Type {
	case "local":
		return NewLocalRegistry(config)
	case "git":
		return NewGitRegistry(config, auth)
	case "https":
		return NewHTTPSRegistry(config, auth)
	case "s3":
		return NewS3Registry(config, auth)
	case "gitlab":
		return NewGitLabRegistry(config, auth)
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", config.Type)
	}
}
