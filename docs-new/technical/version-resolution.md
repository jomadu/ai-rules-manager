# Version Resolution

Technical specification for ARM's semantic versioning and constraint resolution system.

## Overview

ARM supports flexible version resolution using semantic versioning constraints, enabling teams to specify version requirements while maintaining compatibility and stability.

## Supported Version Formats

### Exact Versions
```bash
arm install rules@1.2.3        # Exact version
arm install rules@v1.2.3       # With 'v' prefix
```

### Semantic Version Constraints
```bash
arm install rules@^1.2.0       # >=1.2.0 <2.0.0 (compatible)
arm install rules@~1.2.0       # >=1.2.0 <1.3.0 (patch-level)
arm install rules@>=1.1.0      # Greater than or equal
arm install rules@<2.0.0       # Less than
```

### Range Constraints
```bash
arm install rules@">=1.0.0 <2.0.0"    # Range specification
arm install rules@"1.x"               # Wildcard major
arm install rules@"1.2.x"             # Wildcard minor
```

### Special Versions
```bash
arm install rules@latest       # Latest stable version
arm install rules@main         # Branch name (Git only)
arm install rules@develop      # Development branch
```

## Version Resolution Algorithm

### Resolution Process
```go
func (r *Resolver) ResolveVersion(ctx context.Context, registry, ruleset, constraint string) (string, error) {
    // 1. Parse constraint
    c, err := r.parseConstraint(constraint)
    if err != nil {
        return "", err
    }

    // 2. Get available versions
    versions, err := r.getAvailableVersions(ctx, registry, ruleset)
    if err != nil {
        return "", err
    }

    // 3. Filter and sort
    candidates := r.filterVersions(versions, c)
    if len(candidates) == 0 {
        return "", fmt.Errorf("no versions match constraint %s", constraint)
    }

    // 4. Select best match
    return r.selectBestVersion(candidates, c), nil
}
```

### Constraint Parsing
```go
type Constraint struct {
    Type     ConstraintType  // Exact, Range, Caret, Tilde, etc.
    Version  *semver.Version
    Operator string         // >=, <, =, etc.
    Raw      string         // Original constraint string
}

func (r *Resolver) parseConstraint(constraint string) (*Constraint, error) {
    constraint = strings.TrimSpace(constraint)

    switch {
    case constraint == "latest":
        return &Constraint{Type: Latest}, nil
    case strings.HasPrefix(constraint, "^"):
        return r.parseCaretConstraint(constraint[1:])
    case strings.HasPrefix(constraint, "~"):
        return r.parseTildeConstraint(constraint[1:])
    case strings.HasPrefix(constraint, ">="):
        return r.parseOperatorConstraint(">=", constraint[2:])
    case strings.HasPrefix(constraint, "<="):
        return r.parseOperatorConstraint("<=", constraint[2:])
    case strings.HasPrefix(constraint, ">"):
        return r.parseOperatorConstraint(">", constraint[1:])
    case strings.HasPrefix(constraint, "<"):
        return r.parseOperatorConstraint("<", constraint[1:])
    case strings.Contains(constraint, " "):
        return r.parseRangeConstraint(constraint)
    default:
        return r.parseExactConstraint(constraint)
    }
}
```

## Constraint Types

### Caret Constraints (^)
Compatible within major version:
```go
func (r *Resolver) matchesCaret(version *semver.Version, base *semver.Version) bool {
    if version.Major() != base.Major() {
        return false
    }
    return version.GreaterThanOrEqual(base)
}

// Examples:
// ^1.2.3 matches: 1.2.3, 1.2.4, 1.3.0, 1.9.9
// ^1.2.3 excludes: 1.2.2, 2.0.0, 0.9.9
```

### Tilde Constraints (~)
Compatible within minor version:
```go
func (r *Resolver) matchesTilde(version *semver.Version, base *semver.Version) bool {
    if version.Major() != base.Major() || version.Minor() != base.Minor() {
        return false
    }
    return version.GreaterThanOrEqual(base)
}

// Examples:
// ~1.2.3 matches: 1.2.3, 1.2.4, 1.2.9
// ~1.2.3 excludes: 1.2.2, 1.3.0, 2.0.0
```

### Range Constraints
Multiple constraints combined:
```go
func (r *Resolver) parseRangeConstraint(constraint string) (*Constraint, error) {
    parts := strings.Fields(constraint)
    var constraints []*Constraint

    for _, part := range parts {
        c, err := r.parseConstraint(part)
        if err != nil {
            return nil, err
        }
        constraints = append(constraints, c)
    }

    return &Constraint{
        Type:        Range,
        Constraints: constraints,
        Raw:         constraint,
    }, nil
}

func (r *Resolver) matchesRange(version *semver.Version, constraints []*Constraint) bool {
    for _, c := range constraints {
        if !r.matches(version, c) {
            return false
        }
    }
    return true
}
```

## Registry-Specific Resolution

### Git Registries
```go
func (g *GitRegistry) ListVersions(ctx context.Context, ruleset string) ([]string, error) {
    // Get tags from Git repository
    tags, err := g.getTags(ctx, ruleset)
    if err != nil {
        return nil, err
    }

    // Filter semantic version tags
    var versions []string
    for _, tag := range tags {
        if semver.IsValid(tag) || semver.IsValid("v"+tag) {
            versions = append(versions, tag)
        }
    }

    // Add branches for non-semver versions
    branches, err := g.getBranches(ctx, ruleset)
    if err == nil {
        versions = append(versions, branches...)
    }

    return versions, nil
}
```

