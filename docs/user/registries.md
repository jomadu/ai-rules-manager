# Registry Guide

Complete guide to setting up different types of registries for storing and sharing rulesets.

## Registry Types

ARM supports five registry types:
- **Git** - GitHub, GitLab, and other Git repositories
- **S3** - AWS S3 buckets
- **GitLab** - GitLab Package Registry
- **HTTPS** - Generic HTTP servers with manifest.json
- **Local** - Local file system directories

## Git Registries

### Setup

```bash
# Public repository
arm config add registry test-registry https://github.com/jomadu/ai-rules-manager-test-git-registry --type=git

# Private repository with token
arm config add registry private-registry https://github.com/user/private-rules --type=git --authToken=$GITHUB_TOKEN

# Enable API mode for faster access
arm config add registry api-registry https://github.com/user/rules --type=git --authToken=$GITHUB_TOKEN --apiType=github --apiVersion=2022-11-28
```

### Installing from Git

Git registries require patterns to select files:

```bash
# Install with patterns
arm install rules --patterns "*.md"

# Multiple patterns
arm install rules --patterns "rules/*.md,cursor/*.md"
```

### Version Targeting

```bash
# Latest commit on default branch
arm install rules@latest

# Specific branch
arm install rules@main

# Git tag (supports v1.0.0 and 1.0.0 formats)
arm install rules@^1.0.0

# Specific commit
arm install rules@abc123def
```

## S3 Registries

### Setup

```bash
# Basic S3 registry
arm config add registry koopa-troopa koopa-castle-rules --type=s3 --region=us-east-1

# With custom AWS profile
arm config add registry lakitu-cloud lakitu-rules-bucket --type=s3 --region=us-west-2 --profile=mario-aws

# With prefix
arm config add registry goomba-storage goomba-bucket --type=s3 --region=eu-west-1 --prefix=/rulesets/path
```

### S3 Directory Structure

Your S3 bucket should be organized like this:
```
koopa-castle-rules/
├── power-up-rules/
│   ├── 1.0.0/
│   │   └── ruleset.tar.gz
│   └── 1.1.0/
│       └── ruleset.tar.gz
└── fire-flower-security/
    └── 2.0.0/
        └── ruleset.tar.gz
```

### Installing from S3

```bash
# Install latest version
arm install power-up-rules

# Install specific version
arm install power-up-rules@1.0.0
```

## GitLab Package Registry

### Setup

```bash
# Project-level registry
arm config add registry toad-house https://gitlab.mushroom-kingdom.example/projects/456 --type=gitlab --authToken=$GITLAB_TOKEN

# Group-level registry
arm config add registry yoshi-group https://gitlab.mushroom-kingdom.example/groups/789 --type=gitlab --authToken=$GITLAB_TOKEN
```

### Installing from GitLab

```bash
# Install from GitLab package registry
arm install star-power-performance@~2.1.0
```

## HTTPS Registries

### Setup

```bash
# Generic HTTP registry
arm config add registry warp-zone https://registry.warp-zone.example/rulesets --type=https

# With authentication
arm config add registry pipe-world https://registry.pipe-world.example --type=https --authToken=$HTTP_TOKEN
```

### Server Requirements

Your HTTPS server needs a `manifest.json` at the root:

```json
{
  "rulesets": {
    "power-up-rules": ["1.0.0", "1.1.0"],
    "fire-flower-security": ["2.0.0"]
  }
}
```

## Local Registries

### Setup

```bash
# Absolute path
arm config add registry dev-rules /home/mario/my-rulesets --type=local

# Relative path
arm config add registry project-rules ./local-rulesets --type=local
```

### Local Directory Structure

```
/home/mario/my-rulesets/
├── power-up-rules/
│   ├── 1.0.0/
│   │   └── ruleset.tar.gz
│   └── 1.1.0/
│       └── ruleset.tar.gz
└── fire-flower-security/
    └── 2.0.0/
        └── ruleset.tar.gz
```

## Registry Performance Tuning

### Concurrency Settings

```bash
# Increase parallel operations for fast registries
arm config set registries.test-registry.concurrency 5

# Reduce for rate-limited APIs
arm config set registries.bowser-castle.concurrency 1
```

### Rate Limiting

```bash
# Adjust rate limits per registry
arm config set registries.koopa-troopa.rateLimit 50/minute
arm config set registries.toad-house.rateLimit 100/hour

# Set type defaults
arm config set git.rateLimit 20/minute
arm config set s3.rateLimit 200/hour
```

## Registry Management

### List Configured Registries

```bash
arm config list
```

### Remove Registry

```bash
arm config remove registry old-registry
```

### Search Across Registries

```bash
# Search all registries
arm search "power-up"

# Search specific registries
arm search "security" --registries test-registry,private-registry
```

## Troubleshooting

### Authentication Failed
```bash
Error [AUTH]: Access denied to registry 'private-registry'
Details: HTTP 403 - invalid or expired token
```
**Solution**: Check your authentication token:
```bash
# Verify token is set
echo $GITHUB_TOKEN

# Update token
arm config set registries.private-registry.authToken $NEW_TOKEN
```

### Registry Not Found
```bash
Error [NETWORK]: Failed to connect to registry 'https://github.com/fake-repo'
Details: DNS lookup failed
```
**Solution**: Verify the registry URL is correct.

### S3 Region Mismatch
```bash
Error [REGISTRY]: S3 bucket 'koopa-castle-rules' not found in region 'us-east-1'
```
**Solution**: Check the bucket region:
```bash
arm config set registries.koopa-troopa.region us-west-2
```

### Rate Limited
```bash
Error [NETWORK]: Rate limit exceeded for registry 'private-registry'
Details: 60 requests per hour limit reached
```
**Solution**: Wait for rate limit reset or reduce rate limit:
```bash
arm config set registries.private-registry.rateLimit 50/hour
```
