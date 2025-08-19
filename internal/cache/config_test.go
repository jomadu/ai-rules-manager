package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadCacheConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cache-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("creates default config when none exists", func(t *testing.T) {
		config, err := LoadCacheConfig(tempDir)
		if err != nil {
			t.Fatalf("LoadCacheConfig() failed: %v", err)
		}

		if config.Version != "1.0" {
			t.Errorf("Expected version 1.0, got %s", config.Version)
		}
		if config.TTLHours != 24 {
			t.Errorf("Expected TTL 24 hours, got %d", config.TTLHours)
		}
		if config.MaxSizeMB != 1024 {
			t.Errorf("Expected max size 1024 MB, got %d", config.MaxSizeMB)
		}
		if !config.CleanupEnabled {
			t.Errorf("Expected cleanup enabled to be true")
		}

		// Check that config file was created
		configPath := filepath.Join(tempDir, "config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("Config file was not created")
		}
	})

	t.Run("loads existing config", func(t *testing.T) {
		// First call creates the config
		_, err := LoadCacheConfig(tempDir)
		if err != nil {
			t.Fatalf("Failed to create initial config: %v", err)
		}

		// Second call should load the existing config
		config, err := LoadCacheConfig(tempDir)
		if err != nil {
			t.Fatalf("LoadCacheConfig() failed: %v", err)
		}

		if config.Version != "1.0" {
			t.Errorf("Expected version 1.0, got %s", config.Version)
		}
	})
}

func TestSaveCacheConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cache-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &CacheConfig{
		Version:        "1.0",
		CreatedOn:      time.Now().UTC().Format(time.RFC3339),
		LastUpdatedOn:  time.Now().UTC().Format(time.RFC3339),
		TTLHours:       48,
		MaxSizeMB:      2048,
		CleanupEnabled: false,
	}

	err = SaveCacheConfig(tempDir, config)
	if err != nil {
		t.Fatalf("SaveCacheConfig() failed: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tempDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}

	// Load and verify content
	loadedConfig, err := LoadCacheConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.TTLHours != 48 {
		t.Errorf("Expected TTL 48 hours, got %d", loadedConfig.TTLHours)
	}
	if loadedConfig.MaxSizeMB != 2048 {
		t.Errorf("Expected max size 2048 MB, got %d", loadedConfig.MaxSizeMB)
	}
	if loadedConfig.CleanupEnabled {
		t.Errorf("Expected cleanup enabled to be false")
	}
}

func TestNewRegistryIndex(t *testing.T) {
	registryURL := "https://github.com/user/repo"
	registryType := "git"

	index := NewRegistryIndex(registryURL, registryType)

	if index.NormalizedRegistryURL != registryURL {
		t.Errorf("Expected registry URL %s, got %s", registryURL, index.NormalizedRegistryURL)
	}
	if index.NormalizedRegistryType != registryType {
		t.Errorf("Expected registry type %s, got %s", registryType, index.NormalizedRegistryType)
	}
	if index.Rulesets == nil {
		t.Errorf("Expected rulesets map to be initialized")
	}
	if len(index.Rulesets) != 0 {
		t.Errorf("Expected empty rulesets map, got %d entries", len(index.Rulesets))
	}

	// Check that timestamps are set
	if index.CreatedOn == "" {
		t.Errorf("Expected CreatedOn to be set")
	}
	if index.LastUpdatedOn == "" {
		t.Errorf("Expected LastUpdatedOn to be set")
	}
	if index.LastAccessedOn == "" {
		t.Errorf("Expected LastAccessedOn to be set")
	}
}

func TestNewRulesetCache(t *testing.T) {
	cache := NewRulesetCache()

	if cache.Versions == nil {
		t.Errorf("Expected versions map to be initialized")
	}
	if len(cache.Versions) != 0 {
		t.Errorf("Expected empty versions map, got %d entries", len(cache.Versions))
	}

	// Check that timestamps are set
	if cache.CreatedOn == "" {
		t.Errorf("Expected CreatedOn to be set")
	}
	if cache.LastUpdatedOn == "" {
		t.Errorf("Expected LastUpdatedOn to be set")
	}
	if cache.LastAccessedOn == "" {
		t.Errorf("Expected LastAccessedOn to be set")
	}
}

func TestNewVersionCache(t *testing.T) {
	cache := NewVersionCache()

	// Check that timestamps are set
	if cache.CreatedOn == "" {
		t.Errorf("Expected CreatedOn to be set")
	}
	if cache.LastUpdatedOn == "" {
		t.Errorf("Expected LastUpdatedOn to be set")
	}
	if cache.LastAccessedOn == "" {
		t.Errorf("Expected LastAccessedOn to be set")
	}
}

func TestUpdateAccessTime(t *testing.T) {
	t.Run("RegistryIndex", func(t *testing.T) {
		index := NewRegistryIndex("test-url", "test-type")
		originalTime := index.LastAccessedOn

		// Wait a full second to ensure time difference
		time.Sleep(1100 * time.Millisecond)

		index.UpdateAccessTime()

		if index.LastAccessedOn == originalTime {
			t.Errorf("Expected LastAccessedOn to be updated, original: %s, new: %s", originalTime, index.LastAccessedOn)
		}
	})

	t.Run("RulesetCache", func(t *testing.T) {
		cache := NewRulesetCache()
		originalTime := cache.LastAccessedOn

		// Wait a full second to ensure time difference
		time.Sleep(1100 * time.Millisecond)

		cache.UpdateAccessTime()

		if cache.LastAccessedOn == originalTime {
			t.Errorf("Expected LastAccessedOn to be updated, original: %s, new: %s", originalTime, cache.LastAccessedOn)
		}
	})

	t.Run("VersionCache", func(t *testing.T) {
		cache := NewVersionCache()
		originalTime := cache.LastAccessedOn

		// Wait a full second to ensure time difference
		time.Sleep(1100 * time.Millisecond)

		cache.UpdateAccessTime()

		if cache.LastAccessedOn == originalTime {
			t.Errorf("Expected LastAccessedOn to be updated, original: %s, new: %s", originalTime, cache.LastAccessedOn)
		}
	})
}

func TestLoadRegistryIndex(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "registry-index-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("returns error when index doesn't exist", func(t *testing.T) {
		_, err := LoadRegistryIndex(tempDir)
		if err == nil {
			t.Errorf("Expected error when index doesn't exist")
		}
	})

	t.Run("loads existing index", func(t *testing.T) {
		// Create and save an index
		index := NewRegistryIndex("test-url", "test-type")
		err := SaveRegistryIndex(tempDir, index)
		if err != nil {
			t.Fatalf("Failed to save index: %v", err)
		}

		// Load the index
		loadedIndex, err := LoadRegistryIndex(tempDir)
		if err != nil {
			t.Fatalf("LoadRegistryIndex() failed: %v", err)
		}

		if loadedIndex.NormalizedRegistryURL != "test-url" {
			t.Errorf("Expected registry URL test-url, got %s", loadedIndex.NormalizedRegistryURL)
		}
		if loadedIndex.NormalizedRegistryType != "test-type" {
			t.Errorf("Expected registry type test-type, got %s", loadedIndex.NormalizedRegistryType)
		}
	})
}

func TestSaveRegistryIndex(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "registry-index-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	index := NewRegistryIndex("test-url", "test-type")

	err = SaveRegistryIndex(tempDir, index)
	if err != nil {
		t.Fatalf("SaveRegistryIndex() failed: %v", err)
	}

	// Verify file was created
	indexPath := filepath.Join(tempDir, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("Index file was not created")
	}
}
