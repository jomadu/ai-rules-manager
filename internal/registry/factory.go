package registry

import (
	"fmt"

	"github.com/max-dunn/ai-rules-manager/internal/cache"
	"github.com/max-dunn/ai-rules-manager/internal/config"
)

// CreateRegistry creates a registry instance based on configuration
func CreateRegistry(config *RegistryConfig, auth *AuthConfig) (Registry, error) {
	return CreateRegistryWithCache(config, auth, nil)
}

// CreateRegistryWithCache creates a registry instance with cache manager injection
func CreateRegistryWithCache(config *RegistryConfig, auth *AuthConfig, cacheManager cache.Manager) (Registry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	switch config.Type {
	case "local":
		return NewLocalRegistry(config)
	case "git":
		return NewGitRegistryWithCache(config, auth, cacheManager)
	case "git-local":
		return NewGitLocalRegistry(config, auth)
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

// CreateRegistryWithCacheConfig creates a registry instance with cache configuration
func CreateRegistryWithCacheConfig(registryConfig *RegistryConfig, auth *AuthConfig, cacheManager cache.Manager, cacheConfig *config.CacheConfig, registryName string) (Registry, error) {
	// Perform cache cleanup if cache manager is available
	if cacheManager != nil && cacheConfig != nil {
		// Clean up expired entries
		_ = cacheManager.CleanupExpired(cacheConfig.TTL)

		// Clean up oversized cache
		_ = cacheManager.CleanupOversized(cacheConfig.MaxSize)
	}

	return CreateRegistryWithCache(registryConfig, auth, cacheManager)
}
