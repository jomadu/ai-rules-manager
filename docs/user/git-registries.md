# Git Repository Registries

ARM supports installing rulesets directly from git repositories, allowing you to use collections like awesome-cursorrules or your company's internal rule repositories.

## Quick Start

### Install from Git Repository

```bash
# Install from a specific branch with glob patterns
arm install awesome-rules@main:rules/*.md,docs/*.txt

# Install from a specific commit (pinned, no auto-updates)
arm install awesome-rules@abc1234:**.md

# Install from a semver tag (updates to compatible versions)
arm install awesome-rules@v1.2.0:rules/**
```

## Configuration

### Registry Setup

Configure git repositories as sources in `.armrc`:

```ini
[sources]
awesome-rules = https://github.com/PatrickF1/awesome-cursorrules
company-rules = https://github.com/company/internal-rules

[sources.awesome-rules]
type = git
api = github  # Optional: enables GitHub API optimization
# No authToken needed for public repos

[sources.company-rules]
type = git
api = gitlab  # Optional: enables GitLab API optimization
authToken = $COMPANY_GITHUB_TOKEN
```

### Project Dependencies

Add git-based dependencies to `rules.json`:

```json
{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "awesome-rules@main": {
      "patterns": ["rules/*.md", "examples/*.txt"]
    },
    "company-rules@v2.1.0": {
      "patterns": ["security/**", "typescript/**"]
    }
  }
}
```

## Reference Types

### Branch References (Auto-updating)

```bash
arm install awesome-rules@main:**.md
arm install awesome-rules@develop:rules/**
```

- Updates automatically when `arm update` detects new commits
- Uses HEAD of specified branch if no commit specified

### Commit References (Pinned)

```bash
arm install awesome-rules@abc1234:**.md
arm install awesome-rules@a1b2c3d4e5f6:rules/**
```

- Never updates automatically
- Supports both full and short SHA formats

### Tag References (Semver Updates)

```bash
arm install awesome-rules@v1.2.0:**.md
arm install awesome-rules@1.2.0:rules/**  # v prefix optional
```

- Updates to compatible semver versions
- Only considers valid semver tags (ignores `beta`, `release-2023`, etc.)
- Supports both `v1.0.0` and `1.0.0` formats

### Default Reference

```bash
arm install awesome-rules:**.md  # Uses HEAD of default branch
```

## File Patterns

### Glob Pattern Syntax

ARM supports standard glob patterns for selecting files:

```bash
# Single pattern
awesome-rules@main:rules/*.md

# Multiple patterns (comma-separated)
awesome-rules@main:rules/*.md,docs/*.txt,examples/**

# Recursive patterns
awesome-rules@main:**  # All files
awesome-rules@main:rules/**  # All files under rules/
```

### Pattern Examples

| Pattern | Matches |
|---------|---------|
| `*.md` | All .md files in root |
| `rules/*.md` | All .md files in rules/ directory |
| `**/*.md` | All .md files recursively |
| `rules/**` | All files under rules/ directory |
| `*.{md,txt}` | All .md and .txt files in root |

## File Structure

Git repositories maintain their directory structure when installed:

### Repository Structure
```
awesome-cursorrules/
  rules/
    typescript/
      strict.md
      react.md
    python/
      pep8.md
  docs/
    examples/
      sample.md
```

### Installed Structure
```
.cursorrules/
  arm/
    awesome-rules/
      main/  # or commit/tag
        rules/
          typescript/
            strict.md
            react.md
        python/
          pep8.md
        docs/
          examples/
            sample.md
```

## Update Behavior

### Branch-based Dependencies

```json
{
  "dependencies": {
    "awesome-rules@main": {
      "patterns": ["**.md"]
    }
  }
}
```

- `arm update` checks for new commits on the branch
- Updates if newer commits are available
- `arm outdated` shows if updates are available

### Commit-based Dependencies

```json
{
  "dependencies": {
    "awesome-rules@abc1234": {
      "patterns": ["**.md"]
    }
  }
}
```

