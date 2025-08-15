# Development Setup

## Prerequisites

- **Go 1.23.2+** - [Download](https://golang.org/dl/)
- **Git** - Version control
- **Make** - Build automation

## Quick Start

```bash
# Clone and setup
git clone https://github.com/max-dunn/ai-rules-manager.git
cd ai-rules-manager
go mod download
make install-tools

# Verify
make test
make build
./bin/arm --help
```

## Development Commands

```bash
# Build and test
make build          # Build binary
make test           # Run all tests
make lint           # Run linter
make fmt            # Format code

# Testing
go test ./internal/config/              # Specific package
go test -v ./internal/registry/         # Verbose output
go test -race ./...                     # Race detection

# Integration tests
./tests/integration/git/setup-test-repos.sh
make test-integration
```

## Editor Setup

### VS Code
Install Go extension and add to `.vscode/settings.json`:
```json
{
  "go.lintTool": "golangci-lint",
  "go.formatTool": "gofumpt",
  "go.testFlags": ["-v", "-race"]
}
```

## Debugging

```bash
# Debug tests
dlv test ./internal/config/ -- -test.run TestLoad

# Debug application
dlv exec ./bin/arm -- install test-ruleset
```

## Troubleshooting

```bash
# Module issues
go clean -modcache
go mod download

# Build issues
go version
make clean && make build

# Test failures
go test -v ./...
```
