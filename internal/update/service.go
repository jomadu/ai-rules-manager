package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/max-dunn/ai-rules-manager/internal/config"
	"github.com/max-dunn/ai-rules-manager/internal/install"
	"github.com/max-dunn/ai-rules-manager/internal/registry"
)

// UpdateResult contains the result of an update operation
type UpdateResult struct {
	Registry        string
	Ruleset         string
	Updated         bool
	Version         string
	PreviousVersion string
}

// Service handles ruleset updates
type Service struct {
	config    *config.Config
	installer *install.Installer
}

// New creates a new update service
func New(cfg *config.Config) *Service {
	return &Service{
		config:    cfg,
		installer: install.New(cfg),
	}
}

// CheckOutdated checks if a ruleset is outdated by comparing installed version with latest available
func (s *Service) CheckOutdated(ctx context.Context, rulesetSpec string) (*UpdateResult, error) {
	registry, name, _ := parseRulesetSpec(rulesetSpec)

	// Get current locked version
	if s.config.LockFile == nil || s.config.LockFile.Rulesets[registry] == nil {
		return nil, fmt.Errorf("ruleset '%s/%s' is not installed", registry, name)
	}

	locked := s.config.LockFile.Rulesets[registry][name]
	currentVersion := locked.Version
	currentResolved := locked.Resolved

	// Get version constraint from manifest
	var versionSpec string
	if s.config.Rulesets[registry] != nil && s.config.Rulesets[registry][name].Version != "" {
		versionSpec = s.config.Rulesets[registry][name].Version
	} else {
		versionSpec = currentVersion
	}

	// Resolve latest version for the same version spec
	latestResolved, err := s.resolveLatestVersion(ctx, registry, name, currentResolved, versionSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve latest version: %w", err)
	}

	// For Git registries, compare resolved hashes, not version specs
	isOutdated := false
	if regConfig := s.config.RegistryConfigs[registry]; regConfig != nil && regConfig["type"] == "git" {
		// Compare resolved hashes for Git registries
		isOutdated = currentResolved != latestResolved
	} else {
		// Compare versions for other registry types
		isOutdated = currentVersion != latestResolved
	}

	result := &UpdateResult{
		Registry:        registry,
		Ruleset:         name,
		Version:         latestResolved,
		PreviousVersion: currentVersion,
		Updated:         isOutdated,
	}

	// Don't perform actual update for outdated check
	return result, nil
}

// UpdateRuleset updates a single ruleset to the latest version matching its constraint
func (s *Service) UpdateRuleset(ctx context.Context, rulesetSpec string) (*UpdateResult, error) {
	registry, name, _ := parseRulesetSpec(rulesetSpec)

	// Get current locked version
	if s.config.LockFile == nil || s.config.LockFile.Rulesets[registry] == nil {
		return nil, fmt.Errorf("ruleset '%s/%s' is not installed", registry, name)
	}

	locked := s.config.LockFile.Rulesets[registry][name]
	currentVersion := locked.Version

	// Get version constraint from manifest
	var versionSpec string
	if s.config.Rulesets[registry] != nil && s.config.Rulesets[registry][name].Version != "" {
		versionSpec = s.config.Rulesets[registry][name].Version
	} else {
		versionSpec = "latest"
	}

	// Resolve latest version
	latestVersion, err := s.resolveLatestVersion(ctx, registry, name, currentVersion, versionSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve latest version: %w", err)
	}

	result := &UpdateResult{
		Registry:        registry,
		Ruleset:         name,
		Version:         latestVersion,
		PreviousVersion: currentVersion,
		Updated:         latestVersion != currentVersion,
	}

	// Perform actual file operations if version changed
	if result.Updated {
		if err := s.performUpdate(ctx, registry, name, latestVersion); err != nil {
			return nil, fmt.Errorf("failed to perform update: %w", err)
		}
	}

	return result, nil
}

