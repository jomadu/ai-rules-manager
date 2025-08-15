# Registry System

Comprehensive technical specification for ARM's multi-registry architecture.

## Overview

ARM supports multiple registry types through a unified interface, enabling teams to store and distribute rulesets using their preferred infrastructure. Each registry type implements the core `Registry` interface with optional extensions for advanced features.

## Registry Interface

### Core Interface
```go
type Registry interface {
    GetRulesets(ctx context.Context, patterns []string) ([]RulesetInfo, error)
    GetRuleset(ctx context.Context, name, version string) (*RulesetInfo, error)
    DownloadRuleset(ctx context.Context, name, version, destDir string) error
    DownloadRulesetWithPatterns(ctx context.Context, name, version, destDir string, patterns []string) error
    GetVersions(ctx context.Context, name string) ([]string, error)
    GetType() string
    GetName() string
    Close() error
}
```

### Optional Extensions
```go
type Searcher interface {
    Search(ctx context.Context, query string) ([]SearchResult, error)
}
```

## Registry Types

### 1. Git Registry (`git`)

**Purpose**: GitHub, GitLab, and generic Git repositories

#### Configuration
```ini
[registries]
default = https://github.com/org/rules-repo

[registries.default]
type = git
token = $GITHUB_TOKEN          # Optional for private repos
api_type = github              # Optional: github, gitlab, generic
api_version = 2022-11-28       # Optional: API version
```

#### Implementation Details
- **File**: `internal/registry/git.go`
- **Authentication**: Bearer token for API access
- **Versioning**: Git tags (semver) and branches
- **Pattern Support**: Yes (glob patterns)
- **Search Support**: No (returns error)
- **Caching**: Content and metadata

#### Version Resolution
```go
// Supported version formats
"latest"           // Latest tag
"main"             // Branch name
"v1.2.3"          // Exact tag
"^1.2.0"          // Semver constraint
">=1.0.0 <2.0.0"  // Range constraint
```

#### Pattern Matching
```bash
# Install specific files
arm install ruleset --patterns "rules/*.md,docs/*.md"

# Exclude files
arm install ruleset --patterns "**/*.md,!**/internal/**"
```

#### API Integration
- **GitHub**: Uses GitHub API v4 (GraphQL) and v3 (REST)
- **GitLab**: Uses GitLab API v4
- **Generic**: Falls back to Git operations

### 2. Git-Local Registry (`git-local`)

**Purpose**: Local Git repositories for development and testing

#### Configuration
```ini
[registries]
local-dev = /path/to/local/repo

[registries.local-dev]
type = git-local
```

#### Implementation Details
- **File**: `internal/registry/git_local.go`
- **Authentication**: Filesystem permissions
- **Versioning**: Local Git tags and branches
- **Pattern Support**: Yes
- **Search Support**: No
- **Caching**: File modification time

#### Use Cases
- Development and testing of rulesets
- Offline work
- Custom CI/CD pipelines

### 3. S3 Registry (`s3`)

**Purpose**: AWS S3 buckets with IAM authentication

#### Configuration
```ini
[registries]
s3-rules = my-rules-bucket

[registries.s3-rules]
type = s3
region = us-east-1             # Required
profile = my-aws-profile       # Optional
prefix = /rules/               # Optional
```

#### Implementation Details
- **File**: `internal/registry/s3.go`
- **Authentication**: AWS IAM (profiles, roles, instance metadata)
- **Versioning**: S3 object versioning
- **Pattern Support**: No (structured storage)
- **Search Support**: Limited (prefix-based)
- **Caching**: Object metadata

#### Bucket Structure
```
bucket/
├── prefix/                    # Optional prefix
│   └── ruleset-name/
│       ├── v1.0.0/
│       │   └── ruleset.tar.gz
│       └── v1.1.0/
│           └── ruleset.tar.gz
```

Note: Only `ruleset.tar.gz` files are used. No metadata files or symlinks.

### 4. HTTPS Registry (`https`)

**Purpose**: Generic HTTPS endpoints with RESTful API

#### Configuration
```ini
[registries]
https-registry = https://registry.example.com

[registries.https-registry]
type = https
authToken = $REGISTRY_TOKEN    # Optional
```

#### Implementation Details
- **File**: `internal/registry/https.go`
- **Authentication**: Bearer token
- **Versioning**: API-defined
- **Pattern Support**: No
- **Search Support**: Yes (if API supports)
- **Caching**: Response caching

#### API Endpoints
```
GET /manifest.json                      # Get manifest with rulesets and versions
GET /{name}/{version}/ruleset.tar.gz    # Download ruleset tarball
```

### 5. GitLab Registry (`gitlab`)

**Purpose**: GitLab Package Registry integration

#### Configuration
```ini
[registries]
gitlab-rules = https://gitlab.example.com/projects/123

[registries.gitlab-rules]
type = gitlab
authToken = $GITLAB_TOKEN
apiVersion = 4                 # Optional
```

#### Implementation Details
- **File**: `internal/registry/gitlab.go`
- **Authentication**: Personal access token or CI token
- **Versioning**: GitLab releases and tags
- **Pattern Support**: No (uses pre-packaged tar.gz files)
- **Search Support**: No
- **Caching**: API response caching

### 6. Local Registry (`local`)

**Purpose**: Local filesystem directories

