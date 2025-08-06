# Installation

## Quick Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash
```

This script automatically detects your OS/architecture and installs ARM to `/usr/local/bin/arm`.

## Manual Installation

### Linux/macOS

1. Download the latest release:
```bash
# Linux
wget https://github.com/jomadu/ai-rules-manager/releases/latest/download/arm-linux-amd64.tar.gz
tar -xzf arm-linux-amd64.tar.gz

# macOS
wget https://github.com/jomadu/ai-rules-manager/releases/latest/download/arm-darwin-amd64.tar.gz
tar -xzf arm-darwin-amd64.tar.gz
```

2. Make executable and move to PATH:
```bash
chmod +x arm
sudo mv arm /usr/local/bin/arm
```

### Windows

1. Download `arm-windows-amd64.zip` from [releases](https://github.com/jomadu/ai-rules-manager/releases/latest)
2. Extract `arm.exe` to a directory in your PATH

## Verify Installation

```bash
arm version
arm help
```

## Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/arm

# Remove cache (optional)
rm -rf ~/.arm
```

## Next Steps

Continue to [Getting Started](getting-started.md) to install your first ruleset.
