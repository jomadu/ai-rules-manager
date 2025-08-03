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

	// HealthCheck verifies registry connectivity and authentication
	HealthCheck() error
}

// HealthStatus represents registry health information
type HealthStatus struct {
	Healthy      bool   `json:"healthy"`
	ResponseTime string `json:"responseTime"`
	Error        string `json:"error,omitempty"`
	Version      string `json:"version,omitempty"`
}

// Metadata represents ruleset metadata from registry
type Metadata struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Versions     []Version         `json:"versions"`
	Repository   string            `json:"repository"`
	Homepage     string            `json:"homepage,omitempty"`
	License      string            `json:"license,omitempty"`
	Keywords     []string          `json:"keywords,omitempty"`
	Maintainers  []string          `json:"maintainers,omitempty"`
	Downloads    int64             `json:"downloads,omitempty"`
	LastModified string            `json:"lastModified,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

// Version represents a specific version with metadata
type Version struct {
	Version     string `json:"version"`
	Published   string `json:"published"`
	Checksum    string `json:"checksum"`
	Size        int64  `json:"size,omitempty"`
	Downloads   int64  `json:"downloads,omitempty"`
	Prerelease  bool   `json:"prerelease,omitempty"`
	Description string `json:"description,omitempty"`
}
