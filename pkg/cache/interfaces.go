package cache

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/jomadu/ai-rules-manager/pkg/registry"
)

// CacheKeyGenerator generates cache keys for different registry types
type CacheKeyGenerator interface {
	RegistryKey(url string) string
	RulesetKey(rulesetName string, selector registry.ContentSelector) string
	VersionKey(versionRef registry.VersionRef) string
}

// Cache provides content-addressable storage with TTL
type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, content []byte, ttl time.Duration) error
	Delete(key string) error
	Clear() error
}

// GitCacheKeyGenerator implements CacheKeyGenerator for Git registries
type GitCacheKeyGenerator struct{}

func NewGitCacheKeyGenerator() *GitCacheKeyGenerator {
	return &GitCacheKeyGenerator{}
}

func (g *GitCacheKeyGenerator) RegistryKey(url string) string {
	hash := sha256.Sum256([]byte(url + "git"))
	return fmt.Sprintf("%x", hash)
}

func (g *GitCacheKeyGenerator) RulesetKey(rulesetName string, selector registry.ContentSelector) string {
	hash := sha256.Sum256([]byte(rulesetName + selector.String()))
	return fmt.Sprintf("%x", hash)
}

func (g *GitCacheKeyGenerator) VersionKey(versionRef registry.VersionRef) string {
	return versionRef.ID // Use commit hash directly for Git
}

// FileCache implements Cache for file-based storage
type FileCache struct {
	basePath string
}

func NewFileCache(basePath string) *FileCache {
	return &FileCache{basePath: basePath}
}

func (f *FileCache) Get(key string) ([]byte, error) {
	// TODO: implement
	return nil, nil
}

func (f *FileCache) Set(key string, content []byte, ttl time.Duration) error {
	// TODO: implement
	return nil
}

func (f *FileCache) Delete(key string) error {
	// TODO: implement
	return nil
}

func (f *FileCache) Clear() error {
	// TODO: implement
	return nil
}
