# Registry Metadata Generation

## Overview

ARM's registry system uses a hybrid approach for metadata generation, adapting to each registry's capabilities rather than requiring publishers to manually maintain metadata files.

## Registry-Specific Approaches

### GitLab Package Registry ✅
**Uses native GitLab API** - No manual metadata.json required

- **API**: `/api/v4/projects/{id}/packages?package_name={name}`
- **Generates**: Version lists, file sizes, publication dates
- **Benefits**: Automatic metadata from GitLab's package management
- **Fallback**: None needed - GitLab API is authoritative

```go
// Converts GitLab package data to ARM metadata
func (r *GitLabRegistry) convertGitLabToMetadata(pkg GitLabPackage, name string) *Metadata
```

### Generic HTTP Registry ✅
**No metadata assumptions - direct download only**

- **Approach**: Provides minimal metadata, focuses on direct downloads
- **Requirement**: Publishers specify exact versions in install commands
- **Benefits**: No schema requirements, works with any HTTP server
- **Limitation**: No version discovery or rich metadata

### AWS S3 Registry ✅
**Similar to Generic HTTP**

- **Approach**: Minimal metadata with S3-specific information
- **Requirement**: Exact version specification required
- **Benefits**: Leverages S3's reliability and global distribution
- **Limitation**: No automatic version discovery

### Local Filesystem Registry ✅
**Scans directory structure only**

- **Discovery**: Scans directories for version folders
- **Metadata**: Generated from directory structure only
- **Structure**: `{path}/{package}/{version}/{package}-{version}.tar.gz`
- **No metadata files**: Works purely from filesystem organization

## Metadata Structure

### Core Fields (Always Present)
```json
{
  "name": "package-name",
  "description": "Generated or provided description",
  "versions": [...]
}
```

### Enhanced Fields (Registry-Dependent)
```json
{
  "homepage": "https://...",
  "license": "MIT",
  "keywords": ["ai", "rules"],
  "maintainers": ["user@example.com"],
  "downloads": 1234,
  "lastModified": "2025-01-08T10:00:00Z",
  "extra": {
    "gitlab_id": "123",
    "package_type": "generic"
  }
}
```

## Publisher Requirements

### GitLab Package Registry
- Upload packages via GitLab's package API
- Metadata generated automatically from GitLab

### Generic HTTP/S3 Registries
- Package archive: `{package}-{version}.tar.gz`
- Proper URL structure for direct downloads
- Users must specify exact versions

### Filesystem Registry
- Directory structure: `{path}/{package}/{version}/`
- Package archives in version directories
- No metadata files required

## Implementation Benefits

### For Publishers
- **No metadata burden**: No schema files to maintain
- **Registry-native**: Uses each registry's natural capabilities
- **Simple structure**: Just organize files predictably

### For Users
- **Consistent interface**: Same metadata API across all registries
- **Rich information**: Available from GitLab's native API
- **Reliable downloads**: Direct file access from all registry types

## Future Enhancements

### Planned Registry Support
- **GitHub Releases**: Parse release information for metadata
- **Docker Registry**: Adapt for container-based rulesets

### Metadata Enrichment
- **Download statistics**: Track usage across registries
- **Dependency scanning**: Analyze ruleset dependencies
- **Quality metrics**: Automated ruleset quality assessment

## Best Practices

### For Registry Operators
1. **GitLab**: Use native package registry features
2. **S3/HTTP**: Provide direct download URLs
3. **Filesystem**: Organize with clear directory structure

### For Publishers
1. **GitLab**: Use GitLab's package upload features
2. **S3/HTTP**: Upload packages with predictable naming
3. **Filesystem**: Organize files in version directories