# Product Requirements Document: Content-Based Cache System

## Introduction/Overview

ARM's current cache system uses registry names as cache keys, creating conflicts when different projects use the same registry name for different repositories. This causes cache corruption, incorrect data serving, and unpredictable behavior across projects. The content-based cache system solves this by using SHA-256 hashes of registry URLs as cache keys, ensuring unique identification while maintaining cache sharing benefits and preserving the existing user experience.

## Goals

1. **Eliminate Cache Conflicts**: Prevent registry name collisions between projects using different repositories with same names
2. **Enable Safe Cache Sharing**: Allow projects with identical registry URLs to share cached data efficiently
3. **Maintain User Experience**: Preserve existing command-line interface without exposing internal hash complexity
4. **Provide Cache Transparency**: Enable debugging and cache management through human-readable mapping
5. **Ensure Clean Migration**: Implement seamless transition from name-based to content-based caching

## User Stories

- **As a developer working on multiple projects**, I want ARM to cache registries correctly so that Project A's "default" registry doesn't interfere with Project B's "default" registry when they point to different repositories.

- **As a team member sharing registry URLs**, I want ARM to reuse cached data when multiple projects use the same repository so that subsequent installs are faster and use less disk space.

- **As a developer debugging cache issues**, I want to understand which cached directories correspond to which registry URLs so that I can troubleshoot problems effectively.

- **As a system administrator managing disk space**, I want ARM to automatically deduplicate cache data across projects so that identical repositories don't consume multiple cache entries.

- **As a developer upgrading ARM**, I want the cache migration to happen automatically without losing existing cached data or requiring manual intervention.

## Functional Requirements

### 1. Content-Based Cache Keys
1.1. ARM must generate SHA-256 hashes of normalized registry URLs as cache directory names
1.2. ARM must normalize URLs by removing trailing slashes, converting to lowercase, and standardizing protocols
1.3. ARM must handle local paths by converting to absolute paths before hashing
1.4. ARM must use full SHA-256 hash (64 characters) to eliminate collision probability

### 2. Registry Mapping System
2.1. ARM must maintain a global registry mapping file at `~/.arm/cache/registry-map.json`
2.2. ARM must store rich metadata for each registry including URL, type, creation time, and last access time
2.3. ARM must update mapping file atomically to prevent corruption during concurrent operations
2.4. ARM must provide reverse lookup capability from hash to original URL for debugging

### 3. Cache Directory Structure
3.1. ARM must organize cache using hash-based directory names under `~/.arm/cache/registries/`
3.2. ARM must maintain existing subdirectory structure within each hashed registry directory
3.3. ARM must preserve all current cache file formats (versions.json, metadata.json, cache-info.json)
3.4. ARM must store original registry URL in cache-info.json for local reference

### 4. User Interface Preservation
4.1. ARM must continue accepting registry names in all commands without exposing hashes
4.2. ARM must resolve registry names to URLs using current project configuration
4.3. ARM must provide clear error messages when registry names cannot be resolved
4.4. ARM must maintain all existing command syntax and behavior

### 5. Cache Sharing and Deduplication
5.1. ARM must automatically share cache data between projects using identical registry URLs
5.2. ARM must update last access time when any project accesses shared cache data
5.3. ARM must handle concurrent access to shared cache directories safely
5.4. ARM must preserve cache data when one project removes a registry that others still use

### 6. Migration and Compatibility
6.1. ARM must detect existing name-based cache structure on first run after upgrade
6.2. ARM must clear existing cache and rebuild with new structure during migration
6.3. ARM must create registry mapping file during migration process
6.4. ARM must complete migration transparently without user intervention

## Non-Goals (Out of Scope)

- Migrating existing cached data to new structure (clean rebuild is acceptable)
- Supporting rollback to name-based cache system
- Providing cache import/export functionality
- Implementing cache encryption or security features

## Design Considerations

### Cache Directory Structure
```
~/.arm/cache/
├── registry-map.json           # Global registry mapping
├── registries/
│   ├── a1b2c3d4e5f6.../       # SHA-256 hash of registry URL
│   │   ├── repository/         # Git repository clones
│   │   ├── rulesets/
│   │   ├── metadata.json
│   │   ├── versions.json
│   │   └── cache-info.json
│   └── f6e5d4c3b2a1.../       # Another registry hash
└── temp/
```

### Registry Mapping File Format
```json
{
  "version": "1.0",
  "registries": {
    "a1b2c3d4e5f6789...": {
      "url": "https://github.com/user/repo",
      "type": "git",
      "created": "2024-01-15T10:30:00Z",
      "last_used": "2024-01-16T14:22:00Z"
    }
  }
}
```

