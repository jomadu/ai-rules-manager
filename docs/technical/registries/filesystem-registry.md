# Filesystem Registry Guide

## Overview

ARM's Filesystem registry provides direct access to local directory structures, making it ideal for development, testing, and offline usage. It scans directory hierarchies to discover available package versions.

## Features

- **Version Discovery**: Scans directory structure for available versions
- **Local Access**: Direct filesystem access without network dependencies
- **No Authentication**: Uses filesystem permissions
- **Development Friendly**: Perfect for local testing and development
- **Offline Support**: Works without internet connectivity

## Directory Structure

### Basic Structure
```
/path/to/registry/
├── typescript-rules/
│   ├── 1.0.0/
│   │   └── typescript-rules-1.0.0.tar.gz
│   ├── 1.0.1/
│   │   └── typescript-rules-1.0.1.tar.gz
│   └── 1.1.0/
│       └── typescript-rules-1.1.0.tar.gz
└── security-rules/
    ├── 2.0.0/
    │   └── security-rules-2.0.0.tar.gz
    └── 2.1.0/
        └── security-rules-2.1.0.tar.gz
```

### With Organization Scope
```
/path/to/registry/
├── company/
│   ├── typescript-rules/
│   │   ├── 1.0.0/
│   │   │   └── typescript-rules-1.0.0.tar.gz
│   │   └── 1.0.1/
│   │       └── typescript-rules-1.0.1.tar.gz
│   └── security-rules/
│       └── 1.0.0/
│           └── security-rules-1.0.0.tar.gz
└── opensource/
    └── lint-rules/
        └── 0.1.0/
            └── lint-rules-0.1.0.tar.gz
```

### Development Structure
```
~/Development/arm-packages/
├── my-rules/
│   ├── dev/
│   │   └── my-rules-dev.tar.gz
│   ├── 0.1.0/
│   │   └── my-rules-0.1.0.tar.gz
│   └── latest/
│       └── my-rules-latest.tar.gz
└── team-rules/
    └── 1.0.0/
        └── team-rules-1.0.0.tar.gz
```

## Configuration Examples

### Basic Filesystem Registry
```ini
[sources.local]
type = filesystem
path = /path/to/local/registry
```

### Development Registry
```ini
[sources.dev]
type = filesystem
path = ~/Development/arm-packages
```

### Shared Network Registry
```ini
[sources.shared]
type = filesystem
path = /mnt/shared/arm-registry
```

## Usage Examples

### Install Commands
```bash
# Install latest version (discovers from directories)
arm install typescript-rules

# Install specific version
arm install typescript-rules@1.0.0

# Install scoped package
arm install company@security-rules

# List available versions
arm list
```

### Configuration Usage
```bash
# Configure filesystem registry
arm config set sources.local /path/to/registry

# Install from configured source
arm install local@typescript-rules
```

## Version Discovery Process

### 1. Directory Scanning
ARM scans the package directory for version subdirectories:
```
/registry/typescript-rules/
├── 1.0.0/    # Discovered version
├── 1.0.1/    # Discovered version
└── 1.1.0/    # Discovered version
```

### 2. Version Extraction
ARM extracts version names from directory names:
- `1.0.0/` → version `1.0.0`
- `1.0.1/` → version `1.0.1`
- `dev/` → version `dev`
- `latest/` → version `latest`

### 3. Archive Location
ARM looks for archives in version directories:
```
/registry/typescript-rules/1.0.0/typescript-rules-1.0.0.tar.gz
```

## Publishing Workflow

### Manual Organization
```bash
# Create package structure
mkdir -p /registry/typescript-rules/1.0.0
cp typescript-rules-1.0.0.tar.gz /registry/typescript-rules/1.0.0/

# Create scoped package
mkdir -p /registry/company/security-rules/1.0.0
cp security-rules-1.0.0.tar.gz /registry/company/security-rules/1.0.0/
```

