# Technical Documentation

Comprehensive technical specifications for AI Rules Manager.

## System Architecture
1. **[Architecture Overview](architecture.md)** - High-level system design and component interaction
2. **[Configuration System](configuration.md)** - Hierarchical configuration with INI and JSON files
3. **[Cache System](cache.md)** - Content-based caching with TTL and size limits

## Core Systems
4. **[Registry System](registries.md)** - Multi-registry support with authentication
5. **[Installation System](installation.md)** - Orchestrated installation workflow
6. **[Version Resolution](version-resolution.md)** - Semantic versioning and constraint resolution
7. **[File Patterns](patterns.md)** - Pattern matching for selective file installation

## Reference
8. **[Testing Strategy](testing.md)** - Testing approach and coverage requirements

## Implementation Details

### Registry Types
- **Git**: GitHub, GitLab, and generic Git repositories
- **Git-Local**: Local Git repositories for development
- **S3**: AWS S3 buckets with IAM authentication
- **HTTPS**: Generic HTTPS endpoints
- **Local**: Local filesystem directories

### Configuration Hierarchy
1. Global configuration (`~/.arm/.armrc`, `~/.arm/arm.json`)
2. Local configuration (`.armrc`, `arm.json`)
3. Environment variables
4. Command-line flags

### File Organization
```
project/
├── .armrc              # Registry configuration
├── arm.json            # Channels and rulesets
├── arm.lock            # Locked versions
└── .cursor/rules/      # Channel directory
    └── arm/            # ARM namespace
        └── registry/   # Registry name
            └── ruleset/# Ruleset files
```