- Never updates automatically
- `arm outdated` shows "pinned" status
- Manual update required to change commit

### Tag-based Dependencies

```json
{
  "dependencies": {
    "awesome-rules@v1.2.0": {
      "patterns": ["**.md"]
    }
  }
}
```

- `arm update` looks for newer compatible semver tags
- Follows semver compatibility rules
- `arm outdated` shows available updates

## Authentication

### Public Repositories

No authentication required:

```ini
[sources.awesome-rules]
type = git
url = https://github.com/PatrickF1/awesome-cursorrules
```

### Private Repositories

Use personal access tokens:

```ini
[sources.company-rules]
type = git
url = https://github.com/company/internal-rules
authToken = $COMPANY_GITHUB_TOKEN
```

Set tokens in environment variables:

```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"
export COMPANY_GITHUB_TOKEN="ghp_yyyyyyyyyyyyyyyyyyyy"
```

## Examples

### Installing from awesome-cursorrules

```bash
# Configure the source
cat >> .armrc << EOF
[sources]
awesome-rules = https://github.com/PatrickF1/awesome-cursorrules

[sources.awesome-rules]
type = git
EOF

# Install specific rules
arm install awesome-rules@main:rules/cursor-*.md

# Or add to rules.json
cat > rules.json << EOF
{
  "targets": [".cursorrules"],
  "dependencies": {
    "awesome-rules@main": {
      "patterns": ["rules/cursor-*.md", "rules/typescript-*.md"]
    }
  }
}
EOF

arm install
```

### Company Internal Rules

```bash
# Configure private repository
cat >> .armrc << EOF
[sources]
company-rules = https://github.com/company/coding-standards

[sources.company-rules]
type = git
authToken = $COMPANY_GITHUB_TOKEN
EOF

# Install from specific version
arm install company-rules@v2.1.0:standards/**,security/**
```

### Mixed Registry Setup

```json
{
  "dependencies": {
    "typescript-rules": "^1.0.0",
    "awesome-rules@main": {
      "patterns": ["rules/typescript-*.md"]
    },
    "company-rules@v2.1.0": {
      "patterns": ["security/**"]
    }
  }
}
```

## Troubleshooting

### Authentication Issues

```bash
# Verify token is set
echo $GITHUB_TOKEN

# Test repository access
git clone https://github.com/company/private-repo
```

### Pattern Matching

```bash
# Dry run to see what files would be installed
arm install --dry-run awesome-rules@main:rules/*.md
```

### Update Issues

```bash
# Check what updates are available
arm outdated

# Force update specific ruleset
arm update awesome-rules@main
```

## Performance Optimization

### API-First Approach

ARM automatically uses provider APIs when available for significant performance improvements:

- **GitHub API**: 1000x+ faster file selection vs full repository clone
- **GitLab API**: Direct file access without repository download
- **Generic Git**: Fallback to standard git operations

### Configuration

```ini
[sources.awesome-rules]
type = git
api = github  # Enables GitHub API optimization
url = https://github.com/PatrickF1/awesome-cursorrules

[sources.company-gitlab]
type = git
api = gitlab  # Enables GitLab API optimization
url = https://gitlab.company.com/team/rules-repo
authToken = $GITLAB_TOKEN

[sources.generic-git]
type = git
# No api field = uses git operations
url = https://git.company.com/repo.git
```

## Cache Management

### Cache Structure

```
~/.arm/cache/
  packages/          # Package registry cache
  git/              # Git repository cache
    github.com/
      PatrickF1/awesome-cursorrules/
        .git/         # Bare repository
        metadata.json # Cache metadata
```

### Cleanup

```bash
arm clean --cache    # Removes ALL cache (packages + git repos)
arm clean --dry-run  # Show what would be cleaned
```

## Limitations

- Only token-based authentication supported (no SSH keys yet)
- No support for git submodules
- API rate limiting may apply for hosted providers
- Generic git providers require full repository operations
