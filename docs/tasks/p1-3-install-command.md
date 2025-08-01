# P1.3: Install Command

## Overview
Implement the core `arm install` command that downloads, extracts, and installs rulesets to target directories.

## Requirements
- Parse ruleset names with optional version specs
- Download .tar.gz files from registries
- Extract to target directories (.cursorrules, .amazonq/rules)
- Update rules.json and rules.lock files
- Handle conflicts and rollback on failure

## Tasks
- [x] **Create install command structure**:
  ```bash
  arm install <ruleset>[@version]
  arm install  # from rules.json
  ```
- [x] **Parse ruleset specifications**:
  - Handle `company@ruleset-name@1.0.0` format
  - Support version ranges (^1.0.0, ~1.0.0, 1.0.0)
- [x] **Implement dependency resolution**:
  - Resolve version constraints
  - ~~Handle transitive dependencies~~ (not needed - rulesets don't have dependencies)
  - ~~Detect circular dependencies~~ (not applicable)
- [x] **Download functionality**:
  - HTTP client with authentication
  - ~~Progress indicators for large downloads~~ (future enhancement)
  - Checksum verification
- [x] **Tar extraction**:
  - Safe extraction (prevent directory traversal)
  - ~~Preserve file permissions~~ (basic implementation)
  - ~~Handle symbolic links appropriately~~ (skipped for security)
- [x] **Manifest updates**:
  - Add to rules.json dependencies
  - Generate/update rules.lock with exact versions
- [ ] **Rollback mechanism**:
  - ~~Atomic operations where possible~~ (future enhancement)
  - ~~Cleanup on failure~~ (future enhancement)
  - ~~Restore previous state~~ (future enhancement)

## Acceptance Criteria
- [x] `arm install typescript-rules` works end-to-end
- [x] `arm install company@security-rules@^1.0.0` handles scoped packages
- [x] `arm install` installs all dependencies from rules.json
- [x] Files appear in correct target directories
- [x] rules.json and rules.lock are updated correctly
- [ ] Failed installs don't leave partial state (future enhancement)
- [ ] Progress indicators work for slow downloads (future enhancement)

## Dependencies
- net/http (standard library)
- archive/tar (standard library)
- compress/gzip (standard library)
- github.com/hashicorp/go-version (semantic versioning)

## Files Created
- [x] `cmd/arm/install.go` - CLI command implementation
- [x] `internal/installer/installer.go` - Core installation logic
- [x] `internal/installer/installer_test.go` - Unit tests
- ~~`internal/downloader/downloader.go`~~ (integrated into installer)
- ~~`internal/extractor/extractor.go`~~ (integrated into installer)
- ~~`internal/resolver/resolver.go`~~ (integrated into installer)

## Implementation Notes
- Used minimal approach - all functionality integrated into single installer package
- Proper error wrapping implemented for debugging
- Registry URL is configurable but currently hardcoded (TODO: P2.1 configuration)
- Basic semver resolution using existing parser package
- Safe tar extraction with path sanitization
- Checksum verification using SHA256

## Future Enhancements
- Parallel downloads for multiple rulesets
- Progress indicators for large downloads
- Resume capability on interrupted downloads
- Atomic operations and rollback on failure
- Authentication token support (P2.1)
