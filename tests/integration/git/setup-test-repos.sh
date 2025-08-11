#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default repository name
DEFAULT_REPO="ai-rules-manager-test-git-registry"

usage() {
    echo "Usage: $0 [repo-name]"
    echo ""
    echo "Creates a comprehensive test repository for ARM testing."
    echo ""
    echo "Arguments:"
    echo "  repo-name    - Name for test repository (default: $DEFAULT_REPO)"
    echo ""
    echo "Examples:"
    echo "  $0                    # Use default name"
    echo "  $0 my-test-repo      # Custom name"
    echo ""
    echo "Requirements:"
    echo "  - GitHub CLI (gh) must be installed and authenticated"
    echo "  - Git must be installed"
}

log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}



create_version_1_0_0() {
    mkdir -p rules/advanced cursor amazon-q

    cat > README.md << 'EOF'
# ARM Test Repository

Comprehensive test repository for ARM (AI Rules Manager) testing.

## Repository Structure

- `ghost-hunting.md` - Basic ghost hunting guidelines
- `rules/mansion-maintenance.md` - Maintenance rules
- `rules/advanced/boss-battles.md` - Advanced techniques
- `cursor/its-a-me.md` - Cursor-specific rules
- `amazon-q/luigi-assistant.md` - Amazon Q rules
- `config.json` - Configuration file
EOF

    cat > ghost-hunting.md << 'EOF'
# Ghost Hunting Guidelines

*Basic ghost hunting techniques for code debugging.*

## Rule 1: Use Your Flashlight
- Illuminate dark code with proper debugging tools
- Don't debug in the dark

## Rule 2: Follow the Trail
- Track stack traces like ghost footprints
- Every error leaves a trace
EOF

    cat > rules/mansion-maintenance.md << 'EOF'
# Mansion Maintenance

*Keep your codebase mansion clean.*

## Regular Cleanup
- Remove unused code like dusting furniture
- Update dependencies regularly

## Security Checks
- Lock down APIs like securing mansion doors
- Regular security audits
EOF

    cat > rules/advanced/boss-battles.md << 'EOF'
# Boss Battle Strategies

*Advanced techniques for complex problems.*

## Strategy 1: Preparation
- Study the problem before attacking
- Gather the right tools

## Strategy 2: Persistence
- Don't give up on the first try
- Learn from each attempt
EOF

    cat > cursor/its-a-me.md << 'EOF'
# Cursor Integration Rules

*Basic cursor configuration guidelines.*

## Setup Rules
- Configure cursor for optimal performance
- Use consistent settings across team

## Usage Guidelines
- Follow cursor best practices
- Regular updates and maintenance
EOF

    cat > amazon-q/luigi-assistant.md << 'EOF'
# Amazon Q Assistant Rules

*Basic AI assistant guidelines.*

## Interaction Rules
- Be specific in your questions
- Provide context for better answers

## Code Generation
- Review AI-generated code carefully
- Test before implementing
EOF

    cat > config.json << 'EOF'
{
  "version": "1.0.0",
  "features": ["basic-rules"],
  "settings": {
    "ghostDetection": "enabled",
    "debugging": "basic"
  }
}
EOF
}

create_version_1_1_0() {
    # Enhance existing files with advanced techniques
    cat >> ghost-hunting.md << 'EOF'

## Advanced Techniques (v1.1.0)

### Rule 3: Team Coordination
- Hunt ghosts in pairs when possible
- Share findings with the team

### Rule 4: Pattern Recognition
- Learn common ghost behaviors
- Develop hunting strategies
EOF

    cat >> cursor/its-a-me.md << 'EOF'

## Advanced Cursor Features (v1.1.0)

### Custom Shortcuts
- Set up project-specific shortcuts
- Optimize workflow with macros

### Integration Setup
- Connect with version control
- Configure linting and formatting
EOF

    # Update config
    cat > config.json << 'EOF'
{
  "version": "1.1.0",
  "features": ["basic-rules", "advanced-techniques"],
  "settings": {
    "ghostDetection": "enhanced",
    "debugging": "advanced",
    "teamCoordination": "enabled"
  }
}
EOF
}

create_version_1_2_0() {
    # Add best practices and more configuration
    cat >> rules/mansion-maintenance.md << 'EOF'

## Best Practices (v1.2.0)

### Code Quality
- Implement automated testing
- Use code review processes

### Performance Optimization
- Profile code regularly
- Optimize critical paths
EOF

    cat >> amazon-q/luigi-assistant.md << 'EOF'

## Enhanced AI Workflows (v1.2.0)

### Prompt Engineering
- Craft effective prompts for better results
- Use context and examples

### Code Review Integration
- Use AI for code review assistance
- Combine human and AI insights
EOF

    # Update config with more options
    cat > config.json << 'EOF'
{
  "version": "1.2.0",
  "features": ["basic-rules", "advanced-techniques", "best-practices"],
  "settings": {
    "ghostDetection": "enhanced",
    "debugging": "advanced",
    "teamCoordination": "enabled",
    "codeQuality": "strict",
    "performance": "optimized"
  },
  "integrations": {
    "cursor": "enabled",
    "amazonq": "enabled"
  }
}
EOF
}

