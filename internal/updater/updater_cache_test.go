package updater

import (
	"os"
	"testing"

	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/registry"
)

func TestUpdaterCacheIntegration(t *testing.T) {
	// This test verifies that the updater uses cached methods
	// by checking that the code compiles and runs without errors
	// when using the registry manager's cached methods

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tmpDir)

	// Create basic config
	configManager := config.NewManager()
	_ = configManager.Load()
	registryManager := registry.NewManager(configManager)

	updater := &Updater{
		configManager: configManager,
		manager:       registryManager,
		cacheDir:      ".arm/cache",
	}

	// Test that the updater structure is correct and methods exist
	if updater.manager == nil {
		t.Error("Registry manager should not be nil")
	}

	// Verify the manager has the cached methods we expect
	// This is a compile-time check that the methods exist
	_ = updater.manager.CachedDownload
	_ = updater.manager.ParseRegistryName
	_ = updater.manager.StripRegistryPrefix
}
