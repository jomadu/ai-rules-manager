# Configuration Guide

Complete guide to configuring ARM for your development environment and team.

## Configuration Files

ARM uses a hierarchical configuration system with multiple file types:

### File Types
- **`.armrc`** - INI format for registries and system settings
- **`arm.json`** - JSON format for channels and rulesets
- **`arm.lock`** - JSON format for locked versions (auto-generated)

### File Locations
```
Global:  ~/.arm/.armrc, ~/.arm/arm.json
Local:   .armrc, arm.json, arm.lock
```

### Hierarchy
Local configuration overrides global configuration at the key level.

## Registry Configuration

### Adding Registries

#### Git Registry (GitHub/GitLab)
```bash
# Public repository
arm config add registry public https://github.com/org/rules-repo --type=git

# Private repository with authentication
arm config add registry private https://github.com/org/private-rules --type=git --authToken=$GITHUB_TOKEN

# GitLab with API configuration
arm config add registry gitlab https://gitlab.com/org/rules --type=git --authToken=$GITLAB_TOKEN --apiType=gitlab
```

#### S3 Registry
```bash
# Basic S3 registry
arm config add registry s3-rules my-bucket --type=s3 --region=us-east-1

# S3 with custom profile and prefix
arm config add registry s3-team team-rules-bucket --type=s3 --region=us-west-2 --profile=team-profile --prefix=/rules/
```

#### Local Development Registry
```bash
# Local Git repository
arm config add registry local-dev /path/to/local/repo --type=git-local

# Local directory
arm config add registry local-files /path/to/rules --type=local
```

### Registry Configuration Details

#### .armrc Format
```ini
[registries]
default = https://github.com/myteam/ai-rules
s3-prod = production-rules-bucket
local-dev = /Users/dev/rules

[registries.default]
type = git
authToken = $GITHUB_TOKEN

[registries.s3-prod]
type = s3
region = us-east-1
profile = production

[registries.local-dev]
type = git-local
```

## Channel Configuration

Channels define where rulesets are installed in your project.

### Adding Channels
```bash
# Single directory
arm config add channel cursor --directories .cursor/rules

# Multiple directories
arm config add channel multi --directories ".cursor/rules,.amazonq/rules"

# With environment variables
arm config add channel home --directories "$HOME/.cursor/rules"
```

### Channel Configuration in arm.json
```json
{
  "channels": {
    "cursor": {
      "directories": [".cursor/rules"]
    },
    "q": {
      "directories": [".amazonq/rules"]
    },
    "both": {
      "directories": [".cursor/rules", ".amazonq/rules"]
    }
  }
}
```

## Ruleset Configuration

### Adding Rulesets to Manifest

#### Manual Configuration
Edit `arm.json`:
```json
{
  "rulesets": {
    "default": {
      "coding-standards": {
        "version": "^1.0.0",
        "patterns": ["rules/*.md", "guidelines/*.md"]
      },
      "security-rules": {
        "version": "latest"
      }
    },
    "s3-prod": {
      "team-standards": {
        "version": ">=2.0.0"
      }
    }
  }
}
```

#### Command-Line Installation
```bash
# Install and add to manifest
arm install coding-standards@^1.0.0 --patterns "rules/*.md"

# Install from specific registry
arm install s3-prod/team-standards@latest
```

## Advanced Configuration

### Environment Variables

#### Authentication
```bash
# Git registries
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
export GITLAB_TOKEN=glpat-xxxxxxxxxxxx

# S3 registries
export AWS_PROFILE=my-profile
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=AKIAXXXXXXXX
export AWS_SECRET_ACCESS_KEY=xxxxxxxx

# Custom registries
export REGISTRY_TOKEN=token_xxxxxxxxxxxx
```



### Network Configuration

#### .armrc Network Settings
```ini
[network]
timeout = 30
retry.maxAttempts = 3
retry.backoffMultiplier = 2.0
retry.maxBackoff = 30
```

