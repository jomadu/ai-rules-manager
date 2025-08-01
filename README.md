# Model Rules Manager (MRM)

A package manager for AI coding assistant rulesets. Install, update, and manage coding rules across different AI tools like Cursor and Amazon Q Developer.

## What is MRM?

MRM solves the problem of managing and sharing AI coding rulesets across teams and projects. Instead of manually copying `.cursorrules` files or `.amazonq/rules` directories, MRM provides a centralized way to distribute, version, and update coding rules from multiple registries.

## Quick Start

### Installation

```bash
# Install via curl (coming soon)
curl -sSL https://install.mrm.dev | sh

# Or download binary from releases
wget https://github.com/user/mrm/releases/latest/download/mrm-linux-amd64
chmod +x mrm-linux-amd64
sudo mv mrm-linux-amd64 /usr/local/bin/mrm
```

### Basic Usage

```bash
# Install a ruleset
mrm install company@typescript-rules

# Install from manifest
mrm install

# List installed rulesets
mrm list

# Update all rulesets
mrm update

# Check for outdated rulesets
mrm outdated
```

## Configuration

### Project Configuration (rules.json)

```json
{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "typescript-rules": "^1.0.0",
    "company@security-rules": "^2.1.0"
  }
}
```

### Registry Configuration (.mpmrc)

```ini
[sources]
default = https://registry.mpmjs.org/
company = https://internal.company-registry.local/

[sources.company]
authToken = $COMPANY_REGISTRY_TOKEN
```

## Commands

| Command | Description |
|---------|-------------|
| `mrm install [ruleset]` | Install rulesets |
| `mrm uninstall <ruleset>` | Remove a ruleset |
| `mrm update [ruleset]` | Update rulesets |
| `mrm list` | List installed rulesets |
| `mrm outdated` | Show outdated rulesets |
| `mrm config <action>` | Manage configuration |
| `mrm clean` | Clean cache and unused files |
| `mrm help` | Show help |
| `mrm version` | Show version |

## Supported Targets

- **Cursor IDE**: `.cursorrules`
- **Amazon Q Developer**: `.amazonq/rules/`
- Extensible for future AI coding tools

## Supported Registries

- GitLab package registries
- GitHub package registries
- AWS S3 buckets
- Generic HTTP endpoints
- Local file system

## File Structure

After installation, your project will look like:

```
.mpm/
  cache/
    company/
      typescript-rules/
        1.0.1/
          rule-1.md
          rule-2.md
.cursorrules/
  lrm/
    company/
      typescript-rules/
        1.0.1/
          rule-1.md
          rule-2.md
.amazonq/
  rules/
    lrm/
      company/
        typescript-rules/
          1.0.1/
            rule-1.md
            rule-2.md
rules.json
rules.lock
.mpmrc
```

## Development Status

🚧 **In Development** - MRM is currently being built. See our [development phases](prd.md#timeline):

- **Phase 1**: Core commands (install, uninstall, list)
- **Phase 2**: Configuration and registry support
- **Phase 3**: Update/outdated functionality
- **Phase 4**: Cache management and cleanup
- **Phase 5**: Testing and documentation

## Contributing

This project is implemented in Go for fast, dependency-free distribution. See [prd.md](prd.md) for detailed requirements and architecture decisions.

## License

MIT License - see [LICENSE](LICENSE) for details.