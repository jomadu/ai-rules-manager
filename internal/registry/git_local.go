package registry

import (
	"context"
)

// GitLocalRegistry implements the Registry interface for local Git repositories
type GitLocalRegistry struct {
	*BaseGitRegistry
	operations GitOperations
}

// NewGitLocalRegistry creates a new local Git registry instance
func NewGitLocalRegistry(config *RegistryConfig, auth *AuthConfig) (*GitLocalRegistry, error) {
	base, err := NewBaseGitRegistry(config, auth)
	if err != nil {
		return nil, err
	}

	operations, err := NewLocalGitOperations(config.URL)
	if err != nil {
		return nil, err
	}

	return &GitLocalRegistry{
		BaseGitRegistry: base,
		operations:      operations,
	}, nil
}

// GetRulesets returns available rulesets matching the given patterns
func (g *GitLocalRegistry) GetRulesets(ctx context.Context, patterns []string) ([]RulesetInfo, error) {
	files, err := g.operations.GetFiles(ctx, "latest", patterns)
	if err != nil {
		return nil, err
	}
	return g.ConvertFilesToRulesets(files, "git-local"), nil
}

// GetRuleset returns detailed information about a specific ruleset
func (g *GitLocalRegistry) GetRuleset(ctx context.Context, name, version string) (*RulesetInfo, error) {
	// Resolve version using operations
	resolvedVersion, err := g.operations.ResolveVersion(ctx, version)
	if err != nil {
		return nil, err
	}

	rulesets, err := g.GetRulesets(ctx, []string{name + "*"})
	if err != nil {
		return nil, err
	}
	return g.FindRulesetByName(rulesets, name, resolvedVersion)
}

// DownloadRuleset downloads a ruleset to the specified directory
func (g *GitLocalRegistry) DownloadRuleset(ctx context.Context, name, version, destDir string) error {
	return g.DownloadRulesetWithPatterns(ctx, name, version, destDir, []string{"**/*"})
}

// DownloadRulesetWithPatterns downloads a ruleset with pattern matching
func (g *GitLocalRegistry) DownloadRulesetWithPatterns(ctx context.Context, name, version, destDir string, patterns []string) error {
	return g.BaseGitRegistry.DownloadRulesetWithPatterns(ctx, g.operations, version, destDir, patterns)
}

// GetVersions returns available versions for a ruleset
func (g *GitLocalRegistry) GetVersions(ctx context.Context, name string) ([]string, error) {
	return g.operations.ListVersions(ctx)
}

// ResolveVersion resolves a version specification to a concrete version
func (g *GitLocalRegistry) ResolveVersion(ctx context.Context, version string) (string, error) {
	return g.operations.ResolveVersion(ctx, version)
}

// GetFiles retrieves files matching patterns for a specific version
func (g *GitLocalRegistry) GetFiles(ctx context.Context, version string, patterns []string) (map[string][]byte, error) {
	return g.operations.GetFiles(ctx, version, patterns)
}

// DownloadRulesetWithResult downloads a ruleset and returns structured results
func (g *GitLocalRegistry) DownloadRulesetWithResult(ctx context.Context, name, version, destDir string, patterns []string) (*DownloadResult, error) {
	return g.BaseGitRegistry.DownloadRulesetWithResult(ctx, g.operations, version, destDir, patterns)
}

// GetType returns the registry type
func (g *GitLocalRegistry) GetType() string {
	return "git-local"
}

// Close cleans up any resources
func (g *GitLocalRegistry) Close() error {
	return nil
}
