package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ResolvePath resolves a path by handling tilde expansion, converting relative paths
// to absolute paths, and validating that the resolved path exists.
func ResolvePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Handle tilde expansion (cross-platform)
	path = expandTilde(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Validate path exists with cross-platform error handling
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path doesn't exist: %s", absPath)
		}
		if isPermissionError(err) {
			return "", fmt.Errorf("permission denied accessing path: %s", absPath)
		}
		return "", fmt.Errorf("failed to check path: %w", err)
	}

	return absPath, nil
}

// expandTilde handles tilde expansion cross-platform
func expandTilde(path string) string {
	// Only expand on Unix-like systems or if explicitly using ~/
	if runtime.GOOS == "windows" {
		// On Windows, only expand if it starts with ~/ or ~\
		if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
			if homeDir, err := os.UserHomeDir(); err == nil {
				return filepath.Join(homeDir, path[2:])
			}
		} else if path == "~" {
			if homeDir, err := os.UserHomeDir(); err == nil {
				return homeDir
			}
		}
	} else {
		// Unix-like systems
		if strings.HasPrefix(path, "~/") {
			if homeDir, err := os.UserHomeDir(); err == nil {
				return filepath.Join(homeDir, path[2:])
			}
		} else if path == "~" {
			if homeDir, err := os.UserHomeDir(); err == nil {
				return homeDir
			}
		}
	}
	return path
}

// isPermissionError checks if an error is a permission-related error cross-platform
func isPermissionError(err error) bool {
	if runtime.GOOS == "windows" {
		// Windows permission errors
		return strings.Contains(err.Error(), "Access is denied") ||
			strings.Contains(err.Error(), "access denied")
	}
	// Unix-like systems
	return strings.Contains(err.Error(), "permission denied")
}

// NormalizePath normalizes a path for cross-platform compatibility
func NormalizePath(path string) string {
	return filepath.Clean(path)
}

// IsAbsolutePath checks if a path is absolute cross-platform
func IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}
