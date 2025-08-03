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
- [x] **GitLab Package Registry**:
  - API endpoint: `/api/v4/projects/:id/packages/generic/:package_name/:package_version/:file_name`
  - Authentication via personal access tokens
  - Handle GitLab-specific metadata format
  - Enhanced error handling and health checks
- [x] **GitHub Package Registry**:
  - Removed (see ADR-001)
- [x] **Generic HTTP Registry**:
  - Simple HTTP GET for package downloads
  - Configurable URL patterns
  - Basic and bearer token authentication
  - Health check functionality
- [x] **Local File System Registry**:
  - Local directory structure scanning
  - No authentication required
  - Path validation and health checks
- [x] **AWS S3 Registry**:
  - S3 API integration
  - Access key authentication
  - S3 bucket and prefix configuration
  - Enhanced error handling

## Acceptance Criteria
- [x] All registry types implement the Registry interface
- [x] Authentication works for each registry type
- [x] Metadata fetching works correctly
- [x] Download functionality is reliable
- [x] Error handling is consistent across implementations
- [x] Configuration is properly validated
- [x] Health check functionality implemented
- [x] Enhanced metadata structure with additional fields

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
