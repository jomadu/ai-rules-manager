package registry

import (
	"fmt"

	"github.com/max-dunn/ai-rules-manager/internal/cache"
	"github.com/max-dunn/ai-rules-manager/internal/config"
)

// CreateRegistry creates a registry instance based on configuration
func CreateRegistry(config *RegistryConfig, auth *AuthConfig) (Registry, error) {
	return CreateRegistryWithCache(config, auth, "")
}

// CreateRegistryWithCache creates a registry instance with cache manager injection
func CreateRegistryWithCache(config *RegistryConfig, auth *AuthConfig, cacheRoot string) (Registry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	switch config.Type {
	case "local":
		return NewLocalRegistry(config)
	case "git":
		if cacheRoot != "" {
			cacheManager := cache.NewGitRegistryCacheManager(cacheRoot)
			return NewGitRegistryWithCache(config, auth, cacheManager)
		}
		return NewGitRegistry(config, auth)
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
func CreateRegistryWithCacheConfig(registryConfig *RegistryConfig, auth *AuthConfig, cacheConfig *config.CacheConfig, registryName string) (Registry, error) {
	cacheRoot := ""
	if cacheConfig != nil {
		cacheRoot = config.GetCachePath()
	}

	return CreateRegistryWithCache(registryConfig, auth, cacheRoot)
}