#### Configuration
```ini
[registries]
local-rules = /path/to/rules

[registries.local-rules]
type = local
```

#### Implementation Details
- **File**: `internal/registry/local.go`
- **Authentication**: Filesystem permissions
- **Versioning**: Directory structure
- **Pattern Support**: No (uses pre-packaged tar.gz files)
- **Search Support**: No
- **Caching**: File modification time

#### Directory Structure
```
/path/to/rules/
├── ruleset-1/
│   ├── v1.0.0/
│   │   ├── metadata.json
│   │   └── rules/
│   └── latest -> v1.0.0
└── ruleset-2/
    └── v2.1.0/
        ├── metadata.json
        └── guidelines/
```

## Registry Factory

### Creation Process
```go
func CreateRegistry(config *RegistryConfig, auth *AuthConfig) (Registry, error) {
    switch config.Type {
    case "local":
        return NewLocalRegistry(config)
    case "git":
        return NewGitRegistryWithCache(config, auth, cacheManager)
    case "git-local":
        return NewGitLocalRegistry(config, auth)
    case "https":
        return NewHTTPSRegistry(config, auth)
    case "s3":
        return NewS3Registry(config, auth)
    case "gitlab":
        return NewGitLabRegistry(config, auth)
    default:
        return nil, fmt.Errorf("unsupported registry type: %s", config.Type)
    }
}
```

### Configuration Validation
Each registry type validates its specific requirements:
- **Git**: URL format, optional authentication
- **S3**: Bucket name, region, AWS credentials
- **HTTPS**: URL format, optional authentication
- **Local**: Path existence and permissions

## Authentication

### Authentication Configuration
```go
type AuthConfig struct {
    Token      string `json:"token"`
    Username   string `json:"username"`
    Password   string `json:"password"`
    Profile    string `json:"profile"`     // For AWS profiles
    Region     string `json:"region"`      // For AWS regions
    APIType    string `json:"api_type"`    // For API-specific auth
    APIVersion string `json:"api_version"` // For API versioning
}
```

### Environment Variable Support
```bash
# Git registries
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
export GITLAB_TOKEN=glpat-xxxxxxxxxxxx

# S3 registries
export AWS_PROFILE=my-profile
export AWS_REGION=us-east-1

# HTTPS registries
export REGISTRY_TOKEN=token_xxxxxxxxxxxx
```

### Security Considerations
- Tokens stored in environment variables (not config files)
- HTTPS-only for remote registries
- Certificate validation enabled
- Rate limiting respected

## Caching Integration

### Cache Key Generation
```go
func GenerateCacheKey(registry, ruleset, version string, patterns []string) string {
    h := sha256.New()
    h.Write([]byte(registry))
    h.Write([]byte(ruleset))
    h.Write([]byte(version))
    for _, pattern := range patterns {
        h.Write([]byte(pattern))
    }
    return hex.EncodeToString(h.Sum(nil))
}
```

### Cache Hierarchy
```
cache/
├── registries/
│   └── {registry-name}/
│       ├── metadata.json      # Registry metadata
│       └── rulesets/
│           └── {ruleset-name}/
│               └── versions.json  # Available versions
└── rulesets/
    └── {content-hash}/
        ├── metadata.json      # Content metadata
        └── files/             # Actual files
```

### TTL Configuration
```ini
[cache]
ttl = 24h                      # Default TTL
registryMetadataTTL = 1h       # Registry metadata
rulesetMetadataTTL = 6h        # Ruleset metadata
contentTTL = 7d                # Actual content
```

## Error Handling

### Registry-Specific Errors
```go
type RegistryError struct {
    Type     string
    Registry string
    Ruleset  string
    Version  string
    Cause    error
}

// Common error types
var (
    ErrRulesetNotFound   = errors.New("ruleset not found")
    ErrVersionNotFound   = errors.New("version not found")
    ErrAuthenticationFailed = errors.New("authentication failed")
    ErrNetworkTimeout    = errors.New("network timeout")
    ErrInvalidPattern    = errors.New("invalid pattern")
)
```

### Retry Logic
- **Network errors**: Exponential backoff (3 attempts)
- **Rate limiting**: Respect Retry-After headers
- **Authentication**: Single retry after token refresh

## Performance Characteristics

### Concurrency Limits
```ini
[git]
concurrency = 1               # Conservative for API limits

[s3]
concurrency = 10              # Higher for AWS

[https]
concurrency = 5               # Moderate for generic APIs

[local]
concurrency = 20              # High for filesystem
```

### Rate Limiting
```ini
[git]
rateLimit = 10/minute         # GitHub API limits

[s3]
rateLimit = 100/hour          # AWS request limits

[https]
rateLimit = 30/minute         # Conservative default
```

### Memory Usage
- **Streaming downloads**: Large files processed in chunks
- **Pattern matching**: Efficient glob implementation
- **Cache management**: LRU eviction with size limits

## Testing Strategy

### Unit Tests
- Mock implementations for each registry type
- Authentication flow testing
- Error condition handling
- Pattern matching validation

### Integration Tests
- Real registry interactions (with test data)
- End-to-end download workflows
- Cache behavior validation
- Performance benchmarking

### Test Registries
```bash
# Set up test Git registry
./tests/integration/git/setup-test-repos.sh

# Run registry-specific tests
go test ./internal/registry/git_test.go
go test ./internal/registry/s3_test.go
```
