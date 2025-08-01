# Product Requirements Document: AI Rules Manager (ARM)

## Overview

AI Rules Manager (ARM) is a command-line interface tool designed to manage the installation and distribution of LLM rulesets for various agentic AI coding tools. It provides a centralized way to install, update, and manage coding rules from multiple source registries.

## Problem Statement

AI coding tools like Cursor and Amazon Q Developer use different rule file formats and locations (`.cursorrules`, `.amazonq/rules`). Currently, there's no standardized way to:
- Distribute and share rulesets across teams
- Manage versions of rulesets
- Install rulesets from multiple sources
- Keep rulesets up-to-date across projects

## Solution

ARM provides a package manager-like experience for AI coding rulesets, similar to npm for JavaScript packages. It supports multiple target formats and source registries with version management and caching.

## Target Users

- Software development teams using AI coding assistants
- DevOps engineers managing development environments
- Individual developers wanting to share and reuse coding rulesets

## Core Features

### 1. Ruleset Installation & Management

#### Install Command
- **Purpose**: Install rulesets by name, version, or from manifest
- **Usage**: 
  - `arm install <ruleset-name>[@version]`
  - `arm install` (from rules.json)
- **Behavior**: Downloads .tar.gz rulesets, extracts to configured target locations, updates rules.json and rules.lock

#### Uninstall Command
- **Purpose**: Remove installed rulesets
- **Usage**: `arm uninstall <ruleset-name>`
- **Behavior**: Removes ruleset from all target locations, updates rules.json and rules.lock

#### Update Command
- **Purpose**: Update installed rulesets to latest versions
- **Usage**: 
  - `arm update` (all rulesets)
  - `arm update <ruleset-name>`
- **Behavior**: Checks for newer versions, updates rulesets, and updates rules.lock

#### List Command
- **Purpose**: Display all installed rulesets with versions
- **Usage**: `arm list`
- **Output**: Table showing ruleset name, current version, and source

#### Outdated Command
- **Purpose**: Show rulesets with available updates
- **Usage**: `arm outdated`
- **Output**: Table showing current vs. available versions

### 2. Configuration Management

#### Config Command
- **Purpose**: View and modify ARM configuration
- **Usage**: 
  - `arm config list`
  - `arm config set <key> <value>`
  - `arm config get <key>`
- **Configurable Items**:
  - Registry sources
  - Authentication tokens
  - Cache location
  - Default targets

#### Configuration Files

##### .armrc
- **Location**: Repository root or user home directory
- **Purpose**: Define source registries and authentication
- **Format**: INI-style configuration
- **Example**:
```ini
[sources]
default = https://registry.armjs.org/
company = https://internal.company-registry.local/

[sources.company]
authToken = $AUTH_TOKEN
```

##### rules.json
- **Location**: Project root
- **Purpose**: Define project ruleset dependencies and targets
- **Format**: JSON
- **Schema**:
```json
{
  "targets": ["string[]"],
  "dependencies": {
    "ruleset-name": "version-spec",
    "source@ruleset-name": "version-spec"
  }
}
```

##### rules.lock
- **Location**: Project root (auto-generated)
- **Purpose**: Lock exact versions for reproducible installs
- **Format**: JSON with resolved dependency tree

### 3. Cache & Cleanup

#### Clean Command
- **Purpose**: Remove unused rulesets and clear cache
- **Usage**: `arm clean`
- **Behavior**: Removes cached files not referenced by any project

#### Cache Management
- **Location**: `.arm/cache/`
- **Structure**: Organized by source and version
- **Behavior**: Automatic caching of downloaded .tar.gz rulesets

### 4. Help & Information

#### Help Command
- **Purpose**: Display usage information
- **Usage**: 
  - `arm help`
  - `arm help <command>`

#### Version Command
- **Purpose**: Display ARM version
- **Usage**: `arm version`

## Implementation Decision

### Language Choice: Go

ARM will be implemented as a compiled binary using Go, chosen over alternatives for the following reasons:

**Go vs Python Package:**
- **Zero dependencies**: Users don't need Python installed
- **Fast startup**: No interpreter overhead for CLI operations
- **Single binary distribution**: Easy installation via curl/wget
- **Professional tooling**: Matches user expectations for package managers
- **Performance**: Better for file operations and tar extraction

**Go vs Other Compiled Languages:**
- **Rust**: Go offers faster development cycle and simpler syntax
- **C++**: Go provides memory safety without manual management complexity
- **Zig**: Go has mature ecosystem and proven CLI tool track record
- **Node.js (compiled)**: Go produces smaller binaries with better performance

**Go Benefits for ARM:**
- Excellent standard library for HTTP, JSON, and tar handling
- Simple cross-compilation for multiple platforms
- Rich ecosystem of CLI frameworks (cobra, viper)
- Proven track record for similar tools (kubectl, docker, terraform)

## Technical Requirements

### Supported Target Formats
- `.cursorrules` (Cursor IDE)
- `.amazonq/rules/` (Amazon Q Developer)
- Extensible architecture for future formats

### Supported Source Registries
- GitLab group/project package registries
- GitHub repository package registries
- AWS S3 buckets
- Generic HTTP endpoints
- Local file system

### File System Structure
```
.arm/
  cache/
    <source>/
      <ruleset-name>/
        <version>/
          <rule-files>
<target-directory>/
  arm/
    <source>/
      <ruleset-name>/
        <version>/
          <rule-files>
```

### Authentication
- Token-based authentication for private registries
- Environment variable support
- Secure credential storage

### Version Management
- Semantic versioning support
- Version range specifications (^, ~, exact)
- Dependency resolution
- Lock file generation

## Non-Functional Requirements

### Performance
- Fast installation through caching
- Parallel downloads when possible
- Minimal disk space usage

### Security
- Secure authentication handling
- Integrity verification of downloaded rulesets
- No execution of arbitrary code

### Reliability
- Atomic operations (install/uninstall)
- Rollback capability on failed operations
- Graceful error handling

### Usability
- Clear error messages
- Progress indicators for long operations
- Consistent command interface

## Success Metrics

- Installation time < 5 seconds for cached rulesets
- Support for 3+ registry types at launch
- Zero data loss during operations
- 95% command success rate

## Future Considerations

- Web-based registry browser
- Ruleset validation and linting
- Integration with CI/CD pipelines
- Plugin system for custom targets
- Ruleset dependency management
- Automated ruleset updates

## Dependencies

### Go Libraries
- CLI framework (cobra)
- Configuration management (viper)
- HTTP client (standard library)
- JSON handling (standard library)
- Tar/gzip extraction (standard library)
- Semantic version parsing (go-version)
- Cross-platform file operations (standard library)

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Registry unavailability | High | Local caching, fallback registries |
| Authentication failures | Medium | Clear error messages, token validation |
| File system permissions | Medium | Permission checks, user guidance |
| Network connectivity | Medium | Offline mode, cached operations |

## Timeline

- **Phase 1**: Core commands (install, uninstall, list)
- **Phase 2**: Configuration and registry support
- **Phase 3**: Update/outdated functionality
- **Phase 4**: Cache management and cleanup
- **Phase 5**: Testing and documentation