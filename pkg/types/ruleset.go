package types

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Ruleset represents a ruleset package
type Ruleset struct {
	Name     string   `json:"name" validate:"required"`
	Version  string   `json:"version" validate:"required,semver"`
	Source   string   `json:"source" validate:"required,url"`
	Files    []string `json:"files" validate:"required,min=1"`
	Checksum string   `json:"checksum" validate:"required,len=64"`
}

// ParseRulesetName parses a ruleset name into org and package components
// Supports both "package" and "org@package" formats
func ParseRulesetName(name string) (org, pkg string) {
	if strings.Contains(name, "@") {
		parts := strings.SplitN(name, "@", 2)
		return parts[0], parts[1]
	}
	return "", name
}

// FormatRulesetName formats org and package into a ruleset name
func FormatRulesetName(org, pkg string) string {
	if org == "" {
		return pkg
	}
	return fmt.Sprintf("%s@%s", org, pkg)
}

// CalculateChecksum calculates SHA256 checksum for the ruleset files
func (r *Ruleset) CalculateChecksum(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

// Validate performs basic validation on the ruleset
func (r *Ruleset) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("ruleset name is required")
	}
	if r.Version == "" {
		return fmt.Errorf("ruleset version is required")
	}
	if r.Source == "" {
		return fmt.Errorf("ruleset source is required")
	}
	if len(r.Files) == 0 {
		return fmt.Errorf("ruleset must contain at least one file")
	}
	if r.Checksum == "" {
		return fmt.Errorf("ruleset checksum is required")
	}
	if len(r.Checksum) != 64 {
		return fmt.Errorf("invalid checksum format")
	}
	return nil
}