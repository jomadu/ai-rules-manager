# Release Process

How to create and publish ARM releases.

## Release Types

### Semantic Versioning
- **Major** (v2.0.0): Breaking changes, API changes
- **Minor** (v1.1.0): New features, backward compatible
- **Patch** (v1.0.1): Bug fixes, backward compatible
- **Pre-release** (v1.1.0-alpha.1): Development versions

### Release Cadence
- **Major**: Annually or for significant changes
- **Minor**: Monthly or when features are ready
- **Patch**: As needed for critical fixes
- **Pre-release**: Weekly during development cycles

## Pre-Release Checklist

### Code Quality
- [ ] All tests passing
- [ ] Code coverage ≥80%
- [ ] No critical security vulnerabilities
- [ ] Documentation updated
- [ ] CHANGELOG.md updated

### Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] End-to-end tests pass
- [ ] Manual testing on all platforms
- [ ] Performance regression testing

### Documentation
- [ ] README.md updated
- [ ] API documentation current
- [ ] User guides updated
- [ ] Migration guides (for breaking changes)

## Release Process

### 1. Prepare Release Branch
```bash
# Create release branch from main
git checkout main
git pull origin main
git checkout -b release/v1.2.0

# Update version in code if needed
# Update CHANGELOG.md
# Commit changes
git add .
git commit -m "chore: prepare release v1.2.0"
git push origin release/v1.2.0
```

### 2. Create Release PR
```bash
# Create PR from release branch to main
gh pr create \
  --title "Release v1.2.0" \
  --body "$(cat CHANGELOG.md | sed -n '/## \[1.2.0\]/,/## \[/p' | head -n -1)" \
  --base main \
  --head release/v1.2.0
```

### 3. Review and Merge
- [ ] Code review completed
- [ ] All CI checks pass
- [ ] Documentation review completed
- [ ] Security review (for major releases)
- [ ] Merge PR to main

### 4. Create Git Tag
```bash
# After PR is merged
git checkout main
git pull origin main

# Create annotated tag
git tag -a v1.2.0 -m "Release v1.2.0

## Features
- Add S3 registry support
- Improve cache performance

## Bug Fixes
- Fix pattern matching edge cases
- Resolve authentication timeout issues

## Breaking Changes
- None
"

# Push tag to trigger release
git push origin v1.2.0
```

### 5. Automated Release
GitHub Actions automatically:
- Builds binaries for all platforms
- Creates GitHub release
- Uploads release artifacts
- Updates package managers
- Sends notifications

## Release Automation

### GoReleaser Configuration
`.goreleaser.yml` handles:
- Multi-platform builds
- Archive creation
- Checksum generation
- GitHub release creation
- Changelog generation

### GitHub Actions
`.github/workflows/release.yml`:
```yaml
name: Release
on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v4
        with:
          go-version: '1.23.2'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v4
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Post-Release Tasks

### 1. Update Package Managers
```bash
# Update Homebrew formula
cd homebrew-arm
./update-formula.sh v1.2.0
git add .
git commit -m "Update ARM to v1.2.0"
git push origin main

# Update Chocolatey package
cd chocolatey-arm
./update-package.ps1 v1.2.0
```

### 2. Update Documentation
```bash
# Update documentation site
cd docs-site
./deploy.sh v1.2.0

# Update Docker images
docker build -t arm:v1.2.0 .
docker push arm:v1.2.0
```

### 3. Announce Release
- [ ] GitHub release notes published
- [ ] Blog post written (for major releases)
- [ ] Social media announcement
- [ ] Community notifications
- [ ] Update project website

## Hotfix Process

### Critical Bug Fixes
```bash
# Create hotfix branch from latest release tag
git checkout v1.2.0
git checkout -b hotfix/v1.2.1

# Make minimal fix
# Update CHANGELOG.md
git add .
git commit -m "fix: resolve critical authentication issue"

# Create PR to main
gh pr create \
  --title "Hotfix v1.2.1: Fix authentication issue" \
  --body "Critical fix for authentication timeout" \
  --base main \
  --head hotfix/v1.2.1

