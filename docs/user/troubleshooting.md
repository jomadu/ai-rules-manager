# Troubleshooting

Common issues and solutions for ARM.

## Installation Issues

### "arm: command not found"

ARM is not in your PATH.

**Solution:**
```bash
# Check if ARM is installed
which arm

# If not found, reinstall
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash

# Or add to PATH manually
export PATH="/usr/local/bin:$PATH"
```

### Permission denied during installation

Installation script needs sudo access.

**Solution:**
```bash
# Run with explicit sudo
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | sudo bash

# Or install manually to user directory
mkdir -p ~/bin
wget https://github.com/jomadu/ai-rules-manager/releases/latest/download/arm-linux-amd64.tar.gz
tar -xzf arm-linux-amd64.tar.gz -C ~/bin
export PATH="$HOME/bin:$PATH"
```

## Registry Issues

### "No registries configured"

ARM needs at least one registry configured before installing rulesets.

**Solution:**
```bash
# Configure a registry first
arm config set sources.company https://gitlab.company.com
arm config set sources.company.type gitlab

# Or create .armrc manually
cat > ~/.armrc << EOF
[sources]
company = https://gitlab.company.com

[sources.company]
type = gitlab
projectID = 12345
EOF
```

### "Authentication failed"

Registry authentication is incorrect or expired.

**Solution:**
```bash
# Check current config
arm config list

# Update authentication token
arm config set sources.company.authToken $NEW_TOKEN

# For GitLab, ensure token has read_api scope
# For S3, check AWS credentials
aws sts get-caller-identity
```

### "Ruleset not found"

The requested ruleset doesn't exist in the registry.

**Solution:**
```bash
# Check available rulesets (if registry supports listing)
arm search typescript

# Verify registry configuration
arm config get sources.company

# Try explicit registry
arm install company@typescript-rules
```

## Network Issues

### "Connection timeout"

Network connectivity or firewall issues.

**Solution:**
```bash
# Test connectivity
curl -I https://gitlab.company.com

# Check proxy settings
echo $HTTP_PROXY
echo $HTTPS_PROXY

# Configure proxy if needed
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080
```

### "SSL certificate verification failed"

SSL/TLS certificate issues.

**Solution:**
```bash
# Update CA certificates
sudo apt-get update && sudo apt-get install ca-certificates

# For corporate networks, add custom CA
sudo cp company-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Temporary workaround (not recommended)
export ARM_SKIP_TLS_VERIFY=true
```

## File System Issues

### "Permission denied" writing to target directories

ARM can't write to `.cursorrules` or `.amazonq/rules`.

**Solution:**
```bash
# Check directory permissions
ls -la .cursorrules .amazonq/rules

# Fix permissions
chmod 755 .cursorrules .amazonq/rules

# Create directories if missing
mkdir -p .cursorrules .amazonq/rules
```

### "Disk space full"

Insufficient disk space for cache or target directories.

**Solution:**
```bash
# Check disk space
df -h

# Clean ARM cache
arm clean --all

# Remove old rulesets
arm uninstall old-ruleset
```

## Version Conflicts

### "Version conflict detected"

Multiple rulesets require incompatible versions of the same dependency.

**Solution:**
```bash
# Check current versions
arm list

# Update rules.json with compatible versions
# Change "^1.0.0" to "^2.0.0" if needed

# Force update
arm update --force
```

### "No compatible version found"

Requested version range has no matching releases.

**Solution:**
```bash
# Check available versions
arm info typescript-rules

# Use broader version range in rules.json
"typescript-rules": "*"

# Or specify exact version
arm install typescript-rules@1.2.3
```

## Getting Help

### Enable Verbose Output

```bash
arm --verbose install typescript-rules
arm --verbose config list
```

### Check Configuration

```bash
# View all settings
arm config list

# Check specific values
arm config get sources
arm config get performance
```

### Reset Configuration

```bash
# Remove global config
rm ~/.armrc

# Remove project config
rm rules.json rules.lock

# Clear cache
arm clean --all
```

### Report Issues

If problems persist:

1. Run with `--verbose` flag
2. Check ARM version: `arm version`
3. Include error output and configuration
4. Report at: https://github.com/jomadu/ai-rules-manager/issues
