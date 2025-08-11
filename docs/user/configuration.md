# Configuration Guide

Complete guide to configuring ARM with `.armrc` and `arm.json` files.

## Configuration Files

ARM uses two configuration files:
- **`.armrc`** - Registry settings and defaults (INI format)
- **`arm.json`** - Channels and rulesets (JSON format)

## File Locations

### Global vs Local
- **Global**: `~/.arm/.armrc` and `~/.arm/arm.json`
- **Local**: `./.armrc` and `./arm.json` (current directory)

Local settings override global settings at the key level.

### Generating Stubs

Create starter configuration files:

```bash
# Generate in current directory
arm install

# Generate globally
arm install --global
```

## Registry Configuration (.armrc)

### Basic Registry Setup

```bash
# Add different registry types
arm config add registry bowser-castle https://github.com/bowser-castle/security-rules.example --type=git
arm config add registry koopa-troopa koopa-castle-rules --type=s3 --region=us-east-1
arm config add registry toad-house https://gitlab.mushroom-kingdom.example/projects/456 --type=gitlab --authToken=$GITLAB_TOKEN
```

### Registry-Specific Settings

Override defaults for specific registries:

```bash
# Increase concurrency for a fast registry
arm config set registries.bowser-castle.concurrency 5

# Set custom rate limit
arm config set registries.koopa-troopa.rateLimit 20/minute

# Configure S3 profile
arm config set registries.koopa-troopa.profile mario-aws-profile
```

### Type Defaults

Set defaults for all registries of a type:

```bash
# All Git registries
arm config set git.concurrency 2
arm config set git.rateLimit 15/minute

# All S3 registries
arm config set s3.concurrency 10
arm config set s3.rateLimit 100/hour
```

### Environment Variables

Use environment variables in configuration:

```bash
arm config add registry private-castle https://github.com/private/repo.example --type=git --authToken=$GITHUB_TOKEN
```

## Channel Configuration (arm.json)

### Adding Channels

```bash
# Single directory
arm config add channel cursor --directories ~/.cursor/rules

# Multiple directories
arm config add channel cursor --directories "~/.cursor/rules,./project-rules"

# Environment variables supported
arm config add channel q --directories '$HOME/.aws/amazonq/rules'

# GitHub Copilot channel (uses .github directory)
arm config add channel copilot --directories .github
```

### Manual JSON Editing

You can also edit `arm.json` directly:

```json
{
  "engines": {
    "arm": "^1.2.3"
  },
  "channels": {
    "cursor": {
      "directories": ["~/.cursor/rules", "./custom-cursor"]
    },
    "q": {
      "directories": ["$HOME/.aws/amazonq/rules"]
    },
    "copilot": {
      "directories": [".github"]
    }
  },
  "rulesets": {
    "mushroom-kingdom": {
      "power-up-rules": {
        "version": "^1.0.0",
        "patterns": ["rules/*.md", "docs/*.mdc"]
      }
    }
  }
}
```

### GitHub Copilot Configuration

GitHub Copilot supports custom instructions, chat participants, and prompts through files placed in the `.github` directory:

```bash
# Add Copilot channel
arm config add channel copilot --directories .github

# Install rulesets that include Copilot configurations
arm install copilot-rules --patterns "copilot-*.md,copilot-*.yml"
```

**Supported Copilot Files:**
- `copilot-instructions.md` - General instructions for Copilot behavior
- `copilot-chat-participants.yml` - Custom chat participants and commands
- `copilot-prompts.yml` - Reusable prompts for common tasks

**Example Channel Configuration:**
```json
{
  "channels": {
    "copilot": {
      "directories": [".github"]
    }
  }
}
```

## Configuration Commands

### View Configuration

```bash
# Show merged configuration
arm config list

# Show specific value
arm config get registries.default

# Show global only
arm config list --global
```

### Modify Configuration

```bash
# Set values
arm config set registries.default mushroom-kingdom
arm config set git.concurrency 3

# Remove registries/channels
arm config remove registry old-registry
arm config remove channel unused-channel
```

## Advanced Configuration

### Network Settings

```bash
# Increase timeout for slow connections
arm config set network.timeout 60

# Configure retry behavior
arm config set network.retry.maxAttempts 5
arm config set network.retry.backoffMultiplier 2.0
```

### Cache Settings

```bash
# Change cache location
arm config set cache.path ~/my-arm-cache

# Increase cache size
arm config set cache.maxSize 2GB

# Adjust TTL (time to live)
arm config set cache.ttl 7200
```

## Configuration Validation

ARM validates configuration when you run commands:

```bash
# Check configuration
arm config list
```

Common validation errors and fixes are shown automatically.

## Troubleshooting

### Invalid Registry Type
```bash
Error [CONFIG]: Unknown registry type 'invalid' in registry 'my-registry'
Details: Supported types: git, https, s3, gitlab, local
```
**Solution**: Use a supported registry type.

### Missing Required Field
```bash
Error [CONFIG]: Missing required field 'region' for S3 registry
```
**Solution**: Add the required field:
```bash
arm config set registries.koopa-troopa.region us-east-1
```

### Environment Variable Not Found
If `$GITHUB_TOKEN` is not set, ARM will use an empty string. Set the variable:
```bash
export GITHUB_TOKEN=your_token_here
```
