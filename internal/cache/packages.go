package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type PackageCache struct {
	storage *Storage
	config  CacheConfig
}

func NewPackageCache(storage *Storage, config CacheConfig) *PackageCache {
	return &PackageCache{
		storage: storage,
		config:  config,
	}
}

func (pc *PackageCache) Get(registryURL, ruleset, version string) (string, bool) {
	packagePath := pc.storage.PackagePath(registryURL, ruleset, version)
	tarPath := filepath.Join(packagePath, "package.tar.gz")

	if !pc.storage.Exists(tarPath) {
		return "", false
	}

	if pc.storage.IsExpired(tarPath, pc.config.PackageTTL) {
		return "", false
	}

	return tarPath, true
}

func (pc *PackageCache) Store(registryURL, ruleset, version string, data io.Reader) (string, error) {
	packagePath := pc.storage.PackagePath(registryURL, ruleset, version)
	if err := os.MkdirAll(packagePath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create package cache directory: %w", err)
	}

	tarPath := filepath.Join(packagePath, "package.tar.gz")
	file, err := os.Create(tarPath)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if _, err := io.Copy(file, data); err != nil {
		_ = os.Remove(tarPath) // Clean up on failure
		return "", fmt.Errorf("failed to write cache file: %w", err)
	}

	return tarPath, nil
}
