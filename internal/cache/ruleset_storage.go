package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// RulesetStorage handles storing individual files in the cache with ruleset/version structure
type RulesetStorage struct {
	cacheManager Manager
}

// NewRulesetStorage creates a new ruleset storage manager
func NewRulesetStorage(cacheManager Manager) *RulesetStorage {
	return &RulesetStorage{
		cacheManager: cacheManager,
	}
}

// StoreRulesetFiles stores individual files for a ruleset version with patterns
func (rs *RulesetStorage) StoreRulesetFiles(registryType, registryURL, rulesetName, version string, files map[string][]byte, patterns []string) error {
	rulesetPath, err := rs.GetRulesetVersionPathWithPatterns(registryType, registryURL, rulesetName, version, patterns)
	if err != nil {
		return fmt.Errorf("failed to get ruleset path: %w", err)
	}

	// Create the version directory
	if err := os.MkdirAll(rulesetPath, 0o755); err != nil {
		return fmt.Errorf("failed to create ruleset directory: %w", err)
	}

	// Store each file
	for filename, content := range files {
		filePath := filepath.Join(rulesetPath, filename)

		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return fmt.Errorf("failed to create file directory: %w", err)
		}

		if err := os.WriteFile(filePath, content, 0o644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filename, err)
		}
	}

	return nil
}

// StoreRulesetFilesFromPaths copies files from source paths to cache with patterns
func (rs *RulesetStorage) StoreRulesetFilesFromPaths(registryType, registryURL, rulesetName, version string, filePaths []string, sourceDir string, patterns []string) error {
	rulesetPath, err := rs.GetRulesetVersionPathWithPatterns(registryType, registryURL, rulesetName, version, patterns)
	if err != nil {
		return fmt.Errorf("failed to get ruleset path: %w", err)
	}

	// Create the version directory
	if err := os.MkdirAll(rulesetPath, 0o755); err != nil {
		return fmt.Errorf("failed to create ruleset directory: %w", err)
	}

	// Copy each file
	for _, filePath := range filePaths {
		srcPath := filepath.Join(sourceDir, filePath)
		destPath := filepath.Join(rulesetPath, filePath)

		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return fmt.Errorf("failed to create file directory: %w", err)
		}

		if err := rs.copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy file %s: %w", filePath, err)
		}
	}

	return nil
}

// GetRulesetVersionPath returns the path for a specific ruleset version (legacy method)
func (rs *RulesetStorage) GetRulesetVersionPath(registryType, registryURL, rulesetName, version string) (string, error) {
	return rs.GetRulesetVersionPathWithPatterns(registryType, registryURL, rulesetName, version, nil)
}

// GetRulesetVersionPathWithPatterns returns the path for a specific ruleset version with patterns
func (rs *RulesetStorage) GetRulesetVersionPathWithPatterns(registryType, registryURL, rulesetName, version string, patterns []string) (string, error) {
	rulesetCachePath, err := rs.cacheManager.GetRulesetCachePath(registryType, registryURL, rulesetName, patterns)
	if err != nil {
		return "", fmt.Errorf("failed to get ruleset cache path: %w", err)
	}

	return filepath.Join(rulesetCachePath, version), nil
}

// GetRulesetFiles retrieves all files for a specific ruleset version with patterns
func (rs *RulesetStorage) GetRulesetFiles(registryType, registryURL, rulesetName, version string, patterns []string) (map[string][]byte, error) {
	rulesetPath, err := rs.GetRulesetVersionPathWithPatterns(registryType, registryURL, rulesetName, version, patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to get ruleset path: %w", err)
	}

	if _, err := os.Stat(rulesetPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("ruleset version not found in cache")
	}

	files := make(map[string][]byte)

	err = filepath.Walk(rulesetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path from ruleset directory
		relPath, err := filepath.Rel(rulesetPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", relPath, err)
		}

		files[relPath] = content
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk ruleset directory: %w", err)
	}

	return files, nil
}

