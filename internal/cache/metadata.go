package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// VersionsFile represents the structure of versions.json
type VersionsFile struct {
	CachedAt   time.Time                    `json:"cached_at"`
	TTLSeconds int                          `json:"ttl_seconds"`
	Rulesets   map[string][]string          `json:"rulesets"`
	Mappings   map[string]map[string]string `json:"mappings,omitempty"` // Git only: version -> commit mappings
}

// MetadataFile represents the structure of metadata.json
type MetadataFile struct {
	CachedAt   time.Time                  `json:"cached_at"`
	TTLSeconds int                        `json:"ttl_seconds"`
	Rulesets   map[string]RulesetMetadata `json:"rulesets"`
}

// RulesetMetadata contains metadata for a single ruleset
type RulesetMetadata struct {
	LatestVersion  string `json:"latest_version"`
	FileCount      int    `json:"file_count"`
	TotalSizeBytes int64  `json:"total_size_bytes"`
}

// MetadataManager handles versions.json and metadata.json files
type MetadataManager struct {
	cacheManager Manager
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(cacheManager Manager) *MetadataManager {
	return &MetadataManager{
		cacheManager: cacheManager,
	}
}

// UpdateVersions updates the versions.json file for a registry
func (mm *MetadataManager) UpdateVersions(registryType, registryURL, rulesetName string, versions []string, mappings map[string]string) error {
	cachePath, err := mm.cacheManager.GetCachePath(registryType, registryURL)
	if err != nil {
		return fmt.Errorf("failed to get cache path: %w", err)
	}

	versionsPath := filepath.Join(cachePath, "versions.json")

	// Load existing versions file or create new one
	versionsFile, err := mm.loadVersionsFile(versionsPath)
	if err != nil {
		versionsFile = &VersionsFile{
			CachedAt:   time.Now(),
			TTLSeconds: 3600,
			Rulesets:   make(map[string][]string),
			Mappings:   make(map[string]map[string]string),
		}
	}

	// Update versions for this ruleset
	versionsFile.Rulesets[rulesetName] = versions
	versionsFile.CachedAt = time.Now()

	// Update mappings for git registries
	if registryType == "git" && mappings != nil {
		if versionsFile.Mappings == nil {
			versionsFile.Mappings = make(map[string]map[string]string)
		}
		versionsFile.Mappings[rulesetName] = mappings
	}

	return mm.saveVersionsFile(versionsPath, versionsFile)
}

// UpdateMetadata updates the metadata.json file for a registry
func (mm *MetadataManager) UpdateMetadata(registryType, registryURL, rulesetName, latestVersion string, fileCount int, totalSize int64) error {
	cachePath, err := mm.cacheManager.GetCachePath(registryType, registryURL)
	if err != nil {
		return fmt.Errorf("failed to get cache path: %w", err)
	}

	metadataPath := filepath.Join(cachePath, "metadata.json")

	// Load existing metadata file or create new one
	metadataFile, err := mm.loadMetadataFile(metadataPath)
	if err != nil {
		metadataFile = &MetadataFile{
			CachedAt:   time.Now(),
			TTLSeconds: 3600,
			Rulesets:   make(map[string]RulesetMetadata),
		}
	}

	// Update metadata for this ruleset
	metadataFile.Rulesets[rulesetName] = RulesetMetadata{
		LatestVersion:  latestVersion,
		FileCount:      fileCount,
		TotalSizeBytes: totalSize,
	}
	metadataFile.CachedAt = time.Now()

	return mm.saveMetadataFile(metadataPath, metadataFile)
}

// GetVersions retrieves versions for a ruleset from cache
func (mm *MetadataManager) GetVersions(registryType, registryURL, rulesetName string) (versions []string, mappings map[string]string, err error) {
	cachePath, err := mm.cacheManager.GetCachePath(registryType, registryURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get cache path: %w", err)
	}

	versionsPath := filepath.Join(cachePath, "versions.json")
	versionsFile, err := mm.loadVersionsFile(versionsPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load versions file: %w", err)
	}

	versions, exists := versionsFile.Rulesets[rulesetName]
	if !exists {
		return nil, nil, fmt.Errorf("ruleset %s not found in versions cache", rulesetName)
	}

	var rulesetMappings map[string]string
	if versionsFile.Mappings != nil {
		rulesetMappings = versionsFile.Mappings[rulesetName]
	}

	return versions, rulesetMappings, nil
}

// GetMetadata retrieves metadata for a ruleset from cache
func (mm *MetadataManager) GetMetadata(registryType, registryURL, rulesetName string) (*RulesetMetadata, error) {
	cachePath, err := mm.cacheManager.GetCachePath(registryType, registryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache path: %w", err)
	}

	metadataPath := filepath.Join(cachePath, "metadata.json")
	metadataFile, err := mm.loadMetadataFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata file: %w", err)
	}

	metadata, exists := metadataFile.Rulesets[rulesetName]
	if !exists {
		return nil, fmt.Errorf("ruleset %s not found in metadata cache", rulesetName)
	}

	return &metadata, nil
}

// IsVersionsCacheValid checks if the versions cache is still valid
func (mm *MetadataManager) IsVersionsCacheValid(registryType, registryURL string, ttl time.Duration) (bool, error) {
	cachePath, err := mm.cacheManager.GetCachePath(registryType, registryURL)
	if err != nil {
		return false, fmt.Errorf("failed to get cache path: %w", err)
	}

	versionsPath := filepath.Join(cachePath, "versions.json")
	versionsFile, err := mm.loadVersionsFile(versionsPath)
	if err != nil {
		return false, nil // Cache doesn't exist or is invalid
	}

	return time.Since(versionsFile.CachedAt) < ttl, nil
}

// loadVersionsFile loads versions.json from the specified path
func (mm *MetadataManager) loadVersionsFile(path string) (*VersionsFile, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("versions file does not exist")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions file: %w", err)
	}

	var versionsFile VersionsFile
	if err := json.Unmarshal(data, &versionsFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal versions file: %w", err)
	}

	return &versionsFile, nil
}

// saveVersionsFile saves versions.json to the specified path
func (mm *MetadataManager) saveVersionsFile(path string, versionsFile *VersionsFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(versionsFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal versions file: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// loadMetadataFile loads metadata.json from the specified path
func (mm *MetadataManager) loadMetadataFile(path string) (*MetadataFile, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("metadata file does not exist")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadataFile MetadataFile
	if err := json.Unmarshal(data, &metadataFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata file: %w", err)
	}

	return &metadataFile, nil
}

// saveMetadataFile saves metadata.json to the specified path
func (mm *MetadataManager) saveMetadataFile(path string, metadataFile *MetadataFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(metadataFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata file: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}
