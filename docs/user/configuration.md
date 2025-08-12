# Configuration Guide

Complete guide to configuring ARM with `.armrc` and `arm.json` files.

## Configuration Files

ARM uses two configuration files:
- **`.armrc`** - Registry settings and defaults (INI format)
- **`arm.json`** - Channels and rulesets (JSON format)

## File Locations

### Global vs Local
- **Global**: `~/.arm/.armrc` and `~/.arm/arm.json`
- **Local**: `./.armrc` and `./arm.json` (current directory)

Local settings override global settings at the key level.

### Generating Stubs

Create starter configuration files:

```bash
# Generate in current directory
arm install

# Generate globally
arm install --global
```

## Registry Configuration (.armrc)

### Basic Registry Setup

```bash
# Add different registry types
arm config add registry bowser-castle https://github.com/bowser-castle/security-rules.example --type=git
arm config add registry koopa-troopa koopa-castle-rules --type=s3 --region=us-east-1
arm config add registry toad-house https://gitlab.mushroom-kingdom.example/projects/456 --type=gitlab --authToken=$GITLAB_TOKEN
```

### Registry-Specific Settings

Override defaults for specific registries:

```bash
# Increase concurrency for a fast registry
arm config set registries.bowser-castle.concurrency 5

# Set custom rate limit
arm config set registries.koopa-troopa.rateLimit 20/minute

# Configure S3 profile
arm config set registries.koopa-troopa.profile mario-aws-profile
```

### Type Defaults

Set defaults for all registries of a type:

```bash
# All Git registries
arm config set git.concurrency 2
arm config set git.rateLimit 15/minute

# All S3 registries
arm config set s3.concurrency 10
arm config set s3.rateLimit 100/hour
```

### Environment Variables

Use environment variables in configuration:

```bash
arm config add registry private-castle https://github.com/private/repo.example --type=git --authToken=$GITHUB_TOKEN
```

## Channel Configuration (arm.json)

### Adding Channels

```bash
# Single directory
arm config add channel cursor --directories ~/.cursor/rules

# Multiple directories
arm config add channel cursor --directories "~/.cursor/rules,./project-rules"

# Environment variables supported
arm config add channel q --directories '$HOME/.aws/amazonq/rules'

# GitHub Copilot channel (uses .github directory)
arm config add channel copilot --directories .github
```

### Manual JSON Editing

You can also edit `arm.json` directly:

```json
{
  "engines": {
    "arm": "^1.2.3"
  },
  "channels": {
    "cursor": {
      "directories": ["~/.cursor/rules", "./custom-cursor"]
    },
    "q": {
      "directories": ["$HOME/.aws/amazonq/rules"]
    },
    "copilot": {
      "directories": [".github"]
    }
  },
  "rulesets": {
    "mushroom-kingdom": {
      "power-up-rules": {
        "version": "^1.0.0",
        "patterns": ["rules/*.md", "docs/*.mdc"]
      }
    }
  }
}
```

### GitHub Copilot Configuration

GitHub Copilot supports custom instructions, chat participants, and prompts through files placed in the `.github` directory:

```bash
# Add Copilot channel
arm config add channel copilot --directories .github

# Install rulesets that include Copilot configurations
arm install copilot-rules --patterns "copilot-*.md,copilot-*.yml"
```

#### Supported Copilot Files

**copilot-instructions.md** - General instructions for how Copilot should behave in your repository.

Example:
```markdown
# GitHub Copilot Instructions

You are an AI coding assistant. Follow these guidelines:

## Code Style
- Use meaningful variable names
- Include appropriate comments
- Follow project conventions

## Security
- Never include hardcoded secrets
- Validate user inputs
- Follow security best practices
```

**copilot-chat-participants.yml** - Defines custom chat participants with specialized roles and commands.

Example:
```yaml
participants:
  - name: "code-reviewer"
    description: "AI assistant specialized in code review"
    commands:
      - name: "review"
        description: "Perform comprehensive code review"
      - name: "security"
        description: "Focus on security vulnerabilities"
```

**copilot-prompts.yml** - Reusable prompts for common development tasks.

