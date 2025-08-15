# File Patterns

Technical specification for ARM's pattern matching and file selection system.

## Overview

ARM uses glob patterns to selectively install files from registries, providing fine-grained control over which files are included in ruleset installations. The pattern system supports inclusion, exclusion, and complex matching scenarios.

## Pattern Syntax

### Basic Patterns
- `*.md` - All .md files in current directory
- `**/*.md` - All .md files recursively through subdirectories
- `rules/*.md` - All .md files directly in rules directory
- `rules/**/*.md` - All .md files in rules directory and subdirectories

### Exclusion Patterns
- `**/*.md,!**/drafts/**` - All .md files except those in drafts directories
- `**/*,!**/*.tmp` - All files except temporary files
- `rules/*.md,!rules/old.md` - All .md files in rules except old.md

### Complex Patterns
- `{rules,guidelines}/*.md` - .md files in either rules OR guidelines directories
- `rules/{coding,security}/*.md` - .md files in specific rule subdirectories
- `**/*.{md,txt,rst}` - Files with multiple extensions recursively

## Pattern Implementation

### Core Components

**Pattern Matcher Interface** - Defines the contract for pattern matching operations including single pattern matching, multiple pattern evaluation, and path filtering.

**Glob Matcher** - Primary implementation that converts glob patterns to regular expressions with caching for performance. Handles exclusion patterns (prefixed with `!`) by inverting match results.

**Pattern Compilation** - Converts glob syntax to regex by mapping special characters:
- `*` becomes `[^/]*` (matches within directory)
- `**` becomes `.*` (matches across directories)
- `?` becomes `[^/]` (single character except path separator)
- `{a,b}` becomes `(a|b)` (alternation)
- Escapes regex metacharacters appropriately

## Pattern Processing

### Multi-Pattern Logic

When multiple patterns are provided, ARM processes them with inclusion/exclusion precedence:

1. **Inclusion Phase** - Check if path matches any non-exclusion pattern
2. **Exclusion Phase** - Check if path matches any exclusion pattern (prefixed with `!`)
3. **Final Decision** - Include path only if it matches inclusion patterns AND doesn't match exclusion patterns

### Path Filtering

The filtering process applies patterns to file lists efficiently:
- Empty pattern list includes all files
- Each file path is evaluated against the complete pattern set
- Only matching paths are included in the final result
- Errors in pattern matching halt the entire operation

## Registry Integration

### Git Registry Support

Pattern matching integrates seamlessly with Git registries through a multi-step process:

1. **Repository Cloning** - Clone the target repository to a temporary directory
2. **File Discovery** - Recursively find all files in the cloned repository
3. **Pattern Application** - Apply user-specified patterns to filter the file list
4. **Selective Copy** - Copy only matching files to the destination directory

### Pattern Validation

Patterns are validated before use to prevent runtime errors:
- Each pattern is compiled to verify syntax correctness
- Exclusion prefixes are stripped before validation
- Invalid patterns generate descriptive error messages
- Validation occurs early in the installation process

## Performance Optimizations

### Pattern Caching

ARM implements multi-level caching to optimize pattern matching performance:

**Compiled Pattern Cache** - Stores compiled regular expressions to avoid repeated compilation overhead

**Result Cache** - Caches pattern-path match results with LRU eviction to handle memory constraints

**Thread Safety** - Uses read-write mutexes to allow concurrent read access while protecting write operations

### Batch Processing

For large file sets, ARM processes paths in parallel batches:
- Divides file lists into configurable batch sizes
- Processes batches concurrently using goroutines
- Aggregates results with mutex protection
- Maintains order independence for better parallelization

## Common Pattern Examples

### Documentation Files
- `**/*.md,**/*.rst,**/*.txt` - Include all documentation formats recursively
- `**/*.md,!**/drafts/**,!**/internal/**` - All markdown except drafts and internal docs
- `{docs,documentation}/**/*.md` - Markdown files from standard documentation directories

### Code Rules
- `rules/**/*.md,guidelines/**/*.md` - All rule files from multiple directories
- `rules/{coding,security,testing}/*.md` - Specific rule categories only
- `rules/**/*.md,!**/deprecated/**` - All rules except deprecated ones

### Configuration Files
- `**/*.{json,yaml,yml,toml,ini}` - All common configuration formats
- `**/*.json,!**/*local*.json,!**/.*` - JSON configs excluding local and hidden files
- `{config,configs,configuration}/**/*` - Files from standard config directories

## Error Handling

### Pattern Errors

ARM provides structured error handling for pattern operations:

**PatternError Type** - Captures the specific pattern, path, and underlying cause of failures

**Common Error Types**:
- `ErrInvalidPattern` - Malformed pattern syntax
- `ErrPatternTooComplex` - Pattern exceeds complexity limits
- `ErrNoMatchingFiles` - No files match the provided patterns

### Recovery Mechanisms

Pattern matching includes panic recovery to handle edge cases:
- Logs problematic patterns for debugging
- Prevents crashes from malformed regex compilation
- Provides graceful degradation for complex patterns

## Testing Strategy

### Test Coverage

Comprehensive testing covers multiple scenarios:
- **Basic Matching** - Single patterns against various path structures
- **Exclusion Logic** - Proper handling of negation patterns
- **Complex Patterns** - Alternation, multiple extensions, nested directories
- **Edge Cases** - Empty patterns, malformed syntax, boundary conditions

### Performance Validation

Benchmarking ensures pattern matching scales effectively:
- Tests with large file sets (1000+ paths)
- Measures caching effectiveness
- Validates concurrent processing performance
- Monitors memory usage patterns
