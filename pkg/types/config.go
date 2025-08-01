package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// RegistryConfig represents the .armrc configuration
type RegistryConfig struct {
	Sources map[string]RegistrySource `ini:"sources"`
}

// RegistrySource represents a registry source configuration
type RegistrySource struct {
	URL       string `ini:"url"`
	AuthToken string `ini:"authToken"`
}

// LoadRegistryConfig loads configuration from .armrc files
func LoadRegistryConfig() (*RegistryConfig, error) {
	config := &RegistryConfig{
		Sources: make(map[string]RegistrySource),
	}

	// Set default registry
	config.Sources["default"] = RegistrySource{
		URL: "https://registry.armjs.org/",
	}

	// Load user-level config
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userConfigPath := filepath.Join(homeDir, ".armrc")
		if err := loadConfigFile(config, userConfigPath); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load user config: %w", err)
		}
	}

	// Load project-level config (overrides user config)
	if err := loadConfigFile(config, ".armrc"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}

	return config, nil
}

// loadConfigFile loads a single .armrc file
func loadConfigFile(config *RegistryConfig, path string) error {
	cfg, err := ini.Load(path)
	if err != nil {
		return err
	}

	sourcesSection := cfg.Section("sources")
	for _, key := range sourcesSection.Keys() {
		name := key.Name()
		url := key.Value()

		// Handle environment variable substitution
		url = os.ExpandEnv(url)

		source := RegistrySource{URL: url}

		// Check for auth token in separate section
		authSection := cfg.Section(fmt.Sprintf("sources.%s", name))
		if authSection != nil {
			if authKey := authSection.Key("authToken"); authKey != nil {
				source.AuthToken = os.ExpandEnv(authKey.Value())
			}
		}

		config.Sources[name] = source
	}

	return nil
}

// GetRegistryURL returns the registry URL for a given source name
func (c *RegistryConfig) GetRegistryURL(sourceName string) (string, error) {
	source, exists := c.Sources[sourceName]
	if !exists {
		return "", fmt.Errorf("registry source %s not found", sourceName)
	}
	return strings.TrimSuffix(source.URL, "/"), nil
}

// GetAuthToken returns the auth token for a given source name
func (c *RegistryConfig) GetAuthToken(sourceName string) string {
	source, exists := c.Sources[sourceName]
	if !exists {
		return ""
	}
	return source.AuthToken
}

// ResolveRulesetSource determines which registry source to use for a ruleset
func (c *RegistryConfig) ResolveRulesetSource(rulesetName string) string {
	org, _ := ParseRulesetName(rulesetName)
	if org != "" {
		// Check if we have a specific source for this org
		if _, exists := c.Sources[org]; exists {
			return org
		}
	}
	return "default"
}

// BuildDownloadURL constructs the download URL for a ruleset
func (c *RegistryConfig) BuildDownloadURL(rulesetName, version string) (string, error) {
	sourceName := c.ResolveRulesetSource(rulesetName)
	baseURL, err := c.GetRegistryURL(sourceName)
	if err != nil {
		return "", err
	}

	org, pkg := ParseRulesetName(rulesetName)
	if org != "" {
		// Scoped package: @org/package
		return fmt.Sprintf("%s/@%s/%s/%s/package.tgz", baseURL, org, pkg, version), nil
	}
	// Unscoped package
	return fmt.Sprintf("%s/%s/%s/package.tgz", baseURL, pkg, version), nil
}
