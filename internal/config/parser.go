package config

import (
	"os"
	"regexp"
	"strings"

	"gopkg.in/ini.v1"
)

// ARMConfig represents the parsed .armrc configuration
type ARMConfig struct {
	Sources map[string]Source `ini:"sources"`
	Cache   CacheConfig       `ini:"cache"`
}

// Source represents a registry source configuration
type Source struct {
	URL       string `ini:"-"`
	AuthToken string `ini:"authToken"`
	Timeout   string `ini:"timeout"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Location string `ini:"location"`
	MaxSize  string `ini:"maxSize"`
}

// ParseFile parses an .armrc file and returns the configuration
func ParseFile(path string) (*ARMConfig, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}

	config := &ARMConfig{
		Sources: make(map[string]Source),
	}

	// Parse sources section
	if sourcesSection := cfg.Section("sources"); sourcesSection != nil {
		for _, key := range sourcesSection.Keys() {
			source := Source{URL: key.Value()}
			config.Sources[key.Name()] = source
		}
	}

	// Parse source-specific sections
	for sourceName := range config.Sources {
		sectionName := "sources." + sourceName
		if section := cfg.Section(sectionName); section != nil {
			source := config.Sources[sourceName]
			if authToken := section.Key("authToken"); authToken != nil {
				source.AuthToken = authToken.Value()
			}
			if timeout := section.Key("timeout"); timeout != nil {
				source.Timeout = timeout.Value()
			}
			config.Sources[sourceName] = source
		}
	}

	// Parse cache section
	if cacheSection := cfg.Section("cache"); cacheSection != nil {
		if location := cacheSection.Key("location"); location != nil {
			config.Cache.Location = location.Value()
		}
		if maxSize := cacheSection.Key("maxSize"); maxSize != nil {
			config.Cache.MaxSize = maxSize.Value()
		}
	}

	// Substitute environment variables
	substituteEnvVars(config)

	return config, nil
}

// substituteEnvVars replaces environment variable references in config values
func substituteEnvVars(config *ARMConfig) {
	// Regex patterns for $VAR and ${VAR}
	envVarPattern := regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

	// Substitute in sources
	for name, source := range config.Sources {
		source.URL = substituteString(source.URL, envVarPattern)
		source.AuthToken = substituteString(source.AuthToken, envVarPattern)
		source.Timeout = substituteString(source.Timeout, envVarPattern)
		config.Sources[name] = source
	}

	// Substitute in cache config
	config.Cache.Location = substituteString(config.Cache.Location, envVarPattern)
	config.Cache.MaxSize = substituteString(config.Cache.MaxSize, envVarPattern)
}

// substituteString replaces environment variable references in a string
func substituteString(s string, pattern *regexp.Regexp) string {
	return pattern.ReplaceAllStringFunc(s, func(match string) string {
		var varName string
		if strings.HasPrefix(match, "${") {
			// ${VAR} format
			varName = match[2 : len(match)-1]
		} else {
			// $VAR format
			varName = match[1:]
		}
		return os.Getenv(varName)
	})
}
