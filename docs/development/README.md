# Development Documentation

Technical documentation for ARM developers and contributors.

## Quick Start

```bash
# Clone and setup
git clone https://github.com/jomadu/ai-rules-manager.git
cd ai-rules-manager && make setup

# Run tests and build
make test && make build
```

## Documentation Structure

### 🏗️ [Architecture](architecture/)
- **[Overview](architecture/overview.md)** - High-level system design
- **[Components](architecture/components.md)** - Core components breakdown
- **[Registries](architecture/registries.md)** - Registry system design
- **[Caching](architecture/caching.md)** - Cache architecture
- **[ADRs](architecture/adr/)** - Architecture Decision Records

### 🧪 Testing
- **Testing philosophy and coverage requirements**

## Quick Start

### Prerequisites

- Go 1.22+
- Git
- Make

### Setup

```bash
# Clone repository
git clone https://github.com/jomadu/ai-rules-manager.git
cd ai-rules-manager

# Install development tools
make setup

# Run tests
make test

# Build binary
make build
```

## Architecture Overview

### Core Components

```
cmd/arm/           # CLI commands and main entry point
internal/          # Internal packages
├── cache/         # Global cache management
├── config/        # Configuration parsing
├── installer/     # Package installation logic
├── registry/      # Registry implementations
└── updater/       # Update and version checking
pkg/types/         # Public types and interfaces
```

### Registry System

ARM supports multiple registry types through a common interface:

- **GitLab**: Package registries with API access
- **S3**: AWS S3 bucket-based storage
- **HTTP**: Generic HTTP file servers
- **Filesystem**: Local directory registries

### Caching Strategy

- **Global Cache**: `~/.arm/cache/` for downloaded packages
- **Metadata Cache**: Registry information and version lists
- **Package Cache**: Extracted ruleset files

## Development Workflow

### Feature Development

1. Create feature branch from `main`
2. Implement changes with tests
3. Update documentation
4. Submit pull request
5. Code review and merge

### Testing Philosophy

- **Test-Driven Development** - Write tests before implementation
- **Comprehensive Coverage** - 85%+ test coverage target
- **Fast Feedback** - Quick test execution for development workflow
- **Reliable Tests** - Consistent results across environments

### Code Quality

- **Linting**: golangci-lint with strict rules
- **Formatting**: gofmt and goimports
- **Security**: CodeQL scanning
- **Coverage**: 85%+ test coverage target

## Build and Release

### Local Development

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run all checks
make check
```

### Release Process

1. Update version in code
2. Create release tag
3. Automated CI builds binaries
4. GitHub release created automatically
5. Installation script updated

## Contributing

### Code Style

- Follow Go conventions
- Use meaningful variable names
- Add comments for complex logic
- Keep functions focused and small

### Commit Messages

Use conventional commits:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `refactor:` - Code refactoring

### Pull Request Process

1. Ensure all tests pass
2. Update documentation
3. Add changelog entry
4. Request review from maintainers

## Debugging

### Debug Mode

```bash
export ARM_DEBUG=1
arm install typescript-rules
```

### Common Issues

- **Registry connectivity**: Check network and authentication
- **Cache corruption**: Clear with `arm clean --cache`
- **Version conflicts**: Check semantic versioning constraints

### Profiling

```bash
# CPU profiling
go build -o arm-debug ./cmd/arm
./arm-debug -cpuprofile=cpu.prof install typescript-rules
go tool pprof cpu.prof

# Memory profiling
./arm-debug -memprofile=mem.prof install typescript-rules
go tool pprof mem.prof
```

## External Dependencies

### Core Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/stretchr/testify` - Testing utilities
- `github.com/fatih/color` - Terminal colors

### Registry Dependencies

- AWS SDK for S3 support
- HTTP client for GitLab/HTTP registries
- Archive utilities for package handling

## Security Considerations

### Authentication

- Environment variable storage for tokens
- No credentials in configuration files
- Secure token transmission

### Package Validation

- Checksum verification
- Archive extraction safety
- Path traversal prevention

### Network Security

- HTTPS enforcement
- Certificate validation
- Timeout and retry limits