// ListRulesetVersions returns all cached versions for a ruleset with patterns
func (rs *RulesetStorage) ListRulesetVersions(registryType, registryURL, rulesetName string, patterns []string) ([]string, error) {
	rulesetCachePath, err := rs.cacheManager.GetRulesetCachePath(registryType, registryURL, rulesetName, patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to get ruleset cache path: %w", err)
	}

	if _, err := os.Stat(rulesetCachePath); os.IsNotExist(err) {
		return []string{}, nil // No versions cached
	}

	entries, err := os.ReadDir(rulesetCachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ruleset directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	return versions, nil
}

// ListRulesets returns all cached rulesets for a registry
func (rs *RulesetStorage) ListRulesets(registryType, registryURL string) ([]string, error) {
	rulesetsPath, err := rs.cacheManager.GetRulesetsPath(registryType, registryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get rulesets path: %w", err)
	}

	if _, err := os.Stat(rulesetsPath); os.IsNotExist(err) {
		return []string{}, nil // No rulesets cached
	}

	entries, err := os.ReadDir(rulesetsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rulesets directory: %w", err)
	}

	var rulesets []string
	for _, entry := range entries {
		if entry.IsDir() {
			rulesets = append(rulesets, entry.Name())
		}
	}

	return rulesets, nil
}

// GetRulesetStats returns file count and total size for a ruleset version with patterns
func (rs *RulesetStorage) GetRulesetStats(registryType, registryURL, rulesetName, version string, patterns []string) (fileCount int, totalSize int64, err error) {
	rulesetPath, err := rs.GetRulesetVersionPathWithPatterns(registryType, registryURL, rulesetName, version, patterns)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get ruleset path: %w", err)
	}

	if _, err := os.Stat(rulesetPath); os.IsNotExist(err) {
		return 0, 0, fmt.Errorf("ruleset version not found in cache")
	}

	var count int
	var size int64

	err = filepath.Walk(rulesetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			count++
			size += info.Size()
		}

		return nil
	})

	if err != nil {
		return 0, 0, fmt.Errorf("failed to walk ruleset directory: %w", err)
	}

	return count, size, nil
}

// RemoveRulesetVersion removes a specific version of a ruleset from cache with patterns
func (rs *RulesetStorage) RemoveRulesetVersion(registryType, registryURL, rulesetName, version string, patterns []string) error {
	rulesetPath, err := rs.GetRulesetVersionPathWithPatterns(registryType, registryURL, rulesetName, version, patterns)
	if err != nil {
		return fmt.Errorf("failed to get ruleset path: %w", err)
	}

	if err := os.RemoveAll(rulesetPath); err != nil {
		return fmt.Errorf("failed to remove ruleset version: %w", err)
	}

	return nil
}

// CleanupUnreferencedVersions removes version directories that are no longer referenced
func (rs *RulesetStorage) CleanupUnreferencedVersions(registryType, registryURL, rulesetName string, referencedVersions, patterns []string) error {
	cachedVersions, err := rs.ListRulesetVersions(registryType, registryURL, rulesetName, patterns)
	if err != nil {
		return fmt.Errorf("failed to list cached versions: %w", err)
	}

	// Create a set of referenced versions for quick lookup
	referencedSet := make(map[string]bool)
	for _, version := range referencedVersions {
		referencedSet[version] = true
	}

	// Remove unreferenced versions
	for _, cachedVersion := range cachedVersions {
		if !referencedSet[cachedVersion] {
			if err := rs.RemoveRulesetVersion(registryType, registryURL, rulesetName, cachedVersion, patterns); err != nil {
				// Log error but continue cleanup
				continue
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func (rs *RulesetStorage) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Copy file permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// NormalizeFilePath normalizes file paths for consistent storage
func (rs *RulesetStorage) NormalizeFilePath(path string) string {
	// Convert to forward slashes for consistency
	normalized := filepath.ToSlash(path)

	// Remove leading slashes
	normalized = strings.TrimPrefix(normalized, "/")

	return normalized
}
