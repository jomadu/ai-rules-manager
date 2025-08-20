## PRD Reference

**Source PRD:** `tasks/json-config-format/prd-json-config-format.md`

**Key Requirements Summary:**
- Replace INI configuration format (.armrc) with JSON format (.armrc.json)
- Implement JSON schema validation with helpful error messages
- Maintain flat JSON structure mirroring current INI sections (registries, channels, cache)
- Silently ignore existing .armrc INI files without backward compatibility
- Update all CLI commands to read/write .armrc.json instead of .armrc
- Support all current configuration options in JSON format

## Relevant Files

- `internal/config/config.go` - Main configuration loading and parsing logic that needs JSON migration
- `internal/config/config_test.go` - Unit tests for configuration functionality
- `internal/cli/commands.go` - CLI commands that modify configuration files
- `internal/cli/commands_test.go` - Unit tests for CLI configuration commands
- `go.mod` - May need JSON schema validation library dependency
- `sandbox/.armrc` - Example INI configuration file for reference
- `docs/user/configuration.md` - Documentation that needs updating with JSON examples

### Notes

- JSON schema validation should use a Go library like `github.com/xeipuuv/gojsonschema`
- The current config.go already handles both INI and JSON files but needs to be refactored to JSON-only
- All configuration modification commands in CLI need to write JSON instead of INI format

## Tasks

- [x] 1.0 Replace INI Configuration Parser with JSON-Only Implementation
  - [x] 1.1 Add JSON schema validation library dependency to go.mod
  - [x] 1.2 Remove INI parsing logic from config.go loadINIFile and processSection methods
  - [x] 1.3 Create new JSON configuration structure matching INI sections (registries, channels, cache)
  - [x] 1.4 Update loadConfigFromPaths to only load .armrc.json files, ignore .armrc files
  - [x] 1.5 Refactor mergeConfigs to work with JSON-only configuration data
  - [x] 1.6 Update validateConfig to validate JSON configuration structure
- [ ] 2.0 Implement JSON Schema Validation
  - [ ] 2.1 Define JSON schema for ARM configuration with all current INI sections
  - [ ] 2.2 Implement schema validation function with descriptive error messages
  - [ ] 2.3 Add validation for registry URLs, types, and channel directories per schema
  - [ ] 2.4 Integrate schema validation into configuration loading process
  - [ ] 2.5 Add malformed JSON error handling with helpful messages
- [ ] 3.0 Update CLI Configuration Commands for JSON Format
  - [ ] 3.1 Update getConfigPath to return .armrc.json instead of .armrc
  - [ ] 3.2 Replace loadOrCreateINI with loadOrCreateJSON for all config commands
  - [ ] 3.3 Update handleConfigSet to modify JSON structure instead of INI sections
  - [ ] 3.4 Update handleAddRegistry to write to JSON registries section
  - [ ] 3.5 Update handleAddChannel to write to JSON channels section
  - [ ] 3.6 Update handleRemoveRegistry and handleRemoveChannel for JSON format
- [ ] 4.0 Update Configuration File Generation and Stub Creation
  - [ ] 4.1 Update GenerateStubFiles to create .armrc.json instead of .armrc
  - [ ] 4.2 Replace generateARMRCStub with generateARMRCJSONStub function
  - [ ] 4.3 Create JSON stub content with all current configuration sections
  - [ ] 4.4 Update ensureConfigFiles to check for .armrc.json existence
- [ ] 5.0 Update Documentation and Examples
  - [ ] 5.1 Update README.md examples to show .armrc.json instead of .armrc
  - [ ] 5.2 Create or update configuration documentation with JSON examples
  - [ ] 5.3 Update sandbox configuration files to use JSON format
  - [ ] 5.4 Update any integration tests to use JSON configuration format
