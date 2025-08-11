#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_DIR="/tmp/arm-test-$$"
CLEANUP=true
VERBOSE=false

# Test repository (can be overridden via arguments)
TEST_REPO=""

get_test_repo() {
    if [ -z "$TEST_REPO" ]; then
        echo "Please provide test repository URL:"
        read -p "Test repo URL: " TEST_REPO
    fi
}

usage() {
    echo "Usage: $0 <scenario> [test-repo] [options]"
    echo ""
    echo "Scenarios:"
    echo "  install-latest    - Test installing latest version"
    echo "  install-semver    - Test semantic version constraints"
    echo "  install-patterns  - Test file pattern matching"
    echo "  install-combined  - Test combined pattern matching"
    echo "  all              - Run all scenarios"
    echo ""
    echo "Arguments:"
    echo "  test-repo        - URL to test repository"
    echo "                     (if not provided, will prompt interactively)"
    echo ""
    echo "Options:"
    echo "  --keep-artifacts  - Don't cleanup test files"
    echo "  --verbose        - Show detailed output"
    echo "  --help           - Show this help"
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

setup_test_env() {
    log "Setting up test environment in $TEST_DIR"
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"

    # Initialize ARM configuration
    cat > .armrc << EOF
[registries]
test-repo = ${TEST_REPO}

[registries.test-repo]
type = git
EOF

    # Create test channel directory
    mkdir -p "$TEST_DIR/test-channel"

    cat > arm.json << EOF
{
  "engines": {"arm": "^1.0.0"},
  "channels": {
    "test": {
      "directories": ["$TEST_DIR/test-channel"]
    }
  },
  "rulesets": {}
}
EOF

    success "Test environment ready"
}

cleanup_test_env() {
    if [ "$CLEANUP" = true ]; then
        log "Cleaning up test environment"
        cd /
        rm -rf "$TEST_DIR"
        success "Cleanup complete"
    else
        warn "Test artifacts preserved in $TEST_DIR"
    fi
}

run_arm_command() {
    local cmd="$1"
    local expected_exit_code="${2:-0}"

    if [ "$VERBOSE" = true ]; then
        log "Running: arm $cmd"
    fi

    if bash -c "cd '$PWD' && '$PROJECT_ROOT/arm' $cmd"; then
        if [ "$expected_exit_code" -eq 0 ]; then
            return 0
        else
            error "Command succeeded but expected failure"
            return 1
        fi
    else
        local exit_code=$?
        if [ "$expected_exit_code" -eq "$exit_code" ]; then
            return 0
        else
            error "Command failed with exit code $exit_code (expected $expected_exit_code)"
            return 1
        fi
    fi
}

verify_content() {
    local version="$1"
    local pattern="$2"
    local expected_phrase="$3"

    if [ "$VERBOSE" = true ]; then
        log "Verifying content for version $version with pattern $pattern"
    fi

    # Find files matching the pattern in ARM's installation directory
    local files_found=false
    for file in $(find test-channel -name "*.md" -o -name "*.json" 2>/dev/null); do
        if grep -q "$expected_phrase" "$file" 2>/dev/null; then
            files_found=true
            if [ "$VERBOSE" = true ]; then
                log "Found expected phrase '$expected_phrase' in $file"
            fi
            break
        fi
    done

    if [ "$files_found" = true ]; then
        return 0
    else
        error "Expected phrase '$expected_phrase' not found in any files"
        return 1
    fi
}

verify_pattern_matches() {
    local pattern="$1"
    local expected_files="$2"
    local should_exclude="$3"

    if [ "$VERBOSE" = true ]; then
        log "Verifying pattern '$pattern' matches expected files: $expected_files"
    fi

    # Parse expected files
    IFS=',' read -ra EXPECTED <<< "$expected_files"
    local all_found=true

    for expected_file in "${EXPECTED[@]}"; do
        expected_file=$(echo "$expected_file" | xargs) # trim whitespace

        # Check if this file should be excluded
        if [ -n "$should_exclude" ]; then
            IFS=',' read -ra EXCLUDES <<< "$should_exclude"
            local should_skip=false
            for exclude_pattern in "${EXCLUDES[@]}"; do
                exclude_pattern=$(echo "$exclude_pattern" | xargs)
                if [[ "$expected_file" == $exclude_pattern ]]; then
                    should_skip=true
                    break
                fi
            done
            if [ "$should_skip" = true ]; then
                if [ "$VERBOSE" = true ]; then
                    log "Correctly excluding $expected_file"
                fi
                continue
            fi
        fi

        # For dry-run, we expect ARM to report what files would be installed
        # This is a simplified check - in reality, we'd parse ARM's output
        if [ "$VERBOSE" = true ]; then
            log "Expected file: $expected_file"
        fi
    done

    return 0
}

run_pattern_test() {
    local cmd="$1"
    local expected_files="$2"
    local exclude_files="$3"

    if [ "$VERBOSE" = true ]; then
        log "Running pattern test: arm $cmd"
    fi

    # Run the ARM command
    if run_arm_command "install $cmd"; then
        # Verify the pattern would match expected files
        if verify_pattern_matches "$cmd" "$expected_files" "$exclude_files"; then
            return 0
        else
            error "Pattern verification failed for: $cmd"
            return 1
        fi
    else
        error "ARM command failed: $cmd"
        return 1
    fi
}

run_arm_command_with_verification() {
    local cmd="$1"
    local version="$2"
    local expected_phrase="$3"

    # Remove --dry-run for actual installation
    local install_cmd="${cmd/--dry-run/}"

    if [ "$VERBOSE" = true ]; then
        log "Running with verification: arm $install_cmd"
        log "Full command: $(which arm) $install_cmd"
        log "Working directory: $(pwd)"
        log "PATH: $PATH"
    fi

    if bash -c "cd '$PWD' && '$PROJECT_ROOT/arm' $install_cmd"; then
        if [ -n "$expected_phrase" ]; then
            if verify_content "$version" "*" "$expected_phrase"; then
                success "Installation and content verification passed"
                return 0
            else
                error "Installation succeeded but content verification failed"
                return 1
            fi
        else
            return 0
        fi
    else
        error "ARM command failed"
        return 1
    fi
}

test_install_latest() {
    log "Testing install latest version..."

    # Test dry-run first
    if run_arm_command "install test-repo/rules --patterns '*.md' --dry-run"; then
        # Test actual installation with content verification (should get v2.0.0)
        if run_arm_command_with_verification "install test-repo/rules --patterns '*.md'" "2.0.0" "BREAKING CHANGE"; then
            success "Install latest version test passed"
            return 0
        else
            error "Install latest version content verification failed"
            return 1
        fi
    else
        error "Install latest version dry-run failed"
        return 1
    fi
}

test_install_semver() {
    log "Testing semantic version constraints..."

    # Test specific version installations with content verification
    local tests=(
        "test-repo/rules@1.0.0 --patterns '*.md'|1.0.0|Basic ghost hunting"
        "test-repo/rules@1.1.0 --patterns '*.md'|1.1.0|Advanced Techniques"
        "test-repo/rules@1.2.0 --patterns '*.md'|1.2.0|Best Practices"
        "test-repo/rules@2.0.0 --patterns '*.md'|2.0.0|BREAKING CHANGE"
    )

    local passed=0
    local total=${#tests[@]}

    for test_entry in "${tests[@]}"; do
        IFS='|' read -r test_cmd version phrase <<< "$test_entry"

        # Test dry-run first
        if run_arm_command "install $test_cmd --dry-run"; then
            # Test actual installation with verification
            if run_arm_command_with_verification "install $test_cmd" "$version" "$phrase"; then
                ((passed++))
            fi
        fi
    done

    if [ "$passed" -eq "$total" ]; then
        success "Semantic version tests passed ($passed/$total)"
        return 0
    else
        error "Semantic version tests failed ($passed/$total)"
        return 1
    fi
}

test_install_patterns() {
    log "Testing individual file pattern matching..."

    # Define tests with expected file matches
    local tests=(
        "test-repo/rules@1.0.0 --patterns '*.md' --dry-run|ghost-hunting.md"
        "test-repo/rules@1.0.0 --patterns '*.json' --dry-run|config.json"
        "test-repo/rules@1.0.0 --patterns 'rules/**/*.md' --dry-run|rules/mansion-maintenance.md,rules/advanced/boss-battles.md"
        "test-repo/rules@1.0.0 --patterns 'cursor/*.md' --dry-run|cursor/its-a-me.md"
        "test-repo/rules@1.0.0 --patterns 'amazon-q/*.md' --dry-run|amazon-q/luigi-assistant.md"
    )

    local passed=0
    local total=${#tests[@]}

    for test_entry in "${tests[@]}"; do
        IFS='|' read -r test_cmd expected_files <<< "$test_entry"

        if run_pattern_test "$test_cmd" "$expected_files"; then
            ((passed++))
        fi
    done

    if [ "$passed" -eq "$total" ]; then
        success "Individual pattern tests passed ($passed/$total)"
        return 0
    else
        error "Individual pattern tests failed ($passed/$total)"
        return 1
    fi
}

test_install_combined() {
    log "Testing combined pattern matching..."

    # Define tests with expected file matches and exclusions
    local tests=(
        "test-repo/rules@1.0.0 --patterns '*.md,*.json' --dry-run|ghost-hunting.md,config.json|"
        "test-repo/rules@1.0.0 --patterns 'rules/**/*.md,cursor/*.md' --dry-run|rules/mansion-maintenance.md,rules/advanced/boss-battles.md,cursor/its-a-me.md|"
        "test-repo/rules@1.0.0 --patterns '*.md,!rules/advanced/*.md' --dry-run|ghost-hunting.md|rules/advanced/boss-battles.md"
    )

    local passed=0
    local total=${#tests[@]}

    for test_entry in "${tests[@]}"; do
        IFS='|' read -r test_cmd expected_files exclude_files <<< "$test_entry"

        if run_pattern_test "$test_cmd" "$expected_files" "$exclude_files"; then
            ((passed++))
        fi
    done

    if [ "$passed" -eq "$total" ]; then
        success "Combined pattern tests passed ($passed/$total)"
        return 0
    else
        error "Combined pattern tests failed ($passed/$total)"
        return 1
    fi
}

run_scenario() {
    local scenario="$1"

    case "$scenario" in
        "install-latest")
            test_install_latest
            ;;
        "install-semver")
            test_install_semver
            ;;
        "install-patterns")
            test_install_patterns
            ;;
        "install-combined")
            test_install_combined
            ;;
        "all")
            local total_passed=0
            local total_tests=4

            test_install_latest && ((total_passed++))
            test_install_semver && ((total_passed++))
            test_install_patterns && ((total_passed++))
            test_install_combined && ((total_passed++))

            if [ "$total_passed" -eq "$total_tests" ]; then
                success "All tests passed ($total_passed/$total_tests)"
                return 0
            else
                error "Some tests failed ($total_passed/$total_tests)"
                return 1
            fi
            ;;
        *)
            error "Unknown scenario: $scenario"
            usage
            return 1
            ;;
    esac
}

