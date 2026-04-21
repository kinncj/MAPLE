#!/usr/bin/env bash
set -euo pipefail

REPO="kinncj/MAPLE"
RTK_REPO="rtk-ai/rtk"
INSTALL_DIR="${MAPLE_INSTALL_DIR:-$HOME/.tools/maple/bin}"
SKIP_RTK="${SKIP_RTK:-}"

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
        --skip-rtk)
            SKIP_RTK=1
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: install.sh [--version vX.Y.Z] [--install-dir PATH] [--skip-rtk]"
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

# ── Resolve maple version ──────────────────────────────────────────────────────
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

# ── Download and install maple ─────────────────────────────────────────────────
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

# ── Install RTK token optimizer ────────────────────────────────────────────────
# RTK intercepts Bash tool calls and compresses output 60-90% before it hits the
# LLM context window. Single Rust binary, zero runtime dependencies.
# Skip with: --skip-rtk or SKIP_RTK=1
if [ -z "$SKIP_RTK" ]; then
    if command -v rtk >/dev/null 2>&1; then
        echo "✓ rtk already installed"
    else
        echo "Installing rtk token optimizer…"
        RTK_VERSION=$(curl -fsSL "https://api.github.com/repos/$RTK_REPO/releases?per_page=1" \
            | grep '"tag_name"' | head -1 | cut -d'"' -f4)
        if [ -n "$RTK_VERSION" ]; then
            case "${OS}-${ARCH}" in
                darwin-arm64) RTK_TRIPLE="aarch64-apple-darwin" ;;
                darwin-amd64) RTK_TRIPLE="x86_64-apple-darwin" ;;
                linux-arm64)  RTK_TRIPLE="aarch64-unknown-linux-gnu" ;;
                linux-amd64)  RTK_TRIPLE="x86_64-unknown-linux-gnu" ;;
                *)            RTK_TRIPLE="" ;;
            esac
            if [ -n "$RTK_TRIPLE" ]; then
                RTK_URL="https://github.com/${RTK_REPO}/releases/download/${RTK_VERSION}/rtk-${RTK_TRIPLE}.tar.gz"
                if curl -fsSL "$RTK_URL" -o "$TMP/rtk.tar.gz" 2>/dev/null; then
                    tar -xzf "$TMP/rtk.tar.gz" -C "$TMP"
                    mv "$TMP/rtk" "$INSTALL_DIR/rtk"
                    chmod +x "$INSTALL_DIR/rtk"
                    echo "✓ Installed rtk ${RTK_VERSION}"
                else
                    echo "~ rtk download failed — install manually: https://github.com/rtk-ai/rtk"
                fi
            else
                echo "~ rtk: unsupported platform ${OS}/${ARCH}"
            fi
        else
            echo "~ rtk: could not resolve latest version"
        fi
    fi
    echo ""
fi

# ── PATH reminder ──────────────────────────────────────────────────────────────
if ! echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
    echo "Add to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
fi

echo "Verify with: maple --version"
