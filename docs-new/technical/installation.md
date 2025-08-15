# Installation System

Technical specification for ARM's orchestrated installation workflow.

## Overview

The installation system coordinates downloading, caching, and deploying rulesets across multiple channels with atomic operations and rollback capabilities.

## Architecture

```
Installation Request
        ↓
   Orchestrator
        ↓
┌─────────────────┐
│   Pre-flight    │ ← Validation, permissions
│     Checks      │
└─────────────────┘
        ↓
┌─────────────────┐
│    Download     │ ← Registry interaction
│   & Extract     │
└─────────────────┘
        ↓
┌─────────────────┐
│  Multi-Channel  │ ← Parallel installation
│  Installation   │
└─────────────────┘
        ↓
┌─────────────────┐
│   Lock File     │ ← Version tracking
│    Update       │
└─────────────────┘
```

## Core Components

### Installation Request
```go
type InstallRequest struct {
    Registry        string
    Ruleset         string
    Version         string
    ResolvedVersion string   // Actual commit/version
    SourceFiles     []string
    Channels        []string
    Patterns        []string
}
```

### Installation Result
```go
type InstallResult struct {
    Registry     string
    Ruleset      string
    Version      string
    FilesCount   int
    Channels     []string
    Duration     time.Duration
}
```

## Orchestration Workflow

### 1. Pre-flight Validation
```go
func (o *Orchestrator) validateRequest(req *InstallRequest) error {
    // Validate registry exists
    if _, exists := o.config.Registries[req.Registry]; !exists {
        return fmt.Errorf("registry '%s' not found", req.Registry)
    }

    // Validate channels exist and are writable
    for _, channel := range req.Channels {
        if err := o.validateChannel(channel); err != nil {
            return fmt.Errorf("channel '%s': %w", channel, err)
        }
    }

    // Check for conflicts
    return o.checkConflicts(req)
}
```

### 2. Download and Extract
```go
func (o *Orchestrator) downloadRuleset(req *InstallRequest) ([]string, error) {
    // Create temporary directory
    tempDir, err := os.MkdirTemp("", "arm-install-*")
    if err != nil {
        return nil, err
    }
    defer os.RemoveAll(tempDir)

    // Download from registry
    reg, err := o.createRegistry(req.Registry)
    if err != nil {
        return nil, err
    }
    defer reg.Close()

    // Download with patterns if supported
    if pd, ok := reg.(registry.PatternDownloader); ok && len(req.Patterns) > 0 {
        err = pd.DownloadRulesetWithPatterns(ctx, req.Ruleset, req.Version, tempDir, req.Patterns)
    } else {
        err = reg.DownloadRuleset(ctx, req.Ruleset, req.Version, tempDir)
    }

    if err != nil {
        return nil, err
    }

    // Find downloaded files
    return o.findFiles(tempDir)
}
```

### 3. Multi-Channel Installation
```go
func (o *Orchestrator) installToChannels(req *InstallRequest, sourceFiles []string) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(req.Channels))

    for _, channel := range req.Channels {
        wg.Add(1)
        go func(ch string) {
            defer wg.Done()
            if err := o.installToChannel(req, sourceFiles, ch); err != nil {
                errChan <- fmt.Errorf("channel '%s': %w", ch, err)
            }
        }(channel)
    }

    wg.Wait()
    close(errChan)

    // Collect errors
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        return fmt.Errorf("installation failed: %v", errors)
    }

    return nil
}
```

## File Organization

### ARM Namespace Structure
```
channel-directory/
└── arm/                    # ARM namespace
    └── registry-name/      # Registry namespace
        └── ruleset-name/   # Ruleset files
            ├── file1.md
            ├── file2.md
            └── subdir/
                └── file3.md
```

### Installation Process
```go
func (i *Installer) installToChannel(req *InstallRequest, sourceFiles []string, channel string) error {
    channelConfig := i.config.Channels[channel]

    for _, dir := range channelConfig.Directories {
        targetDir := filepath.Join(dir, "arm", req.Registry, req.Ruleset)

        // Create target directory
        if err := os.MkdirAll(targetDir, 0755); err != nil {
            return err
        }

        // Copy files with structure preservation
        for _, sourceFile := range sourceFiles {
            if err := i.copyFile(sourceFile, targetDir); err != nil {
                return err
            }
        }
    }

    return nil
}
```

