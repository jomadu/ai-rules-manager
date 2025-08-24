# ARM Components and Design Patterns

> **Primary Focus**: ARM is designed primarily for Git repositories. The abstractions below are future-proofing for potential support of other registry types (HTTP, S3, Container registries) but Git repositories remain the core use case and implementation priority.

## 1. Registry Manager

### Purpose
Abstract registry operations and content retrieval

### Responsibilities
- List available versions (tags, branches, packages, etc.)
- Retrieve content using registry-specific selectors
- Handle authentication for private registries
- Abstract away registry implementation details (Git, HTTP, S3, etc.)

### Interface Design
```go
type Registry interface {
    ListVersions() ([]VersionRef, error)
    GetContent(versionRef VersionRef, selector ContentSelector) ([]File, error)
    GetMetadata() RegistryMetadata
}

type VersionRef struct {
    ID       string            // "1.2.3", "main", "abc123", "latest"
    Type     VersionRefType    // Tag, Branch, Commit, Label
    Metadata map[string]string // Registry-specific data
}

type ContentSelector interface {
    // String returns a string representation for cache key generation
    String() string

    // Validate checks if the selector configuration is valid
    Validate() error
}

type File struct {
    Path    string // Relative path within ruleset
    Content []byte // File content
    Size    int64  // File size in bytes
}

// Git-specific content selector
type GitContentSelector struct {
    Patterns    []string // ["rules/amazonq/*.md", "rules/cursor/*.mdc"]
    Excludes    []string // ["**/*.test.md", "**/README.md"] - optional exclusions
}

// Implement ContentSelector interface
func (g GitContentSelector) String() string {
    return fmt.Sprintf("patterns:%v,excludes:%v", g.Patterns, g.Excludes)
}

// Validation method
func (g GitContentSelector) Validate() error {
    if len(g.Patterns) == 0 {
        return errors.New("at least one pattern is required")
    }
    for _, pattern := range g.Patterns {
        if _, err := filepath.Match(pattern, ""); err != nil {
            return fmt.Errorf("invalid pattern %q: %w", pattern, err)
        }
    }
    return nil
}
```

### Registry Types

#### Git Registry (Primary Implementation)
- **Versions**: Tags (`v1.2.3`), branches (`main`), commits (`abc123`)
- **Content Selection**: File patterns (`rules/amazonq/*.md`)
- **Use Case**: Development repositories with file-based rulesets
- **Status**: Core implementation focus

### Design Patterns
- **Repository Pattern**: Abstract registry operations behind interfaces
- **Provider Pattern**: Registry-specific component creation

## 2. Cache System

### Purpose
Content-addressable storage for performance and offline capability

### Responsibilities
- SHA256-based content hashing for deduplication
- TTL-based cache invalidation
- Hierarchical storage (registry → ruleset → version)
- Cache size management and cleanup

### Cache Structure
Cache keys are registry-specific to handle different uniqueness requirements:

#### Git Registry Caching
- Registry cache key: `sha256(url + "git")`
- Ruleset cache key: `sha256(ruleset_name + patterns)` - name + patterns define uniqueness
- Version cache key: commit hash

### Interface Design
```go
type CacheKeyGenerator interface {
    RegistryKey(url string) string
    RulesetKey(rulesetName string, selector ContentSelector) string
    VersionKey(versionRef VersionRef) string
}

// Registry-specific implementations
type GitCacheKeyGenerator struct{}
```

### Design Patterns
- **Content-Addressable Storage**: Registry-specific cache keys for integrity and deduplication
- **Strategy Pattern**: Cache key generation based on registry type

## 3. Version Resolver

### Purpose
Version constraint logic and decision making

### Responsibilities
- Parse version constraints (`^1.0.0`, `~1.0.0`, `=1.0.0`, `>=1.0.0`)
- Apply constraint satisfaction logic to available versions
- Handle registry-agnostic version resolution
- Support different version reference types (tags, branches, labels)
- Determine when to refresh mutable references

### Interface Design
```go
type VersionResolver interface {
    ResolveVersion(constraint string, available []VersionRef) (VersionRef, error)
}

type ContentResolver interface {
    ResolveContent(selector ContentSelector, available []File) ([]File, error)
}
```

### Content Selectors
```go
// Registry-specific implementations - ONLY for content extraction
type PackageContentSelector struct {
    // No selection needed - entire package extracted
}

type ObjectContentSelector struct {
    Prefix string // "rules/amazonq/" - which objects to extract
}
```

