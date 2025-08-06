package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallWorkflowIntegration(t *testing.T) {
	// Build the ARM binary for testing
	armBinary := buildARMBinary(t)

	tests := []struct {
		name           string
		setupFiles     func(tempDir string) error
		command        []string
		expectedFiles  []string
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "scenario_1a_no_config_install",
			setupFiles: func(tempDir string) error {
				return nil // Clean directory
			},
			command:        []string{armBinary, "install"},
			expectedFiles:  []string{".armrc", "rules.json"},
			expectedOutput: []string{"Created stub configuration files", "Please configure your registries"},
			expectError:    false,
		},
		{
			name: "scenario_1b_no_config_install_ruleset",
			setupFiles: func(tempDir string) error {
				return nil // Clean directory
			},
			command:        []string{armBinary, "install", "test-rules@1.0.0"},
			expectedFiles:  []string{".armrc", "rules.json"},
			expectedOutput: []string{"Created stub configuration files", "was not installed due to missing source configuration"},
			expectError:    false,
		},
		{
			name: "scenario_3a_has_armrc_install",
			setupFiles: func(tempDir string) error {
				configContent := `[sources]
default = https://registry.armjs.org/`
				return os.WriteFile(filepath.Join(tempDir, ".armrc"), []byte(configContent), 0644)
			},
			command:        []string{armBinary, "install"},
			expectedFiles:  []string{".armrc", "rules.json"},
			expectedOutput: []string{"Created stub rules.json file", "Please add dependencies"},
			expectError:    false,
		},
		{
			name: "scenario_3b_has_armrc_install_ruleset",
			setupFiles: func(tempDir string) error {
				configContent := `[sources]
default = https://registry.armjs.org/`
				return os.WriteFile(filepath.Join(tempDir, ".armrc"), []byte(configContent), 0644)
			},
			command:        []string{armBinary, "install", "test-rules@1.0.0"},
			expectedFiles:  []string{".armrc", "rules.json"},
			expectedOutput: []string{"Created rules.json file with the specified ruleset", "Installing test-rules@1.0.0"},
			expectError:    true, // Will fail due to network call
		},
		{
			name: "scenario_4a_has_rules_install",
			setupFiles: func(tempDir string) error {
				manifestContent := `{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "existing-rules": "1.0.0"
  }
}`
				return os.WriteFile(filepath.Join(tempDir, "rules.json"), []byte(manifestContent), 0644)
			},
			command:        []string{armBinary, "install"},
			expectedFiles:  []string{".armrc", "rules.json"},
			expectedOutput: []string{"No registry sources configured", "Created stub .armrc file"},
			expectError:    true,
		},
		{
			name: "scenario_4b_has_rules_install_ruleset",
			setupFiles: func(tempDir string) error {
				manifestContent := `{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "existing-rules": "1.0.0"
  }
}`
				return os.WriteFile(filepath.Join(tempDir, "rules.json"), []byte(manifestContent), 0644)
			},
			command:        []string{armBinary, "install", "new-rules@2.0.0"},
			expectedFiles:  []string{".armrc", "rules.json"},
			expectedOutput: []string{"No registry sources configured", "Created stub .armrc file"},
			expectError:    true,
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

			// Run the command
			cmd := exec.Command(tt.command[0], tt.command[1:]...)
			cmd.Dir = tempDir
			output, err := cmd.CombinedOutput()

			// Check error expectation
			if tt.expectError {
				assert.Error(t, err, "Expected command to fail but it succeeded")
			} else {
				assert.NoError(t, err, "Command failed unexpectedly: %s", string(output))
			}

			// Verify expected output
			outputStr := string(output)
			for _, expectedOut := range tt.expectedOutput {
				assert.Contains(t, outputStr, expectedOut, "Expected output not found")
			}

			// Verify expected files were created
			for _, file := range tt.expectedFiles {
				assert.True(t, FileExists(file), "Expected file %s to exist", file)
			}

			// Verify stub file contents only for scenarios that create stub files
			if tt.name == "scenario_1a_no_config_install" || tt.name == "scenario_1b_no_config_install_ruleset" || tt.name == "scenario_4a_has_rules_install" || tt.name == "scenario_4b_has_rules_install_ruleset" {
				if contains(tt.expectedFiles, ".armrc") {
					content, err := os.ReadFile(".armrc")
					require.NoError(t, err)
					assert.Contains(t, string(content), "# Example configuration for ARM")
					assert.Contains(t, string(content), "# [sources]")
				}
			}

			if tt.name == "scenario_1a_no_config_install" || tt.name == "scenario_1b_no_config_install_ruleset" || tt.name == "scenario_3a_has_armrc_install" || tt.name == "scenario_3b_has_armrc_install_ruleset" {
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

func TestInstallWorkflowWithValidRegistry(t *testing.T) {
	// Build the ARM binary for testing
	armBinary := buildARMBinary(t)

	// Create isolated test environment
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	// Create filesystem registry with test package
	registryDir := filepath.Join(tempDir, "registry")
	err = os.MkdirAll(registryDir, 0755)
	require.NoError(t, err)

	packageDir := filepath.Join(registryDir, "test-rules", "1.0.0")
	err = os.MkdirAll(packageDir, 0755)
	require.NoError(t, err)

	// Create test rule file in temp directory
	tempRuleDir := filepath.Join(tempDir, "temp-rules")
	err = os.MkdirAll(tempRuleDir, 0755)
	require.NoError(t, err)

	ruleContent := "# Test Rules\n\nUse proper testing practices."
	err = os.WriteFile(filepath.Join(tempRuleDir, "test-rules.md"), []byte(ruleContent), 0644)
	require.NoError(t, err)

	// Create tar.gz package
	tarPath := filepath.Join(packageDir, "test-rules-1.0.0.tar.gz")
	err = createTarGz(tempRuleDir, tarPath)
	require.NoError(t, err)

	t.Run("scenario_2a_both_files_exist_install", func(t *testing.T) {
		// Create config
		configContent := `[sources]
filesystem = ` + registryDir + `

[sources.filesystem]
type = filesystem
path = ` + registryDir
		err = os.WriteFile(".armrc", []byte(configContent), 0644)
		require.NoError(t, err)

		// Create manifest
		manifestContent := `{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "filesystem@test-rules": "1.0.0"
  }
}`
		err = os.WriteFile("rules.json", []byte(manifestContent), 0644)
		require.NoError(t, err)

		// Run install
		cmd := exec.Command(armBinary, "install")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Should succeed with filesystem registry
		assert.NoError(t, err, "Install failed: %s", string(output))

		// Verify installation
		assert.True(t, FileExists(".cursorrules/arm/test-rules/1.0.0/test-rules.md"))
		assert.True(t, FileExists(".amazonq/rules/arm/test-rules/1.0.0/test-rules.md"))
		assert.True(t, FileExists("rules.lock"))
	})

	t.Run("scenario_2b_both_files_exist_install_ruleset", func(t *testing.T) {
		// Clean up previous test
		_ = os.RemoveAll(".cursorrules")
		_ = os.RemoveAll(".amazonq")
		_ = os.Remove("rules.lock")

		// Run install with specific ruleset
		cmd := exec.Command(armBinary, "install", "filesystem@test-rules@1.0.0")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Should succeed with filesystem registry
		assert.NoError(t, err, "Install failed: %s", string(output))

		// Verify installation
		assert.True(t, FileExists(".cursorrules/arm/test-rules/1.0.0/test-rules.md"))
		assert.True(t, FileExists(".amazonq/rules/arm/test-rules/1.0.0/test-rules.md"))
	})
}

// buildARMBinary builds the ARM binary for integration testing
func buildARMBinary(t *testing.T) string {
	// Find the project root
	wd, err := os.Getwd()
	require.NoError(t, err)

	// Navigate up to find the project root (where go.mod is)
	projectRoot := wd
	for {
		if FileExists(filepath.Join(projectRoot, "go.mod")) {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatal("Could not find project root with go.mod")
		}
		projectRoot = parent
	}

	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "arm")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/arm")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build ARM binary: %s", string(output))

	return binaryPath
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

