# Team Setup Guide

Guide for team leads and platform engineers to distribute AI coding rulesets across development teams.

## Team Workflow Overview

1. **Team Lead**: Creates and maintains rulesets in a shared registry
2. **Team Members**: Install and sync rulesets from the shared registry
3. **Platform Team**: Manages enterprise-wide registries and policies

## Setting Up Team Registries

### Option 1: Git Repository (Recommended)

Create a shared Git repository for your team's rulesets:

```bash
# Team lead sets up the registry
arm config add registry team-registry https://github.com/jomadu/ai-rules-manager-test-git-registry --type=git --authToken=$GITHUB_TOKEN

# Team members add the same registry
arm config add registry team-registry https://github.com/jomadu/ai-rules-manager-test-git-registry --type=git
```

### Option 2: S3 Bucket (Enterprise)

For larger teams with AWS infrastructure:

```bash
# Platform team creates shared S3 registry
arm config add registry corp-registry team-rules-bucket --type=s3 --region=us-east-1 --profile=company-aws

# Team members use the same bucket
arm config add registry corp-registry team-rules-bucket --type=s3 --region=us-east-1
```

## Standardizing Team Configuration

### Shared Configuration Template

Create a template `.armrc` for your team:

```ini
# Team ARM Configuration Template

[registries]
team-registry = https://github.com/jomadu/ai-rules-manager-test-git-registry
corp-registry = team-rules-bucket

[registries.team-registry]
type = git
concurrency = 2
rateLimit = 15/minute

[registries.corp-registry]
type = s3
region = us-east-1
concurrency = 5
rateLimit = 50/minute
```

### Shared arm.json Template

```json
{
  "engines": {
    "arm": "^1.2.3"
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
    "team-registry": {
      "rules": {
        "version": "^2.0.0",
        "patterns": ["*.md"]
      }
    }
  }
}
```

## Team Member Onboarding

### Quick Setup Script

Create an onboarding script for new team members:

```bash
#!/bin/bash
# team-setup.sh

echo "Setting up Team ARM configuration..."

# Add team registries
arm config add registry team-registry https://github.com/jomadu/ai-rules-manager-test-git-registry --type=git
arm config add registry corp-registry team-rules-bucket --type=s3 --region=us-east-1

# Set default registry
arm config set registries.default team-registry

# Add channels
arm config add channel cursor --directories .cursor/rules
arm config add channel q --directories .amazonq/rules

# Install team rulesets
arm install rules --patterns "*.md"

echo "Team setup complete!"
```

### Verification

Team members can verify their setup:

```bash
# Check configuration
arm config list

# Verify installations
arm list

# Check for updates
arm outdated
```

## Managing Team Rulesets

### Publishing Updates

Team leads publish new versions:

```bash
# After updating rulesets in Git repository
git tag v2.1.0
git push origin v2.1.0
```

### Team Sync

Team members stay updated:

```bash
# Check for updates
arm outdated

# Update all rulesets
arm update

# Update specific ruleset
arm update rules
```

## Enterprise Deployment

### CI/CD Integration

Add ARM to your CI/CD pipeline:

```yaml
# .github/workflows/setup-arm.yml
name: Setup ARM
on: [push, pull_request]

jobs:
  setup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install ARM
        run: curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash

      - name: Configure ARM
        run: |
          arm config add registry corp-registry team-rules-bucket --type=s3 --region=us-east-1
          arm config add channel cursor --directories .cursor/rules

      - name: Install rulesets
        run: arm install
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

### Docker Integration

Include ARM in development containers:

```dockerfile
# Dockerfile
FROM node:18

# Install ARM
RUN curl -sSL https://raw.githubusercontent.com/jomadu/ai-rules-manager/main/scripts/install.sh | bash

# Copy team configuration
COPY .armrc /root/.armrc
COPY arm.json /workspace/arm.json

# Install team rulesets
WORKDIR /workspace
RUN arm install

# Continue with your app setup...
```

## Governance and Policies

### Version Pinning Strategy

Choose a versioning strategy for your team:

```bash
# Conservative: Pin exact versions
arm install rules@2.0.0

# Flexible: Use semver ranges
arm install rules@^2.0.0

# Latest: Always use newest (not recommended for production)
arm install rules@latest
```

### Registry Access Control

Control who can publish to team registries:

- **Git**: Use repository permissions and branch protection
- **S3**: Use IAM policies to control bucket access
- **GitLab**: Use project/group permissions

### Audit and Compliance

Track ruleset usage across your team:

```bash
# Generate team usage report
arm list --json > team-rulesets-$(date +%Y%m%d).json

# Check for outdated rulesets
arm outdated --json
```

## Troubleshooting Team Issues

### Different Versions Across Team

```bash
# Team member has wrong version
$ arm list
cursor:
  team-registry:
    - rules@1.9.0  # Should be 2.0.0
```

**Solution**: Update to latest:
```bash
arm update rules
```

### Registry Access Issues

```bash
Error [AUTH]: Access denied to registry 'team-registry'
Details: HTTP 403 - insufficient permissions
```

**Solutions**:
- **Git**: Check if team member has repository access
- **S3**: Verify AWS credentials and IAM permissions
- **GitLab**: Check project/group membership

### Configuration Drift

Team members have different configurations:

**Solution**: Use shared configuration templates and setup scripts.

### Lock File Conflicts

```bash
Error: arm.json changes conflict with arm.lock
```

**Solution**: Re-run install to regenerate lock file:
```bash
arm install
```

## Best Practices

### For Team Leads
- Use semantic versioning for ruleset releases
- Document changes in release notes
- Test rulesets before publishing
- Communicate updates to the team

### For Team Members
- Run `arm outdated` regularly
- Don't modify ARM-managed files manually
- Report issues with rulesets to team leads
- Keep local configuration minimal

### For Platform Teams
- Standardize registry types across the organization
- Provide shared configuration templates
- Monitor registry usage and performance
- Implement backup strategies for critical rulesets
