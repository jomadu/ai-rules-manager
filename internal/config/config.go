package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/ini.v1"
)

// Config represents the ARM configuration
type Config struct {
	Registries map[string]string            // [registries] section
	RegistryConfigs map[string]map[string]string // [registries.name] sections
	TypeDefaults map[string]map[string]string    // [git], [s3], etc. sections
	NetworkConfig map[string]string         // [network] section
	CacheConfig map[string]string           // [cache] section
}

// Load loads the ARM configuration from files
func Load() (*Config, error) {
	cfg := &Config{
		Registries: make(map[string]string),
		RegistryConfigs: make(map[string]map[string]string),
		TypeDefaults: make(map[string]map[string]string),
		NetworkConfig: make(map[string]string),
		CacheConfig: make(map[string]string),
	}

	// Load global config
	globalPath := filepath.Join(os.Getenv("HOME"), ".arm", ".armrc")
	if err := cfg.loadINIFile(globalPath, false); err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Load local config (overrides global)
	localPath := ".armrc"
	if err := cfg.loadINIFile(localPath, false); err != nil {
		return nil, fmt.Errorf("failed to load local config: %w", err)
	}

	return cfg, nil
}

// loadINIFile loads and parses an INI file with environment variable expansion
func (c *Config) loadINIFile(path string, required bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if required {
			return fmt.Errorf("required config file not found: %s", path)
		}
		return nil // Optional file doesn't exist
	}

	cfg, err := ini.Load(path)
	if err != nil {
		return fmt.Errorf("failed to parse INI file %s: %w", path, err)
	}

	// Process sections
	for _, section := range cfg.Sections() {
		sectionName := section.Name()
		if sectionName == "DEFAULT" {
			continue
		}

		if err := c.processSection(section); err != nil {
			return fmt.Errorf("failed to process section [%s]: %w", sectionName, err)
		}
	}

	return nil
}

// processSection processes a single INI section
func (c *Config) processSection(section *ini.Section) error {
	sectionName := section.Name()

	// Handle nested sections like [registries.my-registry]
	if strings.Contains(sectionName, ".") {
		parts := strings.SplitN(sectionName, ".", 2)
		if parts[0] == "registries" {
			return c.processRegistryConfig(parts[1], section)
		}
		return fmt.Errorf("unsupported nested section: %s", sectionName)
	}

	// Handle top-level sections
	switch sectionName {
	case "registries":
		return c.processRegistries(section)
	case "git", "https", "s3", "gitlab", "local":
		return c.processTypeDefaults(sectionName, section)
	case "network":
		return c.processNetworkConfig(section)
	case "cache":
		return c.processCacheConfig(section)
	default:
		return fmt.Errorf("unknown section: %s", sectionName)
	}
}

// processRegistries processes the [registries] section
func (c *Config) processRegistries(section *ini.Section) error {
	for _, key := range section.Keys() {
		value := expandEnvVars(key.String())
		c.Registries[key.Name()] = value
	}
	return nil
}

// processRegistryConfig processes [registries.name] sections
func (c *Config) processRegistryConfig(name string, section *ini.Section) error {
	if c.RegistryConfigs[name] == nil {
		c.RegistryConfigs[name] = make(map[string]string)
	}

	for _, key := range section.Keys() {
		value := expandEnvVars(key.String())
		c.RegistryConfigs[name][key.Name()] = value
	}
	return nil
}

// processTypeDefaults processes registry type default sections
func (c *Config) processTypeDefaults(typeName string, section *ini.Section) error {
	if c.TypeDefaults[typeName] == nil {
		c.TypeDefaults[typeName] = make(map[string]string)
	}

	for _, key := range section.Keys() {
		value := expandEnvVars(key.String())
		c.TypeDefaults[typeName][key.Name()] = value
	}
	return nil
}

// processNetworkConfig processes the [network] section
func (c *Config) processNetworkConfig(section *ini.Section) error {
	for _, key := range section.Keys() {
		value := expandEnvVars(key.String())
		c.NetworkConfig[key.Name()] = value
	}
	return nil
}

// processCacheConfig processes the [cache] section
func (c *Config) processCacheConfig(section *ini.Section) error {
	for _, key := range section.Keys() {
		value := expandEnvVars(key.String())
		c.CacheConfig[key.Name()] = value
	}
	return nil
}

// expandEnvVars expands environment variables in the format $VAR and ${VAR}
func expandEnvVars(s string) string {
	// Pattern matches $VAR and ${VAR}
	pattern := regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)
	
	return pattern.ReplaceAllStringFunc(s, func(match string) string {
		var varName string
		if strings.HasPrefix(match, "${") {
			// ${VAR} format
			varName = match[2 : len(match)-1]
		} else {
			// $VAR format
			varName = match[1:]
		}
		
		// Return environment variable value or empty string if not found
		return os.Getenv(varName)
	})
}
