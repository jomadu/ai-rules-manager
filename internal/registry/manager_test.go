package registry

import (
	"testing"

	"github.com/jomadu/arm/internal/config"
)

func TestManager_GetConcurrency(t *testing.T) {
	// Create test config
	testConfig := &config.ARMConfig{
		Sources: map[string]config.Source{
			"company": {
				Type:        "gitlab",
				Concurrency: 2, // Source-specific override
			},
			"company-2": {
				Type: "gitlab", // No source-specific concurrency
			},
			"s3-backup": {
				Type: "s3", // No source-specific concurrency
			},
		},
		Performance: config.PerformanceConfig{
			DefaultConcurrency: 5,
			RegistryTypes: map[string]config.TypeConfig{
				"gitlab": {Concurrency: 3},
				"s3":     {Concurrency: 8},
			},
		},
	}

	// Create mock config manager
	configManager := &mockConfigManager{config: testConfig}
	manager := &Manager{configManager: configManager}

	tests := []struct {
		name       string
		sourceName string
		expected   int
	}{
		{
			name:       "source-specific override",
			sourceName: "company",
			expected:   2, // Source override
		},
		{
			name:       "registry type default",
			sourceName: "company-2",
			expected:   3, // GitLab type default
		},
		{
			name:       "s3 registry type default",
			sourceName: "s3-backup",
			expected:   8, // S3 type default
		},
		{
			name:       "nonexistent source",
			sourceName: "nonexistent",
			expected:   5, // Global default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.GetConcurrency(tt.sourceName)
			if result != tt.expected {
				t.Errorf("GetConcurrency(%s) = %d, expected %d", tt.sourceName, result, tt.expected)
			}
		})
	}
}

func TestManager_GetConcurrency_HardcodedFallbacks(t *testing.T) {
	// Create test config with zero default concurrency to test hardcoded fallbacks
	testConfig := &config.ARMConfig{
		Sources: map[string]config.Source{
			"gitlab-source":  {Type: "gitlab"},
			"s3-source":      {Type: "s3"},
			"http-source":    {Type: "http"},
			"fs-source":      {Type: "filesystem"},
			"unknown-source": {Type: "unknown"},
		},
		Performance: config.PerformanceConfig{
			DefaultConcurrency: 0, // Force hardcoded fallbacks
			RegistryTypes:      map[string]config.TypeConfig{},
		},
	}

	configManager := &mockConfigManager{config: testConfig}
	manager := &Manager{configManager: configManager}

	tests := []struct {
		name       string
		sourceName string
		expected   int
	}{
		{"gitlab fallback", "gitlab-source", 2},
		{"s3 fallback", "s3-source", 8},
		{"http fallback", "http-source", 4},
		{"filesystem fallback", "fs-source", 10},
		{"unknown fallback", "unknown-source", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.GetConcurrency(tt.sourceName)
			if result != tt.expected {
				t.Errorf("GetConcurrency(%s) = %d, expected %d", tt.sourceName, result, tt.expected)
			}
		})
	}
}

func TestManager_ParseRegistryName(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.ARMConfig
		rulesetName string
		expected    string
	}{
		{
			name: "explicit registry prefix",
			config: &config.ARMConfig{
				Sources: map[string]config.Source{
					"company": {Type: "gitlab"},
				},
			},
			rulesetName: "company@typescript-rules",
			expected:    "company",
		},
		{
			name: "default source exists",
			config: &config.ARMConfig{
				Sources: map[string]config.Source{
					"default": {Type: "http"},
					"company": {Type: "gitlab"},
				},
			},
			rulesetName: "typescript-rules",
			expected:    "default",
		},
		{
			name: "single source auto-detection",
			config: &config.ARMConfig{
				Sources: map[string]config.Source{
					"company": {Type: "gitlab"},
				},
			},
			rulesetName: "typescript-rules",
			expected:    "company",
		},
		{
			name: "multiple sources without default",
			config: &config.ARMConfig{
				Sources: map[string]config.Source{
					"company": {Type: "gitlab"},
					"backup":  {Type: "s3"},
				},
			},
			rulesetName: "typescript-rules",
			expected:    "default", // Will trigger error with helpful message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configManager := &mockConfigManager{config: tt.config}
			manager := &Manager{configManager: configManager}
			result := manager.ParseRegistryName(tt.rulesetName)
			if result != tt.expected {
				t.Errorf("ParseRegistryName(%s) = %s, expected %s", tt.rulesetName, result, tt.expected)
			}
		})
	}
}

// mockConfigManager implements the config manager interface for testing
type mockConfigManager struct {
	config *config.ARMConfig
}

func (m *mockConfigManager) GetConfig() *config.ARMConfig {
	return m.config
}

func (m *mockConfigManager) GetSource(name string) (config.Source, bool) {
	source, exists := m.config.Sources[name]
	return source, exists
}

func (m *mockConfigManager) SetSource(name string, source *config.Source) {
	m.config.Sources[name] = *source
}

func (m *mockConfigManager) Load() error {
	return nil
}
