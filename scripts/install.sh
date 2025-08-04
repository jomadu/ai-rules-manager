#!/bin/bash
set -e

# ARM Installation Script
REPO="jomadu/ai-rules-manager"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="arm"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

detect_platform() {
    local os arch
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          log_error "Unsupported OS: $(uname -s)"; exit 1 ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac

    echo "${os}-${arch}"
}

get_latest_version() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//'
}

install_binary() {
    local platform="$1" version="$2"
    local binary_name="${BINARY_NAME}-${platform}"
    local download_url="https://github.com/${REPO}/releases/download/v${version}/${binary_name}.tar.gz"

    log_info "Downloading ARM v${version} for ${platform}..."

    local temp_dir=$(mktemp -d)
    cd "$temp_dir"

    curl -sL "$download_url" | tar -xz
    chmod +x "${binary_name}"

    if [ -w "$INSTALL_DIR" ]; then
        mv "${binary_name}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv "${binary_name}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    rm -rf "$temp_dir"
    log_info "ARM installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

main() {
    log_info "Installing ARM (AI Rules Manager)..."

    local platform=$(detect_platform)
    local version=$(get_latest_version)

    install_binary "$platform" "$version"

    if command -v "$BINARY_NAME" > /dev/null 2>&1; then
        log_info "ARM is ready! Run 'arm help' to get started"
    fi
}

main "$@"
