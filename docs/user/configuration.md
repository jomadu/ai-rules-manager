# Configuration

ARM uses two configuration files to manage behavior and registries.

## Project Configuration (rules.json)

Project-specific ruleset dependencies and targets.

```json
{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "typescript-rules": "^1.0.0",
    "company@security-rules": "^2.1.0",
    "react-rules": "~1.5.0"
  }
}
```

### Fields

- **targets** - Array of directories where rules are installed
- **dependencies** - Object mapping ruleset names to version ranges

### Version Ranges

- `^1.0.0` - Compatible with 1.x.x (>=1.0.0 <2.0.0)
- `~1.5.0` - Patch-level changes (>=1.5.0 <1.6.0)
- `1.2.3` - Exact version
- `latest` - Always use latest version

## Global Configuration (.armrc)

User-wide settings stored in `~/.armrc`.

```ini
[sources]
default = https://registry.armjs.org/
company = https://gitlab.company.com

[sources.company]
type = gitlab
projectID = 12345
authToken = $COMPANY_REGISTRY_TOKEN
concurrency = 2

[performance]
defaultConcurrency = 3

[performance.gitlab]
concurrency = 3

[performance.s3]
concurrency = 8
```

### Registry Sources

Configure multiple registries:

```bash
# Add registry
arm config set sources.company https://gitlab.company.com

# Set registry type and auth
arm config set sources.company.type gitlab
arm config set sources.company.authToken $TOKEN
```

### Performance Settings

Control download concurrency:

```bash
# Global default
arm config set performance.defaultConcurrency 5

# Registry-specific
arm config set performance.gitlab.concurrency 3
arm config set performance.s3.concurrency 10
```

## Environment Variables

Override configuration with environment variables:

- `ARM_CONFIG_PATH` - Custom config file location
- `ARM_CACHE_DIR` - Custom cache directory (default: `~/.arm/cache`)
- `ARM_REGISTRY_TOKEN` - Default registry authentication token
- `ARM_CONCURRENCY` - Override default concurrency

```bash
export ARM_REGISTRY_TOKEN=your-token
export ARM_CONCURRENCY=8
arm install typescript-rules
```

## Configuration Commands

```bash
# View all settings
arm config list

# Get specific value
arm config get sources.default

# Set value
arm config set sources.company https://internal.company.com

# Remove value
arm config unset sources.company.authToken
```

## File Locations

- **Project config**: `./rules.json`
- **Global config**: `~/.armrc`
- **Lock file**: `./rules.lock`
- **Cache**: `~/.arm/cache/`
