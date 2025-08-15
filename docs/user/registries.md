# Registry Guide

Working with different registry types in ARM.

## Registry Types Overview

| Type | Use Case | Authentication | Versioning | Patterns |
|------|----------|----------------|------------|----------|
| Git | GitHub/GitLab repos | Token (private only) | Tags/branches | Yes |
| Git-Local | Local Git repos | Filesystem | Git tags/branches | Yes |
| S3 | AWS S3 buckets | IAM | Directory structure | No |
| HTTPS | Custom APIs | Token | API-defined | No |
| GitLab | GitLab Package Registry | Token | Package versions | No |
| Local | Local directories | Filesystem | Directory structure | No |

## Git Registries

### GitHub
```bash
# Public repository
arm config add registry public https://github.com/org/rules --type=git

# Private repository
arm config add registry private https://github.com/org/private-rules --type=git --authToken=$GITHUB_TOKEN
```

### GitLab
```bash
arm config add registry gitlab https://gitlab.com/org/rules --type=git --authToken=$GITLAB_TOKEN --apiType=gitlab
```

### Version Formats
```bash
arm install rules@latest        # Latest tag
arm install rules@main          # Branch
arm install rules@v1.2.3        # Specific tag
arm install rules@1.2.3         # Tag without 'v' prefix
arm install rules@^1.0.0        # Semver constraint
```

### Pattern Usage
```bash
# Install specific files
arm install rules --patterns "standards/*.md,guidelines/*.md"

# Exclude files
arm install rules --patterns "**/*.md,!**/drafts/**"
```

## S3 Registries

### Setup
```bash
# Basic S3 registry
arm config add registry s3-rules my-bucket --type=s3 --region=us-east-1

# With custom profile and prefix
arm config add registry s3-team team-bucket --type=s3 --region=us-west-2 --profile=team --prefix=/rules/
```

### Bucket Structure
```
bucket/
├── prefix/                    # Optional
│   └── ruleset-name/
│       ├── v1.0.0/
│       │   └── ruleset.tar.gz
│       └── v1.1.0/
│           └── ruleset.tar.gz
```

Note: "latest" version is resolved by listing all version directories and selecting the most recent one. Only `ruleset.tar.gz` files are used.

### Authentication
```bash
# AWS CLI configuration
aws configure --profile team

# Environment variables
export AWS_PROFILE=team
export AWS_REGION=us-east-1
```

## HTTPS Registries

### Setup
```bash
arm config add registry api https://registry.example.com --type=https --authToken=$REGISTRY_TOKEN
```

### API Endpoints
- `GET /manifest.json` - Get manifest with rulesets and versions
- `GET /{name}/{version}/ruleset.tar.gz` - Download ruleset tarball

### Manifest Format
```json
{
  "rulesets": {
    "coding-standards": ["v1.0.0", "v1.1.0"],
    "security-rules": ["v2.0.0"]
  }
}
```

Note: Patterns are ignored for HTTPS registries as they use pre-packaged tar.gz files.

## Local Registries

### Local Directory
```bash
arm config add registry local-rules /path/to/rules --type=local
```

### Directory Structure
```
/path/to/rules/
├── ruleset-1/
│   ├── v1.0.0/
│   │   └── ruleset.tar.gz
│   └── v1.1.0/
│       └── ruleset.tar.gz
└── ruleset-2/
    └── v2.0.0/
        └── ruleset.tar.gz
```

Note: Local registries expect pre-packaged `ruleset.tar.gz` files in each version directory. "Latest" version is resolved by selecting the most recent version directory. Patterns are ignored.

### Git-Local
```bash
arm config add registry local-dev /path/to/git/repo --type=git-local
```

## GitLab Registries

### Setup
```bash
arm config add registry gitlab-rules https://gitlab.example.com/projects/123 --type=gitlab --authToken=$GITLAB_TOKEN
```

### Implementation
- Uses GitLab Package Registry API
- Supports generic packages
- Requires project ID in URL
- Patterns are ignored (uses pre-packaged tar.gz files)

## Registry Management

### List Registries
```bash
arm config list | grep registries
```

### Test Registry
```bash
arm info registry/test-ruleset
```

### Remove Registry
```bash
arm config remove registry old-registry
```

## Best Practices

### Security
- Use environment variables for tokens
- Rotate tokens regularly
- Use least-privilege IAM policies for S3

### Organization
- Use descriptive registry names
- Separate dev/staging/prod registries
- Document registry purposes

### Performance
- Use local registries for development
- Configure appropriate cache settings
- Monitor rate limits for API-based registries
