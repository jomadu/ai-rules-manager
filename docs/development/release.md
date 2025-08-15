# Release Process

Quick guide for creating and publishing ARM releases.

## Release Types
- **Major** (v2.0.0): Breaking changes
- **Minor** (v1.1.0): New features
- **Patch** (v1.0.1): Bug fixes
- **Pre-release** (v1.1.0-alpha.1): Development versions

## Pre-Release Checklist
- [ ] All tests passing
- [ ] Code coverage ≥80%
- [ ] Documentation updated
- [ ] CHANGELOG.md updated

## Release Process

### 1. Create Release Branch
```bash
git checkout main && git pull
git checkout -b release/v1.2.0
# Update CHANGELOG.md
git add . && git commit -m "chore: prepare release v1.2.0"
git push origin release/v1.2.0
```

### 2. Create & Merge PR
```bash
gh pr create --title "Release v1.2.0" --base main --head release/v1.2.0
# After review and CI passes, merge to main
```

### 3. Tag Release
```bash
git checkout main && git pull
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0
```

### 4. Automated Release
GitHub Actions automatically:
- Builds binaries for all platforms
- Creates GitHub release
- Updates package managers

## Hotfix Process
```bash
# Create hotfix from release tag
git checkout v1.2.0
git checkout -b hotfix/v1.2.1
# Make fix, commit, create PR
gh pr create --title "Hotfix v1.2.1" --base main
# After merge, tag and push
git tag -a v1.2.1 -m "Hotfix v1.2.1"
git push origin v1.2.1
```

## Post-Release
- [ ] Verify release artifacts work
- [ ] Update package managers (Homebrew, Chocolatey)
- [ ] Announce release

## Emergency Rollback
```bash
# Delete problematic release
gh release delete v1.2.0 --yes
git tag -d v1.2.0
git push origin :refs/tags/v1.2.0
```
