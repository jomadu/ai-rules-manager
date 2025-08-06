package registry

import (
	"fmt"
	"log"
	"strings"

	"github.com/jomadu/arm/internal/cache"
	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/errors"
)

// ConfigManager interface for dependency injection
type ConfigManager interface {
	GetConfig() *config.ARMConfig
	GetSource(name string) (config.Source, bool)
	SetSource(name string, source *config.Source)
	Load() error
}

// Manager manages multiple registries and provides registry selection
type Manager struct {
	configManager ConfigManager
	registries    map[string]Registry
	cache         *cache.Manager
}

// NewManager creates a new registry manager
func NewManager(configManager ConfigManager) *Manager {
	cacheManager, err := cache.NewManager()
	if err != nil {
		log.Printf("Warning: Cache initialization failed, performance may be reduced: %v", err)
	}

	return &Manager{
		configManager: configManager,
		registries:    make(map[string]Registry),
		cache:         cacheManager,
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
		return nil, errors.SourceNotFound(name)
	}

	// Create registry based on type
	registry, err := m.createRegistry(&source)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrConfigInvalid, fmt.Sprintf("Failed to create registry '%s'", name)).
			WithContext("registry", name).
			WithContext("type", source.Type)
	}

	// Cache the registry
	m.registries[name] = registry

	return registry, nil
}

// GetRegistryForRuleset determines which registry to use for a ruleset
func (m *Manager) GetRegistryForRuleset(rulesetName string) (Registry, error) {
	registryName := m.ParseRegistryName(rulesetName)
	return m.GetRegistry(registryName)
}

// InvalidateCache clears the registry cache
func (m *Manager) InvalidateCache() {
	m.registries = make(map[string]Registry)
}

// ParseRegistryName extracts registry name from ruleset name
// Examples: "company@typescript-rules" -> "company", "typescript-rules" -> "default" or auto-detected source
func (m *Manager) ParseRegistryName(rulesetName string) string {
	parts := strings.Split(rulesetName, "@")
	if len(parts) >= 2 {
		// Check if first part is a registry name
		if _, exists := m.configManager.GetSource(parts[0]); exists {
			return parts[0]
		}
	}

	// Check if default source exists
	if _, exists := m.configManager.GetSource("default"); exists {
		return "default"
	}

	// Auto-detect single source when default doesn't exist
	config := m.configManager.GetConfig()
	if config != nil && len(config.Sources) == 1 {
		for sourceName := range config.Sources {
			return sourceName
		}
	}

	return "default" // Will trigger SourceNotFound error with helpful message
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

// GetConcurrency returns the concurrency limit for a registry source
func (m *Manager) GetConcurrency(sourceName string) int {
	source, exists := m.configManager.GetSource(sourceName)
	if !exists {
		return m.configManager.GetConfig().Performance.DefaultConcurrency
	}

	// 1. Source-specific override (highest priority)
	if source.Concurrency > 0 {
		return source.Concurrency
	}

	// 2. Registry type default
	if typeConfig, exists := m.configManager.GetConfig().Performance.RegistryTypes[source.Type]; exists && typeConfig.Concurrency > 0 {
		return typeConfig.Concurrency
	}

	// 3. Global default with registry type fallbacks
	defaultConcurrency := m.configManager.GetConfig().Performance.DefaultConcurrency
	if defaultConcurrency <= 0 {
		// Hardcoded fallbacks by registry type
		switch source.Type {
		case "gitlab":
			return 2
		case "s3":
			return 8
		case "http":
			return 4
		case "filesystem":
			return 10
		default:
			return 3
		}
	}

	return defaultConcurrency
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
			return nil, errors.New(errors.ErrConfigInvalid, "GitLab registry requires either projectID or groupID").
				WithSuggestion("Add projectID with: arm config set sources.<name>.projectID <id>").
				WithSuggestion("Or add groupID with: arm config set sources.<name>.groupID <id>")
		}
		return NewGitLab(source.URL, source.AuthToken, source.ProjectID, source.GroupID), nil

	case RegistryTypeS3:
		if source.Bucket == "" || source.Region == "" {
			return nil, errors.New(errors.ErrConfigInvalid, "S3 registry requires bucket and region").
				WithSuggestion("Add bucket with: arm config set sources.<name>.bucket <bucket-name>").
				WithSuggestion("Add region with: arm config set sources.<name>.region <aws-region>")
		}
		return NewS3(source.AuthToken, source.Bucket, source.Region, source.Prefix), nil

	case RegistryTypeFilesystem:
		if source.Path == "" {
			return nil, errors.New(errors.ErrConfigInvalid, "Filesystem registry requires path").
				WithSuggestion("Add path with: arm config set sources.<name>.path <directory-path>")
		}
		return NewFilesystem(source.Path), nil

	case RegistryTypeGit:
		if source.URL == "" {
			return nil, errors.New(errors.ErrConfigInvalid, "Git registry requires URL").
				WithSuggestion("Add URL with: arm config set sources.<n>.url <git-repo-url>")
		}
		return NewGitRegistry(source.Name, source.URL, source.AuthToken, source.APIType)

	case RegistryTypeGeneric:
		return NewGenericHTTP(source.URL, source.AuthToken), nil

	default:
		return nil, errors.New(errors.ErrConfigInvalid, fmt.Sprintf("Unsupported registry type: %s", regType)).
			WithSuggestion("Supported types: gitlab, s3, http, filesystem, git").
			WithSuggestion("Set type with: arm config set sources.<name>.type <type>")
	}
}
