# P5.3: Documentation

## Overview
Create comprehensive documentation including CLI help text, usage examples, registry setup guides, troubleshooting, and API documentation.

## Requirements
- Complete CLI help text
- Usage examples and tutorials
- Registry setup guides
- Troubleshooting documentation
- API documentation for registry implementers

## Tasks
- [ ] **CLI help system**:
  - Command-specific help text
  - Usage examples for each command
  - Flag descriptions and defaults
  - Interactive help navigation
- [ ] **User documentation**:
  - Getting started guide
  - Configuration examples
  - Common workflows
  - Best practices guide
- [ ] **Registry setup guides**:
  - GitLab package registry setup
  - GitHub packages configuration
  - S3 bucket configuration
  - Custom HTTP registry setup
- [ ] **Troubleshooting guide**:
  - Common error scenarios
  - Debug mode usage
  - Network connectivity issues
  - Permission problems
- [ ] **API documentation**:
  - Registry interface specification
  - Custom registry implementation guide
  - Plugin development (future)
  - Configuration schema reference

## Acceptance Criteria
- [ ] All commands have comprehensive help text
- [ ] Documentation covers all major use cases
- [ ] Registry setup guides are complete and tested
- [ ] Troubleshooting guide addresses common issues
- [ ] API documentation enables third-party integrations
- [ ] Documentation is kept up-to-date with code changes

## Files to Create
- `docs/getting-started.md`
- `docs/configuration.md`
- `docs/registries/gitlab.md`
- `docs/registries/github.md`
- `docs/registries/s3.md`
- `docs/troubleshooting.md`
- `docs/api/registry-interface.md`

## CLI Help Examples
```bash
$ arm install --help
Install rulesets from registries

Usage:
  arm install [ruleset[@version]] [flags]

Examples:
  arm install                    # Install from rules.json
  arm install typescript-rules   # Install latest version
  arm install company@rules@1.0  # Install specific version

Flags:
  --dry-run    Show what would be installed
  --force      Overwrite existing rulesets
```

## Documentation Structure
```
docs/
  getting-started.md
  configuration.md
  commands/
    install.md
    update.md
    list.md
  registries/
    gitlab.md
    github.md
    s3.md
  troubleshooting.md
  api/
    registry-interface.md
```

## Notes
- Consider interactive documentation (man pages)
- Plan for documentation versioning
- Implement documentation testing/validation