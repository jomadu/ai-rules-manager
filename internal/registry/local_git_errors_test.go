package registry

import (
	"errors"
	"runtime"
	"strings"
	"testing"
)

func TestLocalGitError_Error(t *testing.T) {
	err := &LocalGitError{
		Type:       "TEST_ERROR",
		Message:    "Test error message",
		Details:    "Additional details",
		Suggestion: "Try this fix",
		Path:       "/test/path",
	}

	errorStr := err.Error()

	// Check that all components are included
	if !strings.Contains(errorStr, "Error [LOCAL-GIT]: Test error message") {
		t.Errorf("Error string should contain main message")
	}
	if !strings.Contains(errorStr, "Details: Additional details") {
		t.Errorf("Error string should contain details")
	}
	if !strings.Contains(errorStr, "Suggestion: Try this fix") {
		t.Errorf("Error string should contain suggestion")
	}
}

func TestLocalGitError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &LocalGitError{
		Type:    "TEST_ERROR",
		Message: "Test error",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Expected unwrapped error to be %v, got %v", cause, unwrapped)
	}
}

func TestNewRepositoryNotFoundError(t *testing.T) {
	path := "/nonexistent/path"
	err := NewRepositoryNotFoundError(path)

	if err.Type != "REPOSITORY_NOT_FOUND" {
		t.Errorf("Expected type REPOSITORY_NOT_FOUND, got %s", err.Type)
	}

	if err.Path != path {
		t.Errorf("Expected path %s, got %s", path, err.Path)
	}

	if !strings.Contains(err.Message, path) {
		t.Errorf("Error message should contain path")
	}

	if !strings.Contains(err.Suggestion, "arm config set") {
		t.Errorf("Suggestion should contain config update command")
	}
}

func TestNewInvalidRepositoryError(t *testing.T) {
	path := "/invalid/repo"
	err := NewInvalidRepositoryError(path)

	if err.Type != "INVALID_REPOSITORY" {
		t.Errorf("Expected type INVALID_REPOSITORY, got %s", err.Type)
	}

	if !strings.Contains(err.Message, "not a valid Git repository") {
		t.Errorf("Message should indicate invalid Git repository")
	}

	if !strings.Contains(err.Suggestion, "git init") {
		t.Errorf("Suggestion should mention git init")
	}
}

func TestNewGitCommandError(t *testing.T) {
	path := "/test/repo"
	command := "rev-parse HEAD"
	cause := errors.New("command failed")

	err := NewGitCommandError(path, command, cause)

	if err.Type != "GIT_COMMAND_FAILED" {
		t.Errorf("Expected type GIT_COMMAND_FAILED, got %s", err.Type)
	}

	if err.Cause != cause {
		t.Errorf("Expected cause to be preserved")
	}

	if !strings.Contains(err.Details, command) {
		t.Errorf("Details should contain command")
	}
}

func TestNewPermissionError(t *testing.T) {
	path := "/restricted/path"
	operation := "read repository"

	err := NewPermissionError(path, operation)

	if err.Type != "PERMISSION_DENIED" {
		t.Errorf("Expected type PERMISSION_DENIED, got %s", err.Type)
	}

	if !strings.Contains(err.Details, operation) {
		t.Errorf("Details should contain operation")
	}

	// Check platform-specific suggestions
	if runtime.GOOS == "windows" {
		if !strings.Contains(err.Suggestion, "Administrator") {
			t.Errorf("Windows suggestion should mention Administrator")
		}
	} else {
		if !strings.Contains(err.Suggestion, "permissions") {
			t.Errorf("Unix suggestion should mention permissions")
		}
	}
}

func TestNewRepositoryMovedError(t *testing.T) {
	path := "/old/path"
	err := NewRepositoryMovedError(path)

	if err.Type != "REPOSITORY_MOVED" {
		t.Errorf("Expected type REPOSITORY_MOVED, got %s", err.Type)
	}

	if !strings.Contains(err.Message, "moved or renamed") {
		t.Errorf("Message should indicate repository was moved")
	}

	if !strings.Contains(err.Suggestion, "arm config set") {
		t.Errorf("Suggestion should contain config update command")
	}
}

