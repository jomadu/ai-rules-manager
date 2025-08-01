# CC.2: Logging and Debugging

## Overview
Implement structured logging, debug mode, diagnostic information collection, and performance profiling capabilities.

## Requirements
- Implement structured logging
- Add debug mode with verbose output
- Create diagnostic information collection
- Add performance profiling capabilities

## Tasks
- [ ] **Structured logging system**:
  - JSON-formatted log output
  - Configurable log levels (debug, info, warn, error)
  - Contextual logging with request IDs
  - Log rotation and retention
- [ ] **Debug mode implementation**:
  - --debug flag for verbose output
  - Detailed operation tracing
  - HTTP request/response logging
  - File system operation logging
- [ ] **Diagnostic information**:
  - System information collection
  - Configuration dump
  - Cache state analysis
  - Registry connectivity tests
- [ ] **Performance profiling**:
  - CPU profiling support
  - Memory usage tracking
  - Operation timing metrics
  - Bottleneck identification

## Acceptance Criteria
- [ ] Logs are structured and parseable
- [ ] Debug mode provides useful troubleshooting info
- [ ] Diagnostic command helps with support issues
- [ ] Performance profiling identifies bottlenecks
- [ ] Log levels can be configured
- [ ] Sensitive information is not logged

## Dependencies
- github.com/sirupsen/logrus (structured logging)
- net/http/pprof (profiling)

## Files to Create
- `internal/logging/logger.go`
- `internal/logging/debug.go`
- `internal/diagnostics/collector.go`
- `internal/profiling/profiler.go`

## Logging Configuration
```go
type LogConfig struct {
    Level      string
    Format     string  // json, text
    Output     string  // stdout, stderr, file
    File       string
    MaxSize    int
    MaxBackups int
}
```

## Debug Output Example
```
DEBUG[2024-01-15T10:30:45Z] Starting install operation
  ruleset=typescript-rules
  version=1.2.3
  source=company

DEBUG[2024-01-15T10:30:45Z] Resolving registry
  registry=https://registry.company.com
  auth=token

DEBUG[2024-01-15T10:30:46Z] HTTP request
  method=GET
  url=https://registry.company.com/typescript-rules/1.2.3
  headers={"Authorization": "[REDACTED]"}
```

## Notes
- Consider log aggregation for enterprise users
- Plan for remote debugging capabilities
- Implement log sanitization for sensitive data