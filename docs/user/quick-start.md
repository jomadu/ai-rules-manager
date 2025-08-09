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

Add a Git registry (we'll use a fake Mario-themed example):

```bash
arm config add registry mushroom-kingdom https://github.com/mushroom-kingdom/cursor-rules.example --type=git
```

Set it as your default registry:

```bash
arm config set registries.default mushroom-kingdom
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

Install a Mario-themed ruleset:

```bash
arm install power-up-rules --patterns "rules/*.md"
```

### 5. Verify Installation

Check what's installed:

```bash
arm list
```

You should see:
```
cursor:
  mushroom-kingdom:
    - power-up-rules@1.0.0

q:
  mushroom-kingdom:
    - power-up-rules@1.0.0
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
Error [REGISTRY]: Registry 'mushroom-kingdom' not configured
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
