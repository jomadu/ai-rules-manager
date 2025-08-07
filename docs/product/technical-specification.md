# Technical Specification: AI Rules Manager (ARM)

## 1. Configuration System

### 1.1 Configuration File Hierarchy

**File Locations:**
- Global: `~/.arm/.armrc`
- Local: `./.armrc` (current directory only, no parent directory traversal)

**Precedence Rules:**
- Configuration files are merged at the key level
- Local configuration takes precedence over global for specific keys
- Missing keys in local configuration inherit from global configuration
- If neither file exists, ARM uses built-in defaults

**Merge Example:**
```ini
# Global ~/.arm/.armrc
[git]
concurrency=1
rateLimit=10

# Local ./.armrc
[git]
concurrency=5

# Effective configuration
[git]
concurrency=5    # from local
rateLimit=10     # from global
```

### 1.2 Configuration File Formats

**INI Format (.armrc):**
- Standard INI format with `#` and `;` comment support
- Section headers: `[section]`
- Key-value pairs: `key=value`
- Environment variable expansion supported: `$HOME`, `$USER`, etc.
- Schema validation with helpful error messages for malformed syntax

**JSON Format (arm.json, arm.lock):**
- Standard JSON format
- Schema validation with detailed error messages
- No comment support (standard JSON)
- No environment variable expansion (literal string values only)

### 1.3 Registry Configuration

**Supported URL Schemes:**
- `git://github.com/user/repo`
- `git://github.com/org/repo`
- `s3://bucket.region.amazonaws.com/`
- `gitlab://gitlab.example.com/project/123`
- `gitlab://gitlab.example.com/group/123`
- `http://example.com/registry`
- `file:///path/to/local/registry`

**Registry Configuration Structure:**
```ini
[registries]
default = git://github.com/user/default-registry
my-git-registry = git://github.com/user/repo
my-s3-registry = s3://my-bucket.us-east-1.amazonaws.com/
my-gitlab-registry = gitlab://gitlab.example.com/project/123

# Git registry configuration
[registries.my-git-registry]
authToken = $GITHUB_TOKEN
apiType = github
apiVersion = 2022-11-28
concurrency = 2
rateLimit = 10

# S3 registry configuration (uses AWS credentials chain)
[registries.my-s3-registry]
profile = my-aws-profile  # optional, uses default profile if omitted
prefix = /registries/path # optional
concurrency = 10
rateLimit = 100

# GitLab registry configuration
[registries.my-gitlab-registry]
authToken = $GITLAB_TOKEN
apiVersion = 4
concurrency = 2
rateLimit = 60
```

**Validation Rules:**
- URL format validation (defer connectivity checks to usage time)
- Registry name validation (must match entry in `[registries]` section)
- Registry-specific parameter validation based on URL scheme

**Built-in Defaults:**
```ini
[git]
concurrency=1
rateLimit=10

[s3]
concurrency=10
rateLimit=100

[gitlab]
concurrency=2
rateLimit=60

[cache]
path=~/.arm/cache
```

### 1.4 Channel Configuration

**Channel Structure (arm.json):**
```json
{
  "channels": {
    "cursor": {
      "directories": [".cursor/rules", ".custom/cursor"]
    },
    "q": {
      "directories": ["~/.aws/amazonq/rules"]
    }
  }
}
```

**Directory Handling:**
- Multiple directories per channel supported
- Literal directory paths only (no environment variable expansion)
- ARM creates directories if they don't exist
- Validates directory write permissions during configuration

### 1.5 Cache Configuration

**Cache Path:**
- Default: `~/.arm/cache`
- Configurable via `[cache] path=` in .armrc
- Environment variable expansion supported
- Write permission validation during configuration

**Stub File Templates:**

