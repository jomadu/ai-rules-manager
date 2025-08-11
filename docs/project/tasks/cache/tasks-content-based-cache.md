# Content-Based Cache System Implementation Tasks

Based on the PRD analysis and current codebase assessment, I've identified the main tasks required to implement the content-based cache system. The current ARM implementation doesn't have explicit caching yet, so we'll be building this from scratch.

## Relevant Files

- `internal/cache/manager.go` - Core cache management with content-based keys (CREATED)
- `internal/cache/manager_test.go` - Unit tests for cache manager (CREATED)
- `internal/cache/registry_map.go` - Registry mapping file management
- `internal/cache/registry_map_test.go` - Unit tests for registry mapping
- `internal/cache/url_normalizer.go` - URL normalization utilities
- `internal/cache/url_normalizer_test.go` - Unit tests for URL normalization
- `internal/config/cache.go` - Cache configuration management
- `internal/config/cache_test.go` - Unit tests for cache configuration
- `internal/registry/base_git_registry.go` - Updated to use content-based cache
- `internal/registry/git.go` - Updated to use content-based cache
- `internal/registry/factory.go` - Updated to inject cache manager
- `internal/cli/commands.go` - Updated clean command for cache management

### Notes

- Unit tests should be placed alongside the code files they are testing
- Use `go test ./internal/cache/...` to run cache-related tests
- Integration tests will validate end-to-end cache behavior

## Tasks

- [x] 1.0 Create Core Cache Infrastructure
  - [x] 1.1 Create cache package with manager interface and basic structure
  - [x] 1.2 Implement URL normalization utilities with comprehensive rules
  - [x] 1.3 Implement SHA-256 hash generation for cache keys
  - [x] 1.4 Create cache directory structure management
- [x] 2.0 Implement Registry Mapping System
  - [x] 2.1 Create registry mapping file structure and JSON schema
  - [x] 2.2 Implement atomic file operations for mapping updates
  - [x] 2.3 Add reverse lookup functionality (hash to URL)
  - [x] 2.4 Implement mapping file validation and recovery
- [x] 3.0 Integrate Cache with Registry System
  - [x] 3.1 Update registry factory to inject cache manager
  - [x] 3.2 Modify Git registry to use content-based cache paths
  - [x] 3.3 Update cache-info.json format with new metadata
  - [x] 3.4 Implement cache sharing logic for identical URLs
- [x] 4.0 Add Cache Configuration Support
  - [x] 4.1 Extend config system with cache settings (FIXED: enabled setting now actually works)
  - [x] 4.2 Add cache path configuration with environment variable expansion
  - [x] 4.3 Implement cache size, TTL, and cleanup interval configuration
  - [x] 4.4 Add registry-specific cache settings
- [x] 5.0 Implement Cache Enforcement Logic
  - [x] 5.1 Add TTL validation to check cache expiration
  - [x] 5.2 Implement cache size monitoring and enforcement
  - [x] 5.3 Add cleanup methods for expired and oversized cache entries
  - [x] 5.4 Integrate cache validation into registry operations
