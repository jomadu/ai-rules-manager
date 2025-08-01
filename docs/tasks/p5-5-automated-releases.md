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
- [ ] **Conventional commit setup**:
  - Configure commitlint for commit message validation
  - Document commit message format (feat, fix, BREAKING CHANGE)
  - Set up pre-commit hooks for validation
- [ ] **Release automation workflow**:
  - GitHub Actions workflow triggered on main branch push
  - Parse commit history since last release
  - Calculate next semantic version
  - Generate changelog from commit messages
- [ ] **Binary build and distribution**:
  - Cross-platform binary compilation
  - Asset signing and checksum generation
  - GitHub release creation with artifacts
  - Automated package registry updates

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
- commitlint

## Files to Create
- `.github/workflows/release.yml`
- `.commitlintrc.json`
- `.goreleaser.yml`
- `scripts/release.sh`

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