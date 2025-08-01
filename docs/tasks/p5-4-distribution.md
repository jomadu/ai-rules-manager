# P5.4: Distribution

## Overview
Set up automated distribution pipeline including GitHub releases, binary signing, installation scripts, and package manager submissions.

## Requirements
- Conventional commit-based automated releases
- Semantic versioning automation
- GitHub releases with cross-platform binaries
- Binary signing and verification
- Installation script creation
- Package manager submissions (brew, apt, etc.)

## Tasks
- [ ] **Conventional commit automation**:
  - Commit message parsing (feat, fix, BREAKING CHANGE)
  - Semantic version calculation
  - Automated changelog generation
  - Release triggering on main branch pushes
- [ ] **GitHub releases automation**:
  - Cross-platform binary builds
  - Release notes from conventional commits
  - Asset upload and organization
  - Version tagging
- [ ] **Binary signing and verification**:
  - Code signing for macOS binaries
  - GPG signing for Linux binaries
  - Checksum generation and verification
  - Signature verification in install scripts
- [ ] **Installation scripts**:
  - curl-based installer script
  - Platform detection and binary selection
  - Verification of downloaded binaries
  - Installation to appropriate directories
- [ ] **Package manager submissions**:
  - Homebrew formula creation
  - Debian package creation
  - RPM package creation
  - Chocolatey package (Windows)
  - Snap package (Linux)

## Acceptance Criteria
- [ ] Releases are created automatically from conventional commits
- [ ] Binaries are signed and verifiable
- [ ] Installation script works on all platforms
- [ ] Package managers have up-to-date packages
- [ ] Users can install via multiple methods
- [ ] Installation process is secure and reliable

## Dependencies
- goreleaser (release automation)
- GitHub Actions (CI/CD)
- semantic-release or custom conventional commit parser
- conventional-changelog for release notes

## Files to Create
- `.goreleaser.yml`
- `scripts/install.sh`
- `packaging/homebrew/arm.rb`
- `packaging/debian/control`
- `.github/workflows/release.yml`

## Installation Methods
```bash
# Curl installer
curl -sSL https://install.arm.dev | sh

# Homebrew
brew install arm

# Debian/Ubuntu
wget https://github.com/user/arm/releases/latest/download/arm_linux_amd64.deb
sudo dpkg -i arm_linux_amd64.deb

# Manual download
wget https://github.com/user/arm/releases/latest/download/arm-linux-amd64
chmod +x arm-linux-amd64
sudo mv arm-linux-amd64 /usr/local/bin/arm
```

## Release Process
1. Push conventional commits to main branch
2. GitHub Actions analyzes commit history
3. Calculate semantic version (major.minor.patch)
4. Generate changelog from commit messages
5. Build cross-platform binaries with goreleaser
6. Sign binaries and create checksums
7. Create GitHub release and version tag automatically
8. Update package managers automatically

## Notes
- Consider Windows installer (MSI)
- Plan for update notifications in CLI
- Implement telemetry for installation methods (opt-in)