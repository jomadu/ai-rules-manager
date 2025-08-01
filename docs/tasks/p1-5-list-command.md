# P1.5: List Command

## Overview
Implement the `mrm list` command to display all installed rulesets with their versions and sources in a formatted table.

## Requirements
- Read installed rulesets from rules.lock
- Display formatted table output
- Show ruleset name, version, and source
- Handle empty/missing manifest files gracefully

## Tasks
- [ ] **Create list command structure**:
  ```bash
  mrm list
  mrm list --format=json  # optional JSON output
  ```
- [ ] **Read installation data**:
  - Parse rules.lock file
  - Handle missing or corrupted files
  - Gather ruleset metadata
- [ ] **Format table output**:
  - Column headers: Name, Version, Source
  - Proper alignment and spacing
  - Color coding for different sources
- [ ] **Handle edge cases**:
  - No rulesets installed
  - Missing rules.lock file
  - Corrupted manifest files
- [ ] **Optional features**:
  - JSON output format
  - Filter by source
  - Sort options (name, version, source)

## Acceptance Criteria
- [ ] `mrm list` shows installed rulesets in table format
- [ ] Empty state shows helpful message
- [ ] Table is properly formatted and readable
- [ ] JSON output option works correctly
- [ ] Handles missing files gracefully
- [ ] Shows accurate version and source information

## Dependencies
- text/tabwriter (standard library)
- encoding/json (standard library)

## Files to Create
- `cmd/mrm/list.go`
- `internal/formatter/table.go`
- `internal/reader/lockfile.go`

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