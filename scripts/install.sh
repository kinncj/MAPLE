#!/usr/bin/env bash
set -euo pipefail

REPO="kinncj/MAPLE"
INSTALL_DIR="${MAPLE_INSTALL_DIR:-$HOME/.tools/maple/bin}"

# ── Parse arguments ────────────────────────────────────────────────────────────
VERSION="${MAPLE_VERSION:-}"
while [[ $# -gt 0 ]]; do
    case "$1" in
        --version|-v)
            VERSION="$2"
            shift 2
            ;;
        --version=*|-v=*)
            VERSION="${1#*=}"
            shift
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: install.sh [--version vX.Y.Z] [--install-dir PATH]"
            exit 1
            ;;
    esac
done

# ── Platform detection ─────────────────────────────────────────────────────────
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)        ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac
case "$OS" in
    linux|darwin) ;;
    *) echo "Unsupported OS: $OS — use install.ps1 on Windows"; exit 1 ;;
esac

# ── Resolve version ────────────────────────────────────────────────────────────
if [ -z "$VERSION" ]; then
    # Fetch up to 100 releases (GitHub API max per page), sort by semver, pick highest.
    # Avoids /releases/latest which sorts by created_at, not version number.
    VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases?per_page=100" \
        | grep '"tag_name"' \
        | cut -d'"' -f4 \
        | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' \
        | sort -V \
        | tail -1)
fi
[ -z "$VERSION" ] && { echo "Could not determine latest version. Set MAPLE_VERSION=vX.Y.Z or pass --version vX.Y.Z"; exit 1; }

# ── Download and install ───────────────────────────────────────────────────────
ARCHIVE="maple-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "Installing maple ${VERSION} (${OS}/${ARCH})"
echo "  → ${INSTALL_DIR}/maple"
echo ""

mkdir -p "$INSTALL_DIR"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/$ARCHIVE"
tar -xzf "$TMP/$ARCHIVE" -C "$TMP"
mv "$TMP/maple" "$INSTALL_DIR/maple"
chmod +x "$INSTALL_DIR/maple"

echo "✓ Installed maple ${VERSION}"
echo ""

if ! echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
    echo "Add to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
fi

echo "Verify with: maple --version"
