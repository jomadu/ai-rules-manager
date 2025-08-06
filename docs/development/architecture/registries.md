# Registry System Design

Architecture and design of ARM's registry abstraction layer.

## Registry Interface

### Core Interface

```go
type Registry interface {
    GetVersions(name string) ([]string, error)
    Download(name, version string) (io.ReadCloser, error)
    GetMetadata(name string) (*Metadata, error)
}
```

### Registry Capabilities

| Feature | GitLab | S3 | HTTP | Filesystem | Git |
|---------|--------|----|----- |------------|-----|
| Version Discovery | ✅ | ✅ | ❌ | ✅ | ✅ |
| Metadata API | ✅ | ❌ | ❌ | ❌ | ❌ |
| Authentication | ✅ | ✅ | ✅ | ❌ | ✅ |
| Caching | ✅ | ✅ | ✅ | ❌ | ✅ |

## Registry Manager

### Responsibilities
- Registry instance creation and management
- Registry selection based on package name
- Configuration parsing and validation
- Authentication handling

### Registry Resolution

```go
// company@package → looks for sources.company
// package → uses sources.default
func (m *Manager) ResolveRegistry(packageName string) (Registry, error) {
    if strings.Contains(packageName, "@") {
        parts := strings.Split(packageName, "@")
        source := parts[0]
        return m.getRegistry(source)
    }
    return m.getRegistry("default")
}
```

## Registry Types

### GitLab Package Registry
- **API-based** - Uses GitLab Package Registry API
- **Authentication** - Personal access tokens
- **Version Discovery** - API endpoint for package versions
- **Metadata** - Full package information available

### AWS S3
- **Prefix-based** - Uses S3 object prefixes for organization
- **Authentication** - AWS credentials (access key/secret)
- **Version Discovery** - S3 prefix listing
- **Metadata** - Minimal, inferred from object structure

### HTTP Registry
- **File-based** - Simple HTTP file server
- **Authentication** - Optional HTTP basic auth or tokens
- **Version Discovery** - Not supported (exact versions required)
- **Metadata** - None available

### Filesystem Registry
- **Directory-based** - Local directory structure
- **Authentication** - File system permissions
- **Version Discovery** - Directory listing
- **Metadata** - None available

### Git Registry
- **Repository-based** - Direct git repository access
- **Authentication** - Git credentials (SSH/HTTPS)
- **Version Discovery** - Git tags and branches
- **Metadata** - None available

## Metadata Handling

### Metadata Structure

```go
type Metadata struct {
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Description string            `json:"description"`
    Author      string            `json:"author"`
    Tags        []string          `json:"tags"`
    Files       []string          `json:"files"`
    Checksum    string            `json:"checksum"`
    Extra       map[string]string `json:"extra"`
}
```

### Registry-Specific Metadata

**GitLab**: Full metadata from Package Registry API
**S3**: Minimal metadata inferred from object properties
**HTTP**: No metadata available
**Filesystem**: No metadata available
**Git**: No metadata available

## Authentication

### Token-Based Authentication

```go
type AuthConfig struct {
    Type     string `ini:"type"`
    Token    string `ini:"authToken"`
    Username string `ini:"username"`
    Password string `ini:"password"`
}
```

### Environment Variable Support

```ini
[sources.company]
type = gitlab
authToken = ${GITLAB_TOKEN}
```

### Security Considerations
- Tokens stored in environment variables
- No credentials in configuration files
- Secure token transmission (HTTPS only)
- Token masking in logs and output

## Version Discovery

### Semantic Version Support

All registries that support version discovery return semantic versions:
- `1.0.0`, `1.0.1`, `1.1.0`
- Pre-release versions: `1.0.0-alpha.1`
- Build metadata: `1.0.0+build.1`

### Version Sorting

Versions are sorted using semantic version rules:
- Latest version first
- Pre-release versions sorted correctly
- Build metadata ignored for comparison

## Caching Strategy

### Registry-Level Caching

- **Version Lists** - Cached to reduce API calls
- **Metadata** - Cached when available
- **Package Downloads** - Cached globally

### Cache Keys

```go
// Version list cache key
versionKey := fmt.Sprintf("registry:%s:versions:%s", registryName, packageName)

// Metadata cache key
metadataKey := fmt.Sprintf("registry:%s:metadata:%s:%s", registryName, packageName, version)

// Package cache key
packageKey := fmt.Sprintf("package:%s:%s:%s", registryName, packageName, version)
```

### Cache Invalidation

- **Time-based** - TTL for version lists and metadata
- **Manual** - `arm clean --cache` command
- **Version-based** - Package cache never expires (immutable)

## Error Handling

### Registry-Specific Errors

```go
type RegistryError struct {
    Registry string
    Package  string
    Version  string
    Err      error
}
```

### Common Error Scenarios
- **Network errors** - Registry unreachable
- **Authentication errors** - Invalid credentials
- **Package not found** - Package doesn't exist
- **Version not found** - Specific version unavailable
- **Permission errors** - Insufficient access rights

### Error Recovery
- Retry logic for transient network errors
- Fallback to cached data when possible
- Clear error messages for user action
