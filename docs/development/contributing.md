# Contributing Guide

How to contribute to AI Rules Manager development.

## Getting Started

### Prerequisites
- Go 1.23.2+
- Git
- Make

### Setup
```bash
# Fork and clone
git clone https://github.com/your-username/ai-rules-manager.git
cd ai-rules-manager

# Install dependencies
go mod download

# Run tests
make test

# Build
make build
```

## Development Workflow

### Branch Strategy
- `main` - stable release branch
- `feature/name` - feature development
- `fix/name` - bug fixes
- `docs/name` - documentation updates

### Commit Format
Use conventional commits:
```
feat: add S3 registry support
fix: resolve cache corruption issue
docs: update configuration guide
test: add integration tests for Git registry
```

### Pull Request Process
1. Create feature branch from `main`
2. Make changes with tests
3. Update documentation
4. Submit PR with description
5. Address review feedback
6. Merge after approval

## Code Standards

### Go Style
- Follow `gofmt` formatting
- Use `golangci-lint` for linting
- Write comprehensive tests
- Document public APIs

### Testing Requirements
- Unit tests for all packages
- Integration tests for workflows
- Minimum 80% coverage
- Benchmarks for performance-critical code

### Documentation
- Update relevant docs for changes
- Include code examples
- Keep README current
- Add ADRs for architectural decisions

## Areas for Contribution

### High Priority
- New registry types (Azure DevOps, Bitbucket)
- Performance optimizations
- Security enhancements
- Error handling improvements

### Medium Priority
- CLI usability features
- Additional search capabilities
- Monitoring/observability
- Configuration validation

### Documentation
- User guides and tutorials
- API documentation
- Architecture documentation
- Troubleshooting guides

## Testing

### Running Tests
```bash
# Unit tests
make test

# Integration tests
make test-integration

# Specific package
go test ./internal/registry/

# With coverage
make test-coverage
```

### Test Structure
```
tests/
├── unit/           # Unit tests alongside code
├── integration/    # End-to-end tests
└── fixtures/       # Test data
```

## Release Process

### Version Numbering
- Semantic versioning (MAJOR.MINOR.PATCH)
- Pre-release: v1.2.3-alpha.1
- Release candidates: v1.2.3-rc.1

### Release Steps
1. Update version in code
2. Update CHANGELOG.md
3. Create release PR
4. Tag release after merge
5. GitHub Actions handles build/publish

## Architecture Guidelines

### Design Principles
- Interface-based design
- Dependency injection
- Error wrapping with context
- Configuration hierarchy
- Content-based caching

### Package Structure
```
internal/
├── cache/      # Caching system
├── cli/        # Command-line interface
├── config/     # Configuration management
├── install/    # Installation orchestration
├── registry/   # Registry implementations
├── update/     # Update service
└── version/    # Version resolution
```

### Adding New Registry Types
1. Implement `Registry` interface
2. Add factory method
3. Add configuration validation
4. Write comprehensive tests
5. Update documentation

## Code Review Guidelines

### For Authors
- Keep PRs focused and small
- Write clear commit messages
- Include tests and documentation
- Respond promptly to feedback

### For Reviewers
- Focus on correctness and maintainability
- Check test coverage
- Verify documentation updates
- Consider security implications

## Community

### Communication
- GitHub Issues for bugs/features
- GitHub Discussions for questions
- Pull Requests for code changes

### Code of Conduct
- Be respectful and inclusive
- Focus on constructive feedback
- Help newcomers get started
- Follow project guidelines
