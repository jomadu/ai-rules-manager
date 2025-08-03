# Registry Capabilities Matrix

## Overview

Different registry types have varying capabilities based on their underlying infrastructure and APIs.

## Capability Matrix

| Feature | GitLab | Generic HTTP | S3 | Filesystem |
|---------|--------|--------------|----|-----------| 
| **Version Discovery** | ✅ | ❌ | ✅ | ✅ |
| **Rich Metadata** | ✅ | ❌ | ❌ | ❌ |
| **Health Checks** | ✅ | ✅ | ✅ | ✅ |
| **Authentication** | ✅ | ✅ | ✅ | ❌ |
| **Direct Downloads** | ✅ | ✅ | ✅ | ✅ |

**Legend**: ✅ Full Support | ❌ Not Supported

## Usage Patterns

### GitLab Package Registry
```bash
# Full feature support
arm install company@typescript-rules          # Latest version
arm install company@typescript-rules@1.0.0   # Specific version
arm list                                      # Shows rich metadata
arm outdated                                  # Version comparison
```

### Generic HTTP Registry
```bash
# Exact version required - no discovery
arm install company@typescript-rules@1.0.0   # Must specify version
arm list                                      # Shows minimal metadata
```

### S3 Registry
```bash
# Version discovery via S3 prefix listing
arm install company@typescript-rules          # Lists S3 prefixes
arm install company@typescript-rules@1.0.0   # Specific version
arm list                                      # Shows S3 info + versions
```

### Filesystem Registry
```bash
# Version discovery supported
arm install company@typescript-rules          # Latest from filesystem
arm install company@typescript-rules@1.0.0   # Specific version
arm list                                      # Shows available versions
```

## Registry Selection Guide

### Choose GitLab When:
- You need version discovery and rich metadata
- You're already using GitLab for source control
- You want integrated package management
- You need download statistics and package information

### Choose Generic HTTP When:
- You have existing HTTP file servers
- You want simple, predictable URLs
- You don't need version discovery
- You prefer minimal infrastructure

### Choose S3 When:
- You need global distribution and reliability
- You're already using AWS infrastructure
- You want to organize packages with prefixes
- You need version discovery with S3's hierarchical structure

### Choose Filesystem When:
- You're developing or testing locally
- You have shared network storage
- You want version discovery without APIs
- You need offline access to rulesets

## Implementation Notes

### Version Discovery
- **GitLab**: Uses package API to list available versions
- **Filesystem**: Scans directory structure for version folders
- **S3**: Uses S3 list-objects-v2 API to discover version prefixes
- **HTTP**: Not supported - requires exact version specification

### Metadata Generation
- **GitLab**: Rich metadata from package API (sizes, dates, downloads)
- **Filesystem**: Basic metadata from directory scanning
- **HTTP/S3**: Minimal metadata (name, repository URL only)

### Authentication
- **GitLab**: Personal access tokens or CI tokens
- **HTTP**: Bearer tokens or basic auth
- **S3**: AWS access keys or IAM roles
- **Filesystem**: No authentication (local access)

## Best Practices

### For Teams
- **Primary**: GitLab for rich package management
- **Scalable**: S3 for reliable distribution with version discovery
- **Development**: Filesystem for local testing

### For Individual Users
- **Simple**: Generic HTTP for basic needs (exact versions)
- **Advanced**: GitLab or S3 for version discovery
- **Offline**: Filesystem for local development