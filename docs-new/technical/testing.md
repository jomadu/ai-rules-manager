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
```
tests/
├── unit/                   # Unit tests (alongside code)
├── integration/            # End-to-end workflow tests
├── fixtures/              # Test data and mock registries
├── benchmarks/            # Performance benchmarks
└── security/              # Security-focused tests
```

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
```go
func TestConfigLoad(t *testing.T) {
    tests := []struct {
        name     string
        setup    func() string // Returns temp dir
        want     *Config
        wantErr  bool
    }{
        {
            name: "valid configuration",
            setup: func() string {
                dir := createTempDir(t)
                writeFile(t, filepath.Join(dir, ".armrc"), validARMRC)
                writeFile(t, filepath.Join(dir, "arm.json"), validARMJSON)
                return dir
            },
            want: &Config{
                Registries: map[string]string{
                    "default": "https://github.com/test/repo",
                },
            },
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tempDir := tt.setup()
            defer os.RemoveAll(tempDir)

            // Change to temp directory
            oldWd, _ := os.Getwd()
            os.Chdir(tempDir)
            defer os.Chdir(oldWd)

            got, err := config.Load()
            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want.Registries, got.Registries)
        })
    }
}
```

### Mock Implementations
```go
type MockRegistry struct {
    rulesets map[string]map[string][]byte // ruleset -> version -> content
    versions map[string][]string          // ruleset -> versions
    errors   map[string]error             // operation -> error
}

func (m *MockRegistry) DownloadRuleset(ctx context.Context, name, version, destDir string) error {
    if err, exists := m.errors["download"]; exists {
        return err
    }

    content, exists := m.rulesets[name][version]
    if !exists {
        return fmt.Errorf("ruleset %s@%s not found", name, version)
    }

    return os.WriteFile(filepath.Join(destDir, "ruleset.tar.gz"), content, 0644)
}

func (m *MockRegistry) ListVersions(ctx context.Context, name string) ([]string, error) {
    if err, exists := m.errors["list"]; exists {
        return nil, err
    }

    return m.versions[name], nil
}
```

## Integration Testing

### Git Registry Tests
```go
func TestGitRegistryIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Set up test repository
    testRepo := setupTestGitRepo(t)
    defer cleanupTestRepo(testRepo)

    // Create registry
    config := &registry.RegistryConfig{
        Name: "test",
        Type: "git",
        URL:  testRepo.URL,
    }

    reg, err := registry.NewGitRegistry(config, &registry.AuthConfig{}, nil)
    require.NoError(t, err)
    defer reg.Close()

    // Test version listing
    versions, err := reg.ListVersions(context.Background(), "test-ruleset")
    require.NoError(t, err)
    assert.Contains(t, versions, "v1.0.0")

    // Test download
    tempDir := t.TempDir()
    err = reg.DownloadRuleset(context.Background(), "test-ruleset", "v1.0.0", tempDir)
    require.NoError(t, err)

    // Verify downloaded content
    files, err := filepath.Glob(filepath.Join(tempDir, "*"))
    require.NoError(t, err)
    assert.NotEmpty(t, files)
}
```

### End-to-End Workflow Tests
```go
func TestInstallWorkflow(t *testing.T) {
    // Set up test environment
    testDir := t.TempDir()
    oldWd, _ := os.Getwd()
    os.Chdir(testDir)
    defer os.Chdir(oldWd)

    // Create test registry
    testRepo := setupTestGitRepo(t)
    defer cleanupTestRepo(testRepo)

    // Initialize ARM configuration
    err := exec.Command("arm", "install").Run()
    require.NoError(t, err)

    // Add test registry
    err = exec.Command("arm", "config", "add", "registry", "test", testRepo.URL, "--type=git").Run()
    require.NoError(t, err)

    // Add channel
    err = exec.Command("arm", "config", "add", "channel", "test", "--directories=.test/rules").Run()
    require.NoError(t, err)

    // Install ruleset
    err = exec.Command("arm", "install", "test/ruleset@v1.0.0", "--patterns=*.md").Run()
    require.NoError(t, err)

    // Verify installation
    files, err := filepath.Glob(".test/rules/arm/test/ruleset/*.md")
    require.NoError(t, err)
    assert.NotEmpty(t, files)

    // Verify lock file
    lockData, err := os.ReadFile("arm.lock")
    require.NoError(t, err)

    var lockFile config.LockFile
    err = json.Unmarshal(lockData, &lockFile)
    require.NoError(t, err)

    assert.Equal(t, "v1.0.0", lockFile.Rulesets["test"]["ruleset"].Version)
}
```

## Performance Testing

