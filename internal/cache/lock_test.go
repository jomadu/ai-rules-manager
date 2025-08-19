package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewRegistryLock(t *testing.T) {
	cacheRoot := t.TempDir()
	registryKey := "test-registry"
	timeout := 5 * time.Second

	lock := NewRegistryLock(cacheRoot, registryKey, timeout)

	expectedPath := filepath.Join(cacheRoot, "locks", "test-registry.lock")
	if lock.lockPath != expectedPath {
		t.Errorf("Expected lock path %s, got %s", expectedPath, lock.lockPath)
	}

	if lock.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, lock.timeout)
	}

	if lock.acquired {
		t.Error("Expected lock to not be acquired initially")
	}
}

func TestRegistryLock_AcquireAndRelease(t *testing.T) {
	cacheRoot := t.TempDir()
	registryKey := "test-registry"
	timeout := 1 * time.Second

	lock := NewRegistryLock(cacheRoot, registryKey, timeout)

	// Test acquire
	err := lock.Acquire("test-operation")
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	if !lock.acquired {
		t.Error("Expected lock to be acquired")
	}

	// Verify lock file exists
	if _, err := os.Stat(lock.lockPath); os.IsNotExist(err) {
		t.Error("Expected lock file to exist")
	}

	// Test release
	err = lock.Release()
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	if lock.acquired {
		t.Error("Expected lock to not be acquired after release")
	}

	// Verify lock file is removed
	if _, err := os.Stat(lock.lockPath); !os.IsNotExist(err) {
		t.Error("Expected lock file to be removed")
	}
}

func TestRegistryLock_AcquireTimeout(t *testing.T) {
	cacheRoot := t.TempDir()
	registryKey := "test-registry"
	timeout := 200 * time.Millisecond

	// Create first lock
	lock1 := NewRegistryLock(cacheRoot, registryKey, timeout)
	err := lock1.Acquire("operation1")
	if err != nil {
		t.Fatalf("Failed to acquire first lock: %v", err)
	}
	defer lock1.Release()

	// Try to acquire second lock (should timeout)
	lock2 := NewRegistryLock(cacheRoot, registryKey, timeout)
	start := time.Now()
	err = lock2.Acquire("operation2")
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected second lock acquisition to fail")
	}

	// Should timeout around the specified duration
	if elapsed < timeout || elapsed > timeout+100*time.Millisecond {
		t.Errorf("Expected timeout around %v, got %v", timeout, elapsed)
	}
}

func TestRegistryLock_CleanupStaleLock(t *testing.T) {
	cacheRoot := t.TempDir()
	registryKey := "test-registry"
	timeout := 1 * time.Second

	lock := NewRegistryLock(cacheRoot, registryKey, timeout)

	// Create lock directory
	err := os.MkdirAll(filepath.Dir(lock.lockPath), 0o755)
	if err != nil {
		t.Fatalf("Failed to create lock directory: %v", err)
	}

	// Create a stale lock (2 hours old)
	staleLockInfo := &LockInfo{
		PID:       12345,
		Hostname:  "test-host",
		CreatedAt: time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339),
		Operation: "stale-operation",
	}

	data, err := json.MarshalIndent(staleLockInfo, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal stale lock info: %v", err)
	}

	err = os.WriteFile(lock.lockPath, data, 0o644)
	if err != nil {
		t.Fatalf("Failed to write stale lock file: %v", err)
	}

	// Try to acquire lock (should clean up stale lock and succeed)
	err = lock.Acquire("new-operation")
	if err != nil {
		t.Fatalf("Failed to acquire lock after stale cleanup: %v", err)
	}
	defer lock.Release()

	if !lock.acquired {
		t.Error("Expected lock to be acquired after stale cleanup")
	}
}

