package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/max-dunn/ai-rules-manager/internal/cache"
)

func TestDefaultCacheConfig(t *testing.T) {
	cfg := DefaultCacheConfig()

	// Should have reasonable defaults
	if cfg.MaxSize != 0 {
		t.Errorf("Expected unlimited max size (0), got %d", cfg.MaxSize)
	}

	if cfg.TTL != 24*time.Hour {
		t.Errorf("Expected 24h TTL, got %v", cfg.TTL)
	}

	if cfg.CleanupInterval != 6*time.Hour {
		t.Errorf("Expected 6h cleanup interval, got %v", cfg.CleanupInterval)
	}

	// Path should be in home directory
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "."
	}
	expectedPath := filepath.Join(homeDir, ".arm", "cache")
	if cfg.Path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, cfg.Path)
	}

}

func TestLoadCacheConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected *CacheConfig
	}{
		{
			name: "default config",
			config: &Config{
				TypeDefaults: make(map[string]map[string]string),
			},
			expected: DefaultCacheConfig(),
		},
		{
			name: "custom cache settings",
			config: &Config{
				TypeDefaults: map[string]map[string]string{
					"cache": {
						"path":            "/custom/cache/path",
						"maxSize":         "1073741824", // 1GB
						"ttl":             "12h",
						"cleanupInterval": "3h",
					},
				},
			},
			expected: &CacheConfig{
				Path:            "/custom/cache/path",
				MaxSize:         1073741824,
				TTL:             12 * time.Hour,
				CleanupInterval: 3 * time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.LoadCacheConfig()

			if result.Path != tt.expected.Path {
				t.Errorf("Expected path %s, got %s", tt.expected.Path, result.Path)
			}

			if result.MaxSize != tt.expected.MaxSize {
				t.Errorf("Expected max size %d, got %d", tt.expected.MaxSize, result.MaxSize)
			}

			if result.TTL != tt.expected.TTL {
				t.Errorf("Expected TTL %v, got %v", tt.expected.TTL, result.TTL)
			}

			if result.CleanupInterval != tt.expected.CleanupInterval {
				t.Errorf("Expected cleanup interval %v, got %v", tt.expected.CleanupInterval, result.CleanupInterval)
			}

		})
	}
}

func TestCachePathConfigurationRespected(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "arm-cache-path-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	customCachePath := filepath.Join(tempDir, "custom-cache")

	// Create a config with custom cache path and other settings
	cfg := &Config{
		TypeDefaults: map[string]map[string]string{
			"cache": {
				"path":            customCachePath,
				"ttl":             "12h",
				"maxSize":         "2147483648", // 2GB
				"cleanupInterval": "3h",
			},
		},
	}

	// Load cache configuration
	cacheConfig := cfg.LoadCacheConfig()

	// Verify all custom settings are respected
	if cacheConfig.Path != customCachePath {
		t.Errorf("Expected cache path %s, got %s", customCachePath, cacheConfig.Path)
	}

	if cacheConfig.TTL != 12*time.Hour {
		t.Errorf("Expected TTL 12h, got %v", cacheConfig.TTL)
	}

	if cacheConfig.MaxSize != 2147483648 {
		t.Errorf("Expected max size 2147483648, got %d", cacheConfig.MaxSize)
	}

	if cacheConfig.CleanupInterval != 3*time.Hour {
		t.Errorf("Expected cleanup interval 3h, got %v", cacheConfig.CleanupInterval)
	}

	// Test that cache manager uses the configured path
	manager := cache.NewGitRegistryCacheManager(cacheConfig.Path)

	registryURL := "https://github.com/test/repo"

	// Get repository path and verify it uses the custom path
	repoPath, err := manager.GetRepositoryPath(registryURL)
	if err != nil {
		t.Fatalf("Failed to get repository path: %v", err)
	}

	if !strings.HasPrefix(repoPath, customCachePath) {
		t.Errorf("Repository path %s does not start with configured path %s", repoPath, customCachePath)
	}
}
