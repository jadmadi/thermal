#!/bin/sh
# install.sh — Zero-dependency standalone installer for Thermal
# Usage: curl -fsSL https://jadmadi.net/thermal/install.sh | sh
# Or with options:
#   sh install.sh --dry-run
#   sh install.sh --version 0.14.0
#   sh install.sh --prefix /opt/bin
set -eu

REPO="jadmadi/thermal"
GITHUB_URL="https://github.com/${REPO}"
GITHUB_API="https://api.github.com/repos/${REPO}/releases/latest"

DRY_RUN=0
CUSTOM_VERSION=""
CUSTOM_PREFIX=""

while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    --version)
      CUSTOM_VERSION="$2"
      shift 2
      ;;
    --prefix)
      CUSTOM_PREFIX="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: install.sh [--dry-run] [--version <v>] [--prefix <dir>]"
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

# Architecture detection
OS="$(uname -s)"
case "${OS}" in
  Linux|linux)
    OS="linux"
    ;;
  Darwin|darwin)
    OS="darwin"
    ;;
  *)
    echo "Error: Unsupported operating system: ${OS}" >&2
    echo "Thermal binaries are available for Linux, macOS (Darwin), and Windows." >&2
    exit 1
    ;;
esac

ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  i386|i686)
    ARCH="386"
    ;;
  *)
    echo "Error: Unsupported architecture: ${ARCH}" >&2
    exit 1
    ;;
esac

# Version resolution
VERSION="${CUSTOM_VERSION:-${VERSION:-}}"
if [ -z "${VERSION}" ]; then
  echo "Resolving latest Thermal release from GitHub..."
  if command -v curl >/dev/null 2>&1; then
    RELEASE_JSON="$(curl -fsSL "${GITHUB_API}" 2>/dev/null || true)"
  elif command -v wget >/dev/null 2>&1; then
    RELEASE_JSON="$(wget -qO- "${GITHUB_API}" 2>/dev/null || true)"
  else
    echo "Error: Neither curl nor wget was found in PATH." >&2
    exit 1
  fi

  if [ -n "${RELEASE_JSON}" ]; then
    VERSION="$(echo "${RELEASE_JSON}" | grep -m1 '"tag_name":' | sed -E 's/.*"tag_name":[[:space:]]*"v?([^"]+)".*/\1/')"
  fi

  if [ -z "${VERSION}" ]; then
    echo "Warning: Could not determine latest version from GitHub API; defaulting to v0.14.0" >&2
    VERSION="0.14.0"
  fi
fi

# Strip optional leading 'v'
VERSION="$(echo "${VERSION}" | sed 's/^v//')"

ARCHIVE="thermal_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="${GITHUB_URL}/releases/download/v${VERSION}/${ARCHIVE}"
CHECKSUM_URL="${GITHUB_URL}/releases/download/v${VERSION}/checksums.txt"

# Target installation directory
INSTALL_DIR="${CUSTOM_PREFIX:-}"
if [ -z "${INSTALL_DIR}" ]; then
  if [ "$(id -u)" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
  elif [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="${HOME}/.local/bin"
  fi
fi

echo "Thermal Installer (https://jadmadi.net/thermal/install.sh):"
echo "  Version:      v${VERSION}"
echo "  OS/Arch:      ${OS}/${ARCH}"
echo "  Target:       ${INSTALL_DIR}/thermal"
echo "  Archive:      ${ARCHIVE}"

if [ "${DRY_RUN}" -eq 1 ]; then
  echo "Dry-run mode enabled; exiting without making changes."
  exit 0
fi

# Create temporary directory
TMP_DIR="$(mktemp -d -t thermal-install-XXXXXX 2>/dev/null || mktemp -d /tmp/thermal-install-XXXXXX)"
cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT INT TERM

echo "Downloading ${ARCHIVE}..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${ARCHIVE}"
  curl -fsSL "${CHECKSUM_URL}" -o "${TMP_DIR}/checksums.txt" || true
elif command -v wget >/dev/null 2>&1; then
  wget -qO "${TMP_DIR}/${ARCHIVE}" "${DOWNLOAD_URL}"
  wget -qO "${TMP_DIR}/checksums.txt" "${CHECKSUM_URL}" || true
fi

# Checksum verification
if [ -f "${TMP_DIR}/checksums.txt" ] && grep -q "${ARCHIVE}" "${TMP_DIR}/checksums.txt"; then
  echo "Verifying SHA-256 checksum..."
  EXPECTED_SUM="$(grep "${ARCHIVE}" "${TMP_DIR}/checksums.txt" | awk '{print $1}')"
  ACTUAL_SUM=""

  if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_SUM="$(sha256sum "${TMP_DIR}/${ARCHIVE}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    ACTUAL_SUM="$(shasum -a 256 "${TMP_DIR}/${ARCHIVE}" | awk '{print $1}')"
  elif command -v openssl >/dev/null 2>&1; then
    ACTUAL_SUM="$(openssl dgst -sha256 "${TMP_DIR}/${ARCHIVE}" | awk '{print $NF}')"
  fi

  if [ -n "${ACTUAL_SUM}" ]; then
    if [ "${ACTUAL_SUM}" != "${EXPECTED_SUM}" ]; then
      echo "Error: Checksum mismatch for ${ARCHIVE}!" >&2
      echo "  Expected: ${EXPECTED_SUM}" >&2
      echo "  Actual:   ${ACTUAL_SUM}" >&2
      exit 1
    fi
    echo "  Checksum verified: ${ACTUAL_SUM}"
  else
    echo "Notice: No sha256 utility found; skipping checksum validation."
  fi
fi

# Extract binary
echo "Extracting binary..."
tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "${TMP_DIR}"

if [ ! -f "${TMP_DIR}/thermal" ]; then
  echo "Error: Binary 'thermal' not found in downloaded archive." >&2
  exit 1
fi

chmod 0755 "${TMP_DIR}/thermal"

# Install binary
mkdir -p "${INSTALL_DIR}"
if [ -w "${INSTALL_DIR}" ]; then
  mv "${TMP_DIR}/thermal" "${INSTALL_DIR}/thermal"
else
  echo "Destination ${INSTALL_DIR} is not writable; attempting sudo..."
  sudo mv "${TMP_DIR}/thermal" "${INSTALL_DIR}/thermal"
fi

echo "Successfully installed Thermal to ${INSTALL_DIR}/thermal"

# PATH check
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*)
    ;;
  *)
    echo "Notice: ${INSTALL_DIR} is not in your \$PATH."
    echo "Add it to your profile (e.g. ~/.bashrc or ~/.zshrc):"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac

"${INSTALL_DIR}/thermal" version || true
