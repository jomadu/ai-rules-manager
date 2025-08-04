package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jomadu/arm/internal/cleaner"
	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/installer"
	"github.com/jomadu/arm/internal/registry"
	"github.com/jomadu/arm/internal/updater"
	"github.com/jomadu/arm/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemWorkflow(t *testing.T) {
	// Create isolated test environment
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	// Create filesystem registry
	registryDir := filepath.Join(tempDir, "registry")
	err = os.MkdirAll(registryDir, 0o755)
	require.NoError(t, err)

	// Create test package
	packageDir := filepath.Join(registryDir, "workflow-rules", "1.0.0")
	err = os.MkdirAll(packageDir, 0o755)
	require.NoError(t, err)

	// Create test rule file and tar.gz
	tempRuleDir := filepath.Join(tempDir, "temp-rules")
	err = os.MkdirAll(tempRuleDir, 0o755)
	require.NoError(t, err)

	ruleContent := "# Workflow Rules\n\nUse proper workflow practices."
	err = os.WriteFile(filepath.Join(tempRuleDir, "workflow-rules.md"), []byte(ruleContent), 0o644)
	require.NoError(t, err)

	tarPath := filepath.Join(packageDir, "workflow-rules-1.0.0.tar.gz")
	err = createTarGz(tempRuleDir, tarPath)
	require.NoError(t, err)

	// Create config
	configContent := `[sources]
filesystem = ` + registryDir + `

[sources.filesystem]
type = filesystem
path = ` + registryDir + `

[performance]
defaultConcurrency = 2`

	err = os.WriteFile(".armrc", []byte(configContent), 0o644)
	require.NoError(t, err)

	// Create manifest
	manifestContent := `{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "filesystem@workflow-rules": "^1.0.0"
  }
}`
	err = os.WriteFile("rules.json", []byte(manifestContent), 0o644)
	require.NoError(t, err)

	// Step 1: Install
	t.Run("install", func(t *testing.T) {
		configManager := config.NewManager()
		err := configManager.Load()
		require.NoError(t, err)

		registryManager := registry.NewManager(configManager)
		installer := installer.NewWithManager(registryManager, "filesystem", "workflow-rules")
		err = installer.Install("workflow-rules", "1.0.0")
		require.NoError(t, err)

		// Verify installation
		assert.True(t, FileExists(".cursorrules/arm/workflow-rules/1.0.0/workflow-rules.md"))
		assert.True(t, FileExists(".amazonq/rules/arm/workflow-rules/1.0.0/workflow-rules.md"))
		assert.True(t, FileExists("rules.lock"))
	})

	// Step 2: List (verify lock file)
	t.Run("list", func(t *testing.T) {
		lock, err := types.LoadLockFile("rules.lock")
		require.NoError(t, err)

		assert.Len(t, lock.Dependencies, 1)
		assert.Contains(t, lock.Dependencies, "workflow-rules")
		assert.Equal(t, "1.0.0", lock.Dependencies["workflow-rules"].Version)
	})

	// Step 3: Update (dry-run)
	t.Run("update", func(t *testing.T) {
		updater, err := updater.New()
		require.NoError(t, err)

		err = updater.Update("workflow-rules", true)
		require.NoError(t, err)
	})

	// Step 4: Clean (dry-run)
	t.Run("clean", func(t *testing.T) {
		cleaner := cleaner.New()

		manifest, err := types.LoadManifest("rules.json")
		require.NoError(t, err)

		lock, err := types.LoadLockFile("rules.lock")
		require.NoError(t, err)

		usedRulesets := make(map[string]bool)
		for name := range lock.Dependencies {
			usedRulesets[name] = true
		}

		err = cleaner.CleanTargets(manifest.Targets, usedRulesets, true)
		require.NoError(t, err)
	})
}

func TestErrorScenarios(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	t.Run("invalid_config", func(t *testing.T) {
		invalidConfigPath := filepath.Join(tempDir, "invalid.armrc")
		err := os.WriteFile(invalidConfigPath, []byte(`[sources
missing_bracket = invalid`), 0o644)
		require.NoError(t, err)

		_, err = config.ParseFile(invalidConfigPath)
		assert.Error(t, err)
	})

	t.Run("nonexistent_ruleset", func(t *testing.T) {
		// Create minimal config
		configContent := `[sources]
filesystem = ` + tempDir + `

[sources.filesystem]
type = filesystem
path = ` + tempDir

		err := os.WriteFile(".armrc", []byte(configContent), 0o644)
		require.NoError(t, err)

		configManager := config.NewManager()
		err = configManager.Load()
		require.NoError(t, err)

		registryManager := registry.NewManager(configManager)
		installer := installer.NewWithManager(registryManager, "filesystem", "nonexistent-ruleset")
		err = installer.Install("nonexistent-ruleset", "1.0.0")
		assert.Error(t, err)
	})
}
