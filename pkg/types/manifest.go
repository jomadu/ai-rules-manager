package types

import (
	"encoding/json"
	"fmt"
	"os"
)

// RulesManifest represents the rules.json file structure
type RulesManifest struct {
	Targets      []string          `json:"targets" validate:"required,min=1"`
	Dependencies map[string]string `json:"dependencies" validate:"required"`
}

// RulesLock represents the rules.lock file structure
type RulesLock struct {
	Version      string                      `json:"version" validate:"required"`
	Dependencies map[string]LockedDependency `json:"dependencies" validate:"required"`
}

// LockedDependency represents a locked dependency in rules.lock
type LockedDependency struct {
	Version  string `json:"version" validate:"required,semver"`
	Source   string `json:"source" validate:"required,url"`
	Checksum string `json:"checksum" validate:"required,len=64"`
}

// LoadManifest loads and validates a rules.json file
func LoadManifest(path string) (*RulesManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest RulesManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return &manifest, nil
}

// SaveManifest saves a rules.json file
func (m *RulesManifest) SaveManifest(path string) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// Validate performs validation on the manifest
func (m *RulesManifest) Validate() error {
	if len(m.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}
	if m.Dependencies == nil {
		m.Dependencies = make(map[string]string)
	}
	return nil
}

// LoadLockFile loads and validates a rules.lock file
func LoadLockFile(path string) (*RulesLock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lock RulesLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	if err := lock.Validate(); err != nil {
		return nil, fmt.Errorf("invalid lock file: %w", err)
	}

	return &lock, nil
}

// SaveLockFile saves a rules.lock file
func (l *RulesLock) SaveLockFile(path string) error {
	if err := l.Validate(); err != nil {
		return fmt.Errorf("invalid lock file: %w", err)
	}

	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// Validate performs validation on the lock file
func (l *RulesLock) Validate() error {
	if l.Version == "" {
		return fmt.Errorf("lock file version is required")
	}
	if l.Dependencies == nil {
		l.Dependencies = make(map[string]LockedDependency)
	}
	for name, dep := range l.Dependencies {
		if dep.Version == "" {
			return fmt.Errorf("dependency %s missing version", name)
		}
		if dep.Source == "" {
			return fmt.Errorf("dependency %s missing source", name)
		}
		if dep.Checksum == "" {
			return fmt.Errorf("dependency %s missing checksum", name)
		}
		if len(dep.Checksum) != 64 {
			return fmt.Errorf("dependency %s has invalid checksum format", name)
		}
	}
	return nil
}
