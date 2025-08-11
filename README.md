![AI Rules Manager](assets/header.png)

# AI Rules Manager (ARM)

A package manager for AI coding assistant rulesets that enables developers and teams to install, update, and manage coding rules across different AI tools like Cursor, Amazon Q Developer, and GitHub Copilot.

## Why ARM?

Stop manually copying `.cursorrules`, `.amazonq/rules`, and `.github/copilot-*` files between projects. ARM provides an npm-like experience for AI coding rules with version control, team synchronization, and multi-registry support.

## Quick Start

```bash
# Install ARM
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash

# Initialize configuration
arm install

# Add a registry
arm config add registry default https://github.com/mushroom-kingdom/cursor-rules.example --type=git

# Add channels for your AI tools
arm config add channel cursor --directories ~/.cursor/rules
arm config add channel q --directories ~/.aws/amazonq/rules
arm config add channel copilot --directories .github

# Install your first ruleset
arm install power-up-rules --patterns "rules/*.md"

# Verify installation
arm list
```

## Features

- **Multi-Registry Support** - Git, S3, GitLab, HTTPS, and Local registries
- **Semantic Versioning** - Version constraints with `^`, `~`, `>=` operators
- **Team Synchronization** - Share standardized rules across development teams
- **Cross-Platform** - Fast, reliable Go implementation
- **Channel Management** - Support multiple AI tools simultaneously
- **Caching & Offline** - Local caching with graceful offline fallback

## Registry Types

| Type | Example | Use Case |
|------|---------|----------|
| **Git** | `github.com/user/repo` | Public/private repositories |
| **S3** | `my-bucket` | AWS-hosted rulesets |
| **GitLab** | `gitlab.com/projects/123` | GitLab Package Registry |
| **HTTPS** | `registry.example.com` | Custom HTTP servers |
| **Local** | `/path/to/rules` | Development and testing |

## Documentation

- [User Guide](docs/user/README.md) - Complete usage documentation
- [Quick Start](docs/user/quick-start.md) - Get started in 5 minutes
- [Team Setup](docs/user/team-setup.md) - Deploy ARM across your team
- [Configuration](docs/user/configuration.md) - Advanced configuration options

## Contributing

ARM is designed for community adoption. Contributions welcome for new registry types, AI tool integrations, and feature improvements.

## License

GPL-3.0 License - see [LICENSE.txt](LICENSE.txt) for details.