#### Registry-Specific Settings
```ini
[git]
concurrency = 1
rateLimit = 10/minute

[s3]
concurrency = 10
rateLimit = 100/hour

[https]
concurrency = 5
rateLimit = 30/minute
```

### Cache Configuration

```ini
[cache]
path = $HOME/.arm/cache
maxSize = 1073741824           # 1GB in bytes (0 = unlimited)
ttl = 24h                      # Time-to-live for cache entries
cleanupInterval = 6h           # How often to run cleanup
```

## Engine Configuration

Specify ARM version requirements:

```json
{
  "engines": {
    "arm": "^1.0.0"
  }
}
```

## Configuration Commands

### Viewing Configuration
```bash
# List all configuration
arm config list

# Get specific value
arm config get registries.default
arm config get git.concurrency

# Check effective configuration (merged)
arm config list
```

### Setting Values
```bash
# Set registry URL
arm config set registries.default https://github.com/new/repo

# Set network timeout
arm config set network.timeout 60

# Set cache size
arm config set cache.maxSize 2147483648  # 2GB
```

### Removing Configuration
```bash
# Remove registry
arm config remove registry old-registry

# Remove channel
arm config remove channel unused-channel
```

## Team Configuration

### Shared Configuration Strategy

#### Recommended: Security-First Approach
```bash
# Commit these files to your repository
arm.json        # Channels and rulesets (safe)
arm.lock        # Locked versions (for reproducible builds)
# Don't commit .armrc (may contain sensitive registry configs)
# Use global ~/.arm/.armrc for team registry setup
```



### Environment-Specific Configuration

#### Development
```json
{
  "channels": {
    "dev": {
      "directories": [".cursor/rules", ".amazonq/rules"]
    }
  },
  "rulesets": {
    "team": {
      "coding-standards": {"version": "latest"},
      "dev-tools": {"version": "^1.0.0"}
    }
  }
}
```

#### Production
```json
{
  "channels": {
    "prod": {
      "directories": [".cursor/rules"]
    }
  },
  "rulesets": {
    "team": {
      "coding-standards": {"version": "1.2.3"},
      "security-rules": {"version": "2.1.0"}
    }
  }
}
```

## Configuration Validation

### Automatic Validation
ARM validates configuration on load:
- Registry type compatibility
- Required fields for each registry type
- Channel directory accessibility
- Version constraint syntax

### Manual Validation
```bash
# Test registry connectivity
arm info team/coding-standards

# Verify channel access
arm list --channels cursor

# Check configuration syntax
arm config list
```

## Troubleshooting Configuration

### Common Issues

#### Registry Not Found
```bash
# Check registry configuration
arm config get registries.myregistry

# List all registries
arm config list | grep registries
```

#### Authentication Failures
```bash
# Verify token is set
echo $GITHUB_TOKEN

# Test with explicit token
arm config add registry test https://github.com/org/repo --type=git --authToken=your_token
```

#### Permission Denied
```bash
# Check directory permissions
ls -la .cursor/rules/

# Create missing directories
mkdir -p .cursor/rules .amazonq/rules
```

#### Version Conflicts
```bash
# Check locked versions
cat arm.lock

# Update to resolve conflicts
arm update
```

### Configuration Reset
```bash
# Reset local configuration
rm .armrc arm.json arm.lock

# Reset global configuration
rm ~/.arm/.armrc ~/.arm/arm.json

# Regenerate defaults
arm install
```

## Best Practices

### Security
- Store tokens in environment variables, not config files
- Use least-privilege AWS IAM policies for S3 registries
- Regularly rotate authentication tokens
- Use HTTPS-only registries in production

### Organization
- Use descriptive registry names
- Group related rulesets in the same registry
- Document registry purposes in team documentation
- Use semantic versioning for ruleset versions

### Performance
- Configure appropriate cache sizes for your team
- Use local registries for development
- Set reasonable concurrency limits
- Monitor network usage and adjust rate limits

### Maintenance
- Regularly update ARM version
- Clean unused cache entries
- Review and update locked versions
- Monitor registry health and availability
