# Commands Reference

## Installation Commands

### `arm install [ruleset]`

Install rulesets from registries or manifest.

```bash
# Install from manifest (rules.json)
arm install

# Install specific ruleset (latest version)
arm install typescript-rules

# Install from specific registry
arm install company@security-rules

# Install specific version
arm install typescript-rules@1.2.0

# Install from git repository
arm install awesome-rules@main:rules/*.md,docs/*.txt
```

**Options:**
- `--dry-run` - Show what would be installed without installing

### `arm uninstall <ruleset>`

Remove installed rulesets.

```bash
# Uninstall specific ruleset
arm uninstall typescript-rules

# Uninstall from specific registry
arm uninstall company@security-rules
```

## Update Commands

### `arm update [ruleset]`

Update rulesets to latest compatible versions.

```bash
# Update all rulesets
arm update

# Update specific ruleset
arm update typescript-rules

# Update from specific registry
arm update company@security-rules
```

### `arm outdated`

Check for available updates.

```bash
# Show outdated rulesets
arm outdated

# JSON output
arm outdated --format=json
```

## Information Commands

### `arm list`

List installed rulesets.

```bash
# Table format (default)
arm list

# JSON format
arm list --format=json
```

**Output includes:**
- Ruleset name and registry
- Installed version
- Target directories
- Installation date

### `arm version`

Show ARM version information.

```bash
arm version
```

### `arm help`

Show help information.

```bash
# General help
arm help

# Command-specific help
arm help install
```

## Configuration Commands

### `arm config list`

Show all configuration values.

```bash
arm config list
```

### `arm config get <key>`

Get specific configuration value.

```bash
arm config get sources.default
arm config get performance.defaultConcurrency
```

### `arm config set <key> <value>`

Set configuration value.

```bash
# Set registry URL
arm config set sources.company https://gitlab.company.com

# Set performance options
arm config set performance.defaultConcurrency 5
```

## Maintenance Commands

### `arm clean`

Clean cache and unused files.

```bash
# Clean package cache
arm clean

# Clean everything including registry cache
arm clean --all
```

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Configuration error
- `3` - Network error
- `4` - File system error
- `5` - Version conflict error

## Global Options

Available for all commands:

- `--verbose` - Enable verbose output
- `--quiet` - Suppress non-error output
- `--config <path>` - Use custom config file location
