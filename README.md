# AI Rules Manager (ARM)

A package manager for AI coding assistant rulesets. Install, update, and manage coding rules across different AI tools like Cursor and Amazon Q Developer.

## What is ARM?

ARM solves the problem of managing and sharing AI coding rulesets across teams and projects. Instead of manually copying `.cursorrules` files or `.amazonq/rules` directories, ARM provides a centralized way to distribute, version, and update coding rules from multiple registries.

## Quick Start

### Installation

```bash
# Install via curl (coming soon)
curl -sSL https://install.arm.dev | sh

# Or download binary from releases
wget https://github.com/user/arm/releases/latest/download/arm-linux-amd64
chmod +x arm-linux-amd64
sudo mv arm-linux-amd64 /usr/local/bin/arm
```

### Basic Usage

```bash
# Install a ruleset
arm install company@typescript-rules

# Install from manifest
arm install

# List installed rulesets
arm list
arm list --format=json  # JSON output

# Update all rulesets
arm update

# Check for outdated rulesets
arm outdated
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

### Registry Configuration (.armrc)

```ini
[sources]
default = https://registry.armjs.org/
company = https://internal.company-registry.local/

[sources.company]
authToken = $COMPANY_REGISTRY_TOKEN
```

## Commands

| Command | Description |
|---------|-------------|
| `arm install [ruleset]` | Install rulesets |
| `arm uninstall <ruleset>` | Remove a ruleset |
| `arm update [ruleset]` | Update rulesets |
| `arm list [--format=table|json]` | List installed rulesets |
| `arm outdated` | Show outdated rulesets |
| `arm config <action>` | Manage configuration |
| `arm clean` | Clean cache and unused files |
| `arm help` | Show help |
| `arm version` | Show version |

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
.arm/
  cache/
    company/
      typescript-rules/
        1.0.1/
          rule-1.md
          rule-2.md
.cursorrules/
  arm/
    company/
      typescript-rules/
        1.0.1/
          rule-1.md
          rule-2.md
.amazonq/
  rules/
    arm/
      company/
        typescript-rules/
          1.0.1/
            rule-1.md
            rule-2.md
rules.json
rules.lock
.armrc
```

## Development Status

✅ **Phase 1 Complete** - Core functionality implemented and tested. See our [development phases](docs/prd.md#timeline):

- **Phase 1**: Core commands (install, uninstall, list) - ✅ **COMPLETED**
- **Phase 2**: Configuration and registry support - 🚧 **IN PROGRESS**
- **Phase 3**: Update/outdated functionality - 📋 **PLANNED**
- **Phase 4**: Cache management and cleanup - 📋 **PLANNED**
- **Phase 5**: Testing and documentation - 📋 **PLANNED**

📋 **Current Focus**: Multi-registry configuration support. See [docs/next-steps.md](docs/next-steps.md) for development priorities.

📈 **Status Report**: See [docs/project-status.md](docs/project-status.md) for detailed project metrics and roadmap.

## Contributing

This project is implemented in Go for fast, dependency-free distribution. See [prd.md](prd.md) for detailed requirements and architecture decisions.

## License

MIT License - see [LICENSE](LICENSE) for details.
