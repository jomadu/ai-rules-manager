# Registry Guide

ARM supports multiple registry types for hosting and distributing rulesets. You must configure at least one registry before installing rulesets.

## Registry Types Comparison

| Registry Type | Version Discovery | Authentication | Use Case |
|---------------|-------------------|----------------|----------|
| GitLab Package Registry | ✅ | Token/OAuth | Enterprise teams |
| AWS S3 | ✅ | AWS credentials | Cloud-native orgs |
| Git Repository | ✅ | SSH/HTTPS | Open source projects |
| HTTP Server | ❌ | Basic/Bearer | Simple file hosting |
| Local Filesystem | ✅ | None | Development/testing |

## GitLab Package Registry

Best for enterprise teams with existing GitLab infrastructure.

### Setup

```bash
# Configure registry
arm config set sources.company https://gitlab.company.com
arm config set sources.company.type gitlab
arm config set sources.company.projectID 12345
arm config set sources.company.authToken $GITLAB_TOKEN
```

### Usage

```bash
# Install from GitLab registry
arm install company@typescript-rules
arm install company@security-rules@2.1.0
```

### Authentication

- **Personal Access Token**: Create with `read_api` scope
- **CI/CD Token**: Use `$CI_JOB_TOKEN` in pipelines
- **OAuth**: For interactive authentication

## AWS S3

Ideal for cloud-native organizations using AWS.

### Setup

```bash
# Configure S3 registry
arm config set sources.s3 s3://my-rules-bucket/packages/
arm config set sources.s3.type s3
arm config set sources.s3.region us-east-1
```

### Usage

```bash
# Install from S3
arm install s3@typescript-rules
```

### Authentication

Uses standard AWS credential chain:
- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- AWS credentials file (`~/.aws/credentials`)
- IAM roles (EC2/ECS)

## Git Repository

Perfect for open source projects and direct repository access.

### Setup

```bash
# No configuration needed for public repos
# For private repos, ensure SSH/HTTPS access is configured

# Configure authentication for higher rate limits
arm config set git.github.token $GITHUB_TOKEN
arm config set git.gitlab.token $GITLAB_TOKEN
```

### Usage

```bash
# Install from git repository
arm install github.com/user/awesome-rules@main:rules/*.md

# Specific patterns and branch
arm install gitlab.com/team/standards@v2.0:typescript/*.md,docs/*.txt

# SSH access
arm install git@github.com:company/private-rules@main:rules/*
```

### Patterns

- `rules/*.md` - All markdown files in rules directory
- `**/*.md` - All markdown files recursively
- `typescript/*.md,react/*.md` - Multiple patterns

## HTTP Server

Simple file server hosting for basic needs.

### Setup

```bash
# Configure HTTP registry
arm config set sources.files https://files.company.com/rules/
```

### Usage

```bash
# Exact version required for HTTP registries
arm install files@typescript-rules@1.0.0
```

### Requirements

- Must specify exact versions
- Server must support directory listing or provide manifest

## Local Filesystem

For development and testing.

### Setup

```bash
# Configure local registry
arm config set sources.local file:///path/to/local/registry/
```

### Usage

```bash
# Install from local filesystem
arm install local@typescript-rules
```

### Structure

```
/path/to/local/registry/
├── typescript-rules/
│   ├── 1.0.0/
│   │   ├── rule1.md
│   │   └── rule2.md
│   └── 1.1.0/
│       ├── rule1.md
│       └── rule3.md
```

## Authentication Examples

### GitLab with Personal Access Token

```bash
export GITLAB_TOKEN=glpat-xxxxxxxxxxxxxxxxxxxx
arm config set sources.company.authToken $GITLAB_TOKEN
```

### AWS S3 with Environment Variables

```bash
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_DEFAULT_REGION=us-east-1
```

### Git with SSH Key

```bash
# Ensure SSH key is configured
ssh-add ~/.ssh/id_rsa
arm install git@gitlab.company.com:team/rules@main:rules/*
```

## Registry Priority

When multiple registries have the same ruleset:

1. Explicit registry specification (`company@ruleset`)
2. Registry order in configuration
3. Default registry (if configured)

```bash
# Explicit registry (highest priority)
arm install company@typescript-rules

# Uses first matching registry in config
arm install typescript-rules
```