### URL Normalization Rules
- Remove trailing slashes: `https://github.com/user/repo/` → `https://github.com/user/repo`
- Convert to lowercase: `HTTPS://GITHUB.COM/User/Repo` → `https://github.com/user/repo`
- Standardize protocols: `git@github.com:user/repo.git` → `https://github.com/user/repo`
- Resolve local paths: `./local/repo` → `/absolute/path/to/local/repo`

## Technical Considerations

### Hash Generation
- Use SHA-256 for cryptographic strength and collision resistance
- Apply normalization before hashing to ensure consistency
- Store both original and normalized URLs in mapping file
- Use full 64-character hash to eliminate collision probability

### Concurrent Access
- Use file locking for registry mapping file updates
- Implement atomic writes using temporary files and rename operations
- Handle race conditions during cache directory creation
- Ensure thread-safe access to shared cache directories

### Performance Impact
- Hash generation adds minimal overhead (microseconds)
- Mapping file lookup adds single JSON parse per operation
- Cache sharing reduces overall disk usage and download time
- Directory scanning for cleanup operations remains unchanged

### Error Handling
- Graceful degradation when mapping file is corrupted
- Clear error messages for hash generation failures
- Automatic mapping file recreation when missing
- Detailed logging for debugging cache operations

### Configuration Integration

**Cache Path Configuration:**
- Default cache location: `~/.arm/cache`
- Configurable via `.armrc`: `[cache] path = /custom/cache/path`
- Environment variable expansion supported: `path = $HOME/.custom/arm/cache`
- Tilde expansion supported: `path = ~/custom/cache`
- Write permission validation during configuration

**Cache Management Settings:**
```ini
# .armrc configuration
[cache]
path = ~/.arm/cache          # Cache directory location
maxSize = 1GB               # Maximum cache size (LRU eviction)
ttl = 3600                  # Cache TTL in seconds (1 hour default)
```

**Registry-Specific Cache Configuration:**
```ini
# Registry type defaults affect cache behavior
[git]
concurrency = 1             # Affects parallel cache operations
rateLimit = 10/minute       # Affects cache refresh frequency

[s3]
concurrency = 10
rateLimit = 100/hour

[network]
timeout = 30                # Network timeout for cache refresh
retry.maxAttempts = 3       # Retry attempts for failed cache operations
retry.backoffMultiplier = 2.0
retry.maxBackoff = 30
```

**Cache File Formats:**

**Enhanced cache-info.json:**
```json
{
  "registry_url": "https://github.com/user/repo",
  "registry_url_normalized": "https://github.com/user/repo",
  "registry_type": "git",
  "cache_key_hash": "a1b2c3d4e5f6789...",
  "last_accessed": "2024-01-15T10:30:00Z",
  "total_size_bytes": 1048576,
  "created": "2024-01-15T10:30:00Z"
}
```

**versions.json (unchanged format):**
```json
{
  "cached_at": "2024-01-15T10:30:00Z",
  "ttl_seconds": 3600,
  "rulesets": {
    "my-rules": ["1.0.0", "1.1.0", "1.2.0"],
    "python-rules": ["2.0.0", "2.1.0"]
  }
}
```

**metadata.json (unchanged format):**
```json
{
  "cached_at": "2024-01-15T10:30:00Z",
  "ttl_seconds": 3600,
  "rulesets": {
    "my-rules": {
      "description": "Python coding rules",
      "latest_version": "1.2.0"
    }
  }
}
```

## Success Metrics

### Primary Metrics
1. **Zero Cache Conflicts**: No instances of incorrect data served due to registry name collisions
2. **Cache Sharing Efficiency**: Identical repositories share cache data across projects (>95% deduplication rate)
3. **Migration Success Rate**: Successful cache migration on upgrade (>99% success rate)
4. **Performance Maintenance**: Cache operations complete within 110% of current performance

### Secondary Metrics
1. **Disk Space Efficiency**: Reduced total cache size due to deduplication
2. **User Experience**: No breaking changes to existing command syntax
3. **Error Rate**: Reduced cache-related errors and support requests
4. **Debug Capability**: Improved troubleshooting through mapping file transparency

## Open Questions

1. **Mapping File Size**: Should we implement cleanup for unused registry entries in the mapping file?
   **Resolved:** Yes, clean unused entries during `arm clean cache` operations

2. **Hash Display**: Should debugging commands show hash values alongside registry names?
   **Resolved:** Only in verbose mode to avoid cluttering normal output

3. **Backward Compatibility**: Should we provide a flag to temporarily use old cache structure?
   **Resolved:** No, clean migration is simpler and more reliable

4. **Cache Validation**: Should we validate cache integrity using checksums?
   **Resolved:** Not in initial implementation, rely on existing validation mechanisms
