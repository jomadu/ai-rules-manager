# Filesystem Registry

Use local directories for ARM rulesets during development.

## Setup

```ini
[sources.local]
type = filesystem
path = ./local-registry
```

## Directory Structure

```
local-registry/
└── ruleset-name/
    ├── 1.0.0/
    │   ├── rule.md
    │   └── package.json
    └── versions.json
```

## Creating Registry

```bash
# Create structure
mkdir -p local-registry/typescript-rules/1.0.0

# Add rules
echo "# TypeScript Rules" > local-registry/typescript-rules/1.0.0/rules.md

# Add metadata
cat > local-registry/typescript-rules/1.0.0/package.json << EOF
{"name": "typescript-rules", "version": "1.0.0"}
EOF

# Add versions
cat > local-registry/typescript-rules/versions.json << EOF
{"versions": ["1.0.0"], "latest": "1.0.0"}
EOF
```

## Usage

```bash
arm install local@typescript-rules
```

## Use Cases

- **Development**: Test rules before publishing
- **Offline**: Work without internet
- **Custom**: Organization-specific rules

## Troubleshooting

- **Path Not Found**: Check path exists and is readable
- **Version Discovery Failed**: Ensure versions.json exists
- **Permission Denied**: Check file permissions
