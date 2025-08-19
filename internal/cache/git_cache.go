package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GitCacheManager implements GitRegistryCacheManager for Git-based registries
type GitCacheManager struct {
	cacheRoot string
}

// NewGitCacheManager creates a new Git registry cache manager
func NewGitCacheManager(cacheRoot string) *GitCacheManager {
	return &GitCacheManager{
		cacheRoot: cacheRoot,
	}
}

// Store stores files for a Git registry with given identifier and version
func (g *GitCacheManager) Store(registryURL string, identifier []string, version string, files map[string][]byte) error {
	return g.StoreRuleset(registryURL, identifier, version, files)
}

// Get retrieves files for a Git registry with given identifier and version
func (g *GitCacheManager) Get(registryURL string, identifier []string, version string) (map[string][]byte, error) {
	return g.GetRuleset(registryURL, identifier, version)
}

// GetPath returns the filesystem path for a Git registry with given identifier and version
func (g *GitCacheManager) GetPath(registryURL string, identifier []string, version string) (string, error) {
	registryKey := GenerateRegistryKey("git", registryURL)
	patternsKey := GeneratePatternsKey(identifier)

	return filepath.Join(g.cacheRoot, "registries", registryKey, "rulesets", patternsKey, version), nil
}

// IsValid checks if cached content is still valid based on TTL
func (g *GitCacheManager) IsValid(registryURL string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return true, nil
	}

	registryKey := GenerateRegistryKey("git", registryURL)
	registryPath := filepath.Join(g.cacheRoot, "registries", registryKey)

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
func (g *GitCacheManager) Cleanup(ttl time.Duration, maxSize int64) error {
	registriesPath := filepath.Join(g.cacheRoot, "registries")

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
			if err := g.cleanupRegistry(registryPath, ttl); err != nil {
				continue // Log error but continue with other registries
			}
		}
	}

	// Check cache size and cleanup if needed
	if maxSize > 0 {
		return g.cleanupBySize(maxSize)
	}

	return nil
}

// StoreRuleset stores ruleset files for a Git registry with patterns and commit hash
func (g *GitCacheManager) StoreRuleset(registryURL string, patterns []string, commitHash string, files map[string][]byte) error {
	registryKey := GenerateRegistryKey("git", registryURL)
	patternsKey := GeneratePatternsKey(patterns)

	registryPath := filepath.Join(g.cacheRoot, "registries", registryKey)
	versionPath := filepath.Join(registryPath, "rulesets", patternsKey, commitHash)

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
	return g.updateRegistryIndex(registryURL, patterns, commitHash)
}

// GetRuleset retrieves ruleset files for a Git registry with patterns and commit hash
func (g *GitCacheManager) GetRuleset(registryURL string, patterns []string, commitHash string) (map[string][]byte, error) {
	versionPath, err := g.GetPath(registryURL, patterns, commitHash)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cached version not found: %s", commitHash)
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
	g.updateAccessTime(registryURL, patterns, commitHash)

	return files, nil
}

// GetRepositoryPath returns the path to the Git repository clone
func (g *GitCacheManager) GetRepositoryPath(registryURL string) (string, error) {
	registryKey := GenerateRegistryKey("git", registryURL)
	return filepath.Join(g.cacheRoot, "registries", registryKey, "repository"), nil
}

// updateRegistryIndex updates the registry index with new ruleset/version information
func (g *GitCacheManager) updateRegistryIndex(registryURL string, patterns []string, commitHash string) error {
	registryKey := GenerateRegistryKey("git", registryURL)
	registryPath := filepath.Join(g.cacheRoot, "registries", registryKey)

	// Load or create registry index
	index, err := LoadRegistryIndex(registryPath)
	if err != nil {
		index = NewRegistryIndex(normalizeURL(registryURL), "git")
	}

	patternsKey := GeneratePatternsKey(patterns)

	// Get or create ruleset cache
	rulesetCache, exists := index.Rulesets[patternsKey]
	if !exists {
		rulesetCache = NewRulesetCache()
		rulesetCache.NormalizedRulesetPatterns = patterns
		index.Rulesets[patternsKey] = rulesetCache
	}

	// Add version cache
	rulesetCache.Versions[commitHash] = NewVersionCache()
	rulesetCache.UpdateAccessTime()
	index.UpdateAccessTime()

	return SaveRegistryIndex(registryPath, index)
}

// updateAccessTime updates access times for registry, ruleset, and version
func (g *GitCacheManager) updateAccessTime(registryURL string, patterns []string, commitHash string) {
	registryKey := GenerateRegistryKey("git", registryURL)
	registryPath := filepath.Join(g.cacheRoot, "registries", registryKey)

	index, err := LoadRegistryIndex(registryPath)
	if err != nil {
		return
	}

	patternsKey := GeneratePatternsKey(patterns)

	if rulesetCache, exists := index.Rulesets[patternsKey]; exists {
		if versionCache, exists := rulesetCache.Versions[commitHash]; exists {
			versionCache.UpdateAccessTime()
		}
		rulesetCache.UpdateAccessTime()
	}

	index.UpdateAccessTime()
	_ = SaveRegistryIndex(registryPath, index) // Ignore error for access time update
}

// cleanupRegistry removes expired entries from a single registry
func (g *GitCacheManager) cleanupRegistry(registryPath string, ttl time.Duration) error {
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
func (g *GitCacheManager) cleanupBySize(maxSize int64) error {
	currentSize, err := g.getCacheSize()
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
	registriesPath := filepath.Join(g.cacheRoot, "registries")

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
			size := g.getDirSize(path)
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
func (g *GitCacheManager) getCacheSize() (int64, error) {
	var size int64
	err := filepath.Walk(g.cacheRoot, func(path string, info os.FileInfo, err error) error {
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
func (g *GitCacheManager) getDirSize(dirPath string) int64 {
	var size int64
	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}
