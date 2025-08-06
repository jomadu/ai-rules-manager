# ADR-001: Remove GitHub Registry Implementation

## Status
Accepted

## Date
2025-08-01

## Context

During implementation of P2.2 Registry Abstraction, we created registry-specific implementations for different package registry types:
- Generic HTTP (file servers)
- GitLab Package Registry (project/group level)
- GitHub Package Registry
- AWS S3

The GitHub registry implementation was based on assumptions about GitHub's package registry API that don't align with reality.

### GitHub Package Registry Reality

GitHub doesn't provide a "generic package registry" suitable for ARM's use case:

1. **GitHub Packages** are ecosystem-specific (npm, Maven, NuGet, Docker)
2. **No generic file hosting** for arbitrary package formats
3. **Complex authentication** requiring specific package formats
4. **URL patterns** don't match our assumed implementation

Our implementation assumed URLs like:
```
https://npm.pkg.github.com/download/@owner/package/version/package-version.tgz
```

But GitHub's actual APIs are:
- **npm packages**: Requires publishing as npm packages with specific metadata
- **GitHub Releases**: Different API pattern focused on source code releases

### Alternative Considered: GitHub Releases

GitHub Releases could work but would require:
- Complex tag naming conventions (`package-v1.0.0`)
- One repository per registry or complex filtering
- Publishers following GitHub release workflow
- More implementation complexity than GitLab's generic packages

## Decision

**Remove GitHub registry implementation** from ARM for the following reasons:

1. **No suitable GitHub API** for ARM's generic package hosting needs
2. **GitLab's generic package registry** is purpose-built for this use case
3. **AWS S3** provides better generic file hosting than GitHub
4. **Complexity vs. value** - GitHub Releases implementation would be complex with limited benefit
5. **Focus resources** on registries that work well with ARM's model

## Consequences

### Positive
- **Cleaner codebase** without non-functional implementation
- **Clear documentation** about actually supported registries
- **Focus on working solutions** (GitLab, S3, generic HTTP)
- **Avoid user confusion** from broken GitHub integration

### Negative
- **Reduced registry options** for users preferring GitHub
- **Future implementation needed** if GitHub adds generic package support

### Mitigation
- **Document decision** clearly in registry documentation
- **Keep door open** for future GitHub implementation if APIs improve
- **Recommend alternatives**: GitLab for teams, S3 for simple hosting

## Implementation

1. Remove `internal/registry/github.go`
2. Remove `RegistryTypeGitHub` from types
3. Remove GitHub case from registry factory
4. Remove GitHub-specific config fields (`owner`)
5. Update all documentation to remove GitHub references
6. Create this ADR to document the decision

## Future Considerations

If GitHub adds a generic package registry or if there's strong user demand for GitHub Releases integration, we can revisit this decision. The registry abstraction architecture supports adding new registry types.

For now, users wanting GitHub-based distribution should:
- Use GitLab's generic package registry
- Host files on S3 with GitHub Actions for publishing
- Use generic HTTP servers with GitHub Pages or similar
