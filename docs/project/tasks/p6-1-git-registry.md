# P6.1 - Git Repository Registry Implementation

**Status**: 📋 **PLANNED**
**Phase**: 6 - Extended Registry Support
**Priority**: Medium
**Estimated Effort**: 3-4 days

## Overview

Implement git repository registry support to allow direct installation of rulesets from git repositories with glob pattern file selection and flexible reference targeting (branches, commits, tags).

## Requirements

### Core Functionality

- **Repository Access**: Support public and private git repositories with token authentication
- **Reference Types**: Support branches, commits (full/short SHA), and semver tags
- **File Selection**: Use glob patterns to select specific files from repositories
- **Version Management**: Different update behaviors based on reference type
- **File Structure**: Maintain repository directory structure in installed files

### Reference Behavior

| Reference Type | Example | Update Behavior |
|---------------|---------|-----------------|
| Branch | `main`, `develop` | Auto-update on new commits |
| Commit | `abc1234`, `a1b2c3d4e5f6` | Never auto-update (pinned) |
| Tag | `v1.0.0`, `1.0.0` | Update to compatible semver versions |
| Default | (none) | Use HEAD of default branch |

### Configuration Format

**.armrc**:
```ini
[sources]
awesome-rules = https://github.com/PatrickF1/awesome-cursorrules
company-rules = https://github.com/company/internal-rules

[sources.awesome-rules]
type = git
# No authToken for public repos

[sources.company-rules]
type = git
authToken = $COMPANY_GITHUB_TOKEN
```

**rules.json**:
```json
{
  "dependencies": {
    "awesome-rules@main": {
      "patterns": ["rules/*.md", "docs/*.txt"]
    },
    "company-rules@v2.1.0": {
      "patterns": ["security/**", "typescript/**"]
    }
  }
}
```

## Technical Implementation

### 1. Git Registry Type

```go
type GitRegistry struct {
    URL       string
    AuthToken string
    client    *git.Client
}

func (r *GitRegistry) GetVersions(name string) ([]string, error)
func (r *GitRegistry) DownloadPackage(name, version string) ([]byte, error)
func (r *GitRegistry) GetMetadata(name string) (*Metadata, error)
```

### 2. Reference Parsing

```go
type GitReference struct {
    Type   ReferenceType // Branch, Commit, Tag
    Value  string
    Parsed ParsedReference
}

type ReferenceType int
const (
    RefTypeBranch ReferenceType = iota
    RefTypeCommit
    RefTypeTag
    RefTypeDefault
)
```

### 3. File Pattern Matching

```go
type PatternMatcher struct {
    patterns []string
}

func (pm *PatternMatcher) MatchFiles(files []string) []string
func ParsePatterns(input string) []string // "*.md,docs/*.txt" -> ["*.md", "docs/*.txt"]
```

### 4. Version Discovery

- **Branches**: Use git ls-remote to check for new commits
- **Tags**: List all tags, filter for valid semver, sort by version
- **Commits**: No version discovery (pinned)

### 5. Update Logic

```go
func (r *GitRegistry) CheckForUpdates(dep Dependency) (*UpdateInfo, error) {
    switch dep.Reference.Type {
    case RefTypeBranch:
        return r.checkBranchUpdates(dep)
    case RefTypeTag:
        return r.checkTagUpdates(dep)
    case RefTypeCommit:
        return nil, nil // No updates for pinned commits
    }
}
```

## Implementation Tasks

### Phase 1: Core Git Operations
- [ ] Implement GitRegistry struct and interface
- [ ] Add git client wrapper (using go-git library)
- [ ] Implement reference parsing and validation
- [ ] Add authentication support (token-based)

### Phase 2: File Operations
- [ ] Implement glob pattern matching
- [ ] Add file extraction from git repositories
- [ ] Maintain directory structure in extracted files
- [ ] Handle multiple pattern support

### Phase 3: Version Management
- [ ] Implement branch-based version discovery
- [ ] Add semver tag parsing and filtering
- [ ] Implement commit-based pinning
- [ ] Add default reference handling (HEAD of default branch)

### Phase 4: Integration
- [ ] Integrate with existing registry system
- [ ] Update configuration parsing for git sources
- [ ] Add git registry to dependency resolution
- [ ] Update install/update/outdated commands

### Phase 5: Testing & Documentation
- [ ] Unit tests for git operations
- [ ] Integration tests with real repositories
- [ ] Update user documentation
- [ ] Add troubleshooting guides

## Dependencies

### External Libraries
- **go-git/go-git**: Git operations in pure Go
- **gobwas/glob**: Glob pattern matching
- **Masterminds/semver**: Semantic version parsing

### Internal Dependencies
- Registry interface (existing)
- Configuration system (existing)
- Cache system (existing)
- File operations (existing)

## Configuration Examples

### Public Repository (awesome-cursorrules)
```ini
[sources]
awesome-rules = https://github.com/PatrickF1/awesome-cursorrules

[sources.awesome-rules]
type = git
```

### Private Company Repository
```ini
[sources]
company-rules = https://github.com/company/coding-standards

[sources.company-rules]
type = git
authToken = $COMPANY_GITHUB_TOKEN
```

### Mixed Dependencies
```json
{
  "dependencies": {
    "typescript-rules": "^1.0.0",
    "awesome-rules@main": {
      "patterns": ["rules/typescript-*.md", "rules/react-*.md"]
    },
    "company-rules@v2.1.0": {
      "patterns": ["security/**", "performance/**"]
    },
    "experimental@abc1234": {
      "patterns": ["experimental/*.md"]
    }
  }
}
```

## Error Handling

### Authentication Errors
- Invalid or expired tokens
- Repository access denied
- Rate limiting

### Repository Errors
- Repository not found
- Invalid references (branch/tag/commit)
- Network connectivity issues

### Pattern Matching Errors
- Invalid glob patterns
- No files match patterns
- Pattern conflicts

## Performance Considerations

### Caching Strategy
- Cache repository metadata (branches, tags)
- Cache file listings for specific references
- Reuse cloned repositories when possible

### Optimization Opportunities
- Shallow clones for specific commits/tags
- Sparse checkout for large repositories
- Parallel pattern matching

## Testing Strategy

### Unit Tests
- Reference parsing and validation
- Pattern matching logic
- Version comparison and sorting
- Authentication handling

### Integration Tests
- Real repository operations (public repos)
- End-to-end install/update workflows
- Error scenarios and recovery

### Performance Tests
- Large repository handling
- Multiple pattern performance
- Concurrent operations

## Documentation Updates

### User Documentation
- [x] Git registry configuration guide
- [x] Pattern syntax documentation
- [x] Authentication setup
- [x] Troubleshooting guide

### Technical Documentation
- [ ] Architecture decision record
- [ ] API documentation
- [ ] Performance benchmarks
- [ ] Security considerations

## Success Criteria

- [ ] Install rulesets from public git repositories
- [ ] Install rulesets from private repositories with authentication
- [ ] Support all reference types (branch, commit, tag, default)
- [ ] Pattern matching works with complex glob patterns
- [ ] Update behavior matches reference type expectations
- [ ] Performance acceptable for typical repository sizes
- [ ] Comprehensive error handling and user feedback
- [ ] Documentation complete and accurate

## Future Enhancements

### Phase 2 Improvements
- SSH key authentication support
- Git submodule support
- Monorepo path targeting
- Custom git providers (GitLab, Bitbucket)

### Advanced Features
- Repository mirroring for performance
- Webhook-based update notifications
- Branch protection and security scanning
- Integration with git hosting APIs
