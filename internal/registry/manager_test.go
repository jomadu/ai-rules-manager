package registry

import (
	"testing"

	"github.com/jomadu/arm/internal/config"
)

func TestManager_parseRegistryName(t *testing.T) {
	// Create mock config manager
	configManager := config.NewManager()
	_ = configManager.Load() // Ignore load errors for testing
	configManager.SetSource("company", &config.Source{URL: "https://company.local"})
	configManager.SetSource("default", &config.Source{URL: "https://registry.armjs.org"})

	manager := NewManager(configManager)

	tests := []struct {
		name     string
		ruleset  string
		expected string
	}{
		{"explicit registry", "company@typescript-rules", "company"},
		{"default registry", "typescript-rules", "default"},
		{"unknown registry", "unknown@typescript-rules", "default"},
		{"org format", "myorg@package@1.0.0", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.parseRegistryName(tt.ruleset)
			if result != tt.expected {
				t.Errorf("parseRegistryName(%s) = %s, want %s", tt.ruleset, result, tt.expected)
			}
		})
	}
}

func TestManager_StripRegistryPrefix(t *testing.T) {
	// Create mock config manager
	configManager := config.NewManager()
	_ = configManager.Load() // Ignore load errors for testing
	configManager.SetSource("company", &config.Source{URL: "https://company.local"})

	manager := NewManager(configManager)

	tests := []struct {
		name     string
		ruleset  string
		expected string
	}{
		{"with registry prefix", "company@typescript-rules", "typescript-rules"},
		{"without registry prefix", "typescript-rules", "typescript-rules"},
		{"unknown registry prefix", "unknown@typescript-rules", "unknown@typescript-rules"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.StripRegistryPrefix(tt.ruleset)
			if result != tt.expected {
				t.Errorf("StripRegistryPrefix(%s) = %s, want %s", tt.ruleset, result, tt.expected)
			}
		})
	}
}
