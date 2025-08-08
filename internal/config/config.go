package config

import (
	"encoding/json"
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
	
	// JSON configuration
	Channels map[string]ChannelConfig       // channels from arm.json
	Rulesets map[string]map[string]RulesetSpec // rulesets from arm.json
	Engines map[string]string               // engines from arm.json
	LockFile *LockFile                      // arm.lock content
}

// ChannelConfig represents a channel configuration
type ChannelConfig struct {
	Directories []string `json:"directories"`
}

// RulesetSpec represents a ruleset specification
type RulesetSpec struct {
	Version  string   `json:"version"`
	Patterns []string `json:"patterns,omitempty"`
}

// ARMConfig represents the arm.json file structure
type ARMConfig struct {
	Engines  map[string]string                    `json:"engines"`
	Channels map[string]ChannelConfig            `json:"channels"`
	Rulesets map[string]map[string]RulesetSpec   `json:"rulesets"`
}

// LockFile represents the arm.lock file structure
type LockFile struct {
	Rulesets map[string]map[string]LockedRuleset `json:"rulesets"`
}

// LockedRuleset represents a locked ruleset entry
type LockedRuleset struct {
	Version  string `json:"version"`
	Resolved string `json:"resolved"`
	Registry string `json:"registry"`
	Type     string `json:"type"`
	Region   string `json:"region,omitempty"`
}

