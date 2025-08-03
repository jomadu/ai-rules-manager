# Registry Documentation

Registry-specific implementation guides for ARM's supported registry types.

## Registry Types

**[gitlab-registry.md](gitlab-registry.md)** - GitLab Package Registry
- Native GitLab API integration with full metadata
- Project and group-level package registries
- Version discovery and rich package information

**[s3-registry.md](s3-registry.md)** - AWS S3 Registry
- S3 prefix-based version discovery
- Global distribution and scalability
- Custom prefix organization

**[http-registry.md](http-registry.md)** - Generic HTTP Registry
- Simple HTTP file server setup
- Direct downloads with predictable URLs
- Minimal infrastructure requirements

**[filesystem-registry.md](filesystem-registry.md)** - Local Filesystem Registry
- Directory-based version discovery
- Development and offline usage
- Local and shared storage support

## Common Concepts

### Package Structure
All registries follow consistent package naming:
```
{package}-{version}.tar.gz
{org}/{package}-{version}.tar.gz  # With organization scope
```

### Configuration Format
```ini
[sources.name]
type = {gitlab|s3|generic|filesystem}
# Registry-specific fields
```

### Version Discovery Support
- ✅ **GitLab**: Native API
- ✅ **S3**: Prefix listing
- ✅ **Filesystem**: Directory scanning
- ❌ **HTTP**: Exact versions required

## Quick Reference

| Registry | Version Discovery | Authentication | Best For |
|----------|------------------|----------------|----------|
| GitLab | ✅ Full | Token | Teams using GitLab |
| S3 | ✅ Prefix | AWS Keys | Scalable distribution |
| HTTP | ❌ None | Bearer/Basic | Simple setups |
| Filesystem | ✅ Directory | None | Development/Testing |
