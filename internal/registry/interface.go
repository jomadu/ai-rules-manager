package registry

import (
	"io"

	"github.com/jomadu/arm/pkg/types"
)

// Registry defines the interface for ruleset registries
type Registry interface {
	// GetRuleset retrieves a specific ruleset version
	GetRuleset(name, version string) (*types.Ruleset, error)

	// ListVersions returns all available versions for a ruleset
	ListVersions(name string) ([]string, error)

	// Download downloads a ruleset archive
	Download(name, version string) (io.ReadCloser, error)

	// GetMetadata retrieves metadata for a ruleset
	GetMetadata(name string) (*Metadata, error)
}

// Metadata represents ruleset metadata from registry
type Metadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Versions    []Version `json:"versions"`
	Repository  string    `json:"repository"`
}

// Version represents a specific version with metadata
type Version struct {
	Version   string `json:"version"`
	Published string `json:"published"`
	Checksum  string `json:"checksum"`
}
