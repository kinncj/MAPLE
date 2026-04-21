#!/usr/bin/env bash
set -euo pipefail

REPO="kinncj/AI-Squad"
INSTALL_DIR="${MAPLE_INSTALL_DIR:-$HOME/.tools/maple/bin}"

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

VERSION="${MAPLE_VERSION:-}"
if [ -z "$VERSION" ]; then
    VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
        | grep '"tag_name"' | head -1 | cut -d'"' -f4)
fi
[ -z "$VERSION" ] && { echo "Could not determine latest version. Set MAPLE_VERSION=vX.Y.Z"; exit 1; }

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
