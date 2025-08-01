# Contributing to ARM

## Development Setup

1. Install Go 1.22 or later
2. Install Python 3.8 or later (for pre-commit hooks)
3. Ensure `$GOPATH/bin` is in your PATH (usually `~/go/bin`)
4. Clone the repository
5. Run setup: `make setup`

## Code Quality

We use several tools to maintain code quality:

- **gofmt** and **goimports** for code formatting
- **golangci-lint** for comprehensive linting
- **pre-commit** hooks for automated checks

> **Note**: We previously used gosec for security scanning but removed it to simplify the development setup. We may reintroduce it in a future task when needed.

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

Examples:
- `feat: add install command`
- `fix: handle missing config file`
- `docs: update README with examples`

## Development Workflow

1. Create a feature branch: `git checkout -b feat/your-feature`
2. Make changes and commit with conventional messages
3. Run checks: `make check`
4. Push and create a pull request

## Testing

- Write tests for new functionality
- Ensure all tests pass: `make test`
- Maintain test coverage

## Pull Requests

- Keep PRs focused and small
- Include tests for new features
- Update documentation as needed
- Ensure CI passes
