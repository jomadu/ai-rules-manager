# Test Registry

This directory contains a minimal test registry for ARM development and testing.

## Structure

```
test/registry/
├── server.go          # Simple HTTP server for testing
├── rulesets/          # Sample rulesets
│   ├── typescript-rules/
│   └── security-rules/
└── packages/          # Generated .tar.gz files
    ├── typescript-rules/
    └── security-rules/
```

## Usage

1. Start test server: `go run test/registry/server.go` (auto-generates packages)
2. Test ARM commands against localhost:8080
