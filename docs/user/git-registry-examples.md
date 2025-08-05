# Git Registry Examples

This document provides practical examples of using ARM with git repositories as registries.

## Configuration Examples

### Basic Git Registry Setup

Create a `.armrc` file in your project or home directory:

```ini
[sources]
default = https://registry.armjs.org/
awesome-rules = https://github.com/PatrickF1/awesome-cursorrules
company-rules = https://github.com/company/internal-rules

[sources.awesome-rules]
type = git
api = github

[sources.company-rules]
type = git
api = github
authToken = $COMPANY_GITHUB_TOKEN

[performance]
defaultConcurrency = 3

[performance.git]
concurrency = 2
```

### GitLab Registry

```ini
[sources]
gitlab-rules = https://gitlab.com/company/coding-standards

[sources.gitlab-rules]
type = git
api = gitlab
authToken = $GITLAB_TOKEN
```

### Generic Git Provider

```ini
[sources]
custom-git = https://git.example.com/team/rules

[sources.custom-git]
type = git
# No api field - uses generic git operations
authToken = $CUSTOM_GIT_TOKEN
```

## Usage Examples

### Installing from Git Repositories

#### Install from Branch (Auto-updating)

```bash
# Install from main branch - will update when new commits are pushed
arm install awesome-rules@main

# Install from develop branch
arm install awesome-rules@develop
```

#### Install from Specific Tag (Semver)

```bash
# Install specific version tag
arm install company-rules@v2.1.0

# Install with semver constraint (updates to compatible versions)
arm install company-rules@^2.1.0
```

#### Install from Specific Commit (Pinned)

```bash
# Install from specific commit - never auto-updates
arm install experimental@abc1234567890abcdef1234567890abcdef123456

# Short commit SHA also works
arm install experimental@abc1234
```

### Using File Patterns

Create a `rules.json` file to specify which files to install:

```json
{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "typescript-rules": "^1.0.0",
    "awesome-rules@main": {
      "patterns": ["rules/*.md", "docs/*.txt"]
    },
    "company-rules@v2.1.0": {
      "patterns": ["security/**", "typescript/**"]
    },
    "experimental@abc1234": {
      "patterns": ["experimental/*.md"]
    }
  }
}
```

Then install all dependencies:

```bash
arm install
```

### Pattern Examples

| Pattern | Description | Matches |
|---------|-------------|---------|
| `*.md` | All markdown files in root | `readme.md`, `guide.md` |
| `rules/*.md` | Markdown files in rules directory | `rules/typescript.md`, `rules/react.md` |
| `**/*.md` | All markdown files recursively | `docs/api/readme.md`, `src/guide.md` |
| `{rules,docs}/*.md` | Markdown files in rules or docs | `rules/style.md`, `docs/api.md` |
| `security/**` | All files in security directory | `security/auth.md`, `security/crypto/keys.md` |

## Reference Types and Update Behavior

### Branch References

```bash
# These will auto-update when new commits are pushed
arm install awesome-rules@main
arm install awesome-rules@develop
arm install awesome-rules@feature/new-rules
```

**Update Behavior**: ARM checks for new commits and updates automatically during `arm update`.

### Tag References (Semver)

```bash
# Exact version - no updates
arm install company-rules@v2.1.0

# Semver constraints - updates to compatible versions
arm install company-rules@^2.1.0  # 2.1.0 <= version < 3.0.0
arm install company-rules@~2.1.0  # 2.1.0 <= version < 2.2.0
```

**Update Behavior**: ARM respects semver constraints and updates to newer compatible versions.

### Commit References

```bash
# Full SHA
arm install experimental@abc1234567890abcdef1234567890abcdef123456

# Short SHA (7+ characters)
arm install experimental@abc1234
```

**Update Behavior**: Never updates - pinned to specific commit.

### Default Reference

```bash
# Uses HEAD of default branch (main or master)
arm install awesome-rules
```

**Update Behavior**: Same as branch reference for the default branch.

## Authentication

### GitHub Personal Access Token

```bash
# Set token for private repositories
arm config set sources.company-rules.authToken $GITHUB_TOKEN

# Or use environment variable in .armrc
# authToken = $GITHUB_TOKEN
```

### GitLab Access Token

```bash
# Set GitLab token
arm config set sources.gitlab-rules.authToken $GITLAB_TOKEN
```

### Generic Git Authentication

For other git providers, ARM uses the token in the URL:

```bash
# GitHub-style
https://token@github.com/owner/repo.git

# GitLab-style (oauth2)
https://oauth2:token@gitlab.com/owner/repo.git
```

## Performance Optimization

### API vs Git Operations

ARM can use hosting provider APIs for better performance:

```ini
[sources.awesome-rules]
type = git
api = github  # Uses GitHub API - 1000x faster for sparse file selection

[sources.gitlab-rules]
type = git
api = gitlab  # Uses GitLab API

[sources.custom-git]
type = git
# No api field - uses git operations (slower but works everywhere)
```

### Concurrency Settings

```ini
[performance]
defaultConcurrency = 3

[performance.git]
concurrency = 2  # Lower for git operations to avoid rate limits
```

## Troubleshooting

### Common Issues

1. **Authentication Failed**
   ```bash
   # Check token permissions
   arm config get sources.company-rules.authToken

   # Update token
   arm config set sources.company-rules.authToken $NEW_TOKEN
   ```

2. **No Files Matched Patterns**
   ```bash
   # List available files in repository
   git ls-tree -r --name-only HEAD

   # Test pattern locally
   find . -name "*.md"
   ```

3. **Rate Limiting**
   ```bash
   # Reduce concurrency
   arm config set performance.git.concurrency 1
   ```

4. **Repository Not Found**
   ```bash
   # Check URL and authentication
   git ls-remote https://github.com/owner/repo.git
   ```

### Debug Mode

```bash
# Enable verbose logging
ARM_DEBUG=1 arm install awesome-rules@main
```

## Best Practices

1. **Use Specific Tags for Production**
   ```json
   {
     "dependencies": {
       "company-rules@v2.1.0": {
         "patterns": ["production/**"]
       }
     }
   }
   ```

2. **Use Branches for Development**
   ```json
   {
     "dependencies": {
       "experimental@develop": {
         "patterns": ["experimental/**"]
       }
     }
   }
   ```

3. **Minimize Pattern Scope**
   ```json
   {
     "dependencies": {
       "large-repo@main": {
         "patterns": ["rules/typescript/*.md"]
       }
     }
   }
   ```

4. **Use API Optimization**
   ```ini
   [sources.github-repo]
   type = git
   api = github  # Much faster than generic git operations
   ```

5. **Set Appropriate Concurrency**
   ```ini
   [performance.git]
   concurrency = 2  # Conservative to avoid rate limits
   ```

## Migration from Other Registries

### From HTTP Registry

```ini
# Before
[sources]
old-registry = https://registry.example.com/

# After
[sources]
git-registry = https://github.com/company/rules

[sources.git-registry]
type = git
api = github
```

### From Local Files

```bash
# Before
cp -r /local/rules .cursorrules/

# After
# Push rules to git repository, then:
arm install company-rules@main
```

This completes the git registry implementation with comprehensive examples and documentation.
