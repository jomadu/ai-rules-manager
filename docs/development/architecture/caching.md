# Cache Architecture

Design and implementation of ARM's caching system.

## Cache Structure

### Global Cache Directory

```
~/.arm/cache/
├── packages/           # Downloaded package archives
│   └── {registry}/
│       └── {package}/
│           └── {version}/
│               └── package.tar.gz
├── registry/           # Registry metadata and version lists
│   └── {registry}/
│       ├── metadata.json
│       └── versions.json
└── backups/           # Installation backups
    └── {package}/
        └── {version}/
            └── {target}/
```

## Cache Types

### Package Cache

**Purpose**: Store downloaded package archives to avoid re-downloading

**Key Format**: `package:{registry}:{package}:{version}`

**Storage**: Compressed tar.gz files

**Lifecycle**: Permanent (packages are immutable)

### Registry Cache

**Purpose**: Cache registry metadata and version lists

**Key Format**: `registry:{registry}:{type}:{package}`

**Storage**: JSON files

**Lifecycle**: TTL-based expiration

### Backup Cache

**Purpose**: Store previous installations for rollback

**Key Format**: `backup:{package}:{version}:{target}`

**Storage**: Directory structure mirrors installation

**Lifecycle**: Cleaned up after successful updates

## Cache Interface

### Core Interface

```go
type Cache interface {
    Get(key string) ([]byte, error)
    Set(key string, data []byte) error
    Delete(key string) error
    Clear() error
    Size() (int64, error)
}
```

### Implementation

```go
type FileCache struct {
    baseDir string
    mutex   sync.RWMutex
}

func (c *FileCache) Get(key string) ([]byte, error) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    path := c.keyToPath(key)
    return os.ReadFile(path)
}
```

## Cache Policies

### TTL Policy

```go
type TTLPolicy struct {
    DefaultTTL time.Duration
    TypeTTL    map[string]time.Duration
}

func (p *TTLPolicy) IsExpired(entry *CacheEntry) bool {
    ttl := p.getTTL(entry.Type)
    return time.Since(entry.Created) > ttl
}
```

### Size Policy

```go
type SizePolicy struct {
    MaxSize    int64
    MaxEntries int
}

func (p *SizePolicy) ShouldEvict(cache *Cache) bool {
    size, _ := cache.Size()
    return size > p.MaxSize
}
```

## Cache Operations

### Cache Population

```go
func (i *Installer) installPackage(name, version string) error {
    // Check cache first
    cacheKey := fmt.Sprintf("package:%s:%s:%s", i.registry.Name(), name, version)

    if data := i.cache.Get(cacheKey); data != nil {
        return i.installFromCache(data)
    }

    // Download and cache
    reader, err := i.registry.Download(name, version)
    if err != nil {
        return err
    }

    data, err := io.ReadAll(reader)
    if err != nil {
        return err
    }

    // Cache for future use
    i.cache.Set(cacheKey, data)

    return i.installFromData(data)
}
```

### Cache Invalidation

```go
func (c *Cache) InvalidateRegistry(registryName string) error {
    pattern := fmt.Sprintf("registry:%s:*", registryName)
    return c.DeletePattern(pattern)
}

func (c *Cache) InvalidatePackage(registry, package string) error {
    // Don't invalidate package cache (immutable)
    // Only invalidate registry metadata
    key := fmt.Sprintf("registry:%s:metadata:%s", registry, package)
    return c.Delete(key)
}
```

## Performance Optimizations

### Concurrent Access

```go
type SafeCache struct {
    cache Cache
    mutex sync.RWMutex
}

func (c *SafeCache) Get(key string) ([]byte, error) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    return c.cache.Get(key)
}
```

### Batch Operations

```go
func (c *Cache) SetBatch(entries map[string][]byte) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    for key, data := range entries {
        if err := c.set(key, data); err != nil {
            return err
        }
    }
    return nil
}
```

### Compression

```go
func (c *FileCache) compress(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    gz := gzip.NewWriter(&buf)

    if _, err := gz.Write(data); err != nil {
        return nil, err
    }

    if err := gz.Close(); err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}
```

## Cache Maintenance

### Cleanup Strategies

**Time-based Cleanup**:
```go
func (c *Cache) CleanExpired() error {
    return c.walkCache(func(entry *CacheEntry) error {
        if c.policy.IsExpired(entry) {
            return c.Delete(entry.Key)
        }
        return nil
    })
}
```

**Size-based Cleanup**:
```go
func (c *Cache) CleanBySize() error {
    if !c.policy.ShouldEvict(c) {
        return nil
    }

    // Remove oldest entries until under size limit
    entries := c.listEntriesByAge()
    for _, entry := range entries {
        c.Delete(entry.Key)
        if !c.policy.ShouldEvict(c) {
            break
        }
    }
    return nil
}
```

### Cache Statistics

```go
type CacheStats struct {
    Hits        int64
    Misses      int64
    Size        int64
    Entries     int
    HitRatio    float64
    LastCleanup time.Time
}

func (c *Cache) Stats() *CacheStats {
    return &CacheStats{
        Hits:     atomic.LoadInt64(&c.hits),
        Misses:   atomic.LoadInt64(&c.misses),
        Size:     c.size(),
        Entries:  c.count(),
        HitRatio: c.hitRatio(),
    }
}
```

## Configuration

### Cache Configuration

```ini
[cache]
directory = ~/.arm/cache
maxSize = 1GB
maxEntries = 1000
defaultTTL = 1h

[cache.registry]
ttl = 15m

[cache.metadata]
ttl = 5m
```

### Environment Variables

```bash
# Override cache directory
export ARM_CACHE_DIR=/tmp/arm-cache

# Disable caching
export ARM_NO_CACHE=1

# Cache debug mode
export ARM_CACHE_DEBUG=1
```

## Security Considerations

### Safe File Operations

- Atomic writes using temporary files
- Path traversal prevention
- Permission checking
- Safe cleanup of temporary files

### Cache Isolation

- User-specific cache directories
- No shared cache between users
- Proper file permissions
- Secure temporary file handling
