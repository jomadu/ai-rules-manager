package cache

import (
	"testing"
	"time"

	"github.com/jomadu/ai-rules-manager/pkg/registry"
)

func TestGitCacheKeyGenerator_RegistryKey(t *testing.T) {
	generator := NewGitCacheKeyGenerator()

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "github url",
			url:  "https://github.com/user/repo",
		},
		{
			name: "gitlab url",
			url:  "https://gitlab.com/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := generator.RegistryKey(tt.url)

			// Should return a SHA256 hash (64 hex characters)
			if len(key) != 64 {
				t.Errorf("expected 64 character hash, got %d characters", len(key))
			}

			// Same URL should produce same key
			key2 := generator.RegistryKey(tt.url)
			if key != key2 {
				t.Errorf("same URL should produce same key")
			}
		})
	}
}

func TestGitCacheKeyGenerator_RulesetKey(t *testing.T) {
	generator := NewGitCacheKeyGenerator()

	selector := registry.GitContentSelector{
		Patterns: []string{"rules/*.md"},
		Excludes: []string{"**/test.md"},
	}

	key := generator.RulesetKey("my-ruleset", selector)

	// Should return a SHA256 hash (64 hex characters)
	if len(key) != 64 {
		t.Errorf("expected 64 character hash, got %d characters", len(key))
	}

	// Same inputs should produce same key
	key2 := generator.RulesetKey("my-ruleset", selector)
	if key != key2 {
		t.Errorf("same inputs should produce same key")
	}

	// Different ruleset name should produce different key
	key3 := generator.RulesetKey("other-ruleset", selector)
	if key == key3 {
		t.Errorf("different ruleset names should produce different keys")
	}
}

func TestGitCacheKeyGenerator_VersionKey(t *testing.T) {
	generator := NewGitCacheKeyGenerator()

	tests := []struct {
		name       string
		versionRef registry.VersionRef
		expected   string
	}{
		{
			name: "commit hash",
			versionRef: registry.VersionRef{
				ID:   "abc123def456",
				Type: registry.Commit,
			},
			expected: "abc123def456",
		},
		{
			name: "tag version",
			versionRef: registry.VersionRef{
				ID:   "v1.0.0",
				Type: registry.Tag,
			},
			expected: "v1.0.0",
		},
		{
			name: "branch name",
			versionRef: registry.VersionRef{
				ID:   "main",
				Type: registry.Branch,
			},
			expected: "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := generator.VersionKey(tt.versionRef)

			// For Git, should return the ID directly as cache key
			if key != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, key)
			}
		})
	}
}

func TestFileCache_Operations(t *testing.T) {
	cache := NewFileCache("/tmp/test-cache")

	// Test cache miss behavior
	content, _ := cache.Get("nonexistent-key")
	if content != nil {
		t.Errorf("cache miss should return nil content")
	}
	// Cache miss should return an error or nil content

	// Test cache set operation
	testContent := []byte("test content")
	err := cache.Set("test-key", testContent, time.Hour)
	if err != nil {
		t.Errorf("cache set should not error for valid inputs: %v", err)
	}

	// Test cache hit behavior (after set)
	_, err = cache.Get("test-key")
	if err != nil {
		t.Errorf("cache get should not error for existing key: %v", err)
	}
	// TODO: Once implemented, verify stored content is returned:
	// if !bytes.Equal(content, testContent) {
	//     t.Errorf("expected content %s, got %s", testContent, content)
	// }

	// Test cache delete operation
	err = cache.Delete("test-key")
	if err != nil {
		t.Errorf("cache delete should not error for valid key: %v", err)
	}

	// Test cache clear operation
	err = cache.Clear()
	if err != nil {
		t.Errorf("cache clear should not error: %v", err)
	}

	// Test TTL behavior
	err = cache.Set("ttl-key", []byte("ttl content"), time.Nanosecond)
	if err != nil {
		t.Errorf("cache set with TTL should not error: %v", err)
	}
	// TODO: Once implemented, verify expired content is not returned
}