// Load loads the ARM configuration from files with hierarchical merging
func Load() (*Config, error) {
	// Load global configuration first
	globalCfg, err := loadConfigFromPaths(
		filepath.Join(os.Getenv("HOME"), ".arm", ".armrc"),
		filepath.Join(os.Getenv("HOME"), ".arm", "arm.json"),
		"arm.lock", // Lock file is always local
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Load local configuration
	localCfg, err := loadConfigFromPaths(".armrc", "arm.json", "arm.lock")
	if err != nil {
		return nil, fmt.Errorf("failed to load local config: %w", err)
	}

	// Merge configurations (local overrides global at key level)
	mergedCfg := mergeConfigs(globalCfg, localCfg)

	// Validate merged configuration
	if err := validateConfig(mergedCfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return mergedCfg, nil
}

// loadConfigFromPaths loads configuration from specified file paths
func loadConfigFromPaths(iniPath, jsonPath, lockPath string) (*Config, error) {
	cfg := &Config{
		Registries: make(map[string]string),
		RegistryConfigs: make(map[string]map[string]string),
		TypeDefaults: make(map[string]map[string]string),
		NetworkConfig: make(map[string]string),
		CacheConfig: make(map[string]string),
		Channels: make(map[string]ChannelConfig),
		Rulesets: make(map[string]map[string]RulesetSpec),
		Engines: make(map[string]string),
	}

	// Load INI file
	if err := cfg.loadINIFile(iniPath, false); err != nil {
		return nil, fmt.Errorf("failed to load INI file %s: %w", iniPath, err)
	}

	// Load JSON file
	if err := cfg.loadARMJSON(jsonPath, false); err != nil {
		return nil, fmt.Errorf("failed to load JSON file %s: %w", jsonPath, err)
	}

	// Load lock file (only for local config)
	if lockPath == "arm.lock" {
		if err := cfg.loadLockFile(lockPath); err != nil {
			return nil, fmt.Errorf("failed to load lock file %s: %w", lockPath, err)
		}
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

// mergeConfigs merges two configurations with local taking precedence at key level
func mergeConfigs(global, local *Config) *Config {
	merged := &Config{
		Registries: make(map[string]string),
		RegistryConfigs: make(map[string]map[string]string),
		TypeDefaults: make(map[string]map[string]string),
		NetworkConfig: make(map[string]string),
		CacheConfig: make(map[string]string),
		Channels: make(map[string]ChannelConfig),
		Rulesets: make(map[string]map[string]RulesetSpec),
		Engines: make(map[string]string),
	}

	// Merge registries (key-level merge)
	mergeStringMaps(merged.Registries, global.Registries, local.Registries)

	// Merge registry configs (nested map merge)
	mergeNestedStringMaps(merged.RegistryConfigs, global.RegistryConfigs, local.RegistryConfigs)

	// Merge type defaults (nested map merge)
	mergeNestedStringMaps(merged.TypeDefaults, global.TypeDefaults, local.TypeDefaults)

	// Merge network and cache configs (key-level merge)
	mergeStringMaps(merged.NetworkConfig, global.NetworkConfig, local.NetworkConfig)
	mergeStringMaps(merged.CacheConfig, global.CacheConfig, local.CacheConfig)

	// Merge engines (key-level merge)
	mergeStringMaps(merged.Engines, global.Engines, local.Engines)

	// Merge channels (key-level merge)
	mergeChannelMaps(merged.Channels, global.Channels, local.Channels)

	// Merge rulesets (nested map merge)
	mergeRulesetMaps(merged.Rulesets, global.Rulesets, local.Rulesets)

	// Lock file is always from local (no merging needed)
	merged.LockFile = local.LockFile

	return merged
}

// mergeStringMaps merges string maps with local taking precedence
func mergeStringMaps(dest, global, local map[string]string) {
	// Copy global values first
	for k, v := range global {
		dest[k] = v
	}
	// Override with local values
	for k, v := range local {
		dest[k] = v
	}
}

// mergeNestedStringMaps merges nested string maps with local taking precedence
func mergeNestedStringMaps(dest, global, local map[string]map[string]string) {
	// Copy global values first
	for k, v := range global {
		dest[k] = make(map[string]string)
		for kk, vv := range v {
			dest[k][kk] = vv
		}
	}
	// Merge with local values (key-level merge within each nested map)
	for k, v := range local {
		if dest[k] == nil {
			dest[k] = make(map[string]string)
		}
		for kk, vv := range v {
			dest[k][kk] = vv
		}
	}
}

// mergeChannelMaps merges channel maps with local taking precedence
func mergeChannelMaps(dest, global, local map[string]ChannelConfig) {
	// Copy global values first
	for k, v := range global {
		dest[k] = v
	}
	// Override with local values
	for k, v := range local {
		dest[k] = v
	}
}

// mergeRulesetMaps merges ruleset maps with local taking precedence
func mergeRulesetMaps(dest, global, local map[string]map[string]RulesetSpec) {
	// Copy global values first
	for k, v := range global {
		dest[k] = make(map[string]RulesetSpec)
		for kk, vv := range v {
			dest[k][kk] = vv
		}
	}
	// Merge with local values (key-level merge within each registry)
	for k, v := range local {
		if dest[k] == nil {
			dest[k] = make(map[string]RulesetSpec)
		}
		for kk, vv := range v {
			dest[k][kk] = vv
		}
	}
}

// loadARMJSON loads and parses an arm.json file
func (c *Config) loadARMJSON(path string, required bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if required {
			return fmt.Errorf("required JSON file not found: %s", path)
		}
		return nil // Optional file doesn't exist
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read JSON file %s: %w", path, err)
	}

	// Expand environment variables in JSON content
	expandedData := expandEnvVarsInJSON(string(data))

	var armConfig ARMConfig
	if err := json.Unmarshal([]byte(expandedData), &armConfig); err != nil {
		return fmt.Errorf("failed to parse JSON file %s: %w", path, err)
	}

	// Merge into config (local overrides global)
	for k, v := range armConfig.Engines {
		c.Engines[k] = v
	}
	for k, v := range armConfig.Channels {
		c.Channels[k] = v
	}
	for registry, rulesets := range armConfig.Rulesets {
		if c.Rulesets[registry] == nil {
			c.Rulesets[registry] = make(map[string]RulesetSpec)
		}
		for name, spec := range rulesets {
			c.Rulesets[registry][name] = spec
		}
	}

	return nil
}

// loadLockFile loads and parses an arm.lock file
func (c *Config) loadLockFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Lock file is optional
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read lock file %s: %w", path, err)
	}

	var lockFile LockFile
	if err := json.Unmarshal(data, &lockFile); err != nil {
		return fmt.Errorf("failed to parse lock file %s: %w", path, err)
	}

	c.LockFile = &lockFile
	return nil
}

// expandEnvVarsInJSON expands environment variables in JSON string content
func expandEnvVarsInJSON(jsonContent string) string {
	return expandEnvVars(jsonContent)
}

// validateConfig validates the merged configuration
func validateConfig(cfg *Config) error {
	// Validate registries
	for name, url := range cfg.Registries {
		if err := validateRegistry(name, url, cfg.RegistryConfigs[name]); err != nil {
			return fmt.Errorf("registry '%s': %w", name, err)
		}
	}

	// Validate engines
	if err := validateEngines(cfg.Engines); err != nil {
		return fmt.Errorf("engines: %w", err)
	}

	// Validate channels
	if err := validateChannels(cfg.Channels); err != nil {
		return fmt.Errorf("channels: %w", err)
	}

	return nil
}

// validateRegistry validates a single registry configuration
func validateRegistry(name, url string, config map[string]string) error {
	if config == nil {
		return fmt.Errorf("missing configuration section [registries.%s]", name)
	}

	registryType, exists := config["type"]
	if !exists {
		return fmt.Errorf("missing required field 'type'")
	}

	// Validate registry type
	validTypes := []string{"git", "https", "s3", "gitlab", "local"}
	if !contains(validTypes, registryType) {
		return fmt.Errorf("unknown registry type '%s'. Supported types: %s", registryType, strings.Join(validTypes, ", "))
	}

	// Type-specific validation
	switch registryType {
	case "s3":
		if _, exists := config["region"]; !exists {
			return fmt.Errorf("missing required field 'region' for S3 registry")
		}
	case "git":
		if url == "" {
			return fmt.Errorf("missing registry URL for Git registry")
		}
		if !strings.HasPrefix(url, "https://") {
			return fmt.Errorf("Git registry URL must use HTTPS protocol")
		}
	case "https":
		if url == "" {
			return fmt.Errorf("missing registry URL for HTTPS registry")
		}
		if !strings.HasPrefix(url, "https://") {
			return fmt.Errorf("HTTPS registry URL must use HTTPS protocol")
		}
	case "gitlab":
		if url == "" {
			return fmt.Errorf("missing registry URL for GitLab registry")
		}
		if !strings.HasPrefix(url, "https://") {
			return fmt.Errorf("GitLab registry URL must use HTTPS protocol")
		}
	case "local":
		if url == "" {
			return fmt.Errorf("missing registry path for Local registry")
		}
	}

	return nil
}

// validateEngines validates the engines configuration
func validateEngines(engines map[string]string) error {
	if len(engines) == 0 {
		return nil // Engines are optional
	}

	// Validate ARM engine version if present
	if armVersion, exists := engines["arm"]; exists {
		if armVersion == "" {
			return fmt.Errorf("ARM engine version cannot be empty")
		}
		// Basic semver pattern validation
		if !regexp.MustCompile(`^[\^~>=<]?\d+\.\d+\.\d+`).MatchString(armVersion) {
			return fmt.Errorf("invalid ARM engine version format: %s", armVersion)
		}
	}

	return nil
}

// validateChannels validates the channels configuration
func validateChannels(channels map[string]ChannelConfig) error {
	for name, config := range channels {
		if len(config.Directories) == 0 {
			return fmt.Errorf("channel '%s' must have at least one directory", name)
		}
		for i, dir := range config.Directories {
			if dir == "" {
				return fmt.Errorf("channel '%s' directory %d cannot be empty", name, i)
			}
		}
	}
	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
