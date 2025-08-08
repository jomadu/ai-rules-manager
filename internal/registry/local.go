package registry

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalRegistry implements the Registry interface for local filesystem registries
type LocalRegistry struct {
	config *RegistryConfig
	path   string
}

// NewLocalRegistry creates a new local filesystem registry instance
func NewLocalRegistry(config *RegistryConfig) (*LocalRegistry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Convert relative paths to absolute paths
	path := config.URL
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve path: %w", err)
		}
		path = absPath
	}

	// Check if path exists and is accessible
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("registry path not accessible: %w", err)
	}

	return &LocalRegistry{
		config: config,
		path:   path,
	}, nil
}

// GetRulesets returns available rulesets by scanning the filesystem
func (l *LocalRegistry) GetRulesets(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	entries, err := os.ReadDir(l.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry directory: %w", err)
	}

	var rulesets []RulesetInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		rulesetName := entry.Name()
		rulesetPath := filepath.Join(l.path, rulesetName)

		// Find the latest version for this ruleset
		versions, err := l.getVersionsForRuleset(rulesetPath)
		if err != nil || len(versions) == 0 || (len(versions) == 1 && versions[0] == "latest") {
			continue
		}

		// Use the latest version (last in sorted order)
		latestVersion := versions[len(versions)-1]

		// Get file info for timestamp
		info, err := entry.Info()
		var updatedAt time.Time
		if err == nil {
			updatedAt = info.ModTime()
		}

		ruleset := RulesetInfo{
			Name:      rulesetName,
			Version:   latestVersion,
			Registry:  l.config.Name,
			Type:      "local",
			UpdatedAt: updatedAt,
			Metadata: map[string]string{
				"path": l.path,
			},
		}
		rulesets = append(rulesets, ruleset)
	}

	return rulesets, nil
}

// GetRuleset returns detailed information about a specific ruleset
func (l *LocalRegistry) GetRuleset(ctx context.Context, name, version string) (*RulesetInfo, error) {
	rulesetPath := filepath.Join(l.path, name)
	if _, err := os.Stat(rulesetPath); err != nil {
		return nil, fmt.Errorf("ruleset %s not found", name)
	}

	versions, err := l.getVersionsForRuleset(rulesetPath)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for ruleset %s", name)
	}

	// If version is "latest", use the last version in the array
	if version == "latest" {
		version = versions[len(versions)-1]
	} else {
		// Validate that the requested version exists
		found := false
		for _, v := range versions {
			if v == version {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("version %s not found for ruleset %s", version, name)
		}
	}

	// Get file info for timestamp
	versionPath := filepath.Join(rulesetPath, version)
	info, err := os.Stat(versionPath)
	var updatedAt time.Time
	if err == nil {
		updatedAt = info.ModTime()
	}

	return &RulesetInfo{
		Name:      name,
		Version:   version,
		Registry:  l.config.Name,
		Type:      "local",
		UpdatedAt: updatedAt,
		Metadata: map[string]string{
			"path": l.path,
		},
	}, nil
}

// DownloadRuleset copies a ruleset from the local filesystem to the specified directory
func (l *LocalRegistry) DownloadRuleset(ctx context.Context, name, version, destDir string) error {
	// Construct source path: path/ruleset/version/ruleset.tar.gz
	sourcePath := filepath.Join(l.path, name, version, "ruleset.tar.gz")

	// Check if source file exists
	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("ruleset file not found: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination file
	destPath := filepath.Join(destDir, "ruleset.tar.gz")
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy content
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// GetVersions returns available versions for a ruleset
func (l *LocalRegistry) GetVersions(ctx context.Context, name string) ([]string, error) {
	rulesetPath := filepath.Join(l.path, name)
	return l.getVersionsForRuleset(rulesetPath)
}

// GetType returns the registry type
func (l *LocalRegistry) GetType() string {
	return "local"
}

// GetName returns the registry name
func (l *LocalRegistry) GetName() string {
	return l.config.Name
}

// Close cleans up any resources
func (l *LocalRegistry) Close() error {
	return nil
}

// getVersionsForRuleset scans a ruleset directory for available versions
func (l *LocalRegistry) getVersionsForRuleset(rulesetPath string) ([]string, error) {
	if _, err := os.Stat(rulesetPath); err != nil {
		return nil, fmt.Errorf("ruleset directory not found: %w", err)
	}

	entries, err := os.ReadDir(rulesetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ruleset directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		versionName := entry.Name()
		// Check if this version directory contains a ruleset.tar.gz file
		rulesetFile := filepath.Join(rulesetPath, versionName, "ruleset.tar.gz")
		if _, err := os.Stat(rulesetFile); err == nil {
			versions = append(versions, versionName)
		}
	}

	if len(versions) == 0 {
		return []string{"latest"}, nil
	}

	// Sort versions (simple string sort for now)
	// TODO: Implement proper semantic version sorting if needed
	return versions, nil
}