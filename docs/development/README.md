# Development Documentation

Documentation for contributors and maintainers of AI Rules Manager.

## Getting Started
1. **[Development Setup](setup.md)** - Set up local development environment
2. **[Contributing Guide](contributing.md)** - How to contribute to ARM
3. **[Build System](build.md)** - Build process and CI/CD pipeline

## Maintenance
4. **[Release Process](release.md)** - How to create and publish releases
5. **[Architecture Decisions](adr/)** - Record of architectural decisions
   - [ADR-003: Patterns in Lock File](adr/003-patterns-in-lock-file.md) - Git registry pattern storage

## Development Workflow

### Quick Start
```bash
# Clone repository
git clone https://github.com/jomadu/ai-rules-manager.git
cd ai-rules-manager

# Install dependencies
go mod download

# Run tests
make test

# Build
make build

# Run locally
./bin/arm --help
```

### Project Structure
```
ai-rules-manager/
├── cmd/arm/              # Main application entry point
├── internal/             # Internal packages
│   ├── cache/           # Caching system
│   ├── cli/             # Command-line interface
│   ├── config/          # Configuration management
│   ├── install/         # Installation orchestration
│   ├── registry/        # Registry implementations
│   ├── update/          # Update service
│   └── version/         # Version resolution
├── scripts/             # Build and utility scripts
├── tests/               # Integration tests
├── docs/                # Current documentation
├── docs-new/            # New documentation structure
└── Makefile             # Build targets
```

## Development Standards

### Code Quality
- **Go version**: 1.23.2+
- **Linting**: golangci-lint with strict configuration
- **Testing**: Minimum 80% coverage
- **Documentation**: All public APIs documented

### Git Workflow
- **Branching**: Feature branches from `main`
- **Commits**: Conventional commit format
- **PRs**: Required for all changes
- **Reviews**: At least one approval required

### Testing Strategy
- **Unit tests**: All packages have comprehensive unit tests
- **Integration tests**: End-to-end workflow testing
- **Performance tests**: Benchmarks for critical paths
- **Security tests**: Static analysis and dependency scanning

## Architecture Principles

### Design Goals
- **Modularity**: Clear separation of concerns
- **Extensibility**: Easy to add new registry types
- **Performance**: Efficient caching and concurrent operations
- **Reliability**: Robust error handling and recovery
- **Security**: Secure by default with proper authentication

### Key Patterns
- **Interface-based design**: Registry abstraction
- **Dependency injection**: Testable components
- **Error wrapping**: Contextual error information
- **Configuration hierarchy**: Flexible configuration merging
- **Content-based caching**: Efficient storage and retrieval

## Contributing Areas

### High Priority
- New registry type implementations
- Performance optimizations
- Documentation improvements
- Test coverage expansion
- Security enhancements

### Medium Priority
- CLI usability improvements
- Additional search capabilities
- Monitoring and observability
- Configuration validation
- Error message improvements

### Future Considerations
- Web UI for registry management
- Plugin system for custom registry types
- Distributed caching
- Registry mirroring
- Advanced pattern matching
