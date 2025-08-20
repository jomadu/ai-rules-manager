package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/max-dunn/ai-rules-manager/internal/version"
)

// Config represents the ARM configuration
type Config struct {
	Registries      map[string]string            // [registries] section
	RegistryConfigs map[string]map[string]string // [registries.name] sections
	TypeDefaults    map[string]map[string]string // [git], [s3], etc. sections
	NetworkConfig   map[string]string            // [network] section

	// JSON configuration
	Channels map[string]ChannelConfig          // channels from arm.json
	Rulesets map[string]map[string]RulesetSpec // rulesets from arm.json
	Engines  map[string]string                 // engines from arm.json
	LockFile *LockFile                         // arm.lock content

	// Cache configuration (loaded from INI sections)
	CacheConfig *CacheConfig // cache settings
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
	Engines  map[string]string                 `json:"engines"`
	Rulesets map[string]map[string]RulesetSpec `json:"rulesets"`
}

// ARMRCConfig represents the .armrc.json file structure
type ARMRCConfig struct {
	Registries map[string]RegistryConfig `json:"registries"`
	Channels   map[string]ChannelConfig  `json:"channels"`
	Cache      *CacheConfig              `json:"cache,omitempty"`
	Network    *NetworkConfig            `json:"network,omitempty"`
	Git        *TypeConfig               `json:"git,omitempty"`
	HTTPS      *TypeConfig               `json:"https,omitempty"`
	S3         *TypeConfig               `json:"s3,omitempty"`
	Gitlab     *TypeConfig               `json:"gitlab,omitempty"`
	Local      *TypeConfig               `json:"local,omitempty"`
}

// RegistryConfig represents a registry configuration
type RegistryConfig struct {
	URL        string `json:"url"`
	Type       string `json:"type"`
	Region     string `json:"region,omitempty"`
	AuthToken  string `json:"authToken,omitempty"`
	APIType    string `json:"apiType,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	Profile    string `json:"profile,omitempty"`
	Prefix     string `json:"prefix,omitempty"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Timeout                string `json:"timeout,omitempty"`
	RetryMaxAttempts       string `json:"retry.maxAttempts,omitempty"`
	RetryBackoffMultiplier string `json:"retry.backoffMultiplier,omitempty"`
	RetryMaxBackoff        string `json:"retry.maxBackoff,omitempty"`
}

// TypeConfig represents registry type defaults
type TypeConfig struct {
	Concurrency string `json:"concurrency,omitempty"`
	RateLimit   string `json:"rateLimit,omitempty"`
}

// LockFile represents the arm.lock file structure
type LockFile struct {
	Rulesets map[string]map[string]LockedRuleset `json:"rulesets"`
}

// LockedRuleset represents a locked ruleset entry
type LockedRuleset struct {
	Version  string   `json:"version"`
	Resolved string   `json:"resolved"`
	Patterns []string `json:"patterns,omitempty"` // Only for git registries
	Registry string   `json:"registry"`
	Type     string   `json:"type"`
	Region   string   `json:"region,omitempty"`
}

