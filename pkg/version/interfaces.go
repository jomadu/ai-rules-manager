package version

import "github.com/jomadu/ai-rules-manager/pkg/registry"

// VersionResolver handles version constraint logic and decision making
type VersionResolver interface {
	ResolveVersion(constraint string, available []registry.VersionRef) (registry.VersionRef, error)
}

// ContentResolver handles content selection from available files
type ContentResolver interface {
	ResolveContent(selector registry.ContentSelector, available []registry.File) ([]registry.File, error)
}

// SemVerResolver implements VersionResolver using semantic versioning
type SemVerResolver struct{}

func NewSemVerResolver() *SemVerResolver {
	return &SemVerResolver{}
}

func (s *SemVerResolver) ResolveVersion(constraint string, available []registry.VersionRef) (registry.VersionRef, error) {
	// TODO: implement
	return registry.VersionRef{}, nil
}

// GitContentResolver implements ContentResolver for Git repositories
type GitContentResolver struct{}

func NewGitContentResolver() *GitContentResolver {
	return &GitContentResolver{}
}

func (g *GitContentResolver) ResolveContent(selector registry.ContentSelector, available []registry.File) ([]registry.File, error) {
	// TODO: implement
	return nil, nil
}
