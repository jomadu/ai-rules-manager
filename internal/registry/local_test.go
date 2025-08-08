package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLocalRegistry(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "local-registry-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config := &RegistryConfig{
		Name:    "test-local",
		Type:    "local",
		URL:     tempDir,
		Timeout: 30 * time.Second,
	}

	registry, err := NewLocalRegistry(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if registry.GetName() != "test-local" {
		t.Errorf("Expected name 'test-local', got %s", registry.GetName())
	}

	if registry.GetType() != "local" {
		t.Errorf("Expected type 'local', got %s", registry.GetType())
	}

	// Test that path was converted to absolute
	if !filepath.IsAbs(registry.path) {
		t.Error("Expected absolute path")
	}
}

func TestNewLocalRegistryRelativePath(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "local-registry-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory to test relative paths
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	_ = os.Chdir(tempDir)

	// Create a subdirectory to use as relative path
	relativeDir := "registry"
	if err := os.Mkdir(relativeDir, 0o755); err != nil {
		t.Fatalf("Failed to create relative dir: %v", err)
	}

	config := &RegistryConfig{
		Name:    "test-local",
		Type:    "local",
		URL:     relativeDir,
		Timeout: 30 * time.Second,
	}

	registry, err := NewLocalRegistry(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test that relative path was converted to absolute
	if !filepath.IsAbs(registry.path) {
		t.Error("Expected absolute path")
	}

	expectedPath := filepath.Join(tempDir, relativeDir)
	// Resolve symlinks for comparison (macOS /var -> /private/var)
	expectedResolved, _ := filepath.EvalSymlinks(expectedPath)
	actualResolved, _ := filepath.EvalSymlinks(registry.path)
	if actualResolved != expectedResolved {
		t.Errorf("Expected path %s, got %s", expectedResolved, actualResolved)
	}
}

func TestNewLocalRegistryInvalidPath(t *testing.T) {
	config := &RegistryConfig{
		Name:    "test-local",
		Type:    "local",
		URL:     "/non/existent/path",
		Timeout: 30 * time.Second,
	}

	_, err := NewLocalRegistry(config)
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestLocalRegistry_GetRulesets(t *testing.T) {
	// Create temp directory structure
	tempDir, err := os.MkdirTemp("", "local-registry-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test registry structure
	setupTestRegistry(t, tempDir)

	config := &RegistryConfig{
		Name:    "test-local",
		Type:    "local",
		URL:     tempDir,
		Timeout: 30 * time.Second,
	}

	registry, err := NewLocalRegistry(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	rulesets, err := registry.GetRulesets(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rulesets) != 2 {
		t.Errorf("Expected 2 rulesets, got %d", len(rulesets))
	}

	// Check that we get the expected rulesets
	rulesetNames := make(map[string]bool)
	for _, ruleset := range rulesets {
		rulesetNames[ruleset.Name] = true
		if ruleset.Type != "local" {
			t.Errorf("Expected type 'local', got %s", ruleset.Type)
		}
		if ruleset.Registry != "test-local" {
			t.Errorf("Expected registry 'test-local', got %s", ruleset.Registry)
		}
	}

	if !rulesetNames["python-rules"] {
		t.Error("Expected to find 'python-rules' ruleset")
	}
	if !rulesetNames["js-rules"] {
		t.Error("Expected to find 'js-rules' ruleset")
	}
}

func TestLocalRegistry_GetRuleset(t *testing.T) {
	// Create temp directory structure
	tempDir, err := os.MkdirTemp("", "local-registry-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test registry structure
	setupTestRegistry(t, tempDir)

	config := &RegistryConfig{
		Name:    "test-local",
		Type:    "local",
		URL:     tempDir,
		Timeout: 30 * time.Second,
	}

	registry, err := NewLocalRegistry(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test getting specific version
	ruleset, err := registry.GetRuleset(context.Background(), "python-rules", "1.0.0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ruleset.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", ruleset.Version)
	}

	// Test getting latest version
	ruleset, err = registry.GetRuleset(context.Background(), "python-rules", "latest")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ruleset.Version != "1.2.0" {
		t.Errorf("Expected latest version 1.2.0, got %s", ruleset.Version)
	}

	// Test non-existent ruleset
	_, err = registry.GetRuleset(context.Background(), "non-existent", "1.0.0")
	if err == nil {
		t.Error("Expected error for non-existent ruleset")
	}

	// Test non-existent version
	_, err = registry.GetRuleset(context.Background(), "python-rules", "999.0.0")
	if err == nil {
		t.Error("Expected error for non-existent version")
	}
}

func TestLocalRegistry_DownloadRuleset(t *testing.T) {
	// Create temp directory structure
	tempDir, err := os.MkdirTemp("", "local-registry-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test registry structure
	setupTestRegistry(t, tempDir)

	config := &RegistryConfig{
		Name:    "test-local",
		Type:    "local",
		URL:     tempDir,
		Timeout: 30 * time.Second,
	}

	registry, err := NewLocalRegistry(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Create destination directory
	destDir, err := os.MkdirTemp("", "download-test")
	if err != nil {
		t.Fatalf("Failed to create dest dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(destDir) }()

	// Test download
	err = registry.DownloadRuleset(context.Background(), "python-rules", "1.0.0", destDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check file was created
	filePath := filepath.Join(destDir, "ruleset.tar.gz")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected ruleset.tar.gz to be created")
	}

	// Check file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != "fake python rules content" {
		t.Errorf("Expected 'fake python rules content', got %s", string(content))
	}
}

func TestLocalRegistry_GetVersions(t *testing.T) {
	// Create temp directory structure
	tempDir, err := os.MkdirTemp("", "local-registry-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test registry structure
	setupTestRegistry(t, tempDir)

	config := &RegistryConfig{
		Name:    "test-local",
		Type:    "local",
		URL:     tempDir,
		Timeout: 30 * time.Second,
	}

	registry, err := NewLocalRegistry(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test getting versions for existing ruleset
	versions, err := registry.GetVersions(context.Background(), "python-rules")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(versions))
	}

	// Check that we have the expected versions
	versionMap := make(map[string]bool)
	for _, version := range versions {
		versionMap[version] = true
	}

	expectedVersions := []string{"1.0.0", "1.1.0", "1.2.0"}
	for _, expected := range expectedVersions {
		if !versionMap[expected] {
			t.Errorf("Expected to find version %s", expected)
		}
	}

	// Test non-existent ruleset
	_, err = registry.GetVersions(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent ruleset")
	}
}

// setupTestRegistry creates a test registry structure in the given directory
func setupTestRegistry(t *testing.T, baseDir string) {
	// Create python-rules with multiple versions
	pythonDir := filepath.Join(baseDir, "python-rules")
	if err := os.MkdirAll(pythonDir, 0o755); err != nil {
		t.Fatalf("Failed to create python-rules dir: %v", err)
	}

	versions := []string{"1.0.0", "1.1.0", "1.2.0"}
	for _, version := range versions {
		versionDir := filepath.Join(pythonDir, version)
		if err := os.MkdirAll(versionDir, 0o755); err != nil {
			t.Fatalf("Failed to create version dir: %v", err)
		}

		rulesetFile := filepath.Join(versionDir, "ruleset.tar.gz")
		if err := os.WriteFile(rulesetFile, []byte("fake python rules content"), 0o644); err != nil {
			t.Fatalf("Failed to create ruleset file: %v", err)
		}
	}

	// Create js-rules with single version
	jsDir := filepath.Join(baseDir, "js-rules")
	if err := os.MkdirAll(jsDir, 0o755); err != nil {
		t.Fatalf("Failed to create js-rules dir: %v", err)
	}

	jsVersionDir := filepath.Join(jsDir, "2.0.0")
	if err := os.MkdirAll(jsVersionDir, 0o755); err != nil {
		t.Fatalf("Failed to create js version dir: %v", err)
	}

	jsRulesetFile := filepath.Join(jsVersionDir, "ruleset.tar.gz")
	if err := os.WriteFile(jsRulesetFile, []byte("fake js rules content"), 0o644); err != nil {
		t.Fatalf("Failed to create js ruleset file: %v", err)
	}

	// Create a directory without ruleset.tar.gz (should be ignored)
	emptyDir := filepath.Join(baseDir, "empty-rules", "1.0.0")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}
}
