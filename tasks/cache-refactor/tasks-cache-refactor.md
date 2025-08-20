## Relevant Files

- `internal/cache/manager.go` - New unified cache manager interface and implementation
- `internal/cache/manager_test.go` - Unit tests for cache manager
- `internal/cache/git_cache.go` - Git-specific cache manager implementation
- `internal/cache/git_cache_test.go` - Unit tests for Git cache manager
- `internal/cache/ruleset_cache.go` - Non-Git registry cache manager implementation
- `internal/cache/ruleset_cache_test.go` - Unit tests for non-Git cache manager
- `internal/cache/config.go` - Cache configuration structures and validation
- `internal/cache/config_test.go` - Unit tests for cache configuration
- `internal/cache/lock.go` - Registry locking implementation
- `internal/cache/lock_test.go` - Unit tests for registry locking

### Notes

- Unit tests should be placed alongside the code files they are testing
- Use `go test ./internal/cache/...` to run all cache-related tests
- The refactor removed existing files: `metadata.go`, `registry_map.go`, `ruleset_map.go`, `ruleset_storage.go`, `url_normalizer.go`
- Integration tests updated to work with new cache structure (index.json instead of ruleset-map.json)
- Cache management commands not implemented in CLI (cache functionality is internal)

**✅ ALL TASKS COMPLETED**

## Tasks

- [x] 1.0 Create New Cache Manager Interfaces and Structures
  - [x] 1.1 Define RegistryCacheManager interface, Git/Non-Git specific interfaces, and factory functions
  - [x] 1.2 Create cache index structures (RegistryIndex, RulesetCache, VersionCache, CacheConfig)
- [x] 2.0 Implement Git Registry Cache Manager with Tests
  - [x] 2.1 Implement GitRegistryCacheManager with commit hash directories and pattern-based cache keys
  - [x] 2.2 Add StoreRuleset, GetRuleset, and GetRepositoryPath methods for Git registries
  - [x] 2.3 Create comprehensive unit tests for Git cache operations
- [x] 3.0 Implement Non-Git Registry Cache Manager with Tests
  - [x] 3.1 Implement RulesetRegistryCacheManager with semantic version directories and ruleset name-based cache keys
  - [x] 3.2 Add StoreRuleset and GetRuleset methods without repository subdirectory logic
  - [x] 3.3 Create comprehensive unit tests for non-Git cache operations
- [x] 4.0 Implement Cache Configuration and Locking
  - [x] 4.1 Add registry locking with timeout, stale lock cleanup, and cache initialization
  - [x] 4.2 Implement TTL-based cleanup and cache size monitoring with cleanup
- [x] 5.0 Replace Existing Cache Implementation
  - [x] 5.1 Update all registry implementations and remove deprecated cache files
  - [x] 5.2 Update installer, orchestrator, and migrate cache validation logic
  - [x] 5.3 Update integration tests to work with new cache system
