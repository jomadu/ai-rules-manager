# Testing ARM

## Test Registry

A minimal test registry is available in `test/registry/` for development and testing.

### Usage

1. **Start test server (auto-generates packages):**
   ```bash
   cd test/registry
   go run server.go
   ```

3. **Test ARM commands:**
   ```bash
   # Install rulesets
   go run ./cmd/arm install typescript-rules@1.0.0
   go run ./cmd/arm install security-rules@1.2.0

   # List installed rulesets
   go run ./cmd/arm list
   go run ./cmd/arm list --format=json
   ```

### Available Test Rulesets

- `typescript-rules@1.0.0` - TypeScript coding standards
- `security-rules@1.2.0` - Security best practices

### Test Results

The install/list workflow is now working end-to-end:
- ✅ Install creates rules.json and rules.lock
- ✅ Rulesets are extracted to both .cursorrules and .amazonq/rules
- ✅ List command displays installed rulesets in table and JSON formats
- ✅ Checksums are calculated and stored for integrity verification
