package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RegistryIndex represents the index file for a registry cache
type RegistryIndex struct {
	CreatedOn              string                   `json:"created_on"`
	LastUpdatedOn          string                   `json:"last_updated_on"`
	LastAccessedOn         string                   `json:"last_accessed_on"`
	NormalizedRegistryURL  string                   `json:"normalized_registry_url"`
	NormalizedRegistryType string                   `json:"normalized_registry_type"`
	Rulesets               map[string]*RulesetCache `json:"rulesets"`
}

// RulesetCache represents cached ruleset information
type RulesetCache struct {
	// Git registries use patterns, non-Git use ruleset name
	NormalizedRulesetPatterns []string                 `json:"normalized_ruleset_patterns,omitempty"`
	NormalizedRulesetName     string                   `json:"normalized_ruleset_name,omitempty"`
	CreatedOn                 string                   `json:"created_on"`
	LastUpdatedOn             string                   `json:"last_updated_on"`
	LastAccessedOn            string                   `json:"last_accessed_on"`
	Versions                  map[string]*VersionCache `json:"versions"`
}

// VersionCache represents cached version information
type VersionCache struct {
	CreatedOn      string `json:"created_on"`
	LastUpdatedOn  string `json:"last_updated_on"`
	LastAccessedOn string `json:"last_accessed_on"`
}

// LoadRegistryIndex loads the registry index file
func LoadRegistryIndex(registryPath string) (*RegistryIndex, error) {
	indexPath := filepath.Join(registryPath, "index.json")

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("registry index not found: %s", indexPath)
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry index: %w", err)
	}

	var index RegistryIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse registry index: %w", err)
	}

	return &index, nil
}

// SaveRegistryIndex saves the registry index file
func SaveRegistryIndex(registryPath string, index *RegistryIndex) error {
	if err := os.MkdirAll(registryPath, 0o755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	indexPath := filepath.Join(registryPath, "index.json")
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry index: %w", err)
	}

	return os.WriteFile(indexPath, data, 0o644)
}

// NewRegistryIndex creates a new registry index
func NewRegistryIndex(registryURL, registryType string) *RegistryIndex {
	now := time.Now().UTC().Format(time.RFC3339)
	return &RegistryIndex{
		CreatedOn:              now,
		LastUpdatedOn:          now,
		LastAccessedOn:         now,
		NormalizedRegistryURL:  registryURL,
		NormalizedRegistryType: registryType,
		Rulesets:               make(map[string]*RulesetCache),
	}
}

// NewRulesetCache creates a new ruleset cache entry
func NewRulesetCache() *RulesetCache {
	now := time.Now().UTC().Format(time.RFC3339)
	return &RulesetCache{
		CreatedOn:      now,
		LastUpdatedOn:  now,
		LastAccessedOn: now,
		Versions:       make(map[string]*VersionCache),
	}
}

// NewVersionCache creates a new version cache entry
func NewVersionCache() *VersionCache {
	now := time.Now().UTC().Format(time.RFC3339)
	return &VersionCache{
		CreatedOn:      now,
		LastUpdatedOn:  now,
		LastAccessedOn: now,
	}
}

// UpdateAccessTime updates the last accessed time for registry index
func (r *RegistryIndex) UpdateAccessTime() {
	r.LastAccessedOn = time.Now().UTC().Format(time.RFC3339)
}

// UpdateAccessTime updates the last accessed time for ruleset cache
func (r *RulesetCache) UpdateAccessTime() {
	r.LastAccessedOn = time.Now().UTC().Format(time.RFC3339)
}

// UpdateAccessTime updates the last accessed time for version cache
func (v *VersionCache) UpdateAccessTime() {
	v.LastAccessedOn = time.Now().UTC().Format(time.RFC3339)
}

// CleanupCache performs TTL-based and size-based cleanup across all cache managers
func CleanupCache(cacheRoot string, ttl time.Duration, maxSize int64) error {
	// Cleanup Git registries
	gitManager := NewGitCacheManager(cacheRoot)
	if err := gitManager.Cleanup(ttl, maxSize); err != nil {
		return fmt.Errorf("failed to cleanup Git cache: %w", err)
	}

	// Cleanup non-Git registries
	rulesetManager := NewRulesetCacheManager(cacheRoot)
	if err := rulesetManager.Cleanup(ttl, maxSize); err != nil {
		return fmt.Errorf("failed to cleanup ruleset cache: %w", err)
	}

	return nil
}

// GetCacheStats returns statistics about cache usage
func GetCacheStats(cacheRoot string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total cache size
	var totalSize int64
	err := filepath.Walk(cacheRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cache size: %w", err)
	}

	stats["total_size_bytes"] = totalSize
	stats["total_size_mb"] = totalSize / (1024 * 1024)

	// Count registries
	registriesPath := filepath.Join(cacheRoot, "registries")
	if entries, err := os.ReadDir(registriesPath); err == nil {
		stats["registry_count"] = len(entries)
	} else {
		stats["registry_count"] = 0
	}

	return stats, nil
}