// resolveLatestVersion resolves a version spec to the latest matching version
func (s *Service) resolveLatestVersion(ctx context.Context, registryName, _, currentVersion, versionSpec string) (string, error) {
	// Create registry configuration
	registryConfig := &registry.RegistryConfig{
		Name: registryName,
		Type: s.config.RegistryConfigs[registryName]["type"],
		URL:  s.config.Registries[registryName],
	}

	// Create auth configuration
	authConfig := &registry.AuthConfig{}
	if regConfig := s.config.RegistryConfigs[registryName]; regConfig != nil {
		authConfig.Token = regConfig["authToken"]
		authConfig.Region = regConfig["region"]
		authConfig.Profile = regConfig["profile"]
	}

	// Create registry instance
	reg, err := registry.CreateRegistry(registryConfig, authConfig)
	if err != nil {
		return currentVersion, fmt.Errorf("failed to create registry: %w", err)
	}
	defer func() { _ = reg.Close() }()

	// For Git registries, resolve the version spec
	if regConfig := s.config.RegistryConfigs[registryName]; regConfig != nil && regConfig["type"] == "git" {
		gitReg, ok := reg.(*registry.GitRegistry)
		if !ok {
			return currentVersion, fmt.Errorf("expected Git registry but got %T", reg)
		}

		resolvedVersion, err := gitReg.ResolveVersion(ctx, versionSpec)
		if err != nil {
			return currentVersion, fmt.Errorf("failed to resolve version: %w", err)
		}

		return resolvedVersion, nil
	}

	// For other registry types, return current version
	return currentVersion, nil
}

// performUpdate performs the actual file operations for an update
func (s *Service) performUpdate(ctx context.Context, registryName, name, newVersion string) error {
	// Get patterns from manifest
	var patterns []string
	if s.config.Rulesets[registryName] != nil && s.config.Rulesets[registryName][name].Patterns != nil {
		patterns = s.config.Rulesets[registryName][name].Patterns
	}

	// Create registry configuration
	registryConfig := &registry.RegistryConfig{
		Name: registryName,
		Type: s.config.RegistryConfigs[registryName]["type"],
		URL:  s.config.Registries[registryName],
	}

	// Create auth configuration
	authConfig := &registry.AuthConfig{}
	if regConfig := s.config.RegistryConfigs[registryName]; regConfig != nil {
		authConfig.Token = regConfig["authToken"]
		authConfig.Region = regConfig["region"]
		authConfig.Profile = regConfig["profile"]
	}

	// Create registry instance
	reg, err := registry.CreateRegistry(registryConfig, authConfig)
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}
	defer func() { _ = reg.Close() }()

	// Download new version
	sourceFiles, err := s.downloadRuleset(ctx, reg, name, newVersion, patterns)
	if err != nil {
		return fmt.Errorf("failed to download ruleset: %w", err)
	}

	// Install using the installer (which handles file replacement and lock file updates)
	req := &install.InstallRequest{
		Registry:    registryName,
		Ruleset:     name,
		Version:     newVersion,
		SourceFiles: sourceFiles,
		Channels:    nil, // Use all configured channels
	}

	_, err = s.installer.Install(req)
	return err
}

// downloadRuleset downloads a ruleset and returns the source files
func (s *Service) downloadRuleset(ctx context.Context, reg registry.Registry, name, version string, patterns []string) ([]string, error) {
	// For Git registries, use structured download
	if gitReg, ok := reg.(*registry.GitRegistry); ok {
		tempDir, err := createTempDir()
		if err != nil {
			return nil, err
		}

		result, err := gitReg.DownloadRulesetWithResult(ctx, name, version, tempDir, patterns)
		if err != nil {
			return nil, err
		}

		return result.Files, nil
	}

	// For other registry types, use standard download
	tempDir, err := createTempDir()
	if err != nil {
		return nil, err
	}

	if err := reg.DownloadRuleset(ctx, name, version, tempDir); err != nil {
		return nil, err
	}

	// Find downloaded files
	return findFiles(tempDir)
}

// createTempDir creates a temporary directory for downloads
func createTempDir() (string, error) {
	return os.MkdirTemp("", "arm-update-*")
}

// findFiles finds all files in a directory
func findFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// parseRulesetSpec parses a ruleset specification into registry, name, and version
func parseRulesetSpec(spec string) (registry, name, version string) {
	// Handle version specification (name@version)
	if idx := strings.LastIndex(spec, "@"); idx != -1 {
		version = spec[idx+1:]
		spec = spec[:idx]
	} else {
		version = "latest"
	}

	// Handle registry specification (registry/name)
	if idx := strings.Index(spec, "/"); idx != -1 {
		registry = spec[:idx]
		name = spec[idx+1:]
	} else {
		name = spec
	}

	return registry, name, version
}