**.armrc Stub:**
```ini
# ARM Configuration File
# Configure registries and default settings

[registries]
# Default registry used when no source is specified
# default = git://github.com/user/registry

# Named registries
# my-git-registry = git://github.com/user/repo
# my-s3-registry = s3://bucket.region.amazonaws.com/
# my-gitlab-registry = gitlab://gitlab.example.com/project/123
# my-http-registry = http://example.com/registry
# my-local-registry = file:///path/to/local/registry

# Registry-specific configuration
# [registries.my-git-registry]
# authToken = $GITHUB_TOKEN
# apiType = github
# apiVersion = 2022-11-28
# concurrency = 2
# rateLimit = 10

# [registries.my-s3-registry]
# profile = my-aws-profile  # optional, uses default AWS profile if omitted
# prefix = /registries/path # optional prefix within bucket
# concurrency = 10
# rateLimit = 100

# [registries.my-gitlab-registry]
# authToken = $GITLAB_TOKEN
# apiVersion = 4
# concurrency = 2
# rateLimit = 60

# Defaults for registry types
# [git]
# concurrency=1
# rateLimit=10

# [s3]
# concurrency=10
# rateLimit=100

# [gitlab]
# concurrency=2
# rateLimit=60

# Cache configuration
# [cache]
# path=~/.arm/cache
```

**arm.json Stub:**
```json
{
  "engines": {
    "arm": "^1.2.3"
  },
  "channels": {},
  "rulesets": {}
}
```

## 2. Registry Types and Integration

### 2.1 Git Repository Registries

**URL Format:**
- `git://github.com/user/repo`
- `git://github.com/org/repo`
- `git://gitlab.example.com/user/repo`

**Operation Modes:**

**Git Operations Mode (Default):**
- Uses local git commands with user's existing git authentication
- Clones/fetches repository using git protocol
- Relies on user's configured git credentials (SSH keys, credential helpers)
- No additional authentication configuration required

**API Mode (Optional):**
- Uses platform-specific APIs (GitHub, GitLab) for file access
- Requires explicit authentication token configuration
- Faster for single file access, no full repository clone needed
- Configured via `apiType`, `apiVersion`, and `authToken` parameters

**Version Resolution:**
- `"latest"` → Tracks default branch (ARM detects main/master automatically)
- `"main"` → Tracks specific named branch
- `"^1.0.0"` → Tracks git tags matching semver pattern (supports both `v1.0.0` and `1.0.0` formats)
- `"53c5307"` → Specific commit hash (full or abbreviated)

**File Selection (matchingPatterns):**
- Applies glob patterns to entire repository file tree
- Supports multiple patterns per ruleset
- Uses standard glob syntax: `*`, `**`, `?`, `[abc]`, `{a,b,c}`
- Example patterns:
  - `"rules/*.md"` → All .md files in rules directory
  - `"**/*.cursorrules"` → All .cursorrules files recursively
  - `".cursor/rules/01-*.md"` → Specific numbered rule files

**Configuration Example:**
```ini
[registries]
my-git-registry = git://github.com/user/repo

[registries.my-git-registry]
authToken = $GITHUB_TOKEN  # optional, for API mode
apiType = github           # optional, enables API mode
apiVersion = 2022-11-28    # optional, API version
concurrency = 2
rateLimit = 10
```

### 2.2 AWS S3 Registries

**URL Format:**
- `s3://bucket.region.amazonaws.com/`
- `s3://my-bucket.us-east-1.amazonaws.com/registries/`

**Directory Structure:**
```
s3://bucket/prefix/
├── ruleset1/
│   ├── 1.0.0/
│   │   └── ruleset.tar.gz
│   └── 1.1.0/
│       └── ruleset.tar.gz
└── ruleset2/
    └── 2.0.0/
        └── ruleset.tar.gz
```

**Authentication:**
- Uses AWS credential chain (environment variables, profiles, IAM roles)
- Optional `profile` parameter to specify named AWS profile
- No explicit tokens required

**Version Discovery:**
- Lists S3 objects with prefix to discover available rulesets and versions
- Parses object keys to extract ruleset names and version numbers
- Supports semantic versioning for version resolution

**Region Detection:**
- Extracts AWS region from bucket URL (e.g., `us-east-1` from `s3://bucket.us-east-1.amazonaws.com/`)
- Falls back to AWS credential chain default region (AWS_DEFAULT_REGION, profile config, or us-east-1)
- Can be explicitly configured via `region` parameter in registry configuration

