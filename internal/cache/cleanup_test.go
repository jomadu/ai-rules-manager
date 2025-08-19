package cache

import (
	"path/filepath"
	"testing"
	"time"
)

func TestGitCacheManager_Cleanup_TTL(t *testing.T) {
	cacheRoot := t.TempDir()
	manager := NewGitCacheManager(cacheRoot)

	// Store some test data
	files := map[string][]byte{
		"rule1.md": []byte("content1"),
		"rule2.md": []byte("content2"),
	}

	err := manager.StoreRuleset("https://github.com/test/repo", []string{"*.md"}, "abc123", files)
	if err != nil {
		t.Fatalf("Failed to store ruleset: %v", err)
	}

	// Manually update access time to be old
	registryKey := GenerateRegistryKey("git", "https://github.com/test/repo")
	registryPath := filepath.Join(cacheRoot, "registries", registryKey)
	
	index, err := LoadRegistryIndex(registryPath)
	if err != nil {
		t.Fatalf("Failed to load registry index: %v", err)
	}

	// Set access time to 2 hours ago
	oldTime := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	for _, rulesetCache := range index.Rulesets {
		rulesetCache.LastAccessedOn = oldTime
		for _, versionCache := range rulesetCache.Versions {
			versionCache.LastAccessedOn = oldTime
		}
	}
	
	err = SaveRegistryIndex(registryPath, index)
	if err != nil {
		t.Fatalf("Failed to save registry index: %v", err)
	}

	// Run cleanup with 1 hour TTL
	err = manager.Cleanup(1*time.Hour, 0)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify data was removed
	_, err = manager.GetRuleset("https://github.com/test/repo", []string{"*.md"}, "abc123")
	if err == nil {
		t.Error("Expected cached data to be removed after TTL cleanup")
	}
}

func TestGitCacheManager_Cleanup_Size(t *testing.T) {
	cacheRoot := t.TempDir()
	manager := NewGitCacheManager(cacheRoot)

	// Store large test data
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	files := map[string][]byte{
		"large_file.md": largeContent,
	}

	err := manager.StoreRuleset("https://github.com/test/repo", []string{"*.md"}, "abc123", files)
	if err != nil {
		t.Fatalf("Failed to store ruleset: %v", err)
	}

	// Run cleanup with small size limit (500KB)
	err = manager.Cleanup(0, 500*1024)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify data was removed due to size limit
	_, err = manager.GetRuleset("https://github.com/test/repo", []string{"*.md"}, "abc123")
	if err == nil {
		t.Error("Expected cached data to be removed after size cleanup")
	}
}

func TestRulesetCacheManager_Cleanup_TTL(t *testing.T) {
	cacheRoot := t.TempDir()
	manager := NewRulesetCacheManager(cacheRoot)

	// Store some test data
	files := map[string][]byte{
		"rule1.md": []byte("content1"),
		"rule2.md": []byte("content2"),
	}

	err := manager.StoreRuleset("https://registry.example.com", "test-rules", "1.0.0", files)
	if err != nil {
		t.Fatalf("Failed to store ruleset: %v", err)
	}

	// Manually update access time to be old
	registryKey := GenerateRegistryKey("ruleset", "https://registry.example.com")
	registryPath := filepath.Join(cacheRoot, "registries", registryKey)
	
	index, err := LoadRegistryIndex(registryPath)
	if err != nil {
		t.Fatalf("Failed to load registry index: %v", err)
	}

	// Set access time to 2 hours ago
	oldTime := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	for _, rulesetCache := range index.Rulesets {
		rulesetCache.LastAccessedOn = oldTime
		for _, versionCache := range rulesetCache.Versions {
			versionCache.LastAccessedOn = oldTime
		}
	}
	
	err = SaveRegistryIndex(registryPath, index)
	if err != nil {
		t.Fatalf("Failed to save registry index: %v", err)
	}

	// Run cleanup with 1 hour TTL
	err = manager.Cleanup(1*time.Hour, 0)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify data was removed
	_, err = manager.GetRuleset("https://registry.example.com", "test-rules", "1.0.0")
	if err == nil {
		t.Error("Expected cached data to be removed after TTL cleanup")
	}
}

