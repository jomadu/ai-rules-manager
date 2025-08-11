package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BaseGitRegistry contains shared logic for all Git registry types
type BaseGitRegistry struct {
	config *RegistryConfig
	auth   *AuthConfig
}

// NewBaseGitRegistry creates a new base Git registry instance
func NewBaseGitRegistry(config *RegistryConfig, auth *AuthConfig) (*BaseGitRegistry, error) {
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &BaseGitRegistry{
		config: config,
		auth:   auth,
	}, nil
}

// GetConfig returns the registry configuration
func (b *BaseGitRegistry) GetConfig() *RegistryConfig {
	return b.config
}

// GetAuth returns the authentication configuration
func (b *BaseGitRegistry) GetAuth() *AuthConfig {
	return b.auth
}

// GetName returns the registry name
func (b *BaseGitRegistry) GetName() string {
	return b.config.Name
}

// WriteFilesToDest writes files from memory to destination directory
func (b *BaseGitRegistry) WriteFilesToDest(files map[string][]byte, destDir string) error {
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for filePath, content := range files {
		// Validate path to prevent traversal
		if !ValidatePath(filePath) {
			continue
		}

		destPath := filepath.Join(destDir, filePath)
		destFileDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destFileDir, 0o700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destFileDir, err)
		}

		if err := os.WriteFile(destPath, content, 0o644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}
	}

	return nil
}

// CreateDownloadResult creates a structured download result
func (b *BaseGitRegistry) CreateDownloadResult(versionSpec, resolvedVersion string, files map[string][]byte, destDir string) *DownloadResult {
	var filePaths []string
	for filePath := range files {
		filePaths = append(filePaths, filepath.Join(destDir, filePath))
	}

	return &DownloadResult{
		VersionSpec:     versionSpec,
		ResolvedVersion: resolvedVersion,
		Files:           filePaths,
	}
}

// ConvertFilesToRulesets converts file map to RulesetInfo slice
func (b *BaseGitRegistry) ConvertFilesToRulesets(files map[string][]byte, registryType string) []RulesetInfo {
	var rulesets []RulesetInfo
	for filePath := range files {
		filename := filepath.Base(filePath)
		ruleset := RulesetInfo{
			Name:      strings.TrimSuffix(filename, filepath.Ext(filename)),
			Version:   "latest",
			Registry:  b.config.Name,
			Type:      registryType,
			Patterns:  []string{filePath},
			UpdatedAt: time.Now(),
		}
		rulesets = append(rulesets, ruleset)
	}
	return rulesets
}

// FindRulesetByName finds a specific ruleset by name from a slice
func (b *BaseGitRegistry) FindRulesetByName(rulesets []RulesetInfo, name, version string) (*RulesetInfo, error) {
	for i := range rulesets {
		if rulesets[i].Name == name {
			rulesets[i].Version = version
			return &rulesets[i], nil
		}
	}
	return nil, fmt.Errorf("ruleset %s not found", name)
}

// DownloadRulesetWithPatterns provides shared download logic with pattern matching
func (b *BaseGitRegistry) DownloadRulesetWithPatterns(ctx context.Context, operations GitOperations, version, destDir string, patterns []string) error {
	if len(patterns) == 0 {
		patterns = []string{"**/*"}
	}

	files, err := operations.GetFiles(ctx, version, patterns)
	if err != nil {
		return err
	}

	return b.WriteFilesToDest(files, destDir)
}

// DownloadRulesetWithResult provides shared download logic returning structured results
func (b *BaseGitRegistry) DownloadRulesetWithResult(ctx context.Context, operations GitOperations, versionSpec, destDir string, patterns []string) (*DownloadResult, error) {
	if len(patterns) == 0 {
		patterns = []string{"**/*"}
	}

	resolvedVersion, err := operations.ResolveVersion(ctx, versionSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve version: %w", err)
	}

	// Use specific repository path if available (for caching)
	var files map[string][]byte
	if remoteOps, ok := operations.(*RemoteGitOperations); ok {
		// Check if destDir looks like a repository cache path
		if strings.Contains(destDir, "/repository") {
			// Calculate rulesets directory path
			rulesetsDir := strings.Replace(destDir, "/repository", "/rulesets", 1)
			files, err = remoteOps.getFilesCloneAtWithRulesetCache(ctx, destDir, rulesetsDir, resolvedVersion, patterns)
		} else {
			files, err = operations.GetFiles(ctx, resolvedVersion, patterns)
		}
	} else {
		files, err = operations.GetFiles(ctx, resolvedVersion, patterns)
	}

	if err != nil {
		return nil, err
	}

	// Don't write files to dest when using repository cache path - they're already there
	if !strings.Contains(destDir, "/repository") {
		if err := b.WriteFilesToDest(files, destDir); err != nil {
			return nil, err
		}
	}

	return b.CreateDownloadResult(versionSpec, resolvedVersion, files, destDir), nil
}