### Development Workflow
```bash
# Build and organize development package
npm run build
tar -czf my-rules-dev.tar.gz dist/
mkdir -p ~/Development/arm-packages/my-rules/dev
mv my-rules-dev.tar.gz ~/Development/arm-packages/my-rules/dev/

# Test installation
arm install dev@my-rules@dev
```

### Automated Build Script
```bash
#!/bin/bash
# build-and-publish.sh

PACKAGE_NAME="typescript-rules"
VERSION="1.0.0"
REGISTRY_PATH="/path/to/registry"

# Build package
npm run build
tar -czf ${PACKAGE_NAME}-${VERSION}.tar.gz dist/

# Create directory structure
mkdir -p ${REGISTRY_PATH}/${PACKAGE_NAME}/${VERSION}

# Move package to registry
mv ${PACKAGE_NAME}-${VERSION}.tar.gz ${REGISTRY_PATH}/${PACKAGE_NAME}/${VERSION}/

echo "Published ${PACKAGE_NAME}@${VERSION} to filesystem registry"
```

## Use Cases

### Local Development
```bash
# Set up development registry
mkdir -p ~/arm-dev-registry
arm config set sources.dev ~/arm-dev-registry

# Test packages locally before publishing
arm install dev@my-experimental-rules
```

### Team Shared Storage
```bash
# Mount shared network drive
sudo mount -t nfs server:/shared/arm-registry /mnt/arm-registry

# Configure shared registry
arm config set sources.team /mnt/arm-registry

# Team members can install shared packages
arm install team@shared-rules
```

### Offline Environments
```bash
# Pre-populate registry for offline use
rsync -av online-registry/ /offline/arm-registry/

# Configure offline registry
arm config set sources.offline /offline/arm-registry

# Works without internet
arm install offline@typescript-rules
```

### CI/CD Testing
```yaml
# GitHub Actions example
- name: Setup Test Registry
  run: |
    mkdir -p test-registry/test-rules/1.0.0
    cp test-package.tar.gz test-registry/test-rules/1.0.0/
    
- name: Test Installation
  run: |
    arm config set sources.test ./test-registry
    arm install test@test-rules@1.0.0
```

## Best Practices

### Directory Organization
- Use semantic versioning for version directories
- Keep consistent naming patterns
- Organize by team or project scope
- Use symbolic links for aliases (latest → 1.0.0)

### File Permissions
- Ensure read access for ARM user
- Use group permissions for team access
- Protect against accidental deletion
- Consider backup strategies

### Development Workflow
- Use separate development and production registries
- Test packages locally before sharing
- Use version tags like `dev`, `staging`, `latest`
- Implement automated build scripts

### Performance
- Keep registry on fast storage (SSD)
- Avoid deeply nested directory structures
- Clean up old versions periodically
- Use local paths for better performance

## Limitations

### No Network Access
Filesystem registries are local only:
- Cannot share across different machines without shared storage
- No remote access capabilities
- Requires physical or network filesystem access

### No Authentication
Uses filesystem permissions only:
- No user-based access control
- No token-based authentication
- Security depends on filesystem permissions

### No Rich Metadata
Provides basic metadata only:
- Version discovery from directory names
- No download counts or statistics
- No package descriptions or maintainer info

## Troubleshooting

### Common Issues
1. **Permission Denied**: Check filesystem permissions
2. **Path Not Found**: Verify registry path exists
3. **No Versions Found**: Check directory structure
4. **Archive Not Found**: Verify archive naming and location

### Debugging Steps
```bash
# Check registry path
ls -la /path/to/registry

# Check package structure
ls -la /path/to/registry/typescript-rules/

# Check permissions
stat /path/to/registry/typescript-rules/1.0.0/

# Test ARM configuration
arm config get sources.local
```

### Directory Structure Validation
```bash
# Validate expected structure
find /path/to/registry -name "*.tar.gz" -type f

# Check for proper nesting
find /path/to/registry -mindepth 3 -maxdepth 3 -name "*.tar.gz"
```

### Performance Issues
- **Slow scanning**: Reduce directory depth or file count
- **Large directories**: Split packages across multiple registries
- **Network storage**: Use local cache or faster network storage