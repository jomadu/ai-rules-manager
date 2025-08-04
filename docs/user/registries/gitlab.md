# GitLab Registry

Use GitLab package registries for ARM rulesets.

## Setup

```ini
[sources.gitlab]
type = gitlab
url = https://gitlab.example.com
projectID = 12345
authToken = $GITLAB_TOKEN
```

```bash
export GITLAB_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
```

## Authentication

1. GitLab → User Settings → Access Tokens
2. Create token with `api` scope
3. Set environment variable

## Publishing

```bash
# Create package
tar -czf package.tar.gz -C rules/ .

# Upload
curl --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
     --upload-file package.tar.gz \
     "https://gitlab.example.com/api/v4/projects/12345/packages/generic/arm-rules/ruleset-name/1.0.0/package.tar.gz"
```

## Troubleshooting

**Authentication Failed**
```bash
curl --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
     "https://gitlab.example.com/api/v4/user"
```

**Package Not Found**
- Check project ID
- Verify token permissions
- Ensure package exists
