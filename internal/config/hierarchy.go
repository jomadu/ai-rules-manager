package config

import (
	"os"
	"path/filepath"
)

// Manager handles configuration loading and hierarchy
type Manager struct {
	config *ARMConfig
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{}
}

// Load loads configuration from the hierarchy (user -> project -> env -> flags)
func (m *Manager) Load() error {
	config := &ARMConfig{
		Sources: make(map[string]Source),
	}

	// 1. Load user-level config (~/.armrc)
	if userConfig, err := loadUserConfig(); err == nil {
		mergeConfigs(config, userConfig)
	}

	// 2. Load project-level config (./.armrc)
	if projectConfig, err := loadProjectConfig(); err == nil {
		mergeConfigs(config, projectConfig)
	}

	// TODO: 3. Environment variables override (future)
	// TODO: 4. Command-line flags override (future)

	m.config = config
	return nil
}

// GetConfig returns the loaded configuration
func (m *Manager) GetConfig() *ARMConfig {
	return m.config
}

// GetSource returns a specific source configuration
func (m *Manager) GetSource(name string) (Source, bool) {
	source, exists := m.config.Sources[name]
	return source, exists
}

// SetSource sets a source configuration
func (m *Manager) SetSource(name string, source *Source) {
	m.config.Sources[name] = *source
}

// loadUserConfig loads configuration from ~/.armrc
func loadUserConfig() (*ARMConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".armrc")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ARMConfig{Sources: make(map[string]Source)}, nil
	}

	return ParseFile(configPath)
}

// loadProjectConfig loads configuration from ./.armrc
func loadProjectConfig() (*ARMConfig, error) {
	configPath := ".armrc"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ARMConfig{Sources: make(map[string]Source)}, nil
	}

	return ParseFile(configPath)
}

// mergeConfigs merges source config into target config (source takes precedence)
func mergeConfigs(target, source *ARMConfig) {
	// Merge sources
	for name, sourceConfig := range source.Sources {
		target.Sources[name] = sourceConfig
	}

	// Merge cache config (only if source has values)
	if source.Cache.Location != "" {
		target.Cache.Location = source.Cache.Location
	}
	if source.Cache.MaxSize != "" {
		target.Cache.MaxSize = source.Cache.MaxSize
	}
}
