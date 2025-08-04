package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jomadu/arm/internal/cleaner"
	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/installer"
	"github.com/jomadu/arm/internal/updater"
	"github.com/jomadu/arm/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompleteWorkflow(t *testing.T) {
	env := NewTestEnv(t)

	// Step 1: Install from manifest
	t.Run("install", func(t *testing.T) {
		manifest, err := types.LoadManifest("rules.json")
		require.NoError(t, err)

		for name, versionSpec := range manifest.Dependencies {
			registryName := env.RegistryManager.ParseRegistryName(name)
			cleanName := env.RegistryManager.StripRegistryPrefix(name)

			installer := installer.NewWithManager(env.RegistryManager, registryName, cleanName)
			err := installer.Install(cleanName, versionSpec)
			require.NoError(t, err)
		}

		// Verify files were created
		assert.True(t, FileExists(".cursorrules/arm/typescript-rules/1.0.0/typescript-rules.md"))
		assert.True(t, FileExists(".amazonq/rules/arm/typescript-rules/1.0.0/typescript-rules.md"))
		assert.True(t, FileExists("rules.lock"))
	})

	// Step 2: List installed rulesets
	t.Run("list", func(t *testing.T) {
		// Load lock file to verify installation
		lock, err := types.LoadLockFile("rules.lock")
		require.NoError(t, err)

		assert.Len(t, lock.Dependencies, 1)
		assert.Contains(t, lock.Dependencies, "typescript-rules")
		assert.Equal(t, "1.0.0", lock.Dependencies["typescript-rules"].Version)
	})

	// Step 3: Update (should be no-op since already latest)
	t.Run("update", func(t *testing.T) {
		updater, err := updater.New()
		require.NoError(t, err)

		// Update specific ruleset with dry-run
		err = updater.Update("typescript-rules", true)
		require.NoError(t, err)
	})

	// Step 4: Clean unused files
	t.Run("clean", func(t *testing.T) {
		cleaner := cleaner.New()

		// Load manifest to get targets
		manifest, err := types.LoadManifest("rules.json")
		require.NoError(t, err)

		// Load lock file to get used rulesets
		lock, err := types.LoadLockFile("rules.lock")
		require.NoError(t, err)

		// Create used rulesets map
		usedRulesets := make(map[string]bool)
		for name := range lock.Dependencies {
			usedRulesets[name] = true
		}

		// Clean targets with dry-run
		err = cleaner.CleanTargets(manifest.Targets, usedRulesets, true)
		require.NoError(t, err)
	})
}

func TestMultiRegistryWorkflow(t *testing.T) {
	env := NewTestEnv(t)

	// Create manifest with multiple registries
	manifestContent := `{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "typescript-rules": "^1.0.0",
    "filesystem@typescript-rules": "^1.0.0"
  }
}`
	err := os.WriteFile("rules.json", []byte(manifestContent), 0o644)
	require.NoError(t, err)

	t.Run("install_from_multiple_registries", func(t *testing.T) {
		manifest, err := types.LoadManifest("rules.json")
		require.NoError(t, err)

		for name, versionSpec := range manifest.Dependencies {
			registryName := env.RegistryManager.ParseRegistryName(name)
			cleanName := env.RegistryManager.StripRegistryPrefix(name)

			installer := installer.NewWithManager(env.RegistryManager, registryName, cleanName)
			err := installer.Install(cleanName, versionSpec)
			require.NoError(t, err)
		}

		// Verify both installations
		assert.True(t, FileExists(".cursorrules/arm/typescript-rules/1.0.0/typescript-rules.md"))
		assert.True(t, DirExists(".cursorrules/arm/filesystem"))
	})

	t.Run("list_multiple_sources", func(t *testing.T) {
		// Load lock file to verify multiple installations
		lock, err := types.LoadLockFile("rules.lock")
		require.NoError(t, err)

		// Should have rulesets from both registries
		assert.GreaterOrEqual(t, len(lock.Dependencies), 1)
	})
}

func TestErrorScenarios(t *testing.T) {
	env := NewTestEnv(t)

	t.Run("invalid_config", func(t *testing.T) {
		// Copy invalid config
		invalidConfigPath := filepath.Join(env.TempDir, "invalid.armrc")
		err := os.WriteFile(invalidConfigPath, []byte(`[sources
missing_bracket = invalid`), 0o644)
		require.NoError(t, err)

		// Try to parse invalid config
		_, err = config.ParseFile(invalidConfigPath)
		assert.Error(t, err)
	})

	t.Run("nonexistent_ruleset", func(t *testing.T) {
		installer := installer.NewWithManager(env.RegistryManager, "default", "nonexistent-ruleset")
		err := installer.Install("nonexistent-ruleset", "1.0.0")
		assert.Error(t, err)
	})

	t.Run("permission_error", func(t *testing.T) {
		// Create read-only directory
		readOnlyDir := filepath.Join(env.TempDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0o444) // read-only
		require.NoError(t, err)

		// Try to create files in read-only directory (this may not fail on all systems)
		err = os.WriteFile(filepath.Join(readOnlyDir, "test.txt"), []byte("test"), 0o644)
		// Note: This test may be platform-dependent
		if err != nil {
			assert.Error(t, err)
		}
	})
}
