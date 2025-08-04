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
- [x] **Create outdated command structure**:
  ```bash
  arm outdated                    # show all outdated rulesets
  arm outdated typescript-rules   # check specific ruleset
  arm outdated --format=json      # JSON output
  arm outdated --format=table     # table output (default)
  arm outdated --only-outdated    # show only outdated rulesets
  arm outdated --no-color         # disable colored output
  ```
- [x] **Version comparison logic**:
  - Extracted shared checker module from update command
  - Compare installed vs latest compatible versions
  - Handle version constraint validation
  - Check registries for latest versions
- [x] **Output formatting**:
  - Table format with columns: Name, Current, Available, Constraint, Status
  - JSON format for programmatic consumption
  - Color coding for different update types (green=up-to-date, yellow=outdated, red=error)
  - Summary statistics with counts
- [x] **Error handling**:
  - Show errors in output with error status (Option A)
  - Continue checking other rulesets on individual failures
  - Provide actionable error information
- [x] **Performance decisions**:
  - Sequential version checking (parallel optimization deferred)
  - Efficient registry queries using existing infrastructure

## Acceptance Criteria
- [x] `arm outdated` shows all rulesets with available updates
- [x] Table format displays clear comparison information
- [x] JSON format provides structured data
- [x] Version constraints are respected
- [x] Performance is acceptable for large numbers of rulesets
- [x] Exit codes indicate presence of outdated rulesets
- [x] Unit tests cover new functionality

## Dependencies
- Reuse existing updater logic
- github.com/hashicorp/go-version (already added)

## Files Created
- `cmd/arm/outdated.go` - Main command implementation
- `internal/updater/checker.go` - Shared version checking logic
- `cmd/arm/outdated_test.go` - Unit tests for command functions
- `internal/updater/checker_test.go` - Unit tests for checker module

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

## Status: ✅ COMPLETED
**Completed**: January 2025
**Dependencies**: P3.1 Update Command (completed)

## Implementation Notes
- Created shared `internal/updater/checker.go` module for version checking logic
- Implemented `cmd/arm/outdated.go` with full feature set
- Added filtering options: `--only-outdated` and specific ruleset checking
- Used exact constraint display from rules.json (Option A)
- Error handling shows individual failures without stopping entire command
- Exit codes: 0 (up to date), 1 (updates available), 2 (error)
- Sequential processing chosen for initial implementation
