# GitHub Copilot Instructions

You are an AI coding assistant integrated with GitHub Copilot. Follow these guidelines when generating code and providing assistance:

## Code Style and Standards

- Follow the project's existing code style and conventions
- Use meaningful variable and function names
- Include appropriate comments for complex logic
- Ensure code is readable and maintainable

## Security Considerations

- Never include hardcoded secrets, API keys, or passwords
- Use environment variables for sensitive configuration
- Follow security best practices for the language/framework
- Validate user inputs appropriately

## Testing

- Suggest unit tests for new functions when appropriate
- Follow the project's existing testing patterns
- Consider edge cases and error conditions

## Documentation

- Generate clear, concise docstrings/comments
- Update README files when adding new features
- Include usage examples for public APIs

## Framework-Specific Guidelines

- For Go projects: Follow Go idioms and error handling patterns
- For JavaScript/TypeScript: Use modern ES features appropriately
- For Python: Follow PEP 8 and use type hints when beneficial

## Project Context

This is the AI Rules Manager (ARM) project - a package manager for AI coding assistant rulesets. When working on this codebase:

- Maintain compatibility with existing registry types (Git, S3, GitLab, HTTPS, Local)
- Ensure new features work with the existing channel system
- Follow the established patterns for CLI commands and configuration
- Consider cross-platform compatibility (Linux, macOS, Windows)