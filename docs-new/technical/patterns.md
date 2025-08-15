# File Patterns

Technical specification for ARM's pattern matching and file selection system.

## Overview

ARM uses glob patterns to selectively install files from Git-based registries, enabling fine-grained control over which files are included in ruleset installations.

## Pattern Syntax

### Basic Patterns
```bash
*.md                    # All .md files in current directory
**/*.md                 # All .md files recursively
rules/*.md              # All .md files in rules directory
rules/**/*.md           # All .md files in rules and subdirectories
```

### Exclusion Patterns
```bash
**/*.md,!**/drafts/**   # All .md files except in drafts directories
**/*,!**/*.tmp          # All files except .tmp files
rules/*.md,!rules/old.md # All .md in rules except old.md
```

### Complex Patterns
```bash
{rules,guidelines}/*.md              # .md files in rules OR guidelines
rules/{coding,security}/*.md         # .md files in specific subdirectories
**/*.{md,txt,rst}                   # Multiple file extensions
```

## Pattern Implementation

### Pattern Matcher Interface
```go
type PatternMatcher interface {
    Match(pattern, path string) (bool, error)
    MatchMultiple(patterns []string, path string) (bool, error)
    FilterPaths(patterns []string, paths []string) ([]string, error)
}
```

### Glob Implementation
```go
type GlobMatcher struct {
    cache map[string]*regexp.Regexp
    mu    sync.RWMutex
}

func (g *GlobMatcher) Match(pattern, path string) (bool, error) {
    // Handle exclusion patterns
    if strings.HasPrefix(pattern, "!") {
        match, err := g.match(pattern[1:], path)
        return !match, err
    }

    return g.match(pattern, path)
}

func (g *GlobMatcher) match(pattern, path string) (bool, error) {
    // Use cached regex if available
    g.mu.RLock()
    regex, exists := g.cache[pattern]
    g.mu.RUnlock()

    if !exists {
        var err error
        regex, err = g.compilePattern(pattern)
        if err != nil {
            return false, err
        }

        g.mu.Lock()
        g.cache[pattern] = regex
        g.mu.Unlock()
    }

    return regex.MatchString(path), nil
}
```

### Pattern Compilation
```go
func (g *GlobMatcher) compilePattern(pattern string) (*regexp.Regexp, error) {
    // Convert glob pattern to regex
    regex := g.globToRegex(pattern)
    return regexp.Compile(regex)
}

func (g *GlobMatcher) globToRegex(pattern string) string {
    var result strings.Builder
    result.WriteString("^")

    for i, char := range pattern {
        switch char {
        case '*':
            if i+1 < len(pattern) && pattern[i+1] == '*' {
                // ** matches any number of directories
                result.WriteString(".*")
                i++ // Skip next *
            } else {
                // * matches within directory
                result.WriteString("[^/]*")
            }
        case '?':
            result.WriteString("[^/]")
        case '[':
            result.WriteString("[")
        case ']':
            result.WriteString("]")
        case '{':
            result.WriteString("(")
        case '}':
            result.WriteString(")")
        case ',':
            result.WriteString("|")
        case '.', '+', '^', '$', '(', ')', '|', '\\':
            result.WriteString("\\")
            result.WriteRune(char)
        default:
            result.WriteRune(char)
        }
    }

    result.WriteString("$")
    return result.String()
}
```

## Pattern Processing

### Multi-Pattern Matching
```go
func (g *GlobMatcher) MatchMultiple(patterns []string, path string) (bool, error) {
    var included bool
    var excluded bool

    for _, pattern := range patterns {
        if strings.HasPrefix(pattern, "!") {
            // Exclusion pattern
            match, err := g.Match(pattern, path)
            if err != nil {
                return false, err
            }
            if match {
                excluded = true
            }
        } else {
            // Inclusion pattern
            match, err := g.Match(pattern, path)
            if err != nil {
                return false, err
            }
            if match {
                included = true
            }
        }
    }

    // Include if matched by inclusion pattern and not excluded
    return included && !excluded, nil
}
```

### Path Filtering
```go
func (g *GlobMatcher) FilterPaths(patterns []string, paths []string) ([]string, error) {
    if len(patterns) == 0 {
        return paths, nil // No patterns = include all
    }

    var filtered []string
    for _, path := range paths {
        match, err := g.MatchMultiple(patterns, path)
        if err != nil {
            return nil, fmt.Errorf("pattern matching failed for %s: %w", path, err)
        }
        if match {
            filtered = append(filtered, path)
        }
    }

    return filtered, nil
}
```

## Registry Integration

### Git Registry Pattern Support
```go
func (g *GitRegistry) DownloadRulesetWithPatterns(ctx context.Context, name, version, destDir string, patterns []string) error {
    // Clone repository to temporary directory
    tempRepo, err := g.cloneRepository(ctx, name, version)
    if err != nil {
        return err
    }
    defer os.RemoveAll(tempRepo)

    // Find all files in repository
    allFiles, err := g.findAllFiles(tempRepo)
    if err != nil {
        return err
    }

    // Apply patterns to filter files
    matcher := NewGlobMatcher()
    filteredFiles, err := matcher.FilterPaths(patterns, allFiles)
    if err != nil {
        return err
    }

    // Copy filtered files to destination
    return g.copyFiles(filteredFiles, tempRepo, destDir)
}
```

