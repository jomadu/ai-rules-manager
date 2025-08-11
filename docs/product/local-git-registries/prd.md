# Local Git Registry Support

## Introduction/Overview

ARM currently supports remote Git repositories as registries for sharing AI coding rulesets. However, developers need the ability to work with local Git repositories for testing rulesets before publishing, working in air-gapped environments, supporting development workflows with private repositories, and enabling integration testing in CI/CD pipelines. This feature extends ARM's existing Git registry functionality to support local file system Git repositories using a new `git-local` registry type.

## Goals

1. Enable developers to add local Git repositories as ARM registries
2. Support installation of rulesets from local Git repositories
3. Maintain consistency with existing ARM registry interface and commands
4. Provide flexible version resolution for local development workflows
5. Ensure reliable error handling for local Git operations

## User Stories

- **As a developer testing rulesets locally**, I want to add my local Git repository as an ARM registry so that I can test ruleset installations before publishing to remote registries.

- **As a developer in an air-gapped environment**, I want to use local Git repositories as ARM registries so that I can manage rulesets without internet connectivity.

- **As a team lead**, I want to configure local Git repositories for internal rulesets so that my team can share rules through our internal Git infrastructure.

- **As a developer working on multiple projects**, I want to install rulesets from both local and remote registries so that I can use a mix of development and production rulesets.

- **As a CI/CD engineer**, I want to test Git-based registry functionality using local repositories in automated pipelines so that I can validate ARM behavior without external dependencies or network access.

## Functional Requirements

1. ARM must support a new registry type `git-local` for local Git repositories
2. ARM must use the same configuration format as remote Git registries with runtime path validation
3. ARM must validate that local paths exist and contain valid Git repositories during registry operations
4. ARM must support all path types (absolute, relative from current directory, relative from ARM config directory) with clear resolution rules
5. ARM must resolve versions using explicit user specification in arm.json (tags, branches, commit hashes) following same logic as remote Git repositories
6. ARM must use hybrid caching strategy (cache metadata, direct file access) for optimal performance
7. ARM must handle local Git repositories independently of workspace Git repositories without conflicts
8. ARM must support mixed local and remote dependencies in ruleset installations
9. ARM must fail fast with clear error messages when local Git operations fail
10. ARM must integrate local Git registries with existing ARM commands (install, list, etc.) without requiring separate command syntax

## Non-Goals (Out of Scope)

- Converting between local and remote Git repositories
- Git repository initialization or creation
- Git authentication mechanisms for local repositories
- Automatic syncing of local repositories with remote counterparts

## Design Considerations

The feature should maintain ARM's existing command-line interface patterns:

```bash
# Registry configuration
arm config add registry local-dev /path/to/repo --type=git-local

# Installation (requires patterns for Git registries)
arm install local-dev/my-rules --patterns "rules/*.md,**/*.mdc"
arm install local-dev/my-rules@main --patterns "rules/*.md"
arm list --registry=local-dev
```

Configuration format should mirror existing Git registries:

```ini
[registries]
local-dev = /path/to/local/repo

[registries.local-dev]
type = git-local
```

## Technical Considerations

- Evaluate existing Git registry implementation for extension vs complete re-architecture based on best practice design principles
- Consider unified Git registry abstraction that handles both local and remote repositories through common interface
- If re-architecture is warranted, design modular Git registry system with pluggable local/remote handlers
- Implement path resolution logic that handles absolute, relative, and tilde-expanded paths
- Use existing ARM cache structure (~/.arm/cache/registries/) for metadata caching while accessing files directly from local repository path
- Cache version discovery results (tags, branches) but read actual files directly from local Git repository without cloning
- Apply same TTL and eviction policies as remote Git registries to cached metadata
- Ensure error messages clearly distinguish between local Git issues and general ARM errors
- Consider file system permissions and cross-platform path handling
- Maintain backward compatibility with existing Git registry configurations during any architectural changes

## Success Metrics

- Developers can successfully configure local Git repositories as ARM registries
- Installation success rate from local Git registries matches remote Git registries (>95%)
- Local Git registry operations complete within 2x the time of equivalent remote operations
- Zero breaking changes to existing ARM registry functionality
- Clear error messages reduce support requests related to local Git configuration

## Open Questions

1. ~~Should ARM automatically detect Git repository type (local vs remote) based on path patterns, or require explicit type specification?~~
   **Resolved:** Require explicit type specification (`--type=git-local`)

2. ~~What should be the default version resolution priority when user doesn't specify preferences?~~
   **Resolved:** Follow same logic as remote Git repositories (tags > branches > HEAD with user-configurable priority)

3. ~~How should ARM handle symbolic links in local Git repository paths?~~
   **Resolved:** Store symlink path in configuration but resolve to actual path during Git operations

4. ~~Should there be size limits or performance warnings for large local Git repositories?~~
   **Resolved:** No size limits or warnings in initial implementation

5. ~~How should ARM handle local Git repositories that are moved or renamed after configuration?~~
   **Resolved:** Fail immediately with clear error message directing user to update registry path
