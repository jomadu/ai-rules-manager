# P3.1: Update Command

## Overview
Implement the `arm update` command to check for and install newer versions of installed rulesets while preserving version constraints.

## Requirements
- Check for newer versions across registries
- Update individual or all rulesets
- Preserve version constraints from rules.json
- Update rules.lock with new versions

## Tasks
- [x] **Create update command structure**:
  ```bash
  arm update                    # update all
  arm update <ruleset-name>     # update specific
  arm update --dry-run          # show what would update
  ```
- [x] **Version checking logic**:
  - Compare installed vs available versions
  - Respect version constraints using hashicorp/go-version
  - Handle pre-release versions appropriately
  - Check multiple registries for latest versions
- [x] **Update process**:
  - Download new versions
  - Backup current installation
  - Install new version
  - Update rules.lock
  - Rollback on failure
- [x] **Constraint preservation**:
  - Keep original version specs in rules.json
  - Only update rules.lock with exact versions
  - Validate new versions satisfy constraints
- [x] **Progress reporting**:
  - Show which rulesets are being checked
  - Display available updates
  - Progress bars for downloads using schollz/progressbar/v3
  - Summary of completed updates

## Acceptance Criteria
- [x] `arm update` updates all outdated rulesets
- [x] `arm update typescript-rules` updates specific ruleset
- [x] Version constraints are respected
- [x] rules.lock is updated with new exact versions
- [x] rules.json constraints remain unchanged
- [x] Failed updates are rolled back
- [x] --dry-run shows planned updates without executing

## Dependencies
- github.com/hashicorp/go-version (version comparison) ✅
- github.com/schollz/progressbar/v3 (progress bars) ✅

## Files Created
- `cmd/arm/update.go` ✅
- `internal/updater/updater.go` ✅
- `internal/updater/updater_test.go` ✅

## Example Output
```
Checking for updates...
✓ typescript-rules: 1.2.3 → 1.2.5
✓ security-rules: 2.1.0 (up to date)
✗ react-rules: 0.5.1 (no compatible updates)

Updating 1 ruleset...
Downloading typescript-rules@1.2.5... ████████████ 100%
✓ Updated typescript-rules to 1.2.5
```

## Implementation Notes
- ✅ Version constraint checking using hashicorp/go-version library
- ✅ Progress bars implemented with schollz/progressbar/v3
- ✅ Backup/restore functionality for failed updates
- ✅ Comprehensive test coverage for version checking and backup/restore
- ✅ Integration with existing registry and config systems
- ✅ Proper error handling with rollback on failure
- ✅ Dry-run mode with clear indication

## Status: ✅ COMPLETED
**Completion Date**: January 2025
**Commit**: b25f34a - feat: implement update command with version constraints and progress bars
