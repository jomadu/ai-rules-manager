# P4.2: Clean Command

## Overview
Implement the `arm clean` command to remove unused rulesets from project targets and clear global cache.

## Requirements ✅ COMPLETED
- Remove unused rulesets from project targets based on rules.lock
- Clear entire global cache with confirmation
- Support --dry-run flag for preview
- Handle both org/package and simple package structures

## Tasks
- [ ] **Create clean command structure**:
  ```bash
  arm clean                    # clean all unused
  arm clean --cache            # clean cache only
  arm clean --dry-run          # show what would be cleaned
  arm clean --all              # aggressive cleanup
  ```
- [ ] **Identify unused cache entries**:
  - Compare cache contents with rules.lock
  - Find rulesets not referenced by any project
  - Identify orphaned partial downloads
  - Detect corrupted cache entries
- [ ] **Cache size analysis**:
  - Calculate total cache size
  - Show size per ruleset/version
  - Identify largest cache consumers
  - Report space that would be freed
- [ ] **Cleanup operations**:
  - Remove unused .tar.gz files
  - Clean up empty directories
  - Remove expired metadata files
  - Clear temporary download files
- [ ] **Safety measures**:
  - Confirm before destructive operations
  - Preserve currently used rulesets
  - Backup critical cache metadata
  - Provide undo capability where possible

## Acceptance Criteria ✅ COMPLETED
- [x] `arm clean` removes unused rulesets from project targets
- [x] `arm clean --cache` clears global cache with confirmation
- [x] --dry-run shows cleanup plan without executing
- [x] Currently used rulesets (in rules.lock) are never removed
- [x] Empty directories are cleaned up after removal
- [x] User confirmation for cache cleanup
- [x] Comprehensive unit test coverage

## Dependencies
- os (standard library)
- path/filepath (standard library)

## Files Created ✅
- `cmd/arm/clean.go` - CLI command implementation
- `internal/cleaner/cleaner.go` - Core cleanup logic
- `cmd/arm/clean_test.go` - Command unit tests
- `internal/cleaner/cleaner_test.go` - Cleaner unit tests

## Example Output
```
Analyzing cache usage...

Cache Summary:
Total size: 245 MB
Unused entries: 12 rulesets (89 MB)
Expired metadata: 5 entries (1.2 MB)

Would remove:
- typescript-rules@1.1.0 (15 MB)
- security-rules@1.9.0 (22 MB)
- react-rules@0.4.0 (8 MB)
...

Continue? [y/N]
```

## Notes
- Consider cache retention policies
- Plan for automatic cleanup scheduling
- Implement cache corruption detection and repair
