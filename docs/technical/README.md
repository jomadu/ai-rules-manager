# Technical Documentation

Implementation guides and technical details for ARM developers and AI coding assistants.

## Implementation Guides

**[adr/](adr/)** - Architecture Decision Records
- Technical decisions and rationale
- Historical context for implementation choices

**[configuration-driven-targets.md](configuration-driven-targets.md)** - Target directory architecture
- How ARM handles multiple AI tool targets
- Configuration-driven installation behavior
- Directory structure and organization

**[armrc-configuration.md](armrc-configuration.md)** - .armrc configuration system
- Registry source configuration and hierarchy
- Environment variable substitution
- Configuration commands and security features

**[testing.md](testing.md)** - Test registry and testing procedures
- Local test registry setup and usage
- End-to-end testing workflows
- Available test rulesets

**[testing-uninstall.md](testing-uninstall.md)** - Uninstall command test results
- Comprehensive test scenarios and results
- Edge case handling verification
- Directory cleanup validation

## For AI Assistants

When working on ARM development:
1. **Check** `../project/tasks.md` for current priorities and status
2. **Reference** `../project/tasks/` for specific task requirements
3. **Use** implementation guides here for technical architecture details
4. **Follow** testing procedures in `testing.md` for validation
