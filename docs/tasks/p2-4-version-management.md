# P2.4: Version Management

## Overview
Implement semantic version parsing, range resolution, dependency tree resolution, and conflict handling.

## Requirements
- Integrate semantic version parsing library
- Implement version range resolution (^, ~, exact)
- Create dependency tree resolution
- Handle version conflicts

## Tasks
- [ ] **Semantic Version Integration**:
  - Use github.com/hashicorp/go-version
  - Parse version strings (1.2.3, v1.2.3-beta.1)
  - Validate version formats
  - Compare versions correctly
- [ ] **Version Range Resolution**:
  - Caret ranges: ^1.2.3 (>=1.2.3 <2.0.0)
  - Tilde ranges: ~1.2.3 (>=1.2.3 <1.3.0)
  - Exact versions: 1.2.3
  - Range operators: >=1.0.0 <2.0.0
- [ ] **Dependency Tree Resolution**:
  - Build dependency graph
  - Resolve transitive dependencies
  - Find compatible version combinations
  - Detect circular dependencies
- [ ] **Conflict Handling**:
  - Identify version conflicts
  - Suggest resolution strategies
  - Allow manual conflict resolution
  - Provide clear error messages

## Acceptance Criteria
- [ ] Version parsing handles all common formats
- [ ] Range resolution works for ^, ~, and exact versions
- [ ] Dependency tree builds correctly
- [ ] Circular dependencies are detected
- [ ] Version conflicts show helpful messages
- [ ] Resolution algorithm is deterministic

## Dependencies
- github.com/hashicorp/go-version (semantic versioning)

## Files to Create
- `internal/version/parser.go`
- `internal/version/resolver.go`
- `internal/version/constraints.go`
- `internal/version/conflicts.go`

## Example Usage
```go
constraint, _ := version.NewConstraint("^1.2.0")
version, _ := version.NewVersion("1.2.5")
if constraint.Check(version) {
    // Version satisfies constraint
}
```

## Test Cases
- [ ] Parse various version formats
- [ ] Resolve caret and tilde ranges
- [ ] Handle pre-release versions
- [ ] Detect circular dependencies
- [ ] Resolve complex dependency trees
