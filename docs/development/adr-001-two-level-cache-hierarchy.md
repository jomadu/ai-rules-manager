# ADR-001: Three-Level Cache Hierarchy for Pattern-Aware Caching

## Status
Accepted

## Context

ARM's cache system had a critical flaw where two projects using the same Git registry URL could experience cache collisions when installing the same ruleset with different patterns. This occurred because the cache key was generated solely from `registry_type + registry_url`, ignoring the patterns parameter.

### Problem Scenario
- Project A: `arm install ruleset-x --patterns="*.md"`
- Project B: `arm install ruleset-x --patterns="*.py"`

Both projects would share the same cache entry, causing Project B to receive files matching Project A's patterns instead of its own.

### Requirements
1. **Pattern Precedence**: CLI patterns override manifest patterns and update the manifest
2. **Cache Isolation**: Different patterns must produce separate cache entries
3. **Pattern Normalization**: Order-independent pattern matching (`["*.md", "*.py"]` == `["*.py", "*.md"]`)
4. **Backward Compatibility**: Existing cache methods must continue working
5. **Clean Migration**: Users handle cache cleanup themselves

## Decision

Implement a **three-level cache hierarchy**:

1. **Registry Level**: `hash(registry_type + registry_url)` (existing)
2. **Ruleset Level**: `hash(ruleset_name + normalized_patterns)` (new)
3. **Version Level**: `{version}` (new)

### Cache Structure
```
/cache/registries/{registry_hash}/rulesets/{ruleset_hash}/{version}/
```

Where:
- `registry_hash` = SHA-256 of `registry_type:normalized_url`
- `ruleset_hash` = SHA-256 of `ruleset_name:sorted_patterns`
- `version` = Resolved version string (commit hash for latest, or semver like "v1.2.0")

### Pattern Normalization Strategy
- **Sort patterns alphabetically** for order independence
- **Trim whitespace** for consistency
- **Empty patterns** handled as empty string in hash

### Implementation Details

#### New Cache Manager Methods
```go
GetRulesetCacheKey(rulesetName string, patterns []string) string
GetRulesetCachePath(registryType, registryURL, rulesetName string, patterns []string) (string, error)
```

#### Updated RulesetStorage Methods
All methods now accept `patterns []string` parameter:
- `StoreRulesetFiles(..., patterns []string)`
- `GetRulesetFiles(..., patterns []string)`
- `ListRulesetVersions(..., patterns []string)`
- `GetRulesetStats(..., patterns []string)`
- `RemoveRulesetVersion(..., patterns []string)`

#### Legacy Compatibility
Original methods maintained as wrappers calling new methods with `nil` patterns.

## Consequences

### Positive
- **Cache Isolation**: Different patterns create separate cache entries
- **Pattern Consistency**: Order-independent pattern matching
- **Registry Efficiency**: Repository clones still shared at registry level
- **Backward Compatibility**: Existing code continues working
- **Clean Architecture**: Three-level hierarchy matches logical structure
- **Version Support**: Consistent structure for future version management

### Negative
- **Increased Storage**: Multiple pattern combinations consume more disk space
- **API Complexity**: Additional patterns parameter in storage methods
- **Migration Required**: Users must clear existing cache for new structure

### Trade-offs Accepted
- **Storage vs Correctness**: Chose correctness over disk space efficiency
- **API Complexity vs Safety**: Added parameters to prevent silent bugs
- **Manual Migration vs Automatic**: Trusted users to handle cache cleanup

## Implementation Notes

### Pattern Update Behavior
- **Manifest install** (`arm install`): Use patterns from `arm.json`
- **CLI install** (`arm install --patterns="*.md"`): CLI patterns replace manifest patterns

### Cache Invalidation
- **Keep both caches**: Old pattern entries expire naturally via TTL
- **No auto-invalidation**: Avoids unnecessary re-downloads

### Validation
- **Basic syntax check**: Validate glob pattern syntax
- **No repository validation**: Trust user input, let glob matching handle errors

## Examples

### Before (Problematic)
```
Project A: --patterns="*.md"
Project B: --patterns="*.py"
Cache Key: dc2d0358c99b8ec27e11446ca61d4158624692072bf207a1f3a9a4332be595da
Result: Both projects share same cache → collision
```

### After (Fixed)
```
Project A: --patterns="*.md"
Cache Path: /registries/dc2d0358.../rulesets/b7aa92eddd49309f.../latest/

Project B: --patterns="*.py"
Cache Path: /registries/dc2d0358.../rulesets/0922776b231a28cc.../latest/

Result: Separate cache entries → no collision
```

## Verification

Test results confirm:
- ✅ Same patterns (different order) → Same cache key
- ✅ Different patterns → Different cache keys
- ✅ Empty patterns → Unique cache key
- ✅ Three-level path structure correctly implemented
- ✅ All existing tests pass

## References
- Issue: Cache collision between projects with different patterns
- Implementation: `internal/cache/manager.go`, `internal/cache/ruleset_storage.go`
- Tests: `internal/cache/ruleset_storage_test.go`
