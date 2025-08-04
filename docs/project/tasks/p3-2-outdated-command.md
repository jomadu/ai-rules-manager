# P3.2: Outdated Command

## Overview
Implement the `arm outdated` command to show available updates for installed rulesets without performing the actual updates.

## Requirements
- Display rulesets with available updates
- Show current vs available versions
- Respect version constraints from rules.json
- Support different output formats (table, JSON)
- Indicate update compatibility

## Tasks
- [ ] **Create outdated command structure**:
  ```bash
  arm outdated                  # show all outdated rulesets
  arm outdated --format=json    # JSON output
  arm outdated --format=table   # table output (default)
  ```
- [ ] **Version comparison logic**:
  - Reuse version checking from update command
  - Compare installed vs latest compatible versions
  - Handle pre-release versions appropriately
  - Check multiple registries for latest versions
- [ ] **Output formatting**:
  - Table format with columns: Name, Current, Available, Constraint
  - JSON format for programmatic consumption
  - Color coding for different update types
  - Summary statistics
- [ ] **Performance optimization**:
  - Parallel version checking
  - Cache version information
  - Efficient registry queries

## Acceptance Criteria
- [ ] `arm outdated` shows all rulesets with available updates
- [ ] Table format displays clear comparison information
- [ ] JSON format provides structured data
- [ ] Version constraints are respected
- [ ] Performance is acceptable for large numbers of rulesets
- [ ] Exit codes indicate presence of outdated rulesets

## Dependencies
- Reuse existing updater logic
- github.com/hashicorp/go-version (already added)

## Files to Create
- `cmd/arm/outdated.go`
- `internal/updater/checker.go` (extract from updater.go)

## Example Output

### Table Format
```
Name              Current  Available  Constraint      Status
typescript-rules  1.2.3    1.2.5      >= 1.0.0, < 2.0.0  Update available
security-rules    2.1.0    2.1.0      >= 2.1.0, < 2.2.0  Up to date
react-rules       0.5.1    1.0.0      >= 0.5.0, < 1.0.0  No compatible update

2 rulesets checked, 1 update available
```

### JSON Format
```json
{
  "rulesets": [
    {
      "name": "typescript-rules",
      "current": "1.2.3",
      "available": "1.2.5",
      "constraint": ">= 1.0.0, < 2.0.0",
      "status": "outdated"
    }
  ],
  "summary": {
    "total": 2,
    "outdated": 1,
    "upToDate": 1
  }
}
```

## Implementation Notes
- Extract version checking logic into shared checker module
- Implement efficient caching to avoid repeated registry calls
- Use consistent exit codes (0 = up to date, 1 = updates available, 2 = error)
- Consider rate limiting for registry API calls

## Status: 📋 PLANNED
**Target Completion**: February 2025
**Dependencies**: P3.1 Update Command (completed)
