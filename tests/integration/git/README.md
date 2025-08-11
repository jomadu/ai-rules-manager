# ARM Testing Scripts

Scripts for testing ARM workflows with a comprehensive git-based test repository.

## Quick Setup

### 1. Create Test Repository Automatically

```bash
# Create comprehensive test repository
./scripts/test/setup-test-repos.sh

# Or with custom name
./scripts/test/setup-test-repos.sh my-test-repo
```

**Requirements:**
- GitHub CLI (`gh`) installed and authenticated
- Git installed

### 2. Run Tests

```bash
# The setup script will show you the exact command to run
./scripts/test/test-workflow.sh all "https://github.com/USERNAME/ai-rules-manager-test-git-registry"

# Or run interactively (will prompt for repo URL)
./scripts/test/test-workflow.sh all
```

## Usage

### Run Individual Test Scenarios

```bash
# Test installing latest version
./scripts/test/test-workflow.sh install-latest "https://github.com/USERNAME/ai-rules-manager-test-git-registry"

# Test semantic version constraints
./scripts/test/test-workflow.sh install-semver "https://github.com/USERNAME/ai-rules-manager-test-git-registry"

# Test individual file pattern matching
./scripts/test/test-workflow.sh install-patterns "https://github.com/USERNAME/ai-rules-manager-test-git-registry"

# Test combined pattern matching
./scripts/test/test-workflow.sh install-combined "https://github.com/USERNAME/ai-rules-manager-test-git-registry"
```

### Run All Tests

```bash
./scripts/test/test-workflow.sh all "https://github.com/USERNAME/ai-rules-manager-test-git-registry"
```

### Options

```bash
# Keep test artifacts for inspection
./scripts/test/test-workflow.sh all "https://github.com/USERNAME/ai-rules-manager-test-git-registry" --keep-artifacts

# Show verbose output
./scripts/test/test-workflow.sh all "https://github.com/USERNAME/ai-rules-manager-test-git-registry" --verbose

# Combine options
./scripts/test/test-workflow.sh install-semver "https://github.com/USERNAME/ai-rules-manager-test-git-registry" --verbose --keep-artifacts
```

## Test Scenarios

### install-latest
Tests installing rulesets without specifying a version (should get latest tag).

### install-semver
Tests semantic version constraints:
- Exact version: `test-repo/rules@1.0.0`
- Version 1.1.0: `test-repo/rules@1.1.0`
- Version 1.2.0: `test-repo/rules@1.2.0`
- Version 2.0.0: `test-repo/rules@2.0.0`

### install-patterns
Tests individual file pattern matching:
- Simple patterns: `'*.md'`, `'*.json'`
- Directory patterns: `'rules/**/*.md'`, `'cursor/*.md'`
- Tool-specific patterns: `'amazon-q/*.md'`

### install-combined
Tests combined pattern matching:
- Multiple patterns: `'*.md,*.json'`
- Complex combinations: `'rules/**/*.md,cursor/*.md'`
- Exclusion patterns: `'*.md,!rules/advanced/*.md'`

## Repository Structure

The setup script creates a comprehensive test repository with version history:

### Repository Content
- `ghost-hunting.md` / `ghost-detection.md` - Basic debugging guidelines
- `rules/mansion-maintenance.md` / `guidelines/maintenance.md` - Maintenance rules
- `rules/advanced/boss-battles.md` / `guidelines/expert-strategies.md` - Advanced techniques
- `cursor/its-a-me.md` / `tools/cursor-pro.md` - Cursor-specific rules
- `amazon-q/luigi-assistant.md` / `ai-assistants/q-developer.md` - AI assistant rules
- `config.json` / `settings.json` - Configuration files

### Version History
- **v1.0.0**: Basic content with fundamental rules
- **v1.1.0**: Enhanced content with advanced techniques
- **v1.2.0**: Best practices and configuration improvements
- **v2.0.0**: Breaking changes with restructured directories and renamed files

## Troubleshooting

### ARM Command Not Found
Make sure ARM is installed and in your PATH:
```bash
which arm
```

### GitHub CLI Issues
Make sure GitHub CLI is installed and authenticated:
```bash
gh --version
gh auth status
# If not authenticated:
gh auth login
```

### Repository Access Issues
The repository is created as public by default for testing compatibility.

### Test Failures
Use `--verbose` and `--keep-artifacts` flags to debug:
```bash
./scripts/test/test-workflow.sh install-semver "https://github.com/USERNAME/ai-rules-manager-test-git-registry" --verbose --keep-artifacts
# Check artifacts in /tmp/arm-test-* directory
```

### Test Configuration
Tests use `.armrc` configuration with registry setup:
```ini
[registries]
test-repo = https://github.com/USERNAME/ai-rules-manager-test-git-registry

[registries.test-repo]
type = git
```

The test repository includes version-specific content that allows tests to verify:
- Correct version resolution by checking for version-specific phrases
- Proper file pattern matching through dry-run commands
- Content integrity across versions
- Breaking change handling in v2.0.0
