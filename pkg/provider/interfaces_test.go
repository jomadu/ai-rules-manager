package provider

import (
	"testing"

	"github.com/jomadu/ai-rules-manager/pkg/config"
)

func TestGitRegistryProvider_CreateComponents(t *testing.T) {
	provider := NewGitRegistryProvider()

	regConfig := &config.RegistryConfig{
		URL:  "https://github.com/user/repo",
		Type: "git",
	}

	// Test CreateRegistry
	registry, err := provider.CreateRegistry(regConfig)
	if err != nil {
		t.Errorf("unexpected error creating registry: %v", err)
	}
	if registry == nil {
		t.Errorf("expected registry, got nil")
	}

	// Test CreateVersionResolver
	versionResolver, err := provider.CreateVersionResolver()
	if err != nil {
		t.Errorf("unexpected error creating version resolver: %v", err)
	}
	if versionResolver == nil {
		t.Errorf("expected version resolver, got nil")
	}

	// Test CreateContentResolver
	contentResolver, err := provider.CreateContentResolver()
	if err != nil {
		t.Errorf("unexpected error creating content resolver: %v", err)
	}
	if contentResolver == nil {
		t.Errorf("expected content resolver, got nil")
	}

	// Test CreateCacheKeyGenerator
	keyGenerator, err := provider.CreateCacheKeyGenerator()
	if err != nil {
		t.Errorf("unexpected error creating cache key generator: %v", err)
	}
	if keyGenerator == nil {
		t.Errorf("expected cache key generator, got nil")
	}
}

func TestGitRegistryProvider_ComponentTypes(t *testing.T) {
	provider := NewGitRegistryProvider()

	regConfig := &config.RegistryConfig{
		URL:  "https://github.com/user/repo",
		Type: "git",
	}

	// Verify that created components implement the expected interfaces
	registry, _ := provider.CreateRegistry(regConfig)
	metadata := registry.GetMetadata()
	if metadata.Type != "git" {
		t.Errorf("expected git registry type, got %s", metadata.Type)
	}
	if metadata.URL != "https://github.com/user/repo" {
		t.Errorf("expected URL https://github.com/user/repo, got %s", metadata.URL)
	}

	// Test that version resolver can be called
	versionResolver, _ := provider.CreateVersionResolver()
	_, err := versionResolver.ResolveVersion("^1.0.0", nil)
	if err != nil {
		t.Errorf("version resolver should handle nil input gracefully, got error: %v", err)
	}

	// Test that content resolver can be called
	contentResolver, _ := provider.CreateContentResolver()
	_, err = contentResolver.ResolveContent(nil, nil)
	if err != nil {
		t.Errorf("content resolver should handle nil input gracefully, got error: %v", err)
	}

	// Test that cache key generator can be called
	keyGenerator, _ := provider.CreateCacheKeyGenerator()
	key := keyGenerator.RegistryKey("https://github.com/user/repo")
	if len(key) != 64 {
		t.Errorf("expected 64 character hash, got %d characters", len(key))
	}
}
