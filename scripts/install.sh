#!/usr/bin/env bash
set -euo pipefail

REPO="kinncj/AI-Squad"
INSTALL_DIR="${SQUAD_INSTALL_DIR:-$HOME/.tools/ai-squad/bin}"

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

VERSION="${SQUAD_VERSION:-}"
if [ -z "$VERSION" ]; then
    VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
        | grep '"tag_name"' | head -1 | cut -d'"' -f4)
fi
[ -z "$VERSION" ] && { echo "Could not determine latest version. Set SQUAD_VERSION=vX.Y.Z"; exit 1; }

ARCHIVE="squad-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "Installing squad ${VERSION} (${OS}/${ARCH})"
echo "  → ${INSTALL_DIR}/squad"
echo ""

mkdir -p "$INSTALL_DIR"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/$ARCHIVE"
tar -xzf "$TMP/$ARCHIVE" -C "$TMP"
mv "$TMP/squad" "$INSTALL_DIR/squad"
chmod +x "$INSTALL_DIR/squad"

echo "✓ Installed squad ${VERSION}"
echo ""

if ! echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
    echo "Add to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo "  export PATH=\"\$HOME/.tools/ai-squad/bin:\$PATH\""
    echo ""
fi

echo "Verify with: squad --version"
