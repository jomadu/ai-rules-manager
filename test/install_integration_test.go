package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/max-dunn/ai-rules-manager/internal/config"
	"github.com/max-dunn/ai-rules-manager/internal/install"
)

func TestInstallWorkflow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test environment
	tempDir, err := os.MkdirTemp("", "arm-install-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory to ensure lock file is created there
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	// Create test configuration
	cfg := createTestConfig(t, tempDir)

	// Create test source files
	sourceFiles := createTestSourceFiles(t, tempDir)

	// Create installer
	installer := install.New(cfg)

	// Test install request
	req := &install.InstallRequest{
		Registry:    "test-registry",
		Ruleset:     "python-rules",
		Version:     "1.2.0",
		SourceFiles: sourceFiles,
		Channels:    []string{"cursor", "q"},
	}

	// Execute install
	result, err := installer.Install(req)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify install result
	verifyInstallResult(t, result, req, len(sourceFiles))

	// Verify files deployed to channels
	verifyChannelDeployment(t, tempDir, req)

	// Verify lock file updated
	verifyLockFile(t, installer, req)

	t.Logf("Install integration test completed successfully")
}

func TestInstallWorkflow_ErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tempDir, err := os.MkdirTemp("", "arm-install-error-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory to ensure lock file is created there
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		setupConfig func() *config.Config
		setupReq    func() *install.InstallRequest
		expectError string
	}{
		{
			name: "missing registry",
			setupConfig: func() *config.Config {
				return createTestConfig(t, tempDir)
			},
			setupReq: func() *install.InstallRequest {
				return &install.InstallRequest{
					Registry:    "",
					Ruleset:     "test-rules",
					Version:     "1.0.0",
					SourceFiles: []string{"test.md"},
				}
			},
			expectError: "registry, ruleset, and version are required",
		},
		{
			name: "no source files",
			setupConfig: func() *config.Config {
				return createTestConfig(t, tempDir)
			},
			setupReq: func() *install.InstallRequest {
				return &install.InstallRequest{
					Registry: "test-registry",
					Ruleset:  "test-rules",
					Version:  "1.0.0",
				}
			},
			expectError: "no source files provided",
		},
		{
			name: "no channels configured",
			setupConfig: func() *config.Config {
				cfg := createTestConfig(t, tempDir)
				cfg.Channels = make(map[string]config.ChannelConfig)
				return cfg
			},
			setupReq: func() *install.InstallRequest {
				return &install.InstallRequest{
					Registry:    "test-registry",
					Ruleset:     "test-rules",
					Version:     "1.0.0",
					SourceFiles: []string{"test.md"},
				}
			},
			expectError: "no channels configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			req := tt.setupReq()
			installer := install.New(cfg)

			_, err := installer.Install(req)
			if err == nil {
				t.Errorf("Expected error but got none")
			} else if err.Error() != tt.expectError {
				t.Errorf("Expected error %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

func createTestConfig(t *testing.T, tempDir string) *config.Config {
	t.Helper()

	// Create channel directories
	cursorDir := filepath.Join(tempDir, "cursor")
	qDir := filepath.Join(tempDir, "q")
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(qDir, 0o755); err != nil {
		t.Fatal(err)
	}

	return &config.Config{
		Registries: map[string]string{
			"test-registry": "https://github.com/test/repo",
		},
		RegistryConfigs: map[string]map[string]string{
			"test-registry": {
				"type": "git",
			},
		},
		Channels: map[string]config.ChannelConfig{
			"cursor": {
				Directories: []string{cursorDir},
			},
			"q": {
				Directories: []string{qDir},
			},
		},
	}
}

func createTestSourceFiles(t *testing.T, tempDir string) []string {
	t.Helper()

	// Create a temporary extraction directory that mimics Git registry extraction
	extractDir := filepath.Join(tempDir, "arm-install-12345")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"python.md":     "# Python Rules\nPython coding standards",
		"javascript.md": "# JavaScript Rules\nJS coding standards",
		"config.json":   `{"version": "1.2.0"}`,
	}

	var sourceFiles []string
	for filename, content := range files {
		filePath := filepath.Join(extractDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		sourceFiles = append(sourceFiles, filePath)
	}

	return sourceFiles
}

func verifyInstallResult(t *testing.T, result *install.InstallResult, req *install.InstallRequest, expectedFileCount int) {
	t.Helper()

	if result.Registry != req.Registry {
		t.Errorf("Expected registry %q, got %q", req.Registry, result.Registry)
	}
	if result.Ruleset != req.Ruleset {
		t.Errorf("Expected ruleset %q, got %q", req.Ruleset, result.Ruleset)
	}
	if result.Version != req.Version {
		t.Errorf("Expected version %q, got %q", req.Version, result.Version)
	}
	// Files are counted per channel, so multiply by number of channels
	expectedTotal := expectedFileCount * len(req.Channels)
	if result.FilesCount != expectedTotal {
		t.Errorf("Expected %d files (%d files × %d channels), got %d", expectedTotal, expectedFileCount, len(req.Channels), result.FilesCount)
	}
	if len(result.Channels) != len(req.Channels) {
		t.Errorf("Expected %d channels, got %d", len(req.Channels), len(result.Channels))
	}
}

func verifyChannelDeployment(t *testing.T, tempDir string, req *install.InstallRequest) {
	t.Helper()

	expectedFiles := []string{"python.md", "javascript.md", "config.json"}

	for _, channel := range req.Channels {
		channelDir := filepath.Join(tempDir, channel)
		versionDir := filepath.Join(channelDir, "arm", req.Registry, req.Ruleset, req.Version)

		// Check version directory exists
		if _, err := os.Stat(versionDir); os.IsNotExist(err) {
			t.Errorf("Version directory not created: %s", versionDir)
			continue
		}

		// Check files deployed (they're in the arm-install-12345 subdirectory)
		for _, expectedFile := range expectedFiles {
			// Files are in the arm-install-12345 subdirectory
			filePath := filepath.Join(versionDir, "arm-install-12345", expectedFile)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Expected file not deployed: %s", filePath)
			}
		}
	}
}

func verifyLockFile(t *testing.T, installer *install.Installer, req *install.InstallRequest) {
	t.Helper()

	lockFile, err := installer.GetLockFile()
	if err != nil {
		t.Fatalf("Failed to get lock file: %v", err)
	}

	if lockFile.Rulesets[req.Registry] == nil {
		t.Errorf("Registry %q not found in lock file", req.Registry)
		return
	}

	lockedRuleset := lockFile.Rulesets[req.Registry][req.Ruleset]
	if lockedRuleset.Version != req.Version {
		t.Errorf("Expected locked version %q, got %q", req.Version, lockedRuleset.Version)
	}
	if lockedRuleset.Type != "git" {
		t.Errorf("Expected locked type 'git', got %q", lockedRuleset.Type)
	}
}
