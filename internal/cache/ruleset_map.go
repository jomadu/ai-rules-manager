package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// RulesetMapping represents the mapping between cache keys and ruleset information
type RulesetMapping struct {
	CacheKey           string    `json:"cache_key"`
	RegistryCacheKey   string    `json:"registry_cache_key"`
	RulesetName        string    `json:"ruleset_name"`
	Patterns           []string  `json:"patterns"`
	NormalizedPatterns string    `json:"normalized_patterns"`
	CreatedAt          time.Time `json:"created_at"`
	LastAccessed       time.Time `json:"last_accessed"`
}

// RulesetMapFile represents the structure of the ruleset mapping file
type RulesetMapFile struct {
	Version  string           `json:"version"`
	Mappings []RulesetMapping `json:"mappings"`
}

// RulesetMapper manages the ruleset mapping file operations
type RulesetMapper struct {
	mapFilePath string
	mutex       sync.RWMutex
}

// NewRulesetMapper creates a new ruleset mapper
func NewRulesetMapper(cacheRoot string) *RulesetMapper {
	return &RulesetMapper{
		mapFilePath: filepath.Join(cacheRoot, "ruleset-map.json"),
	}
}

// AddMapping adds or updates a ruleset mapping
func (rm *RulesetMapper) AddMapping(cacheKey, registryCacheKey, rulesetName string, patterns []string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	mapFile, err := rm.LoadMapFile()
	if err != nil {
		return fmt.Errorf("failed to load map file: %w", err)
	}

	normalizedPatterns := rm.normalizePatterns(patterns)
	now := time.Now()
	mapping := RulesetMapping{
		CacheKey:           cacheKey,
		RegistryCacheKey:   registryCacheKey,
		RulesetName:        rulesetName,
		Patterns:           patterns,
		NormalizedPatterns: normalizedPatterns,
		CreatedAt:          now,
		LastAccessed:       now,
	}

	// Update existing mapping or add new one
	found := false
	for i := range mapFile.Mappings {
		existing := &mapFile.Mappings[i]
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

// GetMapping retrieves a ruleset mapping by cache key
func (rm *RulesetMapper) GetMapping(cacheKey string) (*RulesetMapping, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	mapFile, err := rm.LoadMapFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load map file: %w", err)
	}

	for i := range mapFile.Mappings {
		if mapFile.Mappings[i].CacheKey == cacheKey {
			return &mapFile.Mappings[i], nil
		}
	}

	return nil, fmt.Errorf("mapping not found for cache key: %s", cacheKey)
}

// FindMappingByRuleset finds a ruleset mapping by registry cache key, ruleset name, and patterns
func (rm *RulesetMapper) FindMappingByRuleset(registryCacheKey, rulesetName string, patterns []string) (*RulesetMapping, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	mapFile, err := rm.LoadMapFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load map file: %w", err)
	}

	normalizedPatterns := rm.normalizePatterns(patterns)
	for i := range mapFile.Mappings {
		mapping := &mapFile.Mappings[i]
		if mapping.RegistryCacheKey == registryCacheKey &&
			mapping.RulesetName == rulesetName &&
			mapping.NormalizedPatterns == normalizedPatterns {
			return mapping, nil
		}
	}

	return nil, fmt.Errorf("mapping not found for %s:%s:%s", registryCacheKey, rulesetName, normalizedPatterns)
}

// UpdateLastAccessed updates the last accessed time for a cache key
func (rm *RulesetMapper) UpdateLastAccessed(cacheKey string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	mapFile, err := rm.LoadMapFile()
	if err != nil {
		return fmt.Errorf("failed to load map file: %w", err)
	}

	for i := range mapFile.Mappings {
		if mapFile.Mappings[i].CacheKey == cacheKey {
			mapFile.Mappings[i].LastAccessed = time.Now()
			return rm.saveMapFile(mapFile)
		}
	}

	return fmt.Errorf("mapping not found for cache key: %s", cacheKey)
}

// RemoveMapping removes a ruleset mapping by cache key
func (rm *RulesetMapper) RemoveMapping(cacheKey string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	mapFile, err := rm.LoadMapFile()
	if err != nil {
		return fmt.Errorf("failed to load map file: %w", err)
	}

	for i := range mapFile.Mappings {
		if mapFile.Mappings[i].CacheKey == cacheKey {
			mapFile.Mappings = append(mapFile.Mappings[:i], mapFile.Mappings[i+1:]...)
			return rm.saveMapFile(mapFile)
		}
	}

	return fmt.Errorf("mapping not found for cache key: %s", cacheKey)
}

// ListMappings returns all ruleset mappings
func (rm *RulesetMapper) ListMappings() ([]RulesetMapping, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	mapFile, err := rm.LoadMapFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load map file: %w", err)
	}

	return mapFile.Mappings, nil
}

// ListMappingsByRegistry returns all ruleset mappings for a specific registry
func (rm *RulesetMapper) ListMappingsByRegistry(registryCacheKey string) ([]RulesetMapping, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	mapFile, err := rm.LoadMapFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load map file: %w", err)
	}

	var mappings []RulesetMapping
	for i := range mapFile.Mappings {
		if mapFile.Mappings[i].RegistryCacheKey == registryCacheKey {
			mappings = append(mappings, mapFile.Mappings[i])
		}
	}

	return mappings, nil
}

// normalizePatterns normalizes patterns for consistent comparison
func (rm *RulesetMapper) normalizePatterns(patterns []string) string {
	if len(patterns) == 0 {
		return ""
	}

	normalized := make([]string, len(patterns))
	for i, pattern := range patterns {
		normalized[i] = strings.TrimSpace(pattern)
	}
	sort.Strings(normalized)
	return strings.Join(normalized, ",")
}

// LoadMapFile loads the ruleset mapping file
func (rm *RulesetMapper) LoadMapFile() (*RulesetMapFile, error) {
	if _, err := os.Stat(rm.mapFilePath); os.IsNotExist(err) {
		// Create empty map file if it doesn't exist
		return &RulesetMapFile{
			Version:  "1.0",
			Mappings: []RulesetMapping{},
		}, nil
	}

	data, err := os.ReadFile(rm.mapFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read map file: %w", err)
	}

	var mapFile RulesetMapFile
	if err := json.Unmarshal(data, &mapFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal map file: %w", err)
	}

	return &mapFile, nil
}

// saveMapFile saves the ruleset mapping file atomically
func (rm *RulesetMapper) saveMapFile(mapFile *RulesetMapFile) error {
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
