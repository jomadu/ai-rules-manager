# ADR-003: Patterns in Lock File for Git Registries

## Status
Accepted

## Context
ARM supports multiple registry types (Git, S3, HTTPS, GitLab, Local) with different content delivery mechanisms. Git registries require patterns to filter which files from a repository get installed, while other registry types serve pre-packaged rulesets.

For reproducible builds, we need to determine whether patterns should be stored in the lock file (`arm.lock`) alongside resolved versions.

## Decision
**Include patterns in lock file only for Git registries.**

### Rationale by Registry Type

**Git Registries:**
- Patterns are essential for determining installed content
- Without patterns, entire repository would be installed
- Different patterns = different installed files = different behavior
- Must be locked for reproducibility

**Other Registry Types (S3, HTTPS, GitLab Package Registry, Local):**
- Serve pre-packaged rulesets (tar.gz files)
- Patterns already applied during packaging
- Install-time patterns would be meaningless
- No patterns needed in lock file

## Implementation

### Lock File Structure
```json
{
  "rulesets": {
    "git-registry": {
      "my-rules": {
        "version": "latest",
        "resolved": "abc123...",
        "patterns": ["rules/*.md", "docs/*.md"],
        "type": "git"
      }
    },
    "s3-registry": {
      "packaged-rules": {
        "version": "1.0.0",
        "resolved": "1.0.0",
        "type": "s3"
      }
    }
  }
}
```

### Installation Logic
- Git registries: Use patterns from lock file (or manifest fallback)
- Other registries: Ignore patterns, install entire package

## Consequences

### Positive
- Reproducible builds for Git registries
- Clean separation of concerns by registry type
- Minimal lock file size for non-Git registries

### Negative
- Registry-type specific logic complexity
- Lock file structure varies by registry type

## Alternatives Considered
1. **Always include patterns**: Wasteful for non-Git registries
2. **Never include patterns**: Non-reproducible Git installs
3. **Separate lock files by type**: Unnecessary complexity
