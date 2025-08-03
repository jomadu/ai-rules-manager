package registry

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jomadu/arm/pkg/types"
)

// FilesystemRegistry implements Registry interface for local filesystem
type FilesystemRegistry struct {
	basePath string
}

// NewFilesystem creates a new filesystem registry client
func NewFilesystem(basePath string) *FilesystemRegistry {
	return &FilesystemRegistry{
		basePath: basePath,
	}
}

// GetRuleset retrieves a specific ruleset version
func (r *FilesystemRegistry) GetRuleset(name, version string) (*types.Ruleset, error) {
	return &types.Ruleset{
		Name:     name,
		Version:  version,
		Source:   r.basePath,
		Files:    []string{},
		Checksum: "",
	}, nil
}

// ListVersions returns all available versions for a ruleset
func (r *FilesystemRegistry) ListVersions(name string) ([]string, error) {
	_, pkg := types.ParseRulesetName(name)
	pkgPath := filepath.Join(r.basePath, pkg)

	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	return versions, nil
}

// Download downloads a ruleset archive
func (r *FilesystemRegistry) Download(name, version string) (io.ReadCloser, error) {
	_, pkg := types.ParseRulesetName(name)
	archivePath := filepath.Join(r.basePath, pkg, version, fmt.Sprintf("%s-%s.tar.gz", pkg, version))

	file, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}

	return file, nil
}

// GetMetadata retrieves metadata for a ruleset
func (r *FilesystemRegistry) GetMetadata(name string) (*Metadata, error) {
	versions, err := r.ListVersions(name)
	if err != nil {
		return nil, err
	}

	versionList := make([]Version, len(versions))
	for i, v := range versions {
		versionList[i] = Version{Version: v}
	}

	return &Metadata{
		Name:        name,
		Description: fmt.Sprintf("Local filesystem registry: %s", r.basePath),
		Versions:    versionList,
		Repository:  r.basePath,
	}, nil
}

// HealthCheck verifies filesystem registry accessibility
func (r *FilesystemRegistry) HealthCheck() error {
	if _, err := os.Stat(r.basePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("registry path does not exist: %s", r.basePath)
		}
		return fmt.Errorf("cannot access registry path: %w", err)
	}

	// Check if path is readable
	entries, err := os.ReadDir(r.basePath)
	if err != nil {
		return fmt.Errorf("cannot read registry directory: %w", err)
	}

	// Basic structure validation
	if len(entries) == 0 {
		return fmt.Errorf("registry directory is empty")
	}

	return nil
}
