# Troubleshooting Guide

Common issues and solutions for ARM.

## Installation Issues

### ARM Command Not Found
```bash
# Check if ARM is in PATH
which arm

# If not found, add to PATH or reinstall
export PATH=$PATH:/usr/local/bin
# or
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash
```

### Permission Denied During Installation
```bash
# Use sudo for system-wide installation
sudo curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash

# Or install to user directory
mkdir -p ~/bin
curl -L https://github.com/jomadu/ai-rules-manager/releases/latest/download/arm-linux-amd64 -o ~/bin/arm
chmod +x ~/bin/arm
export PATH=$PATH:~/bin
```

## Configuration Issues

### Registry Not Found
```bash
# Check registry configuration
arm config get registries.myregistry

# List all registries
arm config list | grep registries

# Add missing registry
arm config add registry myregistry https://github.com/org/repo --type=git
```

### Authentication Failures
```bash
# Check if token is set
echo $GITHUB_TOKEN

# Set token and retry
export GITHUB_TOKEN=your_token_here
arm config add registry private https://github.com/org/private --type=git --authToken=$GITHUB_TOKEN

# For S3 registries
aws configure list
export AWS_PROFILE=your-profile
```

### Invalid Configuration
```bash
# Validate configuration
arm config list

# Reset configuration if corrupted
rm .armrc arm.json
arm install  # Regenerates stub files
```

## Installation/Update Issues

### Ruleset Not Found
```bash
# Check available rulesets
arm search ruleset-name

# Check specific registry
arm info registry/ruleset-name

# Verify registry connectivity
arm info registry/any-ruleset --versions
```

### Version Not Found
```bash
# Check available versions
arm info ruleset-name --versions

# Use different version constraint
arm install ruleset-name@latest
arm install ruleset-name@^1.0.0
```

### Permission Denied Writing Files
```bash
# Check directory permissions
ls -la .cursor/rules/

# Create directory if missing
mkdir -p .cursor/rules .amazonq/rules

# Fix permissions
chmod 755 .cursor/rules .amazonq/rules
```

### Pattern Matching Issues
```bash
# Test patterns with dry run
arm install ruleset --patterns "rules/*.md" --dry-run

# Use simpler patterns
arm install ruleset --patterns "*.md"

# Check pattern syntax
arm install ruleset --patterns "**/*.md,!**/internal/**" --verbose
```

## Network Issues

### Connection Timeout
```bash
# Increase timeout
arm config set network.timeout 60

# Check network connectivity
curl -I https://github.com

# Use verbose mode for debugging
arm install ruleset --verbose
```

### Rate Limiting
```bash
# Check rate limit configuration
arm config get git.rateLimit

# Reduce concurrency
arm config set git.concurrency 1

# Wait and retry
sleep 60
arm install ruleset
```

### SSL/TLS Issues
```bash
# Check certificate validity
curl -v https://your-registry.com

# Temporarily allow insecure (not recommended)
arm install ruleset --insecure

# Update certificates
# On macOS: brew install ca-certificates
# On Ubuntu: sudo apt-get update && sudo apt-get install ca-certificates
```

## Cache Issues

### Cache Corruption
```bash
# Clear cache
arm clean cache

# Check cache location
arm config get cache.path

# Manually remove cache
rm -rf ~/.arm/cache
```

### Disk Space Issues
```bash
# Check cache size
du -sh ~/.arm/cache

# Reduce cache size limit
arm config set cache.maxSize 536870912  # 512MB

# Clean unused entries
arm clean unused
```

## Performance Issues

### Slow Downloads
```bash
# Check cache hit rate
arm install ruleset --verbose

# Increase cache TTL
arm config set cache.ttl 48h

# Use local registry for development
arm config add registry local /path/to/local/rules --type=local
```

### High Memory Usage
```bash
# Reduce concurrency
arm config set git.concurrency 1
arm config set s3.concurrency 5

# Clean cache regularly
arm clean cache
```

## Lock File Issues

### Corrupted Lock File
```bash
# Remove and regenerate
rm arm.lock
arm install

# Check lock file syntax
cat arm.lock | jq .
```

### Version Conflicts
```bash
# Check outdated rulesets
arm outdated

# Update conflicting rulesets
arm update

# Force reinstall
arm uninstall problematic-ruleset
arm install problematic-ruleset
```

## Debugging Commands

### Verbose Output
```bash
# Enable detailed logging
arm install ruleset --verbose

# Dry run to see planned actions
arm install ruleset --dry-run
```

### Configuration Debugging
```bash
# Show effective configuration
arm config list

# Test specific registry
arm info registry/test-ruleset

# Validate configuration
arm config list --json | jq .
```

### Network Debugging
```bash
# Test registry connectivity
curl -I https://api.github.com

# Check DNS resolution
nslookup github.com

# Test with different network
# (try mobile hotspot, different WiFi)
```

## Getting Help

### Log Collection
```bash
# Run with verbose output and save log
arm install problematic-ruleset --verbose 2>&1 | tee arm-debug.log

# Include configuration (remove sensitive data)
arm config list >> arm-debug.log
```

### System Information
```bash
# ARM version
arm version

# System information
uname -a
go version  # if building from source

# Environment
env | grep -E "(ARM_|GITHUB_|AWS_|HOME|PATH)"
```

### Common Error Patterns

#### "registry 'X' not found"
- Registry not configured in .armrc
- Typo in registry name
- Configuration file not loaded

#### "authentication failed"
- Missing or invalid token
- Token expired
- Wrong token type for registry

#### "version 'X' not found"
- Version doesn't exist
- Wrong version format
- Registry doesn't support versioning

#### "permission denied"
- Directory doesn't exist
- Insufficient permissions
- Read-only filesystem

#### "pattern 'X' matched no files"
- Incorrect pattern syntax
- Files don't exist in repository
- Case sensitivity issues

### Support Channels
- GitHub Issues: https://github.com/jomadu/ai-rules-manager/issues
- Documentation: https://github.com/jomadu/ai-rules-manager/docs
- Community: GitHub Discussions
