# ARM - AI Rules Manager

ARM manages AI coding assistant rulesets across teams and projects. Install, update, and share coding rules for Cursor IDE and Amazon Q Developer from centralized registries.

## Quick Start

```bash
# Install ARM
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash

# Configure a registry first
arm config set sources.company https://gitlab.company.com
arm config set sources.company.type gitlab

# Install your first ruleset
arm install company@typescript-rules

# Verify installation
arm list
```

Your AI coding rules are now installed and ready to use in your IDE.

## What's Next

- **New to ARM?** → [Getting Started Guide](getting-started.md)
- **Need to install?** → [Installation Guide](installation.md)
- **Command reference** → [Commands](commands.md)
- **Configuration** → [Configuration Guide](configuration.md)
- **Registry setup** → [Registry Guide](registries.md)
- **Having issues?** → [Troubleshooting](troubleshooting.md)

## Key Features

- **Multi-IDE Support**: Works with Cursor (`.cursorrules`) and Amazon Q Developer (`.amazonq/rules`)
- **Multiple Registries**: GitLab, S3, Git repos, HTTP, and local filesystem
- **Version Management**: Semantic versioning with update detection
- **Team Collaboration**: Share rulesets across projects and teams
- **Fast & Lightweight**: Single binary, no dependencies

## Common Commands

```bash
arm install [ruleset]           # Install rulesets
arm update                      # Update all rulesets
arm list                        # Show installed rulesets
arm outdated                    # Check for updates
arm config list                 # View configuration
```

For complete command documentation, see [Commands](commands.md).
