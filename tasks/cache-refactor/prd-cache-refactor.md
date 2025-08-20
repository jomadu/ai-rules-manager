# Cache System Simplification PRD

## Overview

Simplify ARM's cache system by consolidating 5 manager classes into 1, eliminating mapping files, and implementing Git-specific version resolution using commit hashes.

## Goals

1. **Reduce complexity**: Consolidate cache management into single interface
2. **Eliminate redundancy**: Remove mapping files, use filesystem structure for discovery
3. **Git-optimized caching**: Use commit hashes for immutable cache keys
4. **Pattern-based storage**: Maintain pattern-specific caching with simplified hash calculation

## Current Problems

- **5 separate manager classes**: DefaultManager, RegistryMapper, RulesetMapper, MetadataManager, RulesetStorage
- **Redundant mapping files**: `registry-map.json` and `ruleset-map.json` duplicate filesystem structure
- **Complex URL normalization**: Registry-specific normalization adds unnecessary complexity
- **Inconsistent version handling**: Mix of user versions and resolved versions in cache keys

## Proposed Solution

### Simplified Cache Structure

```
cache/
├── config.json                                 # Global cache configuration
└── registries/
    ├── hash(git_registry_url + git)/           # Git registry cache
    │   ├── index.json                          # Git-specific metadata
    │   ├── repository/                         # Git clone
    │   └── rulesets/
    │       └── hash(ruleset_patterns)/                 # Pattern-specific cache
    │           ├── abc123def456.../            # Commit hash directories
    │           ├── def456abc123.../
    │           └── 789abc123def.../
    └── hash(s3_registry_url + s3)/             # Non-Git registry cache
        ├── index.json                          # Simple metadata
        └── rulesets/
            └── hash(ruleset_name)/             # Ruleset-specific cache
                ├── 1.0.0/                      # Semantic version directories
                ├── 1.1.0/
                └── 2.0.0/
```

#### Cache Configuration Structure

```json
{
  "version": "1.0",
  "created_on": "2024-01-15T10:30:00Z",
  "last_updated_on": "2024-01-15T10:30:00Z",
  "ttl_hours": 24,
  "max_size_mb": 1024,
  "cleanup_enabled": true
}
```

#### Git-Based Registry index.json Structure

```json
{
  "created_on": "2024-01-15T10:30:00Z",
  "last_updated_on": "2024-01-15T10:30:00Z",
  "last_accessed_on": "2024-01-15T10:30:00Z",
  "normalized_registry_url": "https://github.com/user/repo",
  "normalized_registry_type": "git",
  "rulesets": {
    "xyz789abc123...": {
      "normalized_ruleset_patterns": ["*.md", "rules/**"],
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "abc123def456...": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    }
  }
}
```

#### Non-Git-Based Registry index.json Structure

```json
{
  "created_on": "2024-01-15T10:30:00Z",
  "last_updated_on": "2024-01-15T10:30:00Z",
  "last_accessed_on": "2024-01-15T10:30:00Z",
  "normalized_registry_url": "s3://my-bucket/rules",
  "normalized_registry_type": "s3",
  "rulesets": {
    "def456abc789...": {
      "normalized_ruleset_name": "power-up-rules",
      "created_on": "2024-01-15T10:30:00Z",
      "last_updated_on": "2024-01-15T10:30:00Z",
      "last_accessed_on": "2024-01-15T10:30:00Z",
      "versions": {
        "1.0.0": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        },
        "1.1.0": {
          "created_on": "2024-01-15T10:30:00Z",
          "last_updated_on": "2024-01-15T10:30:00Z",
          "last_accessed_on": "2024-01-15T10:30:00Z"
        }
      }
    }
  }
}
```

### Cache Manager Interface (Strategy Pattern)

```go
// Simplified base interface for all registry cache managers
type RegistryCacheManager interface {
    Store(registryURL string, identifier []string, version string, files map[string][]byte) error
    Get(registryURL string, identifier []string, version string) (map[string][]byte, error)
    GetPath(registryURL string, identifier []string, version string) (string, error)
    IsValid(registryURL string, ttl time.Duration) (bool, error)
    Cleanup(ttl time.Duration, maxSize int64) error
}

// Git-specific registry cache manager
type GitRegistryCacheManager interface {
    RegistryCacheManager

    // Git-specific operations
    StoreRuleset(registryURL string, patterns []string, commitHash string, files map[string][]byte) error
    GetRuleset(registryURL string, patterns []string, commitHash string) (map[string][]byte, error)
    GetRepositoryPath(registryURL string) (string, error)
}

// Non-Git registry cache manager
type RulesetRegistryCacheManager interface {
    RegistryCacheManager

    // Non-Git-specific operations
    StoreRuleset(registryURL, rulesetName, version string, files map[string][]byte) error
    GetRuleset(registryURL, rulesetName, version string) (map[string][]byte, error)
}

// Factory functions for creating registry managers
func NewGitRegistryCacheManager() GitRegistryCacheManager
func NewS3RegistryCacheManager() RulesetRegistryCacheManager
func NewHTTPSRegistryCacheManager() RulesetRegistryCacheManager

// Registry index structures
type RegistryIndex struct {
    CreatedOn            string                    `json:"created_on"`
    LastUpdatedOn        string                    `json:"last_updated_on"`
    LastAccessedOn       string                    `json:"last_accessed_on"`
    NormalizedRegistryURL string                   `json:"normalized_registry_url"`
    NormalizedRegistryType string                  `json:"normalized_registry_type"`
    Rulesets             map[string]*RulesetCache `json:"rulesets"`
}

type RulesetCache struct {
    // Git registries use patterns, non-Git use ruleset name
    NormalizedRulesetPatterns []string                   `json:"normalized_ruleset_patterns,omitempty"`
    NormalizedRulesetName     string                     `json:"normalized_ruleset_name,omitempty"`
    CreatedOn                 string                     `json:"created_on"`
    LastUpdatedOn             string                     `json:"last_updated_on"`
    LastAccessedOn            string                     `json:"last_accessed_on"`
    Versions                  map[string]*VersionCache  `json:"versions"`
}

type VersionCache struct {
    CreatedOn      string `json:"created_on"`
    LastUpdatedOn  string `json:"last_updated_on"`
    LastAccessedOn string `json:"last_accessed_on"`
}

// Cache configuration structure
type CacheConfig struct {
    Version              string `json:"version"`
    CreatedOn            string `json:"created_on"`
    LastUpdatedOn        string `json:"last_updated_on"`
    TTLHours             int    `json:"ttl_hours"`
    MaxSizeMB            int    `json:"max_size_mb"`
    CleanupEnabled       bool   `json:"cleanup_enabled"`
}
```

