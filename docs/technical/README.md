# Technical Documentation

Implementation guides and technical details for ARM developers and AI coding assistants.

## Documentation Sections

**[core/](core/)** - Core ARM architecture and implementation
- Configuration system and target handling
- Metadata generation and registry capabilities
- Testing procedures and architecture decisions

**[registries/](registries/)** - Registry-specific implementation guides
- GitLab, S3, HTTP, and Filesystem registry setup
- Publishing workflows and troubleshooting
- Registry selection and best practices

**[s3-registry.md](s3-registry.md)** - S3 registry guide
- S3 prefix structure and version discovery
- Configuration, publishing workflows, and best practices
- AWS integration and troubleshooting

**[http-registry.md](http-registry.md)** - HTTP registry guide
- Generic HTTP file server setup and configuration
- URL patterns, authentication, and server examples
- Publishing workflows and limitations

**[filesystem-registry.md](filesystem-registry.md)** - Filesystem registry guide
- Local directory structure and version discovery
- Development workflows and use cases
- Performance considerations and troubleshooting

## For AI Assistants

When working on ARM development:
1. **Check** `../project/tasks.md` for current priorities and status
2. **Reference** `../project/tasks/` for specific task requirements
3. **Use** implementation guides here for technical architecture details
4. **Follow** testing procedures in `testing.md` for validation
