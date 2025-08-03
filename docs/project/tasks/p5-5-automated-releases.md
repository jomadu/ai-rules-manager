# P5.5: Automated Releases

## Overview
Implement automated semantic versioning and releases based on conventional commit messages, similar to npm's semantic-release.

## Requirements
- Conventional commit message parsing
- Semantic version calculation
- Automated changelog generation
- GitHub Actions release workflow
- Cross-platform binary distribution

## Tasks
- [ ] **Conventional commit integration**:
  - Integrate with existing commitlint setup from CC.4
  - Ensure commit message validation is working
  - Document release-triggering commit types
- [ ] **Release automation workflow**:
  - Create .github/workflows/release.yml
  - GitHub Actions workflow triggered on main branch push
  - Parse commit history since last release
  - Calculate next semantic version (major.minor.patch)
  - Generate changelog from conventional commits
- [ ] **Binary build and distribution**:
  - Configure goreleaser for cross-platform builds
  - Set up asset signing and checksum generation
  - Create GitHub release with artifacts
  - Automated package registry updates
- [ ] **Integration with existing CI/CD**:
  - Coordinate with build.yml workflow
  - Ensure no conflicts with existing automation
  - Set up proper workflow dependencies

## Acceptance Criteria
- [ ] `feat:` commits trigger minor version bumps
- [ ] `fix:` commits trigger patch version bumps
- [ ] `BREAKING CHANGE:` commits trigger major version bumps
- [ ] Changelog is automatically generated from commits
- [ ] Releases are created without manual intervention
- [ ] Cross-platform binaries are built and distributed

## Dependencies
- GitHub Actions
- goreleaser
- conventional-changelog
- commitlint (from CC.4 setup)
- semantic-release or custom parser

## Files to Create
- `.github/workflows/release.yml`
- `.goreleaser.yml`
- `scripts/release.sh`
- Update existing `.commitlintrc.json` if needed

## Commit Types
- `feat`: New features (minor bump)
- `fix`: Bug fixes (patch bump)
- `BREAKING CHANGE`: Breaking changes (major bump)
- `docs`: Documentation changes
- `ci`: CI/CD changes
- `refactor`: Code refactoring
- `test`: Test changes

## Notes
- Follow semantic versioning strictly
- Ensure backward compatibility unless BREAKING CHANGE
- Generate detailed release notes from commit messages
