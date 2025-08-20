package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/max-dunn/ai-rules-manager/internal/cache"
)

// CreateRegistry creates a registry instance with hardcoded cache
func CreateRegistry(config *RegistryConfig) (Registry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	cacheRoot := GetCachePath()
	switch config.Type {
	case "git":
		cacheManager := cache.NewGitRegistryCacheManager(cacheRoot)
		return NewGitRegistryWithCache(config, cacheManager)
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", config.Type)
	}
}

// GetCachePath returns the hardcoded cache directory path
func GetCachePath() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".arm", "cache")
}
