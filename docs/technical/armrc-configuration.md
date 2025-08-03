# .armrc Configuration System

## Overview

ARM uses `.armrc` configuration files to manage registry sources, authentication, and cache settings. The configuration system supports hierarchy, environment variable substitution, and secure credential handling.

## Configuration Hierarchy

ARM loads configuration in the following order (later sources override earlier ones):

1. **User-level**: `~/.armrc` (global settings)
2. **Project-level**: `./.armrc` (project-specific overrides)
3. **Environment variables**: Future implementation
4. **Command-line flags**: Future implementation

## File Format

Configuration files use INI format:

```ini
[sources]
default = https://registry.armjs.org/
company = https://internal.company.local/

[sources.company]
authToken = ${COMPANY_REGISTRY_TOKEN}
timeout = 30s

[cache]
location = ~/.arm/cache
maxSize = 1GB
```

## Configuration Sections

### Sources Section

Defines registry sources where rulesets can be downloaded:

```ini
[sources]
default = https://registry.armjs.org/
company = https://internal.company.local/
local = file:///path/to/local/registry
```

### Source-Specific Configuration

Each source can have additional configuration based on registry type:

**Generic HTTP Registry:**
```ini
[sources.company]
type = generic
authToken = secret_token_here
timeout = 30s
```

**GitLab Package Registry (Project-level):**
```ini
[sources.gitlab-project]
type = gitlab
projectID = 12345
authToken = ${GITLAB_TOKEN}
```

**GitLab Package Registry (Group-level):**
```ini
[sources.gitlab-group]
type = gitlab
groupID = 67890
authToken = ${GITLAB_TOKEN}
```

**AWS S3 Registry:**
```ini
[sources.s3]
type = s3
bucket = my-arm-registry
region = us-east-1
prefix = rulesets
authToken = ${AWS_ACCESS_TOKEN}
```

**Local Filesystem Registry:**
```ini
[sources.local]
type = filesystem
path = /path/to/local/rulesets
# Structure: {path}/{package}/{version}/{package}-{version}.tar.gz
```

**Supported fields**:
- `type` - Registry type: `generic`, `gitlab`, `s3`, `filesystem`
- `authToken` - Authentication token for the registry
- `timeout` - Request timeout (e.g., "30s", "1m")
- `projectID` - GitLab project ID (numeric)
- `groupID` - GitLab group ID (numeric)
- `bucket` - S3 bucket name
- `region` - AWS region for S3
- `prefix` - S3 key prefix for organizing rulesets
- `path` - Local filesystem path for filesystem registry

### Cache Section

Controls local cache behavior:

```ini
[cache]
location = ~/.arm/cache
maxSize = 1GB
```

**Supported fields**:
- `location` - Cache directory path
- `maxSize` - Maximum cache size (future implementation)

## Environment Variable Substitution

ARM supports environment variable substitution in configuration values:

```ini
[sources.company]
authToken = $COMPANY_TOKEN          # $VAR syntax
authToken = ${COMPANY_TOKEN}        # ${VAR} syntax
```

**Supported patterns**:
- `$VAR` - Simple variable reference
- `${VAR}` - Braced variable reference

**Note**: Default values (`${VAR:-default}`) are not currently supported.

## Configuration Commands

### List Configuration

Display all current configuration:

```bash
arm config list
```

Output shows merged configuration with masked auth tokens:
```
[sources]
default = https://registry.armjs.org/
company = https://internal.company.local/

[sources.company]
authToken = secr****123
timeout = 30s
```

### Get Configuration Value

Retrieve specific configuration values:

```bash
arm config get sources.default
# Output: https://registry.armjs.org/

arm config get sources.company.authToken
# Output: secr****123 (masked for security)
```

### Set Configuration Value

Update configuration values:

```bash
arm config set sources.company https://new-registry.com/
arm config set sources.company.authToken new_token_here
arm config set cache.location ~/.custom/cache
```

**Behavior**:
- Creates `.armrc` in current directory if it doesn't exist
- Falls back to `~/.armrc` if no project config exists
- Preserves existing configuration structure

## Security Features

### Token Masking

Authentication tokens are automatically masked in output:
- Tokens ≤8 characters: fully masked (`********`)
- Tokens >8 characters: show first 4 and last 4 characters (`abcd****xyz9`)

### Environment Variables

Sensitive values should use environment variables:

```ini
[sources.company]
authToken = ${COMPANY_REGISTRY_TOKEN}
```

This keeps secrets out of configuration files and supports different environments.

## Integration with Existing Systems

### Registry Resolution

When installing scoped packages (`company@package-name`), ARM:

1. Looks for a source named `company` in configuration
2. Uses that source's URL and authentication
3. Falls back to `default` source if no match found

### Backward Compatibility

ARM maintains backward compatibility:
- Works without any `.armrc` configuration
- Uses built-in default registry when no config exists
- Existing `rules.json` and `rules.lock` files unchanged

## Best Practices

### Project Configuration

Create project-level `.armrc` for team settings:

```ini
[sources]
company = https://internal.company.local/

[sources.company]
authToken = ${COMPANY_REGISTRY_TOKEN}
```

### User Configuration

Use user-level `~/.armrc` for personal settings:

```ini
[sources]
default = https://registry.armjs.org/
personal = https://github.com/username/arm-registry

[cache]
location = ~/Development/.arm-cache
```

### Environment Variables

Set environment variables in CI/CD:

```bash
export COMPANY_REGISTRY_TOKEN="your_token_here"
# GitLab registry - supports version discovery
arm install company@security-rules
# Generic HTTP/S3 - requires exact version
arm install company@security-rules@1.2.0
```

## Troubleshooting

### Configuration Not Loading

1. Check file permissions on `.armrc` files
2. Verify INI syntax is correct
3. Use `arm config list` to see merged configuration

### Authentication Issues

1. Verify environment variables are set
2. Check token masking isn't hiding issues
3. Test with `arm config get sources.company.authToken`

### Environment Variable Substitution

1. Ensure variables are exported in shell
2. Check variable names match exactly (case-sensitive)
3. Use `${VAR}` syntax for complex variable names
