package registry

import (
	"testing"

	"github.com/jomadu/arm/internal/config"
)

func TestGetRegistryURL(t *testing.T) {
	configManager := config.NewManager()
	_ = configManager.Load() // Initialize config
	configManager.SetSource("test", &config.Source{URL: "https://test.com"})
	configManager.SetSource("default", &config.Source{URL: "https://registry.armjs.org"})

	manager := NewManager(configManager)

	tests := []struct {
		name         string
		registryName string
		expected     string
	}{
		{"known registry", "test", "https://test.com"},
		{"default registry", "default", "https://registry.armjs.org"},
		{"unknown registry", "unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.getRegistryURL(tt.registryName)
			if result != tt.expected {
				t.Errorf("getRegistryURL(%s) = %s, want %s", tt.registryName, result, tt.expected)
			}
		})
	}
}

func TestConvertMetadata(t *testing.T) {
	// Test convertFromMetadata
	metadata := &Metadata{
		Name:        "test-ruleset",
		Description: "Test description",
		Repository:  "https://github.com/test/repo",
		Homepage:    "https://test.com",
		License:     "MIT",
		Keywords:    []string{"test", "rules"},
	}

	data := convertFromMetadata(metadata)

	if data["name"] != "test-ruleset" {
		t.Errorf("Expected name 'test-ruleset', got %v", data["name"])
	}
	if data["description"] != "Test description" {
		t.Errorf("Expected description 'Test description', got %v", data["description"])
	}

	// Test convertToMetadata
	converted := convertToMetadata(data)

	if converted.Name != "test-ruleset" {
		t.Errorf("Expected name 'test-ruleset', got %s", converted.Name)
	}
	if converted.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got %s", converted.Description)
	}
}

func TestConvertMetadataWithMissingFields(t *testing.T) {
	// Test with minimal data
	data := map[string]interface{}{
		"name": "minimal-ruleset",
	}

	converted := convertToMetadata(data)

	if converted.Name != "minimal-ruleset" {
		t.Errorf("Expected name 'minimal-ruleset', got %s", converted.Name)
	}
	if converted.Description != "" {
		t.Errorf("Expected empty description, got %s", converted.Description)
	}
}

func TestConvertMetadataWithWrongTypes(t *testing.T) {
	// Test with wrong data types (should not panic)
	data := map[string]interface{}{
		"name":        123,                            // wrong type
		"description": []string{"not", "a", "string"}, // wrong type
	}

	converted := convertToMetadata(data)

	// Should handle wrong types gracefully
	if converted.Name != "" {
		t.Errorf("Expected empty name for wrong type, got %s", converted.Name)
	}
	if converted.Description != "" {
		t.Errorf("Expected empty description for wrong type, got %s", converted.Description)
	}
}
