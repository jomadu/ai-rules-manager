# Architecture Overview

AI Rules Manager (ARM) is a modular package manager for AI coding assistant rulesets with hierarchical configuration, multi-registry support, and intelligent caching.

## High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Layer     │    │  Config Layer   │    │  Cache Layer    │
│                 │    │                 │    │                 │
│ • Commands      │◄──►│ • Hierarchical  │◄──►│ • Registry-based│
│ • Validation    │    │ • Merging       │    │ • URL Normalize │
│ • Output        │    │ • Environment   │    │ • TTL & LRU     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Registry Layer  │    │ Install Layer   │    │ Version Layer   │
│                 │    │                 │    │                 │
│ • Multi-type    │◄──►│ • Orchestration │◄──►│ • Semver        │
│ • Auth          │    │ • Rate Limiting │    │ • Git Refs      │
│ • Patterns      │    │ • Concurrency   │    │ • Resolution    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Core Components

### 1. CLI Layer (`internal/cli/`)
- **Purpose**: User interface and command orchestration
- **Key Files**: `commands.go`, `error_handling_test.go`, `pattern_test.go`, `search_test.go`, `update_test.go`
- **Responsibilities**:
  - Command parsing and validation
  - Flag handling and global options
  - Output formatting (text/JSON)
  - Error handling and user feedback
  - Pattern matching and search functionality

### 2. Configuration Layer (`internal/config/`)
- **Purpose**: Hierarchical configuration management
- **Key Files**: `config.go`, `manifest.go`, `cache.go`, `path_resolver.go`
- **Responsibilities**:
  - INI file parsing (`.armrc`) with environment variable expansion
  - JSON file parsing (`arm.json`, `arm.lock`)
  - Configuration merging (global → local) at key level
  - Path resolution and validation
  - Cache configuration management

### 3. Registry Layer (`internal/registry/`)
- **Purpose**: Multi-registry abstraction and implementation
- **Key Files**: `registry.go`, `factory.go`, `base_git_registry.go`, `git_common.go`, `git_local.go`, `git_operations.go`, `remote_git_operations.go`, `gitlab.go`, `https.go`, `s3.go`, `local.go`
- **Responsibilities**:
  - Registry type detection and creation
  - Authentication handling per registry type
  - Git operations (local and remote)
  - Pattern-based file downloading
  - Version resolution and search
  - Registry-specific error handling

### 4. Cache Layer (`internal/cache/`)
- **Purpose**: Registry-based caching with URL normalization
- **Key Files**: `manager.go`, `metadata.go`, `registry_map.go`, `ruleset_map.go`, `ruleset_storage.go`, `url_normalizer.go`
- **Responsibilities**:
  - Registry-specific cache key generation
  - URL normalization for consistent caching
  - Registry and ruleset mapping
  - TTL-based expiration and LRU eviction
  - Cache metadata management

### 5. Installation Layer (`internal/install/`)
- **Purpose**: Orchestrated installation with concurrency control
- **Key Files**: `orchestrator.go`, `installer.go`, `workflow_test.go`
- **Responsibilities**:
  - Multi-channel installation
  - Rate limiting and concurrency control
  - Token bucket rate limiting
  - File copying with ARM namespacing
  - Lock file and manifest management

### 6. Version Layer (`internal/version/`)
- **Purpose**: Multi-strategy version resolution
- **Key Files**: `resolver.go`
- **Responsibilities**:
  - Semver constraint resolution
  - Git reference resolution (branches, commits, latest)
  - Exact version matching
  - Version validation and comparison

### 7. Update Layer (`internal/update/`)
- **Purpose**: Ruleset update management
- **Key Files**: `service.go`
- **Responsibilities**:
  - Outdated ruleset detection
  - Version comparison and updates
  - Update orchestration

## Data Flow

### Installation Flow
```
1. CLI Command
   ↓
2. Config Loading (global + local merge)
   ↓
3. Registry Creation (with auth + cache)
   ↓
4. Version Resolution (semver constraints)
   ↓
5. Content Download (with patterns)
   ↓
6. Cache Storage (content-based)
   ↓
7. Installation Orchestration
   ↓
8. File Copying (to channels with ARM namespace)
   ↓
9. Lock File Update
   ↓
10. Manifest Update
```

### Update Flow
```
1. CLI Command
   ↓
2. Lock File Analysis
   ↓
3. Version Comparison (current vs available)
   ↓
4. Selective Updates (outdated only)
   ↓
5. Installation Flow (for updated rulesets)
```

## Configuration Architecture

### File Hierarchy
```
Global:  ~/.arm/.armrc, ~/.arm/arm.json
Local:   .armrc, arm.json, arm.lock
```

### Merging Strategy
- **Key-level merging**: Local values override global values
- **Nested maps**: Registry configs merge at individual key level
- **Arrays**: Local arrays completely replace global arrays
- **Lock file**: Always local (no merging)

