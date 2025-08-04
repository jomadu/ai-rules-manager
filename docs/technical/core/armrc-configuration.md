# Configuration System

Technical details of ARM's .armrc configuration system.

## Hierarchy

1. `~/.armrc` (user-level)
2. `./.armrc` (project-level, overrides user)

## Format

```ini
[sources]
default = https://registry.armjs.org/
company = https://internal.company.local/

[sources.company]
type = gitlab
projectID = 12345
authToken = ${GITLAB_TOKEN}
```

## Registry Types

**GitLab**:
```ini
[sources.gitlab]
type = gitlab
projectID = 12345
authToken = ${GITLAB_TOKEN}
```

**S3**:
```ini
[sources.s3]
type = s3
bucket = my-registry
region = us-east-1
accessKey = ${AWS_ACCESS_KEY_ID}
secretKey = ${AWS_SECRET_ACCESS_KEY}
```

**HTTP**:
```ini
[sources.http]
type = http
url = https://packages.company.com/
authToken = ${HTTP_TOKEN}
```

**Filesystem**:
```ini
[sources.local]
type = filesystem
path = ./local-registry
```

## Environment Variables

Supports `$VAR` and `${VAR}` substitution.

## Commands

```bash
arm config list                    # Show all config
arm config get sources.default     # Get specific value
arm config set sources.company URL # Set value
```

## Security

- Tokens masked in output (`abcd****xyz9`)
- Environment variables for secrets
- No credentials in config files

## Registry Resolution

`company@package` → looks for `sources.company` → falls back to `sources.default`
