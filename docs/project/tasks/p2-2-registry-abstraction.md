# P2.2: Registry Abstraction

## Overview
Create a registry interface and HTTP client abstraction to support multiple registry types with authentication and metadata fetching.

## Requirements
- Registry interface for different source types
- HTTP registry client with authentication
- Registry metadata fetching
- Error handling and retry logic

## Tasks
- [x] **Define Registry interface**:
  ```go
  type Registry interface {
    GetRuleset(name, version string) (*Ruleset, error)
    ListVersions(name string) ([]string, error)
    Download(name, version string) (io.ReadCloser, error)
    GetMetadata(name string) (*Metadata, error)
  }
  ```
- [x] **HTTP client implementation**:
  - Support for different authentication methods
  - Request/response handling
  - Timeout and retry configuration
  - User-agent and headers
- [x] **Authentication handling**:
  - Bearer tokens
  - Basic authentication
  - Custom headers
  - Token refresh logic
- [x] **Metadata structures**:
  ```go
  type Metadata struct {
    Name        string
    Description string
    Versions    []Version
    Repository  string
  }
  ```
- [x] **Error handling**:
  - Network errors
  - Authentication failures
  - Rate limiting
  - Registry unavailable

## Acceptance Criteria
- [x] Registry interface supports all required operations
- [x] HTTP client handles authentication correctly
- [x] Metadata fetching works for all registry types
- [x] Proper error handling and user feedback
- [x] Request timeouts prevent hanging
- [x] Registry manager with caching and registry selection

## Dependencies
- net/http (standard library)
- context (standard library)

## Files Created
- `internal/registry/interface.go` - Registry interface definition
- `internal/registry/http.go` - Generic HTTP registry implementation
- `internal/registry/gitlab.go` - GitLab package registry implementation
- `internal/registry/s3.go` - AWS S3 registry implementation
- `internal/registry/auth.go` - Authentication providers
- `internal/registry/manager.go` - Registry factory and management
- `internal/registry/types.go` - Registry type constants

## Notes
- Plan for caching of metadata responses
- Consider rate limiting for API calls
- Support for registry health checks
- Design for easy testing with mocks
- **GitHub registry removed**: See ADR-001 for rationale - GitHub lacks suitable generic package registry API
