package registry

import (
	"context"
	"fmt"
)

// GitError represents context-aware Git operation errors
type GitError struct {
	Operation string
	Repo      string
	Version   string
	Cause     error
}

func (e *GitError) Error() string {
	if e.Version != "" {
		return fmt.Sprintf("git %s failed for %s@%s: %v", e.Operation, e.Repo, e.Version, e.Cause)
	}
	return fmt.Sprintf("git %s failed for %s: %v", e.Operation, e.Repo, e.Cause)
}

func (e *GitError) Unwrap() error {
	return e.Cause
}

// VersionResolver handles Git version operations
type VersionResolver interface {
	ResolveVersion(ctx context.Context, constraint string) (string, error)
	ListVersions(ctx context.Context) ([]string, error)
}

// FileProvider handles Git file operations
type FileProvider interface {
	GetFiles(ctx context.Context, version string, patterns []string) (map[string][]byte, error)
}

// GitOperations combines version resolution and file operations
type GitOperations interface {
	VersionResolver
	FileProvider
}
