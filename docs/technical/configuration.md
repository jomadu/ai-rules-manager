# Configuration System

Technical specification for ARM's hierarchical configuration system with INI and JSON file support.

## Overview

ARM uses a multi-file, hierarchical configuration system that supports:
- **Global and local configuration** with key-level merging
- **Multiple file formats** (INI for system settings, JSON for project data)
- **Environment variable expansion** with secure defaults
- **Validation and type checking** with helpful error messages

## Configuration Architecture

### File Types and Purposes

| File | Format | Purpose | Location |
|------|--------|---------|----------|
| `.armrc` | INI | Registries, channels, network, cache settings | Global/Local |
| `arm.json` | JSON | Rulesets, engines | Global/Local |
| `arm.lock` | JSON | Locked versions, patterns, metadata | Local only |

### Hierarchy and Merging

```
1. Global Configuration (~/.arm/)
   ├── .armrc          # Global system settings
   └── arm.json        # Global project defaults

2. Local Configuration (project root)
   ├── .armrc          # Project-specific overrides
   ├── arm.json        # Project channels and rulesets
   └── arm.lock        # Locked versions (auto-generated)

3. Environment Variables
   └── ARM_*, GITHUB_TOKEN, AWS_*, etc.

4. Command-line Flags
   └── --global, --dry-run, etc.
```

## Configuration Loading

### Loading Process
```go
func Load() (*Config, error) {
    // 1. Load global configuration
    globalCfg, err := loadConfigFromPaths(
        filepath.Join(os.Getenv("HOME"), ".arm", ".armrc"),
        filepath.Join(os.Getenv("HOME"), ".arm", "arm.json"),
        "", // No global lock file
    )

    // 2. Load local configuration
    localCfg, err := loadConfigFromPaths(".armrc", "arm.json", "arm.lock")

    // 3. Merge configurations (local overrides global)
    mergedCfg := mergeConfigs(globalCfg, localCfg)

    // 4. Validate merged configuration
    return validateConfig(mergedCfg)
}
```

### Merging Strategy

#### Key-Level Merging
Local configuration overrides global at the individual key level:

```ini
# Global ~/.arm/.armrc
[registries]
default = https://github.com/global/repo
team = https://github.com/team/repo

[registries.default]
type = git
authToken = $GLOBAL_TOKEN

# Local .armrc
[registries]
default = https://github.com/local/repo  # Overrides global

[registries.default]
type = git                               # Keeps global type
authToken = $LOCAL_TOKEN                 # Overrides global token
```

**Result**: Local `default` URL and token, but inherits other global registries.

#### Nested Map Merging
Registry configurations merge at the individual setting level:

```go
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
            dest[k][kk] = vv  // Local overrides global
        }
    }
}
```

## INI Configuration (.armrc)

### File Structure
```ini
# Registry definitions
[registries]
default = https://github.com/org/rules
s3-prod = production-bucket

# Registry-specific configuration
[registries.default]
type = git
authToken = $GITHUB_TOKEN
apiType = github
apiVersion = 2022-11-28

[registries.s3-prod]
type = s3
region = us-east-1
profile = production

# Channel configuration
[channels.cursor]
directories = .cursor/rules

[channels.q]
directories = .amazonq/rules

[channels.both]
directories = .cursor/rules,.amazonq/rules

# Registry type defaults
[git]
concurrency = 1
rateLimit = 10/minute

[s3]
concurrency = 10
rateLimit = 100/hour

# Network configuration
[network]
timeout = 30
retry.maxAttempts = 3
retry.backoffMultiplier = 2.0
retry.maxBackoff = 30

# Cache configuration
[cache]
path = $HOME/.arm/cache
maxSize = 1073741824
ttl = 24h
cleanupInterval = 6h
```

### INI Processing

#### Section Processing
```go
func (c *Config) processSection(section *ini.Section) error {
    sectionName := section.Name()

    // Handle nested sections like [registries.my-registry]
    if strings.Contains(sectionName, ".") {
        parts := strings.SplitN(sectionName, ".", 2)
        if parts[0] == "registries" {
            return c.processRegistryConfig(parts[1], section)
        }
        if parts[0] == "channels" {
            return c.processChannelConfig(parts[1], section)
        }
        return fmt.Errorf("unsupported nested section: %s", sectionName)
    }

    // Handle top-level sections
    switch sectionName {
    case "registries":
        return c.processRegistries(section)
    case "git", "https", "s3", "gitlab", "local", "cache":
        return c.processTypeDefaults(sectionName, section)
    case "network":
        return c.processNetworkConfig(section)
    default:
        return fmt.Errorf("unknown section: %s", sectionName)
    }
}
```

#### Environment Variable Expansion
```go
func expandEnvVars(s string) string {
    // Pattern matches $VAR and ${VAR}
    pattern := regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

    return pattern.ReplaceAllStringFunc(s, func(match string) string {
        var varName string
        if strings.HasPrefix(match, "${") {
            varName = match[2 : len(match)-1]  // ${VAR} format
        } else {
            varName = match[1:]                // $VAR format
        }
        return os.Getenv(varName)
    })
}
```

## JSON Configuration (arm.json)

### File Structure
```json
{
  "engines": {
    "arm": "^1.0.0"
  },
  "rulesets": {
    "default": {
      "coding-standards": {
        "version": "^1.0.0",
        "patterns": ["rules/*.md", "guidelines/*.md"]
      },
      "security-rules": {
        "version": "latest"
      }
    },
    "s3-prod": {
      "team-standards": {
        "version": ">=2.0.0"
      }
    }
  }
}
```

