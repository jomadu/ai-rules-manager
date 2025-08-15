# Architecture Overview

AI Rules Manager (ARM) is designed as a modular package manager for AI coding assistant rulesets with support for multiple registry types, content-based caching, and team synchronization.

## High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Layer     │    │  Config Layer   │    │  Cache Layer    │
│                 │    │                 │    │                 │
│ • Commands      │◄──►│ • Hierarchical  │◄──►│ • Content-based │
│ • Validation    │    │ • Merging       │    │ • TTL & Limits  │
│ • Output        │    │ • Environment   │    │ • Cleanup       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Registry Layer  │    │ Install Layer   │    │ Version Layer   │
│                 │    │                 │    │                 │
│ • Multi-type    │◄──►│ • Orchestration │◄──►│ • Semver        │
│ • Auth          │    │ • File copying  │    │ • Constraints   │
│ • Patterns      │    │ • Namespacing   │    │ • Resolution    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Core Components

### 1. CLI Layer (`internal/cli/`)
- **Purpose**: User interface and command orchestration
- **Key Files**: `commands.go`
- **Responsibilities**:
  - Command parsing and validation
  - Flag handling and global options
  - Output formatting (text/JSON)
  - Error handling and user feedback

### 2. Configuration Layer (`internal/config/`)
- **Purpose**: Hierarchical configuration management
- **Key Files**: `config.go`, `manifest.go`, `cache.go`
- **Responsibilities**:
  - INI file parsing (`.armrc`)
  - JSON file parsing (`arm.json`, `arm.lock`)
  - Environment variable expansion
  - Configuration merging (global → local)
  - Validation and defaults

### 3. Registry Layer (`internal/registry/`)
- **Purpose**: Multi-registry abstraction and implementation
- **Key Files**: `registry.go`, `factory.go`, `git.go`, `s3.go`, etc.
- **Responsibilities**:
  - Registry type detection and creation
  - Authentication handling
  - Content downloading with patterns
  - Version resolution and caching
  - Search capabilities (where supported)

### 4. Cache Layer (`internal/cache/`)
- **Purpose**: Content-based caching with intelligent management
- **Key Files**: `manager.go`, `ruleset_storage.go`, `metadata.go`
- **Responsibilities**:
  - Content-addressable storage
  - TTL-based expiration
  - Size-based eviction
  - Registry and ruleset mapping
  - Cleanup and maintenance

### 5. Installation Layer (`internal/install/`)
- **Purpose**: Orchestrated installation workflow
- **Key Files**: `orchestrator.go`, `installer.go`
- **Responsibilities**:
  - Multi-channel installation
  - File copying with ARM namespacing
  - Lock file management
  - Manifest updates
  - Rollback on failure

### 6. Version Layer (`internal/version/`)
- **Purpose**: Semantic versioning and constraint resolution
- **Key Files**: `resolver.go`
- **Responsibilities**:
  - Semver parsing and validation
  - Constraint matching (`^`, `~`, `>=`, etc.)
  - Latest version resolution
  - Version comparison and sorting

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
    DownloadRuleset(ctx context.Context, name, version, destDir string) error
    ListVersions(ctx context.Context, name string) ([]string, error)
    Close() error
}

type Searcher interface {
    Search(ctx context.Context, query string) ([]SearchResult, error)
}

type PatternDownloader interface {
    DownloadRulesetWithPatterns(ctx context.Context, name, version, destDir string, patterns []string) error
}
```

### Registry Types

#### Git Registry
- **Authentication**: Token-based (GitHub, GitLab)
- **Versioning**: Git tags and branches
- **Patterns**: Glob-based file selection
- **Caching**: Content and metadata caching

#### S3 Registry
- **Authentication**: AWS IAM (profiles, roles)
- **Versioning**: Object versioning
- **Structure**: Bucket/prefix organization
- **Caching**: Object metadata caching

#### HTTPS Registry
- **Authentication**: Bearer tokens
- **Versioning**: API-based
- **Structure**: RESTful endpoints
- **Caching**: Response caching

#### Local Registry
- **Authentication**: Filesystem permissions
- **Versioning**: Directory structure
- **Structure**: Local paths
- **Caching**: File modification time

## Cache Architecture

### Content-Based Storage
- **Key Generation**: SHA-256 of content + metadata
- **Deduplication**: Identical content shares storage
- **Integrity**: Content verification on retrieval

### Two-Level Hierarchy
```
cache/
├── registries/          # Registry metadata
│   └── <registry>/
│       ├── metadata.json
│       └── rulesets/
│           └── <ruleset>/
│               └── versions.json
└── rulesets/           # Actual content
    └── <content-hash>/
        ├── metadata.json
        └── files/
```

### Eviction Strategy
1. **TTL-based**: Remove expired entries
2. **Size-based**: LRU eviction when over limit
3. **Reference counting**: Keep referenced content

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
- **Network Reduction**: ~90% fewer downloads for repeated operations
- **Speed Improvement**: ~10x faster for cached content
- **Bandwidth Savings**: Significant for large rulesets

### Concurrency
- **Registry Operations**: Configurable concurrency limits
- **File Operations**: Parallel copying where safe
- **Cache Operations**: Thread-safe with minimal locking

### Memory Usage
- **Streaming**: Large files processed in chunks
- **Lazy Loading**: Content loaded on demand
- **Cleanup**: Automatic memory management
