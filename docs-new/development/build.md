# Build System

Build process and CI/CD pipeline for ARM.

## Build Architecture

### Build Tools
- **Make** - Primary build automation
- **Go toolchain** - Compilation and testing
- **GitHub Actions** - CI/CD pipeline
- **GoReleaser** - Release automation

### Build Targets
```bash
make build          # Single platform build
make build-all      # Multi-platform build
make test           # Run test suite
make lint           # Code quality checks
make release        # Create release artifacts
```

## Local Build Process

### Development Build
```bash
# Quick development build
make build

# Build with race detection
make build-race

# Build for specific platform
GOOS=linux GOARCH=amd64 make build
```

### Multi-Platform Build
```bash
# Build for all supported platforms
make build-all

# Outputs:
# bin/arm-linux-amd64
# bin/arm-linux-arm64
# bin/arm-darwin-amd64
# bin/arm-darwin-arm64
# bin/arm-windows-amd64.exe
```

### Build Configuration
`Makefile` key sections:
```makefile
# Build variables
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse HEAD)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS = -X github.com/max-dunn/ai-rules-manager/internal/version.Version=$(VERSION) \
          -X github.com/max-dunn/ai-rules-manager/internal/version.Commit=$(COMMIT) \
          -X github.com/max-dunn/ai-rules-manager/internal/version.BuildTime=$(BUILD_TIME)

# Build targets
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/arm ./cmd/arm

build-all:
	$(foreach GOOS_GOARCH,$(PLATFORMS), \
		GOOS=$(word 1,$(subst -, ,$(GOOS_GOARCH))) \
		GOARCH=$(word 2,$(subst -, ,$(GOOS_GOARCH))) \
		CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" \
		-o bin/arm-$(GOOS_GOARCH)$(if $(filter windows,$(word 1,$(subst -, ,$(GOOS_GOARCH)))),.exe) \
		./cmd/arm;)
```

## CI/CD Pipeline

### GitHub Actions Workflow
`.github/workflows/build.yml`:
```yaml
name: Build and Test
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23.2'

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

      - name: Run tests
        run: make test

      - name: Run linter
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23.2'

      - name: Build all platforms
        run: make build-all

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: binaries
          path: bin/
```

### Release Pipeline
`.github/workflows/release.yml`:
```yaml
name: Release
on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v4
        with:
          go-version: '1.23.2'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v4
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Release Configuration

### GoReleaser Configuration
`.goreleaser.yml`:
```yaml
project_name: ai-rules-manager

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: arm
    binary: arm
    main: ./cmd/arm
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/max-dunn/ai-rules-manager/internal/version.Version={{.Version}}
      - -X github.com/max-dunn/ai-rules-manager/internal/version.Commit={{.Commit}}
      - -X github.com/max-dunn/ai-rules-manager/internal/version.BuildTime={{.Date}}

archives:
  - id: default
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    files:
      - README.md
      - LICENSE.txt

checksum:
  name_template: 'checksums.txt'

snapshot:
  name_template: "{{ incpatch .Version }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'

release:
  github:
    owner: max-dunn
    name: ai-rules-manager
  draft: false
  prerelease: auto
```

## Version Management

### Version Injection
Build-time version injection:
```go
// internal/version/version.go
package version

var (
    Version   = "dev"
    Commit    = "unknown"
    BuildTime = "unknown"
)

func GetVersion() string {
    if Version == "dev" {
        return "dev"
    }
    return Version
}

func GetBuildInfo() BuildInfo {
    return BuildInfo{
        Version:   Version,
        Commit:    Commit,
        BuildTime: BuildTime,
    }
}
```

### Semantic Versioning
- **Major** (v1.0.0): Breaking changes
- **Minor** (v1.1.0): New features, backward compatible
- **Patch** (v1.1.1): Bug fixes, backward compatible
- **Pre-release** (v1.1.0-alpha.1): Development versions

## Build Optimization

### Binary Size Optimization
```makefile
# Optimized build flags
LDFLAGS_RELEASE = -s -w -X main.version=$(VERSION)