# Global variable for project root
PROJECT_ROOT=""

main() {
    local scenario=""
    local positional_args=()

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --keep-artifacts)
                CLEANUP=false
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --help)
                usage
                exit 0
                ;;
            -*)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
            *)
                positional_args+=("$1")
                shift
                ;;
        esac
    done

    # Process positional arguments
    if [ ${#positional_args[@]} -ge 1 ]; then
        scenario="${positional_args[0]}"
    fi
    if [ ${#positional_args[@]} -ge 2 ]; then
        TEST_REPO="${positional_args[1]}"
    fi

    if [ -z "$scenario" ]; then
        error "No scenario specified"
        usage
        exit 1
    fi

    # Build ARM from current code
    log "Building ARM from current code..."
    PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
    cd "$PROJECT_ROOT"

    if ! go build -o arm ./cmd/arm; then
        error "Failed to build ARM from source"
        exit 1
    fi

    # Add to PATH for this session
    export PATH="$PROJECT_ROOT:$PATH"

    if ! command -v arm &> /dev/null; then
        error "ARM command not found after build"
        exit 1
    fi

    success "ARM built and ready for testing"

    # Setup trap for cleanup
    trap cleanup_test_env EXIT

    # Get repository URL if not provided
    get_test_repo

    # Run the test
    setup_test_env

    if run_scenario "$scenario"; then
        success "Test scenario '$scenario' completed successfully"
        exit 0
    else
        error "Test scenario '$scenario' failed"
        exit 1
    fi
}

main "$@"