**Configuration Example:**
```ini
[registries]
my-s3-registry = s3://my-bucket.us-east-1.amazonaws.com/

[registries.my-s3-registry]
profile = my-aws-profile  # optional, uses default profile if omitted
region = us-west-2        # optional, overrides region from URL or credential chain
prefix = /registries/path # optional prefix within bucket
concurrency = 10
rateLimit = 100
```

### 2.3 GitLab Package Registries

**URL Format:**
- `gitlab://gitlab.example.com/project/123` → Project-level registry
- `gitlab://gitlab.example.com/group/456` → Group-level registry

**API Integration:**
- Uses GitLab Generic Packages API for simple file storage
- Stores rulesets as tar.gz files with version metadata
- Supports both project and group-level package registries

**Ruleset Naming Schema:**
- Package Name: `{ruleset-name}` (e.g., "my-rules") - uses GitLab's package terminology
- File Name: `ruleset.tar.gz` (fixed filename)
- API Path: `/projects/:id/packages/generic/{ruleset-name}/{version}/ruleset.tar.gz`
- Example: Ruleset "my-rules" version "1.0.0" → `my-rules/1.0.0/ruleset.tar.gz`

**Authentication:**
- Requires GitLab access token with appropriate package permissions
- Token configured via `authToken` parameter
- Supports environment variable expansion

**Version Management:**
- Uses GitLab's package versioning system
- Supports semantic versioning patterns
- Ruleset metadata stored in GitLab registry
- Consistent with S3/HTTP/Local registry structure

**Configuration Example:**
```ini
[registries]
my-gitlab-registry = gitlab://gitlab.example.com/project/123

[registries.my-gitlab-registry]
authToken = $GITLAB_TOKEN
apiVersion = 4  # optional, defaults to latest (v4)
concurrency = 2
rateLimit = 60
```

### 2.4 Generic HTTP Registries

**URL Format:**
- `http://example.com/registry`
- `https://my-registry.example.com/rulesets`

**Directory Structure:**
```
http://example.com/registry/
├── manifest.json          # contains all rulesets and versions
├── ruleset1/
│   ├── 1.0.0/
│   │   └── ruleset.tar.gz
│   └── 1.1.0/
│       └── ruleset.tar.gz
└── ruleset2/
    └── 2.0.0/
        └── ruleset.tar.gz
```

**Version Discovery:**
- Root-level `manifest.json` contains all available rulesets and versions
- Single HTTP request to discover entire registry contents
- Manifest format:
```json
{
  "rulesets": {
    "ruleset1": ["1.0.0", "1.1.0"],
    "ruleset2": ["2.0.0"]
  }
}
```

**Authentication:**
- Basic HTTP authentication supported
- Bearer token authentication supported
- Configured via standard HTTP authentication headers

**Configuration Example:**
```ini
[registries]
my-http-registry = http://example.com/registry

[registries.my-http-registry]
authToken = $HTTP_REGISTRY_TOKEN  # optional, for bearer auth
concurrency = 5
rateLimit = 50
```

### 2.5 Local File System Registries

**URL Format:**
- `file:///absolute/path/to/registry`
- `file://./relative/path/to/registry`

**Directory Structure:**
```
/path/to/registry/
├── ruleset1/
│   ├── 1.0.0/
│   │   └── ruleset.tar.gz
│   └── 1.1.0/
│       └── ruleset.tar.gz
└── ruleset2/
    └── 2.0.0/
        └── ruleset.tar.gz
```

**Version Discovery:**
- Uses filesystem directory listing to discover rulesets and versions
- Supports both absolute and relative paths
- No network requests required

**Ruleset Format:**
- Same tar.gz structure as other registry types
- Maintains consistency across all registry implementations

**Path Handling:**
- Supports both absolute paths (`file:///home/user/registry`)
- Supports relative paths (`file://./local-registry`)
- Path resolution relative to current working directory for relative paths

**Configuration Example:**
```ini
[registries]
my-local-registry = file:///path/to/local/registry

[registries.my-local-registry]
# No additional configuration required
# Inherits default concurrency and rate limit settings
```

