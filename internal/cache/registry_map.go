package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RegistryMapping represents the mapping between cache keys and registry information
type RegistryMapping struct {
	CacheKey      string    `json:"cache_key"`
	RegistryType  string    `json:"registry_type"`
	RegistryURL   string    `json:"registry_url"`
	NormalizedURL string    `json:"normalized_url"`
	CreatedAt     time.Time `json:"created_at"`
	LastAccessed  time.Time `json:"last_accessed"`
}

// RegistryMapFile represents the structure of the registry mapping file
type RegistryMapFile struct {
	Version  string            `json:"version"`
	Mappings []RegistryMapping `json:"mappings"`
}

// RegistryMapper manages the registry mapping file operations
type RegistryMapper struct {
	mapFilePath string
	mutex       sync.RWMutex
}

// NewRegistryMapper creates a new registry mapper
func NewRegistryMapper(cacheRoot string) *RegistryMapper {
	return &RegistryMapper{
		mapFilePath: filepath.Join(cacheRoot, "registry-map.json"),
	}
}

// AddMapping adds or updates a registry mapping
func (rm *RegistryMapper) AddMapping(cacheKey, registryType, registryURL, normalizedURL string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	mapFile, err := rm.loadMapFile()
	if err != nil {
		return fmt.Errorf("failed to load map file: %w", err)
	}

	now := time.Now()
	mapping := RegistryMapping{
		CacheKey:      cacheKey,
		RegistryType:  registryType,
		RegistryURL:   registryURL,
		NormalizedURL: normalizedURL,
		CreatedAt:     now,
		LastAccessed:  now,
	}

	// Update existing mapping or add new one
	found := false
	for i, existing := range mapFile.Mappings {
		if existing.CacheKey == cacheKey {
			mapping.CreatedAt = existing.CreatedAt // Preserve original creation time
			mapFile.Mappings[i] = mapping
			found = true
			break
		}
	}

	if !found {
		mapFile.Mappings = append(mapFile.Mappings, mapping)
	}

	return rm.saveMapFile(mapFile)
}

// GetMapping retrieves a registry mapping by cache key
func (rm *RegistryMapper) GetMapping(cacheKey string) (*RegistryMapping, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	mapFile, err := rm.loadMapFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load map file: %w", err)
	}

	for _, mapping := range mapFile.Mappings {
		if mapping.CacheKey == cacheKey {
			return &mapping, nil
		}
	}

	return nil, fmt.Errorf("mapping not found for cache key: %s", cacheKey)
}

// FindMappingByURL finds a registry mapping by normalized URL and type
func (rm *RegistryMapper) FindMappingByURL(registryType, normalizedURL string) (*RegistryMapping, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	mapFile, err := rm.loadMapFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load map file: %w", err)
	}

	for _, mapping := range mapFile.Mappings {
		if mapping.RegistryType == registryType && mapping.NormalizedURL == normalizedURL {
			return &mapping, nil
		}
	}

	return nil, fmt.Errorf("mapping not found for %s:%s", registryType, normalizedURL)
}

// UpdateLastAccessed updates the last accessed time for a cache key
func (rm *RegistryMapper) UpdateLastAccessed(cacheKey string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	mapFile, err := rm.loadMapFile()
	if err != nil {
		return fmt.Errorf("failed to load map file: %w", err)
	}

	for i, mapping := range mapFile.Mappings {
		if mapping.CacheKey == cacheKey {
			mapFile.Mappings[i].LastAccessed = time.Now()
			return rm.saveMapFile(mapFile)
		}
	}

	return fmt.Errorf("mapping not found for cache key: %s", cacheKey)
}

// RemoveMapping removes a registry mapping by cache key
func (rm *RegistryMapper) RemoveMapping(cacheKey string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	mapFile, err := rm.loadMapFile()
	if err != nil {
		return fmt.Errorf("failed to load map file: %w", err)
	}

	for i, mapping := range mapFile.Mappings {
		if mapping.CacheKey == cacheKey {
			mapFile.Mappings = append(mapFile.Mappings[:i], mapFile.Mappings[i+1:]...)
			return rm.saveMapFile(mapFile)
		}
	}

	return fmt.Errorf("mapping not found for cache key: %s", cacheKey)
}

// ListMappings returns all registry mappings
func (rm *RegistryMapper) ListMappings() ([]RegistryMapping, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	mapFile, err := rm.loadMapFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load map file: %w", err)
	}

	return mapFile.Mappings, nil
}

// ValidateAndRecover validates the mapping file and recovers from corruption
func (rm *RegistryMapper) ValidateAndRecover() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	mapFile, err := rm.loadMapFile()
	if err != nil {
		// If file is corrupted, create backup and start fresh
		if rm.backupCorruptedFile() {
			mapFile = &RegistryMapFile{
				Version:  "1.0",
				Mappings: []RegistryMapping{},
			}
			return rm.saveMapFile(mapFile)
		}
		return fmt.Errorf("failed to recover from corruption: %w", err)
	}

	// Validate mappings and remove invalid ones
	validMappings := make([]RegistryMapping, 0, len(mapFile.Mappings))
	for _, mapping := range mapFile.Mappings {
		if rm.isValidMapping(&mapping) {
			validMappings = append(validMappings, mapping)
		}
	}

	// Save cleaned mappings if any were removed
	if len(validMappings) != len(mapFile.Mappings) {
		mapFile.Mappings = validMappings
		return rm.saveMapFile(mapFile)
	}

	return nil
}

// backupCorruptedFile creates a backup of corrupted file
func (rm *RegistryMapper) backupCorruptedFile() bool {
	if _, err := os.Stat(rm.mapFilePath); os.IsNotExist(err) {
		return true // No file to backup
	}

	backupPath := rm.mapFilePath + ".corrupted." + fmt.Sprintf("%d", time.Now().Unix())
	err := os.Rename(rm.mapFilePath, backupPath)
	return err == nil
}

// isValidMapping validates a registry mapping
func (rm *RegistryMapper) isValidMapping(mapping *RegistryMapping) bool {
	// Check required fields
	if mapping.CacheKey == "" || mapping.RegistryType == "" || mapping.RegistryURL == "" {
		return false
	}

	// Validate cache key format (should be 64-character hex)
	if len(mapping.CacheKey) != 64 {
		return false
	}
	for _, c := range mapping.CacheKey {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}

	// Validate registry type
	validTypes := map[string]bool{
		"git": true, "gitlab": true, "s3": true, "https": true, "local": true,
	}
	return validTypes[mapping.RegistryType]
}

// loadMapFile loads the registry mapping file
func (rm *RegistryMapper) loadMapFile() (*RegistryMapFile, error) {
	if _, err := os.Stat(rm.mapFilePath); os.IsNotExist(err) {
		// Create empty map file if it doesn't exist
		return &RegistryMapFile{
			Version:  "1.0",
			Mappings: []RegistryMapping{},
		}, nil
	}

	data, err := os.ReadFile(rm.mapFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read map file: %w", err)
	}

	var mapFile RegistryMapFile
	if err := json.Unmarshal(data, &mapFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal map file: %w", err)
	}

	return &mapFile, nil
}

// saveMapFile saves the registry mapping file atomically
func (rm *RegistryMapper) saveMapFile(mapFile *RegistryMapFile) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(rm.mapFilePath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(mapFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal map file: %w", err)
	}

	// Write to temporary file first for atomic operation
	tempFile := rm.mapFilePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, rm.mapFilePath); err != nil {
		_ = os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
