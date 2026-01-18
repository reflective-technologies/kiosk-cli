#!/bin/sh
set -e

# Kiosk CLI installer
# Usage: curl -fsSL https://raw.githubusercontent.com/reflective-technologies/kiosk-cli/main/install.sh | sh
#
# Note: Requires the GitHub repo to be public for unauthenticated downloads

REPO="reflective-technologies/kiosk-cli"
INSTALL_DIR="${KIOSK_INSTALL_DIR:-$HOME/.local/bin}"

main() {
    detect_platform
    check_dependencies

    echo "Installing kiosk CLI..."
    echo "  Platform: ${OS}_${ARCH}"
    echo "  Install dir: ${INSTALL_DIR}"
    echo ""

    fetch_latest_version
    download_and_install
    verify_installation

    echo ""
    echo "kiosk ${VERSION} installed successfully!"
    check_path
}

detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"

    case "${OS}" in
        darwin|linux) ;;
        *)
            echo "Error: Unsupported operating system: ${OS}"
            exit 1
            ;;
    esac

    case "${ARCH}" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo "Error: Unsupported architecture: ${ARCH}"
            exit 1
            ;;
    esac
}

check_dependencies() {
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        echo "Error: curl or wget is required"
        exit 1
    fi

    if ! command -v tar >/dev/null 2>&1; then
        echo "Error: tar is required"
        exit 1
    fi
}

fetch_latest_version() {
    echo "Fetching latest version..."

    RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"

    if command -v curl >/dev/null 2>&1; then
        RELEASE_INFO=$(curl -fsSL "${RELEASE_URL}")
    else
        RELEASE_INFO=$(wget -qO- "${RELEASE_URL}")
    fi

    VERSION=$(echo "${RELEASE_INFO}" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "${VERSION}" ]; then
        echo "Error: Could not determine latest version"
        exit 1
    fi

    # Remove 'v' prefix for asset naming
    VERSION_NUM="${VERSION#v}"

    echo "  Latest version: ${VERSION}"
}

download_and_install() {
    ASSET_NAME="kiosk_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET_NAME}"

    TMPDIR=$(mktemp -d)
    trap 'rm -rf "${TMPDIR}"' EXIT

    echo "Downloading ${ASSET_NAME}..."

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "${DOWNLOAD_URL}" -o "${TMPDIR}/${ASSET_NAME}"
    else
        wget -q "${DOWNLOAD_URL}" -O "${TMPDIR}/${ASSET_NAME}"
    fi

    echo "Extracting..."
    tar -xzf "${TMPDIR}/${ASSET_NAME}" -C "${TMPDIR}"

    mkdir -p "${INSTALL_DIR}"
    mv "${TMPDIR}/kiosk" "${INSTALL_DIR}/kiosk"
    chmod +x "${INSTALL_DIR}/kiosk"
}

verify_installation() {
    if [ ! -x "${INSTALL_DIR}/kiosk" ]; then
        echo "Error: Installation failed"
        exit 1
    fi
}

check_path() {
    case ":${PATH}:" in
        *":${INSTALL_DIR}:"*)
            # Already in PATH
            ;;
        *)
            echo "Note: ${INSTALL_DIR} is not in your PATH."
            echo ""
            echo "Add it to your shell profile:"
            echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
            ;;
    esac
}

main "$@"
