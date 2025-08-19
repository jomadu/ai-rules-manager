# Usage Guide

Complete command reference and common workflows for ARM.

## Command Overview

```bash
arm <command> [options] [arguments]
```

### Global Options
- `--global` - Operate on global configuration
- `--quiet` - Suppress non-essential output
- `--verbose` - Show detailed output
- `--dry-run` - Show what would be done without executing
- `--json` - Output machine-readable JSON format
- `--no-color` - Disable colored output
- `--insecure` - Allow insecure HTTP connections

## Core Commands

### `arm install`

Install rulesets from configured registries.

#### Install from Manifest
```bash
# Install all rulesets defined in arm.json
arm install
```

#### Install Specific Ruleset
```bash
# Install latest version
arm install coding-standards

# Install specific version
arm install coding-standards@v1.2.0

# Install from specific registry
arm install myregistry/coding-standards

# Install with version constraint
arm install coding-standards@^1.0.0
```

#### Install Options
```bash
# Install with file patterns (Git registries only)
arm install coding-standards --patterns "rules/*.md,docs/*.md"

# Install excluding files (Git registries only)
arm install coding-standards --patterns "**/*.md,!**/internal/**"

# Install to specific channels
arm install coding-standards --channels cursor,q
```

#### Examples
```bash
# Install team standards to both Cursor and Amazon Q
arm install team/coding-standards@latest --patterns "standards/*.md" --channels cursor,q

# Install security rules from S3 registry
arm install s3-prod/security-rules@>=2.0.0

# Dry run to see what would be installed
arm install coding-standards --dry-run --patterns "rules/*.md"
```

### `arm update`

Update installed rulesets to newer versions.

#### Update All Rulesets
```bash
# Update all rulesets
arm update

# Dry run to see what would be updated
arm update --dry-run
```

#### Update Specific Ruleset
```bash
# Update specific ruleset
arm update coding-standards

# Update from specific registry
arm update myregistry/coding-standards
```

### `arm uninstall`

Remove installed rulesets.

```bash
# Uninstall ruleset from all channels
arm uninstall coding-standards

# Uninstall from specific channels
arm uninstall coding-standards --channels cursor

# Uninstall from specific registry
arm uninstall myregistry/coding-standards

# Dry run
arm uninstall coding-standards --dry-run
```

### `arm list`

List installed rulesets.

```bash
# List all installed rulesets
arm list

# List with JSON output
arm list --json

# List from specific channels
arm list --channels cursor

# List local installations only
arm list --local

# List global installations only
arm list --global
```

### `arm search`

Search for rulesets across registries.

```bash
# Search all registries
arm search "coding standards"

# Search specific registries
arm search "security" --registries=team,s3-prod

# Search with pattern matching
arm search "rules" --registries="team*"

# Limit results
arm search "standards" --limit=10

# JSON output
arm search "coding" --json
```

### `arm info`

Show detailed information about a ruleset.

```bash
# Show ruleset information
arm info coding-standards

# Show from specific registry
arm info team/coding-standards

# Show all available versions
arm info coding-standards --versions

# JSON output
arm info coding-standards --json
```

### `arm outdated`

Show outdated rulesets that have newer versions available.

```bash
# Check for outdated rulesets
arm outdated

# JSON output for scripting
arm outdated --json
```

### `arm outdated`

Show outdated rulesets.

```bash
# Check for outdated rulesets
arm outdated

# JSON output
arm outdated --json
```

### `arm clean`

Clean unused rulesets and cache.

```bash
# Clean all unused rulesets (default)
arm clean

# Clean specific target
arm clean unused
arm clean cache    # No-op (caching disabled)
arm clean all

# Force without confirmation
arm clean --force

# Dry run
arm clean --dry-run
```

## Configuration Commands

### `arm config`

Manage ARM configuration.

#### View Configuration
```bash
# List all configuration
arm config list

# Get specific value
arm config get registries.default
arm config get cache.maxSize

# List global configuration
arm config list --global
```

#### Set Configuration
```bash
# Set registry URL
arm config set registries.default https://github.com/new/repo

# Set cache size (in bytes)
arm config set cache.maxSize 2147483648

# Set network timeout
arm config set network.timeout 60
```

#### Add Registries
```bash
# Add Git registry
arm config add registry team https://github.com/team/rules --type=git

# Add with authentication
arm config add registry private https://github.com/org/private --type=git --authToken=$GITHUB_TOKEN

# Add S3 registry
arm config add registry s3-rules my-bucket --type=s3 --region=us-east-1

# Add local registry
arm config add registry local-dev /path/to/rules --type=local
```

#### Add Channels
```bash
# Add single directory channel
arm config add channel cursor --directories .cursor/rules

# Add multi-directory channel
arm config add channel both --directories ".cursor/rules,.amazonq/rules"
```

#### Remove Configuration
```bash
# Remove registry
arm config remove registry old-registry

# Remove channel
arm config remove channel unused-channel
```

## Common Workflows

### Initial Project Setup

```bash
# 1. Initialize ARM in project
arm install

# 2. Add team registry
arm config add registry team https://github.com/company/ai-rules --type=git --authToken=$GITHUB_TOKEN

# 3. Add channels for your AI tools
arm config add channel cursor --directories .cursor/rules
arm config add channel q --directories .amazonq/rules

# 4. Install team standards
arm install team/coding-standards@latest --patterns "standards/*.md,guidelines/*.md"

# 5. Install security rules
arm install team/security-rules@^2.0.0 --patterns "security/*.md"

# 6. Verify installation
arm list
```

