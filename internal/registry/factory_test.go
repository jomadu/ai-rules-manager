package registry

import (
	"testing"
	"time"

	"github.com/max-dunn/ai-rules-manager/internal/config"
)

func TestCreateRegistryWithCacheConfig(t *testing.T) {

	// Test registry config
	registryConfig := &RegistryConfig{
		Name: "test-registry",
		Type: "git",
		URL:  "https://github.com/test/repo",
	}

	tests := []struct {
		name         string
		cacheConfig  *config.CacheConfig
		registryName string
		expectCache  bool
		description  string
	}{
		{
			name: "cache enabled globally",
			cacheConfig: &config.CacheConfig{
				TTL:     config.Duration(24 * time.Hour),
				MaxSize: 0,
			},
			registryName: "test-registry",
			expectCache:  true,
			description:  "Should use cache when enabled globally",
		},
		{
			name:         "nil cache config",
			cacheConfig:  nil,
			registryName: "test-registry",
			expectCache:  false,
			description:  "Should not create cache manager when no cache config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, err := CreateRegistryWithCacheConfig(
				registryConfig,
				tt.cacheConfig,
				tt.registryName,
			)

			if err != nil {
				t.Fatalf("Failed to create registry: %v", err)
			}

			// Check if the registry is a Git registry
			gitRegistry, ok := registry.(*GitRegistry)
			if !ok {
				t.Fatalf("Expected GitRegistry, got %T", registry)
			}

			// Check if cache manager is set based on expectation
			hasCacheManager := gitRegistry.cacheManager != nil
			if hasCacheManager != tt.expectCache {
				t.Errorf("%s: expected cache manager = %v, got %v",
					tt.description, tt.expectCache, hasCacheManager)
			}
		})
	}
}

func TestCreateRegistryWithCacheConfigInvalidRegistry(t *testing.T) {
	cacheConfig := &config.CacheConfig{
		TTL: config.Duration(24 * time.Hour),
	}

	// Invalid registry config (missing required fields)
	registryConfig := &RegistryConfig{
		Type: "invalid-type",
		URL:  "",
	}

	_, err := CreateRegistryWithCacheConfig(
		registryConfig,
		cacheConfig,
		"test-registry",
	)

	if err == nil {
		t.Error("Expected error for invalid registry config, got nil")
	}
}
