#!/bin/sh
set -e

REPO="jadmadi/thermal"
BINARY="thermal"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *) echo "error: unsupported OS: $OS" >&2; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "error: unsupported arch: $ARCH" >&2; exit 1 ;;
esac

# Get latest version from GitHub API
VERSION=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "error: could not determine latest version" >&2
  exit 1
fi

ARCHIVE="${BINARY}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/${VERSION}/${ARCHIVE}"

echo "Installing $BINARY $VERSION for $OS/$ARCH..."

# Download and install
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$INSTALL_DIR"

curl -sL "$URL" | tar xz -C "$INSTALL_DIR" "$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"

echo "Installed $BINARY to $INSTALL_DIR/$BINARY"

# Check if in PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Note: add $INSTALL_DIR to your PATH" ;;
esac