### JSON Processing
```go
func (c *Config) loadARMJSON(path string, required bool) error {
    // Read and expand environment variables
    data, err := os.ReadFile(path)
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
```

## Lock File (arm.lock)

### Purpose
- **Version Locking**: Record exact versions installed
- **Integrity**: Store checksums and metadata
- **Reproducibility**: Ensure consistent installations
- **Audit Trail**: Track what was installed when

### Structure
```json
{
  "rulesets": {
    "default": {
      "coding-standards": {
        "version": "1.2.3",
        "resolved": "abc123def456...",
        "patterns": ["rules/*.md", "guidelines/*.md"],
        "registry": "default",
        "type": "git",
        "installed": "2024-01-15T10:30:00Z"
      }
    },
    "s3-prod": {
      "team-standards": {
        "version": "2.1.0",
        "resolved": "s3://bucket/team-standards/v2.1.0/",
        "registry": "s3-prod",
        "type": "s3",
        "region": "us-east-1",
        "installed": "2024-01-15T10:35:00Z"
      }
    }
  }
}
```

### Lock File Management
```go
type LockFile struct {
    Rulesets map[string]map[string]LockedRuleset `json:"rulesets"`
}

type LockedRuleset struct {
    Version   string   `json:"version"`    // Installed version
    Resolved  string   `json:"resolved"`   // Resolved identifier (commit hash, S3 path, etc.)
    Patterns  []string `json:"patterns,omitempty"` // Patterns for Git registries
    Registry  string   `json:"registry"`   // Registry name
    Type      string   `json:"type"`       // Registry type
    Region    string   `json:"region,omitempty"` // AWS region for S3
    Installed string   `json:"installed"`  // Installation timestamp
}
```

## Configuration Validation

### Registry Validation
```go
func validateRegistry(name, url string, config map[string]string) error {
    if config == nil {
        return fmt.Errorf("missing configuration section [registries.%s]", name)
    }

    registryType, exists := config["type"]
    if !exists {
        return fmt.Errorf("missing required field 'type'")
    }

    // Type-specific validation
    switch registryType {
    case "git":
        if url == "" {
            return fmt.Errorf("missing registry URL for Git registry")
        }
        if !strings.HasPrefix(url, "https://") {
            return fmt.Errorf("Git registry URL must use HTTPS protocol")
        }
    case "s3":
        if _, exists := config["region"]; !exists {
            return fmt.Errorf("missing required field 'region' for S3 registry")
        }
    // ... other types
    }

    return nil
}
```

### Engine Validation
```go
func validateEngines(engines map[string]string) error {
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
```

### Channel Validation
```go
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
```

## Cache Configuration

### Cache Settings
```go
type CacheConfig struct {
    Path                string        // Cache root directory
    MaxSize             int64         // Maximum cache size in bytes
    TTL                 time.Duration // Default TTL for cache entries
    CleanupInterval     time.Duration // How often to run cleanup
    RegistryMetadataTTL time.Duration // TTL for registry metadata
    RulesetMetadataTTL  time.Duration // TTL for ruleset metadata
    ContentTTL          time.Duration // TTL for actual content
}
```

### Loading Cache Configuration
```go
func (c *Config) LoadCacheConfig() *CacheConfig {
    cacheConfig := &CacheConfig{
        Path:                expandEnvVars(c.TypeDefaults["cache"]["path"]),
        MaxSize:             parseSizeOrDefault(c.TypeDefaults["cache"]["maxSize"], 1<<30), // 1GB
        TTL:                 parseDurationOrDefault(c.TypeDefaults["cache"]["ttl"], 24*time.Hour),
        CleanupInterval:     parseDurationOrDefault(c.TypeDefaults["cache"]["cleanupInterval"], 6*time.Hour),
        RegistryMetadataTTL: parseDurationOrDefault(c.TypeDefaults["cache"]["registryMetadataTTL"], time.Hour),
        RulesetMetadataTTL:  parseDurationOrDefault(c.TypeDefaults["cache"]["rulesetMetadataTTL"], 6*time.Hour),
        ContentTTL:          parseDurationOrDefault(c.TypeDefaults["cache"]["contentTTL"], 7*24*time.Hour),
    }

    // Apply defaults if not configured
    if cacheConfig.Path == "" {
        cacheConfig.Path = filepath.Join(os.Getenv("HOME"), ".arm", "cache")
    }

    return cacheConfig
}
```

## Configuration Generation

### Stub File Generation
```go
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
```

## Security Considerations

### Token Storage
- **Environment Variables**: Preferred method for sensitive data
- **File Permissions**: Configuration files use 0600 permissions
- **No Plaintext**: Tokens never stored in configuration files

### Path Security
- **Path Validation**: Prevent directory traversal attacks
- **Permission Checks**: Validate write permissions before operations
- **Expansion Safety**: Secure environment variable expansion

### Network Security
- **HTTPS Only**: Remote registries must use HTTPS
- **Certificate Validation**: Full TLS certificate verification
- **Timeout Handling**: Prevent hanging network operations

## Performance Characteristics

### Configuration Loading
- **Lazy Loading**: Configuration loaded only when needed
- **Caching**: Parsed configuration cached in memory
- **Validation**: Early validation prevents runtime errors

### Memory Usage
- **Efficient Parsing**: INI and JSON parsers with minimal memory overhead
- **Selective Loading**: Only load required configuration sections
- **Garbage Collection**: Proper cleanup of temporary objects

### File I/O
- **Atomic Writes**: Configuration updates are atomic
- **Backup Creation**: Automatic backup before modifications
- **Error Recovery**: Rollback on write failures
