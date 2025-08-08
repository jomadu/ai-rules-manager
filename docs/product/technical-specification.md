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
rateLimit=10/minute

# Local ./.armrc
[git]
concurrency=5

# Effective configuration
[git]
concurrency=5         # from local
rateLimit=10/minute   # from global
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
- Environment variable expansion supported in string values: `$HOME`, `${USER}`, etc.
- Expansion processing: environment variables → tilde expansion → schema validation
- Missing environment variables resolve to empty strings
- Expanded values cached until file modification

### 1.3 Registry Configuration

**Supported Registry Types:**
- `git` - Git repositories (GitHub, GitLab, etc.) accessed via HTTPS
- `https` - Generic HTTP registries with manifest.json
- `s3` - AWS S3 bucket registries
- `gitlab` - GitLab package registries
- `local` - Local file system registries

**Registry Configuration Structure:**
```ini
[registries]
default = github.com/user/default-registry
my-git-registry = https://github.com/user/repo
my-s3-registry = my-bucket
my-gitlab-registry = https://gitlab.example.com/projects/123

# Required type configuration for all registries
[registries.default]
type = git

[registries.my-git-registry]
type = git
authToken = $GITHUB_TOKEN  # optional, for API mode
apiType = github           # optional, enables API mode
apiVersion = 2022-11-28    # optional, API version
concurrency = 5            # Override git default
rateLimit = 20/minute      # Override git default

[registries.my-s3-registry]
type = s3
region = us-east-1         # required for S3 registries
profile = my-aws-profile   # optional, uses default profile if omitted
prefix = /registries/path  # optional prefix within bucket

[registries.my-gitlab-registry]
type = gitlab
authToken = $GITLAB_TOKEN
apiVersion = 4

# Type-based defaults
[git]
concurrency = 1
rateLimit = 10/minute

[https]
concurrency = 5
rateLimit = 30/minute

[s3]
concurrency = 10
rateLimit = 100/hour

[gitlab]
concurrency = 2
rateLimit = 60/hour

[local]
concurrency = 20
rateLimit = 1000/second
```

**Validation Rules:**
- Registry name validation (must match entry in `[registries]` section)
- Type validation (must be one of: git, https, s3, gitlab, local)
- Required type parameter for all registries
- Type-specific parameter validation (e.g., region required for S3)

**Built-in Defaults:**
```ini
[git]
concurrency = 1
rateLimit = 10/minute

[https]
concurrency = 5
rateLimit = 30/minute

[s3]
concurrency = 10
rateLimit = 100/hour

[gitlab]
concurrency = 2
rateLimit = 60/hour

[local]
concurrency = 20
rateLimit = 1000/second

[network]
timeout = 30
retry.maxAttempts = 3
retry.backoffMultiplier = 2.0
retry.maxBackoff = 30

[cache]
path = ~/.arm/cache
maxSize = 1GB
ttl = 3600
```

### 1.4 Channel Configuration

**Channel Structure (arm.json):**
```json
{
  "channels": {
    "cursor": {
      "directories": ["$HOME/.cursor/rules", "${PROJECT_ROOT}/custom"]
    },
    "q": {
      "directories": ["~/.aws/amazonq/rules"]
    }
  }
}
```

**Directory Handling:**
- Multiple directories per channel supported
- Environment variable expansion supported (`$HOME`, `${PROJECT_ROOT}`)
- Tilde expansion supported (`~/.aws/amazonq/rules`)
- Both expansions can be used together
- ARM creates directories if they don't exist
- Validates directory write permissions during configuration

### 1.5 Ruleset Configuration

**Ruleset Structure (arm.json):**
```json
{
  "engines": {
    "arm": "^1.2.3"
  },
  "channels": {
    "cursor": {
      "directories": ["$HOME/.cursor/rules", "${PROJECT_ROOT}/custom"]
    }
  },
  "rulesets": {
    "default": {
      "my-rules": {
        "version": "^1.0.0",
        "patterns": ["rules/*.md", "**/*.mdc"]
      },
      "python-rules": {
        "version": "~2.1.0"
      }
    },
    "my-registry": {
      "custom-rules": {
        "version": "latest"
      }
    }
  }
}
```

**Registry Namespacing:**
- Rulesets are organized by registry name
- Registry names must match those configured in .armrc
- Each registry can contain multiple rulesets
- Ruleset names are unique within each registry namespace

**Environment Variable Support:**
- All string values support `$VAR` and `${VAR}` syntax
- Missing variables resolve to empty strings
- Processing order: environment expansion → tilde expansion → validation
- Examples:
  - `"$HOME/.cursor/rules"` → `"/Users/john/.cursor/rules"`
  - `"${PROJECT_ROOT}/custom"` → `"/workspace/my-project/custom"`
  - `"$MISSING_VAR/path"` → `"/path"` (empty string substitution)

### 1.6 Cache Configuration

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
# default = github.com/user/registry

# Named registries
# my-git-registry = https://github.com/user/repo
# my-s3-registry = my-bucket
# my-gitlab-registry = https://gitlab.example.com/projects/123
# my-https-registry = https://example.com/registry
# my-local-registry = /path/to/local/registry

# Required type configuration for all registries
# [registries.default]
# type = git

# [registries.my-git-registry]
# type = git
# authToken = $GITHUB_TOKEN  # optional, for API mode
# apiType = github           # optional, enables API mode
# apiVersion = 2022-11-28    # optional, API version

# [registries.my-s3-registry]
# type = s3
# region = us-east-1         # required for S3 registries
# profile = my-aws-profile   # optional, uses default AWS profile if omitted
# prefix = /registries/path  # optional prefix within bucket

# [registries.my-gitlab-registry]
# type = gitlab
# authToken = $GITLAB_TOKEN
# apiVersion = 4

# [registries.my-https-registry]
# type = https

# [registries.my-local-registry]
# type = local

# Type-based defaults (optional - ARM has built-in defaults)
# [git]
# concurrency = 1
# rateLimit = 10/minute

# [https]
# concurrency = 5
# rateLimit = 30/minute

# [s3]
# concurrency = 10
# rateLimit = 100/hour

# [gitlab]
# concurrency = 2
# rateLimit = 60/hour

# [local]
# concurrency = 20
# rateLimit = 1000/second

# Network configuration
# [network]
# timeout = 30
# retry.maxAttempts = 3
# retry.backoffMultiplier = 2.0
# retry.maxBackoff = 30

