package registry

import (
	"testing"
)

func TestCreateRegistry(t *testing.T) {
	registryConfig := &RegistryConfig{
		Name: "test-registry",
		Type: "git",
		URL:  "https://github.com/test/repo",
	}

	registry, err := CreateRegistry(registryConfig)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	gitRegistry, ok := registry.(*GitRegistry)
	if !ok {
		t.Fatalf("Expected GitRegistry, got %T", registry)
	}

	if gitRegistry.cacheManager == nil {
		t.Error("Expected cache manager to be set")
	}
}

func TestCreateRegistryInvalid(t *testing.T) {
	registryConfig := &RegistryConfig{
		Type: "invalid-type",
		URL:  "",
	}

	_, err := CreateRegistry(registryConfig)
	if err == nil {
		t.Error("Expected error for invalid registry config, got nil")
	}
}
