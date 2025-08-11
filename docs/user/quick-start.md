# Quick Start Guide

Get ARM up and running in 5 minutes with Mario-themed examples.

## Installation

Download and install ARM:

```bash
# Install via script
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash

# Or download binary manually
wget https://github.com/jomadu/ai-rules-manager/releases/latest/download/arm-linux-amd64.tar.gz
tar -xzf arm-linux-amd64.tar.gz
chmod +x arm-linux-amd64
sudo mv arm-linux-amd64 /usr/local/bin/arm
```

## First Steps

### 1. Initialize Configuration

Generate starter configuration files:

```bash
arm install
```

This creates `.armrc` and `arm.json` stub files in your current directory.

### 2. Configure Your First Registry

Add a Git registry:

```bash
arm config add registry default https://github.com/jomadu/ai-rules-manager-test-git-registry --type=git
```

### 3. Configure Channels

Tell ARM where to install rulesets for your AI tools:

```bash
# For Cursor
arm config add channel cursor --directories ~/.cursor/rules

# For Amazon Q Developer
arm config add channel q --directories ~/.aws/amazonq/rules
```

### 4. Install Your First Ruleset

Install a ruleset:

```bash
arm install rules --patterns "*.md"
```

### 5. Verify Installation

Check what's installed:

```bash
arm list
```

You should see:
```
cursor:
  default:
    - rules@2.0.0

q:
  default:
    - rules@2.0.0
```

## What Just Happened?

1. **ARM created config files** with your registry and channel settings
2. **Downloaded ruleset files** from the Git repository using your patterns
3. **Installed files** to both Cursor and Q Developer directories
4. **Created a lock file** (`arm.lock`) to track installed versions

## Next Steps

- **[Configuration Guide](configuration.md)** - Deep dive into .armrc and arm.json
- **[Registry Guide](registries.md)** - Set up different registry types
- **[Team Guide](team-setup.md)** - Share rulesets across your team

## Troubleshooting

### Registry Not Found
```bash
$ arm install power-up-rules
Error [REGISTRY]: Registry 'default' not configured
```
**Solution**: Make sure you've added the registry with `arm config add registry`.

### No Channels Configured
```bash
$ arm install power-up-rules
Error: no channels configured
```
**Solution**: Add at least one channel with `arm config add channel`.

### Permission Denied
```bash
$ arm install power-up-rules
Error [FILESYSTEM]: Permission denied writing to ~/.cursor/rules
```
**Solution**: Check directory permissions or create the directory manually:
```bash
mkdir -p ~/.cursor/rules
```
