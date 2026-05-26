#!/bin/bash
set -e

# unicli - Universal CLI installer
# Usage: curl -fsSL https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.sh | bash
# Or with version: curl -fsSL https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.sh | bash -s -- v1.0.0

set -e

VERSION="${1:-latest}"
BINARY_NAME="unicli"
INSTALL_DIR=""
REPO="neko233-com/unicli"

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        CYGWIN*)    echo "windows" ;;
        MINGW*)     echo "windows" ;;
        MSYS*)      echo "windows" ;;
        *)         echo "unsupported" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64)     echo "amd64" ;;
        aarch64)    echo "arm64" ;;
        arm64)      echo "arm64" ;;
        amd64)      echo "amd64" ;;
        *)         echo "amd64" ;;
    esac
}

# Get latest version from GitHub API
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | \
        grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | tr -d 'v' || echo ""
}

# Install on Linux
install_linux() {
    INSTALL_DIR="/usr/local/bin"
    TARBALL="${BINARY_NAME}-linux-${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${TARBALL}"

    echo "Installing unicli v${VERSION} for Linux (${ARCH})..."

    TMPDIR=$(mktemp -d)
    cd "$TMPDIR"

    curl -fsSL "$DOWNLOAD_URL" -o "${TARBALL}"
    tar -xzf "${TARBALL}"
    rm -f "${TARBALL}"

    if [ -w "$INSTALL_DIR" ]; then
        mv -f "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv -f "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    cd /
    rm -rf "$TMPDIR"
}

# Install on macOS
install_darwin() {
    INSTALL_DIR="/usr/local/bin"
    TARBALL="${BINARY_NAME}-darwin-${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${TARBALL}"

    echo "Installing unicli v${VERSION} for macOS (${ARCH})..."

    TMPDIR=$(mktemp -d)
    cd "$TMPDIR"

    curl -fsSL "$DOWNLOAD_URL" -o "${TARBALL}"
    tar -xzf "${TARBALL}"
    rm -f "${TARBALL}"

    if [ -w "$INSTALL_DIR" ]; then
        mv -f "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv -f "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    cd /
    rm -rf "$TMPDIR"
}

# Install on Windows (using PowerShell)
install_windows() {
    echo "Downloading unicli for Windows..."

    local TARBALL="${BINARY_NAME}-windows-${ARCH}.zip"
    local DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${TARBALL}"
    local TEMP_DIR=$(mktemp -d)
    local INSTALL_DIR="${LOCALAPPDATA}\unicli"

    cd "$TEMP_DIR"
    curl -fsSL "$DOWNLOAD_URL" -o "${TARBALL}"

    powershell -Command "Expand-Archive -Path '${TARBALL}' -DestinationPath '${TEMP_DIR}' -Force"
    rm -f "${TARBALL}"

    mkdir -p "$INSTALL_DIR"
    mv -f "${TEMP_DIR}/unicli.exe" "${INSTALL_DIR}/unicli.exe" 2>/dev/null || \
    mv -f "${TEMP_DIR}/${BINARY_NAME}.exe" "${INSTALL_DIR}/unicli.exe"

    # Add to PATH
    local USER_PATH=$(powershell -Command "[Environment]::GetEnvironmentVariable('Path', 'User')")
    if [[ ! "$USER_PATH" == *"${INSTALL_DIR}"* ]]; then
        powershell -Command "[Environment]::SetEnvironmentVariable('Path', \"\${USER_PATH};${INSTALL_DIR}\", 'User')"
        export PATH="${INSTALL_DIR}:$PATH"
    fi

    cd /
    rm -rf "$TEMP_DIR"
}

# Main
main() {
    OS=$(detect_os)
    ARCH=$(detect_arch)

    if [ "$OS" = "unsupported" ]; then
        echo "Unsupported operating system."
        exit 1
    fi

    if [ "$VERSION" = "latest" ] || [ -z "$VERSION" ]; then
        VERSION=$(get_latest_version)
        if [ -z "$VERSION" ]; then
            VERSION="1.0.0"
        fi
    fi

    echo "Detected: ${OS}/${ARCH}"
    echo "Installing unicli v${VERSION}..."

    case "$OS" in
        linux)   install_linux ;;
        darwin)  install_darwin ;;
        windows) install_windows ;;
    esac

    echo ""
    echo "Installed successfully!"
    echo "Run 'unicli --help' to get started."
}

main "$@"