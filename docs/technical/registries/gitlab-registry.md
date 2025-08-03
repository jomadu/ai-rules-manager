# GitLab Registry Guide

## Overview

ARM's GitLab registry integrates with GitLab's native Package Registry, providing rich metadata and version discovery through GitLab's API. It supports both project-level and group-level package registries.

## Features

- **Native API Integration**: Uses GitLab's package API for metadata
- **Version Discovery**: Automatic version listing from GitLab
- **Rich Metadata**: Package information, sizes, publication dates
- **Project & Group Support**: Both project and group-level registries
- **Authentication**: GitLab personal access tokens or CI tokens

## Configuration Examples

### Project-Level Registry
```ini
[sources.gitlab-project]
type = gitlab
url = https://gitlab.example.com
projectID = 12345
authToken = ${GITLAB_TOKEN}
```

### Group-Level Registry
```ini
[sources.gitlab-group]
type = gitlab
url = https://gitlab.example.com
groupID = 67890
authToken = ${GITLAB_TOKEN}
```

### GitLab.com Registry
```ini
[sources.gitlab-com]
type = gitlab
url = https://gitlab.com
projectID = 12345
authToken = ${GITLAB_TOKEN}
```

## Usage Examples

### Install Commands
```bash
# Install latest version (version discovery supported)
arm install typescript-rules

# Install specific version
arm install typescript-rules@1.0.0

# Install scoped package
arm install company@security-rules

# List packages with rich metadata
arm list
```

### Configuration Usage
```bash
# Configure GitLab registry
arm config set sources.company https://gitlab.company.com
arm config set sources.company.projectID 12345
arm config set sources.company.authToken ${GITLAB_TOKEN}

# Install from configured source
arm install company@typescript-rules
```

## API Integration

### Package Discovery
ARM uses GitLab's packages API:
```
GET /api/v4/projects/{id}/packages?package_name={name}
GET /api/v4/groups/{id}/packages?package_name={name}
```

### Metadata Extraction
GitLab API provides rich package information:
- Package ID and name
- Version information
- File sizes and checksums
- Publication and update dates
- Package type and status

### Download URLs
Packages are downloaded via GitLab's generic package API:
```
GET /api/v4/projects/{id}/packages/generic/{name}/{version}/{file}
GET /api/v4/groups/{id}/packages/generic/{name}/{version}/{file}
```

## Publishing Workflow

### Using GitLab CI/CD
```yaml
# .gitlab-ci.yml
publish:
  stage: deploy
  script:
    - npm run build
    - tar -czf ${PACKAGE_NAME}-${CI_COMMIT_TAG}.tar.gz dist/
    - |
      curl --header "JOB-TOKEN: $CI_JOB_TOKEN" \
           --upload-file ${PACKAGE_NAME}-${CI_COMMIT_TAG}.tar.gz \
           "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/packages/generic/${PACKAGE_NAME}/${CI_COMMIT_TAG}/${PACKAGE_NAME}-${CI_COMMIT_TAG}.tar.gz"
  only:
    - tags
```

### Manual Upload
```bash
# Upload package to GitLab registry
curl --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
     --upload-file typescript-rules-1.0.0.tar.gz \
     "https://gitlab.example.com/api/v4/projects/12345/packages/generic/typescript-rules/1.0.0/typescript-rules-1.0.0.tar.gz"
```

### Group-Level Publishing
```bash
# Upload to group registry
curl --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
     --upload-file security-rules-1.0.0.tar.gz \
     "https://gitlab.example.com/api/v4/groups/67890/packages/generic/security-rules/1.0.0/security-rules-1.0.0.tar.gz"
```

## Authentication

### Personal Access Tokens
Create token with `api` scope:
1. Go to GitLab → User Settings → Access Tokens
2. Create token with `api` scope
3. Use in ARM configuration

### CI/CD Tokens
Use built-in CI job tokens:
```yaml
# In GitLab CI/CD
script:
  - export GITLAB_TOKEN=$CI_JOB_TOKEN
  - arm install company@typescript-rules
```

### Project vs Group Tokens
- **Project tokens**: Access to specific project packages
- **Group tokens**: Access to all packages in group
- **Personal tokens**: Access based on user permissions

## Best Practices

### Token Management
- Use environment variables for tokens
- Rotate tokens regularly
- Use minimal required scopes
- Consider CI/CD tokens for automated workflows

### Registry Organization
- Use group registries for shared packages
- Use project registries for project-specific packages
- Implement consistent naming conventions
- Tag releases for version management

### CI/CD Integration
- Automate publishing on git tags
- Use semantic versioning
- Include package validation steps
- Set up automated testing

### Security
- Enable package registry security scanning
- Use protected branches for releases
- Implement approval workflows
- Monitor package access logs

## Limitations

### GitLab Version Requirements
- Requires GitLab 13.5+ for generic packages
- Some features require newer GitLab versions
- API rate limits may apply

### Package Size Limits
- Default 5GB per package file
- Configurable by GitLab administrators
- Consider splitting large packages

### API Rate Limits
- GitLab applies API rate limits
- May affect bulk operations
- Consider caching for frequent access

## Troubleshooting

### Common Issues
1. **401 Unauthorized**: Check token validity and scopes
2. **403 Forbidden**: Verify project/group permissions
3. **404 Not Found**: Confirm project/group ID and package existence
4. **Package Not Found**: Check package name and version

### Debugging Steps
```bash
# Test GitLab API access
curl --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
     "https://gitlab.example.com/api/v4/projects/12345"

# List packages in project
curl --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
     "https://gitlab.example.com/api/v4/projects/12345/packages"

# Test package download
curl --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
     "https://gitlab.example.com/api/v4/projects/12345/packages/generic/typescript-rules/1.0.0/typescript-rules-1.0.0.tar.gz"
```

### Health Check Validation
ARM's health check verifies:
- GitLab instance accessibility
- Authentication token validity
- Project/group existence and permissions
- API endpoint availability