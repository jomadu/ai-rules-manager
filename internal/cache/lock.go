package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LockInfo represents information about a cache lock
type LockInfo struct {
	PID       int    `json:"pid"`
	Hostname  string `json:"hostname"`
	CreatedAt string `json:"created_at"`
	Operation string `json:"operation"`
}

// RegistryLock manages locking for registry cache operations
type RegistryLock struct {
	lockPath string
	timeout  time.Duration
	lockInfo *LockInfo
	acquired bool
}

// NewRegistryLock creates a new registry lock
func NewRegistryLock(cacheRoot, registryKey string, timeout time.Duration) *RegistryLock {
	lockPath := filepath.Join(cacheRoot, "locks", fmt.Sprintf("%s.lock", registryKey))
	return &RegistryLock{
		lockPath: lockPath,
		timeout:  timeout,
		acquired: false,
	}
}

// Acquire attempts to acquire the registry lock with timeout
func (l *RegistryLock) Acquire(operation string) error {
	// Clean up stale locks first
	if err := l.cleanupStaleLock(); err != nil {
		return fmt.Errorf("failed to cleanup stale lock: %w", err)
	}

	// Create lock directory
	if err := os.MkdirAll(filepath.Dir(l.lockPath), 0o755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Try to acquire lock with timeout
	start := time.Now()
	for time.Since(start) < l.timeout {
		if err := l.tryAcquire(operation); err == nil {
			l.acquired = true
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("failed to acquire lock within timeout %v", l.timeout)
}

// Release releases the registry lock
func (l *RegistryLock) Release() error {
	if !l.acquired {
		return nil
	}

	if err := os.Remove(l.lockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	l.acquired = false
	return nil
}

// tryAcquire attempts to acquire the lock once
func (l *RegistryLock) tryAcquire(operation string) error {
	// Check if lock already exists
	if _, err := os.Stat(l.lockPath); err == nil {
		return fmt.Errorf("lock already exists")
	}

	// Create lock info
	hostname, _ := os.Hostname()
	lockInfo := &LockInfo{
		PID:       os.Getpid(),
		Hostname:  hostname,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Operation: operation,
	}

	// Write lock file atomically
	data, err := json.MarshalIndent(lockInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock info: %w", err)
	}

	tempPath := l.lockPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp lock file: %w", err)
	}

	if err := os.Rename(tempPath, l.lockPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	l.lockInfo = lockInfo
	return nil
}

// cleanupStaleLock removes stale locks based on age
func (l *RegistryLock) cleanupStaleLock() error {
	lockInfo, err := l.readLockInfo()
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No lock to clean up
		}
		return err
	}

	createdAt, err := time.Parse(time.RFC3339, lockInfo.CreatedAt)
	if err != nil {
		// Invalid timestamp, remove the lock
		return os.Remove(l.lockPath)
	}

	// Remove locks older than 1 hour (stale lock threshold)
	if time.Since(createdAt) > time.Hour {
		return os.Remove(l.lockPath)
	}

	return nil
}

// readLockInfo reads lock information from the lock file
func (l *RegistryLock) readLockInfo() (*LockInfo, error) {
	data, err := os.ReadFile(l.lockPath)
	if err != nil {
		return nil, err
	}

	var lockInfo LockInfo
	if err := json.Unmarshal(data, &lockInfo); err != nil {
		return nil, fmt.Errorf("failed to parse lock info: %w", err)
	}

	return &lockInfo, nil
}

// InitializeCache ensures cache directory structure exists and is properly configured
func InitializeCache(cacheRoot string) error {
	// Create main cache directory
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return fmt.Errorf("failed to create cache root: %w", err)
	}

	// Create subdirectories
	subdirs := []string{"locks", "registries"}
	for _, subdir := range subdirs {
		path := filepath.Join(cacheRoot, subdir)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("failed to create cache subdirectory %s: %w", subdir, err)
		}
	}

	// Load or create cache configuration
	_, err := LoadCacheConfig(cacheRoot)
	if err != nil {
		return fmt.Errorf("failed to initialize cache config: %w", err)
	}

	return nil
}

// CleanupStaleLocks removes all stale lock files from the cache
func CleanupStaleLocks(cacheRoot string) error {
	locksDir := filepath.Join(cacheRoot, "locks")

	// Check if locks directory exists
	if _, err := os.Stat(locksDir); os.IsNotExist(err) {
		return nil // No locks directory, nothing to clean
	}

	entries, err := os.ReadDir(locksDir)
	if err != nil {
		return fmt.Errorf("failed to read locks directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".lock" {
			lockPath := filepath.Join(locksDir, entry.Name())
			if err := cleanupSingleStaleLock(lockPath); err != nil {
				// Log error but continue with other locks
				continue
			}
		}
	}

	return nil
}

// cleanupSingleStaleLock removes a single stale lock file
func cleanupSingleStaleLock(lockPath string) error {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return err
	}

	var lockInfo LockInfo
	if err := json.Unmarshal(data, &lockInfo); err != nil {
		// Invalid lock file, remove it
		return os.Remove(lockPath)
	}

	createdAt, err := time.Parse(time.RFC3339, lockInfo.CreatedAt)
	if err != nil {
		// Invalid timestamp, remove the lock
		return os.Remove(lockPath)
	}

	// Remove locks older than 1 hour
	if time.Since(createdAt) > time.Hour {
		return os.Remove(lockPath)
	}

	return nil
}