### Design Patterns
- **Strategy Pattern**: Version resolution and content selection based on registry type

## 4. File System Manager

### Purpose
Atomic file operations for ruleset installation

### Responsibilities
- Install files to configured sink directories from `.armrc.json`
- Create ARM directory structure: `{sink_dir}/arm/{registry}/{ruleset}/{version}/`
- Atomic directory operations with rollback
- Multi-sink installation based on sink configuration
- Cleanup of orphaned installations and empty ARM directories

### Design Patterns
- **Atomic Operations**: File system operations with transaction-like behavior
  - Create temporary directories
  - Perform operations in temp space
  - Atomic move on success
  - Cleanup on failure

## 5. Configuration Manager

### Purpose
Three-tier project configuration system

### Responsibilities
- Infrastructure config (`.armrc.json`) - registry definitions and sink mappings
- Dependency config (`arm.json`) - ruleset dependencies and version constraints
- Lock file (`arm.lock`) - resolved versions, commit hashes, and reproducible state
- Configuration validation, merging, and migration
- Project-scoped configuration (no global user settings)

### Configuration Files

#### `.armrc.json` - Infrastructure Configuration
**Purpose**: "Where can I find rulesets and where should I install them?"
- Registry definitions: `ai-rules` → `https://github.com/my-user/ai-rules`
- Sink mappings: `cursor` → `.cursor/rules`, `q` → `.amazonq/rules`
- Project-specific infrastructure setup

#### `arm.json` - Dependency Configuration
**Purpose**: "What rulesets do I want and what versions?"
- Ruleset dependencies with version constraints
- Content selectors (patterns, packages, prefixes)
- Similar to `package.json` dependencies section

#### `arm.lock` - Resolved State
**Purpose**: "Exactly what was installed and from where?"
- Resolved commit hashes and exact versions
- Registry URLs and metadata for reproducible installs
- Similar to `package-lock.json` or `yarn.lock`

## Cross-Component Design Patterns

### Provider Pattern
Registry-specific component creation:
```go
type RegistryProvider interface {
    CreateRegistry(config RegistryConfig) (Registry, error)
    CreateVersionResolver() (VersionResolver, error)
    CreateContentResolver() (ContentResolver, error)
}

// Registry-specific implementations
type GitRegistryProvider struct{}
type HTTPRegistryProvider struct{}
type S3RegistryProvider struct{}
```

### Command Pattern
CLI operations with validation and rollback:
```go
type Command interface {
    Validate() error
    Execute() error
    Rollback() error
}
```

## Component Interactions

### Installation Flow
1. **Configuration Manager** parses registry and ruleset info
2. **Registry Manager** lists available versions
3. **Version Resolver** applies constraint logic to determine target version
4. **Cache System** checks for cached content
5. **Registry Manager** fetches content using content selector if cache miss
6. **Content Resolver** applies registry-specific content selection
7. **File System Manager** installs files to configured sinks atomically
8. **Configuration Manager** updates lock file

### Update Flow
1. **Configuration Manager** reads current state
2. **Registry Manager** fetches latest versions
3. **Version Resolver** finds updates within constraints
4. **Cache System** validates cached content
5. **File System Manager** performs atomic updates to configured sinks
6. **Configuration Manager** updates resolved versions

### Outdated Flow
1. **Configuration Manager** reads current lock file state
2. **Registry Manager** fetches latest available versions
3. **Version Resolver** compares current resolved versions with latest compatible versions
4. **Version Resolver** identifies outdated rulesets within constraint boundaries
5. **Configuration Manager** generates outdated report showing:
   - Current version (what's installed)
   - Wanted version (latest within constraints)
   - Latest version (newest available, may break constraints)

### Uninstall Flow
1. **Configuration Manager** validates ruleset exists in project config
2. **File System Manager** identifies installed files from lock file metadata
3. **File System Manager** creates backup of files to be removed
4. **File System Manager** atomically removes ruleset files and directories from sinks
5. **Configuration Manager** removes ruleset from `arm.json` and `arm.lock`
6. **File System Manager** cleans up empty ARM directories if no other rulesets remain
7. **Cache System** optionally retains cached content for potential reinstall break constraints)

### Cache Management
1. **Cache System** monitors size and TTL
2. **Registry Manager** validates cache integrity
3. **File System Manager** performs cleanup operations
4. **Configuration Manager** tracks cache metadata
