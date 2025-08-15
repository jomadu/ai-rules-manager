# Development Setup

Set up your local development environment for ARM.

## Prerequisites

### Required Tools
- **Go 1.23.2+** - [Download](https://golang.org/dl/)
- **Git** - Version control
- **Make** - Build automation
- **Docker** (optional) - For integration testing

### Recommended Tools
- **golangci-lint** - Code linting
- **gofumpt** - Code formatting
- **gotestsum** - Enhanced test output
- **delve** - Go debugger

## Initial Setup

### 1. Clone Repository
```bash
git clone https://github.com/max-dunn/ai-rules-manager.git
cd ai-rules-manager
```

### 2. Install Dependencies
```bash
# Download Go modules
go mod download

# Install development tools
make install-tools
```

### 3. Verify Setup
```bash
# Run tests
make test

# Build binary
make build

# Check binary works
./bin/arm --help
```

## Development Tools

### Install Development Tools
```bash
# Install all tools at once
make install-tools

# Or install individually
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest
go install gotest.tools/gotestsum@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

### Editor Setup

#### VS Code
Install recommended extensions:
- Go (official Go extension)
- golangci-lint
- Test Explorer UI

Settings (`.vscode/settings.json`):
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "gofumpt",
  "go.useLanguageServer": true,
  "go.testFlags": ["-v", "-race"],
  "go.buildFlags": ["-race"]
}
```

#### GoLand/IntelliJ
- Enable Go modules support
- Configure golangci-lint as external tool
- Set gofumpt as code formatter

## Build System

### Make Targets
```bash
# Development
make build          # Build binary
make test           # Run tests
make test-coverage  # Run tests with coverage
make lint           # Run linter
make fmt            # Format code

# Release
make build-all      # Build for all platforms
make release        # Create release artifacts

# Maintenance
make clean          # Clean build artifacts
make deps           # Update dependencies
make install-tools  # Install development tools
```

### Build Configuration
The build system uses:
- **Go modules** for dependency management
- **Make** for build automation
- **golangci-lint** for code quality
- **GitHub Actions** for CI/CD

## Testing Setup

### Test Categories
```bash
# Unit tests (fast)
make test-unit

# Integration tests (requires network)
make test-integration

# End-to-end tests (slow)
make test-e2e

# All tests
make test
```

### Test Environment
```bash
# Set up test registries
./tests/integration/git/setup-test-repos.sh

# Run specific test package
go test ./internal/config/

# Run with verbose output
go test -v ./internal/registry/

# Run with race detection
go test -race ./...
```

## Debugging

### Using Delve
```bash
# Debug specific test
dlv test ./internal/config/ -- -test.run TestLoad

# Debug main application
dlv exec ./bin/arm -- install test-ruleset

# Attach to running process
dlv attach <pid>
```

### Debug Configuration
Create `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug ARM",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/arm",
      "args": ["install", "test-ruleset", "--dry-run", "--verbose"]
    },
    {
      "name": "Debug Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/config",
      "args": ["-test.run", "TestLoad"]
    }
  ]
}
```

## Code Quality

### Linting Configuration
`.golangci.yml`:
```yaml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck

linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/max-dunn/ai-rules-manager
```

### Pre-commit Hooks
Install pre-commit hooks:
```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

`.pre-commit-config.yaml`:
```yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: gofumpt -w
        language: system
        files: \.go$

      - id: go-lint
        name: go lint
        entry: golangci-lint run
        language: system
        files: \.go$
        pass_filenames: false

      - id: go-test
        name: go test
        entry: go test -short
        language: system
        files: \.go$
        pass_filenames: false
```

## Environment Configuration

### Development Environment Variables
```bash
# Create .env file for development
cat > .env << EOF
# Test registry tokens (optional)
GITHUB_TOKEN=your_github_token
GITLAB_TOKEN=your_gitlab_token

# AWS credentials for S3 testing (optional)
AWS_PROFILE=development
AWS_REGION=us-east-1
EOF

# Load environment
source .env
```

### Shell Configuration
Add to your shell profile (`.bashrc`, `.zshrc`):
```bash
# Go development
export GOPATH=$HOME/go
export GO111MODULE=on
export PATH=$PATH:$HOME/go/bin

# Aliases
alias arm-dev='go run ./cmd/arm'
alias arm-test='make test'
alias arm-lint='make lint'
```

## Integration Testing

### Test Registry Setup
```bash
# Create test Git repository
./tests/integration/git/setup-test-repos.sh

# This creates:
# - Local Git repository with test rulesets
# - Multiple versions and tags
# - Test patterns and file structures
```

### Mock Services
For testing without external dependencies:
```bash
# Start mock Git server
docker run -d -p 3000:3000 gitea/gitea:latest

# Start mock S3 server
docker run -d -p 9000:9000 minio/minio server /data
```

## Performance Profiling

### CPU Profiling
```bash
# Build with profiling
go build -o arm-prof ./cmd/arm

# Run with CPU profiling
./arm-prof -cpuprofile=cpu.prof install test-ruleset

# Analyze profile
go tool pprof cpu.prof
```

### Memory Profiling
```bash
# Run with memory profiling
./arm-prof -memprofile=mem.prof install test-ruleset

# Analyze profile
go tool pprof mem.prof
```

### Benchmarking
```bash
# Run benchmarks
go test -bench=. -benchmem ./internal/cache/

# Compare benchmarks
go test -bench=. -count=5 ./internal/registry/ | tee bench.txt
benchstat bench.txt
```

## Troubleshooting

### Common Issues

#### Module Download Failures
```bash
# Clear module cache
go clean -modcache

# Re-download modules
go mod download
```

#### Build Failures
```bash
# Check Go version
go version

# Verify GOPATH and GOROOT
go env

# Clean and rebuild
make clean
make build
```

#### Test Failures
```bash
# Run tests with verbose output
go test -v ./...

# Run specific failing test
go test -run TestSpecificFunction ./internal/package/

# Check for race conditions
go test -race ./...
```

### Getting Help
- Check existing [GitHub Issues](https://github.com/max-dunn/ai-rules-manager/issues)
- Review [Contributing Guide](contributing.md)
- Ask questions in [GitHub Discussions](https://github.com/max-dunn/ai-rules-manager/discussions)
