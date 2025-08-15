# Build System

Build process and CI/CD pipeline for ARM.

## Quick Start

```bash
make build          # Single platform build
make build-all      # Multi-platform build
make test           # Run test suite
make lint           # Code quality checks
```

## Build Tools

- **Make** - Primary build automation
- **Go toolchain** - Compilation and testing
- **GitHub Actions** - CI/CD pipeline
- **GoReleaser** - Release automation

## Local Development

### Basic Build
```bash
# Development build
make build

# Multi-platform build
make build-all
# Outputs: bin/arm-{linux,darwin,windows}-{amd64,arm64}
```

### Build Configuration
Version injection via Makefile:
```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS = -X github.com/max-dunn/ai-rules-manager/internal/version.Version=$(VERSION)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/arm ./cmd/arm
```

## CI/CD Pipeline

### GitHub Actions
Two main workflows:

1. **Build & Test** (on push/PR):
   - Run tests with coverage
   - Lint code
   - Build all platforms
   - Cache Go modules

2. **Release** (on tag push):
   - Use GoReleaser for automated releases
   - Generate checksums
   - Create GitHub releases

### GoReleaser Configuration
```yaml
builds:
  - binary: arm
    main: ./cmd/arm
    env: [CGO_ENABLED=0]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags: [-s -w -X ...Version={{.Version}}]

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
```

## Quality Gates

### Pre-build Checks
```makefile
check: lint test
lint: golangci-lint run
test: go test -race -coverprofile=coverage.out ./...
```

### Requirements
- 80% test coverage minimum
- All linter checks pass
- Security scan with gosec
- Binary validation post-build

## Distribution

### Installation Script
Cross-platform installer:
```bash
# Auto-detect platform and download latest release
curl -sSL https://raw.githubusercontent.com/max-dunn/ai-rules-manager/main/scripts/install.sh | bash
```

### Package Managers
- **Homebrew** formula for macOS/Linux
- **Chocolatey** package for Windows
- Direct binary downloads from GitHub releases

## Troubleshooting

```bash
# Clean build issues
go clean -modcache && go mod download
go clean -cache -testcache

# Debug builds
go build -v ./cmd/arm  # Verbose output
go build -x ./cmd/arm  # Show commands
```

## Versioning

Semantic versioning:
- **Major** (v1.0.0): Breaking changes
- **Minor** (v1.1.0): New features
- **Patch** (v1.1.1): Bug fixes
- **Pre-release** (v1.1.0-alpha.1): Development versions
