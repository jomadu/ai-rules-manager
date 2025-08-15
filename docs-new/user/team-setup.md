# Team Setup Guide

Deploy ARM across your development team for consistent AI coding rules.

## Team Configuration Strategy

### Recommended: Security-First Approach
Commit safe configuration files to your repository for team consistency.

```bash
# Files to commit
arm.json        # Channels and rulesets (safe)
arm.lock        # Locked versions (for reproducible builds)

# Files NOT to commit
.armrc          # May contain sensitive registry configs
                # Use global ~/.arm/.armrc for team registry setup
```

### Alternative: Global + Local Split
Use global configuration for team defaults with sensitive data, local for project-specific overrides.

## Step-by-Step Team Setup

### 1. Create Team Registry

#### GitHub Organization
```bash
# Create a repository for your team's rules
# Example: https://github.com/company/ai-coding-rules

# Structure:
# ai-coding-rules/
# ├── standards/
# │   ├── clean-code.md
# │   └── security.md
# ├── guidelines/
# │   └── best-practices.md
# └── README.md
```

#### S3 Bucket (Enterprise)
```bash
# Create S3 bucket for team rules
aws s3 mb s3://company-ai-rules --region us-east-1

# Upload rulesets
aws s3 sync ./rulesets s3://company-ai-rules/
```

### 2. Configure Team Registry

#### Global ~/.arm/.armrc (Team Registry Setup)
```ini
[registries]
team = https://github.com/company/ai-coding-rules

[registries.team]
type = git
authToken = $GITHUB_TOKEN
```

#### arm.json (Project Configuration)
```json
{
  "engines": {
    "arm": "^1.0.0"
  },
  "channels": {
    "cursor": {
      "directories": [".cursor/rules"]
    },
    "q": {
      "directories": [".amazonq/rules"]
    }
  },
  "rulesets": {
    "team": {
      "coding-standards": {
        "version": "^1.0.0",
        "patterns": ["standards/*.md", "guidelines/*.md"]
      },
      "security-rules": {
        "version": "latest",
        "patterns": ["security/*.md"]
      }
    }
  }
}
```

### 3. Team Onboarding Script

Create `setup-arm.sh` for new team members:

```bash
#!/bin/bash
set -e

echo "Setting up ARM for team development..."

# Install ARM
curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash

# Verify installation
arm version

# Install team rulesets
echo "Installing team rulesets..."
arm install

echo "ARM setup complete!"
echo "Team rulesets installed in .cursor/rules and .amazonq/rules"
```

## Authentication Management

### GitHub Token Setup
```bash
# Each team member creates a personal access token
# with 'repo' scope for private repositories

# Set in environment (recommended)
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx

# Or in shell profile
echo 'export GITHUB_TOKEN=ghp_xxxxxxxxxxxx' >> ~/.bashrc

# Configure in global .armrc (not committed)
arm config add registry team https://github.com/company/ai-coding-rules --type=git --authToken=$GITHUB_TOKEN
```

### AWS Credentials (S3)
```bash
# Configure AWS CLI
aws configure --profile company

# Or use environment variables
export AWS_PROFILE=company
export AWS_REGION=us-east-1
```

## Monitoring and Maintenance

### Regular Updates
```bash
# Weekly team rule updates
arm update

# Check for outdated rules
arm outdated
```

### Cache Management
```bash
# Clean unused cache periodically
arm clean unused

# Monitor cache size
du -sh ~/.arm/cache
```

### Version Tracking
- Use semantic versioning for rule releases
- Document changes in team repository
- Communicate updates to team

## Troubleshooting Team Issues

### Inconsistent Rules
```bash
# Check effective configuration
arm config list

# Verify lock file consistency
cat arm.lock | jq .

# Reinstall if needed
rm arm.lock
arm install
```

### Authentication Issues
```bash
# Verify token access
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Test registry access
arm info team/test-ruleset
```

### Permission Problems
```bash
# Check directory permissions
ls -la .cursor/rules/

# Fix permissions
chmod -R 755 .cursor/rules .amazonq/rules
```

## Best Practices

### Repository Management
- Use semantic versioning for rule releases
- Maintain changelog for rule updates
- Review rule changes through pull requests
- Tag stable releases

### Team Communication
- Announce rule updates in team channels
- Document rule purposes and usage
- Provide migration guides for breaking changes
- Regular team reviews of coding standards

### Security
- Store tokens in environment variables, not config files
- Use organization-level GitHub tokens when possible
- Use HTTPS-only registries in production
- Regularly rotate authentication tokens
- Regularly audit team member access
- Monitor rule repository for unauthorized changes
- Use branch protection for rule repositories