# Cache configuration
# [cache]
# path = ~/.arm/cache
# maxSize = 1GB
# ttl = 3600
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

**Registry Value Format:**
- `https://github.com/user/repo`
- `https://github.com/org/repo`
- `https://gitlab.example.com/user/repo`

**Type Configuration:**
- `type = git` (required)

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
  - `"**/*.mdc"` → All .mdc files recursively
  - `".cursor/rules/01-*.md"` → Specific numbered rule files

**Configuration Example:**
```ini
[registries]
my-git-registry = https://github.com/user/repo

[registries.my-git-registry]
type = git                 # required
authToken = $GITHUB_TOKEN  # optional, for API mode
apiType = github           # optional, enables API mode
apiVersion = 2022-11-28    # optional, API version
concurrency = 2            # override git defaults
rateLimit = 10/minute      # override git defaults
```

### 2.2 AWS S3 Registries

**Registry Value Format:**
- `my-bucket` (bucket name only)
- `my-bucket.with.dots` (literal bucket name)

**Type Configuration:**
- `type = s3` (required)
- `region = us-east-1` (required)

**Directory Structure:**
```
bucket/prefix/
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

**Region Configuration:**
- `region` parameter is required for all S3 registries
- Falls back to AWS credential chain default region if not specified
- No region extraction from bucket name (bucket name is literal)

**Configuration Example:**
```ini
[registries]
my-s3-registry = my-bucket

[registries.my-s3-registry]
type = s3                 # required
region = us-east-1        # required
profile = my-aws-profile  # optional, uses default profile if omitted
prefix = /registries/path # optional prefix within bucket
concurrency = 10          # override s3 defaults
rateLimit = 100/hour      # override s3 defaults
```

### 2.3 GitLab Package Registries

**Registry Value Format:**
- `https://gitlab.example.com/projects/123` → Project-level registry
- `https://gitlab.example.com/groups/456` → Group-level registry

**Type Configuration:**
- `type = gitlab` (required)

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
my-gitlab-registry = https://gitlab.example.com/projects/123

[registries.my-gitlab-registry]
type = gitlab             # required
authToken = $GITLAB_TOKEN # required for GitLab package registry access
apiVersion = 4            # optional, defaults to latest (v4)
concurrency = 2           # override gitlab defaults
rateLimit = 60/hour       # override gitlab defaults
```

### 2.4 Generic HTTP Registries

**Registry Value Format:**
- `https://example.com/registry`
- `https://my-registry.example.com/rulesets`

**Type Configuration:**
- `type = https` (required)

**Directory Structure:**
```
https://example.com/registry/
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
my-https-registry = https://example.com/registry

[registries.my-https-registry]
type = https                      # required
authToken = $HTTP_REGISTRY_TOKEN  # optional, for bearer auth
concurrency = 5                   # override https defaults
rateLimit = 30/minute             # override https defaults
```

### 2.5 Local File System Registries

**Registry Value Format:**
- `/absolute/path/to/registry`
- `./relative/path/to/registry`
- `relative/path/to/registry`

**Type Configuration:**
- `type = local` (required)

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
- Supports absolute paths (`/home/user/registry`)
- Supports relative paths (`./local-registry`, `local-registry`)
- Path resolution relative to current working directory for relative paths

**Configuration Example:**
```ini
[registries]
my-local-registry = /path/to/local/registry

[registries.my-local-registry]
type = local              # required
# Inherits local defaults: concurrency=20, rateLimit=1000/second
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
arm config set registries.default github.com/user/repo
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
default = github.com/user/repo (local)
my-registry = my-bucket (global)

[registries.default]
type = git (local)

[registries.my-registry]
type = s3 (global)
region = us-east-1 (global)

[git]
concurrency = 5 (local)
rateLimit = 10 (global)
```

**Add Registry:**
```bash
arm config add registry <name> <value> --type=<type> [options] [--global]
arm config add registry my-git https://github.com/user/repo --type=git --authToken=$TOKEN --apiType=github
arm config add registry my-s3 my-bucket --type=s3 --region=us-east-1 --profile=myprofile --prefix=/rules
arm config add registry my-gitlab https://gitlab.com/projects/123 --type=gitlab --authToken=$GITLAB_TOKEN
arm config add registry my-https https://example.com/registry --type=https --authToken=$TOKEN
arm config add registry my-local /path/to/registry --type=local
```

**Remove Registry:**
```bash
arm config remove registry <name> [--global]
arm config remove registry my-git
```

**Add Channel:**
```bash
arm config add channel <name> --directories dir1,dir2 [--global]
arm config add channel cursor --directories .cursor/rules,.custom/cursor
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

**Git Registry Pattern Handling:**
```bash
# Install new Git ruleset (requires patterns)
arm install awesome-cursorrules/rules-new-python --patterns "rules-new/python-*.mdc"

# Install with multiple patterns
arm install cursor-rules/base-devops --patterns ".cursor/rules/01-base-devops.mdc,docs/*.md"

