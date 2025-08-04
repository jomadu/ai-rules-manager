package integration

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/installer"
	"github.com/jomadu/arm/internal/registry"
	"github.com/jomadu/arm/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicIntegration(t *testing.T) {
	// Create isolated test environment
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	t.Run("filesystem_registry_workflow", func(t *testing.T) {
		// Create filesystem registry
		registryDir := filepath.Join(tempDir, "registry")
		err := os.MkdirAll(registryDir, 0o755)
		require.NoError(t, err)

		// Create test package structure
		packageDir := filepath.Join(registryDir, "test-rules", "1.0.0")
		err = os.MkdirAll(packageDir, 0o755)
		require.NoError(t, err)

		// Create test rule file in temp directory
		tempRuleDir := filepath.Join(tempDir, "temp-rules")
		err = os.MkdirAll(tempRuleDir, 0o755)
		require.NoError(t, err)

		ruleContent := "# Test Rules\n\nUse proper testing practices."
		err = os.WriteFile(filepath.Join(tempRuleDir, "test-rules.md"), []byte(ruleContent), 0o644)
		require.NoError(t, err)

		// Create tar.gz package
		tarPath := filepath.Join(packageDir, "test-rules-1.0.0.tar.gz")
		err = createTarGz(tempRuleDir, tarPath)
		require.NoError(t, err)

		// Create config
		configContent := `[sources]
filesystem = ` + registryDir + `

[sources.filesystem]
type = filesystem
path = ` + registryDir + `

[performance]
defaultConcurrency = 1`

		err = os.WriteFile(".armrc", []byte(configContent), 0o644)
		require.NoError(t, err)

		// Create manifest
		manifestContent := `{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "filesystem@test-rules": "^1.0.0"
  }
}`
		err = os.WriteFile("rules.json", []byte(manifestContent), 0o644)
		require.NoError(t, err)

		// Test installation
		configManager := config.NewManager()
		err = configManager.Load()
		require.NoError(t, err)

		registryManager := registry.NewManager(configManager)

		installer := installer.NewWithManager(registryManager, "filesystem", "test-rules")
		err = installer.Install("test-rules", "1.0.0")
		require.NoError(t, err)

		// Verify installation
		assert.True(t, FileExists(".cursorrules/arm/test-rules/1.0.0/test-rules.md"))
		assert.True(t, FileExists(".amazonq/rules/arm/test-rules/1.0.0/test-rules.md"))
		assert.True(t, FileExists("rules.lock"))

		// Verify lock file content
		lock, err := types.LoadLockFile("rules.lock")
		require.NoError(t, err)
		assert.Contains(t, lock.Dependencies, "test-rules")
		assert.Equal(t, "1.0.0", lock.Dependencies["test-rules"].Version)
	})

	t.Run("config_parsing", func(t *testing.T) {
		// Test performance config parsing
		configContent := `[sources]
test = http://example.com

[sources.test]
type = http
concurrency = 5

[performance]
defaultConcurrency = 3

[performance.http]
concurrency = 4`

		configPath := filepath.Join(tempDir, "test.armrc")
		err := os.WriteFile(configPath, []byte(configContent), 0o644)
		require.NoError(t, err)

		config, err := config.ParseFile(configPath)
		require.NoError(t, err)

		// Test performance settings
		assert.Equal(t, 3, config.Performance.DefaultConcurrency)
		assert.Equal(t, 4, config.Performance.RegistryTypes["http"].Concurrency)
		assert.Equal(t, 5, config.Sources["test"].Concurrency)
	})

	t.Run("registry_concurrency", func(t *testing.T) {
		// Test concurrency resolution
		testConfig := &config.ARMConfig{
			Sources: map[string]config.Source{
				"test":  {Type: "http", Concurrency: 5},
				"test2": {Type: "http"},
			},
			Performance: config.PerformanceConfig{
				DefaultConcurrency: 3,
				RegistryTypes: map[string]config.TypeConfig{
					"http": {Concurrency: 4},
				},
			},
		}

		configManager := &mockConfigManager{config: testConfig}
		registryManager := registry.NewManager(configManager)

		// Test source-specific override
		assert.Equal(t, 5, registryManager.GetConcurrency("test"))

		// Test type default
		assert.Equal(t, 4, registryManager.GetConcurrency("test2"))

		// Test global default
		assert.Equal(t, 3, registryManager.GetConcurrency("unknown"))
	})
}

func TestErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	t.Run("invalid_config_syntax", func(t *testing.T) {
		invalidConfig := `[sources
missing_bracket = invalid`

		configPath := filepath.Join(tempDir, "invalid.armrc")
		err := os.WriteFile(configPath, []byte(invalidConfig), 0o644)
		require.NoError(t, err)

		_, err = config.ParseFile(configPath)
		assert.Error(t, err)
	})

	t.Run("missing_manifest", func(t *testing.T) {
		_, err := types.LoadManifest("nonexistent.json")
		assert.Error(t, err)
	})

	t.Run("missing_lock_file", func(t *testing.T) {
		_, err := types.LoadLockFile("nonexistent.lock")
		assert.Error(t, err)
	})
}

// mockConfigManager for testing
type mockConfigManager struct {
	config *config.ARMConfig
}

func (m *mockConfigManager) GetConfig() *config.ARMConfig {
	return m.config
}

func (m *mockConfigManager) GetSource(name string) (config.Source, bool) {
	source, exists := m.config.Sources[name]
	return source, exists
}

func (m *mockConfigManager) SetSource(name string, source *config.Source) {
	m.config.Sources[name] = *source
}

func (m *mockConfigManager) Load() error {
	return nil
}

// createTarGz creates a tar.gz archive from a directory
func createTarGz(sourceDir, targetPath string) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	gzWriter := gzip.NewWriter(file)
	defer func() { _ = gzWriter.Close() }()

	tarWriter := tar.NewWriter(gzWriter)
	defer func() { _ = tarWriter.Close() }()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		header := &tar.Header{
			Name: relPath,
			Size: info.Size(),
			Mode: int64(info.Mode()),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(tarWriter, file)
		return err
	})
}
