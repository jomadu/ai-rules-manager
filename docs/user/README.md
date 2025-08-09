# User Documentation

Complete guides for using AI Rules Manager (ARM) to manage your AI coding assistant rulesets.

## Getting Started

### [Quick Start Guide](quick-start.md)
Get ARM up and running in 5 minutes with step-by-step examples.

### [Configuration Guide](configuration.md)
Complete guide to configuring ARM with `.armrc` and `arm.json` files.

### [Registry Guide](registries.md)
Set up different types of registries (Git, S3, GitLab, HTTPS, Local) for storing rulesets.

### [Team Setup Guide](team-setup.md)
Guide for team leads and platform engineers to distribute rulesets across development teams.

## Quick Reference

### Common Commands
```bash
# Install ARM and generate config files
arm install

# Add a registry
arm config add registry my-registry https://github.com/user/repo.example --type=git

# Add a channel
arm config add channel cursor --directories ~/.cursor/rules

# Install a ruleset
arm install my-rules --patterns "rules/*.md"

# List installed rulesets
arm list

# Check for updates
arm outdated

# Update rulesets
arm update
```

### Registry Types
- **Git** - GitHub, GitLab, and other Git repositories
- **S3** - AWS S3 buckets
- **GitLab** - GitLab Package Registry
- **HTTPS** - Generic HTTP servers
- **Local** - Local file system directories

## Need Help?

Each guide includes troubleshooting sections for common issues. If you encounter problems not covered in the guides, please check the project issues or create a new one.
