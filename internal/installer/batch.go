package installer

import (
	"fmt"
	"strings"

	"github.com/jomadu/arm/internal/errors"
	"github.com/jomadu/arm/pkg/types"
)

// BatchResult represents the result of a batch installation
type BatchResult struct {
	Successful []string
	Failed     map[string]error
}

// BatchInstaller handles installation of multiple packages
type BatchInstaller struct {
	manager RegistryManager
}

// NewBatchInstaller creates a new batch installer
func NewBatchInstaller(manager RegistryManager) *BatchInstaller {
	return &BatchInstaller{
		manager: manager,
	}
}

// InstallFromManifest installs all dependencies from a manifest file
func (b *BatchInstaller) InstallFromManifest(manifestPath string) (*BatchResult, error) {
	manifest, err := types.LoadManifest(manifestPath)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrManifestInvalid, "Failed to load manifest").
			WithContext("file", manifestPath).
			WithSuggestion("Check manifest file exists and is valid JSON").
			WithSuggestion("Create manifest with: arm init")
	}

	if err := manifest.Validate(); err != nil {
		return nil, errors.Wrap(err, errors.ErrManifestInvalid, "Invalid manifest").
			WithContext("file", manifestPath).
			WithSuggestion("Fix manifest validation errors").
			WithSuggestion("Check targets and dependencies format")
	}

	return b.InstallPackages(manifest.Dependencies)
}

// InstallPackages installs multiple packages and collects all errors
func (b *BatchInstaller) InstallPackages(packages map[string]string) (*BatchResult, error) {
	result := &BatchResult{
		Successful: []string{},
		Failed:     make(map[string]error),
	}

	for name, versionSpec := range packages {
		err := b.installSinglePackage(name, versionSpec)
		if err != nil {
			result.Failed[name] = err
		} else {
			result.Successful = append(result.Successful, name)
		}
	}

	// If all packages failed, return an error
	if len(result.Successful) == 0 && len(result.Failed) > 0 {
		return result, b.createBatchError(result.Failed)
	}

	return result, nil
}

// installSinglePackage installs a single package
func (b *BatchInstaller) installSinglePackage(name, versionSpec string) error {
	// Parse registry name from package name
	registryName := "default"
	cleanName := name

	if strings.Contains(name, "@") && b.manager != nil {
		// This is a scoped package, extract registry name
		parts := strings.Split(name, "@")
		if len(parts) >= 2 {
			registryName = parts[0]
			cleanName = strings.Join(parts[1:], "@")
		}
	}

	installer := NewWithManager(b.manager, registryName, cleanName)
	return installer.Install(cleanName, versionSpec)
}

// createBatchError creates a structured error for batch failures
func (b *BatchInstaller) createBatchError(failures map[string]error) error {
	_ = failures // We use the count, not individual messages

	return errors.New(errors.ErrPackageNotFound, fmt.Sprintf("Failed to install %d package(s)", len(failures))).
		WithContext("failed_count", fmt.Sprintf("%d", len(failures))).
		WithSuggestion("Check individual package errors above").
		WithSuggestion("Verify package names and versions").
		WithSuggestion("Check registry connectivity")
}

// PrintResults prints the batch installation results
func (b *BatchInstaller) PrintResults(result *BatchResult) {
	if len(result.Successful) > 0 {
		fmt.Printf("✓ Successfully installed %d package(s):\n", len(result.Successful))
		for _, pkg := range result.Successful {
			fmt.Printf("  - %s\n", pkg)
		}
	}

	if len(result.Failed) > 0 {
		fmt.Printf("\n✗ Failed to install %d package(s):\n", len(result.Failed))
		for pkg, err := range result.Failed {
			fmt.Printf("  - %s: %v\n", pkg, err)
		}
	}

	if len(result.Successful) > 0 && len(result.Failed) > 0 {
		fmt.Printf("\nPartial success: %d/%d packages installed\n",
			len(result.Successful), len(result.Successful)+len(result.Failed))
	}
}