### Configuration Types
- **INI Format** (`.armrc`): Registries, type defaults, network settings
- **JSON Format** (`arm.json`): Channels, rulesets, engines
- **Lock Format** (`arm.lock`): Resolved versions and metadata

## Registry Architecture

### Registry Interface
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

type Searcher interface {
    Search(ctx context.Context, query string) ([]SearchResult, error)
}
```

### Registry Types

#### Git Registry
- **Authentication**: Token-based (GitHub, GitLab)
- **Versioning**: Git tags, branches, and commit hashes
- **Patterns**: Glob-based file selection
- **Operations**: Local and remote Git operations

#### Git-Local Registry
- **Authentication**: Filesystem permissions
- **Versioning**: Local Git repository references
- **Structure**: Local Git repository paths
- **Operations**: Local Git operations only

#### S3 Registry
- **Authentication**: AWS IAM (profiles, roles)
- **Versioning**: Object versioning
- **Structure**: Bucket/prefix organization
- **Operations**: S3 API calls

#### HTTPS Registry
- **Authentication**: Bearer tokens
- **Versioning**: API-based
- **Structure**: RESTful endpoints
- **Operations**: HTTP requests

#### Local Registry
- **Authentication**: Filesystem permissions
- **Versioning**: Directory structure
- **Structure**: Local paths
- **Operations**: File system operations

## Cache Architecture

### Registry-Based Storage
- **Key Generation**: SHA-256 of registry type + normalized URL
- **URL Normalization**: Consistent URL formatting per registry type
- **Registry Mapping**: Track cache keys to registry configurations

### Cache Structure
```
cache/
├── registries/
│   └── <registry-hash>/
│       ├── cache-info.json
│       ├── repository/      # Git clones
│       └── rulesets/
│           └── <ruleset-hash>/
│               └── <version>/
├── temp/                    # Temporary files
└── registry-map.json        # Registry mappings
```

### Eviction Strategy
1. **TTL-based**: Remove expired entries based on last accessed time
2. **Size-based**: LRU eviction when over configured limit
3. **Cleanup**: Automatic cleanup of expired and oversized entries

## Security Considerations

### Authentication
- **Token Storage**: Environment variables preferred
- **Credential Isolation**: Per-registry authentication
- **Secure Defaults**: HTTPS-only for remote registries

### File Operations
- **Path Validation**: Prevent directory traversal
- **Permission Checks**: Validate write permissions
- **Atomic Operations**: Prevent partial installations

### Network Security
- **TLS Verification**: Certificate validation
- **Timeout Handling**: Prevent hanging operations
- **Rate Limiting**: Respect API limits

## Installation Architecture

### Orchestration
- **Concurrency Control**: Per-registry concurrency limits
- **Rate Limiting**: Token bucket algorithm per registry
- **Progress Tracking**: Installation progress callbacks
- **Error Handling**: Graceful failure handling

### Multi-Channel Installation
- **Channel Configuration**: Multiple target directories per channel
- **ARM Namespacing**: Files installed under `arm/<registry>/<ruleset>/`
- **Atomic Operations**: Rollback on failure

## Version Resolution

### Resolution Strategies
1. **Semver Resolver**: Handles semantic version constraints (`^1.0.0`, `~1.2.0`, `>=1.0.0`)
2. **Git Resolver**: Handles Git references (`latest`, branch names, commit hashes)
3. **Exact Resolver**: Handles exact version matches (`=1.2.3`)

### Version Types
- **Semantic Versions**: Standard semver with constraint matching
- **Git References**: Branches, tags, commit hashes, and `latest`
- **Exact Matches**: Precise version specifications

## Security Considerations

### Authentication
- **Token Storage**: Environment variables preferred
- **Credential Isolation**: Per-registry authentication
- **Secure Defaults**: HTTPS-only for remote registries

### File Operations
- **Path Validation**: Prevent directory traversal
- **Permission Checks**: Validate write permissions
- **Atomic Operations**: Prevent partial installations

### Network Security
- **TLS Verification**: Certificate validation
- **Timeout Handling**: Prevent hanging operations
- **Rate Limiting**: Respect API limits

## Performance Characteristics

### Caching Benefits
- **Network Reduction**: Avoids repeated downloads for same registry/version
- **Speed Improvement**: Faster access to previously downloaded content
- **Bandwidth Savings**: Reduces network usage for large rulesets

### Concurrency
- **Registry Operations**: Configurable concurrency limits per registry type
- **Rate Limiting**: Token bucket algorithm prevents API abuse
- **Parallel Installation**: Multiple rulesets installed concurrently

### Memory Usage
- **Streaming**: Large files processed in chunks
- **Lazy Loading**: Content loaded on demand
- **Cleanup**: Automatic memory management and cache cleanup