# Override patterns for existing ruleset
arm install cursor-rules/base-agentic --patterns "different/*.mdc"
```

**Pattern Requirements:**
- **Git registries**: `--patterns` flag required for all Git rulesets (updates manifest with new patterns)
- **Non-Git registries**: Patterns not applicable (S3, GitLab, HTTP, Local use pre-packaged files)
- **Error handling**: Installing Git ruleset without `--patterns` flag results in error
- **Manifest updates**: Git rulesets automatically added/updated in local `arm.json` with specified patterns

**Options:**
- `--global` - Install to global configuration
- `--dry-run` - Show what would be installed
- `--channels channel1,channel2` - Install to specific channels only (default: all channels)
- `--patterns pattern1,pattern2` - Glob patterns for Git registry rulesets (comma-separated)

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
- `--channels channel1,channel2` - Remove from specific channels only (default: all channels)

### 3.5 Search Command

**Syntax:**
```bash
arm search <query> [options]
```

**Registry Filtering:**
```bash
arm search "python rules"                    # Search all registries
arm search "python" --registries my-git      # Search specific registry
arm search "python" --registries my-git,my-s3  # Multiple registries
arm search "python" --registries "my-*,yours-*"  # Glob pattern support
```

**Options:**
- `--registries <names>` - Search specific registries (comma-separated)
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
- `--channels channel1,channel2` - Filter by specific channels (default: show all channels)

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

**Supported Version Formats:**

**Semantic Version Ranges:**
- `^1.0.0` - Compatible version (>=1.0.0 <2.0.0)
- `~1.2.0` - Patch releases only (>=1.2.0 <1.3.0)
- `>=1.1.0` - Greater than or equal
- `<=2.0.0` - Less than or equal
- `>1.0.0` - Greater than
- `<2.0.0` - Less than
- `=1.2.3` - Exact version match

**Git-Specific Versions:**
- `latest` - HEAD of default branch (resolved and locked)
- `main` - HEAD of named branch (resolved and locked)
- `develop` - HEAD of named branch (resolved and locked)
- `abc123def` - Specific commit hash

**Version Resolution Priority:**
- Highest satisfying version within range
- Example: `^1.0.0` with available `[1.0.0, 1.2.0, 2.0.0]` resolves to `1.2.0`
- Example: `>=1.1.0` with available `[1.0.0, 1.2.0, 2.0.0]` resolves to `2.0.0`

**Pre-release Versions:**
- Not supported in MVP (e.g., `1.0.0-rc.1`)

### 4.2 Version Resolution

**No Inter-Ruleset Dependencies:**
- Rulesets cannot depend on other rulesets
- No circular dependency resolution needed
- Each ruleset is resolved independently

**Resolution Process:**
1. Parse version specification from `arm.json`
2. Query registry for available versions
3. Apply version range logic to find highest satisfying version
4. Lock resolved version in `arm.lock`
5. Cache resolution until explicit update command

### 4.3 Lock File Management

**Lock File Structure (`arm.lock`):**
```json
{
  "rulesets": {
    "default": {
      "my-rules": {
        "version": "1.2.0",
        "resolved": "2024-01-15T10:30:00Z",
        "registry": "my-bucket",
        "type": "s3",
        "region": "us-east-1"
      }
    },
    "my-git": {
      "python-rules": {
        "version": "abc123def",
        "resolved": "2024-01-15T10:30:00Z",
        "registry": "https://github.com/user/repo",
        "type": "git"
      }
    }
  }
}
```

**Lock File Behavior:**
- **Updates**: Modified on any `install`, `update`, or `uninstall` command
- **Conflict Resolution**: `arm.json` changes require explicit `arm install` or error with manual resolution
- **Registry Grouping**: Mirrors installation directory structure to avoid naming collisions
- **Metadata**: Includes resolution timestamp and registry source for debugging
- **No Checksums**: Simplified for MVP, relies on registry integrity

### 4.4 Package Installation Process

**Parallel Processing:**
- Install multiple rulesets concurrently based on registry `concurrency` settings
- Apply `rateLimit` per registry instance to respect API limits
- Queue additional rulesets when concurrency limit reached

**Installation Flow:**
1. **Version Resolution**: Resolve version specs to concrete versions
2. **Download**: Fetch ruleset from registry (tar.gz for packaged, git clone for Git)
3. **Extract**: Extract files according to patterns (Git) or tar contents (packaged)
4. **Install**: Copy files to channel directories
5. **Update Lock**: Record successful installation in `arm.lock`
6. **Cleanup**: Remove previous version (keep 1 previous version for rollback)

**Failure Handling:**
- **Per-Ruleset Rollback**: Failed installations don't affect successful ones
- **Continue on Failure**: Process remaining rulesets even if some fail
- **Lock File Accuracy**: Only record successfully installed rulesets
- **Partial Cleanup**: Clean up failed installation artifacts

**Progress Indication:**
- **Aggregate Progress**: Show overall percentage complete
- **Current Operation**: Display current step (downloading, extracting, installing)
- **No Persistence**: Progress resets on command interruption (Ctrl+C)

**File Conflict Resolution:**
- **Same Name, Different Rulesets**: Error and require manual resolution
- **Existing Non-ARM Files**: Overwrite without warning
- **ARM-Managed Files**: Replace with new version

### 4.5 File System Layout

**Installation Directory Structure:**
```
.cursor/rules/
├── arm/                        # ARM-managed rulesets
│   ├── registry-name/
│   │   └── ruleset-name/
│   │       ├── current-version/
│   │       │   ├── file1.md
│   │       │   └── file2.mdc
│   │       └── previous-version/
│   │           ├── file1.md
│   │           └── file2.mdc
│   └── another-registry/
│       └── another-ruleset/
│           └── version/
│               └── files...
└── user-file.md                # User-managed files
```

**Version Management:**
- **Current Version**: Active version used by channels
- **Previous Version**: Kept for rollback capability until next successful installation
- **Version Cleanup**: Remove previous version only after new version successfully installs
- **Directory Names**: Use exact version strings (e.g., `1.2.0`, `abc123def`, `main`)

**Channel Deployment:**
- Files are copied (not symlinked) from version directories to channel directories under `arm/` folder
- Each channel maintains its own copy of ruleset files in `{channel-dir}/arm/registry/ruleset/version/` structure
- Channel directories specified in `arm.json` channels configuration
- ARM automatically creates `arm/` subdirectory within each configured channel directory

**Cache Directory Structure:**
```
~/.arm/cache/
├── registries/
│   ├── registry-name/
│   │   ├── repository/          # Git repository clones
│   │   ├── rulesets/
│   │   │   └── ruleset-name/
│   │   │       └── version/
│   │   │           └── ruleset.tar.gz
│   │   ├── metadata.json        # Registry metadata cache
│   │   ├── versions.json        # Available versions cache
│   │   └── cache-info.json      # Cache timestamps and TTL
│   └── another-registry/
└── temp/                        # Temporary extraction directories
```

**Permissions:**
- **Configuration Files**: 600 (user read/write only)
- **Ruleset Files**: 644 (user read/write, group/other read)
- **Directories**: 755 (standard directory permissions)

## 5. Caching and Registry Resolution

### 5.1 Cache Structure

**Cache Directory Layout:**
```
~/.arm/cache/
├── registries/
│   ├── registry-name/
│   │   ├── repository/          # Git repository clones
│   │   ├── rulesets/
│   │   │   └── ruleset-name/
│   │   │       └── version/
│   │   │           └── ruleset.tar.gz
│   │   ├── metadata.json        # Registry metadata cache
│   │   ├── versions.json        # Available versions cache
│   │   └── cache-info.json      # Cache timestamps and TTL
│   └── another-registry/
└── temp/                        # Temporary extraction directories
```

**Cache File Structure:**

**versions.json:**
```json
{
  "cached_at": "2024-01-15T10:30:00Z",
  "ttl_seconds": 3600,
  "rulesets": {
    "my-rules": ["1.0.0", "1.1.0", "1.2.0"],
    "python-rules": ["2.0.0", "2.1.0"]
  }
}
```

**metadata.json:**
```json
{
  "cached_at": "2024-01-15T10:30:00Z",
  "ttl_seconds": 3600,
  "rulesets": {
    "my-rules": {
      "description": "Python coding rules",
      "latest_version": "1.2.0"
    }
  }
}
```

**cache-info.json:**
```json
{
  "registry_url": "s3://bucket.amazonaws.com/",
  "last_accessed": "2024-01-15T10:30:00Z",
  "total_size_bytes": 1048576
}
```

### 5.2 Registry Resolution Logic

**Ruleset Resolution:**
- **No Registry Specified**: Use default registry only
- **Registry Specified**: Use exact registry match (e.g., `my-git/python-rules`)
- **No Collisions**: Each ruleset belongs to exactly one registry
- **Resolution Failure**: Fail immediately if specified registry is unreachable

**Search Resolution:**
- **Parallel Queries**: Query all configured registries simultaneously
- **Registry Filtering**: Support `--registries` flag with glob patterns
- **Result Aggregation**: Combine results from all queried registries
- **Timeout Handling**: Continue with partial results if some registries timeout

**Version Resolution Process:**
1. **Check Lock File**: Use locked version if available and not updating
2. **Check Cache**: Use cached version list if within TTL
3. **Network Query**: Fetch fresh version list if cache expired or missing
4. **Apply Constraints**: Filter versions based on specification (^, ~, >=, etc.)
5. **Select Version**: Choose highest satisfying version
6. **Update Cache**: Store fresh version data with timestamp

### 5.3 Fallback Mechanisms

**Network Failure Handling:**
- **Fresh Cache Available**: Use cached data if within TTL (1 hour)
- **Stale Cache Available**: Fail with "network required" error
- **No Cache Available**: Fail with "network required" error
- **Registry Unreachable**: Fail immediately, no fallback to other registries

**Git Repository Issues:**
- **Merge Conflicts**: Remove repository and re-clone from scratch
- **Corrupted Repository**: Remove repository and re-clone from scratch
- **Authentication Failures**: Fail immediately with clear error message
- **Branch/Tag Missing**: Fail immediately with version not found error

**Cache Corruption:**
- **Invalid JSON**: Remove corrupted cache file and re-fetch
- **Missing Files**: Treat as cache miss and re-fetch
- **Disk Space Issues**: Clean LRU cache entries and retry

### 5.4 Cache Management

**Cache TTL (Time To Live):**
- **Default TTL**: 1 hour (3600 seconds) for all cached data
- **Version Lists**: 1 hour TTL
- **Registry Metadata**: 1 hour TTL
- **Downloaded Rulesets**: Cleaned up immediately after extraction
- **Git Repositories**: Persistent, updated with `git pull`

**Cache Invalidation:**
- **arm update**: Invalidate version caches to find latest versions
- **arm install**: Use cached data if available and within TTL
- **arm search**: Use cached metadata if available and within TTL
- **Manual Invalidation**: `arm clean cache` removes all cached data

**LRU Eviction:**
- **Default Size Limit**: 1GB total cache size
- **Configurable**: Set via `arm config set cache.maxSize 2GB`
- **Eviction Strategy**: Remove least recently accessed registry caches
- **Eviction Order**:
  1. Downloaded tar.gz files (oldest first)
  2. Git repositories (least recently accessed)
  3. Metadata caches (least recently accessed)
- **Protected Items**: Currently installing rulesets are never evicted

**Cache Operations:**

**Automatic Cleanup:**
- **Post-Installation**: Remove downloaded tar.gz files after extraction
- **Size Monitoring**: Check cache size after each download
- **LRU Eviction**: Remove old entries when size limit exceeded

**Manual Cleanup:**
```bash
arm clean cache              # Remove all cached data
arm clean unused             # Remove unused cached rulesets
arm clean all                # Clean cache and unused rulesets
```

**Cache Configuration:**
```bash
arm config set cache.maxSize 2GB        # Set cache size limit
arm config set cache.ttl 7200           # Set TTL to 2 hours
arm config get cache.maxSize             # View current cache limit
```

**Offline Behavior:**
- **Fresh Cache**: Operations succeed using cached data
- **Stale Cache**: Operations fail with "Network connectivity required" error
- **No Cache**: Operations fail with "Network connectivity required" error
- **Lock File Present**: Install operations use locked versions if cached

**Git Repository Management:**
- **Initial Clone**: Full repository clone to cache directory
- **Updates**: `git pull` to update existing repository
- **Conflict Resolution**: Remove and re-clone on any git conflicts
- **Branch Switching**: `git checkout` to switch between branches/tags
- **Storage**: Repositories persist until manually cleaned or LRU evicted

## 6. Error Handling and User Experience

### 6.1 Error Message Standards

**Structured Error Format:**
```
Error [CATEGORY]: Primary error message
Details: Additional context or technical details
Suggestion: Recommended corrective action
```

**Error Categories:**
- `[NETWORK]` - Network connectivity, timeouts, DNS resolution
- `[AUTH]` - Authentication failures, invalid credentials
- `[CONFIG]` - Configuration file errors, invalid settings
- `[REGISTRY]` - Registry-specific errors, unavailable registries
- `[RULESET]` - Ruleset not found, version conflicts
- `[FILESYSTEM]` - File permissions, disk space, path issues
- `[VALIDATION]` - Input validation, malformed data
- `[DEPENDENCY]` - Missing system dependencies (git, tar)

**Example Error Messages:**
```bash
# Network Error
Error [NETWORK]: Failed to connect to registry 's3://bucket.amazonaws.com/'
Details: Connection timeout after 30 seconds
Suggestion: Check your internet connection and registry URL

