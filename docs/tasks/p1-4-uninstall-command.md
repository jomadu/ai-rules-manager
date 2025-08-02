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
- [x] **Create uninstall command structure**:
  ```bash
  arm uninstall <ruleset-name>
  ```
- [x] **Identify installed files**:
  - Read from rules.lock to find exact files
  - Handle multiple target directories
  - Graceful handling when ruleset not found
- [x] **Safe removal process**:
  - Remove ruleset directory from target paths
  - Clean up empty parent directories automatically
  - Preserve files from other rulesets
- [x] **Manifest updates**:
  - Remove from rules.json dependencies
  - Update rules.lock to remove entry
  - Handle missing manifest files gracefully
- [x] **Partial failure handling**:
  - Continue on target directory failures
  - Report which targets succeeded/failed
  - Update manifest even with partial failures
- [x] **Graceful validation**:
  - Idempotent behavior (safe to run multiple times)
  - No error for non-existent rulesets
  - Cache preservation for future reinstalls

## Acceptance Criteria
- [x] `arm uninstall typescript-rules` removes all files
- [x] rules.json and rules.lock are updated correctly
- [x] Other rulesets remain unaffected
- [x] Cache is preserved for future reinstalls
- [x] Graceful handling of non-existent rulesets
- [x] Partial failure reporting with clear messages
- [x] Empty parent directory cleanup

## Dependencies
- os (standard library)
- path/filepath (standard library)

## Files Created
- [x] `cmd/arm/uninstall.go` - CLI command implementation
- [x] `internal/uninstaller/uninstaller.go` - Core uninstall logic
- [x] `internal/uninstaller/uninstaller_test.go` - Unit tests

## Implementation Notes
- Implemented graceful approach without pre-validation
- Cache preservation follows principle of least surprise
- Partial failure handling provides clear user feedback
- Idempotent behavior makes command safe to retry
- Automatic cleanup of empty directories maintains clean structure
