# Getting Started

ARM manages AI coding assistant rulesets across projects.

## Installation

```bash
# Download and install
wget https://github.com/user/arm/releases/latest/download/arm-linux-amd64
chmod +x arm-linux-amd64
sudo mv arm-linux-amd64 /usr/local/bin/arm
```

## Quick Start

1. **Create rules.json**:
```json
{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "typescript-rules": "^1.0.0"
  }
}
```

2. **Install rulesets**:
```bash
arm install
```

3. **Update rulesets**:
```bash
arm update
```

## Commands

| Command | Description |
|---------|-------------|
| `arm install [ruleset]` | Install rulesets |
| `arm list` | Show installed |
| `arm update` | Update all |
| `arm outdated` | Check updates |
| `arm clean` | Remove unused |

## Next Steps

- [Configuration](configuration.md) - Registry setup
- [Registries](registries/) - Registry types
- [Troubleshooting](troubleshooting.md) - Common issues
