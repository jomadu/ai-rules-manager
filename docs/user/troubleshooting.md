# Troubleshooting Guide

Common issues and solutions when using ARM.

## Installation Issues

### Binary Not Found

**Problem**: `arm: command not found`

**Solution**:
```bash
# Check if binary is in PATH
which arm

# Add to PATH if needed
export PATH=$PATH:/usr/local/bin

# Or move binary to PATH location
sudo mv arm /usr/local/bin/
```

### Permission Denied

**Problem**: `permission denied: ./arm`

**Solution**:
```bash
chmod +x arm
```

## Configuration Issues

### Registry Not Found

**Problem**: `registry not found: company`

**Solution**:
```bash
# Check configuration
arm config list

# Add missing registry
arm config set sources.company https://internal.company.local/
```

### Authentication Failed

**Problem**: `authentication failed for registry`

**Solution**:
```bash
# Check environment variables
echo $GITLAB_TOKEN
echo $AWS_ACCESS_KEY_ID

# Set missing tokens
export GITLAB_TOKEN="your-token-here"
```

## Installation Issues

### Package Not Found

**Problem**: `package not found: typescript-rules`

**Solutions**:
1. Check package name spelling
2. Verify registry contains the package
3. Check registry configuration

```bash
# Debug mode for more details
export ARM_DEBUG=1
arm install typescript-rules
```

### Version Constraint Issues

**Problem**: `no compatible version found`

**Solutions**:
1. Check available versions: `arm outdated typescript-rules`
2. Relax version constraints in `rules.json`
3. Use `latest` version

```json
{
  "dependencies": {
    "typescript-rules": "latest"
  }
}
```

### Network Issues

**Problem**: `connection timeout` or `network unreachable`

**Solutions**:
1. Check internet connection
2. Verify registry URLs are accessible
3. Check corporate firewall/proxy settings

```bash
# Test registry connectivity
curl -I https://registry.armjs.org/

# Use proxy if needed
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080
```

## File System Issues

### Permission Denied

**Problem**: `permission denied: .cursorrules`

**Solutions**:
```bash
# Check file permissions
ls -la .cursorrules

# Fix permissions
chmod 755 .cursorrules
```

### Disk Space

**Problem**: `no space left on device`

**Solutions**:
```bash
# Check disk space
df -h

# Clean ARM cache
arm clean --cache

# Clean project targets
arm clean
```

## Lock File Issues

### Corrupted Lock File

**Problem**: `lock file corrupted`

**Solution**:
```bash
# Remove and regenerate
rm rules.lock
arm install
```

### Lock File Conflicts

**Problem**: Git merge conflicts in `rules.lock`

**Solution**:
```bash
# Delete lock file and reinstall
rm rules.lock
arm install
git add rules.lock
git commit -m "fix: regenerate lock file"
```

## Update Issues

### Update Failures

**Problem**: Updates fail partway through

**Solutions**:
1. Use dry-run to check what would be updated
2. Update one ruleset at a time
3. Check for breaking changes

```bash
# Check what would be updated
arm update --dry-run

# Update specific ruleset
arm update typescript-rules
```

### Version Conflicts

**Problem**: `version conflict detected`

**Solution**:
```bash
# Check outdated packages
arm outdated

# Update rules.json constraints
# Then reinstall
rm rules.lock
arm install
```

## Performance Issues

### Slow Downloads

**Problem**: Downloads are very slow

**Solutions**:
1. Increase concurrency settings
2. Use closer registry mirrors
3. Check network bandwidth

```bash
# Increase concurrency
arm config set performance.defaultConcurrency 8

# Check registry performance
time curl -I https://registry.armjs.org/
```

### High Memory Usage

**Problem**: ARM uses too much memory

**Solutions**:
1. Reduce concurrency
2. Clean cache regularly
3. Process fewer packages at once

```bash
# Reduce concurrency
arm config set performance.defaultConcurrency 2

# Clean cache
arm clean --cache
```

## Debug Mode

Enable debug mode for detailed logging:

```bash
export ARM_DEBUG=1
arm install typescript-rules
```

## Getting Help

### Check Version

```bash
arm version
```

### View Configuration

```bash
arm config list
```

### Check Installed Packages

```bash
arm list --format=json
```

### Test Registry Connectivity

```bash
# Test default registry
curl -I https://registry.armjs.org/

# Test custom registry
curl -H "Authorization: Bearer $TOKEN" -I https://internal.company.local/
```

## Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `registry not found` | Missing registry config | Add registry to `.armrc` |
| `authentication failed` | Invalid/missing token | Check environment variables |
| `package not found` | Wrong package name/registry | Verify package exists |
| `version constraint not satisfied` | No compatible version | Update version constraints |
| `permission denied` | File/directory permissions | Fix file permissions |
| `network timeout` | Network connectivity | Check internet/proxy settings |
| `disk full` | No disk space | Clean cache or free disk space |

## Reporting Issues

When reporting issues, include:

1. ARM version: `arm version`
2. Operating system and version
3. Configuration: `arm config list`
4. Error message and full command
5. Debug output: `ARM_DEBUG=1 arm <command>`
