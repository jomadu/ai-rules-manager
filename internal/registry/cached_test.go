package registry

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/pkg/types"
)

// Mock registry for testing
type mockRegistry struct {
	versions []string
	metadata *Metadata
	data     string
	err      error
}

func (m *mockRegistry) GetRuleset(name, version string) (*types.Ruleset, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockRegistry) ListVersions(name string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.versions, nil
}

func (m *mockRegistry) Download(name, version string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(strings.NewReader(m.data)), nil
}

func (m *mockRegistry) GetMetadata(name string) (*Metadata, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.metadata, nil
}

func (m *mockRegistry) HealthCheck() error {
	return m.err
}

func TestCachedListVersions(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	_ = configManager.Load() // Initialize config
	configManager.SetSource("test", &config.Source{URL: "https://test.com"})

	manager := NewManager(configManager)
	mockReg := &mockRegistry{
		versions: []string{"1.0.0", "1.1.0", "2.0.0"},
	}
	manager.registries["test"] = mockReg

	// Test cache miss
	versions, err := manager.CachedListVersions("test", "typescript-rules")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(versions))
	}

	// Test cache hit (modify mock to return different data)
	mockReg.versions = []string{"3.0.0"}
	versions, err = manager.CachedListVersions("test", "typescript-rules")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Should still return cached versions
	if len(versions) != 3 {
		t.Errorf("Expected cached 3 versions, got %d", len(versions))
	}
}

func TestCachedDownload(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	_ = configManager.Load() // Initialize config
	configManager.SetSource("test", &config.Source{URL: "https://test.com"})

	manager := NewManager(configManager)
	mockReg := &mockRegistry{
		data: "test package data",
	}
	manager.registries["test"] = mockReg

	// Test cache miss
	reader, err := manager.CachedDownload("test", "typescript-rules", "1.0.0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer func() { _ = reader.Close() }()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}
	if string(data) != "test package data" {
		t.Errorf("Expected 'test package data', got %q", string(data))
	}
}

func TestCachedGetMetadata(t *testing.T) {
	// Setup
	configManager := config.NewManager()
	_ = configManager.Load() // Initialize config
	configManager.SetSource("test", &config.Source{URL: "https://test.com"})

	manager := NewManager(configManager)
	mockReg := &mockRegistry{
		metadata: &Metadata{
			Name:        "typescript-rules",
			Description: "TypeScript coding rules",
		},
	}
	manager.registries["test"] = mockReg

	// Test cache miss
	metadata, err := manager.CachedGetMetadata("test", "typescript-rules")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if metadata.Name != "typescript-rules" {
		t.Errorf("Expected name 'typescript-rules', got %q", metadata.Name)
	}
}

func TestCachedOperationsWithErrors(t *testing.T) {
	// Setup with different registry name to avoid cache conflicts
	configManager := config.NewManager()
	_ = configManager.Load() // Initialize config
	configManager.SetSource("error-test", &config.Source{URL: "https://error-test.com"})

	manager := NewManager(configManager)
	mockReg := &mockRegistry{
		err: fmt.Errorf("registry error"),
	}
	manager.registries["error-test"] = mockReg

	// Test ListVersions error
	_, err := manager.CachedListVersions("error-test", "error-ruleset")
	if err == nil {
		t.Error("Expected error from ListVersions")
	}

	// Test Download error
	_, err = manager.CachedDownload("error-test", "error-ruleset", "1.0.0")
	if err == nil {
		t.Error("Expected error from Download")
	}

	// Test GetMetadata error
	_, err = manager.CachedGetMetadata("error-test", "error-ruleset")
	if err == nil {
		t.Error("Expected error from GetMetadata")
	}
}

func TestCachedOperationsWithoutCache(t *testing.T) {
	// Setup manager without cache
	configManager := config.NewManager()
	_ = configManager.Load() // Initialize config
	configManager.SetSource("test", &config.Source{URL: "https://test.com"})

	manager := &Manager{
		configManager: configManager,
		registries:    make(map[string]Registry),
		cache:         nil, // No cache
	}

	mockReg := &mockRegistry{
		versions: []string{"1.0.0"},
		data:     "test data",
		metadata: &Metadata{Name: "test"},
	}
	manager.registries["test"] = mockReg

	// All operations should still work without cache
	versions, err := manager.CachedListVersions("test", "typescript-rules")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}

	reader, err := manager.CachedDownload("test", "typescript-rules", "1.0.0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	_ = reader.Close()

	metadata, err := manager.CachedGetMetadata("test", "typescript-rules")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if metadata.Name != "test" {
		t.Errorf("Expected name 'test', got %q", metadata.Name)
	}
}
