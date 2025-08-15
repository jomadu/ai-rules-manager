# Testing Strategy

Comprehensive testing approach and coverage requirements for ARM.

## Testing Philosophy

- **Test-Driven Development**: Write tests before implementation
- **Comprehensive Coverage**: Minimum 80% code coverage
- **Real-World Scenarios**: Test with actual registries and workflows
- **Performance Validation**: Benchmark critical paths
- **Security Testing**: Validate authentication and permissions

## Test Structure

### Test Organization

- `tests/unit/` - Unit tests (alongside code)
- `tests/integration/` - End-to-end workflow tests
- `tests/fixtures/` - Test data and mock registries
- `tests/benchmarks/` - Performance benchmarks
- `tests/security/` - Security-focused tests

### Test Categories

#### Unit Tests
- **Location**: Alongside source code (`*_test.go`)
- **Scope**: Individual functions and methods
- **Coverage**: 90%+ for core packages
- **Execution**: Fast (<1s total)

#### Integration Tests
- **Location**: `tests/integration/`
- **Scope**: Complete workflows and registry interactions
- **Coverage**: All major user scenarios
- **Execution**: Moderate (5-30s per test)

#### End-to-End Tests
- **Location**: `tests/e2e/`
- **Scope**: Full ARM installation and usage
- **Coverage**: Critical user journeys
- **Execution**: Slow (30s-5m per test)

## Unit Testing

### Test Structure
- Use table-driven tests for comprehensive coverage
- Test setup functions create isolated environments
- Mock implementations for external dependencies
- Validate both success and error conditions

### Mock Implementations
- MockRegistry for testing registry operations
- Mock file systems for installation testing
- Mock HTTP clients for network operations
- Configurable error injection for failure scenarios

## Integration Testing

### Git Registry Tests
- Set up temporary Git repositories
- Test version listing and downloading
- Verify authentication mechanisms
- Clean up test repositories after execution

### End-to-End Workflow Tests
- Full ARM command execution
- Registry configuration and management
- Ruleset installation and verification
- Lock file generation and validation

## Performance Testing

### Benchmarks
- Cache retrieval performance
- Pattern matching efficiency
- Registry operation timing
- Memory allocation tracking

### Load Testing
- Concurrent installation scenarios
- High-volume registry operations
- Resource usage under stress
- Error handling during load

## Security Testing

### Authentication Tests
- Valid and invalid token scenarios
- Empty credential handling
- Registry access control
- Token refresh mechanisms

### Permission Tests
- File system permission validation
- Directory access restrictions
- Installation to protected locations
- Error handling for permission failures

## Test Data Management

### Fixture Generation
- Automated test repository creation
- Version tagging and branching
- Test file content generation
- Cleanup and teardown procedures

### Mock Data
- Configuration file templates
- Registry response samples
- Error condition simulations
- Edge case data sets

## Continuous Integration

### GitHub Actions Workflow
- Multi-platform testing (Ubuntu, macOS, Windows)
- Unit, integration, and E2E test execution
- Coverage reporting and validation
- Automated test result publishing

## Coverage Requirements

### Package Coverage Targets
- **Core packages** (`config`, `registry`, `install`): 90%+
- **CLI package**: 80%+
- **Cache package**: 85%+
- **Utility packages**: 75%+
- **Overall project**: 80%+

### Coverage Reporting
- Generate HTML coverage reports
- Automated threshold checking
- Coverage trend tracking
- Integration with CI/CD pipeline

## Test Execution

### Make Targets
- `test` - Run unit and integration tests
- `test-unit` - Unit tests only
- `test-integration` - Integration tests only
- `test-e2e` - End-to-end tests
- `test-coverage` - Generate coverage reports

### Test Execution Guidelines
- Run unit tests frequently during development
- Run integration tests before committing
- Run full test suite before releasing
- Use `-short` flag for quick feedback loops
- Use `-race` flag to detect race conditions
