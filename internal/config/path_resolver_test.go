package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolvePath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "path-resolver-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name        string
		input       string
		setup       func() string // Returns the expected resolved path
		expectError bool
	}{
		{
			name:        "empty_path",
			input:       "",
			expectError: true,
		},
		{
			name:  "absolute_path",
			input: tempDir,
			setup: func() string {
				return tempDir
			},
		},
		{
			name:  "relative_path",
			input: ".",
			setup: func() string {
				wd, _ := os.Getwd()
				return wd
			},
		},
		{
			name:        "nonexistent_path",
			input:       "/nonexistent/path/that/does/not/exist",
			expectError: true,
		},
	}

	// Add tilde expansion tests based on platform
	homeDir, err := os.UserHomeDir()
	if err == nil {
		tests = append(tests, struct {
			name        string
			input       string
			setup       func() string
			expectError bool
		}{
			name:  "tilde_expansion",
			input: "~",
			setup: func() string {
				return homeDir
			},
		})

		// Test ~/subpath if we can create a test subdirectory
		testSubDir := filepath.Join(homeDir, "test-subdir-for-path-resolver")
		if err := os.MkdirAll(testSubDir, 0o755); err == nil {
			defer func() { _ = os.RemoveAll(testSubDir) }()

			tests = append(tests, struct {
				name        string
				input       string
				setup       func() string
				expectError bool
			}{
				name:  "tilde_with_subpath",
				input: "~/test-subdir-for-path-resolver",
				setup: func() string {
					return testSubDir
				},
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var expected string
			if tt.setup != nil {
				expected = tt.setup()
			}

			result, err := ResolvePath(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if expected != "" && result != expected {
				t.Errorf("Expected %s, got %s", expected, result)
			}

			// Verify result is absolute
			if !filepath.IsAbs(result) {
				t.Errorf("Result should be absolute path: %s", result)
			}
		})
	}
}

func TestExpandTilde(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde_only",
			input:    "~",
			expected: homeDir,
		},
		{
			name:     "no_tilde",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative_path",
			input:    "relative/path",
			expected: "relative/path",
		},
	}

	// Platform-specific tilde expansion tests
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "windows_tilde_slash",
				input:    "~/Documents",
				expected: filepath.Join(homeDir, "Documents"),
			},
			{
				name:     "windows_tilde_backslash",
				input:    "~\\Documents",
				expected: filepath.Join(homeDir, "Documents"),
			},
		}...)
	} else {
		tests = append(tests, struct {
			name     string
			input    string
			expected string
		}{
			name:     "unix_tilde_slash",
			input:    "~/Documents",
			expected: filepath.Join(homeDir, "Documents"),
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandTilde(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsPermissionError(t *testing.T) {
	tests := []struct {
		name     string
		error    string
		expected bool
	}{
		{
			name:     "permission_denied",
			error:    "permission denied",
			expected: true,
		},
		{
			name:     "not_permission_error",
			error:    "file not found",
			expected: false,
		},
	}

	// Add Windows-specific tests
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name     string
			error    string
			expected bool
		}{
			{
				name:     "windows_access_denied",
				error:    "Access is denied",
				expected: true,
			},
			{
				name:     "windows_access_denied_lowercase",
				error:    "access denied",
				expected: true,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &testError{message: tt.error}
			result := isPermissionError(err)
			if result != tt.expected {
				t.Errorf("Expected %v for error '%s', got %v", tt.expected, tt.error, result)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean_path",
			input:    "/path/to/file",
			expected: "/path/to/file",
		},
		{
			name:     "path_with_dots",
			input:    "/path/./to/../file",
			expected: "/path/file",
		},
		{
			name:     "path_with_double_slashes",
			input:    "/path//to///file",
			expected: "/path/to/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsAbsolutePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "unix_absolute",
			path:     "/absolute/path",
			expected: true,
		},
		{
			name:     "relative_path",
			path:     "relative/path",
			expected: false,
		},
		{
			name:     "current_directory",
			path:     ".",
			expected: false,
		},
	}

	// Add Windows-specific tests
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name     string
			path     string
			expected bool
		}{
			{
				name:     "windows_absolute_drive",
				path:     "C:\\absolute\\path",
				expected: true,
			},
			{
				name:     "windows_absolute_unc",
				path:     "\\\\server\\share",
				expected: true,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAbsolutePath(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v for path %s, got %v", tt.expected, tt.path, result)
			}
		})
	}
}

func TestResolvePathPermissionError(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "permission-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a subdirectory
	restrictedDir := filepath.Join(tempDir, "restricted")
	if err := os.MkdirAll(restrictedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Try to make it unreadable (Unix only)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(restrictedDir, 0o000); err != nil {
			t.Skip("Cannot change permissions for test")
		}
		defer func() { _ = os.Chmod(restrictedDir, 0o755) }() // Restore for cleanup

		// Test accessing the restricted directory
		_, err = ResolvePath(restrictedDir)
		if err == nil {
			t.Skip("Expected permission error but got none (may be running as root)")
		}

		if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("Expected permission denied error, got: %v", err)
		}
	}
}

// Helper type for testing error messages
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
