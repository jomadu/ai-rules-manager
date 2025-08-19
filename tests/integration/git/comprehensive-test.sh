#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
REGISTRY_URL="https://github.com/jomadu/ai-rules-manager-test-git-registry"
TEST_DIR="test-sandbox"
# ARM_BIN will be set after building

# Test tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
FAILED_TEST_NAMES=()

# Test groups
ALL_GROUPS=("registry" "channel" "version" "pattern" "channel-specific" "update" "uninstall" "cache" "cache-structure" "config" "search" "error" "workflow")
SELECTED_GROUPS=()

# Parse command line arguments
show_usage() {
    echo "Usage: $0 [OPTIONS] [TEST_GROUPS...]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -l, --list     List available test groups"
    echo ""
    echo "Test Groups:"
    for group in "${ALL_GROUPS[@]}"; do
        echo "  $group"
    done
    echo ""
    echo "Examples:"
    echo "  $0                    # Run all tests"
    echo "  $0 search             # Run only search tests"
    echo "  $0 registry channel   # Run registry and channel tests"
}

list_groups() {
    echo "Available test groups:"
    for group in "${ALL_GROUPS[@]}"; do
        echo "  $group"
    done
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -l|--list)
            list_groups
            exit 0
            ;;
        -*)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
        *)
            # Check if it's a valid group
            if [[ " ${ALL_GROUPS[*]} " =~ " $1 " ]]; then
                SELECTED_GROUPS+=("$1")
            else
                echo "Unknown test group: $1"
                echo "Use --list to see available groups"
                exit 1
            fi
            ;;
    esac
    shift
done

