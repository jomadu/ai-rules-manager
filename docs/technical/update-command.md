# Update Command Technical Documentation

## Overview

The `arm update` command provides version-aware updating of installed rulesets while respecting semantic version constraints defined in `rules.json`.

## Architecture

### Core Components

```
cmd/arm/update.go           # CLI command interface
internal/updater/updater.go # Core update logic
internal/updater/updater_test.go # Comprehensive tests
```

### Dependencies

- `github.com/hashicorp/go-version` - Semantic version parsing and constraint checking
- `github.com/schollz/progressbar/v3` - Progress reporting during operations
- Existing registry and config systems

## Command Interface

```bash
arm update                    # Update all outdated rulesets
arm update <ruleset-name>     # Update specific ruleset
arm update --dry-run          # Show planned updates without executing
```

## Implementation Details

### Version Constraint Handling

The updater uses `hashicorp/go-version` for robust semantic version handling:

```go
// Parse constraint from rules.json
constraints, err := version.NewConstraint(">= 1.0.0, < 2.0.0")

// Find latest valid version
for _, vStr := range availableVersions {
    v, err := version.NewVersion(vStr)
    if constraints.Check(v) && v.GreaterThan(currentVer) {
        // Valid update candidate
    }
}
```

### Update Process Flow

1. **Load Configuration**
   - Read `rules.lock` for installed rulesets
   - Read `rules.json` for version constraints

2. **Version Discovery**
   - Query registries for available versions
   - Filter by version constraints
   - Identify update candidates

3. **Backup Creation**
   - Create backup of current installation
   - Store in `.arm/cache/backups/`

4. **Update Execution**
   - Download new version
   - Install to target directories
   - Update `rules.lock`

5. **Error Handling**
   - Restore backup on failure
   - Continue with other rulesets
   - Report failures at end

### Progress Reporting

Uses `schollz/progressbar/v3` for user feedback:

```go
bar := progressbar.NewOptions(len(rulesets),
    progressbar.OptionSetDescription("Checking for updates"),
    progressbar.OptionSetWidth(50),
    progressbar.OptionShowCount(),
)
```

### Backup and Restore

The updater implements a simple backup strategy:

```go
// Backup structure: .arm/cache/backups/{name}/{version}/{target}/
backupDir := filepath.Join(u.cacheDir, "backups", name, version)

// Copy current installation
for _, target := range manifest.Targets {
    targetPath := filepath.Join(target, "arm", name, version)
    backupTargetPath := filepath.Join(backupDir, filepath.Base(target))
    copyDir(targetPath, backupTargetPath)
}
```

## Error Handling

### Failure Scenarios

1. **Network Errors** - Registry unreachable
2. **Version Conflicts** - No valid versions found
3. **File System Errors** - Backup/restore failures
4. **Authentication Errors** - Registry access denied

### Recovery Strategy

- Individual ruleset failures don't stop the overall process
- Automatic rollback to previous version on failure
- Clear error reporting with actionable messages
- Backup cleanup on successful updates

## Testing Strategy

### Test Coverage

- **Version Constraint Logic** - Various constraint scenarios
- **Backup/Restore Operations** - File system operations
- **Error Handling** - Network and file system failures
- **Integration** - End-to-end update scenarios

### Mock Implementation

Tests use a mock version checker to avoid external dependencies:

```go
func mockCheckRulesetUpdate(ruleset InstalledRuleset, availableVersions []string) UpdateResult {
    // Simulate version checking without registry calls
}
```

## Performance Considerations

### Optimization Strategies

- **Parallel Version Checking** - Check multiple rulesets concurrently
- **Registry Caching** - Cache version lists to reduce API calls
- **Incremental Updates** - Only update changed rulesets

### Current Limitations

- Sequential processing of updates
- No persistent caching of version information
- Full re-download of packages (no delta updates)

## Future Enhancements

### Planned Improvements

1. **Parallel Processing** - Concurrent update operations
2. **Smart Caching** - Persistent version cache
3. **Delta Updates** - Only download changed files
4. **Update Scheduling** - Automated update checks
5. **Rollback History** - Multiple backup versions

### Integration Points

- **Notification System** - Alert on available updates
- **CI/CD Integration** - Automated update workflows
- **Dependency Resolution** - Handle transitive dependencies

## Configuration

### Version Constraints

Supports standard semantic version constraints:

```json
{
  "dependencies": {
    "typescript-rules": ">= 1.0.0, < 2.0.0",
    "security-rules": "~2.1.0",
    "react-rules": "^1.5.0"
  }
}
```

### Registry Integration

Works with all supported registry types:
- GitLab Package Registry
- AWS S3 buckets
- Generic HTTP endpoints
- Local file system

## Monitoring and Debugging

### Logging

- Progress reporting during operations
- Detailed error messages with context
- Summary reporting of results

### Debug Information

- Version constraint evaluation
- Registry query results
- File system operation status
- Backup/restore operations

## Security Considerations

### Safe Operations

- Backup before modifications
- Atomic updates where possible
- Validation of downloaded content
- Secure temporary file handling

### Authentication

- Inherits registry authentication from config
- No credential storage in update process
- Secure token handling for registry access
