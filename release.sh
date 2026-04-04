#!/bin/bash
set -euo pipefail

# Usage: ./release.sh v0.1.0

if [ -z "${1:-}" ]; then
  echo "Usage: ./release.sh <version>"
  echo "Example: ./release.sh v0.1.0"
  exit 1
fi

VERSION="$1"

echo "==> Building release binaries for $VERSION..."
make clean
VERSION="$VERSION" make release

echo "==> Creating GitHub release $VERSION..."
gh release create "$VERSION" \
  --title "$VERSION" \
  --generate-notes \
  bin/llmvlt-linux-amd64 \
  bin/llmvlt-linux-arm64 \
  bin/llmvlt-macos-arm64 \
  bin/llmvlt-windows-amd64.exe

echo "==> Done. Release $VERSION published."
echo "    https://github.com/moronim/llmvlt/releases/tag/$VERSION"
