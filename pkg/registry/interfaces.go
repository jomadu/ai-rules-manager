package registry

import "fmt"

// Registry abstracts registry operations and content retrieval
type Registry interface {
	ListVersions() ([]VersionRef, error)
	GetContent(versionRef VersionRef, selector ContentSelector) ([]File, error)
	GetMetadata() RegistryMetadata
}

// VersionRef represents a version reference
type VersionRef struct {
	ID       string            // "1.2.3", "main", "abc123", "latest"
	Type     VersionRefType    // Tag, Branch, Commit, Label
	Metadata map[string]string // Registry-specific data
}

// VersionRefType defines the type of version reference
type VersionRefType int

const (
	Tag VersionRefType = iota
	Branch
	Commit
	Label
)

// ContentSelector defines how to select content from a registry
type ContentSelector interface {
	String() string
	Validate() error
}

// File represents a file from the registry
type File struct {
	Path    string // Relative path within ruleset
	Content []byte // File content
	Size    int64  // File size in bytes
}

// RegistryMetadata contains registry information
type RegistryMetadata struct {
	URL  string
	Type string
}

// GitContentSelector implements ContentSelector for Git repositories
type GitContentSelector struct {
	Patterns []string // ["rules/amazonq/*.md", "rules/cursor/*.mdc"]
	Excludes []string // ["**/*.test.md", "**/README.md"] - optional exclusions
}

func (g GitContentSelector) String() string {
	return fmt.Sprintf("patterns:%v,excludes:%v", g.Patterns, g.Excludes)
}

func (g GitContentSelector) Validate() error {
	if len(g.Patterns) == 0 {
		return fmt.Errorf("at least one pattern is required")
	}
	return nil
}

// GitRegistry implements Registry for Git repositories
type GitRegistry struct {
	url string
}

func NewGitRegistry(url string) *GitRegistry {
	return &GitRegistry{url: url}
}

func (g *GitRegistry) ListVersions() ([]VersionRef, error) {
	// TODO: implement
	return nil, nil
}

func (g *GitRegistry) GetContent(versionRef VersionRef, selector ContentSelector) ([]File, error) {
	// TODO: implement
	return nil, nil
}

func (g *GitRegistry) GetMetadata() RegistryMetadata {
	return RegistryMetadata{URL: g.url, Type: "git"}
}