# Configuration Error
Error [CONFIG]: Invalid configuration in .armrc
Details: Line 5: Unknown registry type 'invalid'
Suggestion: Valid registry types are: git, s3, gitlab, http, local

# Dependency Error
Error [DEPENDENCY]: Required tool 'git' not found in PATH
Details: Git is required for Git registry operations
Suggestion: Install git using your package manager (e.g., 'brew install git')
```

**Exit Codes:**
- `0` - Success
- `1` - General error (config, validation, user input)
- `2` - System error (network, filesystem, dependencies)

**Verbosity Levels:**

**Default Output:**
```bash
$ arm install my-rules
Installing my-rules@1.2.0...
✓ Downloaded my-rules@1.2.0
✓ Installed to .cursor/rules/
```

**Quiet Mode (--quiet):**
```bash
$ arm install my-rules --quiet
# No output on success, only critical errors
```

**Verbose Mode (--verbose):**
```bash
$ arm install my-rules --verbose
[DEBUG] Loading configuration from .armrc
[DEBUG] Cache hit: versions.json (fresh)
[DEBUG] Resolving version ^1.0.0 -> 1.2.0
[DEBUG] Downloading from s3://bucket/my-rules-1.2.0.tar.gz
[DEBUG] HTTP GET 200 (1.2MB in 0.8s)
[DEBUG] Extracting to /tmp/arm-extract-abc123
[DEBUG] Copying 3 files to .cursor/rules/default/my-rules/1.2.0/
[DEBUG] Updating arm.lock
✓ Installed my-rules@1.2.0
```

### 6.2 Configuration Validation

**Validation Timing:**
- **On Access**: Validate configuration files only when accessed or modified
- **Not Every Command**: Skip validation for commands that don't need config
- **Lazy Loading**: Load and validate config sections as needed

**Validation Process:**
1. **Syntax Check**: Verify INI/JSON syntax is valid
2. **Schema Validation**: Check required fields and data types
3. **Semantic Validation**: Verify registry URLs, paths exist
4. **Dependency Check**: Ensure required tools are available

**Invalid Configuration Behavior:**
- **Corrupted Files**: Fail completely with clear error message
- **Missing Files**: Create stub files with default values
- **Invalid Values**: Fail with specific field-level errors
- **No Fallback**: Never silently use defaults for invalid config

**Configuration Error Examples:**
```bash
# Syntax Error
Error [CONFIG]: Invalid INI syntax in ~/.armrc
Details: Line 12: Expected '=' after key 'url'
Suggestion: Check INI file syntax

