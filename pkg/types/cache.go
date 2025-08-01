package types

import (
	"fmt"
	"os"
	"path/filepath"
)

// CacheManager handles cache directory operations
type CacheManager struct {
	CacheDir string
}

// NewCacheManager creates a new cache manager
func NewCacheManager() (*CacheManager, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return nil, err
	}
	
	return &CacheManager{
		CacheDir: cacheDir,
	}, nil
}

// GetCacheDir returns the cache directory path
func GetCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	
	cacheDir := filepath.Join(homeDir, ".arm", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	return cacheDir, nil
}

// GetRulesetCachePath returns the cache path for a specific ruleset version
func (cm *CacheManager) GetRulesetCachePath(rulesetName, version string) string {
	org, pkg := ParseRulesetName(rulesetName)
	if org != "" {
		return filepath.Join(cm.CacheDir, "@"+org, pkg, version)
	}
	return filepath.Join(cm.CacheDir, pkg, version)
}

// GetPackagePath returns the path to the cached package file
func (cm *CacheManager) GetPackagePath(rulesetName, version string) string {
	cachePath := cm.GetRulesetCachePath(rulesetName, version)
	return filepath.Join(cachePath, "package.tar.gz")
}

// EnsureCacheDir creates the cache directory for a ruleset if it doesn't exist
func (cm *CacheManager) EnsureCacheDir(rulesetName, version string) error {
	cachePath := cm.GetRulesetCachePath(rulesetName, version)
	return os.MkdirAll(cachePath, 0755)
}

// IsCached checks if a ruleset version is cached
func (cm *CacheManager) IsCached(rulesetName, version string) bool {
	packagePath := cm.GetPackagePath(rulesetName, version)
	_, err := os.Stat(packagePath)
	return err == nil
}

// GetTargetPath returns the target installation path for a ruleset
func GetTargetPath(target, rulesetName, version string) string {
	org, pkg := ParseRulesetName(rulesetName)
	if org != "" {
		return filepath.Join(target, "arm", "@"+org, pkg, version)
	}
	return filepath.Join(target, "arm", pkg, version)
}