// Load loads the ARM configuration from files with hierarchical merging
func Load() (*Config, error) {
	// Load global configuration first
	globalCfg, err := loadConfigFromPaths(
		filepath.Join(os.Getenv("HOME"), ".arm", ".armrc.json"),
		filepath.Join(os.Getenv("HOME"), ".arm", "arm.json"),
		"arm.lock", // Lock file is always local
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Load local configuration
	localCfg, err := loadConfigFromPaths(".armrc.json", "arm.json", "arm.lock")
	if err != nil {
		return nil, fmt.Errorf("failed to load local config: %w", err)
	}

	// Merge configurations (local overrides global at key level)
	mergedCfg := mergeConfigs(globalCfg, localCfg)

	// Validate merged configuration
	if err := validateConfig(mergedCfg, globalCfg.TypeDefaults["cache"]); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return mergedCfg, nil
}

// loadConfigFromPaths loads configuration from specified file paths
func loadConfigFromPaths(armrcJSONPath, jsonPath, lockPath string) (*Config, error) {
	cfg := &Config{
		Registries:      make(map[string]string),
		RegistryConfigs: make(map[string]map[string]string),
		TypeDefaults:    make(map[string]map[string]string),
		NetworkConfig:   make(map[string]string),
		Channels:        make(map[string]ChannelConfig),
		Rulesets:        make(map[string]map[string]RulesetSpec),
		Engines:         make(map[string]string),
	}

	// Load .armrc.json file
	if err := cfg.loadARMRCJSON(armrcJSONPath, false); err != nil {
		return nil, fmt.Errorf("failed to load JSON config file %s: %w", armrcJSONPath, err)
	}

	// Load arm.json file
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

// mergeConfigs merges two configurations with local taking precedence at key level
func mergeConfigs(global, local *Config) *Config {
	merged := &Config{
		Registries:      make(map[string]string),
		RegistryConfigs: make(map[string]map[string]string),
		TypeDefaults:    make(map[string]map[string]string),
		NetworkConfig:   make(map[string]string),
		Channels:        make(map[string]ChannelConfig),
		Rulesets:        make(map[string]map[string]RulesetSpec),
		Engines:         make(map[string]string),
	}

	// Merge registries (key-level merge)
	mergeStringMaps(merged.Registries, global.Registries, local.Registries)

	// Merge registry configs (nested map merge)
	mergeNestedStringMaps(merged.RegistryConfigs, global.RegistryConfigs, local.RegistryConfigs)

	// Merge type defaults (nested map merge) - exclude cache settings (global only)
	mergeNestedStringMapsExcluding(merged.TypeDefaults, global.TypeDefaults, local.TypeDefaults, []string{"cache"})

	// Merge network config (key-level merge)
	mergeStringMaps(merged.NetworkConfig, global.NetworkConfig, local.NetworkConfig)

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
	mergeNestedStringMapsExcluding(dest, global, local, nil)
}

// mergeNestedStringMapsExcluding merges nested string maps excluding specified keys
func mergeNestedStringMapsExcluding(dest, global, local map[string]map[string]string, exclude []string) {
	// Copy global values first
	for k, v := range global {
		dest[k] = make(map[string]string)
		for kk, vv := range v {
			dest[k][kk] = vv
		}
	}
	// Merge with local values (key-level merge within each nested map)
	for k, v := range local {
		// Skip excluded keys
		if contains(exclude, k) {
			continue
		}
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

// loadARMRCJSON loads and parses an .armrc.json file
func (c *Config) loadARMRCJSON(path string, required bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if required {
			return fmt.Errorf("required JSON config file not found: %s", path)
		}
		return nil // Optional file doesn't exist
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read JSON config file %s: %w", path, err)
	}

	// Expand environment variables in JSON content
	expandedData := expandEnvVarsInJSON(string(data))

	var armrcConfig ARMRCConfig
	if err := json.Unmarshal([]byte(expandedData), &armrcConfig); err != nil {
		return fmt.Errorf("failed to parse JSON config file %s: %w", path, err)
	}

	// Map JSON structure to existing Config fields
	for name, regConfig := range armrcConfig.Registries {
		c.Registries[name] = regConfig.URL
		if c.RegistryConfigs[name] == nil {
			c.RegistryConfigs[name] = make(map[string]string)
		}
		c.RegistryConfigs[name]["type"] = regConfig.Type
		if regConfig.Region != "" {
			c.RegistryConfigs[name]["region"] = regConfig.Region
		}
		if regConfig.AuthToken != "" {
			c.RegistryConfigs[name]["authToken"] = regConfig.AuthToken
		}
		if regConfig.APIType != "" {
			c.RegistryConfigs[name]["apiType"] = regConfig.APIType
		}
		if regConfig.APIVersion != "" {
			c.RegistryConfigs[name]["apiVersion"] = regConfig.APIVersion
		}
		if regConfig.Profile != "" {
			c.RegistryConfigs[name]["profile"] = regConfig.Profile
		}
		if regConfig.Prefix != "" {
			c.RegistryConfigs[name]["prefix"] = regConfig.Prefix
		}
	}

	// Map channels
	for name, channelConfig := range armrcConfig.Channels {
		c.Channels[name] = channelConfig
	}

	// Map type defaults
	if armrcConfig.Git != nil {
		if c.TypeDefaults["git"] == nil {
			c.TypeDefaults["git"] = make(map[string]string)
		}
		if armrcConfig.Git.Concurrency != "" {
			c.TypeDefaults["git"]["concurrency"] = armrcConfig.Git.Concurrency
		}
		if armrcConfig.Git.RateLimit != "" {
			c.TypeDefaults["git"]["rateLimit"] = armrcConfig.Git.RateLimit
		}
	}
	if armrcConfig.HTTPS != nil {
		if c.TypeDefaults["https"] == nil {
			c.TypeDefaults["https"] = make(map[string]string)
		}
		if armrcConfig.HTTPS.Concurrency != "" {
			c.TypeDefaults["https"]["concurrency"] = armrcConfig.HTTPS.Concurrency
		}
		if armrcConfig.HTTPS.RateLimit != "" {
			c.TypeDefaults["https"]["rateLimit"] = armrcConfig.HTTPS.RateLimit
		}
	}
	if armrcConfig.S3 != nil {
		if c.TypeDefaults["s3"] == nil {
			c.TypeDefaults["s3"] = make(map[string]string)
		}
		if armrcConfig.S3.Concurrency != "" {
			c.TypeDefaults["s3"]["concurrency"] = armrcConfig.S3.Concurrency
		}
		if armrcConfig.S3.RateLimit != "" {
			c.TypeDefaults["s3"]["rateLimit"] = armrcConfig.S3.RateLimit
		}
	}
	if armrcConfig.Gitlab != nil {
		if c.TypeDefaults["gitlab"] == nil {
			c.TypeDefaults["gitlab"] = make(map[string]string)
		}
		if armrcConfig.Gitlab.Concurrency != "" {
			c.TypeDefaults["gitlab"]["concurrency"] = armrcConfig.Gitlab.Concurrency
		}
		if armrcConfig.Gitlab.RateLimit != "" {
			c.TypeDefaults["gitlab"]["rateLimit"] = armrcConfig.Gitlab.RateLimit
		}
	}
	if armrcConfig.Local != nil {
		if c.TypeDefaults["local"] == nil {
			c.TypeDefaults["local"] = make(map[string]string)
		}
		if armrcConfig.Local.Concurrency != "" {
			c.TypeDefaults["local"]["concurrency"] = armrcConfig.Local.Concurrency
		}
		if armrcConfig.Local.RateLimit != "" {
			c.TypeDefaults["local"]["rateLimit"] = armrcConfig.Local.RateLimit
		}
	}

	// Map network config
	if armrcConfig.Network != nil {
		if armrcConfig.Network.Timeout != "" {
			c.NetworkConfig["timeout"] = armrcConfig.Network.Timeout
		}
		if armrcConfig.Network.RetryMaxAttempts != "" {
			c.NetworkConfig["retry.maxAttempts"] = armrcConfig.Network.RetryMaxAttempts
		}
		if armrcConfig.Network.RetryBackoffMultiplier != "" {
			c.NetworkConfig["retry.backoffMultiplier"] = armrcConfig.Network.RetryBackoffMultiplier
		}
		if armrcConfig.Network.RetryMaxBackoff != "" {
			c.NetworkConfig["retry.maxBackoff"] = armrcConfig.Network.RetryMaxBackoff
		}
	}

	// Map cache config - store as strings for compatibility with existing LoadCacheConfig
	if armrcConfig.Cache != nil {
		if c.TypeDefaults["cache"] == nil {
			c.TypeDefaults["cache"] = make(map[string]string)
		}
		if armrcConfig.Cache.MaxSize != 0 {
			c.TypeDefaults["cache"]["maxSize"] = fmt.Sprintf("%d", armrcConfig.Cache.MaxSize)
		}
		if armrcConfig.Cache.TTL != 0 {
			c.TypeDefaults["cache"]["ttl"] = time.Duration(armrcConfig.Cache.TTL).String()
		}
		if armrcConfig.Cache.CleanupInterval != 0 {
			c.TypeDefaults["cache"]["cleanupInterval"] = time.Duration(armrcConfig.Cache.CleanupInterval).String()
		}
	}

	return nil
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
func validateConfig(cfg *Config, globalCacheSettings map[string]string) error {
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

	// Load cache configuration from global settings only (not overridable by local)
	cfg.CacheConfig = LoadCacheConfig(globalCacheSettings)

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
	validTypes := []string{"git", "git-local", "https", "s3", "gitlab", "local"}
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
	case "git-local":
		if url == "" {
			return fmt.Errorf("missing registry path for Git-Local registry")
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
			return fmt.Errorf("arm engine version cannot be empty")
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

// GenerateStubFiles generates stub configuration files if they don't exist
func GenerateStubFiles(global bool) error {
	var armrcPath, jsonPath string

	if global {
		homeDir := os.Getenv("HOME")
		armDir := filepath.Join(homeDir, ".arm")
		if err := os.MkdirAll(armDir, 0o755); err != nil {
			return fmt.Errorf("failed to create .arm directory: %w", err)
		}
		armrcPath = filepath.Join(armDir, ".armrc")
		jsonPath = filepath.Join(armDir, "arm.json")
	} else {
		armrcPath = ".armrc"
		jsonPath = "arm.json"
	}

	// Generate .armrc stub if it doesn't exist
	if _, err := os.Stat(armrcPath); os.IsNotExist(err) {
		if err := generateARMRCStub(armrcPath); err != nil {
			return fmt.Errorf("failed to generate .armrc stub: %w", err)
		}
	}

	// Generate arm.json stub if it doesn't exist
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		if err := generateARMJSONStub(jsonPath); err != nil {
			return fmt.Errorf("failed to generate arm.json stub: %w", err)
		}
	}

	return nil
}

// generateARMRCStub generates a stub .armrc file
func generateARMRCStub(path string) error {
	stubContent := `# ARM Configuration File
# Configure registries and default settings

[registries]
# Default registry used when no source is specified
# default = github.com/user/registry

# Named registries
# my-git-registry = https://github.com/user/repo
# my-s3-registry = my-bucket
# my-gitlab-registry = https://gitlab.example.com/projects/123
# my-https-registry = https://example.com/registry
# my-local-registry = /path/to/local/registry

# Required type configuration for all registries
# [registries.default]
# type = git

# [registries.my-git-registry]
# type = git
# authToken = $GITHUB_TOKEN  # optional, for API mode
# apiType = github           # optional, enables API mode
# apiVersion = 2022-11-28    # optional, API version

# [registries.my-s3-registry]
# type = s3
# region = us-east-1         # required for S3 registries
# profile = my-aws-profile   # optional, uses default AWS profile if omitted
# prefix = /registries/path  # optional prefix within bucket

# [registries.my-gitlab-registry]
# type = gitlab
# authToken = $GITLAB_TOKEN
# apiVersion = 4

# [registries.my-https-registry]
# type = https

# [registries.my-local-registry]
# type = local

# Type-based defaults (optional - ARM has built-in defaults)
# [git]
# concurrency = 1
# rateLimit = 10/minute

# [https]
# concurrency = 5
# rateLimit = 30/minute

# [s3]
# concurrency = 10
# rateLimit = 100/hour

# [gitlab]
# concurrency = 2
# rateLimit = 60/hour

# [local]
# concurrency = 20
# rateLimit = 1000/second

# Network configuration
# [network]
# timeout = 30
# retry.maxAttempts = 3
# retry.backoffMultiplier = 2.0
# retry.maxBackoff = 30

# Cache configuration (GLOBAL ONLY - configure in ~/.arm/.armrc, cannot be overridden by local .armrc)
# [cache]
# maxSize = 1GB                  # Max cache size (supports GB, MB, KB, or bytes)
# ttl = 24h                      # Time-to-live for cache entries
# cleanupInterval = 6h           # How often to run cleanup

# Channel configuration (where to install rulesets)
# [channels.cursor]
# directories = .cursor/rules

# [channels.q]
# directories = .amazonq/rules

# [channels.custom]
# directories = ./ai-rules,./shared-rules

`

	return os.WriteFile(path, []byte(stubContent), 0o600)
}

// generateARMJSONStub generates a stub arm.json file
func generateARMJSONStub(path string) error {
	// Get current ARM version (will be injected by build system)
	armVersion := GetCurrentARMVersion()

	stubContent := fmt.Sprintf(`{
  "engines": {
    "arm": "^%s"
  },
  "rulesets": {}
}
`, armVersion)

	return os.WriteFile(path, []byte(stubContent), 0o600)
}

// GetCurrentARMVersion returns the current ARM version
func GetCurrentARMVersion() string {
	currentVersion := version.GetVersion()
	// Clean up version string (remove 'v' prefix and git info if present)
	currentVersion = strings.TrimPrefix(currentVersion, "v")
	// Remove git commit info (e.g., "1.2.0-26-g2869e3f-dirty" -> "1.2.0")
	if idx := strings.Index(currentVersion, "-"); idx != -1 {
		currentVersion = currentVersion[:idx]
	}
	// Fallback to 1.0.0 if version is "dev" or empty
	if currentVersion == "dev" || currentVersion == "" {
		currentVersion = "1.0.0"
	}
	return currentVersion
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
