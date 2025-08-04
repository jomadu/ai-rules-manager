package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type MetadataCache struct {
	storage *Storage
	config  CacheConfig
}

func NewMetadataCache(storage *Storage, config CacheConfig) *MetadataCache {
	return &MetadataCache{
		storage: storage,
		config:  config,
	}
}

func (mc *MetadataCache) GetVersions(registryURL string) ([]string, bool) {
	var versions []string
	if mc.get(registryURL, "versions.json", mc.config.VersionTTL, &versions) {
		return versions, true
	}
	return nil, false
}

func (mc *MetadataCache) StoreVersions(registryURL string, versions []string) error {
	return mc.store(registryURL, "versions.json", versions)
}

func (mc *MetadataCache) GetMetadata(registryURL string) (map[string]interface{}, bool) {
	var metadata map[string]interface{}
	if mc.get(registryURL, "metadata.json", mc.config.MetadataTTL, &metadata) {
		return metadata, true
	}
	return nil, false
}

func (mc *MetadataCache) StoreMetadata(registryURL string, metadata map[string]interface{}) error {
	return mc.store(registryURL, "metadata.json", metadata)
}

func (mc *MetadataCache) get(registryURL, filename string, ttl time.Duration, target interface{}) bool {
	metadataPath := mc.storage.MetadataPath(registryURL)
	filePath := filepath.Join(metadataPath, filename)

	if !mc.storage.Exists(filePath) {
		return false
	}

	if mc.storage.IsExpired(filePath, ttl) {
		return false
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	return json.Unmarshal(data, target) == nil
}

func (mc *MetadataCache) store(registryURL, filename string, data interface{}) error {
	metadataPath := mc.storage.MetadataPath(registryURL)
	if err := os.MkdirAll(metadataPath, 0o755); err != nil {
		return fmt.Errorf("failed to create metadata cache directory: %w", err)
	}

	filePath := filepath.Join(metadataPath, filename)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write metadata cache: %w", err)
	}

	return nil
}