Example:
```yaml
prompts:
  - name: "explain-code"
    description: "Explain what code does in simple terms"
    content: |
      Please explain this code:
      - What does it do?
      - How does it work?
      - Any potential issues?

  - name: "add-tests"
    description: "Generate tests for code"
    content: |
      Write comprehensive tests for this code:
      - Cover happy path scenarios
      - Include edge cases
      - Follow testing conventions
```

#### Copilot Configuration Examples

**Basic Setup:**
```json
{
  "channels": {
    "copilot": {
      "directories": [".github"]
    }
  },
  "rulesets": {
    "my-company": {
      "copilot-rules": {
        "version": "^1.0.0",
        "patterns": ["copilot-*.md", "copilot-*.yml"]
      }
    }
  }
}
```

**Multi-Tool Setup:**
```json
{
  "channels": {
    "cursor": {
      "directories": ["~/.cursor/rules"]
    },
    "q": {
      "directories": ["~/.aws/amazonq/rules"]
    },
    "copilot": {
      "directories": [".github"]
    }
  },
  "rulesets": {
    "my-company": {
      "universal-rules": {
        "version": "^2.1.0",
        "patterns": ["**/*.md", "**/*.yml"]
      }
    }
  }
}
```

#### Copilot Best Practices

**File Organization:**
- Keep Copilot files focused and specific
- Use descriptive names for participants and prompts
- Group related commands under appropriate participants

**Content Guidelines:**
- Write clear, actionable instructions
- Include examples where helpful
- Consider your team's coding standards
- Keep security considerations in mind

**Version Control:**
- Include `.github/copilot-*` files in your repository
- Use ARM to sync rules across team members
- Version your ruleset changes appropriately

## Configuration Commands

### View Configuration

```bash
# Show merged configuration
arm config list

# Show specific value
arm config get registries.default

# Show global only
arm config list --global
```

### Modify Configuration

```bash
# Set values
arm config set registries.default mushroom-kingdom
arm config set git.concurrency 3

# Remove registries/channels
arm config remove registry old-registry
arm config remove channel unused-channel
```

## Advanced Configuration

### Network Settings

```bash
# Increase timeout for slow connections
arm config set network.timeout 60

# Configure retry behavior
arm config set network.retry.maxAttempts 5
arm config set network.retry.backoffMultiplier 2.0
```

### Cache Settings

```bash
# Change cache location
arm config set cache.path ~/my-arm-cache

# Increase cache size
arm config set cache.maxSize 2GB

# Adjust TTL (time to live)
arm config set cache.ttl 7200
```

## Configuration Validation

ARM validates configuration when you run commands:

```bash
# Check configuration
arm config list
```

Common validation errors and fixes are shown automatically.

## Troubleshooting

### Invalid Registry Type
```bash
Error [CONFIG]: Unknown registry type 'invalid' in registry 'my-registry'
Details: Supported types: git, https, s3, gitlab, local
```
**Solution**: Use a supported registry type.

### Missing Required Field
```bash
Error [CONFIG]: Missing required field 'region' for S3 registry
```
**Solution**: Add the required field:
```bash
arm config set registries.koopa-troopa.region us-east-1
```

### Environment Variable Not Found
If `$GITHUB_TOKEN` is not set, ARM will use an empty string. Set the variable:
```bash
export GITHUB_TOKEN=your_token_here
```

### GitHub Copilot Files Not Applied
If Copilot isn't using your custom files:
1. Verify files are in `.github` directory
2. Check file names match expected patterns (`copilot-*.md`, `copilot-*.yml`)
3. Ensure files have correct YAML/Markdown syntax
4. Restart your IDE/editor

### Copilot Directory Permission Issues
```bash
Error [FILESYSTEM]: Permission denied writing to .github
```
**Solution**: Create the directory if it doesn't exist:
```bash
mkdir -p .github
```

### Copilot YAML Syntax Errors
```bash
Error: Invalid YAML in copilot-prompts.yml
```
**Solution**: Validate YAML syntax using online tools or:
```bash
python -c "import yaml; yaml.safe_load(open('.github/copilot-prompts.yml'))"
```
