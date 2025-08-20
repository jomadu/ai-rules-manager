package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/max-dunn/ai-rules-manager/internal/config"
	"github.com/max-dunn/ai-rules-manager/internal/update"
)

func TestHandleConfigSet(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test setting a configuration value
	err = handleConfigSet("git.concurrency", "5", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the file was created and contains the value
	content, err := os.ReadFile(".armrc.json")
	if err != nil {
		t.Fatalf("Failed to read .armrc.json: %v", err)
	}

	if !strings.Contains(string(content), `"git"`) {
		t.Error("Expected git section in .armrc.json")
	}
	if !strings.Contains(string(content), `"concurrency": "5"`) {
		t.Error("Expected concurrency = 5 in .armrc.json")
	}
}

func TestHandleAddRegistry(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test adding a Git registry
	err = handleAddRegistry("my-git", "https://github.com/user/repo", "git", false, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the file was created and contains the registry
	content, err := os.ReadFile(".armrc.json")
	if err != nil {
		t.Fatalf("Failed to read .armrc.json: %v", err)
	}

	expectedContent := []string{
		`"registries"`,
		`"my-git"`,
		`"url": "https://github.com/user/repo"`,
		`"type": "git"`,
	}

	for _, expected := range expectedContent {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Expected '%s' in .armrc.json, got:\n%s", expected, string(content))
		}
	}
}

func TestHandleAddChannel(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test adding a channel
	err = handleAddChannel("cursor", ".cursor/rules,custom/cursor", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the file was created and contains the channel
	content, err := os.ReadFile(".armrc.json")
	if err != nil {
		t.Fatalf("Failed to read .armrc.json: %v", err)
	}

	if !strings.Contains(string(content), `"cursor"`) {
		t.Error("Expected 'cursor' channel in .armrc.json")
	}
	if !strings.Contains(string(content), `".cursor/rules"`) && !strings.Contains(string(content), `"custom/cursor"`) {
		t.Error("Expected '.cursor/rules' and 'custom/cursor' directories in .armrc.json")
	}
}

func TestHandleRemoveRegistry(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

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
	content, err := os.ReadFile(".armrc.json")
	if err != nil {
		t.Fatalf("Failed to read .armrc.json: %v", err)
	}

	if strings.Contains(string(content), "test-registry") {
		t.Error("Expected 'test-registry' to be removed from .armrc.json")
	}
}

func TestHandleRemoveChannel(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

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
	content, err := os.ReadFile(".armrc.json")
	if err != nil {
		t.Fatalf("Failed to read .armrc.json: %v", err)
	}

	if strings.Contains(string(content), "test-channel") {
		t.Error("Expected 'test-channel' to be removed from .armrc.json")
	}
}

func TestGetConfigValue(t *testing.T) {
	cfg := &config.Config{
		Registries: map[string]string{
			"default": "https://github.com/user/repo",
		},
		RegistryConfigs: map[string]map[string]string{
			"default": {
				"type": "git",
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
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"registries.default", "https://github.com/user/repo"},
		{"registries.default.type", "git"},

		{"git.concurrency", "1"},
		{"git.rateLimit", "10/minute"},
		{"network.timeout", "30"},

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
	localPath := getConfigPath(".armrc.json", false)
	if localPath != ".armrc.json" {
		t.Errorf("Expected '.armrc.json', got %s", localPath)
	}

	// Test global path
	globalPath := getConfigPath(".armrc.json", true)
	expectedGlobal := filepath.Join(os.Getenv("HOME"), ".arm", ".armrc.json")
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
		{"my-rules", "", "my-rules", ""},
		{"my-rules@1.0.0", "", "my-rules", "1.0.0"},
		{"registry/my-rules", "registry", "my-rules", ""},
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
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test with no configuration (should generate stubs)
	err = handleInstallFromManifest(false, true, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify stub files were created
	if _, err := os.Stat(".armrc.json"); os.IsNotExist(err) {
		t.Error("Expected .armrc.json to be created")
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
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Create basic configuration
	armrcContent := `{
  "registries": {
    "default": {
      "url": "https://github.com/user/repo",
      "type": "git"
    }
  },
  "channels": {}
}`
	err = os.WriteFile(".armrc.json", []byte(armrcContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create .armrc.json: %v", err)
	}

	armJSONContent := `{"engines":{"arm":"^1.0.0"},"channels":{},"rulesets":{}}`
	err = os.WriteFile("arm.json", []byte(armJSONContent), 0o600)
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

func TestHandleInfo(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "info-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Create basic configuration
	armrcContent := `{
  "registries": {
    "default": {
      "url": "https://github.com/user/repo",
      "type": "git"
    }
  },
  "channels": {}
}`
	err = os.WriteFile(".armrc.json", []byte(armrcContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create .armrc.json: %v", err)
	}

	armJSONContent := `{"engines":{"arm":"^1.0.0"},"channels":{},"rulesets":{}}`
	err = os.WriteFile("arm.json", []byte(armJSONContent), 0o600)
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
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Create configuration with rulesets
	armrcContent := `{
  "registries": {
    "default": {
      "url": "https://github.com/user/repo",
      "type": "git"
    }
  },
  "channels": {}
}`
	err = os.WriteFile(".armrc.json", []byte(armrcContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create .armrc.json: %v", err)
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
	err = os.WriteFile("arm.json", []byte(armJSONContent), 0o600)
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

// Mock update service for testing
type mockUpdateService struct {
	updateResults map[string]*update.UpdateResult
	updateErrors  map[string]error
}

func (m *mockUpdateService) UpdateRuleset(ctx context.Context, rulesetSpec string) (*update.UpdateResult, error) {
	if err, exists := m.updateErrors[rulesetSpec]; exists {
		return nil, err
	}
	if result, exists := m.updateResults[rulesetSpec]; exists {
		return result, nil
	}
	return &update.UpdateResult{
		Registry:        "default",
		Ruleset:         "unknown",
		Version:         "1.0.0",
		PreviousVersion: "1.0.0",
		Updated:         false,
	}, nil
}

// updateServiceInterface allows mocking the update service
type updateServiceInterface interface {
	UpdateRuleset(ctx context.Context, rulesetSpec string) (*update.UpdateResult, error)
}

// NetworkError represents a network-related error for testing
type NetworkError struct {
	Message string
}

func (e *NetworkError) Error() string {
	return e.Message
}

func TestHandleOutdated(t *testing.T) {
	tests := []struct {
		name           string
		setupLockFile  func(tempDir string) error
		mockResults    map[string]*update.UpdateResult
		mockErrors     map[string]error
		jsonOutput     bool
		expectError    bool
		expectedOutput []string
	}{
		{
			name: "no lock file",
			setupLockFile: func(tempDir string) error {
				return nil // Don't create lock file
			},
			expectError:    true,
			expectedOutput: []string{"no lock file found"},
		},
		{
			name: "all rulesets up to date",
			setupLockFile: func(tempDir string) error {
				lockFile := config.LockFile{
					Rulesets: map[string]map[string]config.LockedRuleset{
						"default": {
							"my-rules":     {Version: "1.0.0"},
							"python-rules": {Version: "2.1.0"},
						},
					},
				}
				data, _ := json.MarshalIndent(lockFile, "", "  ")
				return os.WriteFile(filepath.Join(tempDir, "arm.lock"), data, 0o600)
			},
			mockResults: map[string]*update.UpdateResult{
				"default/my-rules": {
					Registry:        "default",
					Ruleset:         "my-rules",
					Version:         "1.0.0",
					PreviousVersion: "1.0.0",
					Updated:         false,
				},
				"default/python-rules": {
					Registry:        "default",
					Ruleset:         "python-rules",
					Version:         "2.1.0",
					PreviousVersion: "2.1.0",
					Updated:         false,
				},
			},
			expectedOutput: []string{"All rulesets are up to date"},
		},
		{
			name: "some rulesets outdated - human readable",
			setupLockFile: func(tempDir string) error {
				lockFile := config.LockFile{
					Rulesets: map[string]map[string]config.LockedRuleset{
						"default": {
							"my-rules":     {Version: "1.0.0"},
							"python-rules": {Version: "2.0.0"},
						},
						"my-git": {
							"js-rules": {Version: "abc123"},
						},
					},
				}
				data, _ := json.MarshalIndent(lockFile, "", "  ")
				return os.WriteFile(filepath.Join(tempDir, "arm.lock"), data, 0o600)
			},
			mockResults: map[string]*update.UpdateResult{
				"default/my-rules": {
					Registry:        "default",
					Ruleset:         "my-rules",
					Version:         "1.2.0",
					PreviousVersion: "1.0.0",
					Updated:         true,
				},
				"default/python-rules": {
					Registry:        "default",
					Ruleset:         "python-rules",
					Version:         "2.0.0",
					PreviousVersion: "2.0.0",
					Updated:         false,
				},
				"my-git/js-rules": {
					Registry:        "my-git",
					Ruleset:         "js-rules",
					Version:         "def456",
					PreviousVersion: "abc123",
					Updated:         true,
				},
			},
			expectedOutput: []string{
				"Found 2 outdated ruleset(s)",
				"default/my-rules",
				"Current: 1.0.0",
				"Latest:  1.2.0",
				"Update:  arm update default/my-rules",
				"my-git/js-rules",
				"Current: abc123",
				"Latest:  def456",
				"Update:  arm update my-git/js-rules",
			},
		},
		{
			name: "network errors ignored",
			setupLockFile: func(tempDir string) error {
				lockFile := config.LockFile{
					Rulesets: map[string]map[string]config.LockedRuleset{
						"default": {
							"my-rules":      {Version: "1.0.0"},
							"failing-rules": {Version: "1.0.0"},
						},
					},
				}
				data, _ := json.MarshalIndent(lockFile, "", "  ")
				return os.WriteFile(filepath.Join(tempDir, "arm.lock"), data, 0o600)
			},
			mockResults: map[string]*update.UpdateResult{
				"default/my-rules": {
					Registry:        "default",
					Ruleset:         "my-rules",
					Version:         "1.2.0",
					PreviousVersion: "1.0.0",
					Updated:         true,
				},
			},
			mockErrors: map[string]error{
				"default/failing-rules": &NetworkError{Message: "connection timeout"},
			},
			expectedOutput: []string{
				"Found 1 outdated ruleset(s)",
				"default/my-rules",
				"Current: 1.0.0",
				"Latest:  1.2.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir, err := os.MkdirTemp("", "outdated-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tempDir) }()

			// Change to temp directory
			originalWd, _ := os.Getwd()
			defer func() { _ = os.Chdir(originalWd) }()
			_ = os.Chdir(tempDir)

			// Setup test environment
			if err := tt.setupLockFile(tempDir); err != nil {
				t.Fatalf("Failed to setup lock file: %v", err)
			}

			// Create basic configuration
			armrcContent := `{
  "registries": {
    "default": {
      "url": "https://github.com/user/repo",
      "type": "git"
    },
    "my-git": {
      "url": "https://github.com/other/repo",
      "type": "git"
    }
  },
  "channels": {}
}`
			err = os.WriteFile(".armrc.json", []byte(armrcContent), 0o600)
			if err != nil {
				t.Fatalf("Failed to create .armrc.json: %v", err)
			}

			// Test with mock service
			mockService := &mockUpdateService{
				updateResults: tt.mockResults,
				updateErrors:  tt.mockErrors,
			}

			// Execute the command with mock
			err = handleOutdatedWithMockService(false, tt.jsonOutput, mockService)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else {
					// Verify error message contains expected text
					for _, expected := range tt.expectedOutput {
						if !strings.Contains(err.Error(), expected) {
							t.Errorf("Expected error to contain '%s', got: %v", expected, err)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestHandleOutdatedJSONValidation(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "outdated-json-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Setup lock file with outdated rulesets
	lockFile := config.LockFile{
		Rulesets: map[string]map[string]config.LockedRuleset{
			"default": {
				"my-rules": {Version: "1.0.0"},
			},
		},
	}
	data, _ := json.MarshalIndent(lockFile, "", "  ")
	err = os.WriteFile("arm.lock", data, 0o600)
	if err != nil {
		t.Fatalf("Failed to create lock file: %v", err)
	}

	// Create basic configuration
	armrcContent := `{
  "registries": {
    "default": {
      "url": "https://github.com/user/repo",
      "type": "git"
    }
  },
  "channels": {}
}`
	err = os.WriteFile(".armrc.json", []byte(armrcContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create .armrc.json: %v", err)
	}

	// Mock service with outdated ruleset
	mockService := &mockUpdateService{
		updateResults: map[string]*update.UpdateResult{
			"default/my-rules": {
				Registry:        "default",
				Ruleset:         "my-rules",
				Version:         "1.2.0",
				PreviousVersion: "1.0.0",
				Updated:         true,
			},
		},
	}

	// Execute with JSON output and capture result
	var capturedOutput string
	handleOutdatedWithCapture := func(_, jsonOutput bool, updateService updateServiceInterface) (string, error) {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return "", fmt.Errorf("failed to load configuration: %w", err)
		}

		// Check if we have a lock file
		if cfg.LockFile == nil {
			return "", fmt.Errorf("no lock file found - no rulesets installed")
		}

		type outdatedInfo struct {
			Registry       string `json:"registry"`
			Name           string `json:"name"`
			CurrentVersion string `json:"current_version"`
			LatestVersion  string `json:"latest_version"`
			UpdateCommand  string `json:"update_command"`
		}

		var outdatedRulesets []outdatedInfo

		// Check each installed ruleset
		for registry, rulesets := range cfg.LockFile.Rulesets {
			for name := range rulesets {
				rulesetSpec := fmt.Sprintf("%s/%s", registry, name)
				result, err := updateService.UpdateRuleset(context.Background(), rulesetSpec)
				if err != nil {
					continue // Skip failed resolutions
				}
				if result.Updated {
					outdatedRulesets = append(outdatedRulesets, outdatedInfo{
						Registry:       result.Registry,
						Name:           result.Ruleset,
						CurrentVersion: result.PreviousVersion,
						LatestVersion:  result.Version,
						UpdateCommand:  fmt.Sprintf("arm update %s/%s", result.Registry, result.Ruleset),
					})
				}
			}
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(map[string]interface{}{
				"outdated": outdatedRulesets,
			}, "", "  ")
			return string(data), nil
		}

		return "", nil
	}

	// Execute with JSON output
	capturedOutput, err = handleOutdatedWithCapture(false, true, mockService)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	// Validate JSON structure
	var result map[string]interface{}
	err = json.Unmarshal([]byte(capturedOutput), &result)
	if err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Validate JSON schema
	outdated, exists := result["outdated"]
	if !exists {
		t.Error("JSON output missing 'outdated' field")
	}

	outdatedList, ok := outdated.([]interface{})
	if !ok {
		t.Error("'outdated' field is not an array")
	}

	if len(outdatedList) != 1 {
		t.Errorf("Expected 1 outdated ruleset, got %d", len(outdatedList))
	}

	// Validate first outdated ruleset structure
	firstRuleset, ok := outdatedList[0].(map[string]interface{})
	if !ok {
		t.Error("Outdated ruleset is not an object")
	}

	requiredFields := []string{"registry", "name", "current_version", "latest_version", "update_command"}
	for _, field := range requiredFields {
		if _, exists := firstRuleset[field]; !exists {
			t.Errorf("Missing required field '%s' in JSON output", field)
		}
	}

	// Validate field values
	if firstRuleset["registry"] != "default" {
		t.Errorf("Expected registry 'default', got %v", firstRuleset["registry"])
	}
	if firstRuleset["name"] != "my-rules" {
		t.Errorf("Expected name 'my-rules', got %v", firstRuleset["name"])
	}
	if firstRuleset["current_version"] != "1.0.0" {
		t.Errorf("Expected current_version '1.0.0', got %v", firstRuleset["current_version"])
	}
	if firstRuleset["latest_version"] != "1.2.0" {
		t.Errorf("Expected latest_version '1.2.0', got %v", firstRuleset["latest_version"])
	}
	if firstRuleset["update_command"] != "arm update default/my-rules" {
		t.Errorf("Expected update_command 'arm update default/my-rules', got %v", firstRuleset["update_command"])
	}
}

// handleOutdatedWithMockService is a testable version of handleOutdated that accepts a mock service
func handleOutdatedWithMockService(_, jsonOutput bool, updateService updateServiceInterface) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if we have a lock file
	if cfg.LockFile == nil {
		return fmt.Errorf("no lock file found - no rulesets installed")
	}

	type outdatedInfo struct {
		Registry       string `json:"registry"`
		Name           string `json:"name"`
		CurrentVersion string `json:"current_version"`
		LatestVersion  string `json:"latest_version"`
		UpdateCommand  string `json:"update_command"`
	}

	var outdatedRulesets []outdatedInfo

	// Check each installed ruleset
	for registry, rulesets := range cfg.LockFile.Rulesets {
		for name := range rulesets {
			rulesetSpec := fmt.Sprintf("%s/%s", registry, name)
			result, err := updateService.UpdateRuleset(context.Background(), rulesetSpec)
			if err != nil {
				continue // Skip failed resolutions
			}
			if result.Updated {
				outdatedRulesets = append(outdatedRulesets, outdatedInfo{
					Registry:       result.Registry,
					Name:           result.Ruleset,
					CurrentVersion: result.PreviousVersion,
					LatestVersion:  result.Version,
					UpdateCommand:  fmt.Sprintf("arm update %s/%s", result.Registry, result.Ruleset),
				})
			}
		}
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"outdated": outdatedRulesets,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(outdatedRulesets) == 0 {
		fmt.Println("All rulesets are up to date")
		return nil
	}

	fmt.Printf("Found %d outdated ruleset(s):\n\n", len(outdatedRulesets))
	for _, info := range outdatedRulesets {
		fmt.Printf("%s/%s\n", info.Registry, info.Name)
		fmt.Printf("  Current: %s\n", info.CurrentVersion)
		fmt.Printf("  Latest:  %s\n", info.LatestVersion)
		fmt.Printf("  Update:  %s\n\n", info.UpdateCommand)
	}

	return nil
}