## Functional Requirements

### 1. Git Registry Version Resolution

1.1. **Commit hash directories**: Git registries MUST use full commit hashes as version directory names

1.2. **Fresh resolution**: Git versions (branches, tags, latest) MUST be resolved to commit hashes at request time

1.3. **No resolution caching**: System MUST NOT cache branch/tag to commit hash mappings

1.4. **Immutable cache keys**: Only resolved commit hashes MUST be used as cache directory names

### 2. Caching Strategy

2.1. **Git pattern-based caching**: Git registries MUST use `hash(patterns)` for cache keys to enable cross-project sharing

2.2. **Non-Git ruleset-based caching**: Non-Git registries MUST use `hash(ruleset_name)` for cache keys

2.3. **Empty pattern handling**: Git registries with empty patterns `[]` MUST use `hash("__EMPTY__")` to avoid collisions

### 3. Cache Resolution Flow

3.1. **Git resolution flow**: User requests "main" → Git resolve to "abc123..." → Check cache for `hash(patterns)/abc123.../` (branch/tag resolution is always fresh, commit-based cache entries are persistent)

3.2. **Non-Git direct lookup**: User requests "1.0.0" → Check cache for `hash(ruleset_name)/1.0.0/`

3.3. **No version mapping**: System MUST NOT maintain user-version to resolved-version mappings

### 4. Non-Git Registry Support

4.1. **Semantic versions**: Non-Git registries MUST use semantic versions as directory names

4.2. **No repository directory**: Non-Git registries MUST NOT have a `repository/` subdirectory

4.3. **Direct version lookup**: Non-Git registries MUST use version strings directly without resolution

### 5. Migration Strategy

5.1. **Fresh start**: System MUST ignore existing complex cache structure

5.2. **Cache rebuild**: System MUST rebuild cache on first access after migration

## Non-Goals

- Backward compatibility with existing cache structure
- Complex URL normalization beyond basic cleanup
- Mapping file migration or preservation
- Support for multiple cache versions simultaneously

## Success Metrics

- **Code reduction**: Reduce cache-related code by 60%+
- **File count reduction**: Eliminate 2 mapping files per registry
- **Performance**: Maintain or improve cache hit rates
- **Reliability**: Zero cache corruption issues during migration

## Technical Considerations

### Implementation Priority

1. **Replace existing cache entirely** - implement new cache and update all callers
2. **Single manager consolidation** - merge 5 classes into 1
3. **Git-specific optimizations** - commit hash directories and version mappings
4. **Pattern-based improvements** - simplified hash calculation

### Edge Cases

- **Hash collisions**: Accept SHA-256 collision risk as negligible
- **Long commit hashes**: Use full 40-character SHA for guaranteed uniqueness
- **Branch deletion**: Return error when requested branch/tag no longer exists
- **Network failures**: Graceful degradation when Git remote is unavailable

## Implementation Decisions

1. **Cleanup strategy**: TTL-based cleanup using last accessed time only
2. **Directory structure**: Pure hash-based paths for maximum performance
3. **Hash algorithm**: SHA-256 for both pattern and URL hashing (consistent security)
4. **Configuration**: Load once at startup, require restart for changes
5. **Error handling**: Categorized errors - cache corruption (rebuild), network failures (retry), invalid input (fail fast)
6. **Concurrent access**: Directory-based atomic locking with 30-second timeout and stale lock cleanup
7. **Cache validation**: Basic validation of config.json and index.json on startup
8. **Initialization**: Auto-create cache directory and config with defaults
9. **URL normalization**: Minimal - trim whitespace and ensure consistent protocol
10. **File permissions**: Use system defaults (755/644)

### Registry Locking Implementation

```go
type RegistryLock struct {
    lockPath string
    acquired bool
}

func (r *RegistryLock) Acquire(timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if err := os.Mkdir(r.lockPath, 0755); err == nil {
            r.acquired = true
            return nil
        }
        // Check for stale locks (>5 minutes old)
        if info, err := os.Stat(r.lockPath); err == nil {
            if time.Since(info.ModTime()) > 5*time.Minute {
                os.RemoveAll(r.lockPath) // Remove stale lock
            }
        }
        time.Sleep(50 * time.Millisecond)
    }
    return fmt.Errorf("failed to acquire lock within %v", timeout)
}

func (r *RegistryLock) Release() error {
    if r.acquired {
        r.acquired = false
        return os.RemoveAll(r.lockPath)
    }
    return nil
}
```
