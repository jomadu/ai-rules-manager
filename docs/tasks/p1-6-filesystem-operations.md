# P1.6: File System Operations

## Overview
Implement core file system operations for managing target directories, cache, and atomic file operations across platforms.

## Requirements
- Create target directory structure (.cursorrules, .amazonq/rules)
- Implement atomic file operations
- Handle file permissions and cross-platform paths
- Create cache management system (.mpm/cache/)

## Tasks
- [ ] **Target directory management**:
  - Create `.cursorrules/lrm/` structure
  - Create `.amazonq/rules/lrm/` structure
  - Handle nested source/ruleset/version directories
- [ ] **Atomic operations**:
  - Atomic file writes (write to temp, then rename)
  - Atomic directory operations
  - Transaction-like behavior for multi-file operations
- [ ] **Cross-platform compatibility**:
  - Handle Windows vs Unix path separators
  - Manage file permissions appropriately
  - Deal with case-sensitive vs case-insensitive filesystems
- [ ] **Cache directory structure**:
  ```
  .mpm/
    cache/
      <source>/
        <ruleset>/
          <version>/
            ruleset.tar.gz
            metadata.json
  ```
- [ ] **File operations utilities**:
  - Safe file copying
  - Directory traversal
  - Cleanup operations
  - Permission handling

## Acceptance Criteria
- [ ] Target directories are created correctly on all platforms
- [ ] File operations are atomic and safe
- [ ] Cache structure is consistent and organized
- [ ] Proper error handling for permission issues
- [ ] Cross-platform path handling works correctly
- [ ] Cleanup operations remove empty directories

## Dependencies
- os (standard library)
- path/filepath (standard library)
- io/fs (standard library)

## Files to Create
- `internal/filesystem/operations.go`
- `internal/filesystem/atomic.go`
- `internal/filesystem/cache.go`
- `internal/filesystem/paths.go`

## Notes
- Consider using file locks for concurrent access
- Implement proper cleanup of temporary files
- Plan for future backup/restore functionality
- Handle symbolic links appropriately
