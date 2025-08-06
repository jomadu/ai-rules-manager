# Getting Started

This guide walks you through your first ARM setup and ruleset installation.

## Prerequisites

- ARM installed ([Installation Guide](installation.md))
- Cursor IDE or Amazon Q Developer

## Step 1: Initialize Your Project

Navigate to your project directory:

```bash
cd your-project
```

ARM works without initialization, but you can create a manifest for team sharing:

```bash
# Optional: Create rules.json for team projects
cat > rules.json << EOF
{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "typescript-rules": "^1.0.0"
  }
}
EOF
```

## Step 2: Configure a Registry

Before installing rulesets, configure at least one registry:

```bash
# Configure GitLab registry
arm config set sources.company https://gitlab.company.com
arm config set sources.company.type gitlab
arm config set sources.company.projectID 12345

# Or configure S3 registry
arm config set sources.s3 s3://my-rules-bucket/packages/
arm config set sources.s3.type s3
```

## Step 3: Install Your First Ruleset

```bash
# Install from configured registry
arm install company@typescript-rules

# Or install from git repository (no config needed)
arm install github.com/user/awesome-rules@main:rules/*.md

# For higher rate limits, configure git authentication
arm config set sources.github.authToken $GITHUB_TOKEN
arm config set sources.gitlab.authToken $GITLAB_TOKEN
```

## Step 4: Verify Installation

```bash
# List installed rulesets
arm list

# Check your IDE directories
ls -la .cursorrules/arm/
ls -la .amazonq/rules/arm/
```

You should see your ruleset files organized by registry and version.

## Step 5: Use in Your IDE

### Cursor IDE
Your rules are automatically available in `.cursorrules/` - Cursor will load them automatically.

### Amazon Q Developer
Your rules are in `.amazonq/rules/` - Amazon Q will use them for context.

## Step 6: Team Collaboration

Share your `rules.json` with your team:

```bash
# Team members can install all dependencies
arm install

# Keep rulesets updated
arm update
```

## Common Workflows

### Check for Updates
```bash
arm outdated
arm update
```

### Add More Rulesets
```bash
arm install react-rules
arm install company@python-standards
```

### Configure Registries
```bash
arm config set sources.company https://gitlab.company.com
```

## Next Steps

- [Commands Reference](commands.md) - Complete command documentation
- [Configuration](configuration.md) - Customize ARM behavior
- [Registries](registries.md) - Set up custom registries
