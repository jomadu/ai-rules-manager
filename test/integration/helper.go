package integration

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/registry"
	"github.com/stretchr/testify/require"
)

// TestEnv represents an isolated test environment
type TestEnv struct {
	TempDir         string
	ConfigPath      string
	ManifestPath    string
	HTTPServer      *httptest.Server
	ConfigManager   *config.Manager
	RegistryManager *registry.Manager
}

// NewTestEnv creates a new isolated test environment
func NewTestEnv(t *testing.T) *TestEnv {
	tempDir := t.TempDir()

	// Create HTTP test server
	server := httptest.NewServer(NewTestServer())

	// Create config file
	configPath := filepath.Join(tempDir, ".armrc")
	configContent := fmt.Sprintf(`[sources]
default = %s
filesystem = %s

[sources.filesystem]
type = filesystem
path = %s

[performance]
defaultConcurrency = 2
`, server.URL, filepath.Join(tempDir, "fs-registry"), filepath.Join(tempDir, "fs-registry"))

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Create filesystem registry directory
	fsRegistryDir := filepath.Join(tempDir, "fs-registry")
	err = os.MkdirAll(fsRegistryDir, 0o755)
	require.NoError(t, err)

	// Copy test packages to filesystem registry
	setupFilesystemRegistry(t, fsRegistryDir)

	// Create manifest
	manifestPath := filepath.Join(tempDir, "rules.json")
	manifestContent := `{
  "targets": [".cursorrules", ".amazonq/rules"],
  "dependencies": {
    "typescript-rules": "^1.0.0"
  }
}`
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0o644)
	require.NoError(t, err)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Cleanup function
	t.Cleanup(func() {
		server.Close()
		_ = os.Chdir(originalDir)
	})

	// Create config manager
	configManager := config.NewManager()
	err = configManager.Load()
	require.NoError(t, err)

	registryManager := registry.NewManager(configManager)

	return &TestEnv{
		TempDir:         tempDir,
		ConfigPath:      configPath,
		ManifestPath:    manifestPath,
		HTTPServer:      server,
		ConfigManager:   configManager,
		RegistryManager: registryManager,
	}
}

// setupFilesystemRegistry copies test packages to filesystem registry
func setupFilesystemRegistry(t *testing.T, fsDir string) {
	// Create typescript-rules package
	packageDir := filepath.Join(fsDir, "typescript-rules")
	err := os.MkdirAll(packageDir, 0o755)
	require.NoError(t, err)

	// Create version directory and files
	versionDir := filepath.Join(packageDir, "1.0.0")
	err = os.MkdirAll(versionDir, 0o755)
	require.NoError(t, err)

	// Create test rule file
	ruleContent := "# TypeScript Rules\n\nUse strict typing and proper interfaces."
	err = os.WriteFile(filepath.Join(versionDir, "typescript-rules.md"), []byte(ruleContent), 0o644)
	require.NoError(t, err)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