# Missing Required Field
Error [CONFIG]: Missing required field in registry 'my-s3'
Details: Field 'type' is required for all registries
Suggestion: Add 'type = s3' to [registries.my-s3] section

# Missing S3 Region
Error [CONFIG]: Missing required field in registry 'my-s3'
Details: Field 'region' is required for S3 registries
Suggestion: Add 'region = us-east-1' to [registries.my-s3] section

# Invalid Registry Type
Error [CONFIG]: Unknown registry type 'ftp' in registry 'my-ftp'
Details: Supported types: git, https, s3, gitlab, local
Suggestion: Change type to one of the supported registry types
```

**Dependency Validation:**
```bash
# Missing Git
Error [DEPENDENCY]: Git not found in system PATH
Details: Git is required for Git registry operations
Suggestion: Install git:
  macOS: brew install git
  Ubuntu: sudo apt install git
  Windows: Download from https://git-scm.com/

# Missing Tar
Error [DEPENDENCY]: Tar not found in system PATH
Details: Tar is required for extracting ruleset archives
Suggestion: Install tar using your system package manager
```

### 6.3 Network Failure Handling

**Retry Configuration:**
```ini
# .armrc configuration
[network]
timeout = 30                    # Global timeout in seconds
retry.maxAttempts = 3          # Maximum retry attempts
retry.backoffMultiplier = 2.0  # Exponential backoff multiplier
retry.maxBackoff = 30          # Maximum backoff time in seconds
```

**Retry Logic:**
- **Exponential Backoff**: 1s, 2s, 4s, 8s, 16s, 30s (capped)
- **Maximum Wait**: 30 seconds total backoff time
- **Retryable Errors**: Network timeouts, temporary DNS failures, 5xx HTTP errors
- **Non-Retryable**: 4xx HTTP errors, authentication failures, malformed URLs

**Network Error Handling:**

**Connection Timeout:**
```bash
Error [NETWORK]: Connection timeout to registry 'my-s3'
Details: No response after 30 seconds
Suggestion: Check internet connection or increase timeout with:
  arm config set network.timeout 60
```

**DNS Resolution:**
```bash
Error [NETWORK]: Failed to resolve hostname for registry 'my-git'
Details: DNS lookup failed for https://github.com/user/repo
Suggestion: Verify registry configuration is correct in .armrc
```

**HTTP Errors:**
```bash
# 404 Not Found
Error [REGISTRY]: Ruleset 'my-rules@1.5.0' not found
Details: HTTP 404 from S3 bucket 'my-bucket' in region 'us-east-1'
Suggestion: Check available versions with 'arm info my-rules --versions'

# 403 Forbidden
Error [AUTH]: Access denied to registry 'my-s3'
Details: HTTP 403 - insufficient permissions for S3 bucket 'my-bucket'
Suggestion: Check AWS credentials or S3 bucket permissions
```

**Partial Failure Handling:**
```bash
$ arm install ruleset1 ruleset2 ruleset3
✓ Installed ruleset1@1.0.0
✗ Failed to install ruleset2: [NETWORK] Connection timeout
✓ Installed ruleset3@2.1.0

Warning: 1 of 3 rulesets failed to install
Run 'arm install ruleset2' to retry failed installation
```

### 6.4 User Guidance Messages

**Command Suggestions:**

**Typo Detection (Fuzzy Matching):**
```bash
$ arm instal my-rules
Error [VALIDATION]: Unknown command 'instal'
Did you mean: install

$ arm install my-ruls
Error [RULESET]: Ruleset 'my-ruls' not found
Did you mean: my-rules
Suggestion: Run 'arm search my-ruls' to find similar rulesets
```

**Missing Arguments:**
```bash
$ arm install
Error [VALIDATION]: Missing required argument <ruleset-name>
Usage: arm install <ruleset-name> [options]
Example: arm install my-rules

$ arm config add registry
Error [VALIDATION]: Missing required arguments
Usage: arm config add registry <name> <value> --type=<type> [options]
Example: arm config add registry my-git https://github.com/user/repo --type=git
```

**Helpful Context:**
```bash
# No registries configured
$ arm search python
Error [CONFIG]: No registries configured
Suggestion: Add a registry first:
  arm config add registry default s3://your-bucket/
  arm config add registry my-git git://github.com/user/repo

# No rulesets installed
$ arm list
No rulesets installed
Suggestion: Install rulesets with 'arm install <ruleset-name>'
           Search available rulesets with 'arm search <query>'
```

**Progress Interruption (Ctrl+C):**
```bash
$ arm install large-ruleset
Downloading large-ruleset@1.0.0... 45%
^C
Interrupted by user
Cleaning up partial downloads...
✓ Cleanup complete

