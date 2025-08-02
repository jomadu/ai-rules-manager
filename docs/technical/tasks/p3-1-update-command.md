# P3.1: Update Command

## Overview
Implement the `arm update` command to check for and install newer versions of installed rulesets while preserving version constraints.

## Requirements
- Check for newer versions across registries
- Update individual or all rulesets
- Preserve version constraints from rules.json
- Update rules.lock with new versions

## Tasks
- [ ] **Create update command structure**:
  ```bash
  arm update                    # update all
  arm update <ruleset-name>     # update specific
  arm update --dry-run          # show what would update
  ```
- [ ] **Version checking logic**:
  - Compare installed vs available versions
  - Respect version constraints (^1.0.0 allows 1.x.x)
  - Handle pre-release versions appropriately
  - Check multiple registries for latest versions
- [ ] **Update process**:
  - Download new versions
  - Backup current installation
  - Install new version
  - Update rules.lock
  - Rollback on failure
- [ ] **Constraint preservation**:
  - Keep original version specs in rules.json
  - Only update rules.lock with exact versions
  - Validate new versions satisfy constraints
- [ ] **Progress reporting**:
  - Show which rulesets are being checked
  - Display available updates
  - Progress bars for downloads
  - Summary of completed updates

## Acceptance Criteria
- [ ] `arm update` updates all outdated rulesets
- [ ] `arm update typescript-rules` updates specific ruleset
- [ ] Version constraints are respected
- [ ] rules.lock is updated with new exact versions
- [ ] rules.json constraints remain unchanged
- [ ] Failed updates are rolled back
- [ ] --dry-run shows planned updates without executing

## Dependencies
- github.com/hashicorp/go-version (version comparison)

## Files to Create
- `cmd/arm/update.go`
- `internal/updater/updater.go`
- `internal/updater/checker.go`

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

## Notes
- Consider parallel update checking for performance
- Plan for update notifications/scheduling
- Handle registry connectivity issues gracefully