func TestCleanupCache(t *testing.T) {
	cacheRoot := t.TempDir()

	// Initialize cache
	err := InitializeCache(cacheRoot)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Store some test data
	gitManager := NewGitCacheManager(cacheRoot)
	files := map[string][]byte{
		"rule.md": []byte("content"),
	}

	err = gitManager.StoreRuleset("https://github.com/test/repo", []string{"*.md"}, "abc123", files)
	if err != nil {
		t.Fatalf("Failed to store ruleset: %v", err)
	}

	// Run global cleanup
	err = CleanupCache(cacheRoot)
	if err != nil {
		t.Fatalf("Global cleanup failed: %v", err)
	}

	// Verify config was updated
	config, err := LoadCacheConfig(cacheRoot)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.LastUpdatedOn == "" {
		t.Error("Expected LastUpdatedOn to be set after cleanup")
	}
}

func TestGetCacheStats(t *testing.T) {
	cacheRoot := t.TempDir()

	// Initialize cache
	err := InitializeCache(cacheRoot)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Store some test data
	gitManager := NewGitCacheManager(cacheRoot)
	files := map[string][]byte{
		"rule.md": []byte("test content"),
	}

	err = gitManager.StoreRuleset("https://github.com/test/repo", []string{"*.md"}, "abc123", files)
	if err != nil {
		t.Fatalf("Failed to store ruleset: %v", err)
	}

	// Get cache stats
	stats, err := GetCacheStats(cacheRoot)
	if err != nil {
		t.Fatalf("Failed to get cache stats: %v", err)
	}

	// Verify stats
	if stats["total_size_bytes"] == nil {
		t.Error("Expected total_size_bytes in stats")
	}

	if stats["total_size_mb"] == nil {
		t.Error("Expected total_size_mb in stats")
	}

	if stats["registry_count"] == nil {
		t.Error("Expected registry_count in stats")
	}

	// Should have at least 1 registry
	if registryCount, ok := stats["registry_count"].(int); !ok || registryCount < 1 {
		t.Errorf("Expected at least 1 registry, got %v", stats["registry_count"])
	}
}

func TestCleanupCache_DisabledCleanup(t *testing.T) {
	cacheRoot := t.TempDir()

	// Initialize cache with cleanup disabled
	config := &CacheConfig{
		Version:        "1.0",
		CreatedOn:      time.Now().UTC().Format(time.RFC3339),
		LastUpdatedOn:  time.Now().UTC().Format(time.RFC3339),
		TTLHours:       24,
		MaxSizeMB:      1024,
		CleanupEnabled: false,
	}

	err := SaveCacheConfig(cacheRoot, config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Run cleanup (should do nothing)
	err = CleanupCache(cacheRoot)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify config wasn't updated (cleanup was skipped)
	newConfig, err := LoadCacheConfig(cacheRoot)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if newConfig.LastUpdatedOn != config.LastUpdatedOn {
		t.Error("Expected LastUpdatedOn to remain unchanged when cleanup is disabled")
	}
}

func TestGitCacheManager_CleanupPartialVersions(t *testing.T) {
	cacheRoot := t.TempDir()
	manager := NewGitCacheManager(cacheRoot)

	// Store multiple versions
	files := map[string][]byte{
		"rule.md": []byte("content"),
	}

	patterns := []string{"*.md"}
	err := manager.StoreRuleset("https://github.com/test/repo", patterns, "old123", files)
	if err != nil {
		t.Fatalf("Failed to store old ruleset: %v", err)
	}

	err = manager.StoreRuleset("https://github.com/test/repo", patterns, "new456", files)
	if err != nil {
		t.Fatalf("Failed to store new ruleset: %v", err)
	}

	// Manually set old version access time
	registryKey := GenerateRegistryKey("git", "https://github.com/test/repo")
	registryPath := filepath.Join(cacheRoot, "registries", registryKey)
	
	index, err := LoadRegistryIndex(registryPath)
	if err != nil {
		t.Fatalf("Failed to load registry index: %v", err)
	}

	patternsKey := GeneratePatternsKey(patterns)
	rulesetCache := index.Rulesets[patternsKey]
	
	// Set old version access time to 2 hours ago
	oldTime := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	rulesetCache.Versions["old123"].LastAccessedOn = oldTime
	
	err = SaveRegistryIndex(registryPath, index)
	if err != nil {
		t.Fatalf("Failed to save registry index: %v", err)
	}

	// Run cleanup with 1 hour TTL
	err = manager.Cleanup(1*time.Hour, 0)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify old version was removed but new version remains
	_, err = manager.GetRuleset("https://github.com/test/repo", patterns, "old123")
	if err == nil {
		t.Error("Expected old version to be removed")
	}

	_, err = manager.GetRuleset("https://github.com/test/repo", patterns, "new456")
	if err != nil {
		t.Errorf("Expected new version to remain: %v", err)
	}
}