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
	registriesPath := filepath.Join(r.cacheRoot, "registries")

	// Check if registries directory exists
	if _, err := os.Stat(registriesPath); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(registriesPath)
	if err != nil {
		return fmt.Errorf("failed to read registries directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			registryPath := filepath.Join(registriesPath, entry.Name())
			if err := r.cleanupRegistry(registryPath, ttl); err != nil {
				continue // Log error but continue with other registries
			}
		}
	}

	// Check cache size and cleanup if needed
	if maxSize > 0 {
		return r.cleanupBySize(maxSize)
	}

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
	_ = SaveRegistryIndex(registryPath, index) // Ignore error for access time update
}

// cleanupRegistry removes expired entries from a single registry
func (r *RulesetCacheManager) cleanupRegistry(registryPath string, ttl time.Duration) error {
	index, err := LoadRegistryIndex(registryPath)
	if err != nil {
		return err
	}

	changed := false
	for rulesetKey, rulesetCache := range index.Rulesets {
		lastAccessed, err := time.Parse(time.RFC3339, rulesetCache.LastAccessedOn)
		if err != nil || time.Since(lastAccessed) > ttl {
			// Remove entire ruleset if expired
			rulesetPath := filepath.Join(registryPath, "rulesets", rulesetKey)
			_ = os.RemoveAll(rulesetPath) // Ignore cleanup errors
			delete(index.Rulesets, rulesetKey)
			changed = true
			continue
		}

		// Check individual versions
		for version, versionCache := range rulesetCache.Versions {
			lastAccessed, err := time.Parse(time.RFC3339, versionCache.LastAccessedOn)
			if err != nil || time.Since(lastAccessed) > ttl {
				versionPath := filepath.Join(registryPath, "rulesets", rulesetKey, version)
				_ = os.RemoveAll(versionPath) // Ignore cleanup errors
				delete(rulesetCache.Versions, version)
				changed = true
			}
		}

		// Remove ruleset if no versions left
		if len(rulesetCache.Versions) == 0 {
			rulesetPath := filepath.Join(registryPath, "rulesets", rulesetKey)
			_ = os.RemoveAll(rulesetPath) // Ignore cleanup errors
			delete(index.Rulesets, rulesetKey)
			changed = true
		}
	}

	if changed {
		return SaveRegistryIndex(registryPath, index)
	}
	return nil
}

// cleanupBySize removes oldest entries until cache is under size limit
func (r *RulesetCacheManager) cleanupBySize(maxSize int64) error {
	currentSize, err := r.getCacheSize()
	if err != nil {
		return err
	}

	if currentSize <= maxSize {
		return nil
	}

	// Get all version entries with access times
	type versionEntry struct {
		path       string
		accessTime time.Time
		size       int64
	}

	var entries []versionEntry
	registriesPath := filepath.Join(r.cacheRoot, "registries")

	err = filepath.Walk(registriesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return err
		}

		// Check if this is a version directory (has files)
		hasFiles := false
		_ = filepath.Walk(path, func(subPath string, subInfo os.FileInfo, subErr error) error {
			if subErr == nil && !subInfo.IsDir() {
				hasFiles = true
			}
			return nil
		})

		if hasFiles {
			size := r.getDirSize(path)
			accessTime := info.ModTime() // Use modification time as fallback
			entries = append(entries, versionEntry{
				path:       path,
				accessTime: accessTime,
				size:       size,
			})
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Sort by access time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].accessTime.After(entries[j].accessTime) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove oldest entries until under size limit
	for _, entry := range entries {
		if currentSize <= maxSize {
			break
		}
		_ = os.RemoveAll(entry.path) // Ignore cleanup errors
		currentSize -= entry.size
	}

	return nil
}

// getCacheSize returns the total size of the cache in bytes
func (r *RulesetCacheManager) getCacheSize() (int64, error) {
	var size int64
	err := filepath.Walk(r.cacheRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// getDirSize returns the size of a directory in bytes
func (r *RulesetCacheManager) getDirSize(dirPath string) int64 {
	var size int64
	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}
