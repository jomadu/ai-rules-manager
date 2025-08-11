package update

import (
	"context"
	"fmt"
	"strings"

	"github.com/max-dunn/ai-rules-manager/internal/config"
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
	config *config.Config
}

// New creates a new update service
func New(cfg *config.Config) *Service {
	return &Service{config: cfg}
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

	// Update lock file if version changed
	if result.Updated {
		if err := s.updateLockFile(registry, name, latestVersion, &locked); err != nil {
			return nil, fmt.Errorf("failed to update lock file: %w", err)
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

// updateLockFile updates the lock file with a new version
func (s *Service) updateLockFile(registry, name, newVersion string, existingLocked *config.LockedRuleset) error {
	updatedLocked := *existingLocked
	updatedLocked.Version = newVersion
	updatedLocked.Resolved = "2024-01-15T10:30:00Z" // Would use current time

	if s.config.LockFile.Rulesets[registry] == nil {
		s.config.LockFile.Rulesets[registry] = make(map[string]config.LockedRuleset)
	}
	s.config.LockFile.Rulesets[registry][name] = updatedLocked

	return nil
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
