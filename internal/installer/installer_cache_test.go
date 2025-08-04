package installer

import (
	"io"
	"strings"
	"testing"

	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/registry"
	"github.com/jomadu/arm/pkg/types"
)

// Test registry manager for testing
type testRegistryManager struct {
	downloadData string
	downloadErr  error
}

func (m *testRegistryManager) CachedDownload(registryName, rulesetName, version string) (io.ReadCloser, error) {
	if m.downloadErr != nil {
		return nil, m.downloadErr
	}
	return io.NopCloser(strings.NewReader(m.downloadData)), nil
}

func TestNewWithManager(t *testing.T) {
	configManager := config.NewManager()
	manager := registry.NewManager(configManager)

	installer := NewWithManager(manager, "test-registry", "test-ruleset")

	if installer.registryManager != manager {
		t.Error("Expected registry manager to be set")
	}
	if installer.registryName != "test-registry" {
		t.Errorf("Expected registry name 'test-registry', got %q", installer.registryName)
	}
	if installer.rulesetName != "test-ruleset" {
		t.Errorf("Expected ruleset name 'test-ruleset', got %q", installer.rulesetName)
	}
}

func TestInstallerDownloadWithCache(t *testing.T) {
	// Create installer with mock manager
	mockManager := &testRegistryManager{
		downloadData: "test tar.gz data",
	}

	installer := &Installer{
		registryManager: mockManager,
		registryName:    "test-registry",
		rulesetName:     "test-ruleset",
	}

	// Test cached download path
	data, err := installer.downloadRuleset("", "test-pkg", "1.0.0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(data) != "test tar.gz data" {
		t.Errorf("Expected 'test tar.gz data', got %q", string(data))
	}
}

func TestInstallerDownloadFallback(t *testing.T) {
	// Create mock registry
	mockReg := &mockRegistry{
		data: "fallback data",
	}

	// Create installer with direct registry (no manager)
	installer := &Installer{
		registry: mockReg,
	}

	// Test fallback to direct registry
	data, err := installer.downloadRuleset("", "test-pkg", "1.0.0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(data) != "fallback data" {
		t.Errorf("Expected 'fallback data', got %q", string(data))
	}
}

// Mock registry for fallback testing
type mockRegistry struct {
	data string
	err  error
}

func (m *mockRegistry) GetRuleset(name, version string) (*types.Ruleset, error) {
	return nil, m.err
}

func (m *mockRegistry) ListVersions(name string) ([]string, error) {
	return nil, m.err
}

func (m *mockRegistry) Download(name, version string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(strings.NewReader(m.data)), nil
}

func (m *mockRegistry) GetMetadata(name string) (*registry.Metadata, error) {
	return nil, m.err
}

func (m *mockRegistry) HealthCheck() error {
	return m.err
}
