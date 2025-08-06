# Core Components

Detailed breakdown of ARM's core components and their responsibilities.

## Component Structure

```
cmd/arm/           # CLI commands and main entry point
internal/          # Internal packages
├── cache/         # Global cache management
├── config/        # Configuration parsing
├── installer/     # Package installation logic
├── registry/      # Registry implementations
└── updater/       # Update and version checking
pkg/types/         # Public types and interfaces
```

## CLI Layer (cmd/arm/)

### Responsibilities
- Command-line argument parsing
- User interface and output formatting
- Error handling and user feedback
- Command orchestration

### Key Files
- `main.go` - Application entry point and root command
- `install.go` - Install command implementation
- `update.go` - Update command implementation
- `list.go` - List command implementation
- `clean.go` - Clean command implementation

## Configuration (internal/config/)

### Responsibilities
- .armrc file parsing and hierarchy
- rules.json parsing and validation
- Environment variable substitution
- Configuration command handling

### Key Components
- **Config Parser** - INI format parsing with sections
- **Environment Substitution** - ${VAR} expansion
- **Validation** - Configuration correctness checking
- **Hierarchy Management** - User vs project config precedence

## Cache Management (internal/cache/)

### Responsibilities
- Global cache directory management
- Package caching and retrieval
- Metadata caching
- Cache cleanup and maintenance

### Cache Types
- **Package Cache** - Downloaded package archives
- **Metadata Cache** - Registry version information
- **Backup Cache** - Previous versions for rollback

### Key Features
- Thread-safe operations
- Configurable cache policies
- Automatic cleanup of stale entries
- Cache integrity verification

## Installation Logic (internal/installer/)

### Responsibilities
- Package installation to target directories
- File system operations and safety
- Target directory management
- Installation state tracking

### Key Operations
- **Install** - Deploy package to targets
- **Uninstall** - Remove package from targets
- **Verify** - Check installation integrity
- **Backup** - Create installation backups

### Safety Features
- Atomic operations where possible
- Path traversal prevention
- Permission checking
- Rollback capability

## Registry Management (internal/registry/)

### Responsibilities
- Registry abstraction and interface
- Registry-specific implementations
- Authentication handling
- Version discovery and metadata

### Registry Types
- **GitLab** - Package registry with API access
- **S3** - AWS S3 bucket storage
- **HTTP** - Generic HTTP file servers
- **Filesystem** - Local directory registries
- **Git** - Direct git repository access

## Update System (internal/updater/)

### Responsibilities
- Version constraint checking
- Update candidate identification
- Backup and restore operations
- Progress reporting

### Key Features
- Semantic version handling
- Constraint satisfaction
- Parallel version checking
- Rollback on failure

## Public Types (pkg/types/)

### Responsibilities
- Public API definitions
- Shared data structures
- Interface definitions
- Type safety

### Key Types
- **Registry** - Registry interface definition
- **Package** - Package metadata structure
- **Version** - Version information
- **Config** - Configuration structures

## Inter-Component Communication

### Data Flow
1. **CLI** → **Config** → Load configuration
2. **CLI** → **Registry** → Resolve package location
3. **Registry** → **Cache** → Check for cached packages
4. **Registry** → **Installer** → Install package
5. **Installer** → **Cache** → Update cache state

### Error Propagation
- Structured error types with context
- Error wrapping for debugging
- User-friendly error messages
- Detailed logging for troubleshooting

## Component Dependencies

### Dependency Graph
```
CLI Commands
    ├── Config (configuration parsing)
    ├── Registry Manager
    │   ├── Individual Registries
    │   └── Cache
    ├── Installer
    │   ├── Cache
    │   └── File System
    └── Updater
        ├── Registry Manager
        ├── Installer
        └── Cache
```

### Design Principles
- **Loose Coupling** - Components interact through interfaces
- **Single Responsibility** - Each component has a focused purpose
- **Dependency Injection** - Dependencies passed explicitly
- **Testability** - Components can be tested in isolation
