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