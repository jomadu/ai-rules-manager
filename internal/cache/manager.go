package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Manager defines the interface for content-based cache management
type Manager interface {
	// GetCachePath returns the cache directory path for a registry configuration
	GetCachePath(registryType, registryURL string) (string, error)

	// GetRepositoryPath returns the repository subdirectory path for git clones
	GetRepositoryPath(registryType, registryURL string) (string, error)

	// GetRulesetsPath returns the rulesets subdirectory path for extracted files
	GetRulesetsPath(registryType, registryURL string) (string, error)

	// GetCacheKey generates a SHA-256 hash key for a registry configuration
	GetCacheKey(registryType, registryURL string) (string, error)

	// EnsureCacheDir creates the cache directory structure if it doesn't exist
	EnsureCacheDir(registryType, registryURL string) error

	// UpdateCacheInfo updates cache metadata for a registry
	UpdateCacheInfo(registryType, registryURL, version string) error

	// NormalizeURL normalizes a registry URL for consistent hashing
	NormalizeURL(url string) string

	// IsCacheValid checks if cached content is still valid based on TTL
	IsCacheValid(registryType, registryURL string, ttl time.Duration) (bool, error)

	// GetCacheSize returns the total size of cache directory in bytes
	GetCacheSize() (int64, error)

	// CleanupExpired removes expired cache entries based on TTL
	CleanupExpired(ttl time.Duration) error

	// CleanupOversized removes oldest cache entries to stay under size limit
	CleanupOversized(maxSize int64) error

	// GetRulesetVersionPath returns the path for a specific ruleset version
	GetRulesetVersionPath(registryType, registryURL, rulesetName, version string) (string, error)

	// GetMetadataManager returns the metadata manager for this cache
	GetMetadataManager() *MetadataManager

	// GetRulesetStorage returns the ruleset storage manager for this cache
	GetRulesetStorage() *RulesetStorage

	// GetRulesetCacheKey generates a SHA-256 hash key for a ruleset with patterns
	GetRulesetCacheKey(rulesetName string, patterns []string) string

	// GetRulesetCachePath returns the cache directory path for a specific ruleset
	GetRulesetCachePath(registryType, registryURL, rulesetName string, patterns []string) (string, error)

	// GetRulesetMapper returns the ruleset mapper for this cache
	GetRulesetMapper() *RulesetMapper
}

