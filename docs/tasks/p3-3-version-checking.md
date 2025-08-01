# P3.3: Version Checking

## Overview
Implement parallel version checking with caching, rate limiting, and progress indicators for efficient registry queries.

## Requirements
- Implement parallel version checking
- Cache version information
- Handle registry rate limits
- Provide progress indicators

## Tasks
- [ ] **Parallel version checking**:
  - Concurrent goroutines for multiple rulesets
  - Worker pool pattern for controlled concurrency
  - Context-based cancellation
  - Error aggregation from multiple sources
- [ ] **Version information caching**:
  - Cache latest version data locally
  - TTL-based cache expiration
  - Cache invalidation strategies
  - Persistent cache across runs
- [ ] **Rate limiting**:
  - Respect registry rate limits
  - Exponential backoff on failures
  - Queue requests when limits exceeded
  - Per-registry rate limiting
- [ ] **Progress indicators**:
  - Progress bars for long operations
  - Status updates during checking
  - ETA calculations
  - Cancellation support

## Acceptance Criteria
- [ ] Version checking runs in parallel efficiently
- [ ] Cache reduces redundant registry calls
- [ ] Rate limits are respected and handled gracefully
- [ ] Progress indicators provide useful feedback
- [ ] Errors don't stop other checks from completing
- [ ] Context cancellation works properly

## Dependencies
- context (standard library)
- sync (standard library)
- time (standard library)

## Files to Create
- `internal/checker/parallel.go`
- `internal/checker/cache.go`
- `internal/checker/ratelimit.go`
- `internal/checker/progress.go`

## Implementation Details
```go
type VersionChecker struct {
    registries []Registry
    cache      Cache
    limiter    RateLimiter
    workers    int
}

func (vc *VersionChecker) CheckVersions(ctx context.Context, rulesets []string) <-chan VersionResult {
    // Implementation with worker pool
}
```

## Notes
- Consider adaptive concurrency based on registry performance
- Plan for offline mode with cached data
- Implement proper cleanup of goroutines
