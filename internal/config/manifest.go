package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ManifestManager handles arm.json file operations
type ManifestManager struct {
	path string
}

// NewManifestManager creates a new manifest manager
func NewManifestManager(global bool) *ManifestManager {
	path := "arm.json"
	if global {
		path = filepath.Join(os.Getenv("HOME"), ".arm", "arm.json")
	}
	return &ManifestManager{path: path}
}

// AddRuleset adds or updates a ruleset in the manifest
func (m *ManifestManager) AddRuleset(registry, name, version string, patterns []string) error {
	armConfig, err := m.loadOrCreate()
	if err != nil {
		return err
	}

	// Initialize registry map if needed
	if armConfig.Rulesets[registry] == nil {
		armConfig.Rulesets[registry] = make(map[string]RulesetSpec)
	}

	// Update ruleset entry
	armConfig.Rulesets[registry][name] = RulesetSpec{
		Version:  version,
		Patterns: patterns,
	}

	return m.save(armConfig)
}

// RemoveRuleset removes a ruleset from the manifest
func (m *ManifestManager) RemoveRuleset(registry, name string) error {
	armConfig, err := m.loadOrCreate()
	if err != nil {
		return err
	}

	if armConfig.Rulesets[registry] != nil {
		delete(armConfig.Rulesets[registry], name)
		// Remove registry if empty
		if len(armConfig.Rulesets[registry]) == 0 {
			delete(armConfig.Rulesets, registry)
		}
	}

	return m.save(armConfig)
}

// loadOrCreate loads existing manifest or creates a new one
func (m *ManifestManager) loadOrCreate() (*ARMConfig, error) {
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
			return nil, err
		}
		return &ARMConfig{
			Engines:  make(map[string]string),
			Rulesets: make(map[string]map[string]RulesetSpec),
		}, nil
	}

	data, err := os.ReadFile(m.path)
	if err != nil {
		return nil, err
	}

	var armConfig ARMConfig
	if err := json.Unmarshal(data, &armConfig); err != nil {
		return nil, err
	}

	// Initialize maps if nil
	if armConfig.Engines == nil {
		armConfig.Engines = make(map[string]string)
	}
	if armConfig.Rulesets == nil {
		armConfig.Rulesets = make(map[string]map[string]RulesetSpec)
	}

	return &armConfig, nil
}

// save saves the manifest to disk
func (m *ManifestManager) save(armConfig *ARMConfig) error {
	data, err := json.MarshalIndent(armConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	return os.WriteFile(m.path, data, 0o600)
}
