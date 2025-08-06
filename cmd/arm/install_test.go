package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallWorkflows(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(tempDir string) error
		args           []string
		expectedFiles  []string
		expectedOutput string
		expectError    bool
	}{
		{
			name: "scenario_1a_no_config_files_install",
			setupFiles: func(tempDir string) error {
				// No files created - clean directory
				return nil
			},
			args:          []string{},
			expectedFiles: []string{".armrc", "rules.json"},
			expectedOutput: "Created stub configuration files (.armrc and rules.json).\nPlease configure your registries in .armrc and add dependencies to rules.json.",
			expectError:   false,
		},
		{
			name: "scenario_1b_no_config_files_install_ruleset",
			setupFiles: func(tempDir string) error {
				// No files created - clean directory
				return nil
			},
			args:          []string{"typescript-rules@1.0.0"},
			expectedFiles: []string{".armrc", "rules.json"},
			expectedOutput: "Created stub configuration files (.armrc and rules.json).\nRuleset typescript-rules@1.0.0 was not installed due to missing source configuration.\nPlease configure your registries in .armrc and run the install command again.",
			expectError:   false,
		},
		{
			name: "scenario_3a_has_armrc_no_rules_install",
			setupFiles: func(tempDir string) error {
				configContent := `[sources]
default = https://registry.armjs.org/`
				return os.WriteFile(filepath.Join(tempDir, ".armrc"), []byte(configContent), 0644)
			},
			args:          []string{},
			expectedFiles: []string{".armrc", "rules.json"},
			expectedOutput: "Created stub rules.json file.\nPlease add dependencies to rules.json and run 'arm install' again.",
			expectError:   false,
		},
		{
			name: "scenario_3b_has_armrc_no_rules_install_ruleset",
			setupFiles: func(tempDir string) error {
				configContent := `[sources]
default = https://registry.armjs.org/`
				return os.WriteFile(filepath.Join(tempDir, ".armrc"), []byte(configContent), 0644)
			},
			args:          []string{"security-rules@2.1.0"},
			expectedFiles: []string{".armrc", "rules.json"},
			expectedOutput: "Created rules.json file with the specified ruleset.\nInstalling security-rules@2.1.0...",
			expectError:   true, // Will fail due to network call, but that's expected
		},
		{
			name: "scenario_4a_has_rules_no_armrc_install",
			setupFiles: func(tempDir string) error {
				manifestContent := `{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "existing-rules": "1.0.0"
  }
}`
				return os.WriteFile(filepath.Join(tempDir, "rules.json"), []byte(manifestContent), 0644)
			},
			args:          []string{},
			expectedFiles: []string{".armrc", "rules.json"},
			expectedOutput: "No registry sources configured. Please configure a source in .armrc file.\nCreated stub .armrc file in",
			expectError:   true,
		},
		{
			name: "scenario_4b_has_rules_no_armrc_install_ruleset",
			setupFiles: func(tempDir string) error {
				manifestContent := `{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "existing-rules": "1.0.0"
  }
}`
				return os.WriteFile(filepath.Join(tempDir, "rules.json"), []byte(manifestContent), 0644)
			},
			args:          []string{"new-rules@1.5.0"},
			expectedFiles: []string{".armrc", "rules.json"},
			expectedOutput: "No registry sources configured. Please configure a source in .armrc file.\nCreated stub .armrc file in",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)

			err = os.Chdir(tempDir)
			require.NoError(t, err)

			t.Cleanup(func() {
				_ = os.Chdir(originalDir)
			})

			// Setup test files
			err = tt.setupFiles(tempDir)
			require.NoError(t, err)

			// Run the install command
			var result error
			if len(tt.args) == 0 {
				result = installFromManifest()
			} else {
				result = installRuleset(tt.args[0])
			}

			// Check error expectation
			if tt.expectError {
				assert.Error(t, result)
			} else {
				assert.NoError(t, result)
			}

			// Verify expected files were created
			for _, file := range tt.expectedFiles {
				assert.True(t, fileExists(file), "Expected file %s to exist", file)
			}

			// Verify stub file contents only for scenarios that create stub files
			if strings.Contains(tt.name, "scenario_1") || strings.Contains(tt.name, "scenario_4") {
				if contains(tt.expectedFiles, ".armrc") {
					content, err := os.ReadFile(".armrc")
					require.NoError(t, err)
					assert.Contains(t, string(content), "# Example configuration for ARM")
					assert.Contains(t, string(content), "# [sources]")
					assert.Contains(t, string(content), "# my-rules = https://github.com/username/my-rules")
				}
			}

			if strings.Contains(tt.name, "scenario_1") || strings.Contains(tt.name, "scenario_3") {
				if contains(tt.expectedFiles, "rules.json") {
					content, err := os.ReadFile("rules.json")
					require.NoError(t, err)
					assert.Contains(t, string(content), `"targets"`)
					assert.Contains(t, string(content), `".cursorrules"`)
					assert.Contains(t, string(content), `".amazonq/rules"`)
				}
			}
		})
	}
}

