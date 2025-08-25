package provider

import (
	"github.com/jomadu/ai-rules-manager/pkg/cache"
	"github.com/jomadu/ai-rules-manager/pkg/config"
	"github.com/jomadu/ai-rules-manager/pkg/registry"
	"github.com/jomadu/ai-rules-manager/pkg/version"
)

// RegistryProvider creates registry-specific components
type RegistryProvider interface {
	CreateRegistry(config *config.RegistryConfig) (registry.Registry, error)
	CreateVersionResolver() (version.VersionResolver, error)
	CreateContentResolver() (version.ContentResolver, error)
	CreateCacheKeyGenerator() (cache.CacheKeyGenerator, error)
}

// GitRegistryProvider implements RegistryProvider for Git repositories
type GitRegistryProvider struct{}

func NewGitRegistryProvider() *GitRegistryProvider {
	return &GitRegistryProvider{}
}

func (g *GitRegistryProvider) CreateRegistry(config *config.RegistryConfig) (registry.Registry, error) {
	// TODO: implement
	return registry.NewGitRegistry(config.URL), nil
}

func (g *GitRegistryProvider) CreateVersionResolver() (version.VersionResolver, error) {
	// TODO: implement
	return version.NewSemVerResolver(), nil
}

func (g *GitRegistryProvider) CreateContentResolver() (version.ContentResolver, error) {
	// TODO: implement
	return version.NewGitContentResolver(), nil
}

func (g *GitRegistryProvider) CreateCacheKeyGenerator() (cache.CacheKeyGenerator, error) {
	// TODO: implement
	return cache.NewGitCacheKeyGenerator(), nil
}