### S3 Registries
```go
func (s *S3Registry) ListVersions(ctx context.Context, ruleset string) ([]string, error) {
    // List objects with prefix
    prefix := fmt.Sprintf("%s/%s/", s.prefix, ruleset)
    objects, err := s.listObjects(ctx, prefix)
    if err != nil {
        return nil, err
    }

    // Extract version directories
    var versions []string
    for _, obj := range objects {
        if version := s.extractVersion(obj.Key, prefix); version != "" {
            versions = append(versions, version)
        }
    }

    return versions, nil
}
```

## Version Sorting and Selection

### Semantic Version Sorting
```go
func (r *Resolver) sortVersions(versions []string) []*semver.Version {
    var semVersions []*semver.Version

    for _, v := range versions {
        if sv, err := semver.NewVersion(v); err == nil {
            semVersions = append(semVersions, sv)
        }
    }

    // Sort in descending order (newest first)
    sort.Slice(semVersions, func(i, j int) bool {
        return semVersions[i].GreaterThan(semVersions[j])
    })

    return semVersions
}
```

### Best Version Selection
```go
func (r *Resolver) selectBestVersion(candidates []*semver.Version, constraint *Constraint) string {
    switch constraint.Type {
    case Latest:
        // Return newest version
        return candidates[0].String()
    case Exact:
        // Return exact match
        for _, v := range candidates {
            if v.Equal(constraint.Version) {
                return v.String()
            }
        }
    case Caret, Tilde, Range:
        // Return newest matching version
        for _, v := range candidates {
            if r.matches(v, constraint) {
                return v.String()
            }
        }
    }

    return ""
}
```

## Caching and Performance

### Version Cache
```go
type VersionCache struct {
    cache map[string]*CacheEntry
    ttl   time.Duration
    mu    sync.RWMutex
}

type CacheEntry struct {
    Versions  []string
    Timestamp time.Time
}

func (c *VersionCache) Get(key string) ([]string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, exists := c.cache[key]
    if !exists {
        return nil, false
    }

    // Check TTL
    if time.Since(entry.Timestamp) > c.ttl {
        return nil, false
    }

    return entry.Versions, true
}
```

### Cache Key Generation
```go
func (r *Resolver) generateCacheKey(registry, ruleset string) string {
    h := sha256.New()
    h.Write([]byte(registry))
    h.Write([]byte(ruleset))
    return hex.EncodeToString(h.Sum(nil))[:16]
}
```

## Error Handling

### Version Resolution Errors
```go
var (
    ErrNoVersionsFound     = errors.New("no versions found")
    ErrNoMatchingVersions  = errors.New("no versions match constraint")
    ErrInvalidConstraint   = errors.New("invalid version constraint")
    ErrInvalidVersion      = errors.New("invalid semantic version")
)

type VersionError struct {
    Registry   string
    Ruleset    string
    Constraint string
    Cause      error
}

func (e *VersionError) Error() string {
    return fmt.Sprintf("version resolution failed for %s/%s@%s: %v",
        e.Registry, e.Ruleset, e.Constraint, e.Cause)
}
```

### Fallback Strategies
```go
func (r *Resolver) ResolveVersionWithFallback(ctx context.Context, registry, ruleset, constraint string) (string, error) {
    // Try primary resolution
    version, err := r.ResolveVersion(ctx, registry, ruleset, constraint)
    if err == nil {
        return version, nil
    }

    // Fallback strategies
    switch constraint {
    case "latest":
        // Try "main" branch for Git registries
        if r.isGitRegistry(registry) {
            return "main", nil
        }
    default:
        // Try relaxing constraint
        if relaxed := r.relaxConstraint(constraint); relaxed != constraint {
            return r.ResolveVersion(ctx, registry, ruleset, relaxed)
        }
    }

    return "", err
}
```

## Testing and Validation

### Constraint Testing
```go
func TestConstraintMatching(t *testing.T) {
    tests := []struct {
        constraint string
        version    string
        matches    bool
    }{
        {"^1.2.0", "1.2.3", true},
        {"^1.2.0", "1.3.0", true},
        {"^1.2.0", "2.0.0", false},
        {"~1.2.0", "1.2.3", true},
        {"~1.2.0", "1.3.0", false},
        {">=1.0.0 <2.0.0", "1.5.0", true},
        {">=1.0.0 <2.0.0", "2.0.0", false},
    }

    resolver := NewResolver()
    for _, test := range tests {
        constraint, _ := resolver.parseConstraint(test.constraint)
        version, _ := semver.NewVersion(test.version)

        result := resolver.matches(version, constraint)
        assert.Equal(t, test.matches, result)
    }
}
```

### Performance Benchmarks
```go
func BenchmarkVersionResolution(b *testing.B) {
    resolver := NewResolver()
    versions := generateTestVersions(1000) // 1000 versions

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = resolver.selectBestVersion(versions, &Constraint{
            Type: Caret,
            Version: semver.MustParse("1.0.0"),
        })
    }
}
```
