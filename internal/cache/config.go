package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheConfig represents the global cache configuration
type CacheConfig struct {
	Version        string `json:"version"`
	CreatedOn      string `json:"created_on"`
	LastUpdatedOn  string `json:"last_updated_on"`
	TTLHours       int    `json:"ttl_hours"`
	MaxSizeMB      int    `json:"max_size_mb"`
	CleanupEnabled bool   `json:"cleanup_enabled"`
}

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

// LoadCacheConfig loads the global cache configuration
func LoadCacheConfig(cacheRoot string) (*CacheConfig, error) {
	configPath := filepath.Join(cacheRoot, "config.json")

	// Create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := &CacheConfig{
			Version:        "1.0",
			CreatedOn:      time.Now().UTC().Format(time.RFC3339),
			LastUpdatedOn:  time.Now().UTC().Format(time.RFC3339),
			TTLHours:       24,
			MaxSizeMB:      1024,
			CleanupEnabled: true,
		}
		if err := SaveCacheConfig(cacheRoot, defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config CacheConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveCacheConfig saves the global cache configuration
func SaveCacheConfig(cacheRoot string, config *CacheConfig) error {
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	configPath := filepath.Join(cacheRoot, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0o644)
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
