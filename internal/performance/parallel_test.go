package performance

import (
	"errors"
	"testing"

	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/registry"
)

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

func TestParallelDownloader_DownloadAll(t *testing.T) {
	// Create test config
	testConfig := &config.ARMConfig{
		Sources: map[string]config.Source{
			"default": {Type: "http"},
		},
		Performance: config.PerformanceConfig{
			DefaultConcurrency: 3,
			RegistryTypes:      make(map[string]config.TypeConfig),
		},
	}

	configManager := &mockConfigManager{config: testConfig}
	registryManager := registry.NewManager(configManager)
	downloader := NewParallelDownloader(registryManager)

	tests := []struct {
		name     string
		jobs     []DownloadJob
		expected int
	}{
		{
			name:     "empty jobs",
			jobs:     []DownloadJob{},
			expected: 0,
		},
		{
			name: "single job",
			jobs: []DownloadJob{
				{
					Name:            "test-ruleset",
					VersionSpec:     "1.0.0",
					RegistryName:    "default",
					CleanName:       "test-ruleset",
					RegistryManager: registryManager,
				},
			},
			expected: 1,
		},
		{
			name: "multiple jobs",
			jobs: []DownloadJob{
				{
					Name:            "test-ruleset-1",
					VersionSpec:     "1.0.0",
					RegistryName:    "default",
					CleanName:       "test-ruleset-1",
					RegistryManager: registryManager,
				},
				{
					Name:            "test-ruleset-2",
					VersionSpec:     "2.0.0",
					RegistryName:    "default",
					CleanName:       "test-ruleset-2",
					RegistryManager: registryManager,
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := downloader.DownloadAll(tt.jobs)
			if len(results) != tt.expected {
				t.Errorf("DownloadAll() returned %d results, expected %d", len(results), tt.expected)
			}
		})
	}
}

func TestPrintResults(t *testing.T) {
	tests := []struct {
		name        string
		results     []DownloadResult
		expectError bool
	}{
		{
			name:        "no results",
			results:     []DownloadResult{},
			expectError: false,
		},
		{
			name: "all successful",
			results: []DownloadResult{
				{
					Job:   DownloadJob{Name: "test1", VersionSpec: "1.0.0"},
					Error: nil,
				},
				{
					Job:   DownloadJob{Name: "test2", VersionSpec: "2.0.0"},
					Error: nil,
				},
			},
			expectError: false,
		},
		{
			name: "some failures",
			results: []DownloadResult{
				{
					Job:   DownloadJob{Name: "test1", VersionSpec: "1.0.0"},
					Error: nil,
				},
				{
					Job:   DownloadJob{Name: "test2", VersionSpec: "2.0.0"},
					Error: errors.New("download failed"),
				},
			},
			expectError: true,
		},
		{
			name: "all failures",
			results: []DownloadResult{
				{
					Job:   DownloadJob{Name: "test1", VersionSpec: "1.0.0"},
					Error: errors.New("download failed"),
				},
				{
					Job:   DownloadJob{Name: "test2", VersionSpec: "2.0.0"},
					Error: errors.New("download failed"),
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PrintResults(tt.results)
			if (err != nil) != tt.expectError {
				t.Errorf("PrintResults() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