### Pattern Validation
```go
func ValidatePatterns(patterns []string) error {
    matcher := NewGlobMatcher()

    for _, pattern := range patterns {
        // Test pattern compilation
        _, err := matcher.compilePattern(strings.TrimPrefix(pattern, "!"))
        if err != nil {
            return fmt.Errorf("invalid pattern '%s': %w", pattern, err)
        }
    }

    return nil
}
```

## Performance Optimizations

### Pattern Caching
```go
type PatternCache struct {
    compiled map[string]*regexp.Regexp
    results  map[string]map[string]bool // pattern -> path -> result
    mu       sync.RWMutex
    maxSize  int
}

func (c *PatternCache) GetResult(pattern, path string) (bool, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if pathResults, exists := c.results[pattern]; exists {
        if result, exists := pathResults[path]; exists {
            return result, true
        }
    }

    return false, false
}

func (c *PatternCache) SetResult(pattern, path string, result bool) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.results[pattern] == nil {
        c.results[pattern] = make(map[string]bool)
    }

    c.results[pattern][path] = result

    // Implement LRU eviction if cache is too large
    if len(c.results) > c.maxSize {
        c.evictOldest()
    }
}
```

### Batch Processing
```go
func (g *GlobMatcher) FilterPathsBatch(patterns []string, paths []string, batchSize int) ([]string, error) {
    var filtered []string
    var mu sync.Mutex
    var wg sync.WaitGroup

    // Process paths in batches
    for i := 0; i < len(paths); i += batchSize {
        end := i + batchSize
        if end > len(paths) {
            end = len(paths)
        }

        wg.Add(1)
        go func(batch []string) {
            defer wg.Done()

            var batchFiltered []string
            for _, path := range batch {
                if match, err := g.MatchMultiple(patterns, path); err == nil && match {
                    batchFiltered = append(batchFiltered, path)
                }
            }

            mu.Lock()
            filtered = append(filtered, batchFiltered...)
            mu.Unlock()
        }(paths[i:end])
    }

    wg.Wait()
    return filtered, nil
}
```

## Common Pattern Examples

### Documentation Files
```bash
# Include all documentation
**/*.md,**/*.rst,**/*.txt

# Exclude drafts and internal docs
**/*.md,!**/drafts/**,!**/internal/**

# Specific documentation directories
{docs,documentation}/**/*.md
```

### Code Rules
```bash
# All rule files
rules/**/*.md,guidelines/**/*.md

# Specific rule categories
rules/{coding,security,testing}/*.md

# Exclude deprecated rules
rules/**/*.md,!**/deprecated/**
```

### Configuration Files
```bash
# All config files
**/*.{json,yaml,yml,toml,ini}

# Exclude local configs
**/*.json,!**/*local*.json,!**/.*

# Specific config directories
{config,configs,configuration}/**/*
```

## Error Handling

### Pattern Errors
```go
type PatternError struct {
    Pattern string
    Path    string
    Cause   error
}

func (e *PatternError) Error() string {
    return fmt.Sprintf("pattern '%s' failed for path '%s': %v",
        e.Pattern, e.Path, e.Cause)
}

var (
    ErrInvalidPattern    = errors.New("invalid pattern syntax")
    ErrPatternTooComplex = errors.New("pattern too complex")
    ErrNoMatchingFiles   = errors.New("no files match patterns")
)
```

### Validation and Recovery
```go
func (g *GlobMatcher) MatchWithRecovery(pattern, path string) (bool, error) {
    defer func() {
        if r := recover(); r != nil {
            // Log pattern that caused panic
            log.Printf("Pattern matching panic: pattern=%s, path=%s, error=%v",
                pattern, path, r)
        }
    }()

    return g.Match(pattern, path)
}
```

## Testing Strategy

### Pattern Test Cases
```go
func TestPatternMatching(t *testing.T) {
    tests := []struct {
        pattern string
        path    string
        matches bool
    }{
        {"*.md", "README.md", true},
        {"*.md", "docs/README.md", false},
        {"**/*.md", "docs/README.md", true},
        {"rules/*.md", "rules/coding.md", true},
        {"rules/*.md", "rules/sub/coding.md", false},
        {"rules/**/*.md", "rules/sub/coding.md", true},
        {"!**/drafts/**", "docs/drafts/temp.md", true}, // exclusion
    }

    matcher := NewGlobMatcher()
    for _, test := range tests {
        result, err := matcher.Match(test.pattern, test.path)
        assert.NoError(t, err)
        assert.Equal(t, test.matches, result)
    }
}
```

### Performance Benchmarks
```go
func BenchmarkPatternMatching(b *testing.B) {
    matcher := NewGlobMatcher()
    patterns := []string{"**/*.md", "!**/drafts/**", "rules/**/*.md"}
    paths := generateTestPaths(1000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = matcher.FilterPaths(patterns, paths)
    }
}
```
