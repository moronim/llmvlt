#!/bin/sh
set -eu

REPO="moronim/llmvlt"
BINARY="llmvlt"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

info()  { printf '%s\n' "==> $*"; }
error() { printf '%s\n' "error: $*" >&2; exit 1; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || error "Missing required command: $1"
}

detect_platform() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)

  case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="macos" ;;
    *) error "Unsupported OS: $OS. Use 'go install github.com/${REPO}@latest'." ;;
  esac

  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
  esac

  case "${OS}-${ARCH}" in
    linux-amd64|linux-arm64|macos-arm64) ;;
    macos-amd64) error "macOS Intel is not supported. Build from source." ;;
    *) error "No prebuilt binary for ${OS}-${ARCH}. Use 'go install github.com/${REPO}@latest'." ;;
  esac
}

get_release_info() {
  API_URL="https://api.github.com/repos/${REPO}/releases/latest"
  RELEASE_JSON=$(curl -fsSL "$API_URL") || error "Could not fetch release metadata"
  VERSION=$(printf '%s\n' "$RELEASE_JSON" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)
  [ -n "$VERSION" ] || error "Could not determine latest version"
}

find_asset_url() {
  ASSET="${BINARY}-${OS}-${ARCH}"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
  # Verify the asset exists
  if ! curl -fsSL --head "$URL" >/dev/null 2>&1; then
    error "No release asset found for ${ASSET} in ${VERSION}"
  fi
}

install_binary() {
  TMP=$(mktemp -d) || error "Could not create temp directory"
  trap 'rm -rf "$TMP"' EXIT INT HUP TERM

  info "Downloading ${BINARY} ${VERSION} for ${OS}/${ARCH}"
  curl -fL "$URL" -o "${TMP}/${BINARY}" || error "Download failed"
  chmod 0755 "${TMP}/${BINARY}"

  if [ -w "$INSTALL_DIR" ]; then
    install -m 0755 "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}" 2>/dev/null \
      || cp "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  else
    command -v sudo >/dev/null 2>&1 || error "Need write access to ${INSTALL_DIR} or sudo"
    sudo mkdir -p "$INSTALL_DIR"
    sudo install -m 0755 "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}" 2>/dev/null \
      || {
        sudo cp "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
        sudo chmod 0755 "${INSTALL_DIR}/${BINARY}"
      }
  fi

  command -v "$BINARY" >/dev/null 2>&1 || info "Installed to ${INSTALL_DIR}/${BINARY} (not yet on PATH)"
  info "Installed ${BINARY} ${VERSION}"
}

need_cmd curl
need_cmd uname
need_cmd mktemp
detect_platform
get_release_info
find_asset_url
install_binary
