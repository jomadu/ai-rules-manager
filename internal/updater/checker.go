package updater

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/jomadu/arm/internal/registry"
	"github.com/jomadu/arm/pkg/types"
)

type CheckResult struct {
	Name       string
	Current    string
	Available  string
	Constraint string
	Status     CheckStatus
	Error      error
}

type CheckStatus int

const (
	CheckUpToDate CheckStatus = iota
	CheckOutdated
	CheckError
	CheckNoCompatible
)

func (s CheckStatus) String() string {
	switch s {
	case CheckUpToDate:
		return "Up to date"
	case CheckOutdated:
		return "Update available"
	case CheckError:
		return "Error"
	case CheckNoCompatible:
		return "No compatible update"
	default:
		return "Unknown"
	}
}

type Checker struct {
	manager *registry.Manager
}

func NewChecker(manager *registry.Manager) *Checker {
	return &Checker{manager: manager}
}

func (c *Checker) CheckRuleset(ruleset InstalledRuleset, constraint string) CheckResult {
	// Parse current version
	currentVer, err := version.NewVersion(ruleset.Version)
	if err != nil {
		return CheckResult{
			Name:       ruleset.Name,
			Current:    ruleset.Version,
			Constraint: constraint,
			Status:     CheckError,
			Error:      fmt.Errorf("invalid current version: %w", err),
		}
	}

	// Parse constraint
	constraints, err := version.NewConstraint(constraint)
	if err != nil {
		return CheckResult{
			Name:       ruleset.Name,
			Current:    ruleset.Version,
			Constraint: constraint,
			Status:     CheckError,
			Error:      fmt.Errorf("invalid version constraint: %w", err),
		}
	}

	// Get available versions using cache
	versions, err := c.manager.CachedListVersions(ruleset.Source, ruleset.Name)
	if err != nil {
		return CheckResult{
			Name:       ruleset.Name,
			Current:    ruleset.Version,
			Constraint: constraint,
			Status:     CheckError,
			Error:      fmt.Errorf("failed to list versions: %w", err),
		}
	}

	// Find latest compatible version
	var latestValid *version.Version
	for _, vStr := range versions {
		v, err := version.NewVersion(vStr)
		if err != nil {
			continue
		}

		if constraints.Check(v) && (latestValid == nil || v.GreaterThan(latestValid)) {
			latestValid = v
		}
	}

	if latestValid == nil {
		return CheckResult{
			Name:       ruleset.Name,
			Current:    ruleset.Version,
			Constraint: constraint,
			Status:     CheckNoCompatible,
			Error:      fmt.Errorf("no compatible versions found"),
		}
	}

	// Check if update is available
	if latestValid.GreaterThan(currentVer) {
		return CheckResult{
			Name:       ruleset.Name,
			Current:    ruleset.Version,
			Available:  latestValid.String(),
			Constraint: constraint,
			Status:     CheckOutdated,
		}
	}

	return CheckResult{
		Name:       ruleset.Name,
		Current:    ruleset.Version,
		Available:  latestValid.String(),
		Constraint: constraint,
		Status:     CheckUpToDate,
	}
}

func (c *Checker) CheckAll() ([]CheckResult, error) {
	// Load lock file
	lockFile, err := types.LoadLockFile("rules.lock")
	if err != nil {
		return nil, fmt.Errorf("failed to load lock file: %w", err)
	}

	// Load manifest for constraints
	manifest, err := types.LoadManifest("rules.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	var results []CheckResult
	for name, dep := range lockFile.Dependencies {
		constraint, exists := manifest.Dependencies[name]
		if !exists {
			results = append(results, CheckResult{
				Name:    name,
				Current: dep.Version,
				Status:  CheckError,
				Error:   fmt.Errorf("no constraint found in rules.json"),
			})
			continue
		}

		ruleset := InstalledRuleset{
			Name:    name,
			Version: dep.Version,
			Source:  dep.Source,
		}

		result := c.CheckRuleset(ruleset, constraint)
		results = append(results, result)
	}

	return results, nil
}
