# Configuration Guide

## Configuration Files

- **`.armrc`** - INI format for registries and system settings
- **`arm.json`** - JSON format for channels and rulesets
- **`arm.lock`** - Auto-generated locked versions

**Locations**: Global (`~/.arm/`) and local (project root). Local overrides global at the key level.

## Registry Configuration

### Adding Registries

**Git**: `arm config add registry name https://github.com/org/repo --type=git --authToken=$TOKEN`

**S3**: `arm config add registry name bucket-name --type=s3 --region=us-east-1`

**Local**: `arm config add registry name /path/to/rules --type=local`

### Registry Types
- **git** - GitHub/GitLab repositories
- **s3** - AWS S3 buckets
- **local** - Local directories
- **https** - HTTP registries

## Channel Configuration

Channels define where rulesets are installed.

**Add channels**: `arm config add channel cursor --directories .cursor/rules`

**Multiple directories**: Use comma-separated paths or add multiple channels.

## Ruleset Configuration

**Install rulesets**: `arm install ruleset-name@version --patterns "*.md"`

**Version constraints**: Use `^1.0.0`, `~1.0.0`, `>=1.0.0`, or `latest`

**Specific registry**: `arm install registry/ruleset@version`

## Environment Variables

**Authentication**: Set `GITHUB_TOKEN`, `GITLAB_TOKEN`, `AWS_PROFILE`, etc.

**Network**: Configure timeout, retry attempts, and rate limits in `.armrc`

**Cache**: Set cache path, size limits, and TTL in `.armrc`

## Engine Configuration

Specify ARM version requirements in `arm.json` engines section.

## Configuration Commands

**View**: `arm config list`, `arm config get key`

**Set**: `arm config set key value`

**Remove**: `arm config remove registry name`, `arm config remove channel name`

## Team Configuration

**Commit to repo**: `arm.json` and `arm.lock` for reproducible builds

**Keep private**: `.armrc` files with sensitive tokens

**Environment-specific**: Use different `arm.json` files for dev/prod environments

## Configuration Validation

ARM automatically validates configuration on load. Test manually with:
- `arm info ruleset-name` - Test registry connectivity
- `arm list` - Verify configuration
- `arm config list` - Check syntax

## Troubleshooting

**Registry issues**: Check with `arm config get registries.name`

**Authentication**: Verify environment variables are set

**Permissions**: Ensure directories exist and are writable

**Version conflicts**: Check `arm.lock` and run `arm update`

**Reset**: Delete config files and run `arm install` to regenerate

## Best Practices

**Security**: Store tokens in environment variables, use HTTPS registries, rotate tokens regularly

**Organization**: Use descriptive names, group related rulesets, document registry purposes

**Performance**: Configure appropriate cache sizes, use local registries for development

**Maintenance**: Update ARM regularly, clean cache, review locked versions
