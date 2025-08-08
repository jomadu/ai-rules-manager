package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/max-dunn/ai-rules-manager/internal/config"
)

func TestHandleConfigSet(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Test setting a configuration value
	err = handleConfigSet("git.concurrency", "5", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the file was created and contains the value
	content, err := os.ReadFile(".armrc")
	if err != nil {
		t.Fatalf("Failed to read .armrc: %v", err)
	}

	if !strings.Contains(string(content), "[git]") {
		t.Error("Expected [git] section in .armrc")
	}
	if !strings.Contains(string(content), "concurrency = 5") {
		t.Error("Expected concurrency = 5 in .armrc")
	}
}

func TestHandleAddRegistry(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Test adding a Git registry
	err = handleAddRegistry("my-git", "https://github.com/user/repo", "git", false, map[string]string{
		"authToken": "test-token",
		"apiType":   "github",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the file was created and contains the registry
	content, err := os.ReadFile(".armrc")
	if err != nil {
		t.Fatalf("Failed to read .armrc: %v", err)
	}

	expectedContent := []string{
		"[registries]",
		"my-git",
		"https://github.com/user/repo",
		"[registries.my-git]",
		"type",
		"git",
		"authToken",
		"test-token",
		"apiType",
		"github",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Expected '%s' in .armrc, got:\n%s", expected, string(content))
		}
	}
}

func TestHandleAddChannel(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Test adding a channel
	err = handleAddChannel("cursor", ".cursor/rules,custom/cursor", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the file was created and contains the channel
	content, err := os.ReadFile("arm.json")
	if err != nil {
		t.Fatalf("Failed to read arm.json: %v", err)
	}

	if !strings.Contains(string(content), "cursor") {
		t.Error("Expected 'cursor' channel in arm.json")
	}
	if !strings.Contains(string(content), ".cursor/rules") {
		t.Error("Expected '.cursor/rules' directory in arm.json")
	}
	if !strings.Contains(string(content), "custom/cursor") {
		t.Error("Expected 'custom/cursor' directory in arm.json")
	}
}

func TestHandleRemoveRegistry(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// First add a registry
	err = handleAddRegistry("test-registry", "https://example.com", "https", false, map[string]string{})
	if err != nil {
		t.Fatalf("Failed to add registry: %v", err)
	}

	// Then remove it
	err = handleRemoveRegistry("test-registry", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the registry was removed
	content, err := os.ReadFile(".armrc")
	if err != nil {
		t.Fatalf("Failed to read .armrc: %v", err)
	}

	if strings.Contains(string(content), "test-registry") {
		t.Error("Expected 'test-registry' to be removed from .armrc")
	}
}

func TestHandleRemoveChannel(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// First add a channel
	err = handleAddChannel("test-channel", "test/dir", false)
	if err != nil {
		t.Fatalf("Failed to add channel: %v", err)
	}

	// Then remove it
	err = handleRemoveChannel("test-channel", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the channel was removed
	content, err := os.ReadFile("arm.json")
	if err != nil {
		t.Fatalf("Failed to read arm.json: %v", err)
	}

	if strings.Contains(string(content), "test-channel") {
		t.Error("Expected 'test-channel' to be removed from arm.json")
	}
}

func TestGetConfigValue(t *testing.T) {
	cfg := &config.Config{
		Registries: map[string]string{
			"default": "https://github.com/user/repo",
		},
		RegistryConfigs: map[string]map[string]string{
			"default": {
				"type":      "git",
				"authToken": "test-token",
			},
		},
		TypeDefaults: map[string]map[string]string{
			"git": {
				"concurrency": "1",
				"rateLimit":   "10/minute",
			},
		},
		NetworkConfig: map[string]string{
			"timeout": "30",
		},
		CacheConfig: map[string]string{
			"path": "~/.arm/cache",
		},
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"registries.default", "https://github.com/user/repo"},
		{"registries.default.type", "git"},
		{"registries.default.authToken", "test-token"},
		{"git.concurrency", "1"},
		{"git.rateLimit", "10/minute"},
		{"network.timeout", "30"},
		{"cache.path", "~/.arm/cache"},
		{"nonexistent.key", ""},
	}

	for _, test := range tests {
		result := getConfigValue(cfg, test.key)
		if result != test.expected {
			t.Errorf("getConfigValue(%s) = %s, expected %s", test.key, result, test.expected)
		}
	}
}

func TestGetConfigPath(t *testing.T) {
	// Test local path
	localPath := getConfigPath(".armrc", false)
	if localPath != ".armrc" {
		t.Errorf("Expected '.armrc', got %s", localPath)
	}

	// Test global path
	globalPath := getConfigPath(".armrc", true)
	expectedGlobal := filepath.Join(os.Getenv("HOME"), ".arm", ".armrc")
	if globalPath != expectedGlobal {
		t.Errorf("Expected %s, got %s", expectedGlobal, globalPath)
	}
}

func TestParseRulesetSpec(t *testing.T) {
	tests := []struct {
		spec             string
		expectedRegistry string
		expectedName     string
		expectedVersion  string
	}{
		{"my-rules", "", "my-rules", "latest"},
		{"my-rules@1.0.0", "", "my-rules", "1.0.0"},
		{"registry/my-rules", "registry", "my-rules", "latest"},
		{"registry/my-rules@1.0.0", "registry", "my-rules", "1.0.0"},
	}

	for _, test := range tests {
		registry, name, version := parseRulesetSpec(test.spec)
		if registry != test.expectedRegistry {
			t.Errorf("parseRulesetSpec(%s) registry = %s, expected %s", test.spec, registry, test.expectedRegistry)
		}
		if name != test.expectedName {
			t.Errorf("parseRulesetSpec(%s) name = %s, expected %s", test.spec, name, test.expectedName)
		}
		if version != test.expectedVersion {
			t.Errorf("parseRulesetSpec(%s) version = %s, expected %s", test.spec, version, test.expectedVersion)
		}
	}
}

func TestHandleInstallFromManifest(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "install-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Test with no configuration (should generate stubs)
	err = handleInstallFromManifest(false, true, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify stub files were created
	if _, err := os.Stat(".armrc"); os.IsNotExist(err) {
		t.Error("Expected .armrc to be created")
	}
	if _, err := os.Stat("arm.json"); os.IsNotExist(err) {
		t.Error("Expected arm.json to be created")
	}
}

func TestHandleInstallRuleset(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "install-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create basic configuration
	armrcContent := `[registries]
default = https://github.com/user/repo

[registries.default]
type = git
`
	err = os.WriteFile(".armrc", []byte(armrcContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create .armrc: %v", err)
	}

	armJSONContent := `{"engines":{"arm":"^1.0.0"},"channels":{},"rulesets":{}}`
	err = os.WriteFile("arm.json", []byte(armJSONContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create arm.json: %v", err)
	}

	// Test installing from default registry (should require patterns for Git)
	err = handleInstallRuleset("my-rules", false, true, "", "")
	if err == nil {
		t.Error("Expected error for Git registry without patterns")
	}

	// Test with patterns (dry run should succeed)
	err = handleInstallRuleset("my-rules", false, true, "", "*.md")
	if err != nil {
		t.Fatalf("Expected no error with patterns, got %v", err)
	}

	// Test with specific registry
	err = handleInstallRuleset("default/my-rules@1.0.0", false, true, "", "*.md")
	if err != nil {
		t.Fatalf("Expected no error with specific registry, got %v", err)
	}
}

func TestHandleSearch(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "search-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create basic configuration
	armrcContent := `[registries]
default = https://github.com/user/repo
my-git = https://github.com/other/repo

[registries.default]
type = git

[registries.my-git]
type = git
`
	err = os.WriteFile(".armrc", []byte(armrcContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create .armrc: %v", err)
	}

	armJSONContent := `{"engines":{"arm":"^1.0.0"},"channels":{},"rulesets":{}}`
	err = os.WriteFile("arm.json", []byte(armJSONContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create arm.json: %v", err)
	}

	// Test search with no registry filter
	err = handleSearch("python", "", false, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test search with specific registry
	err = handleSearch("python", "default", false, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test search with glob pattern
	err = handleSearch("python", "my-*", false, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHandleInfo(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "info-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create basic configuration
	armrcContent := `[registries]
default = https://github.com/user/repo

[registries.default]
type = git
`
	err = os.WriteFile(".armrc", []byte(armrcContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create .armrc: %v", err)
	}

	armJSONContent := `{"engines":{"arm":"^1.0.0"},"channels":{},"rulesets":{}}`
	err = os.WriteFile("arm.json", []byte(armJSONContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create arm.json: %v", err)
	}

	// Test info with default registry
	err = handleInfo("my-rules", false, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test info with specific registry and version
	err = handleInfo("default/my-rules@1.0.0", false, true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test info with JSON output
	err = handleInfo("my-rules", true, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHandleList(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "list-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create configuration with rulesets
	armrcContent := `[registries]
default = https://github.com/user/repo

[registries.default]
type = git
`
	err = os.WriteFile(".armrc", []byte(armrcContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create .armrc: %v", err)
	}

	armJSONContent := `{
  "engines": {"arm": "^1.0.0"},
  "channels": {},
  "rulesets": {
    "default": {
      "my-rules": {"version": "^1.0.0", "patterns": ["*.md"]},
      "python-rules": {"version": "~2.1.0"}
    }
  }
}`
	err = os.WriteFile("arm.json", []byte(armJSONContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create arm.json: %v", err)
	}

	// Test list with rulesets
	err = handleList(false, false, false, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test list with JSON output
	err = handleList(false, false, true, "cursor")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestGetTargetRegistries(t *testing.T) {
	allRegistries := map[string]string{
		"default":   "https://github.com/user/repo",
		"my-git":    "https://github.com/other/repo",
		"my-s3":     "my-bucket",
		"test-repo": "https://test.com/repo",
	}

	tests := []struct {
		filter   string
		expected []string
	}{
		{"", []string{"default", "my-git", "my-s3", "test-repo"}},
		{"default", []string{"default"}},
		{"default,my-git", []string{"default", "my-git"}},
		{"my-*", []string{"my-git", "my-s3"}},
		{"*-repo", []string{"test-repo"}},
		{"nonexistent", []string{}},
	}

	for _, test := range tests {
		result := getTargetRegistries(allRegistries, test.filter)
		if len(result) != len(test.expected) {
			t.Errorf("getTargetRegistries(%s) returned %d items, expected %d", test.filter, len(result), len(test.expected))
			continue
		}

		// Convert to map for easier comparison
		resultMap := make(map[string]bool)
		for _, r := range result {
			resultMap[r] = true
		}

		for _, expected := range test.expected {
			if !resultMap[expected] {
				t.Errorf("getTargetRegistries(%s) missing expected result: %s", test.filter, expected)
			}
		}
	}
}
