package updater

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/jomadu/arm/pkg/types"
)

func TestCheckRulesetUpdate(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Create test manifest
	manifest := &types.RulesManifest{
		Targets: []string{".cursorrules", ".amazonq/rules"},
		Dependencies: map[string]string{
			"typescript-rules": ">= 1.0.0, < 2.0.0",
			"security-rules":   ">= 2.1.0, < 2.2.0",
		},
	}
	err := manifest.SaveManifest("rules.json")
	if err != nil {
		t.Fatalf("Failed to save test manifest: %v", err)
	}

	tests := []struct {
		name           string
		ruleset        InstalledRuleset
		availVersions  []string
		expectedStatus UpdateStatus
		expectedNew    string
	}{
		{
			name: "update available within constraint",
			ruleset: InstalledRuleset{
				Name:    "typescript-rules",
				Version: "1.0.0",
				Source:  "default",
			},
			availVersions:  []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"},
			expectedStatus: UpdateSuccess,
			expectedNew:    "1.2.0",
		},
		{
			name: "no update needed",
			ruleset: InstalledRuleset{
				Name:    "typescript-rules",
				Version: "1.2.0",
				Source:  "default",
			},
			availVersions:  []string{"1.0.0", "1.1.0", "1.2.0"},
			expectedStatus: UpdateNotNeeded,
		},
		{
			name: "patch update within tilde constraint",
			ruleset: InstalledRuleset{
				Name:    "security-rules",
				Version: "2.1.0",
				Source:  "default",
			},
			availVersions:  []string{"2.1.0", "2.1.5", "2.2.0"},
			expectedStatus: UpdateSuccess,
			expectedNew:    "2.1.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the version checking logic
			result := mockCheckRulesetUpdate(tt.ruleset, tt.availVersions)
			
			if result.Status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v (error: %v)", tt.expectedStatus, result.Status, result.Error)
			}
			
			if tt.expectedNew != "" && result.NewVersion != tt.expectedNew {
				t.Errorf("Expected new version %s, got %s", tt.expectedNew, result.NewVersion)
			}
		})
	}
}

// mockCheckRulesetUpdate simulates the version checking logic for testing
func mockCheckRulesetUpdate(ruleset InstalledRuleset, availableVersions []string) UpdateResult {
	// Load manifest to get constraints
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateFailed,
			Error:      err,
		}
	}

	constraint, exists := manifest.Dependencies[ruleset.Name]
	if !exists {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateSkipped,
		}
	}

	// Parse current version
	currentVer, err := version.NewVersion(ruleset.Version)
	if err != nil {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateFailed,
			Error:      err,
		}
	}

	// Parse constraint
	constraints, err := version.NewConstraint(constraint)
	if err != nil {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateFailed,
			Error:      err,
		}
	}

	// Find the latest version that satisfies constraints
	var latestValid *version.Version
	for _, vStr := range availableVersions {
		v, err := version.NewVersion(vStr)
		if err != nil {
			continue
		}

		if constraints.Check(v) && (latestValid == nil || v.GreaterThan(latestValid)) {
			latestValid = v
		}
	}

	if latestValid == nil {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateSkipped,
		}
	}

	// Check if update is needed
	if !latestValid.GreaterThan(currentVer) {
		return UpdateResult{
			Name:       ruleset.Name,
			OldVersion: ruleset.Version,
			Status:     UpdateNotNeeded,
		}
	}

	return UpdateResult{
		Name:       ruleset.Name,
		OldVersion: ruleset.Version,
		NewVersion: latestValid.String(),
		Status:     UpdateSuccess,
	}
}

func TestBackupRestore(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Create test manifest
	manifest := &types.RulesManifest{
		Targets: []string{".cursorrules", ".amazonq/rules"},
		Dependencies: map[string]string{
			"test-rules": "1.0.0",
		},
	}
	err := manifest.SaveManifest("rules.json")
	if err != nil {
		t.Fatalf("Failed to save test manifest: %v", err)
	}

	// Create mock installation
	testFile := filepath.Join(".cursorrules", "arm", "test-rules", "1.0.0", "rule.md")
	err = os.MkdirAll(filepath.Dir(testFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	updater := &Updater{
		cacheDir: ".arm/cache",
	}

	// Test backup creation
	backupPath, err := updater.createBackup("test-rules", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup exists
	backupFile := filepath.Join(backupPath, ".cursorrules", "rule.md")
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Errorf("Backup file not created: %s", backupFile)
		// Debug: list backup contents
		if entries, err := os.ReadDir(backupPath); err == nil {
			t.Logf("Backup directory contents: %v", entries)
		}
	}

	// Remove original file
	err = os.RemoveAll(filepath.Join(".cursorrules", "arm", "test-rules"))
	if err != nil {
		t.Fatalf("Failed to remove original: %v", err)
	}

	// Test restore
	err = updater.restoreBackup("test-rules", backupPath)
	if err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}

	// Verify restore worked - check the restored path structure
	restoredFile := filepath.Join(".cursorrules", "arm", "test-rules", "rule.md")
	if _, err := os.Stat(restoredFile); os.IsNotExist(err) {
		t.Errorf("File not restored: %s", restoredFile)
		// Debug: list what was actually restored
		if entries, err := os.ReadDir(filepath.Join(".cursorrules", "arm")); err == nil {
			t.Logf("Contents of .cursorrules/arm: %v", entries)
		}
	}
}