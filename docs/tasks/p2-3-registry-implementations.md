# P2.3: Registry Implementations

## Overview
Implement specific registry types for GitLab, GitHub, S3, HTTP endpoints, and local file system sources.

## Requirements
- GitLab package registry support
- GitHub package registry support
- Generic HTTP endpoint support
- Local file system registry
- AWS S3 bucket support

## Tasks
- [ ] **GitLab Package Registry**:
  - API endpoint: `/api/v4/projects/:id/packages/generic/:package_name/:package_version/:file_name`
  - Authentication via personal access tokens
  - Handle GitLab-specific metadata format
- [ ] **GitHub Package Registry**:
  - GitHub Packages API integration
  - Support for both public and private repositories
  - GitHub token authentication
- [ ] **Generic HTTP Registry**:
  - Simple HTTP GET for package downloads
  - Configurable URL patterns
  - Basic and bearer token authentication
- [ ] **Local File System Registry**:
  - File:// URL scheme support
  - Local directory structure scanning
  - No authentication required
- [ ] **AWS S3 Registry**:
  - S3 API integration
  - IAM role and access key authentication
  - S3 bucket and prefix configuration

## Acceptance Criteria
- [ ] All registry types implement the Registry interface
- [ ] Authentication works for each registry type
- [ ] Metadata fetching works correctly
- [ ] Download functionality is reliable
- [ ] Error handling is consistent across implementations
- [ ] Configuration is properly validated

## Dependencies
- github.com/aws/aws-sdk-go-v2 (S3 support)
- net/http (standard library)

## Files to Create
- `internal/registry/gitlab.go`
- `internal/registry/github.go`
- `internal/registry/http.go`
- `internal/registry/filesystem.go`
- `internal/registry/s3.go`

## Configuration Examples
```ini
[sources.gitlab]
type = gitlab
url = https://gitlab.company.com
projectId = 123
authToken = ${GITLAB_TOKEN}

[sources.s3]
type = s3
bucket = my-rulesets
region = us-east-1
prefix = rulesets/
```