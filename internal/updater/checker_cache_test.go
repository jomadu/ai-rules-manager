package updater

import (
	"testing"

	"github.com/jomadu/arm/internal/config"
	"github.com/jomadu/arm/internal/registry"
)

func TestCheckerCacheMethodsExist(t *testing.T) {
	// This test verifies that the checker can use cached methods
	// by checking that the required methods exist on the registry manager

	configManager := config.NewManager()
	_ = configManager.Load()
	registryManager := registry.NewManager(configManager)

	checker := NewChecker(registryManager)

	// Test that the checker structure is correct
	if checker.manager == nil {
		t.Error("Registry manager should not be nil")
	}

	// Verify the manager has the cached methods we expect
	// This is a compile-time check that the methods exist
	_ = checker.manager.CachedListVersions
}
