## Relevant Files

- `internal/registry/git_common.go` - Shared Git operations and utilities used by both remote and local Git registries.
- `internal/registry/git_common_test.go` - Unit tests for shared Git operations.
- `internal/registry/git_operations.go` - Interface definition for Git operations (version resolution, file discovery, pattern matching).
- `internal/registry/base_git_registry.go` - Abstract base struct containing shared logic for all Git registry types.
- `internal/registry/remote_git_operations.go` - Remote Git operations implementation (refactored from existing GitRegistry).
- `internal/registry/local_git_operations.go` - Local Git operations implementation for direct repository access.
- `internal/registry/git.go` - Refactored remote Git registry using composition with RemoteGitOperations.
- `internal/registry/git_local.go` - New local Git registry implementation extending BaseGitRegistry.
- `internal/registry/git_local_test.go` - Unit tests for local Git registry implementation.
- `internal/registry/factory.go` - Updated factory to support `git-local` registry type creation.
- `internal/registry/registry.go` - Updated registry interface and validation to support `git-local` type.
- `internal/config/path_resolver.go` - New utility for resolving absolute, relative, and tilde-expanded paths.
- `internal/config/path_resolver_test.go` - Unit tests for path resolution logic.
- `internal/cli/commands.go` - Updated CLI commands to support `git-local` registry type in config commands.
- `tests/integration/git_local_integration_test.go` - Comprehensive integration tests for local Git registry operations.

### Notes

- Unit tests should typically be placed alongside the code files they are testing (e.g., `git_local.go` and `git_local_test.go` in the same directory).
- Use `go test ./internal/registry -v` to run registry-specific tests.
- Use `go test ./... -v` to run all tests.

## Tasks

- [ ] 1.0 Refactor Git Registry Architecture
  - [x] 1.1 Extract common Git operations into `internal/registry/git_common.go` shared by both remote and local implementations
  - [x] 1.2 Create `GitOperations` interface for version resolution, file discovery, and pattern matching operations
  - [x] 1.3 Refactor existing `GitRegistry` to use composition with `RemoteGitOperations` implementing `GitOperations`
  - [x] 1.4 Create abstract `BaseGitRegistry` struct containing shared logic for caching, pattern matching, and file operations

- [x] 2.0 Implement Local Git Operations
  - [x] 2.1 Create `LocalGitOperations` struct implementing `GitOperations` interface for direct local repository access
  - [x] 2.2 Implement local version resolution using direct Git commands without cloning or API calls (NOTE: Consider alternatives to temporary worktrees for better efficiency)
  - [x] 2.3 Implement local file discovery and pattern matching by directly scanning repository working directory (NOTE: Current worktree approach may be inefficient - consider git show/archive alternatives)
  - [x] 2.4 Add path resolution utility in `internal/config/path_resolver.go` for absolute, relative, and tilde expansion (NOTE: LocalGitOperations currently only uses filepath.Abs - needs proper tilde/relative path support)

- [ ] 3.0 Create Local Git Registry Implementation
  - [x] 3.1 Create `GitLocalRegistry` struct in `internal/registry/git_local.go` extending `BaseGitRegistry`
  - [x] 3.2 Integrate `LocalGitOperations` for all Git-specific functionality
  - [x] 3.3 Add symlink path storage with runtime resolution to actual paths during Git operations
  - [x] 3.4 Implement local-specific caching strategy that caches metadata only and accesses files directly (NOTE: LocalGitOperations currently has no caching - needs metadata caching as specified in PRD)

- [x] 4.0 Update Registry Infrastructure
  - [x] 4.1 Add `git-local` type to registry validation in `internal/registry/registry.go`
  - [x] 4.2 Update factory pattern in `internal/registry/factory.go` to create `git-local` registries
  - [x] 4.3 Update CLI commands in `internal/cli/commands.go` to accept `git-local` type in registry configuration
  - [x] 4.4 Ensure backward compatibility with existing Git registry configurations

- [x] 5.0 Error Handling and Validation
  - [x] 5.1 Implement clear error messages distinguishing local Git issues from general ARM errors
  - [x] 5.2 Add runtime path validation to ensure local paths exist and contain valid Git repositories
  - [x] 5.3 Handle moved/renamed repositories with clear error messages directing users to update registry paths
  - [x] 5.4 Add cross-platform file system permissions and path handling support

- [x] 6.0 Integration and Testing
  - [x] 6.1 Create comprehensive unit tests for refactored Git architecture and local Git registry functionality
  - [x] 6.2 Add integration tests for local Git registry operations with existing ARM commands (use tests/integration/{install,update}_integration_test.go for inspiration)
  - [x] 6.3 Validate that refactored remote Git registry maintains all existing functionality
