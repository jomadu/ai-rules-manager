# AI Rules Manager (ARM) Implementation Tasks

## Relevant Files

- `cmd/arm/main.go` - Main entry point for the ARM CLI application
- `cmd/arm/main_test.go` - Unit tests for main CLI entry point
- `go.mod` - Go module definition with dependencies
- `go.sum` - Go module checksums
- `internal/config/config.go` - Configuration management (INI and JSON parsing, validation)
- `internal/config/config_test.go` - Unit tests for configuration management
- `internal/registry/registry.go` - Registry interface and common functionality
- `internal/registry/registry_test.go` - Unit tests for registry interface
- `internal/registry/git.go` - Git registry implementation
- `internal/registry/git_test.go` - Unit tests for Git registry
- `internal/registry/s3.go` - S3 registry implementation
- `internal/registry/s3_test.go` - Unit tests for S3 registry
- `internal/registry/gitlab.go` - GitLab registry implementation
- `internal/registry/gitlab_test.go` - Unit tests for GitLab registry
- `internal/registry/https.go` - HTTPS registry implementation
- `internal/registry/https_test.go` - Unit tests for HTTPS registry
- `internal/registry/local.go` - Local filesystem registry implementation
- `internal/registry/local_test.go` - Unit tests for local registry
- `internal/cli/commands.go` - CLI command definitions and routing
- `internal/cli/commands_test.go` - Unit tests for CLI commands
- `internal/cache/cache.go` - Caching system for registry data and downloads
- `internal/cache/cache_test.go` - Unit tests for caching system
- `internal/version/resolver.go` - Version resolution and semantic versioning logic
- `internal/version/resolver_test.go` - Unit tests for version resolution
- `internal/install/installer.go` - Package installation and file management
- `internal/install/installer_test.go` - Unit tests for package installer
- `go.mod` - Go module definition with dependencies
- `go.sum` - Go module checksums

### Notes

- Unit tests should be placed alongside the code files they are testing in the same directory
- Use `go test ./...` to run all tests, or `go test ./internal/config` to run specific package tests
- The project uses Go 1.23+ as specified in .golangci.yml
- Follow the existing Makefile targets for building, testing, and linting

## Tasks

- [x] 1.0 Core Infrastructure Setup
  - [x] 1.1 Initialize Go module with required dependencies (cobra, viper, aws-sdk-go-v2, go-git)
  - [x] 1.2 Create internal package structure (config, registry, cli, cache, version, install)
  - [x] 1.3 Set up main.go entry point with basic CLI framework
  - [x] 1.4 Configure build system and update Makefile for ARM binary
- [x] 2.0 Configuration System Implementation
  - [x] 2.1 Implement INI parser for .armrc files with environment variable expansion
  - [x] 2.2 Implement JSON parser for arm.json and arm.lock files
  - [x] 2.3 Create hierarchical configuration merger (global + local)
  - [x] 2.4 Add configuration validation with registry type checking
  - [x] 2.5 Implement stub file generation for .armrc and arm.json
- [ ] 3.0 Registry System Implementation
  - [x] 3.1 Define registry interface and common authentication handling
  - [ ] 3.2 Implement Git registry with clone/API modes and pattern matching
  - [ ] 3.3 Implement S3 registry with AWS credential chain and region handling
  - [ ] 3.4 Implement GitLab registry with package API integration
  - [ ] 3.5 Implement HTTPS registry with manifest.json discovery
  - [ ] 3.6 Implement Local filesystem registry with directory scanning
- [ ] 4.0 CLI Command Interface Implementation
  - [ ] 4.1 Implement config command (set, get, list, add/remove registry/channel)
  - [ ] 4.2 Implement install command with stub generation and pattern support
  - [ ] 4.3 Implement search, info, and list commands with registry filtering
  - [ ] 4.4 Implement update, outdated, and uninstall commands
  - [ ] 4.5 Implement clean command and version/help utilities
- [ ] 5.0 Package Management and Installation System
  - [ ] 5.1 Implement semantic version resolution with range operators
  - [ ] 5.2 Create caching system with TTL and LRU eviction
  - [ ] 5.3 Build file installer with ARM namespace directory structure
  - [ ] 5.4 Implement lock file management and dependency resolution
  - [ ] 5.5 Add parallel processing with rate limiting and progress indication