create_version_2_0_0() {
    # Breaking changes: restructure and rename files
    rm -rf rules cursor amazon-q
    mkdir -p guidelines tools ai-assistants

    # Move and rename files with breaking changes
    cat > ghost-detection.md << 'EOF'
# Ghost Detection System (v2.0.0)

*BREAKING CHANGE: Renamed from ghost-hunting.md*

## Detection Methods
- Automated ghost detection
- Real-time monitoring
- Alert systems

## Advanced Techniques
- Machine learning detection
- Pattern analysis
- Predictive modeling
EOF

    cat > guidelines/maintenance.md << 'EOF'
# Maintenance Guidelines (v2.0.0)

*BREAKING CHANGE: Moved from rules/mansion-maintenance.md*

## Automated Maintenance
- Scheduled cleanup tasks
- Dependency updates
- Security patches

## Monitoring
- Health checks
- Performance metrics
- Error tracking
EOF

    cat > guidelines/expert-strategies.md << 'EOF'
# Expert Strategies (v2.0.0)

*BREAKING CHANGE: Renamed from rules/advanced/boss-battles.md*

## Master-Level Techniques
- Complex problem solving
- System architecture
- Performance optimization

## Leadership Skills
- Team coordination
- Knowledge sharing
- Mentoring
EOF

    cat > tools/cursor-pro.md << 'EOF'
# Cursor Pro Configuration (v2.0.0)

*BREAKING CHANGE: Renamed from cursor/its-a-me.md*

## Professional Setup
- Enterprise configurations
- Team synchronization
- Advanced workflows

## Productivity Features
- Custom extensions
- Automation scripts
- Integration pipelines
EOF

    cat > ai-assistants/q-developer.md << 'EOF'
# Q Developer Integration (v2.0.0)

*BREAKING CHANGE: Renamed from amazon-q/luigi-assistant.md*

## Professional AI Workflows
- Enterprise AI integration
- Advanced prompt strategies
- Code generation pipelines

## Quality Assurance
- AI-assisted testing
- Automated reviews
- Continuous improvement
EOF

    # Breaking config changes
    cat > settings.json << 'EOF'
{
  "version": "2.0.0",
  "breaking_changes": [
    "Renamed ghost-hunting.md to ghost-detection.md",
    "Moved rules/ to guidelines/",
    "Renamed cursor/ to tools/",
    "Renamed amazon-q/ to ai-assistants/",
    "Renamed config.json to settings.json"
  ],
  "features": ["automated-detection", "professional-workflows", "enterprise-integration"],
  "settings": {
    "detectionSystem": "ml-powered",
    "automation": "full",
    "integration": "enterprise"
  }
}
EOF
}

check_dependencies() {
    log "Checking dependencies..."

    if ! command -v gh &> /dev/null; then
        error "GitHub CLI (gh) not found!"
        echo "Please install it from: https://cli.github.com/"
        echo "Then run: gh auth login"
        return 1
    fi

    if ! command -v git &> /dev/null; then
        error "Git not found!"
        echo "Please install Git first."
        return 1
    fi

    # Check if authenticated
    if ! gh auth status &> /dev/null; then
        error "GitHub CLI not authenticated!"
        echo "Please run: gh auth login"
        return 1
    fi

    success "Dependencies check passed"
}

create_test_repo() {
    local repo_name="$1"
    local temp_dir="/tmp/arm-setup-$$"

    log "Checking if repository exists: $repo_name"

    # Check if repo already exists
    if gh repo view "$repo_name" &> /dev/null; then
        error "Repository $repo_name already exists!"
        echo "Please choose a different name or delete the existing repository."
        echo "To delete: gh repo delete $repo_name"
        return 1
    fi

    log "Creating test repository: $repo_name"

    mkdir -p "$temp_dir"
    cd "$temp_dir"

    git init

    # Create v1.0.0 - Basic content
    create_version_1_0_0
    git add .
    git commit -m "feat: initial ARM test repository with basic ghost hunting rules"
    git tag v1.0.0

    # Create v1.1.0 - Enhanced content
    create_version_1_1_0
    git add .
    git commit -m "feat: add advanced techniques and cursor integration"
    git tag v1.1.0

    # Create v1.2.0 - More improvements
    create_version_1_2_0
    git add .
    git commit -m "feat: enhance rules with best practices and config options"
    git tag v1.2.0

    # Create v2.0.0 - Breaking changes
    create_version_2_0_0
    git add .
    git commit -m "feat!: restructure repository with breaking changes"
    git tag v2.0.0

    # Create and push repository
    gh repo create "$repo_name" --public --source=. --remote=origin --push
    git push origin --tags

    cd /
    rm -rf "$temp_dir"

    success "Test repository created: https://github.com/$(gh api user --jq .login)/$repo_name"
}



main() {
    local repo_name="${1:-$DEFAULT_REPO}"

    # Check for help
    if [[ "$1" == "--help" || "$1" == "-h" ]]; then
        usage
        exit 0
    fi

    log "Setting up ARM test repository..."
    log "Repository: $repo_name"

    if ! check_dependencies; then
        exit 1
    fi

    create_test_repo "$repo_name"

    success "🎉 Test repository created successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Test your setup:"
    echo "   ./scripts/test/test-workflow.sh all \"https://github.com/\$(gh api user --jq .login)/$repo_name\""
    echo ""
    echo "2. Or run interactively:"
    echo "   ./scripts/test/test-workflow.sh all"
    echo ""
    echo "Your test repository is ready for action!"
}

main "$@"
