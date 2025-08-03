# P1.5: List Command

## Overview
Implement the `arm list` command to display all installed rulesets with their versions and sources in a formatted table.

## Requirements
- Read installed rulesets from rules.lock
- Display formatted table output
- Show ruleset name, version, and source
- Handle empty/missing manifest files gracefully

## Tasks
- [x] **Create list command structure**:
  ```bash
  arm list
  arm list --format=json  # optional JSON output
  ```
- [x] **Read installation data**:
  - Parse rules.lock file
  - Handle missing or corrupted files
  - Gather ruleset metadata
- [x] **Format table output**:
  - Column headers: Name, Version, Source
  - Proper alignment and spacing
  - ~~Color coding for different sources~~ (not implemented)
- [x] **Handle edge cases**:
  - No rulesets installed
  - Missing rules.lock file
  - Corrupted manifest files
- [x] **Optional features**:
  - JSON output format
  - ~~Filter by source~~ (not implemented)
  - Sort by name (implemented)

## Acceptance Criteria
- [x] `arm list` shows installed rulesets in table format
- [x] Empty state shows helpful message
- [x] Table is properly formatted and readable
- [x] JSON output option works correctly
- [x] Handles missing files gracefully
- [x] Shows accurate version and source information

## Dependencies
- text/tabwriter (standard library)
- encoding/json (standard library)

## Files Created
- `cmd/arm/list.go` ✅
- `cmd/arm/list_test.go` ✅
- ~~`internal/formatter/table.go`~~ (implemented inline)
- ~~`internal/reader/lockfile.go`~~ (uses existing types.LoadLockFile)

## Example Output
```
NAME                VERSION    SOURCE
typescript-rules    1.2.3      company
security-rules      2.1.0      default
react-rules         0.5.1      github
```

## Notes
- Consider adding --verbose flag for more details
- Plan for future filtering and sorting options
- Ensure consistent formatting across different terminal sizes