## 3. Command Line Interface

### 3.1 Command Priority and Structure

**Command Priority (Help Order):**
1. `config` - Configuration management
2. `install` - Install rulesets
3. `uninstall` - Remove rulesets
4. `search` - Search for rulesets
5. `info` - Show ruleset information
6. `outdated` - Show outdated rulesets
7. `update` - Update rulesets
8. `clean` - Clean cache and unused rulesets
9. `list` - List installed rulesets
10. `version` - Show ARM version
11. `help` - Show help information

**Global Flags:**
- `--global` - Operate on global configuration (default: local)
- `--quiet` - Suppress non-essential output
- `--verbose` - Show detailed output
- `--dry-run` - Show what would be done without executing (destructive operations)
- `--json` - Output machine-readable JSON format
- `--no-color` - Disable colored output

**Exit Codes:**
- `0` - Success
- `1` - General error
- `2` - Usage error (invalid arguments, missing required parameters)

### 3.2 Config Command

**Syntax:**
```bash
arm config <subcommand> [options]
```

**Subcommands:**

**Set Configuration Value:**
```bash
arm config set <key> <value> [--global]
arm config set registries.default git://github.com/user/repo
arm config set git.concurrency 5
```

**Get Configuration Value:**
```bash
arm config get <key> [--global]
arm config get registries.default
arm config get git.concurrency
```

**List Configuration:**
```bash
arm config list [--global]
```
Shows merged configuration with source indicators:
```
[registries]
default = git://github.com/user/repo (local)
my-registry = s3://bucket.amazonaws.com/ (global)

[git]
concurrency = 5 (local)
rateLimit = 10 (global)
```

**Add Registry:**
```bash
arm config add registry <name> <url> [options] [--global]
arm config add registry my-git git://github.com/user/repo --authToken=$TOKEN --apiType=github
arm config add registry my-s3 s3://bucket.amazonaws.com/ --profile=myprofile --prefix=/rules
arm config add registry my-gitlab gitlab://gitlab.com/project/123 --authToken=$GITLAB_TOKEN
```

**Remove Registry:**
```bash
arm config remove registry <name> [--global]
arm config remove registry my-git
```

**Add Channel:**
```bash
arm config add channel <name> --directory dir1 --directory dir2 [--global]
arm config add channel cursor --directory .cursor/rules --directory .custom/cursor
```

**Remove Channel:**
```bash
arm config remove channel <name> [--global]
arm config remove channel cursor
```

**Validation:**
- URL format validation (immediate)
- Registry name validation (must not conflict with existing)
- No connectivity testing during configuration
- Creates directories if they don't exist
- Sets appropriate file permissions (600 for config files)

### 3.3 Install Command

**Syntax:**
```bash
arm install [ruleset-spec] [options]
```

**Behavior Scenarios:**

**No Arguments:**
```bash
arm install
```
- Check both scopes and proceed if config exists in either location
- If config exists with rulesets: Install all rulesets from manifest
- If config exists but empty rulesets: Show current status
- If no default registry and no named registries: Error with configuration guidance

**Install from Default Registry:**
```bash
arm install my-rules
arm install my-rules@^1.0.0
arm install my-rules@latest
```

**Install from Specific Registry:**
```bash
arm install my-registry/my-rules
arm install my-registry/my-rules@1.2.3
```

**Stub Generation Logic:**
- Check if `.armrc` exists in `./` OR `~/.arm/` → `.armrc` available = true/false
- Check if `arm.json` exists in `./` OR `~/.arm/` → `arm.json` available = true/false
- Generate stubs only for file types missing from both locations
- Target scope: `--global` flag generates in `~/.arm/`, otherwise in `./`
- Creates parent directories if needed (`~/.arm/` for global)
- Sets file permissions: 600 for config files
- No `--force` flag support for overwriting existing files

**Example Scenarios:**
- `.armrc` in `./`, `arm.json` in `~/.arm/` → No stubs needed, proceed
- `.armrc` in `~/.arm/`, no `arm.json` anywhere → Generate `arm.json` in target scope
- No `.armrc` anywhere, `arm.json` in `./` → Generate `.armrc` in target scope
- No files anywhere → Generate both files in target scope