To resume: arm install large-ruleset
```

**Update Suggestions:**
```bash
$ arm outdated
Outdated rulesets found:
  my-rules: 1.0.0 → 1.2.0 (2 versions behind)
  python-rules: 2.1.0 → 2.3.1 (2 versions behind)

Run 'arm update' to update all rulesets
Run 'arm update my-rules' to update specific ruleset
```

**Configuration Guidance:**
```bash
# First time setup
$ arm install my-rules
Error [CONFIG]: No configuration found
Suggestion: Initialize ARM configuration:
  arm config add registry default my-bucket --type=s3 --region=us-east-1

# Missing patterns for Git registry
$ arm install git-registry/my-rules
Error [VALIDATION]: Git registry rulesets require --patterns flag
Example: arm install git-registry/my-rules --patterns "*.md,*.mdc"
Suggestion: Specify file patterns to extract from the Git repository

# Missing type configuration
$ arm config add registry my-new https://github.com/user/repo
Error [VALIDATION]: Missing required --type flag
Example: arm config add registry my-new https://github.com/user/repo --type=git
Suggestion: All registries require explicit type configuration
```

**JSON Output for Scripting:**
```bash
$ arm install my-rules --json
{
  "success": false,
  "error": {
    "category": "NETWORK",
    "message": "Connection timeout to registry 'my-s3'",
    "details": "No response after 30 seconds from S3 bucket 'my-bucket' in region 'us-east-1'",
    "suggestion": "Check internet connection or increase timeout"
  },
  "exit_code": 2
}
```

## 7. File Management

### 7.1 Directory Structure

**Channel Directory Layout:**
```
.cursor/rules/
├── arm/                        # ARM-managed rulesets
│   ├── registry-name/
│   │   └── ruleset-name/
│   │       └── 1.2.0/          # Version directory (actual version string)
│   │           ├── file1.md
│   │           ├── file2.mdc
│   │           └── subdir/
│   │               └── file3.txt
│   └── another-registry/
│       └── another-ruleset/
│           └── 2.1.0/          # Version directory (actual version string)
│               └── rules.md
└── user-file.md                # User-managed files
```

**Directory Creation:**
- **Automatic Creation**: ARM creates `arm/` parent directory and registry/ruleset/version subdirectories automatically
- **Path Structure**: `arm/registry/ruleset/version/` structure for clear separation from user files
- **Version Directories**: Use actual version strings (e.g., `1.2.0`, `abc123def`, `main`) as directory names
- **Nested Directories**: Preserve subdirectory structure from source rulesets within version directories
- **User Separation**: All ARM-managed content lives under `arm/` folder, leaving root directory available for user files
- **Backward Compatibility**: Existing installations remain in their current locations; only new installations use the `arm/` structure

**Installation Directory Mapping:**
```
# Source structure in ruleset
ruleset.tar.gz:
├── python-rules.md
├── javascript-rules.mdc
└── advanced/
    └── security.txt

# Installed to directory
.cursor/rules/arm/my-registry/python-rules/1.2.0/
├── python-rules.md
├── javascript-rules.mdc
└── advanced/
    └── security.txt
```

### 7.2 File Namespacing

**Namespace Strategy:**
- **Directory-Based**: Rely on registry/ruleset directory structure for namespacing
- **No File Renaming**: Preserve original filenames from rulesets
- **Natural Separation**: Different rulesets cannot conflict due to directory isolation
- **Subdirectory Preservation**: Maintain internal directory structure of rulesets

**File Extension Filtering:**
- **Allowed Extensions**: `.md`, `.mdc`, `.txt`
- **Rejected Extensions**: All other file types (`.json`, `.yaml`, `.py`, `.js`, etc.)
- **Installation Behavior**: Skip unsupported files with notification
- **Case Insensitive**: Extensions matched case-insensitively (`.MD`, `.Txt` allowed)

**File Extension Handling:**
```bash
# During installation
$ arm install my-rules --verbose
[DEBUG] Processing ruleset files:
✓ Installed: python-rules.md
✓ Installed: config.mdc
✓ Installed: readme.txt
⚠ Skipped: config.json (unsupported extension)
⚠ Skipped: script.py (unsupported extension)
⚠ Skipped: binary-file (no extension)

Installed 3 files, skipped 3 unsupported files
```

**Namespace Examples:**
```
# Multiple rulesets with same filename - no conflict
.cursor/rules/arm/
├── registry-a/
│   └── python-rules/
│       └── 1.0.0/
│           └── main.md      # From registry-a/python-rules
└── registry-b/
    └── python-rules/
        └── 2.1.0/
            └── main.md      # From registry-b/python-rules (different content)
```

### 7.3 Conflict Resolution

**No File-Level Conflicts:**
- **Directory Isolation**: Each ruleset installs to its own directory path
- **Registry Namespacing**: Registry names prevent cross-registry conflicts
- **Ruleset Namespacing**: Ruleset names prevent same-registry conflicts
- **Channel Independence**: Same ruleset in different channels = independent copies

**Conflict Scenarios and Resolution:**

**Same Ruleset, Multiple Channels:**
```bash
# Installing to multiple channels creates independent copies
$ arm install my-rules --channels channel1,channel2

# Results in files copied to each channel's directories:
# Channel 1: /path/to/channel1/default/my-rules/file.md
# Channel 2: /path/to/channel2/default/my-rules/file.md
# ^ Independent files, can diverge over time
```

**Existing Non-ARM Files:**
```bash
# ARM overwrites existing files in deployment directories
$ ls /path/to/channel/default/my-rules/
manual-file.md  # User-created file

