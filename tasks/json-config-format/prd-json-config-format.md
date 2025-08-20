# PRD: Migrate ARM Configuration from INI to JSON Format

## Introduction/Overview

ARM currently uses an INI format for its configuration file (`.armrc`). This PRD outlines migrating to JSON format (`.armrc.json`) to achieve consistency with supported formats and modern tooling standards. This will be a breaking change requiring users to manually recreate their configuration files.

## Goals

- Replace INI configuration format with JSON for consistency with supported formats
- Implement JSON schema validation for configuration integrity
- Update all configuration-related documentation
- Ensure robust JSON parsing and validation

## User Stories

- As a developer using ARM, I want my configuration in JSON format so that I have better IDE tooling support
- As a team member, I want schema validation so that I get clear error messages for invalid configurations
- As a new ARM user, I want JSON configuration examples in documentation so that I can set up ARM correctly
- As an existing ARM user, I want to manually recreate my configuration in the new format

## Functional Requirements

1. ARM must read configuration from `.armrc.json` instead of `.armrc`
2. JSON structure must use flat format mirroring current INI sections (registries, channels, cache)
3. ARM must validate JSON against a defined schema with helpful error messages
4. ARM must ignore existing `.armrc` INI files silently
5. All CLI commands modifying configuration must write to `.armrc.json`
6. Schema must validate registry URLs, types, and channel directories
7. ARM must handle malformed JSON with descriptive error messages
8. All current configuration options must be supported in JSON format

## Non-Goals (Out of Scope)

- Automatic migration from INI to JSON
- Backward compatibility with INI format
- Supporting multiple configuration file formats
- Migration commands or tooling

## Design Considerations

### JSON Structure
```json
{
  "registries": {
    "default": {
      "url": "https://github.com/user/repo",
      "type": "git"
    }
  },
  "channels": {
    "cursor": {
      "directories": [".cursor/rules"]
    }
  },
  "cache": {
    "ttl": "24h",
    "maxSize": "100MB"
  }
}
```

## Technical Considerations

- Replace INI parser with `encoding/json` in Go codebase
- Implement JSON schema validation library
- Update all configuration read/write operations
- Update configuration file constants and paths

## Success Metrics

- All ARM commands work with `.armrc.json` without regression
- Schema validation provides helpful error messages
- Documentation updated with working JSON examples
- JSON parsing handles edge cases gracefully

## Open Questions

- Which Go JSON schema validation library to use?
- Should we add `arm config validate` command for configuration checking?