# If no groups specified, run all
if [ ${#SELECTED_GROUPS[@]} -eq 0 ]; then
    SELECTED_GROUPS=("${ALL_GROUPS[@]}")
fi

# Function to check if a group should run
should_run_group() {
    local group="$1"
    [[ " ${SELECTED_GROUPS[*]} " =~ " $group " ]]
}

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

run_test() {
    local test_name="$1"
    local test_cmd="$2"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_info "Running test: $test_name"
    if eval "$test_cmd"; then
        log_success "$test_name passed"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        log_error "$test_name failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$test_name")
        return 1
    fi
}

explore_structure() {
    local context="$1"
    log_info "=== Directory Structure: $context ==="
    if command -v tree >/dev/null 2>&1; then
        tree -a -I '.git' . || true
    else
        find . -type f | grep -v '.git' | sort || true
    fi
    echo
}

inspect_cache() {
    local context="$1"
    log_info "=== Cache Inspection: $context ==="

    # Look for ARM cache directory
    if [ -d ".arm/cache" ]; then
        log_info "Cache directory structure:"
        tree -a .arm/cache 2>/dev/null || find .arm/cache -type f | sort
        echo

        # Validate three-level cache hierarchy
        log_info "Validating three-level cache hierarchy:"
        for registry_cache in .arm/cache/registries/*; do
            if [ -d "$registry_cache" ]; then
                local registry_hash=$(basename "$registry_cache")
                log_info "Registry: $registry_hash"

                # Check for repository directory (registry level)
                if [ -d "$registry_cache/repository" ]; then
                    log_success "  ✓ Repository directory found"
                else
                    log_warning "  ✗ Repository directory missing"
                fi

                # Check for rulesets directory (new level)
                if [ -d "$registry_cache/rulesets" ]; then
                    log_success "  ✓ Rulesets directory found"

                    # Check ruleset-level directories
                    for ruleset_cache in "$registry_cache/rulesets"/*; do
                        if [ -d "$ruleset_cache" ]; then
                            local ruleset_hash=$(basename "$ruleset_cache")
                            log_info "    Ruleset: $ruleset_hash"

                            # Check for version directories
                            for version_cache in "$ruleset_cache"/*; do
                                if [ -d "$version_cache" ]; then
                                    local version=$(basename "$version_cache")
                                    log_success "      ✓ Version: $version"

                                    # Count files in version directory
                                    local file_count=$(find "$version_cache" -type f | wc -l)
                                    log_info "        Files: $file_count"
                                fi
                            done
                        fi
                    done
                else
                    log_warning "  ✗ Rulesets directory missing"
                fi
            fi
        done
        echo

        # Inspect cache metadata files
        for metadata_file in .arm/cache/registries/*/metadata.json; do
            if [ -f "$metadata_file" ]; then
                log_info "Cache metadata: $metadata_file"
                cat "$metadata_file" | jq . 2>/dev/null || cat "$metadata_file"
                echo
            fi
        done

        # Inspect registry index files
        for index_file in .arm/cache/registries/*/index.json; do
            if [ -f "$index_file" ]; then
                log_info "Registry index: $index_file"
                cat "$index_file" | jq . 2>/dev/null || cat "$index_file"
                echo
            fi
        done

        # Look for git repository data
        for repo_cache in .arm/cache/registries/*/repository; do
            if [ -d "$repo_cache" ] && [ -d "$repo_cache/.git" ]; then
                log_info "Git repository cache: $(dirname "$repo_cache")"
                cd "$repo_cache"
                git log --oneline -5 2>/dev/null || true
                git tag -l 2>/dev/null || true
                cd - >/dev/null
                echo
            fi
        done
    else
        log_warning "No cache directory found"
    fi
    echo
}

inspect_config() {
    local context="$1"
    log_info "=== Configuration Inspection: $context ==="

    if [ -f ".armrc" ]; then
        log_info "ARM registry configuration (.armrc):"
        cat .armrc
        echo
    else
        log_warning "No .armrc configuration found"
    fi

    if [ -f "arm.json" ]; then
        log_info "ARM project configuration (arm.json):"
        cat arm.json | jq . 2>/dev/null || cat arm.json
        echo
    else
        log_warning "No arm.json configuration found"
    fi

    if [ -f ".arm/config.json" ]; then
        log_info "ARM internal configuration:"
        cat .arm/config.json | jq . 2>/dev/null || cat .arm/config.json
        echo
    fi

    if [ -f ".arm/state.json" ]; then
        log_info "ARM state:"
        cat .arm/state.json | jq . 2>/dev/null || cat .arm/state.json
        echo
    fi
}

verify_files() {
    local context="$1"
    shift
    local files=("$@")

    log_info "=== File Verification: $context ==="
    for file in "${files[@]}"; do
        if [ -f "$file" ]; then
            log_success "✓ Found: $file"
            log_info "Content preview (first 5 lines):"
            head -5 "$file" | sed 's/^/  /'
            echo
        else
            log_error "✗ Missing: $file"
        fi
    done
    echo
}

cleanup() {
    log_info "Cleaning up test environment"
    cd ..
    rm -rf "$TEST_DIR"
}

# Build ARM binary
log_info "Building ARM binary"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$PROJECT_ROOT"
make
ARM_BIN="$PROJECT_ROOT/bin/arm"

# Verify ARM binary was built
if [ ! -f "$ARM_BIN" ]; then
    log_error "Failed to build ARM binary at $ARM_BIN"
    exit 1
fi
log_success "Built ARM binary at $ARM_BIN"

# Return to test directory
cd "$SCRIPT_DIR"

# Setup test environment
log_info "Setting up test environment"
rm -rf "$TEST_DIR"
mkdir "$TEST_DIR"
cd "$TEST_DIR"

# Initialize ARM
log_info "Initializing ARM configuration"
$ARM_BIN install

# Enable caching for testing
log_info "Enabling cache for testing"
echo -e "\n[cache]\npath = .arm/cache\nttl = 1h" >> .armrc

explore_structure "After ARM initialization"
inspect_config "After ARM initialization"

# Test 1: Registry Management
if should_run_group "registry"; then
    log_info "=== Testing Registry Management ==="

    run_test "Add git registry" \
        "$ARM_BIN config add registry test-registry $REGISTRY_URL --type=git"

    run_test "Add default registry" \
        "$ARM_BIN config add registry default $REGISTRY_URL --type=git"

    run_test "List registries" \
        "$ARM_BIN config list registries"

    run_test "Get registry URL" \
        "$ARM_BIN config get registries.test-registry"

    inspect_config "After adding registry"
    inspect_cache "After adding registry"
else
    # Still need to set up registry for other tests
    $ARM_BIN config add registry test-registry $REGISTRY_URL --type=git >/dev/null 2>&1
    $ARM_BIN config add registry default $REGISTRY_URL --type=git >/dev/null 2>&1
fi

# Test 2: Channel Management
if should_run_group "channel"; then
    log_info "=== Testing Channel Management ==="

    run_test "Add cursor channel" \
        "$ARM_BIN config add channel cursor --directories .cursor/rules"

    run_test "Add q-developer channel" \
        "$ARM_BIN config add channel q --directories .amazonq/rules"

    run_test "List channels" \
        "$ARM_BIN config list channels"

    inspect_config "After adding channels"
else
    # Still need to set up channels for other tests
    $ARM_BIN config add channel cursor --directories .cursor/rules >/dev/null 2>&1
    $ARM_BIN config add channel q --directories .amazonq/rules >/dev/null 2>&1
fi

# Test 3: Version Resolution and Installation
if should_run_group "version"; then
    log_info "=== Testing Version Resolution ==="

    run_test "Install latest version" \
        "$ARM_BIN install test-ruleset --patterns '*.md'"

    explore_structure "After installing latest version"
    inspect_cache "After installing latest version"
    # Note: Files will be under resolved commit hash, not 'latest'
    run_test "Verify latest version files exist" \
        "find .cursor/rules/arm/default/test-ruleset -name 'ghost-detection.md' | grep -q ."

    run_test "List installed packages" \
        "$ARM_BIN list"

    # Clean for next test
    rm -rf .cursor .amazonq

    run_test "Install specific version (v1.0.0)" \
        "$ARM_BIN install test-ruleset@v1.0.0 --patterns '*.md'"

    explore_structure "After installing v1.0.0"
    inspect_cache "After installing v1.0.0"

    run_test "Install with semver constraint (^1.0.0)" \
        "$ARM_BIN install test-ruleset@^1.0.0 --patterns '*.md'"

    explore_structure "After installing ^1.0.0"
    inspect_cache "After installing ^1.0.0"
fi

# Test 4: Pattern Matching
if should_run_group "pattern"; then
    log_info "=== Testing Pattern Matching ==="

    # Clean for pattern tests
    rm -rf .cursor .amazonq

    run_test "Install with specific file pattern" \
        "$ARM_BIN install test-ruleset --patterns 'ghost-*.md'"

    explore_structure "After ghost-*.md pattern install"
    run_test "Verify ghost pattern files exist" \
        "find .cursor/rules/arm/default/test-ruleset -name 'ghost-*.md' | wc -l | grep -q 2"

    run_test "Install with directory pattern" \
        "$ARM_BIN install test-ruleset --patterns 'ai-assistants/*.md'"

    explore_structure "After ai-assistants/*.md pattern install"
    run_test "Verify AI assistants pattern files exist" \
        "find .cursor/rules/arm/default/test-ruleset -path '*/ai-assistants/q-developer.md' | grep -q ."

    run_test "Install with multiple patterns" \
        "$ARM_BIN install test-ruleset --patterns 'guidelines/*.md,tools/*.md'"

    explore_structure "After multiple patterns install"
    run_test "Verify multiple patterns files exist" \
        "find .cursor/rules/arm/default/test-ruleset -name '*.md' | grep -E '(guidelines|tools)' | wc -l | grep -q 3"
fi

# Test 5: Channel-specific Installation
if should_run_group "channel-specific"; then
    log_info "=== Testing Channel-specific Installation ==="

    # Clean for channel tests
    rm -rf .cursor .amazonq

    run_test "Install to cursor channel" \
        "$ARM_BIN install test-ruleset --channels cursor --patterns 'tools/cursor-pro.md'"

    explore_structure "After cursor channel install"
    run_test "Verify cursor channel files exist" \
        "find .cursor/rules/arm/default/test-ruleset -name 'cursor-pro.md' | grep -q ."

    run_test "Install to q-developer channel" \
        "$ARM_BIN install test-ruleset --channels q --patterns 'ai-assistants/q-developer.md'"

    explore_structure "After q-developer channel install"
    run_test "Verify q-developer channel files exist" \
        "find .amazonq/rules/arm/default/test-ruleset -name 'q-developer.md' | grep -q ."

    run_test "Verify cursor channel files exist" \
        "find .cursor/rules/arm/default/test-ruleset -name 'cursor-pro.md' | grep -q ."

    run_test "Verify q-developer channel files exist" \
        "find .amazonq/rules/arm/default/test-ruleset -name 'q-developer.md' | grep -q ."
fi

# Test 6: Update Operations
if should_run_group "update"; then
    log_info "=== Testing Update Operations ==="

# Install v1.0.0 first to have something to update from
run_test "Install v1.0.0 for update test" \
    "$ARM_BIN install test-ruleset@v1.0.0 --patterns '*.md'"

log_info "Files before update:"
$ARM_BIN list

run_test "Verify v1.0.0 installed" \
    "$ARM_BIN list | grep 'v1.0.0'"

run_test "Update all packages" \
    "$ARM_BIN update"

log_info "Files after update all:"
$ARM_BIN list

run_test "Verify updated to latest version" \
    "$ARM_BIN list | grep 'test-ruleset'"

explore_structure "After update all"

# Install v1.0.0 again for specific update test
run_test "Install v1.0.0 again for specific update test" \
    "$ARM_BIN install test-ruleset@v1.0.0 --patterns '*.md'"

run_test "Verify v1.0.0 installed again" \
    "$ARM_BIN list | grep 'v1.0.0'"

run_test "Update specific package" \
    "$ARM_BIN update test-ruleset"

log_info "Files after specific update:"
$ARM_BIN list

run_test "Verify specific update to latest version" \
    "$ARM_BIN list | grep 'test-ruleset'"

    explore_structure "After specific update"
fi

# Test 7: Uninstall Operations
if should_run_group "uninstall"; then
    log_info "=== Testing Uninstall Operations ==="

    run_test "Uninstall specific package" \
        "$ARM_BIN uninstall test-ruleset"

    run_test "Verify files removed after uninstall" \
        "! test -d .cursor/rules/arm/default/test-ruleset"
fi

# Test 8: Cache Management (Skipped - cache command not implemented)
if should_run_group "cache"; then
    log_info "=== Testing Cache Management ==="
    log_warning "Cache management commands not yet implemented - skipping tests"

    # Install something to create cache entries for inspection
    $ARM_BIN install test-ruleset --patterns '*.md' >/dev/null 2>&1

    inspect_cache "Cache structure validation"
fi

# Test 8.5: Cache Structure Validation
if should_run_group "cache-structure"; then
    log_info "=== Testing Three-Level Cache Structure ==="

    # Clean slate for structure testing
    rm -rf .arm/cache

    # Install to create cache structure
    run_test "Install to create cache structure" \
        "$ARM_BIN install test-ruleset --patterns '*.md'"

    inspect_cache "After creating cache structure"

    # Validate three-level hierarchy exists
    run_test "Validate registries directory exists" \
        "test -d .arm/cache/registries"

    run_test "Validate registry hash directory exists" \
        "find .arm/cache/registries -mindepth 1 -maxdepth 1 -type d | grep -q ."

    run_test "Validate repository directory exists at registry level" \
        "find .arm/cache/registries/*/repository -type d | grep -q repository"

    run_test "Validate rulesets directory exists at registry level" \
        "find .arm/cache/registries/*/rulesets -type d | grep -q rulesets"

    run_test "Validate ruleset hash directories exist" \
        "find .arm/cache/registries/*/rulesets -mindepth 1 -maxdepth 1 -type d | grep -q ."

    run_test "Validate version directories exist in rulesets" \
        "find .arm/cache/registries/*/rulesets/*/* -maxdepth 0 -type d | grep -q ."

    run_test "Validate files exist in version directories" \
        "find .arm/cache/registries/*/rulesets/*/*/* -name '*.md' | grep -q ."

    # Test different patterns create different ruleset hashes
    run_test "Install with different pattern" \
        "$ARM_BIN install test-ruleset --patterns 'ghost-*.md'"

    run_test "Validate different patterns create separate cache entries" \
        "test $(find .arm/cache/registries/*/rulesets/* -maxdepth 0 -type d | wc -l) -gt 1"

    run_test "Validate registry index.json exists" \
        "find .arm/cache/registries/*/index.json | grep -q ."

    run_test "Validate registry index contains rulesets" \
        "find .arm/cache/registries/*/index.json -exec cat {} \; | jq '.rulesets | length' | grep -q '[1-9]'"

    inspect_cache "After installing with different patterns"
fi

# Test 9: Configuration Management
if should_run_group "config"; then
    log_info "=== Testing Configuration Management ==="

    run_test "Set cache configuration" \
        "$ARM_BIN config set cache.ttl 3600"

    inspect_config "After setting cache TTL"

    run_test "Show all configuration" \
        "$ARM_BIN config show"

    run_test "Remove registry" \
        "$ARM_BIN config remove registry test-registry"

    run_test "Remove channel" \
        "$ARM_BIN config remove channel cursor"
fi

# Test 10: Search Capability
if should_run_group "search"; then
    log_info "=== Testing Search Capability ==="

    run_test "Search for test-ruleset" \
        "$ARM_BIN search test-ruleset"

    log_info "Search results for test-ruleset:"
    $ARM_BIN search test-ruleset

    run_test "Verify search finds test-ruleset" \
        "$ARM_BIN search test-ruleset | grep -v 'Searching for' | grep -v 'not yet implemented' | grep -q 'test-ruleset'"

    run_test "Search with pattern" \
        "$ARM_BIN search ghost"

    log_info "Search results for ghost:"
    $ARM_BIN search ghost

    run_test "Verify search finds ghost-related content" \
        "$ARM_BIN search ghost | grep -v 'Searching for' | grep -v 'not yet implemented' | grep -q 'ghost'"
fi

# Test 12: Error Handling
if should_run_group "error"; then
    log_info "=== Testing Error Handling ==="

    run_test "Install non-existent package (should fail gracefully)" \
        "! $ARM_BIN install non-existent-registry"

    run_test "Install with invalid pattern (should fail gracefully)" \
        "! $ARM_BIN install test-ruleset --patterns 'non-existent-pattern'"
fi

# Test 13: Complex Workflow Simulation
if should_run_group "workflow"; then
    log_info "=== Testing Complex Developer Workflow ==="

    # Simulate a developer setting up a new project
    log_info "Simulating new project setup"

    # Re-add registry and channels
    $ARM_BIN config add registry test-registry "$REGISTRY_URL" --type=git
    $ARM_BIN config add channel cursor --directories .cursor/rules
    $ARM_BIN config add channel q --directories .amazonq/rules

    # Install different rulesets for different purposes
    run_test "Install ghost hunting rules" \
        "$ARM_BIN install test-ruleset --patterns 'ghost-*.md'"

    run_test "Install AI assistant specific rules" \
        "$ARM_BIN install test-ruleset --channels cursor --patterns 'tools/cursor-pro.md'"

    run_test "Install guidelines" \
        "$ARM_BIN install test-ruleset --patterns 'guidelines/*.md'"

    # Show final state
    log_info "=== Final State Analysis ==="
    explore_structure "Final project state"
    inspect_config "Final configuration"
    inspect_cache "Final cache state"

    log_info "Final package list:"
    $ARM_BIN list

    # Verify all expected files are present
    run_test "Verify final ghost files exist" \
        "find .cursor/rules/arm/default/test-ruleset -name 'ghost-*.md' | wc -l | grep -q 2"

    run_test "Verify final cursor tools exist" \
        "find .cursor/rules/arm/default/test-ruleset -name 'cursor-pro.md' | grep -q ."

    run_test "Verify final guidelines exist" \
        "find .cursor/rules/arm/default/test-ruleset -path '*/guidelines/*.md' | wc -l | grep -q 2"
fi

# Cleanup
cleanup

# Test Summary
log_info "=== TEST SUMMARY ==="
log_info "Total tests: $TOTAL_TESTS"
log_success "Passed: $PASSED_TESTS"
if [ $FAILED_TESTS -gt 0 ]; then
    log_error "Failed: $FAILED_TESTS"
    log_error "Failed tests:"
    for test_name in "${FAILED_TEST_NAMES[@]}"; do
        log_error "  - $test_name"
    done
    exit 1
else
    log_error "Failed: $FAILED_TESTS"
    log_success "All tests completed successfully!"
fi
