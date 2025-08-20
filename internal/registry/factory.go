package registry

import (
	"fmt"

	"github.com/max-dunn/ai-rules-manager/internal/cache"
	"github.com/max-dunn/ai-rules-manager/internal/config"
)

// CreateRegistry creates a registry instance based on configuration
func CreateRegistry(config *RegistryConfig) (Registry, error) {
	return CreateRegistryWithCache(config, "")
}

// CreateRegistryWithCache creates a registry instance with cache manager injection
func CreateRegistryWithCache(config *RegistryConfig, cacheRoot string) (Registry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	switch config.Type {
	case "git":
		if cacheRoot != "" {
			cacheManager := cache.NewGitRegistryCacheManager(cacheRoot)
			return NewGitRegistryWithCache(config, cacheManager)
		}
		return NewGitRegistry(config)
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", config.Type)
	}
}

// CreateRegistryWithCacheConfig creates a registry instance with cache configuration
func CreateRegistryWithCacheConfig(registryConfig *RegistryConfig, cacheConfig *config.CacheConfig, registryName string) (Registry, error) {
	cacheRoot := ""
	if cacheConfig != nil {
		cacheRoot = config.GetCachePath()
	}

	return CreateRegistryWithCache(registryConfig, cacheRoot)
}
