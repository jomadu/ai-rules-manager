# ADR-002: Git Repository Registry Design

**Status**: Accepted
**Date**: 2024-01-15
**Deciders**: Development Team

## Context

ARM needs to support installing rulesets directly from git repositories (like awesome-cursorrules) with flexible file selection and reference targeting. This requires decisions on:

1. Registry type naming and configuration
2. Performance optimization strategies
3. Cache structure and cleanup
4. API vs git operations

## Decision

### 1. Registry Type Structure

**Decision**: Use `type = git` as primary registry type with optional `api` field for optimization.

```ini
[sources.awesome-rules]
type = git
url = https://github.com/PatrickF1/awesome-cursorrules
api = github  # Optional: enables GitHub API optimization
authToken = $GITHUB_TOKEN

[sources.company-gitlab]
type = git
url = https://gitlab.company.com/team/rules-repo
api = gitlab  # Optional: enables GitLab API optimization
authToken = $GITLAB_TOKEN

[sources.generic-git]
type = git
url = https://git.company.com/repo.git
# No api field = fallback to git operations
authToken = $GIT_TOKEN
```

**Rationale**:
- Avoids namespace collision with existing `type = gitlab` (package registry)
- Clear separation: `git` = repositories, `gitlab` = package registries
- API optimization is optional enhancement, not core requirement
- Extensible for future providers (`api = bitbucket`, `api = azure-devops`)

### 2. Performance Optimization Strategy

**Decision**: API-first approach with git fallback.

**GitHub API Operations**:
```
GET /repos/owner/repo/git/trees/{sha}?recursive=1  # List files
GET /repos/owner/repo/contents/{path}?ref={branch} # Download files
GET /repos/owner/repo/branches/{branch}            # Get branch info
GET /repos/owner/repo/tags                         # Get tags
```

**GitLab API Operations**:
```
GET /projects/{id}/repository/tree?recursive=true&ref={branch}
GET /projects/{id}/repository/files/{path}/raw?ref={branch}
GET /projects/{id}/repository/branches
GET /projects/{id}/repository/tags
```

**Rationale**:
- 1000x+ performance improvement for sparse file selection
- Only download needed files instead of entire repository
- Fallback to git operations ensures compatibility with any git provider

### 3. Cache Structure

**Decision**: Separate cache directories for packages and git repositories.

```
~/.arm/cache/
  packages/                    # Existing package cache
    registry.armjs.org/
      typescript-rules/1.0.1/package.tar.gz
  git/                        # New git repository cache
    github.com/
      PatrickF1/awesome-cursorrules/
        .git/                 # Bare repository
        metadata.json         # Cache metadata
    gitlab.company.com/
      team/rules-repo/
        .git/
        metadata.json
```

**Rationale**:
- Different access patterns: packages (download once) vs git (clone, fetch, checkout)
- Different cleanup logic: packages by age/size vs git by age/usage
- Clear boundaries for developers and tooling
- Bare repositories for space efficiency

### 4. Cache Cleanup Strategy

**Decision**: Simple time-based cleanup with unified `--cache` flag.

```bash
arm clean              # Remove unused project files (existing behavior)
arm clean --cache      # Remove ALL cache (packages + git repos)
arm clean --dry-run    # Show what would be cleaned
```

**Rationale**:
- Rejected complex "unused repository detection" as overly complicated
- `--cache` flag removes entire cache directory (packages + git)
- Simple time-based cleanup (30+ days) for automatic maintenance
- User confirmation required for destructive operations

### 5. Reference Types and Versioning

**Decision**: Support branches, commits, and semver tags with different update behaviors.

| Reference Type | Example | Update Behavior |
|---------------|---------|-----------------|
| Branch | `main`, `develop` | Auto-update on new commits |
| Commit | `abc1234`, `a1b2c3d4e5f6` | Never auto-update (pinned) |
| Tag | `v1.0.0`, `1.0.0` | Update to compatible semver versions |
| Default | (none) | Use HEAD of default branch |

**Configuration**:
```json
{
  "dependencies": {
    "awesome-rules@main": {
      "patterns": ["rules/*.md", "docs/*.txt"]
    },
    "company-rules@v2.1.0": {
      "patterns": ["security/**", "typescript/**"]
    },
    "experimental@abc1234": {
      "patterns": ["experimental/*.md"]
    }
  }
}
```

**Rationale**:
- Flexible reference targeting for different use cases
- Clear update semantics based on reference type
- Semver tag filtering (ignore non-semver tags like `beta`, `release-2023`)
- Support both `v1.0.0` and `1.0.0` tag formats

## Consequences

### Positive
- Clear separation between package registries and git repositories
- Significant performance improvements for API-supported providers
- Simple and predictable cache cleanup
- Flexible reference targeting for different workflows
- Extensible design for future git providers

### Negative
- Additional complexity in registry implementation
- API rate limiting considerations for hosted providers
- Larger cache size due to git repository storage
- Token management for private repositories

### Neutral
- Requires documentation updates for new registry type
- Need to implement glob pattern matching
- Authentication limited to tokens initially (no SSH keys)

## Implementation Notes

### Phase 1: Core Implementation
- Implement `GitRegistry` with git operations fallback
- Add glob pattern matching for file selection
- Implement reference parsing and validation
- Add basic authentication support

### Phase 2: API Optimization
- Add GitHub API client integration
- Add GitLab API client integration
- Implement API-based file listing and downloading
- Add rate limiting and error handling

### Phase 3: Cache Integration
- Integrate git cache with existing cache system
- Update clean command to handle git repositories
- Add cache metadata tracking
- Implement time-based cleanup

## Alternatives Considered

### Alternative 1: Unified Cache Structure
```
~/.arm/cache/
  registry.armjs.org/typescript-rules/1.0.1/package.tar.gz
  github.com/PatrickF1/awesome-cursorrules/main/extracted-files/
```

**Rejected**: Different access patterns and cleanup logic make separation cleaner.

### Alternative 2: Git-Only Operations
Always use git clone/fetch operations without API optimization.

**Rejected**: Performance would be unacceptable for large repositories with sparse file selection.

### Alternative 3: Complex Usage Tracking
Track which git repositories are "active" across all projects for intelligent cleanup.

**Rejected**: Overly complex for minimal benefit. Simple time-based cleanup is sufficient.

### Alternative 4: Separate Registry Types
Use `type = github` and `type = gitlab` for git repositories.

**Rejected**: Creates confusion with existing package registry types and limits extensibility.