build-release:
	CGO_ENABLED=0 go build \
		-ldflags "$(LDFLAGS_RELEASE)" \
		-trimpath \
		-o bin/arm ./cmd/arm
```

### Build Caching
```yaml
# GitHub Actions cache configuration
- name: Cache Go build cache
  uses: actions/cache@v3
  with:
    path: ~/.cache/go-build
    key: ${{ runner.os }}-go-build-${{ hashFiles('**/*.go') }}

- name: Cache Go modules
  uses: actions/cache@v3
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-mod-${{ hashFiles('**/go.sum') }}
```

## Quality Gates

### Pre-build Checks
```makefile
check: lint test
	@echo "All checks passed"

lint:
	golangci-lint run

test:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total | awk '{if($$3 < 80.0) exit 1}'

security:
	gosec ./...
	nancy sleuth
```

### Build Validation
```bash
# Validate binary after build
./bin/arm version
./bin/arm --help

# Test basic functionality
./bin/arm config list
```

## Distribution

### Package Managers

#### Homebrew Formula
```ruby
class AiRulesManager < Formula
  desc "Package manager for AI coding assistant rulesets"
  homepage "https://github.com/max-dunn/ai-rules-manager"
  url "https://github.com/max-dunn/ai-rules-manager/archive/v1.0.0.tar.gz"
  sha256 "..."
  license "GPL-3.0"

  depends_on "go" => :build

  def install
    system "make", "build"
    bin.install "bin/arm"
  end

  test do
    system "#{bin}/arm", "version"
  end
end
```

#### Chocolatey Package
```xml
<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd">
  <metadata>
    <id>ai-rules-manager</id>
    <version>1.0.0</version>
    <title>AI Rules Manager</title>
    <authors>Max Dunn</authors>
    <description>Package manager for AI coding assistant rulesets</description>
    <projectUrl>https://github.com/max-dunn/ai-rules-manager</projectUrl>
    <licenseUrl>https://github.com/max-dunn/ai-rules-manager/blob/main/LICENSE.txt</licenseUrl>
    <requireLicenseAcceptance>false</requireLicenseAcceptance>
    <tags>ai coding rules package-manager</tags>
  </metadata>
  <files>
    <file src="bin\arm.exe" target="tools" />
  </files>
</package>
```

### Installation Scripts
```bash
#!/bin/bash
# scripts/install.sh

set -e

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Download URL
BINARY="arm-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY="${BINARY}.exe"
fi

URL="https://github.com/max-dunn/ai-rules-manager/releases/latest/download/${BINARY}"

# Download and install
echo "Downloading ARM for ${OS}/${ARCH}..."
curl -sSL "$URL" -o arm
chmod +x arm

# Install to system
if [ -w "/usr/local/bin" ]; then
    mv arm /usr/local/bin/
    echo "ARM installed to /usr/local/bin/arm"
else
    sudo mv arm /usr/local/bin/
    echo "ARM installed to /usr/local/bin/arm (with sudo)"
fi

# Verify installation
arm version
echo "ARM installation complete!"
```

## Build Monitoring

### Build Metrics
- **Build time**: Track build duration trends
- **Binary size**: Monitor size increases
- **Test coverage**: Maintain coverage thresholds
- **Security scan**: Track vulnerability counts

### Notifications
```yaml
# Slack notification on build failure
- name: Notify Slack on failure
  if: failure()
  uses: 8398a7/action-slack@v3
  with:
    status: failure
    channel: '#builds'
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

## Troubleshooting Builds

### Common Build Issues
```bash
# Module issues
go clean -modcache
go mod download

# Build cache issues
go clean -cache
go clean -testcache

# Cross-compilation issues
go env GOOS GOARCH
CGO_ENABLED=0 go build
```

### Debug Build Problems
```bash
# Verbose build output
go build -v ./cmd/arm

# Show build commands
go build -x ./cmd/arm

# Check build constraints
go list -f '{{.GoFiles}}' ./cmd/arm
```
