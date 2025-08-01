# P1.1: Project Setup

## Overview
Initialize the Go project structure and set up core development tools for the MRM CLI application.

## Requirements
- Go module with proper naming and structure
- Cobra CLI framework for command handling
- Viper for configuration management
- Standard Go project layout
- Cross-compilation CI/CD pipeline

## Tasks
- [ ] **Initialize Go module**: `go mod init github.com/user/mrm`
- [ ] **Install Cobra CLI**: Add cobra dependency and generate basic structure
- [ ] **Install Viper**: Add viper for configuration management
- [ ] **Create project structure**:
  ```
  cmd/
    mrm/
      main.go
  internal/
    config/
    registry/
    cache/
  pkg/
    types/
  ```
- [ ] **Set up CI/CD**: GitHub Actions for cross-compilation (linux, darwin, windows)

## Acceptance Criteria
- [ ] `go build` produces working binary
- [ ] Basic `mrm --help` command works
- [ ] CI builds binaries for all target platforms
- [ ] Project follows Go standard layout

## Dependencies
- Go 1.21+
- github.com/spf13/cobra
- github.com/spf13/viper

## Files to Create
- `cmd/mrm/main.go`
- `internal/config/config.go`
- `.github/workflows/build.yml`
- `go.mod`, `go.sum`

## Notes
- Use semantic versioning from start
- Set up proper module path for future distribution
- Consider using goreleaser for release automation