// CacheInfo represents metadata stored in cache-info.json
type CacheInfo struct {
	CacheKey      string    `json:"cache_key"`
	RegistryType  string    `json:"registry_type"`
	RegistryURL   string    `json:"registry_url"`
	NormalizedURL string    `json:"normalized_url"`
	Version       string    `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	LastAccessed  time.Time `json:"last_accessed"`
	LastUpdated   time.Time `json:"last_updated"`
}

// DefaultManager implements the Manager interface
type DefaultManager struct {
	cacheRoot       string
	normalizer      *URLNormalizer
	mapper          *RegistryMapper
	rulesetMapper   *RulesetMapper
	metadataManager *MetadataManager
	rulesetStorage  *RulesetStorage
}

// NewManager creates a new cache manager with the specified cache root directory
func NewManager(cacheRoot string) *DefaultManager {
	manager := &DefaultManager{
		cacheRoot:     cacheRoot,
		normalizer:    NewURLNormalizer(),
		mapper:        NewRegistryMapper(cacheRoot),
		rulesetMapper: NewRulesetMapper(cacheRoot),
	}
	manager.metadataManager = NewMetadataManager(manager)
	manager.rulesetStorage = NewRulesetStorage(manager)
	return manager
}

// GetCachePath returns the cache directory path for a registry configuration
func (m *DefaultManager) GetCachePath(registryType, registryURL string) (string, error) {
	key, err := m.GetCacheKey(registryType, registryURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate cache key: %w", err)
	}

	return filepath.Join(m.cacheRoot, "registries", key), nil
}

// GetRepositoryPath returns the repository subdirectory path for git clones
func (m *DefaultManager) GetRepositoryPath(registryType, registryURL string) (string, error) {
	cachePath, err := m.GetCachePath(registryType, registryURL)
	if err != nil {
		return "", err
	}
	return filepath.Join(cachePath, "repository"), nil
}

// GetRulesetsPath returns the rulesets subdirectory path for extracted files
func (m *DefaultManager) GetRulesetsPath(registryType, registryURL string) (string, error) {
	cachePath, err := m.GetCachePath(registryType, registryURL)
	if err != nil {
		return "", err
	}
	return filepath.Join(cachePath, "rulesets"), nil
}

// GetCacheKey generates a SHA-256 hash key for a registry configuration
func (m *DefaultManager) GetCacheKey(registryType, registryURL string) (string, error) {
	normalized := m.normalizer.NormalizeURL(registryType, registryURL)
	// Include registry type to prevent conflicts between different access methods
	cacheInput := fmt.Sprintf("%s:%s", registryType, normalized)
	hash := sha256.Sum256([]byte(cacheInput))
	return fmt.Sprintf("%x", hash), nil
}

// EnsureCacheDir creates the cache directory structure if it doesn't exist
func (m *DefaultManager) EnsureCacheDir(registryType, registryURL string) error {
	cachePath, err := m.GetCachePath(registryType, registryURL)
	if err != nil {
		return fmt.Errorf("failed to get cache path: %w", err)
	}

	// Create main cache directory
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create repository subdirectory for git clones
	repositoryPath := filepath.Join(cachePath, "repository")
	if err := os.MkdirAll(repositoryPath, 0o755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Create rulesets subdirectory for extracted files
	rulesetsPath := filepath.Join(cachePath, "rulesets")
	if err := os.MkdirAll(rulesetsPath, 0o755); err != nil {
		return fmt.Errorf("failed to create rulesets directory: %w", err)
	}

	// Create temp directory at cache root level
	tempPath := filepath.Join(m.cacheRoot, "temp")
	if err := os.MkdirAll(tempPath, 0o755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	return nil
}

// UpdateCacheInfo updates cache metadata for a registry
func (m *DefaultManager) UpdateCacheInfo(registryType, registryURL, version string) error {
	cacheKey, err := m.GetCacheKey(registryType, registryURL)
	if err != nil {
		return fmt.Errorf("failed to generate cache key: %w", err)
	}

	normalizedURL := m.normalizer.NormalizeURL(registryType, registryURL)

	// Update registry mapping
	if err := m.mapper.AddMapping(cacheKey, registryType, registryURL, normalizedURL); err != nil {
		return fmt.Errorf("failed to update registry mapping: %w", err)
	}

	// Create cache-info.json in cache directory
	cachePath, err := m.GetCachePath(registryType, registryURL)
	if err != nil {
		return fmt.Errorf("failed to get cache path: %w", err)
	}

	cacheInfo := CacheInfo{
		CacheKey:      cacheKey,
		RegistryType:  registryType,
		RegistryURL:   registryURL,
		NormalizedURL: normalizedURL,
		Version:       version,
		CreatedAt:     time.Now(),
		LastAccessed:  time.Now(),
		LastUpdated:   time.Now(),
	}

	// Load existing cache info to preserve creation time
	if existingInfo, err := m.loadCacheInfo(cachePath); err == nil {
		cacheInfo.CreatedAt = existingInfo.CreatedAt
	}

	return m.saveCacheInfo(cachePath, &cacheInfo)
}

// loadCacheInfo loads cache-info.json from cache directory
func (m *DefaultManager) loadCacheInfo(cachePath string) (*CacheInfo, error) {
	infoPath := filepath.Join(cachePath, "cache-info.json")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, err
	}

	var info CacheInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// saveCacheInfo saves cache-info.json to cache directory
func (m *DefaultManager) saveCacheInfo(cachePath string, info *CacheInfo) error {
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	infoPath := filepath.Join(cachePath, "cache-info.json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache info: %w", err)
	}

	return os.WriteFile(infoPath, data, 0o644)
}

// NormalizeURL normalizes a registry URL for consistent hashing
// Deprecated: Use URLNormalizer.NormalizeURL directly for type-specific normalization
func (m *DefaultManager) NormalizeURL(url string) string {
	// Fallback to generic normalization for backward compatibility
	return m.normalizer.normalizeGenericURL(url)
}

// IsCacheValid checks if cached content is still valid based on TTL
func (m *DefaultManager) IsCacheValid(registryType, registryURL string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return true, nil // No expiration if TTL is 0 or negative
	}

	cachePath, err := m.GetCachePath(registryType, registryURL)
	if err != nil {
		return false, err
	}

	cacheInfo, err := m.loadCacheInfo(cachePath)
	if err != nil {
		return false, nil // Cache doesn't exist or is invalid
	}

	// Update last accessed time
	cacheKey, _ := m.GetCacheKey(registryType, registryURL)
	_ = m.mapper.UpdateLastAccessed(cacheKey)

	return time.Since(cacheInfo.LastUpdated) < ttl, nil
}

// GetCacheSize returns the total size of cache directory in bytes
func (m *DefaultManager) GetCacheSize() (int64, error) {
	cacheDir := filepath.Join(m.cacheRoot, "registries")
	var totalSize int64

	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible files
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

// CleanupExpired removes expired cache entries based on TTL
func (m *DefaultManager) CleanupExpired(ttl time.Duration) error {
	if ttl <= 0 {
		return nil // No cleanup if TTL is 0 or negative
	}

	mappings, err := m.mapper.ListMappings()
	if err != nil {
		return err
	}

	for _, mapping := range mappings {
		if time.Since(mapping.LastAccessed) > ttl {
			cachePath := filepath.Join(m.cacheRoot, "registries", mapping.CacheKey)
			_ = os.RemoveAll(cachePath)
			_ = m.mapper.RemoveMapping(mapping.CacheKey)
		}
	}

	return nil
}

// CleanupOversized removes oldest cache entries to stay under size limit
func (m *DefaultManager) CleanupOversized(maxSize int64) error {
	if maxSize <= 0 {
		return nil // No size limit
	}

	currentSize, err := m.GetCacheSize()
	if err != nil || currentSize <= maxSize {
		return err
	}

	mappings, err := m.mapper.ListMappings()
	if err != nil {
		return err
	}

	// Sort by last accessed time (oldest first)
	for i := 0; i < len(mappings)-1; i++ {
		for j := i + 1; j < len(mappings); j++ {
			if mappings[i].LastAccessed.After(mappings[j].LastAccessed) {
				mappings[i], mappings[j] = mappings[j], mappings[i]
			}
		}
	}

	// Remove oldest entries until under size limit
	for _, mapping := range mappings {
		if currentSize <= maxSize {
			break
		}
		cachePath := filepath.Join(m.cacheRoot, "registries", mapping.CacheKey)
		if _, err := os.Stat(cachePath); err == nil {
			currentSize -= getDirSize(cachePath)
			_ = os.RemoveAll(cachePath)
			_ = m.mapper.RemoveMapping(mapping.CacheKey)
		}
	}

	return nil
}

// getDirSize calculates the total size of a directory
func getDirSize(path string) int64 {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// GetRulesetVersionPath returns the path for a specific ruleset version (legacy method)
func (m *DefaultManager) GetRulesetVersionPath(registryType, registryURL, rulesetName, version string) (string, error) {
	// Use empty patterns for backward compatibility
	rulesetCachePath, err := m.GetRulesetCachePath(registryType, registryURL, rulesetName, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get ruleset cache path: %w", err)
	}
	return filepath.Join(rulesetCachePath, version), nil
}

// GetMetadataManager returns the metadata manager for this cache
func (m *DefaultManager) GetMetadataManager() *MetadataManager {
	return m.metadataManager
}

// GetRulesetStorage returns the ruleset storage manager for this cache
func (m *DefaultManager) GetRulesetStorage() *RulesetStorage {
	return m.rulesetStorage
}

// GetRulesetMapper returns the ruleset mapper for this cache
func (m *DefaultManager) GetRulesetMapper() *RulesetMapper {
	return m.rulesetMapper
}

// GetRulesetCacheKey generates a SHA-256 hash key for a ruleset with patterns
func (m *DefaultManager) GetRulesetCacheKey(rulesetName string, patterns []string) string {
	// Normalize patterns: sort and trim whitespace for consistency
	normalizedPatterns := make([]string, len(patterns))
	for i, pattern := range patterns {
		normalizedPatterns[i] = strings.TrimSpace(pattern)
	}
	sort.Strings(normalizedPatterns)

	// Create cache input with ruleset name and normalized patterns
	patternsStr := strings.Join(normalizedPatterns, ",")
	cacheInput := fmt.Sprintf("%s:%s", rulesetName, patternsStr)
	hash := sha256.Sum256([]byte(cacheInput))
	return fmt.Sprintf("%x", hash)
}

// GetRulesetCachePath returns the cache directory path for a specific ruleset
func (m *DefaultManager) GetRulesetCachePath(registryType, registryURL, rulesetName string, patterns []string) (string, error) {
	// Get registry-level cache path and key
	registryCachePath, err := m.GetCachePath(registryType, registryURL)
	if err != nil {
		return "", fmt.Errorf("failed to get registry cache path: %w", err)
	}

	registryCacheKey, err := m.GetCacheKey(registryType, registryURL)
	if err != nil {
		return "", fmt.Errorf("failed to get registry cache key: %w", err)
	}

	// Generate ruleset-level cache key
	rulesetKey := m.GetRulesetCacheKey(rulesetName, patterns)

	// Update ruleset mapper
	if err := m.rulesetMapper.AddMapping(rulesetKey, registryCacheKey, rulesetName, patterns); err != nil {
		// Log error but don't fail - mapping is for convenience
		_ = err
	}

	// Return three-level path: /cache/registries/{registry_hash}/rulesets/{ruleset_hash}/
	return filepath.Join(registryCachePath, "rulesets", rulesetKey), nil
}

// CreateRulesetCache creates the cache structure for a ruleset
func (m *DefaultManager) CreateRulesetCache(registryType, registryURL, rulesetName, version string, patterns []string) error {
	// Ensure registry cache directory exists
	if err := m.EnsureCacheDir(registryType, registryURL); err != nil {
		return fmt.Errorf("failed to ensure registry cache dir: %w", err)
	}

	// Get ruleset cache path
	rulesetCachePath, err := m.GetRulesetCachePath(registryType, registryURL, rulesetName, patterns)
	if err != nil {
		return fmt.Errorf("failed to get ruleset cache path: %w", err)
	}

	// Create version directory
	versionPath := filepath.Join(rulesetCachePath, version)
	if err := os.MkdirAll(versionPath, 0o755); err != nil {
		return fmt.Errorf("failed to create version directory: %w", err)
	}

	return nil
}

// LoadRulesetMap loads the ruleset mapping file
func (m *DefaultManager) LoadRulesetMap() (*RulesetMapFile, error) {
	return m.rulesetMapper.LoadMapFile()
}
