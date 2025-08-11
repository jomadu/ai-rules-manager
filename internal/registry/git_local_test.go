package registry

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalGitOperations_SymlinkHandling(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git-local-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a mock Git repository
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create .git directory to make it look like a Git repo
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a symlink to the repository
	symlinkPath := filepath.Join(tempDir, "repo-symlink")
	if err := os.Symlink(repoDir, symlinkPath); err != nil {
		t.Skip("Symlinks not supported on this system")
	}

	// Test creating LocalGitOperations with symlink path
	ops, err := NewLocalGitOperations(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create LocalGitOperations with symlink: %v", err)
	}

	// Verify original path is stored
	if ops.GetOriginalPath() != symlinkPath {
		t.Errorf("Expected original path %s, got %s", symlinkPath, ops.GetOriginalPath())
	}

	// Verify resolved path points to actual directory
	expectedResolved, _ := filepath.EvalSymlinks(symlinkPath)
	if ops.GetResolvedPath() != expectedResolved {
		t.Errorf("Expected resolved path %s, got %s", expectedResolved, ops.GetResolvedPath())
	}
}

func TestLocalGitOperations_NonSymlinkPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git-local-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a mock Git repository
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create .git directory
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Test creating LocalGitOperations with regular path
	ops, err := NewLocalGitOperations(repoDir)
	if err != nil {
		t.Fatalf("Failed to create LocalGitOperations: %v", err)
	}

	// Both paths should resolve to the same absolute path for non-symlink
	expectedResolved, _ := filepath.EvalSymlinks(repoDir)
	if ops.GetOriginalPath() != repoDir {
		t.Errorf("Expected original path %s, got %s", repoDir, ops.GetOriginalPath())
	}
	if ops.GetResolvedPath() != expectedResolved {
		t.Errorf("Expected resolved path %s, got %s", expectedResolved, ops.GetResolvedPath())
	}
}
func TestLocalGitOperations_PathValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupPath   func(t *testing.T) string
		expectError bool
		errorType   string
	}{
		{
			name: "nonexistent_path",
			setupPath: func(t *testing.T) string {
				return "/nonexistent/path/to/repo"
			},
			expectError: true,
			errorType:   "REPOSITORY_NOT_FOUND",
		},
		{
			name: "path_exists_but_not_git_repo",
			setupPath: func(t *testing.T) string {
				tempDir, err := os.MkdirTemp("", "not-git-repo-*")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.RemoveAll(tempDir) })
				return tempDir
			},
			expectError: true,
			errorType:   "INVALID_REPOSITORY",
		},
		{
			name: "valid_git_repository",
			setupPath: func(t *testing.T) string {
				tempDir, err := os.MkdirTemp("", "valid-git-repo-*")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

				// Create .git directory
				gitDir := filepath.Join(tempDir, ".git")
				if err := os.MkdirAll(gitDir, 0o755); err != nil {
					t.Fatal(err)
				}
				return tempDir
			},
			expectError: false,
			errorType:   "",
		},
		{
			name: "relative_path_to_valid_repo",
			setupPath: func(t *testing.T) string {
				tempDir, err := os.MkdirTemp("", "relative-repo-*")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

				// Create .git directory
				gitDir := filepath.Join(tempDir, ".git")
				if err := os.MkdirAll(gitDir, 0o755); err != nil {
					t.Fatal(err)
				}

				// Change to parent directory and return relative path
				parentDir := filepath.Dir(tempDir)
				repoName := filepath.Base(tempDir)
				originalWd, _ := os.Getwd()
				t.Cleanup(func() { _ = os.Chdir(originalWd) })
				_ = os.Chdir(parentDir)

				return "./" + repoName
			},
			expectError: false,
			errorType:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupPath(t)

			ops, err := NewLocalGitOperations(path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if !IsLocalGitError(err) {
					t.Errorf("Expected LocalGitError but got %T", err)
					return
				}

				errorType := GetLocalGitErrorType(err)
				if errorType != tt.errorType {
					t.Errorf("Expected error type %s but got %s", tt.errorType, errorType)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}

				if ops == nil {
					t.Error("Expected valid operations but got nil")
				}
			}
		})
	}
}

func TestGitLocalRegistry_PathValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupPath   func(t *testing.T) string
		expectError bool
		errorType   string
	}{
		{
			name: "registry_with_nonexistent_path",
			setupPath: func(t *testing.T) string {
				return "/nonexistent/registry/path"
			},
			expectError: true,
			errorType:   "REPOSITORY_NOT_FOUND",
		},
		{
			name: "registry_with_invalid_git_repo",
			setupPath: func(t *testing.T) string {
				tempDir, err := os.MkdirTemp("", "invalid-git-registry-*")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.RemoveAll(tempDir) })
				return tempDir
			},
			expectError: true,
			errorType:   "INVALID_REPOSITORY",
		},
		{
			name: "registry_with_valid_git_repo",
			setupPath: func(t *testing.T) string {
				tempDir, err := os.MkdirTemp("", "valid-git-registry-*")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

				// Create .git directory
				gitDir := filepath.Join(tempDir, ".git")
				if err := os.MkdirAll(gitDir, 0o755); err != nil {
					t.Fatal(err)
				}
				return tempDir
			},
			expectError: false,
			errorType:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupPath(t)

			config := &RegistryConfig{
				Name: "test-registry",
				URL:  path,
				Type: "git-local",
			}

			registry, err := NewGitLocalRegistry(config, nil)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				// Check if the error is wrapped - unwrap to find LocalGitError
				var localGitErr *LocalGitError
				if !errors.As(err, &localGitErr) {
					t.Errorf("Expected LocalGitError but got %T: %v", err, err)
					return
				}

				if localGitErr.Type != tt.errorType {
					t.Errorf("Expected error type %s but got %s", tt.errorType, localGitErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}

				if registry == nil {
					t.Error("Expected valid registry but got nil")
				}
			}
		})
	}
}
