package registry

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// LocalGitError represents errors specific to local Git registry operations
type LocalGitError struct {
	Type       string // Error type for categorization
	Message    string // Primary error message
	Details    string // Additional context
	Suggestion string // Recommended corrective action
	Path       string // Repository path for context
	Cause      error  // Underlying error if any
}

func (e *LocalGitError) Error() string {
	msg := fmt.Sprintf("Error [LOCAL-GIT]: %s", e.Message)
	if e.Details != "" {
		msg += fmt.Sprintf("\nDetails: %s", e.Details)
	}
	if e.Suggestion != "" {
		msg += fmt.Sprintf("\nSuggestion: %s", e.Suggestion)
	}
	return msg
}

func (e *LocalGitError) Unwrap() error {
	return e.Cause
}

// Error constructors for common local Git scenarios

func NewRepositoryNotFoundError(path string) *LocalGitError {
	return &LocalGitError{
		Type:       "REPOSITORY_NOT_FOUND",
		Message:    fmt.Sprintf("Repository path does not exist: %s", path),
		Details:    fmt.Sprintf("The specified path '%s' could not be found on the filesystem", path),
		Suggestion: fmt.Sprintf("Verify the path exists or update registry configuration with:\n  arm config set registries.%s /correct/path/to/repository", filepath.Base(path)),
		Path:       path,
	}
}

func NewInvalidRepositoryError(path string) *LocalGitError {
	return &LocalGitError{
		Type:       "INVALID_REPOSITORY",
		Message:    fmt.Sprintf("Path is not a valid Git repository: %s", path),
		Details:    fmt.Sprintf("The path '%s' exists but does not contain a valid Git repository (.git directory not found)", path),
		Suggestion: "Initialize a Git repository with 'git init' or update registry configuration to point to a valid Git repository",
		Path:       path,
	}
}

func NewGitCommandError(path, command string, cause error) *LocalGitError {
	return &LocalGitError{
		Type:       "GIT_COMMAND_FAILED",
		Message:    fmt.Sprintf("Git command failed in repository: %s", path),
		Details:    fmt.Sprintf("Command 'git %s' failed: %v", command, cause),
		Suggestion: "Check if Git is installed and the repository is not corrupted. Try running the command manually to diagnose the issue",
		Path:       path,
		Cause:      cause,
	}
}

func NewPermissionError(path, operation string) *LocalGitError {
	suggestion := "Check file system permissions or run with appropriate privileges"
	if runtime.GOOS == "windows" {
		suggestion = "Check file system permissions or run as Administrator if needed"
	}
	return &LocalGitError{
		Type:       "PERMISSION_DENIED",
		Message:    fmt.Sprintf("Permission denied accessing repository: %s", path),
		Details:    fmt.Sprintf("Insufficient permissions to %s in repository at '%s'", operation, path),
		Suggestion: suggestion,
		Path:       path,
	}
}

func NewRepositoryMovedError(path string) *LocalGitError {
	return &LocalGitError{
		Type:       "REPOSITORY_MOVED",
		Message:    fmt.Sprintf("Repository appears to have been moved or renamed: %s", path),
		Details:    fmt.Sprintf("The repository at '%s' was previously accessible but is no longer found", path),
		Suggestion: fmt.Sprintf("Update registry configuration with the new path:\n  arm config set registries.%s /new/path/to/repository", filepath.Base(path)),
		Path:       path,
	}
}

func NewCorruptedRepositoryError(path string, cause error) *LocalGitError {
	return &LocalGitError{
		Type:       "CORRUPTED_REPOSITORY",
		Message:    fmt.Sprintf("Repository appears to be corrupted: %s", path),
		Details:    fmt.Sprintf("Git operations failed due to repository corruption: %v", cause),
		Suggestion: "Try running 'git fsck' to check repository integrity, or re-clone the repository if it's corrupted beyond repair",
		Path:       path,
		Cause:      cause,
	}
}

// IsLocalGitError checks if an error is a LocalGitError
func IsLocalGitError(err error) bool {
	_, ok := err.(*LocalGitError)
	return ok
}

// GetLocalGitErrorType returns the error type if it's a LocalGitError
func GetLocalGitErrorType(err error) string {
	if localErr, ok := err.(*LocalGitError); ok {
		return localErr.Type
	}
	return ""
}

// IsPermissionError checks if an error is permission-related cross-platform
func IsPermissionError(err error) bool {
	errStr := strings.ToLower(err.Error())
	if runtime.GOOS == "windows" {
		return strings.Contains(errStr, "access is denied") ||
			strings.Contains(errStr, "access denied") ||
			strings.Contains(errStr, "the system cannot find the path")
	}
	return strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "operation not permitted")
}

// IsPathNotFoundError checks if an error indicates a missing path cross-platform
func IsPathNotFoundError(err error) bool {
	errStr := strings.ToLower(err.Error())
	if runtime.GOOS == "windows" {
		return strings.Contains(errStr, "the system cannot find the file") ||
			strings.Contains(errStr, "the system cannot find the path") ||
			strings.Contains(errStr, "cannot find the file")
	}
	return strings.Contains(errStr, "no such file or directory")
}
