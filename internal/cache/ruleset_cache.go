package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RulesetCacheManager implements RulesetRegistryCacheManager for non-Git registries
type RulesetCacheManager struct {
	cacheRoot string
}

// NewRulesetCacheManager creates a new non-Git registry cache manager
func NewRulesetCacheManager(cacheRoot string) *RulesetCacheManager {
	return &RulesetCacheManager{
		cacheRoot: cacheRoot,
	}
}

// Store stores files for a non-Git registry with given identifier and version
func (r *RulesetCacheManager) Store(registryURL string, identifier []string, version string, files map[string][]byte) error {
	if len(identifier) == 0 {
		return fmt.Errorf("ruleset name is required for non-Git registries")
	}
	return r.StoreRuleset(registryURL, identifier[0], version, files)
}

// Get retrieves files for a non-Git registry with given identifier and version
func (r *RulesetCacheManager) Get(registryURL string, identifier []string, version string) (map[string][]byte, error) {
	if len(identifier) == 0 {
		return nil, fmt.Errorf("ruleset name is required for non-Git registries")
	}
	return r.GetRuleset(registryURL, identifier[0], version)
}

// GetPath returns the filesystem path for a non-Git registry with given identifier and version
func (r *RulesetCacheManager) GetPath(registryURL string, identifier []string, version string) (string, error) {
	if len(identifier) == 0 {
		return "", fmt.Errorf("ruleset name is required for non-Git registries")
	}
	
	registryKey := GenerateRegistryKey("ruleset", registryURL)
	rulesetKey := GenerateRulesetKey(identifier[0])
	
	return filepath.Join(r.cacheRoot, "registries", registryKey, "rulesets", rulesetKey, version), nil
}

// IsValid checks if cached content is still valid based on TTL
func (r *RulesetCacheManager) IsValid(registryURL string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return true, nil
	}

	registryKey := GenerateRegistryKey("ruleset", registryURL)
	registryPath := filepath.Join(r.cacheRoot, "registries", registryKey)
	
	index, err := LoadRegistryIndex(registryPath)
	if err != nil {
		return false, nil
	}

	lastAccessed, err := time.Parse(time.RFC3339, index.LastAccessedOn)
	if err != nil {
		return false, nil
	}

	return time.Since(lastAccessed) < ttl, nil
}

// Cleanup removes expired cache entries based on TTL and size limits
func (r *RulesetCacheManager) Cleanup(ttl time.Duration, maxSize int64) error {
	// Implementation will be added in task 4.2
	return nil
}

// StoreRuleset stores ruleset files for a non-Git registry with ruleset name and version
func (r *RulesetCacheManager) StoreRuleset(registryURL, rulesetName, version string, files map[string][]byte) error {
	registryKey := GenerateRegistryKey("ruleset", registryURL)
	rulesetKey := GenerateRulesetKey(rulesetName)
	
	registryPath := filepath.Join(r.cacheRoot, "registries", registryKey)
	versionPath := filepath.Join(registryPath, "rulesets", rulesetKey, version)
	
	// Create directory structure
	if err := os.MkdirAll(versionPath, 0o755); err != nil {
		return fmt.Errorf("failed to create version directory: %w", err)
	}

	// Store files
	for filename, content := range files {
		filePath := filepath.Join(versionPath, filename)
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return fmt.Errorf("failed to create file directory: %w", err)
		}
		if err := os.WriteFile(filePath, content, 0o644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filename, err)
		}
	}

	// Update registry index
	return r.updateRegistryIndex(registryURL, rulesetName, version)
}

// GetRuleset retrieves ruleset files for a non-Git registry with ruleset name and version
func (r *RulesetCacheManager) GetRuleset(registryURL, rulesetName, version string) (map[string][]byte, error) {
	versionPath, err := r.GetPath(registryURL, []string{rulesetName}, version)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cached version not found: %s", version)
	}

	files := make(map[string][]byte)
	err = filepath.Walk(versionPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		relPath, err := filepath.Rel(versionPath, path)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		files[relPath] = content
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to read cached files: %w", err)
	}

	// Update access time
	r.updateAccessTime(registryURL, rulesetName, version)
	
	return files, nil
}

// updateRegistryIndex updates the registry index with new ruleset/version information
func (r *RulesetCacheManager) updateRegistryIndex(registryURL, rulesetName, version string) error {
	registryKey := GenerateRegistryKey("ruleset", registryURL)
	registryPath := filepath.Join(r.cacheRoot, "registries", registryKey)
	
	// Load or create registry index
	index, err := LoadRegistryIndex(registryPath)
	if err != nil {
		index = NewRegistryIndex(normalizeURL(registryURL), "ruleset")
	}

	rulesetKey := GenerateRulesetKey(rulesetName)
	
	// Get or create ruleset cache
	rulesetCache, exists := index.Rulesets[rulesetKey]
	if !exists {
		rulesetCache = NewRulesetCache()
		rulesetCache.NormalizedRulesetName = rulesetName
		index.Rulesets[rulesetKey] = rulesetCache
	}

	// Add version cache
	rulesetCache.Versions[version] = NewVersionCache()
	rulesetCache.UpdateAccessTime()
	index.UpdateAccessTime()

	return SaveRegistryIndex(registryPath, index)
}

// updateAccessTime updates access times for registry, ruleset, and version
func (r *RulesetCacheManager) updateAccessTime(registryURL, rulesetName, version string) {
	registryKey := GenerateRegistryKey("ruleset", registryURL)
	registryPath := filepath.Join(r.cacheRoot, "registries", registryKey)
	
	index, err := LoadRegistryIndex(registryPath)
	if err != nil {
		return
	}

	rulesetKey := GenerateRulesetKey(rulesetName)
	
	if rulesetCache, exists := index.Rulesets[rulesetKey]; exists {
		if versionCache, exists := rulesetCache.Versions[version]; exists {
			versionCache.UpdateAccessTime()
		}
		rulesetCache.UpdateAccessTime()
	}
	
	index.UpdateAccessTime()
	SaveRegistryIndex(registryPath, index)
}