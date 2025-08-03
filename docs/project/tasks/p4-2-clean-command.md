# P4.2: Clean Command

## Overview
Implement the `arm clean` command to remove unused cached rulesets, orphaned entries, and expired metadata.

## Requirements
- Remove unused cached rulesets
- Clean orphaned cache entries
- Clear expired metadata cache
- Provide cache size reporting

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

## Acceptance Criteria
- [ ] `arm clean` removes only unused cache entries
- [ ] Cache size is reported accurately
- [ ] --dry-run shows cleanup plan without executing
- [ ] Currently used rulesets are never removed
- [ ] Empty directories are cleaned up
- [ ] User confirmation for large cleanups
- [ ] Progress indication for long operations

## Dependencies
- os (standard library)
- path/filepath (standard library)

## Files to Create
- `cmd/arm/clean.go`
- `internal/cleaner/cleaner.go`
- `internal/cleaner/analyzer.go`

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
