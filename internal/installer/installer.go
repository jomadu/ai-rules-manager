package installer

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jomadu/arm/internal/parser"
	"github.com/jomadu/arm/pkg/types"
)

type Installer struct {
	registryURL string
}

func New(registryURL string) *Installer {
	return &Installer{
		registryURL: registryURL,
	}
}

func (i *Installer) Install(name, versionSpec string) error {
	// Parse org and package from name
	org, pkg := types.ParseRulesetName(name)

	// Resolve version
	version := i.resolveVersion(org, pkg, versionSpec)

	// Download ruleset
	tarData, err := i.downloadRuleset(org, pkg, version)
	if err != nil {
		return fmt.Errorf("failed to download ruleset: %w", err)
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", sha256.Sum256(tarData))

	// Extract to target directories
	if err := i.extractRuleset(org, pkg, version, tarData); err != nil {
		return fmt.Errorf("failed to extract ruleset: %w", err)
	}

	// Update manifest and lock files
	if err := i.updateManifest(name, versionSpec); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	if err := i.updateLockFile(name, version, checksum); err != nil {
		return fmt.Errorf("failed to update lock file: %w", err)
	}

	fmt.Printf("Successfully installed %s@%s\n", name, version)
	return nil
}

func (i *Installer) resolveVersion(_, _, versionSpec string) string {
	if versionSpec == "latest" {
		// TODO: Fetch latest version from registry
		return "1.0.0"
	}

	// For now, treat version specs as exact versions
	// TODO: Implement proper semver resolution
	return parser.NormalizeVersion(versionSpec)
}

func (i *Installer) downloadRuleset(org, pkg, version string) ([]byte, error) {
	url := i.buildDownloadURL(org, pkg, version)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d from %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

func (i *Installer) buildDownloadURL(org, pkg, version string) string {
	if org == "" {
		return fmt.Sprintf("%s/%s/%s.tar.gz", i.registryURL, pkg, version)
	}
	return fmt.Sprintf("%s/%s/%s/%s.tar.gz", i.registryURL, org, pkg, version)
}

func (i *Installer) extractRuleset(org, pkg, version string, tarData []byte) error {
	// Load manifest to get target directories
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	for _, target := range manifest.Targets {
		targetDir := filepath.Join(target, "arm")
		if org != "" {
			targetDir = filepath.Join(targetDir, org)
		}
		targetDir = filepath.Join(targetDir, pkg, version)

		if err := i.extractTarGz(tarData, targetDir); err != nil {
			return fmt.Errorf("failed to extract to %s: %w", targetDir, err)
		}
	}

	return nil
}

func (i *Installer) extractTarGz(data []byte, targetDir string) error {
	// Create target directory
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	// Create gzip reader
	gzReader, err := gzip.NewReader(strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gzReader.Close() }()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Sanitize file path
		filename := filepath.Base(header.Name)
		if filename == "" || strings.Contains(filename, "..") {
			continue
		}

		filePath := filepath.Join(targetDir, filename)

		// Create file
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}

		// Copy content
		if _, err := io.Copy(file, tarReader); err != nil {
			_ = file.Close()
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
		_ = file.Close()
	}

	return nil
}

func (i *Installer) updateManifest(name, versionSpec string) error {
	manifestPath := "rules.json"

	// Load existing manifest or create new one
	manifest, err := types.LoadManifest(manifestPath)
	if err != nil {
		// Create new manifest if it doesn't exist
		// Create new manifest with default targets
		manifest = &types.RulesManifest{
			Targets:      types.GetDefaultTargets(),
			Dependencies: make(map[string]string),
		}
	}

	// Add dependency
	manifest.Dependencies[name] = versionSpec

	// Save manifest
	return manifest.SaveManifest(manifestPath)
}

func (i *Installer) updateLockFile(name, version, checksum string) error {
	lockPath := "rules.lock"

	// Load existing lock file or create new one
	lock, err := types.LoadLockFile(lockPath)
	if err != nil {
		// Create new lock file if it doesn't exist
		lock = &types.RulesLock{
			Version:      "1",
			Dependencies: make(map[string]types.LockedDependency),
		}
	}

	// Add locked dependency
	lock.Dependencies[name] = types.LockedDependency{
		Version:  version,
		Source:   i.registryURL,
		Checksum: checksum,
	}

	// Save lock file
	return lock.SaveLockFile(lockPath)
}