$ arm install my-rules
# If my-rules contains manual-file.md, it will be overwritten
# No warning given - ARM assumes ownership of deployment directories
```

**ARM-Managed File Updates:**
```bash
# Updating rulesets replaces existing ARM-managed files
$ arm update my-rules
# All deployed files are replaced with new version
# Previous version kept in installation directory until next update
```

**Directory Conflicts:**
- **Deployment Directory Exists**: ARM uses existing directory, creates subdirectories as needed
- **Registry Directory Exists**: ARM uses existing directory, adds ruleset subdirectories
- **Ruleset Directory Exists**: ARM replaces contents during installation/update

### 7.4 Permission Handling

**Standardized Permissions:**
- **Files**: 644 (user read/write, group/other read)
- **Directories**: 755 (user read/write/execute, group/other read/execute)
- **No Inheritance**: Ignore source permissions from tar.gz or git repositories
- **Consistent Security**: All ARM-managed files use same permission model

**Permission Application:**
```bash
# During installation, ARM sets standard permissions
$ arm install my-rules
$ ls -la .cursor/rules/default/my-rules/
drwxr-xr-x  3 user group   96 Jan 15 10:30 .
drwxr-xr-x  3 user group   96 Jan 15 10:30 ..
-rw-r--r--  1 user group 1024 Jan 15 10:30 rules.md
-rw-r--r--  1 user group  512 Jan 15 10:30 config.mdc
drwxr-xr-x  2 user group   64 Jan 15 10:30 advanced/
```

**Permission Edge Cases:**
- **Existing Files**: ARM changes permissions to standard values during installation
- **User Modifications**: User permission changes are overwritten on next install/update
- **System Restrictions**: If ARM cannot set permissions, installation fails with clear error
- **Readonly Filesystems**: ARM detects and fails gracefully with appropriate error message

**Directory Creation Process:**
1. **Check Parent Path**: Verify channel directory path is valid
2. **Create Missing Directories**: Create channel, registry, and ruleset directories as needed
3. **Set Permissions**: Apply 755 permissions to all created directories
4. **Verify Access**: Ensure ARM can write to target directories
5. **Fail Gracefully**: Clear error messages if directory creation fails

**File Installation Process:**
1. **Filter Extensions**: Skip files with unsupported extensions
2. **Create Target Path**: Ensure target directory exists
3. **Copy File**: Copy from source to target location
4. **Set Permissions**: Apply 644 permissions to installed file
5. **Preserve Structure**: Maintain subdirectory hierarchy from source

**Permission Error Handling:**
```bash
# Permission denied during installation
Error [FILESYSTEM]: Cannot create directory '.cursor/rules/default/'
Details: Permission denied (errno 13)
Suggestion: Check directory permissions or run with appropriate privileges

# Readonly filesystem
Error [FILESYSTEM]: Cannot write to readonly filesystem
Details: Target path '.cursor/rules/' is on readonly mount
Suggestion: Choose a writable location or remount filesystem with write access
```

## 8. Authentication and Security

### 8.1 Authentication Methods by Registry Type

**Git Registry Authentication:**
- **Credential Delegation**: Use existing user Git credentials (SSH keys, tokens)
- **No ARM-Specific Config**: ARM relies on system Git configuration
- **SSH Key Support**: Automatic use of SSH keys configured in ~/.ssh/
- **HTTPS Token Support**: Use tokens configured in Git credential manager
- **Authentication Flow**: Git operations use standard Git authentication mechanisms

**Git Authentication Examples:**
```bash
# ARM uses existing Git credentials automatically
$ git config --global credential.helper store
$ git clone https://github.com/user/repo  # Configure credentials
$ arm install git-registry/my-rules       # Uses same credentials
```

**S3 Registry Authentication:**
- **AWS Credential Chain**: Use standard AWS credential resolution order
- **Credential Sources**: AWS CLI, environment variables, IAM roles, instance profiles
- **No Explicit Keys**: ARM does not store AWS access keys directly
- **Region Detection**: Automatic region detection from bucket URL or AWS config

**AWS Credential Chain Order:**
1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
2. AWS credentials file (~/.aws/credentials)
3. AWS config file (~/.aws/config)
4. IAM roles for EC2 instances
5. IAM roles for ECS tasks
6. IAM roles for Lambda functions

**S3 Authentication Examples:**
```bash
# ARM uses AWS credential chain automatically
$ aws configure  # Set up AWS credentials
$ arm install my-s3-rules  # Uses AWS credentials
```

**GitLab Registry Authentication:**
- **Personal Access Tokens**: Store tokens in .armrc configuration
- **Token Permissions**: Requires 'read_api' and 'read_repository' scopes
- **Environment Variable Expansion**: Support ${GITLAB_TOKEN} in configuration
- **Per-Registry Tokens**: Different tokens for different GitLab instances

**GitLab Configuration:**
```ini
# .armrc configuration
[registries.my-gitlab]
authToken = $GITLAB_TOKEN
```

**HTTP Registry Authentication:**
- **Bearer Tokens**: Support Authorization: Bearer <token> headers
- **Token Storage**: Store tokens in .armrc configuration files
- **Environment Variable Expansion**: Support ${API_TOKEN} syntax
- **Per-Registry Authentication**: Independent authentication per HTTP registry

**HTTP Authentication Configuration:**
```ini
# .armrc configuration
[registries.my-http]
authToken = $API_TOKEN
```

**Local Registry Authentication:**
- **File System Permissions**: Rely on standard file system access controls
- **No Authentication**: Local registries require no additional authentication
- **Path Validation**: Ensure ARM has read access to specified directories

### 8.2 Credential Storage and Management

**Storage Location:**
- **Configuration Files**: Store credentials in .armrc files
- **File System Security**: Rely on 600 permissions for credential protection
- **No Encryption**: Credentials stored in plain text (secured by file permissions)
- **Environment Variables**: Support expansion of environment variables in config

**Environment Variable Expansion:**
- **Supported Syntax**: Both `${VAR}` and `$VAR` formats
- **Full Expansion**: Expand all environment variables, not just credential-related
- **Expansion Timing**: Variables expanded at runtime, not config load time
- **Missing Variables**: Treat missing environment variables as empty strings

**Environment Variable Examples:**
```ini
# .armrc with environment variable expansion
[registries.gitlab-prod]
authToken = $GITLAB_PROD_TOKEN

[registries.gitlab-dev]
authToken = $GITLAB_DEV_TOKEN

[registries.http-api]
authToken = $API_TOKEN_PREFIX_$ENVIRONMENT
```

**Credential Security:**
- **File Permissions**: .armrc files must have 600 permissions (user read/write only)
- **Permission Enforcement**: ARM validates and corrects file permissions on access
- **No Network Transmission**: Credentials never logged or transmitted in plain text
- **Memory Handling**: Clear credentials from memory after use

**Permission Validation:**
```bash
# ARM automatically fixes insecure permissions
$ ls -la ~/.armrc
-rw-rw-rw- 1 user group 256 Jan 15 10:30 .armrc

