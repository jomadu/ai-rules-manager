# User Documentation

Guides for ARM end users.

## Getting Started

- **[Getting Started](getting-started.md)** - Installation and basic usage
- **[Configuration](configuration.md)** - Registry setup
- **[Troubleshooting](troubleshooting.md)** - Common issues

## Registry Types

- **[GitLab](registries/gitlab.md)** - GitLab package registries
- **[S3](registries/s3.md)** - AWS S3 registries
- **[HTTP](registries/http.md)** - Generic HTTP registries
- **[Filesystem](registries/filesystem.md)** - Local registries

## Commands

| Command | Description |
|---------|-------------|
| `arm install [ruleset]` | Install rulesets |
| `arm list` | Show installed |
| `arm update` | Update rulesets |
| `arm config list` | Show configuration |
| `arm clean` | Remove unused files |

## Files

- `rules.json` - Project dependencies
- `rules.lock` - Exact versions (auto-generated)
- `.armrc` - Configuration (optional)
- `~/.arm/cache/` - Global cache