## Lock File Management

### Lock File Structure
```json
{
  "rulesets": {
    "registry-name": {
      "ruleset-name": {
        "version": "1.2.3",
        "resolved": "abc123def456...",
        "registry": "registry-name",
        "type": "git",
        "installed": "2024-01-15T10:30:00Z"
      }
    }
  }
}
```

### Update Process
```go
func (l *LockManager) UpdateLock(req *InstallRequest, result *InstallResult) error {
    l.mu.Lock()
    defer l.mu.Unlock()

    if l.lockFile.Rulesets == nil {
        l.lockFile.Rulesets = make(map[string]map[string]LockedRuleset)
    }

    if l.lockFile.Rulesets[req.Registry] == nil {
        l.lockFile.Rulesets[req.Registry] = make(map[string]LockedRuleset)
    }

    l.lockFile.Rulesets[req.Registry][req.Ruleset] = LockedRuleset{
        Version:   req.Version,
        Resolved:  req.ResolvedVersion,
        Registry:  req.Registry,
        Type:      l.getRegistryType(req.Registry),
        Installed: time.Now().Format(time.RFC3339),
    }

    return l.saveLockFile()
}
```

## Error Handling and Rollback

### Atomic Installation
```go
func (o *Orchestrator) Install(req *InstallRequest) (*InstallResult, error) {
    // Create rollback context
    rollback := NewRollbackContext()
    defer rollback.Execute()

    // Pre-flight checks
    if err := o.validateRequest(req); err != nil {
        return nil, err
    }

    // Download and extract
    sourceFiles, err := o.downloadRuleset(req)
    if err != nil {
        return nil, err
    }
    rollback.AddCleanup(func() { os.RemoveAll(filepath.Dir(sourceFiles[0])) })

    // Install to channels
    if err := o.installToChannels(req, sourceFiles); err != nil {
        return nil, err
    }
    rollback.AddCleanup(func() { o.removeFromChannels(req) })

    // Update lock file
    result := &InstallResult{
        Registry:   req.Registry,
        Ruleset:    req.Ruleset,
        Version:    req.Version,
        FilesCount: len(sourceFiles),
        Channels:   req.Channels,
    }

    if err := o.lockManager.UpdateLock(req, result); err != nil {
        return nil, err
    }

    // Success - disable rollback
    rollback.Disable()
    return result, nil
}
```

### Rollback Context
```go
type RollbackContext struct {
    cleanupFuncs []func()
    disabled     bool
    mu           sync.Mutex
}

func (r *RollbackContext) AddCleanup(fn func()) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.cleanupFuncs = append(r.cleanupFuncs, fn)
}

func (r *RollbackContext) Execute() {
    r.mu.Lock()
    defer r.mu.Unlock()

    if r.disabled {
        return
    }

    // Execute cleanup functions in reverse order
    for i := len(r.cleanupFuncs) - 1; i >= 0; i-- {
        r.cleanupFuncs[i]()
    }
}
```

## Performance Optimizations

### Parallel Channel Installation
- Concurrent installation to multiple channels
- Shared source files to reduce I/O
- Error aggregation for comprehensive reporting

### File Operation Optimization
```go
func (i *Installer) copyFile(src, dstDir string) error {
    // Use hard links when possible (same filesystem)
    if i.canHardLink(src, dstDir) {
        return os.Link(src, filepath.Join(dstDir, filepath.Base(src)))
    }

    // Fall back to copy
    return i.copyFileContents(src, dstDir)
}
```

### Cache Integration
- Content-based caching reduces downloads
- Metadata caching speeds up version resolution
- Intelligent cache warming for common operations

## Validation and Safety

### Pre-installation Checks
- Registry connectivity
- Channel directory permissions
- Disk space availability
- Version constraint validation

### Conflict Detection
```go
func (o *Orchestrator) checkConflicts(req *InstallRequest) error {
    // Check if different version already installed
    if locked, exists := o.getLocked(req.Registry, req.Ruleset); exists {
        if locked.Version != req.Version {
            return fmt.Errorf("version conflict: %s installed, %s requested",
                locked.Version, req.Version)
        }
    }

    // Check for file conflicts in channels
    return o.checkFileConflicts(req)
}
```

### Integrity Verification
- Checksum validation for downloaded content
- File count verification
- Directory structure validation
