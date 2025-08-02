package installer

import (
	"os"
	"testing"

	"github.com/jomadu/arm/pkg/types"
)

func TestInstallerUsesManifestTargets(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Create a manifest with custom targets
	customTargets := []string{".custom-rules", ".another-target/rules"}
	manifest := &types.RulesManifest{
		Targets:      customTargets,
		Dependencies: make(map[string]string),
	}

	// Save the manifest
	if err := manifest.SaveManifest("rules.json"); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Load the manifest to verify it loads correctly
	loadedManifest, err := types.LoadManifest("rules.json")
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	// Verify the loaded manifest has the custom targets
	if len(loadedManifest.Targets) != len(customTargets) {
		t.Errorf("Expected %d targets, got %d", len(customTargets), len(loadedManifest.Targets))
	}

	for i, target := range customTargets {
		if loadedManifest.Targets[i] != target {
			t.Errorf("Expected target %s, got %s", target, loadedManifest.Targets[i])
		}
	}

	// Verify that GetDefaultTargets returns the expected defaults
	defaultTargets := types.GetDefaultTargets()
	expectedDefaults := []string{".cursorrules", ".amazonq/rules"}
	if len(defaultTargets) != len(expectedDefaults) {
		t.Errorf("Expected %d default targets, got %d", len(expectedDefaults), len(defaultTargets))
	}

	for i, expected := range expectedDefaults {
		if defaultTargets[i] != expected {
			t.Errorf("Expected default target %s, got %s", expected, defaultTargets[i])
		}
	}
}
