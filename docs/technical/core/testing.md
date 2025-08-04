# Testing

Testing strategy and test registry for ARM development.

## Test Registry

Local test server in `test/registry/`:

```bash
cd test/registry
go run server.go
```

## Test Commands

```bash
# Install test rulesets
go run ./cmd/arm install typescript-rules@1.0.0
go run ./cmd/arm install security-rules@1.2.0

# Verify installation
go run ./cmd/arm list
```

## Test Rulesets

- `typescript-rules@1.0.0`
- `security-rules@1.2.0`

## Test Coverage

- Unit tests: `go test ./...`
- Integration tests: `test/integration/`
- End-to-end: Manual testing with test registry