func TestRegistryLock_CleanupRecentLock(t *testing.T) {
	cacheRoot := t.TempDir()
	registryKey := "test-registry"
	timeout := 200 * time.Millisecond

	lock := NewRegistryLock(cacheRoot, registryKey, timeout)

	// Create lock directory
	err := os.MkdirAll(filepath.Dir(lock.lockPath), 0o755)
	if err != nil {
		t.Fatalf("Failed to create lock directory: %v", err)
	}

	// Create a recent lock (5 minutes old)
	recentLockInfo := &LockInfo{
		PID:       12345,
		Hostname:  "test-host",
		CreatedAt: time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339),
		Operation: "recent-operation",
	}

	data, err := json.MarshalIndent(recentLockInfo, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal recent lock info: %v", err)
	}

	err = os.WriteFile(lock.lockPath, data, 0o644)
	if err != nil {
		t.Fatalf("Failed to write recent lock file: %v", err)
	}

	// Try to acquire lock (should timeout because lock is not stale)
	err = lock.Acquire("new-operation")
	if err == nil {
		t.Error("Expected lock acquisition to fail for recent lock")
		lock.Release()
	}
}

func TestInitializeCache(t *testing.T) {
	cacheRoot := t.TempDir()

	err := InitializeCache(cacheRoot)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Check that directories were created
	expectedDirs := []string{
		cacheRoot,
		filepath.Join(cacheRoot, "locks"),
		filepath.Join(cacheRoot, "registries"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}

	// Cache initialization complete - directories created
}

func TestCleanupStaleLocks(t *testing.T) {
	cacheRoot := t.TempDir()
	locksDir := filepath.Join(cacheRoot, "locks")

	// Create locks directory
	err := os.MkdirAll(locksDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create locks directory: %v", err)
	}

	// Create stale lock
	staleLockPath := filepath.Join(locksDir, "stale.lock")
	staleLockInfo := &LockInfo{
		PID:       12345,
		Hostname:  "test-host",
		CreatedAt: time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339),
		Operation: "stale-operation",
	}
	data, _ := json.MarshalIndent(staleLockInfo, "", "  ")
	err = os.WriteFile(staleLockPath, data, 0o644)
	if err != nil {
		t.Fatalf("Failed to write stale lock: %v", err)
	}

	// Create recent lock
	recentLockPath := filepath.Join(locksDir, "recent.lock")
	recentLockInfo := &LockInfo{
		PID:       12346,
		Hostname:  "test-host",
		CreatedAt: time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339),
		Operation: "recent-operation",
	}
	data, _ = json.MarshalIndent(recentLockInfo, "", "  ")
	err = os.WriteFile(recentLockPath, data, 0o644)
	if err != nil {
		t.Fatalf("Failed to write recent lock: %v", err)
	}

	// Create invalid lock
	invalidLockPath := filepath.Join(locksDir, "invalid.lock")
	err = os.WriteFile(invalidLockPath, []byte("invalid json"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write invalid lock: %v", err)
	}

	// Run cleanup
	err = CleanupStaleLocks(cacheRoot)
	if err != nil {
		t.Fatalf("Failed to cleanup stale locks: %v", err)
	}

	// Verify stale lock was removed
	if _, err := os.Stat(staleLockPath); !os.IsNotExist(err) {
		t.Error("Expected stale lock to be removed")
	}

	// Verify recent lock still exists
	if _, err := os.Stat(recentLockPath); os.IsNotExist(err) {
		t.Error("Expected recent lock to still exist")
	}

	// Verify invalid lock was removed
	if _, err := os.Stat(invalidLockPath); !os.IsNotExist(err) {
		t.Error("Expected invalid lock to be removed")
	}
}

func TestCleanupStaleLocks_NoLocksDirectory(t *testing.T) {
	cacheRoot := t.TempDir()

	// Run cleanup on non-existent locks directory
	err := CleanupStaleLocks(cacheRoot)
	if err != nil {
		t.Fatalf("Expected no error when locks directory doesn't exist, got: %v", err)
	}
}

func TestRegistryLock_ReleaseNotAcquired(t *testing.T) {
	cacheRoot := t.TempDir()
	registryKey := "test-registry"
	timeout := 1 * time.Second

	lock := NewRegistryLock(cacheRoot, registryKey, timeout)

	// Try to release without acquiring
	err := lock.Release()
	if err != nil {
		t.Errorf("Expected no error when releasing non-acquired lock, got: %v", err)
	}
}
