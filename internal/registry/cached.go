package registry

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// CachedListVersions wraps ListVersions with caching
func (m *Manager) CachedListVersions(registryName, rulesetName string) ([]string, error) {
	registry, err := m.GetRegistry(registryName)
	if err != nil {
		return nil, err
	}

	// Get registry URL for cache key
	registryURL := m.getRegistryURL(registryName)
	cacheKey := fmt.Sprintf("%s/%s", registryURL, rulesetName)

	// Try cache first
	if m.cache != nil {
		if versions, found := m.cache.GetVersions(cacheKey); found {
			log.Printf("✓ Using cached versions for %s", rulesetName)
			return versions, nil
		}
	}

	// Cache miss - fetch from registry
	log.Printf("⬇ Fetching versions for %s", rulesetName)
	versions, err := registry.ListVersions(rulesetName)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if m.cache != nil {
		if err := m.cache.StoreVersions(cacheKey, versions); err != nil {
			log.Printf("Warning: Failed to cache versions for %s: %v", rulesetName, err)
		}
	}

	return versions, nil
}

// CachedDownload wraps Download with caching
func (m *Manager) CachedDownload(registryName, rulesetName, version string) (io.ReadCloser, error) {
	registry, err := m.GetRegistry(registryName)
	if err != nil {
		return nil, err
	}

	// Get registry URL for cache key
	registryURL := m.getRegistryURL(registryName)

	// Try cache first
	if m.cache != nil {
		if path, found := m.cache.GetPackage(registryURL, rulesetName, version); found {
			log.Printf("✓ Using cached package %s@%s", rulesetName, version)
			return openCachedFile(path)
		}
	}

	// Cache miss - download from registry
	log.Printf("⬇ Downloading %s@%s", rulesetName, version)
	reader, err := registry.Download(rulesetName, version)
	if err != nil {
		return nil, err
	}

	// Store in cache if possible
	if m.cache != nil {
		return m.cacheAndReturn(registryURL, rulesetName, version, reader)
	}

	return reader, nil
}

// CachedGetMetadata wraps GetMetadata with caching
func (m *Manager) CachedGetMetadata(registryName, rulesetName string) (*Metadata, error) {
	registry, err := m.GetRegistry(registryName)
	if err != nil {
		return nil, err
	}

	// Get registry URL for cache key
	registryURL := m.getRegistryURL(registryName)
	cacheKey := fmt.Sprintf("%s/%s", registryURL, rulesetName)

	// Try cache first
	if m.cache != nil {
		if data, found := m.cache.GetMetadata(cacheKey); found {
			log.Printf("✓ Using cached metadata for %s", rulesetName)
			return convertToMetadata(data), nil
		}
	}

	// Cache miss - fetch from registry
	log.Printf("⬇ Fetching metadata for %s", rulesetName)
	metadata, err := registry.GetMetadata(rulesetName)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if m.cache != nil {
		if err := m.cache.StoreMetadata(cacheKey, convertFromMetadata(metadata)); err != nil {
			log.Printf("Warning: Failed to cache metadata for %s: %v", rulesetName, err)
		}
	}

	return metadata, nil
}

// Helper functions

func (m *Manager) getRegistryURL(registryName string) string {
	if source, exists := m.configManager.GetSource(registryName); exists {
		return source.URL
	}
	return registryName // fallback
}

func openCachedFile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (m *Manager) cacheAndReturn(registryURL, rulesetName, version string, reader io.ReadCloser) (io.ReadCloser, error) {
	// Read all data first
	data, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		return nil, err
	}

	// Store in cache in background
	go func() {
		if _, err := m.cache.StorePackage(registryURL, rulesetName, version, strings.NewReader(string(data))); err != nil {
			log.Printf("Warning: Failed to cache package %s@%s: %v", rulesetName, version, err)
		}
	}()

	// Return the data as a reader
	return io.NopCloser(strings.NewReader(string(data))), nil
}

func convertToMetadata(data map[string]interface{}) *Metadata {
	// Convert generic map back to Metadata struct
	// This is a simplified implementation
	metadata := &Metadata{}
	if name, ok := data["name"].(string); ok {
		metadata.Name = name
	}
	if desc, ok := data["description"].(string); ok {
		metadata.Description = desc
	}
	return metadata
}

func convertFromMetadata(metadata *Metadata) map[string]interface{} {
	// Convert Metadata struct to generic map for caching
	return map[string]interface{}{
		"name":        metadata.Name,
		"description": metadata.Description,
		"versions":    metadata.Versions,
		"repository":  metadata.Repository,
		"homepage":    metadata.Homepage,
		"license":     metadata.License,
		"keywords":    metadata.Keywords,
	}
}