**Options:**
- `--global` - Install to global configuration
- `--dry-run` - Show what would be installed
- `--channel channel1 --channel channel2` - Install to specific channels only (default: all channels)

### 3.4 Uninstall Command

**Syntax:**
```bash
arm uninstall <ruleset-name> [options]
arm uninstall my-rules
arm uninstall my-registry/my-rules
```

**Options:**
- `--global` - Uninstall from global configuration
- `--dry-run` - Show what would be removed
- `--channel channel1 --channel channel2` - Remove from specific channels only (default: all channels)

### 3.5 Search Command

**Syntax:**
```bash
arm search <query> [options]
```

**Registry Filtering:**
```bash
arm search "python rules"                    # Search all registries
arm search "python" --registry my-git        # Search specific registry
arm search "python" --registry my-git --registry my-s3  # Multiple registries
arm search "python" --registry "my-*"        # Glob pattern support
```

**Options:**
- `--registry <name>` - Search specific registry (repeatable)
- `--json` - JSON output format
- `--limit <n>` - Limit number of results

### 3.6 Info Command

**Syntax:**
```bash
arm info <ruleset-spec> [options]
arm info my-rules
arm info my-registry/my-rules@1.0.0
```

**Shows:**
- Ruleset description and metadata
- Available versions
- Installation status
- Registry information
- File contents preview

**Options:**
- `--json` - JSON output format
- `--versions` - Show all available versions

### 3.7 Outdated Command

**Syntax:**
```bash
arm outdated [options]
```

**Shows:**
- Currently installed version
- Latest available version
- Update command suggestion

**Options:**
- `--global` - Check global installations
- `--json` - JSON output format

### 3.8 Update Command

**Syntax:**
```bash
arm update [ruleset-name] [options]
arm update                    # Update all rulesets
arm update my-rules          # Update specific ruleset
```

**Options:**
- `--global` - Update global installations
- `--dry-run` - Show what would be updated

### 3.9 Clean Command

**Syntax:**
```bash
arm clean [target] [options]
```

**Targets:**
```bash
arm clean cache              # Clean download cache
arm clean unused             # Remove unused rulesets
arm clean all                # Clean cache and unused rulesets
```

**Options:**
- `--global` - Clean global installations
- `--dry-run` - Show what would be cleaned
- `--force` - Skip confirmation prompts

### 3.10 List Command

**Syntax:**
```bash
arm list [options]
```

**Shows:**
- Installed rulesets with versions
- Installation location (global/local)
- Channel assignments
- Update availability indicators

**Options:**
- `--global` - List global installations only
- `--local` - List local installations only
- `--json` - JSON output format
- `--channel channel1 --channel channel2` - Filter by specific channels (default: show all channels)

### 3.11 Version and Help Commands

**Version Command:**
```bash
arm version
arm --version
arm -v
```
Shows ARM version, Go version, and build information.

**Help Command:**
```bash
arm help
arm help <command>
arm <command> --help
arm <command> -h
```
Shows command usage, options, and examples in priority order.

## 4. Package Management

### 4.1 Version Specification Format
### 4.2 Dependency Resolution Algorithm
### 4.3 Lock File Management
### 4.4 Package Installation Process
### 4.5 File System Layout

## 5. Caching and Registry Resolution

### 5.1 Cache Structure
### 5.2 Registry Resolution Logic
### 5.3 Fallback Mechanisms
### 5.4 Cache Management

## 6. Error Handling and User Experience

### 6.1 Error Message Standards
### 6.2 Configuration Validation
### 6.3 Network Failure Handling
### 6.4 User Guidance Messages

## 7. File Management

### 7.1 Directory Structure
### 7.2 File Namespacing
### 7.3 Conflict Resolution
### 7.4 Permission Handling

## 8. Authentication and Security

### 8.1 Authentication Methods by Registry Type
### 8.2 Credential Management
### 8.3 Environment Variable Support
### 8.4 Security Considerations