# After merge, tag and release
git checkout main
git pull origin main
git tag -a v1.2.1 -m "Hotfix v1.2.1: Fix authentication issue"
git push origin v1.2.1
```

## Release Validation

### Pre-Release Testing
```bash
# Build release candidate
make build-all

# Test on multiple platforms
./test-release.sh v1.2.0-rc.1

# Performance testing
./benchmark-release.sh v1.2.0-rc.1

# Security scanning
./security-scan.sh v1.2.0-rc.1
```

### Post-Release Validation
```bash
# Verify release artifacts
curl -sSL https://github.com/max-dunn/ai-rules-manager/releases/download/v1.2.0/arm-linux-amd64 -o arm-test
chmod +x arm-test
./arm-test version

# Test installation script
curl -sSL https://raw.githubusercontent.com/max-dunn/ai-rules-manager/main/scripts/install.sh | bash

# Verify package managers
brew install max-dunn/arm/arm
choco install ai-rules-manager
```

## Rollback Procedures

### GitHub Release Rollback
```bash
# Delete problematic release
gh release delete v1.2.0 --yes

# Delete tag
git tag -d v1.2.0
git push origin :refs/tags/v1.2.0

# Create new release with fixes
git tag -a v1.2.1 -m "Release v1.2.1 (fixes v1.2.0 issues)"
git push origin v1.2.1
```

### Package Manager Rollback
```bash
# Homebrew
cd homebrew-arm
git revert HEAD
git push origin main

# Chocolatey
# Contact Chocolatey maintainers for package removal
```

## Release Metrics

### Track Release Health
- Download counts by platform
- Installation success rates
- User feedback and issues
- Performance impact
- Security vulnerability reports

### Release Dashboard
Monitor:
- Build success rates
- Test coverage trends
- Release frequency
- Time to fix critical issues
- User adoption rates

## Communication

### Release Notes Template
```markdown
# ARM v1.2.0

## 🚀 Features
- **S3 Registry Support**: Install rulesets from AWS S3 buckets
- **Improved Caching**: 50% faster repeated operations
- **Pattern Exclusions**: Support for `!pattern` exclusion syntax

## 🐛 Bug Fixes
- Fix authentication timeout for large repositories
- Resolve pattern matching edge cases with nested directories
- Correct version resolution for pre-release tags

## 📚 Documentation
- New S3 registry setup guide
- Updated team configuration examples
- Improved troubleshooting documentation

## 🔧 Internal Changes
- Refactored cache implementation for better performance
- Added comprehensive integration tests
- Improved error handling and logging

## 📦 Installation
```bash
# Install script
curl -sSL https://raw.githubusercontent.com/max-dunn/ai-rules-manager/main/scripts/install.sh | bash

# Homebrew
brew install max-dunn/arm/arm

# Chocolatey
choco install ai-rules-manager
```

## 🔄 Upgrade Notes
No breaking changes in this release. Existing configurations will continue to work.

## 📈 Metrics
- Binary size: 12.3MB (↓5% from v1.1.0)
- Test coverage: 87% (↑3% from v1.1.0)
- Build time: 2m 15s (↓30s from v1.1.0)

## 🙏 Contributors
Thanks to all contributors who made this release possible!
```

### Notification Channels
- GitHub Releases (automatic)
- Project Discord/Slack
- Twitter/LinkedIn announcements
- Developer newsletters
- Community forums

## Emergency Procedures

### Critical Security Issues
1. **Immediate Response**
   - Assess severity and impact
   - Create private security advisory
   - Develop fix in private repository

2. **Coordinated Disclosure**
   - Notify affected users privately
   - Prepare security release
   - Coordinate with security researchers

3. **Security Release**
   - Fast-track release process
   - Clear security advisory
   - Update all distribution channels
   - Monitor for exploitation

### Release Failure Recovery
1. **Identify Issue**
   - Monitor release metrics
   - Respond to user reports
   - Analyze failure patterns

2. **Quick Response**
   - Hotfix for critical issues
   - Communication to users
   - Rollback if necessary

3. **Post-Mortem**
   - Document what went wrong
   - Improve release process
   - Update testing procedures