### Daily Development

```bash
# Check for outdated rulesets
arm outdated

# Update all rulesets
arm update

# Install new ruleset
arm install team/new-feature-rules --patterns "features/*.md"

# Search for specific rules
arm search "testing guidelines"
```

### Team Synchronization

```bash
# Pull latest team configuration
git pull

# Install any new rulesets
arm install

# Update existing rulesets
arm update

# Clean unused rulesets
arm clean
```

### Registry Management

```bash
# Add new registry
arm config add registry staging https://github.com/team/staging-rules --type=git

# Test registry connectivity
arm info staging/test-ruleset

# Switch default registry
arm config set registries.default staging

# List all registries
arm config list | grep registries
```

## Advanced Usage

### Pattern Matching

#### Include Patterns
```bash
# Specific files
arm install rules --patterns "coding-standards.md,security-guidelines.md"

# Wildcard patterns
arm install rules --patterns "standards/*.md,guidelines/**/*.md"

# Multiple patterns
arm install rules --patterns "rules/*.md,docs/*.md,examples/*.md"
```

#### Exclude Patterns
```bash
# Exclude specific directories
arm install rules --patterns "**/*.md,!**/internal/**,!**/drafts/**"

# Exclude file types
arm install rules --patterns "**/*,!**/*.tmp,!**/*.bak"
```

### Version Constraints

```bash
# Exact version
arm install rules@1.2.3

# Latest version
arm install rules@latest

# Semver constraints
arm install rules@^1.0.0    # >=1.0.0 <2.0.0
arm install rules@~1.2.0    # >=1.2.0 <1.3.0
arm install rules@>=1.1.0   # >=1.1.0

# Range constraints
arm install rules@">=1.0.0 <2.0.0"
```

### Multi-Registry Operations

```bash
# Install from different registries
arm install team/coding-standards@latest
arm install s3-prod/security-rules@^2.0.0
arm install local-dev/experimental-rules@main

# Search across specific registries
arm search "standards" --registries=team,s3-prod

# Update from specific registry
arm update team/coding-standards
```

### Channel Management

```bash
# Install to specific channels
arm install rules --channels cursor
arm install rules --channels cursor,q

# List installations by channel
arm list --channels cursor
arm list --channels q

# Uninstall from specific channels
arm uninstall rules --channels cursor
```

### Batch Operations

```bash
# Install multiple rulesets
arm install coding-standards security-rules testing-guidelines

# Update multiple rulesets
arm update coding-standards security-rules

# Uninstall multiple rulesets
arm uninstall old-rules deprecated-standards
```

## Output Formats

### Human-Readable Output
```bash
arm list
# Output:
# Installed rulesets (scope: both):
#
# Configured rulesets:
#   team/coding-standards@^1.0.0
#     Patterns: standards/*.md, guidelines/*.md
#   team/security-rules@latest
```

### JSON Output
```bash
arm list --json
# Output:
# {
#   "scope": "both",
#   "channels": null,
#   "rulesets": [
#     {
#       "registry": "team",
#       "name": "coding-standards",
#       "version": "^1.0.0",
#       "patterns": ["standards/*.md", "guidelines/*.md"]
#     }
#   ]
# }
```

## Error Handling

### Common Errors and Solutions

#### Registry Not Found
```bash
# Error: registry 'team' not found
# Solution: Add the registry
arm config add registry team https://github.com/team/rules --type=git
```

#### Authentication Failed
```bash
# Error: authentication failed for registry 'private'
# Solution: Set authentication token
export GITHUB_TOKEN=your_token
arm config add registry private https://github.com/org/private --type=git --authToken=$GITHUB_TOKEN
```

#### Permission Denied
```bash
# Error: permission denied writing to .cursor/rules
# Solution: Create directory or fix permissions
mkdir -p .cursor/rules
chmod 755 .cursor/rules
```

#### Version Not Found
```bash
# Error: version 'v2.0.0' not found for ruleset 'coding-standards'
# Solution: Check available versions (feature not yet fully implemented)
arm info coding-standards --versions
```

### Debugging

```bash
# Enable verbose output
arm install coding-standards --verbose

# Dry run to see what would happen
arm install coding-standards --dry-run

# Check configuration
arm config list

# Test registry connectivity
arm info team/test-ruleset
```

## Performance Tips

### Caching
- ARM uses content-based caching for registry operations
- Use `arm clean cache` to clear cache (currently no-op as caching is disabled)
- Configure cache settings in `.armrc` configuration file

### Concurrency
- Configure registry-specific concurrency limits
- Git registries default to 1 concurrent operation (API limits)
- S3 registries can handle higher concurrency

### Network
- Configure timeouts for slow networks
- Set retry limits for unreliable connections
- Use local registries for development

## Best Practices

### Version Management
- Use semantic version constraints (`^1.0.0`) for flexibility
- Pin exact versions (`1.2.3`) for critical production environments
- Regularly update rulesets with `arm update`

### Registry Organization
- Use descriptive registry names
- Separate development and production registries
- Document registry purposes and access requirements

### Team Workflow
- Commit `arm.json` and `arm.lock` to version control for reproducible builds
- Don't commit `.armrc` (may contain sensitive registry configurations)
- Use global `~/.arm/.armrc` for team registry setup with authentication
- Use local configuration for project-specific overrides
