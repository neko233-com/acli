#!/bin/bash
set -e

# acli - macOS/Linux installer
# Usage: curl -fsSL https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.sh | bash
# Or:    curl -fsSL .../install.sh | bash -s -- v1.0.0
# Windows: use scripts/install.ps1 instead

VERSION="${1:-latest}"
BINARY_NAME="acli"
REPO="neko233-com/unicli"

detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        *)          echo "unsupported" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *)              echo "amd64" ;;
    esac
}

normalize_version() {
    local v="$1"
    v="${v#v}"
    v="${v#V}"
    echo "$v"
}

install_binary() {
    local os="$1"
    local arch="$2"
    local ver="$3"

    local asset="${BINARY_NAME}-${os}-${arch}"
    local url
    if [ "$ver" = "latest" ]; then
        url="https://github.com/${REPO}/releases/latest/download/${asset}"
    else
        url="https://github.com/${REPO}/releases/download/v${ver}/${asset}"
    fi
    local install_dir="/usr/local/bin"
    local target="${BINARY_NAME}"

    echo "Downloading ${url}..."
    TMPDIR=$(mktemp -d)
    curl -fsSL "$url" -o "${TMPDIR}/${target}"

    if [ -w "$install_dir" ]; then
        mv -f "${TMPDIR}/${target}" "${install_dir}/${target}"
    else
        sudo mv -f "${TMPDIR}/${target}" "${install_dir}/${target}"
    fi

    chmod +x "${install_dir}/${target}"
    rm -rf "$TMPDIR"

    echo "Installed to ${install_dir}/${target}"
    echo "Run: acli --help"
}

main() {
    OS=$(detect_os)
    ARCH=$(detect_arch)

    if [ "$OS" = "unsupported" ]; then
        echo "Unsupported operating system."
        echo "Windows users: run install.ps1 in PowerShell or CMD."
        echo "  irm https://raw.githubusercontent.com/${REPO}/main/scripts/install.ps1 | iex"
        exit 1
    fi

    if [ "$VERSION" != "latest" ] && [ -n "$VERSION" ]; then
        VERSION=$(normalize_version "$VERSION")
    else
        VERSION="latest"
    fi

    echo "Detected: ${OS}/${ARCH}"
    echo "Installing acli (${VERSION})..."

    install_binary "$OS" "$ARCH" "$VERSION"

    echo ""
    echo "Installed successfully!"
}

main "$@"
