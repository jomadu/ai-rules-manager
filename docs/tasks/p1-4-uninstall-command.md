# P1.4: Uninstall Command

## Overview
Implement the `arm uninstall` command to safely remove rulesets from target directories and update manifest files.

## Requirements
- Remove ruleset files from all target directories
- Update rules.json to remove dependency
- Update rules.lock file
- Clean up orphaned cache entries
- Validate removal operations

## Tasks
- [ ] **Create uninstall command structure**:
  ```bash
  arm uninstall <ruleset-name>
  ```
- [ ] **Identify installed files**:
  - Read from rules.lock to find exact files
  - Handle multiple target directories
  - Check for shared dependencies
- [ ] **Safe removal process**:
  - Verify files belong to specified ruleset
  - Remove files and empty directories
  - Preserve files from other rulesets
- [ ] **Manifest updates**:
  - Remove from rules.json dependencies
  - Update rules.lock to remove entry
  - Recalculate dependency tree
- [ ] **Cache cleanup**:
  - Remove cached .tar.gz if no longer needed
  - Clean up metadata cache entries
- [ ] **Validation**:
  - Confirm ruleset is actually installed
  - Check for dependent rulesets
  - Warn about breaking dependencies

## Acceptance Criteria
- [ ] `arm uninstall typescript-rules` removes all files
- [ ] rules.json and rules.lock are updated correctly
- [ ] Other rulesets remain unaffected
- [ ] Cache is cleaned up appropriately
- [ ] Error if trying to uninstall non-existent ruleset
- [ ] Warning if removal breaks dependencies
- [ ] Dry-run option shows what would be removed

## Dependencies
- os (standard library)
- path/filepath (standard library)

## Files to Create
- `cmd/arm/uninstall.go`
- `internal/uninstaller/uninstaller.go`
- `internal/validator/validator.go`

## Notes
- Consider --force flag to override dependency warnings
- Implement --dry-run for preview
- Log all removal operations for debugging