func TestStubFileCreation(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	t.Run("createStubConfig", func(t *testing.T) {
		err := createStubConfig()
		require.NoError(t, err)

		content, err := os.ReadFile(".armrc")
		require.NoError(t, err)

		assert.Contains(t, string(content), "# Example configuration for ARM (AI Rules Manager)")
		assert.Contains(t, string(content), "# [sources]")
		assert.Contains(t, string(content), "# my-rules = https://github.com/username/my-rules")
		assert.Contains(t, string(content), "# type = git")
		assert.Contains(t, string(content), "# api = github")
		assert.Contains(t, string(content), "# # authToken = $GITHUB_TOKEN")
	})

	t.Run("createStubManifest", func(t *testing.T) {
		err := createStubManifest()
		require.NoError(t, err)

		content, err := os.ReadFile("rules.json")
		require.NoError(t, err)

		assert.Contains(t, string(content), `"targets"`)
		assert.Contains(t, string(content), `".cursorrules"`)
		assert.Contains(t, string(content), `".amazonq/rules"`)
		assert.Contains(t, string(content), `"dependencies": {}`)
	})

	t.Run("createManifestWithRuleset", func(t *testing.T) {
		err := createManifestWithRuleset("test-rules", "2.0.0")
		require.NoError(t, err)

		content, err := os.ReadFile("rules.json")
		require.NoError(t, err)

		assert.Contains(t, string(content), `"test-rules": "2.0.0"`)
		assert.Contains(t, string(content), `".cursorrules"`)
		assert.Contains(t, string(content), `".amazonq/rules"`)
	})

	t.Run("createStubFiles", func(t *testing.T) {
		// Clean up first
		_ = os.Remove(".armrc")
		_ = os.Remove("rules.json")

		err := createStubFiles()
		require.NoError(t, err)

		assert.True(t, fileExists(".armrc"))
		assert.True(t, fileExists("rules.json"))
	})
}

func TestParseRulesetSpec(t *testing.T) {
	tests := []struct {
		input           string
		expectedName    string
		expectedVersion string
	}{
		{"typescript-rules", "typescript-rules", "latest"},
		{"typescript-rules@1.0.0", "typescript-rules", "1.0.0"},
		{"company@typescript-rules@2.1.0", "company@typescript-rules", "2.1.0"},
		{"@company/rules", "", "company/rules"},
		{"simple", "simple", "latest"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, version := parseRulesetSpec(tt.input)
			assert.Equal(t, tt.expectedName, name)
			assert.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestConfigDetection(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	t.Run("fileExists", func(t *testing.T) {
		assert.False(t, fileExists("nonexistent.txt"))

		err := os.WriteFile("test.txt", []byte("test"), 0644)
		require.NoError(t, err)
		assert.True(t, fileExists("test.txt"))
	})

	t.Run("hasValidConfig_no_config", func(t *testing.T) {
		assert.False(t, hasValidConfig())
	})

	t.Run("hasValidConfig_empty_config", func(t *testing.T) {
		err := os.WriteFile(".armrc", []byte("# Just comments"), 0644)
		require.NoError(t, err)
		assert.False(t, hasValidConfig())
	})

	t.Run("hasValidConfig_valid_config", func(t *testing.T) {
		configContent := `[sources]
default = https://registry.armjs.org/`
		err := os.WriteFile(".armrc", []byte(configContent), 0644)
		require.NoError(t, err)
		assert.True(t, hasValidConfig())
	})
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}