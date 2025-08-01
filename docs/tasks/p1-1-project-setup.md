# P1.1: Project Setup

## Overview
Initialize the Go project structure and set up core development tools for the ARM CLI application.

## Requirements
- Go module with proper naming and structure
- Cobra CLI framework for command handling
- Viper for configuration management
- Standard Go project layout
- Cross-compilation CI/CD pipeline
- Conventional commit workflow setup

## Tasks
- [x] **Initialize Go module**: `go mod init github.com/jomadu/arm`
- [x] **Install Cobra CLI**: Add cobra dependency and generate basic structure
- [x] **Install Viper**: Add viper for configuration management
- [x] **Create project structure**:
  ```
  cmd/
    arm/
      main.go
  internal/
    config/
    registry/
    cache/
  pkg/
    types/
  ```
- [x] **Set up CI/CD**: GitHub Actions for cross-compilation (linux, darwin, windows)
- [ ] **Configure conventional commits**: Set up commit message linting and release automation

## Acceptance Criteria
- [x] `go build` produces working binary
- [x] Basic `arm --help` command works
- [x] CI builds binaries for all target platforms
- [x] Project follows Go standard layout

## Dependencies
- Go 1.21+
- github.com/spf13/cobra
- github.com/spf13/viper

## Files to Create
- `cmd/arm/main.go`
- `internal/config/config.go`
- `.github/workflows/build.yml`
- `go.mod`, `go.sum`

## Notes
- Use semantic versioning from start
- Set up proper module path for future distribution
- Use conventional commits for automated releases
- Configure commit message linting (commitlint)
- Set up automated semantic versioning