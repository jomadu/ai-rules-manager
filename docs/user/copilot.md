# GitHub Copilot Integration

ARM supports GitHub Copilot through custom instructions, chat participants, and prompts placed in the `.github` directory of your repository.

## Setup

### 1. Add Copilot Channel

Configure ARM to install rulesets to your `.github` directory:

```bash
arm config add channel copilot --directories .github
```

### 2. Install Copilot Rulesets

Install rulesets that include Copilot configurations:

```bash
# Install specific copilot files
arm install copilot-rules --patterns "copilot-*.md,copilot-*.yml"

# Install all rules including copilot (if registry includes them)
arm install company-rules --patterns "**/*"
```

## Supported File Types

GitHub Copilot recognizes these files in the `.github` directory:

### copilot-instructions.md
General instructions for how Copilot should behave in your repository.

**Example:**
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

### copilot-chat-participants.yml
Defines custom chat participants with specialized roles and commands.

**Example:**
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

### copilot-prompts.yml
Reusable prompts for common development tasks.

**Example:**
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

## Configuration Examples

### Basic Setup
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

### Multi-Tool Setup
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

## Best Practices

### File Organization
- Keep Copilot files focused and specific
- Use descriptive names for participants and prompts
- Group related commands under appropriate participants

### Content Guidelines
- Write clear, actionable instructions
- Include examples where helpful
- Consider your team's coding standards
- Keep security considerations in mind

### Version Control
- Include `.github/copilot-*` files in your repository
- Use ARM to sync rules across team members
- Version your ruleset changes appropriately

## Troubleshooting

### Files Not Applied
If Copilot isn't using your custom files:
1. Verify files are in `.github` directory
2. Check file names match expected patterns
3. Ensure files have correct YAML/Markdown syntax
4. Restart your IDE/editor

### Permission Issues
```bash
Error [FILESYSTEM]: Permission denied writing to .github
```
**Solution**: Create the directory if it doesn't exist:
```bash
mkdir -p .github
```

### Syntax Errors
```bash
Error: Invalid YAML in copilot-prompts.yml
```
**Solution**: Validate YAML syntax using online tools or:
```bash
python -c "import yaml; yaml.safe_load(open('.github/copilot-prompts.yml'))"
```

## Examples

See the [examples directory](../examples/) for complete examples of:
- [copilot-instructions.md](../examples/copilot-instructions.md)
- [copilot-chat-participants.yml](../examples/copilot-chat-participants.yml)
- [copilot-prompts.yml](../examples/copilot-prompts.yml)