func TestNewCorruptedRepositoryError(t *testing.T) {
	path := "/corrupted/repo"
	cause := errors.New("git fsck failed")

	err := NewCorruptedRepositoryError(path, cause)

	if err.Type != "CORRUPTED_REPOSITORY" {
		t.Errorf("Expected type CORRUPTED_REPOSITORY, got %s", err.Type)
	}

	if err.Cause != cause {
		t.Errorf("Expected cause to be preserved")
	}

	if !strings.Contains(err.Suggestion, "git fsck") {
		t.Errorf("Suggestion should mention git fsck")
	}
}

func TestIsLocalGitError(t *testing.T) {
	localErr := &LocalGitError{Type: "TEST_ERROR", Message: "test"}
	regularErr := errors.New("regular error")

	if !IsLocalGitError(localErr) {
		t.Errorf("Should identify LocalGitError")
	}

	if IsLocalGitError(regularErr) {
		t.Errorf("Should not identify regular error as LocalGitError")
	}

	if IsLocalGitError(nil) {
		t.Errorf("Should not identify nil as LocalGitError")
	}
}

func TestGetLocalGitErrorType(t *testing.T) {
	localErr := &LocalGitError{Type: "TEST_ERROR", Message: "test"}
	regularErr := errors.New("regular error")

	errorType := GetLocalGitErrorType(localErr)
	if errorType != "TEST_ERROR" {
		t.Errorf("Expected TEST_ERROR, got %s", errorType)
	}

	errorType = GetLocalGitErrorType(regularErr)
	if errorType != "" {
		t.Errorf("Expected empty string for regular error, got %s", errorType)
	}

	errorType = GetLocalGitErrorType(nil)
	if errorType != "" {
		t.Errorf("Expected empty string for nil error, got %s", errorType)
	}
}

func TestIsPermissionError(t *testing.T) {
	tests := []struct {
		name     string
		error    string
		expected bool
	}{
		{
			name:     "unix_permission_denied",
			error:    "permission denied",
			expected: true,
		},
		{
			name:     "unix_operation_not_permitted",
			error:    "operation not permitted",
			expected: true,
		},
		{
			name:     "regular_error",
			error:    "file not found",
			expected: false,
		},
	}

	// Add Windows-specific tests if running on Windows
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
			{
				name:     "windows_path_not_found",
				error:    "The system cannot find the path",
				expected: true,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.error)
			result := IsPermissionError(err)
			if result != tt.expected {
				t.Errorf("Expected %v for error '%s', got %v", tt.expected, tt.error, result)
			}
		})
	}
}

func TestIsPathNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		error    string
		expected bool
	}{
		{
			name:     "unix_no_such_file",
			error:    "no such file or directory",
			expected: true,
		},
		{
			name:     "regular_error",
			error:    "permission denied",
			expected: false,
		},
	}

	// Add Windows-specific tests if running on Windows
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name     string
			error    string
			expected bool
		}{
			{
				name:     "windows_cannot_find_file",
				error:    "The system cannot find the file",
				expected: true,
			},
			{
				name:     "windows_cannot_find_path",
				error:    "The system cannot find the path",
				expected: true,
			},
			{
				name:     "windows_cannot_find_file_short",
				error:    "cannot find the file",
				expected: true,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.error)
			result := IsPathNotFoundError(err)
			if result != tt.expected {
				t.Errorf("Expected %v for error '%s', got %v", tt.expected, tt.error, result)
			}
		})
	}
}

func TestLocalGitError_ErrorChaining(t *testing.T) {
	// Test error chaining with errors.As
	cause := errors.New("underlying cause")
	localErr := NewGitCommandError("/test/repo", "test command", cause)

	// Should be able to unwrap to find the cause
	if errors.Unwrap(localErr) != cause {
		t.Errorf("Should be able to unwrap LocalGitError to find cause")
	}

	// Should be able to find LocalGitError in chain
	var foundLocalErr *LocalGitError
	if !errors.As(localErr, &foundLocalErr) {
		t.Errorf("Should be able to find LocalGitError in chain")
	}

	if foundLocalErr.Type != "GIT_COMMAND_FAILED" {
		t.Errorf("Found LocalGitError should have correct type")
	}
}