### Benchmarks
```go
func BenchmarkCacheRetrieval(b *testing.B) {
    cache := setupTestCache(b)
    key := "test-key"
    content := make([]byte, 1024*1024) // 1MB

    // Store content
    cache.Store(key, content, cache.Metadata{})

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _, _, err := cache.Retrieve(key)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkPatternMatching(b *testing.B) {
    matcher := patterns.NewGlobMatcher()
    patterns := []string{"**/*.md", "!**/drafts/**", "rules/**/*.md"}
    paths := generateTestPaths(1000)

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _, err := matcher.FilterPaths(patterns, paths)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Load Testing
```go
func TestConcurrentInstallations(t *testing.T) {
    const numGoroutines = 10
    const numInstallations = 5

    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines*numInstallations)

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            for j := 0; j < numInstallations; j++ {
                err := performInstallation(fmt.Sprintf("test-ruleset-%d-%d", id, j))
                if err != nil {
                    errors <- err
                }
            }
        }(i)
    }

    wg.Wait()
    close(errors)

    var errorList []error
    for err := range errors {
        errorList = append(errorList, err)
    }

    if len(errorList) > 0 {
        t.Fatalf("concurrent installations failed: %v", errorList)
    }
}
```

## Security Testing

### Authentication Tests
```go
func TestGitRegistryAuthentication(t *testing.T) {
    tests := []struct {
        name      string
        token     string
        wantError bool
    }{
        {"valid token", "valid-token", false},
        {"invalid token", "invalid-token", true},
        {"empty token", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := &registry.RegistryConfig{
                Name: "test",
                Type: "git",
                URL:  "https://github.com/private/repo",
            }

            auth := &registry.AuthConfig{
                Token: tt.token,
            }

            reg, err := registry.NewGitRegistry(config, auth, nil)
            if tt.wantError {
                assert.Error(t, err)
                return
            }

            require.NoError(t, err)
            defer reg.Close()

            // Test authenticated operation
            _, err = reg.ListVersions(context.Background(), "test-ruleset")
            if tt.wantError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Permission Tests
```go
func TestFilePermissions(t *testing.T) {
    tempDir := t.TempDir()

    // Create directory with restricted permissions
    restrictedDir := filepath.Join(tempDir, "restricted")
    err := os.Mkdir(restrictedDir, 0000)
    require.NoError(t, err)
    defer os.Chmod(restrictedDir, 0755) // Cleanup

    // Test installation to restricted directory
    installer := install.New(&config.Config{
        Channels: map[string]config.ChannelConfig{
            "restricted": {
                Directories: []string{restrictedDir},
            },
        },
    })

    req := &install.InstallRequest{
        Registry:    "test",
        Ruleset:     "test-ruleset",
        Version:     "1.0.0",
        SourceFiles: []string{filepath.Join(tempDir, "test.md")},
        Channels:    []string{"restricted"},
    }

    _, err = installer.Install(req)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "permission denied")
}
```

## Test Data Management

### Fixture Generation
```go
func setupTestGitRepo(t *testing.T) *TestRepo {
    tempDir := t.TempDir()

    // Initialize Git repository
    cmd := exec.Command("git", "init")
    cmd.Dir = tempDir
    require.NoError(t, cmd.Run())

    // Create test files
    testFiles := map[string]string{
        "README.md":           "# Test Ruleset",
        "rules/coding.md":     "# Coding Standards",
        "rules/security.md":   "# Security Guidelines",
        "guidelines/style.md": "# Style Guide",
    }

    for path, content := range testFiles {
        fullPath := filepath.Join(tempDir, path)
        require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
        require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
    }

    // Commit files
    cmd = exec.Command("git", "add", ".")
    cmd.Dir = tempDir
    require.NoError(t, cmd.Run())

    cmd = exec.Command("git", "commit", "-m", "Initial commit")
    cmd.Dir = tempDir
    require.NoError(t, cmd.Run())

    // Create tag
    cmd = exec.Command("git", "tag", "v1.0.0")
    cmd.Dir = tempDir
    require.NoError(t, cmd.Run())

    return &TestRepo{
        Path: tempDir,
        URL:  "file://" + tempDir,
    }
}
```

### Mock Data
```go
var (
    validARMRC = `
[registries]
default = https://github.com/test/repo

[registries.default]
type = git
`

    validARMJSON = `{
  "engines": {"arm": "^1.0.0"},
  "channels": {
    "test": {"directories": [".test/rules"]}
  },
  "rulesets": {
    "default": {
      "test-ruleset": {"version": "^1.0.0"}
    }
  }
}`
)
```

## Continuous Integration

### GitHub Actions Workflow
```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.23.2'

      - name: Run unit tests
        run: make test-unit

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.23.2'

      - name: Run integration tests
        run: make test-integration
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  e2e-tests:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.23.2'

      - name: Build ARM
        run: make build

      - name: Run E2E tests
        run: make test-e2e
```

## Coverage Requirements

### Package Coverage Targets
- **Core packages** (`config`, `registry`, `install`): 90%+
- **CLI package**: 80%+
- **Cache package**: 85%+
- **Utility packages**: 75%+
- **Overall project**: 80%+

### Coverage Reporting
```bash
# Generate coverage report
make test-coverage

# View HTML report
make coverage-html

# Check coverage threshold
make coverage-check
```

## Test Execution

### Make Targets
```makefile
.PHONY: test test-unit test-integration test-e2e test-coverage

test: test-unit test-integration

test-unit:
	go test -short -race -coverprofile=coverage.out ./...

test-integration:
	go test -tags=integration ./tests/integration/...

test-e2e:
	go test -tags=e2e ./tests/e2e/...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

coverage-check:
	go tool cover -func=coverage.out | grep total | awk '{if($$3 < 80.0) exit 1}'
```

### Test Execution Guidelines
- Run unit tests frequently during development
- Run integration tests before committing
- Run full test suite before releasing
- Use `-short` flag for quick feedback loops
- Use `-race` flag to detect race conditions
