package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRulesetMapper_AddAndGetMapping(t *testing.T) {
	tempDir := t.TempDir()
	mapper := NewRulesetMapper(tempDir)

	cacheKey := "abc123"
	registryCacheKey := "def456"
	rulesetName := "test-ruleset"
	patterns := []string{"*.md", "*.txt"}

	// Add mapping
	err := mapper.AddMapping(cacheKey, registryCacheKey, rulesetName, patterns)
	if err != nil {
		t.Fatalf("Failed to add mapping: %v", err)
	}

	// Get mapping
	mapping, err := mapper.GetMapping(cacheKey)
	if err != nil {
		t.Fatalf("Failed to get mapping: %v", err)
	}

	if mapping.CacheKey != cacheKey {
		t.Errorf("Expected cache key %s, got %s", cacheKey, mapping.CacheKey)
	}
	if mapping.RegistryCacheKey != registryCacheKey {
		t.Errorf("Expected registry cache key %s, got %s", registryCacheKey, mapping.RegistryCacheKey)
	}
	if mapping.RulesetName != rulesetName {
		t.Errorf("Expected ruleset name %s, got %s", rulesetName, mapping.RulesetName)
	}
	if len(mapping.Patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(mapping.Patterns))
	}
}

func TestRulesetMapper_FindMappingByRuleset(t *testing.T) {
	tempDir := t.TempDir()
	mapper := NewRulesetMapper(tempDir)

	cacheKey := "abc123"
	registryCacheKey := "def456"
	rulesetName := "test-ruleset"
	patterns := []string{"*.md", "*.txt"}

	// Add mapping
	err := mapper.AddMapping(cacheKey, registryCacheKey, rulesetName, patterns)
	if err != nil {
		t.Fatalf("Failed to add mapping: %v", err)
	}

	// Find mapping by ruleset
	mapping, err := mapper.FindMappingByRuleset(registryCacheKey, rulesetName, patterns)
	if err != nil {
		t.Fatalf("Failed to find mapping: %v", err)
	}

	if mapping.CacheKey != cacheKey {
		t.Errorf("Expected cache key %s, got %s", cacheKey, mapping.CacheKey)
	}
}

func TestRulesetMapper_PatternNormalization(t *testing.T) {
	tempDir := t.TempDir()
	mapper := NewRulesetMapper(tempDir)

	// Test that different order of patterns produces same normalized result
	patterns1 := []string{"*.md", "*.txt"}
	patterns2 := []string{"*.txt", "*.md"}

	normalized1 := mapper.normalizePatterns(patterns1)
	normalized2 := mapper.normalizePatterns(patterns2)

	if normalized1 != normalized2 {
		t.Errorf("Pattern normalization failed: %s != %s", normalized1, normalized2)
	}
}

func TestRulesetMapper_UpdateLastAccessed(t *testing.T) {
	tempDir := t.TempDir()
	mapper := NewRulesetMapper(tempDir)

	cacheKey := "abc123"
	registryCacheKey := "def456"
	rulesetName := "test-ruleset"
	patterns := []string{"*.md"}

	// Add mapping
	err := mapper.AddMapping(cacheKey, registryCacheKey, rulesetName, patterns)
	if err != nil {
		t.Fatalf("Failed to add mapping: %v", err)
	}

	// Get initial mapping
	mapping1, err := mapper.GetMapping(cacheKey)
	if err != nil {
		t.Fatalf("Failed to get mapping: %v", err)
	}

	// Wait a bit and update last accessed
	time.Sleep(10 * time.Millisecond)
	err = mapper.UpdateLastAccessed(cacheKey)
	if err != nil {
		t.Fatalf("Failed to update last accessed: %v", err)
	}

	// Get updated mapping
	mapping2, err := mapper.GetMapping(cacheKey)
	if err != nil {
		t.Fatalf("Failed to get updated mapping: %v", err)
	}

	if !mapping2.LastAccessed.After(mapping1.LastAccessed) {
		t.Error("Last accessed time was not updated")
	}
}

func TestRulesetMapper_RemoveMapping(t *testing.T) {
	tempDir := t.TempDir()
	mapper := NewRulesetMapper(tempDir)

	cacheKey := "abc123"
	registryCacheKey := "def456"
	rulesetName := "test-ruleset"
	patterns := []string{"*.md"}

	// Add mapping
	err := mapper.AddMapping(cacheKey, registryCacheKey, rulesetName, patterns)
	if err != nil {
		t.Fatalf("Failed to add mapping: %v", err)
	}

	// Remove mapping
	err = mapper.RemoveMapping(cacheKey)
	if err != nil {
		t.Fatalf("Failed to remove mapping: %v", err)
	}

	// Try to get removed mapping
	_, err = mapper.GetMapping(cacheKey)
	if err == nil {
		t.Error("Expected error when getting removed mapping")
	}
}

func TestRulesetMapper_ListMappingsByRegistry(t *testing.T) {
	tempDir := t.TempDir()
	mapper := NewRulesetMapper(tempDir)

	registryCacheKey := "def456"

	// Add multiple mappings for same registry
	err := mapper.AddMapping("key1", registryCacheKey, "ruleset1", []string{"*.md"})
	if err != nil {
		t.Fatalf("Failed to add mapping 1: %v", err)
	}

	err = mapper.AddMapping("key2", registryCacheKey, "ruleset2", []string{"*.txt"})
	if err != nil {
		t.Fatalf("Failed to add mapping 2: %v", err)
	}

	err = mapper.AddMapping("key3", "other-registry", "ruleset3", []string{"*.py"})
	if err != nil {
		t.Fatalf("Failed to add mapping 3: %v", err)
	}

	// List mappings for specific registry
	mappings, err := mapper.ListMappingsByRegistry(registryCacheKey)
	if err != nil {
		t.Fatalf("Failed to list mappings: %v", err)
	}

	if len(mappings) != 2 {
		t.Errorf("Expected 2 mappings for registry, got %d", len(mappings))
	}
}

func TestRulesetMapper_FileOperations(t *testing.T) {
	tempDir := t.TempDir()
	mapper := NewRulesetMapper(tempDir)

	// Add mapping to create file
	err := mapper.AddMapping("key1", "reg1", "ruleset1", []string{"*.md"})
	if err != nil {
		t.Fatalf("Failed to add mapping: %v", err)
	}

	// Check file exists
	mapFilePath := filepath.Join(tempDir, "ruleset-map.json")
	if _, err := os.Stat(mapFilePath); os.IsNotExist(err) {
		t.Error("Ruleset map file was not created")
	}

	// Create new mapper instance to test file loading
	mapper2 := NewRulesetMapper(tempDir)
	mapping, err := mapper2.GetMapping("key1")
	if err != nil {
		t.Fatalf("Failed to load mapping from file: %v", err)
	}

	if mapping.RulesetName != "ruleset1" {
		t.Errorf("Expected ruleset name 'ruleset1', got '%s'", mapping.RulesetName)
	}
}
