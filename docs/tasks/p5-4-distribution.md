# P5.4: Distribution

## Overview
Set up automated distribution pipeline including GitHub releases, binary signing, installation scripts, and package manager submissions.

## Requirements
- GitHub releases automation
- Binary signing and verification
- Installation script creation
- Package manager submissions (brew, apt, etc.)

## Tasks
- [ ] **GitHub releases automation**:
  - Automated releases on version tags
  - Cross-platform binary builds
  - Release notes generation
  - Asset upload and organization
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
- [ ] Releases are created automatically on tags
- [ ] Binaries are signed and verifiable
- [ ] Installation script works on all platforms
- [ ] Package managers have up-to-date packages
- [ ] Users can install via multiple methods
- [ ] Installation process is secure and reliable

## Dependencies
- goreleaser (release automation)
- GitHub Actions (CI/CD)

## Files to Create
- `.goreleaser.yml`
- `scripts/install.sh`
- `packaging/homebrew/mrm.rb`
- `packaging/debian/control`
- `.github/workflows/release.yml`

## Installation Methods
```bash
# Curl installer
curl -sSL https://install.mrm.dev | sh

# Homebrew
brew install mrm

# Debian/Ubuntu
wget https://github.com/user/mrm/releases/latest/download/mrm_linux_amd64.deb
sudo dpkg -i mrm_linux_amd64.deb

# Manual download
wget https://github.com/user/mrm/releases/latest/download/mrm-linux-amd64
chmod +x mrm-linux-amd64
sudo mv mrm-linux-amd64 /usr/local/bin/mrm
```

## Release Process
1. Create version tag
2. GitHub Actions triggers build
3. Goreleaser builds cross-platform binaries
4. Binaries are signed
5. Release is created with assets
6. Package managers are updated

## Notes
- Consider Windows installer (MSI)
- Plan for update notifications in CLI
- Implement telemetry for installation methods (opt-in)