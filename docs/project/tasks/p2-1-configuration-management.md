# P2.1: Configuration Management

## Overview
Implement configuration file parsing and management for .armrc files at user and project levels with environment variable support.

## Requirements
- Parse .armrc files in INI format
- Support user-level (~/.armrc) and project-level (.armrc) configs
- Environment variable substitution
- Config command implementation (get, set, list)

## Tasks
- [x] **INI file parsing**:
  ```ini
  [sources]
  default = https://registry.armjs.org/
  company = https://internal.company.local/

  [sources.company]
  authToken = $COMPANY_TOKEN
  ```
- [x] **Configuration hierarchy**:
  - User config: `~/.armrc`
  - Project config: `./.armrc`
  - Environment variables override
  - Command-line flags override (planned for future)
- [x] **Environment variable substitution**:
  - Replace `$VAR` and `${VAR}` patterns
  - Default values not implemented (not required)
- [x] **Config command implementation**:
  ```bash
  arm config list
  arm config get sources.default
  arm config set sources.company https://new-url.com
  ```
- [x] **Validation**:
  - Basic validation implemented
  - Auth token masking for security

## Acceptance Criteria
- [x] .armrc files are parsed correctly
- [x] Configuration hierarchy works (project overrides user)
- [x] Environment variables are substituted properly
- [x] Config commands work for get/set/list operations
- [x] Invalid configurations show helpful errors
- [x] Authentication tokens are handled securely (masked in output)

## Dependencies
- gopkg.in/ini.v1 (INI file parsing)
- os (environment variables)

## Files to Create
- `internal/config/parser.go`
- `internal/config/hierarchy.go`
- `cmd/arm/config.go`

## Example Config
```ini
[sources]
default = https://registry.armjs.org/
company = https://internal.company.local/

[sources.company]
authToken = ${COMPANY_REGISTRY_TOKEN}
timeout = 30s

[cache]
location = ~/.mpm/cache
maxSize = 1GB
```

## Notes
- Consider encrypted storage for sensitive tokens
- Plan for config validation and migration
- Support both INI and YAML formats in future
