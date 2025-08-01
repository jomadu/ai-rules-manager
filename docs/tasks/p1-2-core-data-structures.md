# P1.2: Core Data Structures

## Overview
Define the fundamental data structures for rulesets, manifests, and configuration that form the backbone of MRM.

## Requirements
- Ruleset metadata structure
- rules.json schema and parsing
- rules.lock file format
- Registry source configuration types
- Cache directory structure design

## Tasks
- [ ] **Define Ruleset struct**:
  ```go
  type Ruleset struct {
    Name     string
    Version  string
    Source   string
    Files    []string
    Checksum string
  }
  ```
- [ ] **Implement RulesManifest** for rules.json:
  ```go
  type RulesManifest struct {
    Targets      []string
    Dependencies map[string]string
  }
  ```
- [ ] **Create RulesLock** structure for rules.lock:
  ```go
  type RulesLock struct {
    Version      string
    Dependencies map[string]LockedDependency
  }
  ```
- [ ] **Define RegistryConfig** for .mpmrc:
  ```go
  type RegistryConfig struct {
    Sources map[string]RegistrySource
  }
  ```
- [ ] **Design cache directory structure** and metadata format

## Acceptance Criteria
- [ ] All structs have proper JSON tags
- [ ] Validation methods for each structure
- [ ] Parse/serialize methods work correctly
- [ ] Comprehensive unit tests for all structures
- [ ] Documentation with examples

## Dependencies
- encoding/json (standard library)
- gopkg.in/yaml.v3 (for .mpmrc INI format alternative)

## Files to Create
- `pkg/types/ruleset.go`
- `pkg/types/manifest.go`
- `pkg/types/config.go`
- `internal/parser/manifest.go`
- `internal/parser/lockfile.go`

## Notes
- Consider using struct tags for validation
- Plan for schema evolution (version field)
- Ensure thread-safe operations for concurrent access