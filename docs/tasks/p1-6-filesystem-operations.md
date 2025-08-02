# P1.6: File System Operations

## Overview
Implement core file system operations for managing target directories, cache, and atomic file operations across platforms.

## Requirements
- ✅ Create target directory structure (now configuration-driven via rules.json)
- Implement atomic file operations
- Handle file permissions and cross-platform paths
- Create cache management system (.arm/cache/)

## Tasks
- [x] **Target directory management**:
  - ✅ Configuration-driven target directories (via rules.json)
  - ✅ Handle nested source/ruleset/version directories
  - ✅ Support for custom target paths
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
- [x] Target directories are created correctly on all platforms (configuration-driven)
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
- ✅ Target directories now use configuration from rules.json instead of hardcoded paths
- Consider using file locks for concurrent access
- Implement proper cleanup of temporary files
- Plan for future backup/restore functionality
- Handle symbolic links appropriately

## Completed Work
- ✅ Replaced hardcoded .cursorrules and .amazonq/rules paths with manifest-driven targets
- ✅ Added GetDefaultTargets() function for configurable defaults
- ✅ Updated installer and uninstaller to read from rules.json
- ✅ Added tests and documentation for configuration-driven approach
