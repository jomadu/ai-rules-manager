# Quick Start Guide

Get ARM up and running in 5 minutes with your first ruleset installation.

## Prerequisites

- macOS, Linux, or Windows
- Git (for Git-based registries)
- Internet connection

## Installation

### Option 1: Install Script (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash
```

### Option 2: Manual Download
1. Download the latest release from [GitHub Releases](https://github.com/jomadu/ai-rules-manager/releases)
2. Extract and move `arm` to your PATH

### Option 3: Build from Source
```bash
git clone https://github.com/jomadu/ai-rules-manager.git
cd ai-rules-manager
make build
sudo mv bin/arm /usr/local/bin/
```

## Initial Setup

### 1. Initialize Configuration
```bash
arm install
```
This creates stub configuration files if they don't exist.

### 2. Add a Registry
```bash
# Add a Git registry
arm config add registry default https://github.com/your-org/ai-rules-registry --type=git

# Or add an S3 registry
arm config add registry s3-rules my-rules-bucket --type=s3 --region=us-east-1
```

### 3. Add Channels for Your AI Tools
```bash
# For Cursor
arm config add channel cursor --directories .cursor/rules

# For Amazon Q Developer
arm config add channel q --directories .amazonq/rules

# For both
arm config add channel cursor --directories .cursor/rules
arm config add channel q --directories .amazonq/rules
```

## Install Your First Ruleset

### From Git Registry
```bash
# Install with specific patterns
arm install my-coding-rules --patterns "rules/*.md,guidelines/*.md"

# Install specific version
arm install my-coding-rules@v1.2.0 --patterns "rules/*.md"
```

### From S3 Registry
```bash
# Install latest version
arm install team-standards

# Install to specific channels
arm install team-standards --channels cursor,q
```

## Verify Installation

### Check Installed Rulesets
```bash
arm list
```

### Check for Updates
```bash
arm outdated
```

### Check Configuration
```bash
arm config list
```

### View Ruleset Information
```bash
arm info my-coding-rules
```

## Example Workflow

Here's a complete example setting up ARM for a team:

```bash
# 1. Install ARM
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash

# 2. Initialize
arm install

# 3. Add team registry
arm config add registry team https://github.com/myteam/ai-rules --type=git

# 4. Add channels for both Cursor and Amazon Q
arm config add channel cursor --directories .cursor/rules
arm config add channel q --directories .amazonq/rules

# 5. Install team coding standards
arm install coding-standards --patterns "standards/*.md,best-practices/*.md"

# 6. Install security rules
arm install security-rules --patterns "security/*.md"

# 7. Verify installation
arm list
```

## File Structure After Setup

```
your-project/
├── .armrc                 # Registry and channel configuration
├── arm.json              # Rulesets and engines
├── arm.lock              # Locked versions with patterns for Git registries
│   # Example content:
│   # {
│   #   "rulesets": {
│   #     "team": {
│   #       "coding-standards": {
│   #         "version": "latest",
│   #         "resolved": "abc123...",
│   #         "patterns": ["standards/*.md"],
│   #         "type": "git"
│   #       }
│   #     }
│   #   }
│   # }
├── .cursor/rules/        # Cursor rules
│   └── arm/              # ARM namespace
│       └── team/         # Registry name
│           ├── coding-standards/
│           │   └── latest/       # Version
│           │       ├── standards/
│           │       │   └── clean-code.md
│           │       └── best-practices/
│           │           └── naming-conventions.md
│           └── security-rules/
│               └── latest/       # Version
│                   └── security/
│                       ├── input-validation.md
│                       ├── auth-patterns.md
│                       └── secure-coding.md
└── .amazonq/rules/       # Amazon Q rules
    └── arm/              # ARM namespace
        └── team/         # Registry name
            ├── coding-standards/
            │   └── latest/       # Version
            │       ├── standards/
            │       │   └── clean-code.md
            │       └── best-practices/
            │           └── naming-conventions.md
            └── security-rules/
                └── latest/       # Version
                    └── security/
                        ├── input-validation.md
                        ├── auth-patterns.md
                        └── secure-coding.md
```

## Next Steps

- **[Configuration Guide](configuration.md)** - Learn about advanced configuration options
- **[Registry Guide](registries.md)** - Set up different types of registries
- **[Team Setup Guide](team-setup.md)** - Deploy ARM across your development team
- **[Usage Guide](usage.md)** - Explore all available commands

## Common Issues

### Registry Not Found
```bash
# Check your registry configuration
arm config get registries.default

# List all registries
arm config list
```

### Permission Denied
```bash
# Check directory permissions
ls -la .cursor/rules/

# Create directory if it doesn't exist
mkdir -p .cursor/rules
```

### Git Authentication
```bash
# Set up GitHub token for private repositories
export GITHUB_TOKEN=your_token_here
arm config add registry private https://github.com/org/private-repo --type=git --authToken=$GITHUB_TOKEN
```

Need help? Check the [Troubleshooting Guide](troubleshooting.md) or open an issue on GitHub.
