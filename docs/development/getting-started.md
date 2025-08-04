# Development Getting Started

Set up your development environment for contributing to ARM.

## Prerequisites

### Required Tools

- **Go 1.22+** - Primary development language
- **Git** - Version control
- **Make** - Build automation
- **Python 3.8+** - Pre-commit hooks

### Optional Tools

- **golangci-lint** - Code linting (installed via make)
- **goimports** - Import formatting (installed via make)
- **pre-commit** - Git hooks (installed via make)

## Environment Setup

### 1. Clone Repository

```bash
git clone https://github.com/user/arm.git
cd arm
```

### 2. Install Development Tools

```bash
# Install all development dependencies
make setup

# This installs:
# - golangci-lint (linting)
# - goimports (import formatting)
# - pre-commit hooks (git hooks)
```

### 3. Verify Setup

```bash
# Run tests
make test

# Run linter
make lint

# Build binary
make build

# Test binary
./arm version
```

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Changes

Edit code, add tests, update documentation.

### 3. Run Checks

```bash
# Format code
make fmt

# Run all checks (format, lint, test)
make check
```

### 4. Commit Changes

Pre-commit hooks will run automatically:

```bash
git add .
git commit -m "feat: add new feature"
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
# Create pull request on GitHub
```

## Project Structure

```
arm/
├── cmd/arm/              # CLI commands
│   ├── main.go          # Main entry point
│   ├── install.go       # Install command
│   ├── list.go          # List command
│   └── ...
├── internal/            # Internal packages
│   ├── cache/           # Cache management
│   ├── config/          # Configuration
│   ├── installer/       # Installation logic
│   ├── registry/        # Registry implementations
│   └── ...
├── pkg/types/           # Public types
├── test/                # Test utilities and fixtures
├── docs/                # Documentation
└── scripts/             # Build and utility scripts
```

## Testing

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/cache/

# With coverage
go test -cover ./...

# Integration tests
go test ./test/integration/
```

### Writing Tests

#### Unit Tests

```go
func TestInstaller_Install(t *testing.T) {
    // Arrange
    installer := NewInstaller()

    // Act
    err := installer.Install("test-ruleset", "1.0.0")

    // Assert
    assert.NoError(t, err)
}
```

#### Integration Tests

```go
func TestInstallWorkflow(t *testing.T) {
    // Use test fixtures and temporary directories
    tempDir := t.TempDir()
    // ... test full workflow
}
```

## Code Style

### Go Conventions

- Use `gofmt` for formatting
- Follow effective Go guidelines
- Use meaningful variable names
- Keep functions focused and small

### Naming Conventions

```go
// Types: PascalCase
type RegistryManager struct {}

// Functions: PascalCase (exported), camelCase (private)
func NewManager() *Manager {}
func (m *Manager) parseConfig() error {}

// Constants: PascalCase or UPPER_CASE
const DefaultTimeout = 30 * time.Second
const MAX_RETRIES = 3
```

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to install ruleset %s: %w", name, err)
}

// Use custom error types when appropriate
type RegistryError struct {
    Registry string
    Err      error
}
```

## Debugging

### Debug Mode

```bash
export ARM_DEBUG=1
./arm install typescript-rules
```

### Using Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug binary
dlv exec ./arm -- install typescript-rules

# Debug tests
dlv test ./internal/installer/
```

### Logging

```go
import "log"

// Debug logging (only when ARM_DEBUG=1)
if os.Getenv("ARM_DEBUG") != "" {
    log.Printf("Debug: %s", message)
}
```

## Common Tasks

### Adding a New Command

1. Create command file in `cmd/arm/`
2. Add command to `main.go`
3. Implement command logic
4. Add tests
5. Update documentation

### Adding a New Registry Type

1. Implement `Registry` interface in `internal/registry/`
2. Add configuration parsing
3. Add to registry manager
4. Add tests
5. Update documentation

### Updating Dependencies

```bash
# Update go.mod
go get -u ./...
go mod tidy

# Test everything still works
make check
```

## Troubleshooting

### Common Issues

**Tests Failing**
- Check Go version (1.22+ required)
- Run `go mod tidy`
- Clear test cache: `go clean -testcache`

**Linter Errors**
- Run `make fmt` to fix formatting
- Check golangci-lint configuration
- Update linter: `make install-tools`

**Build Failures**
- Check Go version compatibility
- Verify all dependencies available
- Clear module cache: `go clean -modcache`

### Getting Help

- Check existing issues on GitHub
- Review documentation
- Ask questions in discussions
- Contact maintainers

## Contributing Guidelines

### Before Submitting

- [ ] Tests pass locally
- [ ] Code is formatted (`make fmt`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated
- [ ] Commit messages follow convention

### Pull Request Checklist

- [ ] Clear description of changes
- [ ] Tests added for new functionality
- [ ] Documentation updated
- [ ] Breaking changes noted
- [ ] Changelog entry added
