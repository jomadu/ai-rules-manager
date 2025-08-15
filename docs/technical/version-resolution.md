# Version Resolution

Technical specification for ARM's semantic versioning and constraint resolution system.

## Overview

ARM implements a sophisticated version resolution system that follows semantic versioning principles while providing flexible constraint matching. The system prioritizes stability and predictability, ensuring teams can specify version requirements with confidence.

## Supported Version Formats

### Exact Versions
Specify precise versions for deterministic installations:
- `rules@1.2.3` - Exact semantic version
- `rules@v1.2.3` - Version with 'v' prefix (common in Git tags)

### Semantic Version Constraints
Flexible constraints that balance stability with updates:
- **Caret (`^`)** - Compatible within major version (e.g., `^1.2.0` allows 1.2.0 to 1.9.9)
- **Tilde (`~`)** - Compatible within minor version (e.g., `~1.2.0` allows 1.2.0 to 1.2.9)
- **Comparison operators** - Direct version comparisons (`>=1.1.0`, `<2.0.0`)

### Range Constraints
Combine multiple constraints for precise control:
- Range specifications like `">=1.0.0 <2.0.0"`
- Wildcard patterns (`1.x`, `1.2.x`) for flexible matching

### Special Versions
Non-semantic version references:
- `latest` - Most recent stable release
- Branch names (`main`, `develop`) - Direct branch references (Git registries only)

## Version Resolution Algorithm

### Resolution Process
The version resolver follows a four-step process to find the best matching version:

1. **Constraint Parsing** - Analyzes the version constraint to understand the matching requirements
2. **Version Discovery** - Retrieves all available versions from the target registry
3. **Filtering and Sorting** - Applies constraint rules and sorts candidates by semantic version precedence
4. **Best Match Selection** - Chooses the optimal version based on the constraint type and available options

### Constraint Types and Behaviors
Each constraint type implements specific matching logic:

**Latest Constraint** - Always selects the newest available semantic version, ignoring pre-release versions unless no stable versions exist.

**Exact Constraints** - Match only the specified version, supporting both standard format (1.2.3) and v-prefixed format (v1.2.3).

**Operator Constraints** - Apply mathematical comparisons (>=, <=, >, <) against the version number, following semantic versioning precedence rules.

**Range Constraints** - Combine multiple constraints with logical AND operations, requiring all conditions to be satisfied.

## Constraint Matching Behaviors

### Caret Constraints (^)
**Purpose**: Allow compatible updates within the same major version

**Matching Rules**:
- Must have the same major version number
- Must be greater than or equal to the specified version
- Excludes versions with different major numbers

**Examples**:
- `^1.2.3` matches: 1.2.3, 1.2.4, 1.3.0, 1.9.9
- `^1.2.3` excludes: 1.2.2, 2.0.0, 0.9.9

**Use Case**: Ideal for production environments where you want bug fixes and new features but need to avoid breaking changes.

### Tilde Constraints (~)
**Purpose**: Allow compatible updates within the same minor version

**Matching Rules**:
- Must have the same major and minor version numbers
- Must be greater than or equal to the specified version
- Only allows patch-level updates

**Examples**:
- `~1.2.3` matches: 1.2.3, 1.2.4, 1.2.9
- `~1.2.3` excludes: 1.2.2, 1.3.0, 2.0.0

**Use Case**: Conservative approach for critical systems where only bug fixes are acceptable.

### Wildcard Patterns
**Purpose**: Flexible matching with explicit version component control

**Behaviors**:
- `1.x` matches any version in the 1.x.x series
- `1.2.x` matches any patch version in the 1.2.x series
- Always selects the highest matching version available

## Registry-Specific Resolution

### Git Registries
**Version Discovery Process**:
- Scans Git tags for semantic version patterns
- Supports both standard (1.2.3) and v-prefixed (v1.2.3) tag formats
- Includes branch names as valid version targets for non-semantic constraints
- Prioritizes semantic version tags over branch references

**Special Behaviors**:
- Branch names (main, develop, feature/xyz) are treated as exact version matches
- Tag validation ensures only properly formatted semantic versions are considered
- Remote repository access is cached to minimize network requests

### S3 Registries
**Version Discovery Process**:
- Lists objects using registry prefix patterns
- Extracts version identifiers from S3 object key structures
- Supports nested directory structures for version organization
- Validates semantic version format before inclusion

**Directory Structure Expectations**:
- Versions organized as separate directories or prefixes
- Object keys must follow predictable naming conventions
- Metadata extraction from S3 object properties when available

## Version Sorting and Selection

### Semantic Version Precedence
**Sorting Behavior**:
- Versions are sorted in descending order (newest first)
- Follows semantic versioning precedence rules (major.minor.patch)
- Pre-release versions are ranked lower than stable releases
- Build metadata is ignored during comparison

**Version Validation**:
- Invalid semantic versions are filtered out during sorting
- Both standard and v-prefixed formats are normalized
- Non-semantic versions (branch names) are handled separately

### Selection Strategy
**Latest Constraints**: Always return the highest semantic version available, preferring stable releases over pre-release versions.

**Exact Constraints**: Return the specific version if found, otherwise fail with a clear error message.

**Range Constraints**: Return the highest version that satisfies all constraint conditions, ensuring optimal compatibility.

**Fallback Behavior**: When no semantic versions match, attempt to resolve branch names or alternative version formats based on registry capabilities.

## Caching and Performance

### Version Cache Strategy
**Cache Behavior**:
- Version lists are cached per registry-ruleset combination
- Time-to-live (TTL) prevents stale version information
- Thread-safe operations support concurrent resolution requests
- Cache keys are generated using cryptographic hashing for uniqueness

**Performance Optimizations**:
- Registry queries are minimized through intelligent caching
- Version parsing and sorting results are memoized
- Concurrent resolution requests share cached data
- Cache size limits prevent memory exhaustion

**Cache Invalidation**:
- TTL-based expiration ensures fresh version data
- Manual cache clearing for immediate updates
- Registry-specific invalidation strategies
- Graceful degradation when cache is unavailable

## Error Handling

### Error Categories
**Resolution Errors**:
- **No Versions Found** - Registry contains no versions for the specified ruleset
- **No Matching Versions** - Available versions don't satisfy the constraint
- **Invalid Constraint** - Malformed version constraint syntax
- **Invalid Version** - Non-semantic version format encountered

**Error Context**:
All version resolution errors include complete context (registry, ruleset, constraint) to aid in debugging and user feedback.

### Fallback Strategies
**Graceful Degradation**:
- Latest constraint falls back to main branch for Git registries when no semantic versions exist
- Constraint relaxation attempts broader version matching when strict constraints fail
- Registry-specific fallbacks handle different version discovery mechanisms

**User Experience**:
- Clear error messages explain why resolution failed
- Suggested alternatives when appropriate
- Detailed logging for troubleshooting complex scenarios

## Testing and Validation

### Constraint Validation
**Test Coverage**:
- Comprehensive constraint matching scenarios
- Edge cases for version boundary conditions
- Registry-specific version discovery patterns
- Performance benchmarks for large version sets

**Validation Approach**:
- Unit tests verify individual constraint behaviors
- Integration tests validate end-to-end resolution workflows
- Performance tests ensure scalability with large version catalogs
- Regression tests prevent breaking changes to resolution logic
