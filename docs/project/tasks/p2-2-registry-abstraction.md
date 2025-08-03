# P2.2: Registry Abstraction

## Overview
Create a registry interface and HTTP client abstraction to support multiple registry types with authentication and metadata fetching.

## Requirements
- Registry interface for different source types
- HTTP registry client with authentication
- Registry metadata fetching
- Error handling and retry logic

## Tasks
- [ ] **Define Registry interface**:
  ```go
  type Registry interface {
    GetRuleset(name, version string) (*Ruleset, error)
    ListVersions(name string) ([]string, error)
    Download(name, version string) (io.ReadCloser, error)
    GetMetadata(name string) (*Metadata, error)
  }
  ```
- [ ] **HTTP client implementation**:
  - Support for different authentication methods
  - Request/response handling
  - Timeout and retry configuration
  - User-agent and headers
- [ ] **Authentication handling**:
  - Bearer tokens
  - Basic authentication
  - Custom headers
  - Token refresh logic
- [ ] **Metadata structures**:
  ```go
  type Metadata struct {
    Name        string
    Description string
    Versions    []Version
    Repository  string
  }
  ```
- [ ] **Error handling**:
  - Network errors
  - Authentication failures
  - Rate limiting
  - Registry unavailable

## Acceptance Criteria
- [ ] Registry interface supports all required operations
- [ ] HTTP client handles authentication correctly
- [ ] Metadata fetching works for all registry types
- [ ] Proper error handling and user feedback
- [ ] Retry logic for transient failures
- [ ] Request timeouts prevent hanging

## Dependencies
- net/http (standard library)
- context (standard library)

## Files to Create
- `internal/registry/interface.go`
- `internal/registry/http.go`
- `internal/registry/auth.go`
- `internal/registry/metadata.go`

## Notes
- Plan for caching of metadata responses
- Consider rate limiting for API calls
- Support for registry health checks
- Design for easy testing with mocks
