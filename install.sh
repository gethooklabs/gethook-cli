#!/bin/sh
# GetHook CLI installer
# Usage: curl -fsSL https://cli.gethook.dev/install.sh | sh
set -e

REPO="gethook/gethook-cli"
BINARY="gethook"
INSTALL_DIR="${GETHOOK_INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux|darwin) ;;
  *)
    echo "Unsupported OS: $OS — download manually from https://github.com/$REPO/releases"
    exit 1
    ;;
esac

# Get latest release version
echo "Fetching latest GetHook CLI release..."
VERSION="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' \
  | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')"

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version. Check https://github.com/$REPO/releases"
  exit 1
fi

echo "Installing GetHook CLI $VERSION for $OS/$ARCH..."

TARBALL="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$TARBALL"
TMP_DIR="$(mktemp -d)"

curl -fsSL "$URL" -o "$TMP_DIR/$TARBALL"
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"

# Install binary
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
  echo "Installing to $INSTALL_DIR (may require sudo)..."
  sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"
rm -rf "$TMP_DIR"

echo ""
echo "✓ GetHook CLI $VERSION installed to $INSTALL_DIR/$BINARY"
echo ""
echo "Get started:"
echo "  gethook login"
echo "  gethook listen --forward-to http://localhost:3000/webhooks"
echo ""
echo "Docs: https://docs.gethook.dev/cli"
