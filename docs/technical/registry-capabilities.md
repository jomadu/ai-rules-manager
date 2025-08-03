# Registry Capabilities Matrix

## Overview

Different registry types have varying capabilities based on their underlying infrastructure and APIs.

## Capability Matrix

| Feature | GitLab | Generic HTTP | S3 | Filesystem |
|---------|--------|--------------|----|-----------| 
| **Version Discovery** | ✅ | ❌ | ❌ | ✅ |
| **Rich Metadata** | ✅ | ❌ | ❌ | ❌ |
| **Health Checks** | ✅ | ✅ | ✅ | ✅ |
| **Authentication** | ✅ | ✅ | ✅ | ❌ |
| **Direct Downloads** | ✅ | ✅ | ✅ | ✅ |

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
# Exact version required
arm install company@typescript-rules@1.0.0   # Must specify version
arm list                                      # Shows minimal metadata
```

### S3 Registry
```bash
# Exact version required
arm install company@typescript-rules@1.0.0   # Must specify version
arm list                                      # Shows S3-specific info
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
- You don't need version discovery

### Choose Filesystem When:
- You're developing or testing locally
- You have shared network storage
- You want version discovery without APIs
- You need offline access to rulesets

## Implementation Notes

### Version Discovery
- **GitLab**: Uses package API to list available versions
- **Filesystem**: Scans directory structure for version folders
- **HTTP/S3**: Not supported - users must specify exact versions

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
- **Fallback**: S3 for reliable distribution
- **Development**: Filesystem for local testing

### For Individual Users
- **Simple**: Generic HTTP for basic needs
- **Advanced**: GitLab for full feature set
- **Offline**: Filesystem for local development