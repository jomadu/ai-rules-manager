# P2.1: Configuration Management

## Overview
Implement configuration file parsing and management for .armrc files at user and project levels with environment variable support.

## Requirements
- Parse .armrc files in INI format
- Support user-level (~/.armrc) and project-level (.armrc) configs
- Environment variable substitution
- Config command implementation (get, set, list)

## Tasks
- [ ] **INI file parsing**:
  ```ini
  [sources]
  default = https://registry.armjs.org/
  company = https://internal.company.local/
  
  [sources.company]
  authToken = $COMPANY_TOKEN
  ```
- [ ] **Configuration hierarchy**:
  - User config: `~/.armrc`
  - Project config: `./.armrc`
  - Environment variables override
  - Command-line flags override
- [ ] **Environment variable substitution**:
  - Replace `$VAR` and `${VAR}` patterns
  - Support default values: `${VAR:-default}`
- [ ] **Config command implementation**:
  ```bash
  arm config list
  arm config get sources.default
  arm config set sources.company https://new-url.com
  ```
- [ ] **Validation**:
  - Validate URL formats
  - Check required fields
  - Warn about missing authentication

## Acceptance Criteria
- [ ] .armrc files are parsed correctly
- [ ] Configuration hierarchy works (project overrides user)
- [ ] Environment variables are substituted properly
- [ ] Config commands work for get/set/list operations
- [ ] Invalid configurations show helpful errors
- [ ] Authentication tokens are handled securely

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