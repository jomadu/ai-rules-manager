package registry

import (
	"fmt"
	"strings"

	"github.com/jomadu/arm/internal/config"
)

// Manager manages multiple registries and provides registry selection
type Manager struct {
	configManager *config.Manager
	registries    map[string]Registry
}

// NewManager creates a new registry manager
func NewManager(configManager *config.Manager) *Manager {
	return &Manager{
		configManager: configManager,
		registries:    make(map[string]Registry),
	}
}

// GetRegistry returns a registry by name, creating it if needed
func (m *Manager) GetRegistry(name string) (Registry, error) {
	// Check cache first
	if registry, exists := m.registries[name]; exists {
		return registry, nil
	}

	// Get source configuration
	source, exists := m.configManager.GetSource(name)
	if !exists {
		return nil, fmt.Errorf("registry '%s' not found in configuration", name)
	}

	// Create registry based on type
	registry, err := m.createRegistry(&source)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry '%s': %w", name, err)
	}

	// Cache the registry
	m.registries[name] = registry

	return registry, nil
}

// GetRegistryForRuleset determines which registry to use for a ruleset
func (m *Manager) GetRegistryForRuleset(rulesetName string) (Registry, error) {
	registryName := m.parseRegistryName(rulesetName)
	return m.GetRegistry(registryName)
}

// InvalidateCache clears the registry cache
func (m *Manager) InvalidateCache() {
	m.registries = make(map[string]Registry)
}

// parseRegistryName extracts registry name from ruleset name
// Examples: "company@typescript-rules" -> "company", "typescript-rules" -> "default"
func (m *Manager) parseRegistryName(rulesetName string) string {
	parts := strings.Split(rulesetName, "@")
	if len(parts) >= 2 {
		// Check if first part is a registry name
		if _, exists := m.configManager.GetSource(parts[0]); exists {
			return parts[0]
		}
	}
	return "default"
}

// StripRegistryPrefix removes registry prefix from ruleset name
// Examples: "company@typescript-rules" -> "typescript-rules", "typescript-rules" -> "typescript-rules"
func (m *Manager) StripRegistryPrefix(rulesetName string) string {
	parts := strings.Split(rulesetName, "@")
	if len(parts) >= 2 {
		// Check if first part is a registry name
		if _, exists := m.configManager.GetSource(parts[0]); exists {
			return strings.Join(parts[1:], "@")
		}
	}
	return rulesetName
}

// createRegistry creates a registry instance based on source configuration
func (m *Manager) createRegistry(source *config.Source) (Registry, error) {
	regType := RegistryType(source.Type)
	if regType == "" {
		regType = RegistryTypeGeneric
	}

	switch regType {
	case RegistryTypeGitLab:
		if source.ProjectID == "" && source.GroupID == "" {
			return nil, fmt.Errorf("either projectID or groupID is required for GitLab registry")
		}
		return NewGitLab(source.URL, source.AuthToken, source.ProjectID, source.GroupID), nil

	case RegistryTypeS3:
		if source.Bucket == "" || source.Region == "" {
			return nil, fmt.Errorf("bucket and region are required for S3 registry")
		}
		return NewS3(source.AuthToken, source.Bucket, source.Region, source.Prefix), nil

	case RegistryTypeFilesystem:
		if source.Path == "" {
			return nil, fmt.Errorf("path is required for filesystem registry")
		}
		return NewFilesystem(source.Path), nil

	case RegistryTypeGeneric:
		return NewGenericHTTP(source.URL, source.AuthToken), nil

	default:
		return nil, fmt.Errorf("unsupported registry type: %s", regType)
	}
}
