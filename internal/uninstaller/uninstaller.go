package uninstaller

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jomadu/arm/pkg/types"
)

type Uninstaller struct{}

func New() *Uninstaller {
	return &Uninstaller{}
}

func (u *Uninstaller) Uninstall(name string) error {
	// Load lock file to get installed version
	lock, err := types.LoadLockFile("rules.lock")
	if err != nil {
		fmt.Printf("No lock file found. Ruleset %s is not installed\n", name)
		return nil
	}

	lockedDep, exists := lock.Dependencies[name]
	if !exists {
		fmt.Printf("Ruleset %s is not installed\n", name)
		return nil
	}

	// Load manifest to get target directories
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Remove from target directories
	var failures []string

	for _, target := range manifest.Targets {
		targetPath := types.GetTargetPath(target, name, lockedDep.Version)
		if err := u.removeRuleset(targetPath); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", target, err))
		}
	}

	// Update manifest and lock files
	if err := u.updateManifest(name); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	if err := u.updateLockFile(name); err != nil {
		return fmt.Errorf("failed to update lock file: %w", err)
	}

	// Report results
	if len(failures) > 0 {
		fmt.Printf("Partially uninstalled %s@%s (failures: %s)\n",
			name, lockedDep.Version, strings.Join(failures, ", "))
	} else {
		fmt.Printf("Successfully uninstalled %s@%s\n", name, lockedDep.Version)
	}

	return nil
}

func (u *Uninstaller) removeRuleset(targetPath string) error {
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return nil // Already removed, success
	}

	// Remove the ruleset directory
	if err := os.RemoveAll(targetPath); err != nil {
		return err
	}

	// Clean up empty parent directories
	return u.cleanupEmptyDirs(filepath.Dir(targetPath))
}

func (u *Uninstaller) cleanupEmptyDirs(dir string) error {
	// Don't remove the base "arm" directory
	if filepath.Base(dir) == "arm" {
		return nil
	}

	// Check if directory is empty
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil // Directory doesn't exist, that's fine
	}

	if len(entries) == 0 {
		if err := os.Remove(dir); err != nil {
			return nil // Ignore cleanup failures
		}
		// Recursively clean parent
		return u.cleanupEmptyDirs(filepath.Dir(dir))
	}

	return nil
}

func (u *Uninstaller) updateManifest(name string) error {
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		return nil // No manifest to update
	}

	delete(manifest.Dependencies, name)
	return manifest.SaveManifest("rules.json")
}

func (u *Uninstaller) updateLockFile(name string) error {
	lock, err := types.LoadLockFile("rules.lock")
	if err != nil {
		return err
	}

	delete(lock.Dependencies, name)
	return lock.SaveLockFile("rules.lock")
}
