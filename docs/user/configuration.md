# Configuration Guide

ARM uses configuration files to manage registries, authentication, and performance settings.

## Configuration Files

### .armrc (Registry Configuration)

ARM looks for `.armrc` files in this order:
1. Current directory
2. Parent directories (up to project root)
3. Home directory (`~/.armrc`)

### rules.json (Project Manifest)

Defines project dependencies and target directories:

```json
{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "typescript-rules": "^1.0.0",
    "company@security-rules": "^2.1.0"
  }
}
```

### rules.lock (Lock File)

Auto-generated file with exact versions and sources. Don't edit manually.

## Registry Configuration

### Basic Registry Setup

```ini
[sources]
default = https://registry.armjs.org/
company = https://internal.company.local/
```

### GitLab Package Registry

```ini
[sources.gitlab]
type = gitlab
url = https://gitlab.company.com
projectID = 12345
authToken = $GITLAB_TOKEN
```

### AWS S3 Registry

```ini
[sources.s3]
type = s3
bucket = my-arm-registry
region = us-east-1
prefix = packages/
accessKey = $AWS_ACCESS_KEY_ID
secretKey = $AWS_SECRET_ACCESS_KEY
```

### HTTP Registry

```ini
[sources.http]
type = http
url = https://packages.company.com/arm/
authToken = $HTTP_TOKEN
```

### Filesystem Registry (Development)

```ini
[sources.local]
type = filesystem
path = /path/to/local/registry
```

### Git Repository Registry

```ini
[sources.awesome-rules]
type = git
url = https://github.com/PatrickF1/awesome-cursorrules
authToken = $GITHUB_TOKEN

[sources.company-rules]
type = git
url = https://github.com/company/internal-rules
authToken = $COMPANY_GITHUB_TOKEN
```

**Note**: Public repositories don't require authentication tokens.

## Authentication

### Environment Variables

ARM supports these environment variables:

- `GITLAB_TOKEN` - GitLab personal access token
- `AWS_ACCESS_KEY_ID` - AWS access key
- `AWS_SECRET_ACCESS_KEY` - AWS secret key
- `HTTP_TOKEN` - Generic HTTP authentication token

### Token Storage

Store tokens in environment variables, not in `.armrc` files:

```bash
export GITLAB_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"
export COMPANY_GITHUB_TOKEN="ghp_yyyyyyyyyyyyyyyyyyyy"
```

## Performance Settings

### Global Performance

```ini
[performance]
defaultConcurrency = 3
```

### Registry-Specific Performance

```ini
[performance.gitlab]
concurrency = 2

[performance.s3]
concurrency = 8
```

## Target Configuration

### Default Targets

If no `rules.json` exists, ARM uses these default targets:
- `.cursorrules` (Cursor IDE)
- `.amazonq/rules` (Amazon Q Developer)

### Custom Targets

```json
{
  "targets": [
    ".cursorrules",
    ".amazonq/rules",
    "custom-ai-tool/rules"
  ]
}
```

## Version Constraints

ARM supports semantic versioning constraints:

```json
{
  "dependencies": {
    "exact-version": "1.0.0",
    "caret-range": "^1.0.0",
    "tilde-range": "~1.0.0",
    "latest": "latest"
  }
}
```

### Constraint Types

- `1.0.0` - Exact version
- `^1.0.0` - Compatible version (>=1.0.0 <2.0.0)
- `~1.0.0` - Patch-level changes (>=1.0.0 <1.1.0)
- `latest` - Latest available version

## Configuration Commands

### List All Configuration

```bash
arm config list
```

### Get Specific Value

```bash
arm config get sources.default
arm config get performance.defaultConcurrency
```

### Set Configuration

```bash
arm config set sources.company https://internal.company.local/
arm config set sources.company.authToken $COMPANY_TOKEN
arm config set performance.defaultConcurrency 5
```

## Configuration Examples

### Enterprise Setup

```ini
[sources]
default = https://registry.armjs.org/
internal = https://gitlab.company.com/api/v4/projects/123/packages/generic/arm-rules
public = https://packages.company.com/arm/

[sources.internal]
type = gitlab
projectID = 123
authToken = $GITLAB_TOKEN

[sources.public]
type = http
authToken = $COMPANY_TOKEN

[performance]
defaultConcurrency = 4

[performance.internal]
concurrency = 2
```

### Development Setup

```ini
[sources]
default = https://registry.armjs.org/
local = /home/dev/arm-registry

[sources.local]
type = filesystem

[performance]
defaultConcurrency = 1
```

## Troubleshooting Configuration

### Check Configuration

```bash
arm config list
```

### Validate Registry Access

```bash
arm install --dry-run
```

### Debug Authentication

Set debug environment variable:

```bash
export ARM_DEBUG=1
arm install typescript-rules
```
