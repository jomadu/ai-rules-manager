# P3.2: Outdated Command

## Overview
Implement the `arm outdated` command to display installed rulesets that have newer versions available.

## Requirements
- Compare installed vs available versions
- Display outdated rulesets in table format
- Show current and latest available versions
- Handle registry connectivity issues

## Tasks
- [ ] **Create outdated command structure**:
  ```bash
  arm outdated                 # show all outdated
  arm outdated --json          # JSON output
  arm outdated <ruleset>       # check specific ruleset
  ```
- [ ] **Version comparison logic**:
  - Read current versions from rules.lock
  - Fetch latest versions from registries
  - Compare respecting version constraints
  - Handle pre-release versions
- [ ] **Table formatting**:
  - Columns: Name, Current, Latest, Source
  - Color coding for different update types
  - Sort by name or update priority
- [ ] **Registry connectivity**:
  - Handle offline registries gracefully
  - Show partial results when some registries fail
  - Provide clear error messages
  - Implement timeout handling

## Acceptance Criteria
- [ ] `arm outdated` shows all rulesets with updates
- [ ] Table format is clear and readable
- [ ] JSON output is properly formatted
- [ ] Registry failures don't crash the command
- [ ] Version constraints are respected
- [ ] Empty state shows helpful message

## Dependencies
- github.com/hashicorp/go-version (version comparison)
- text/tabwriter (table formatting)

## Files to Create
- `cmd/arm/outdated.go`
- `internal/outdated/checker.go`
- `internal/outdated/formatter.go`

## Example Output
```
NAME                CURRENT    LATEST     SOURCE
typescript-rules    1.2.3      1.2.5      company
security-rules      2.1.0      2.2.0      default
react-rules         0.5.1      up to date github

2 rulesets can be updated.
Run 'arm update' to update all.
```

## Notes
- Consider caching version information
- Plan for update notifications
- Handle rate limiting from registries