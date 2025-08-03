# P4.1: Cache System

## Overview
Implement comprehensive caching system for downloaded rulesets, registry metadata, and version information with expiration policies.

## Requirements
- Implement cache storage and retrieval
- Cache downloaded tar.gz files
- Cache registry metadata
- Implement cache expiration policies

## Tasks
- [ ] **Cache storage structure**:
  ```
  .mpm/cache/
    packages/
      <source>/
        <ruleset>/
          <version>/
            package.tar.gz
            metadata.json
    registry/
      <source>/
        metadata.json
        versions.json
    temp/
      downloads/
  ```
- [ ] **Package caching**:
  - Store downloaded .tar.gz files
  - Cache package metadata
  - Implement content-addressable storage
  - Verify integrity on retrieval
- [ ] **Registry metadata caching**:
  - Cache registry responses
  - Store version lists
  - Cache authentication tokens (securely)
  - Handle registry-specific metadata
- [ ] **Expiration policies**:
  - TTL-based expiration
  - LRU eviction for size limits
  - Manual cache invalidation
  - Configurable retention policies

## Acceptance Criteria
- [ ] Downloaded packages are cached correctly
- [ ] Cache retrieval is faster than re-download
- [ ] Expired cache entries are cleaned up
- [ ] Cache size limits are enforced
- [ ] Cache corruption is detected and handled
- [ ] Cache statistics are available

## Dependencies
- crypto/sha256 (content hashing)
- encoding/json (metadata storage)
- time (expiration handling)

## Files to Create
- `internal/cache/storage.go`
- `internal/cache/packages.go`
- `internal/cache/metadata.go`
- `internal/cache/expiration.go`

## Cache Configuration
```go
type CacheConfig struct {
    Location    string
    MaxSize     int64
    PackageTTL  time.Duration
    MetadataTTL time.Duration
    CleanupInterval time.Duration
}
```

## Notes
- Consider compression for cached metadata
- Plan for cache migration between versions
- Implement cache repair functionality
