package cache

import (
	"container/list"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache represents the ARM caching system
type Cache struct {
	mu       sync.RWMutex
	basePath string
	maxSize  int64
	ttl      time.Duration
	lru      *list.List
	items    map[string]*list.Element
}

// CacheItem represents a cached item with TTL
type CacheItem struct {
	key       string
	path      string
	size      int64
	cachedAt  time.Time
	accessedAt time.Time
}

// CacheInfo represents cache metadata
type CacheInfo struct {
	RegistryURL    string    `json:"registry_url"`
	LastAccessed   time.Time `json:"last_accessed"`
	TotalSizeBytes int64     `json:"total_size_bytes"`
}

// VersionCache represents cached version data
type VersionCache struct {
	CachedAt    time.Time            `json:"cached_at"`
	TTLSeconds  int                  `json:"ttl_seconds"`
	Rulesets    map[string][]string  `json:"rulesets"`
}

// MetadataCache represents cached metadata
type MetadataCache struct {
	CachedAt    time.Time                    `json:"cached_at"`
	TTLSeconds  int                          `json:"ttl_seconds"`
	Rulesets    map[string]map[string]string `json:"rulesets"`
}

// New creates a new cache instance
func New(basePath string, maxSize int64, ttl time.Duration) *Cache {
	return &Cache{
		basePath: basePath,
		maxSize:  maxSize,
		ttl:      ttl,
		lru:      list.New(),
		items:    make(map[string]*list.Element),
	}
}

// GetVersions retrieves cached version data for a registry
func (c *Cache) GetVersions(registryName string) (*VersionCache, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := filepath.Join(c.basePath, "registries", registryName, "versions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var versions VersionCache
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, err
	}

	// Check TTL
	if time.Since(versions.CachedAt) > time.Duration(versions.TTLSeconds)*time.Second {
		return nil, fmt.Errorf("cache expired")
	}

	return &versions, nil
}

// SetVersions stores version data in cache
func (c *Cache) SetVersions(registryName string, rulesets map[string][]string, ttlSeconds int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	registryDir := filepath.Join(c.basePath, "registries", registryName)
	if err := os.MkdirAll(registryDir, 0o755); err != nil {
		return err
	}

	versions := VersionCache{
		CachedAt:   time.Now(),
		TTLSeconds: ttlSeconds,
		Rulesets:   rulesets,
	}

	data, err := json.Marshal(versions)
	if err != nil {
		return err
	}

	path := filepath.Join(registryDir, "versions.json")
	return os.WriteFile(path, data, 0o644)
}

// GetMetadata retrieves cached metadata for a registry
func (c *Cache) GetMetadata(registryName string) (*MetadataCache, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := filepath.Join(c.basePath, "registries", registryName, "metadata.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata MetadataCache
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	// Check TTL
	if time.Since(metadata.CachedAt) > time.Duration(metadata.TTLSeconds)*time.Second {
		return nil, fmt.Errorf("cache expired")
	}

	return &metadata, nil
}

// SetMetadata stores metadata in cache
func (c *Cache) SetMetadata(registryName string, rulesets map[string]map[string]string, ttlSeconds int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	registryDir := filepath.Join(c.basePath, "registries", registryName)
	if err := os.MkdirAll(registryDir, 0o755); err != nil {
		return err
	}

	metadata := MetadataCache{
		CachedAt:   time.Now(),
		TTLSeconds: ttlSeconds,
		Rulesets:   rulesets,
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	path := filepath.Join(registryDir, "metadata.json")
	return os.WriteFile(path, data, 0o644)
}

// GetRuleset retrieves a cached ruleset file
func (c *Cache) GetRuleset(registryName, rulesetName, version string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.basePath, "registries", registryName, "rulesets", rulesetName, version, "ruleset.tar.gz")
	if _, err := os.Stat(path); err != nil {
		return "", err
	}

	// Update access time for LRU
	c.updateAccess(path)
	return path, nil
}

// SetRuleset stores a ruleset file in cache
func (c *Cache) SetRuleset(registryName, rulesetName, version, filePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	rulesetDir := filepath.Join(c.basePath, "registries", registryName, "rulesets", rulesetName, version)
	if err := os.MkdirAll(rulesetDir, 0o755); err != nil {
		return err
	}

	destPath := filepath.Join(rulesetDir, "ruleset.tar.gz")

	// Copy file
	srcData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(destPath, srcData, 0o644); err != nil {
		return err
	}

	c.addItem(destPath, int64(len(srcData)))
	return nil
}

// Clean removes all cached data
func (c *Cache) Clean() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := os.RemoveAll(c.basePath); err != nil {
		return err
	}

	c.lru.Init()
	c.items = make(map[string]*list.Element)
	return nil
}

// Size returns the current cache size in bytes
func (c *Cache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sizeInternal()
}

// sizeInternal returns the current cache size without acquiring locks
func (c *Cache) sizeInternal() int64 {
	var total int64
	for e := c.lru.Front(); e != nil; e = e.Next() {
		item := e.Value.(*CacheItem)
		total += item.size
	}
	return total
}

// addItem adds an item to the cache and handles LRU eviction
func (c *Cache) addItem(path string, size int64) {
	item := &CacheItem{
		key:        path,
		path:       path,
		size:       size,
		cachedAt:   time.Now(),
		accessedAt: time.Now(),
	}

	// Remove existing item if present
	if elem, exists := c.items[path]; exists {
		c.lru.Remove(elem)
		delete(c.items, path)
	}

	// Add new item
	elem := c.lru.PushFront(item)
	c.items[path] = elem

	// Evict if over size limit
	c.evictIfNeeded()
}

// updateAccess updates the access time and moves item to front
func (c *Cache) updateAccess(path string) {
	if elem, exists := c.items[path]; exists {
		item := elem.Value.(*CacheItem)
		item.accessedAt = time.Now()
		c.lru.MoveToFront(elem)
	}
}

// evictIfNeeded removes least recently used items if over size limit
func (c *Cache) evictIfNeeded() {
	for c.sizeInternal() > c.maxSize && c.lru.Len() > 0 {
		elem := c.lru.Back()
		if elem == nil {
			break
		}

		item := elem.Value.(*CacheItem)

		// Remove file
		_ = os.Remove(item.path)

		// Remove from LRU
		c.lru.Remove(elem)
		delete(c.items, item.key)
	}
}