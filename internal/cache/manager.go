package cache

import (
	"fmt"
	"io"
	"log"
)

type Manager struct {
	storage  *Storage
	packages *PackageCache
	metadata *MetadataCache
	config   CacheConfig
}

func NewManager(config ...CacheConfig) (*Manager, error) {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	storage, err := New()
	if err != nil {
		log.Printf("Cache initialization failed, falling back to no-cache mode: %v", err)
		return &Manager{config: cfg}, nil // Return manager without storage for fallback
	}

	return &Manager{
		storage:  storage,
		packages: NewPackageCache(storage, cfg),
		metadata: NewMetadataCache(storage, cfg),
		config:   cfg,
	}, nil
}

func (m *Manager) GetPackage(registryURL, ruleset, version string) (string, bool) {
	if m.packages == nil {
		return "", false
	}
	return m.packages.Get(registryURL, ruleset, version)
}

func (m *Manager) StorePackage(registryURL, ruleset, version string, data io.Reader) (string, error) {
	if m.packages == nil {
		return "", fmt.Errorf("cache not available")
	}
	return m.packages.Store(registryURL, ruleset, version, data)
}

func (m *Manager) GetVersions(registryURL string) ([]string, bool) {
	if m.metadata == nil {
		return nil, false
	}
	return m.metadata.GetVersions(registryURL)
}

func (m *Manager) StoreVersions(registryURL string, versions []string) error {
	if m.metadata == nil {
		return nil // Silently ignore if cache not available
	}
	return m.metadata.StoreVersions(registryURL, versions)
}

func (m *Manager) GetMetadata(registryURL string) (map[string]interface{}, bool) {
	if m.metadata == nil {
		return nil, false
	}
	return m.metadata.GetMetadata(registryURL)
}

func (m *Manager) StoreMetadata(registryURL string, metadata map[string]interface{}) error {
	if m.metadata == nil {
		return nil // Silently ignore if cache not available
	}
	return m.metadata.StoreMetadata(registryURL, metadata)
}