$ arm install my-rules
Warning: Fixed insecure permissions on ~/.armrc (was 666, now 600)
```

### 8.3 Transport Security

**HTTPS Enforcement:**
- **Required by Default**: All registry connections must use HTTPS
- **HTTP Rejection**: ARM rejects HTTP URLs with clear error messages
- **Development Override**: `--insecure` flag allows HTTP for testing
- **Global Flag**: `--insecure` applies to all registries in the command

**HTTPS Enforcement Examples:**
```bash
# HTTP URLs rejected by default
$ arm config add registry test http://insecure.example.com/
Error [CONFIG]: HTTP URLs not allowed for security
Details: Registry URLs must use HTTPS protocol
Suggestion: Use https://insecure.example.com/ or add --insecure flag for testing

# Override for development/testing
$ arm install test-rules --insecure
Warning: Using insecure HTTP connections (--insecure flag)
```

**Certificate Validation:**
- **Strict by Default**: Validate SSL certificates against trusted CAs
- **Self-Signed Rejection**: Reject self-signed certificates by default
- **Development Override**: `--insecure` flag allows self-signed certificates
- **Certificate Errors**: Clear error messages for certificate validation failures

**Certificate Validation Examples:**
```bash
# Self-signed certificate rejected
$ arm install my-rules
Error [NETWORK]: SSL certificate verification failed
Details: Self-signed certificate for 'registry.example.com'
Suggestion: Use --insecure flag for testing or install proper SSL certificate

# Override for development
$ arm install my-rules --insecure
Warning: Skipping SSL certificate verification (--insecure flag)
```

**TLS Security:**
- **Minimum TLS Version**: Require TLS 1.2 or higher
- **Cipher Suite Validation**: Use secure cipher suites only
- **Certificate Pinning**: Not implemented in MVP (rely on CA validation)
- **HSTS Support**: Honor HTTP Strict Transport Security headers

### 8.4 Rate Limiting and Resource Controls

**Rate Limiting Configuration:**
```ini
# .armrc rate limiting settings
[registries]
my-api = http://api.example.com/
gitlab-instance = gitlab://gitlab.company.com/project/123

# Type-based defaults
[http]
concurrency = 5
rateLimit = 30/minute

[gitlab]
concurrency = 2
rateLimit = 60/hour

# Registry-specific overrides
[registries.my-api]
authToken = $API_TOKEN
rateLimit = 10/minute          # Override http default
concurrency = 2                # Override http default

[registries.gitlab-instance]
authToken = $GITLAB_TOKEN
rateLimit = 100/hour           # Override gitlab default
concurrency = 5                # Override gitlab default
```

**Rate Limit Formats:**
- **Per Second**: `30/second`, `30/s`
- **Per Minute**: `100/minute`, `100/min`, `100/m`
- **Per Hour**: `1000/hour`, `1000/hr`, `1000/h`
- **Per Day**: `10000/day`, `10000/d`

**Rate Limit Behavior:**
- **Queue Requests**: When rate limit hit, queue additional requests
- **Wait and Retry**: Automatically wait for rate limit window to reset
- **Exponential Backoff**: Use exponential backoff for repeated rate limit hits
- **Progress Indication**: Show "Rate limited, waiting..." in progress output

**Rate Limiting Examples:**
```bash
# Rate limit hit during installation
$ arm install multiple-rulesets --verbose
[DEBUG] Installing ruleset 1/5...
[DEBUG] Installing ruleset 2/5...
[DEBUG] Rate limit reached for registry 'my-api' (10/minute)
[INFO] Rate limited, waiting 45 seconds...
[DEBUG] Installing ruleset 3/5...
```

**Download Size Limits:**
- **Default Limit**: 100MB per ruleset download
- **Configurable**: Set via `maxDownloadSize` in registry configuration
- **Size Check**: Validate Content-Length header before download
- **Streaming Validation**: Monitor download size during transfer

**Download Size Configuration:**
```ini
# .armrc download size limits
[registries]
large-rulesets = s3://large-bucket/
small-api = http://api.example.com/

# Type-based defaults
[s3]
maxDownloadSize = 100MB        # Default for all S3 registries

[http]
maxDownloadSize = 100MB        # Default for all HTTP registries

# Registry-specific overrides
[registries.large-rulesets]
maxDownloadSize = 500MB        # Allow larger downloads

[registries.small-api]
maxDownloadSize = 10MB         # Restrict download size
```

**Resource Control Examples:**
```bash
# Download size exceeded
$ arm install huge-ruleset
Error [REGISTRY]: Download size limit exceeded
Details: Ruleset size 150MB exceeds limit of 100MB
Suggestion: Increase maxDownloadSize in registry configuration or contact registry owner

# Size validation during download
$ arm install large-ruleset --verbose
[DEBUG] Downloading large-ruleset@1.0.0 (45MB)...
[DEBUG] Download progress: 45MB/45MB (100%)
✓ Downloaded large-ruleset@1.0.0
```

### 8.5 Security Error Handling

**Authentication Errors:**
```bash
# Git authentication failure
Error [AUTH]: Git authentication failed for 'github.com/user/repo'
Details: Permission denied (publickey)
Suggestion: Check SSH key configuration or use HTTPS with token

# S3 authentication failure
Error [AUTH]: AWS authentication failed for S3 bucket
Details: The AWS Access Key Id you provided does not exist
Suggestion: Check AWS credentials with 'aws sts get-caller-identity'

# GitLab token invalid
Error [AUTH]: GitLab authentication failed
Details: HTTP 401 - Invalid token
Suggestion: Check token permissions (requires 'read_api' and 'read_repository')

# HTTP bearer token invalid
Error [AUTH]: HTTP authentication failed
Details: HTTP 401 - Unauthorized
Suggestion: Check bearer token in registry configuration
```

**Security Configuration Errors:**
```bash
# Insecure file permissions
Error [CONFIG]: Insecure permissions on configuration file
Details: ~/.armrc has permissions 644 (should be 600)
Suggestion: Run 'chmod 600 ~/.armrc' to fix permissions

# Missing environment variable
Error [CONFIG]: Environment variable not found
Details: Variable 'GITLAB_TOKEN' referenced in .armrc but not set
Suggestion: Set environment variable or update configuration
```

**Network Security Errors:**
```bash
# HTTP URL rejected
Error [NETWORK]: Insecure protocol not allowed
Details: HTTP URLs are not permitted for security
Suggestion: Use HTTPS URL or add --insecure flag for testing

# Certificate validation failure
Error [NETWORK]: SSL certificate verification failed
Details: Certificate has expired for 'registry.example.com'
Suggestion: Contact registry administrator or use --insecure flag for testing
``` Registry Type
### 8.2 Credential Management
### 8.3 Environment Variable Support
### 8.4 Security Considerations
