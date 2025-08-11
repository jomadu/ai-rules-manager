package registry

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Pattern matching utilities

// MatchesAnyPattern checks if a file path matches any of the given patterns
func MatchesAnyPattern(filePath string, patterns []string) bool {
	if len(patterns) == 0 {
		return true // Empty patterns match everything
	}

	for _, pattern := range patterns {
		// Direct filepath.Match for exact patterns
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return true
		}

		// Check if pattern matches just the filename
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}

		// Handle glob patterns with ** and *
		if MatchesGlobPattern(filePath, pattern) {
			return true
		}
	}
	return false
}

// MatchesGlobPattern handles glob patterns including **
func MatchesGlobPattern(filePath, pattern string) bool {
	// Handle ** patterns
	if strings.Contains(pattern, "**") {
		// Convert glob pattern to regex
		regexPattern := regexp.QuoteMeta(pattern)
		// Replace ** with .* (matches any characters including /)
		regexPattern = strings.ReplaceAll(regexPattern, `\*\*`, ".*")
		// Replace single * with [^/]* (matches any characters except /)
		regexPattern = strings.ReplaceAll(regexPattern, `\*`, "[^/]*")
		regexPattern = "^" + regexPattern + "$"

		if matched, _ := regexp.MatchString(regexPattern, filePath); matched {
			return true
		}
	}

	// Handle simple * patterns
	if strings.Contains(pattern, "*") && !strings.Contains(pattern, "**") {
		// For patterns like *.md, check if it matches the full path or just the filename
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
		// Also check the full path
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return true
		}
	}

	return false
}

// File operations utilities

// CopyFile copies a single file
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CopyMatchingFiles copies the matching files to destination directory preserving structure
func CopyMatchingFiles(repoDir string, matchingFiles []string, destDir string) error {
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return err
	}

	for _, relPath := range matchingFiles {
		// Validate path to prevent traversal
		if strings.Contains(relPath, "..") {
			continue // Skip potentially malicious paths
		}

		srcPath := filepath.Join(repoDir, relPath)
		// Preserve directory structure in destination
		destPath := filepath.Join(destDir, relPath)

		// Create destination directory
		destFileDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destFileDir, 0o700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destFileDir, err)
		}

		// Copy file
		if err := CopyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", relPath, err)
		}
	}

	return nil
}

// FindMatchingFiles finds files in repository that match the given patterns
func FindMatchingFiles(repoDir string, patterns []string) ([]string, error) {
	var matchingFiles []string

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory and other hidden directories
		if info.IsDir() && (info.Name() == ".git" || strings.HasPrefix(info.Name(), ".")) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path from repo root
		relPath, err := filepath.Rel(repoDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Validate path to prevent traversal
		if strings.Contains(relPath, "..") {
			return nil // Skip potentially malicious paths
		}

		// Check if file matches any pattern
		if MatchesAnyPattern(relPath, patterns) {
			matchingFiles = append(matchingFiles, relPath)
		}

		return nil
	})

	return matchingFiles, err
}

// Semver utilities

// IsSemverPattern checks if version is a semver pattern
func IsSemverPattern(version string) bool {
	return strings.HasPrefix(version, "^") || strings.HasPrefix(version, "~") ||
		strings.HasPrefix(version, ">=") || strings.HasPrefix(version, "<=") ||
		strings.HasPrefix(version, ">") || strings.HasPrefix(version, "<")
}

// ResolveSemverPattern resolves a semver pattern to the highest matching version
func ResolveSemverPattern(versionSpec string, availableVersions []string) (string, error) {
	constraint, err := semver.NewConstraint(versionSpec)
	if err != nil {
		return "", fmt.Errorf("invalid semver constraint: %w", err)
	}

	// Parse and filter valid semantic versions
	var candidates []*semver.Version
	for _, v := range availableVersions {
		if v == "latest" {
			continue
		}
		if ver, err := semver.NewVersion(v); err == nil {
			candidates = append(candidates, ver)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no valid semantic versions found")
	}

	// Sort versions (highest first)
	sort.Sort(sort.Reverse(semver.Collection(candidates)))

	// Find the highest version that satisfies the constraint
	for _, candidate := range candidates {
		if constraint.Check(candidate) {
			return candidate.String(), nil
		}
	}

	return "", fmt.Errorf("no versions satisfy constraint: %s", versionSpec)
}

// ResolveLatestVersion resolves "latest" to the highest semantic version
func ResolveLatestVersion(availableVersions []string) (string, error) {
	// Parse and filter valid semantic versions
	var candidates []*semver.Version
	for _, v := range availableVersions {
		if v == "latest" {
			continue
		}
		if ver, err := semver.NewVersion(v); err == nil {
			candidates = append(candidates, ver)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no valid semantic versions found")
	}

	// Sort versions (highest first) and return the latest
	sort.Sort(sort.Reverse(semver.Collection(candidates)))
	return candidates[0].String(), nil
}

// IsVersionNumber checks if a string looks like a version number
func IsVersionNumber(version string) bool {
	// Match patterns like "1.0.0", "v1.0.0", "2.1.3", etc.
	matched, _ := regexp.MatchString(`^v?\d+\.\d+\.\d+`, version)
	return matched
}

// Path validation utilities

// IsHexString checks if a string contains only hexadecimal characters
func IsHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// ValidatePath validates a path to prevent directory traversal attacks
func ValidatePath(path string) bool {
	return !strings.Contains(path, "..")
}
