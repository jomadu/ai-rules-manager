# Configuration-Driven Target Directories

## Overview

ARM now uses configuration-driven target directories instead of hardcoded paths. This allows users to customize where rulesets are installed based on their project needs.

## How It Works

### Target Configuration

Target directories are specified in the `rules.json` manifest file:

```json
{
  "targets": [".cursorrules", ".amazonq/rules", ".custom-ai-tool/rules"],
  "dependencies": {
    "typescript-rules": "^1.0.0"
  }
}
```

### Default Targets

If no manifest exists, ARM will create one with default targets:
- `.cursorrules` (for Cursor IDE)
- `.amazonq/rules` (for Amazon Q Developer)

### Installation Behavior

When installing rulesets, ARM will:

1. Load the `rules.json` manifest
2. Extract rulesets to all specified target directories
3. Create the directory structure: `{target}/arm/{org}/{package}/{version}/`

### Uninstallation Behavior

When uninstalling rulesets, ARM will:

1. Load the `rules.json` manifest to determine target directories
2. Remove rulesets from all specified target directories
3. Clean up empty parent directories

## Benefits

- **Flexibility**: Support for any AI coding tool by configuring custom targets
- **Consistency**: All operations respect the same configuration
- **Maintainability**: No hardcoded paths in the codebase
- **Future-proof**: Easy to add support for new AI tools

## Migration

Existing projects with hardcoded assumptions will continue to work as the default targets match the previous hardcoded values. Users can customize targets by editing their `rules.json` file.
