#!/bin/bash

# astat installation script
# This script downloads the latest released binary from GitHub and installs it.

set -e

OWNER="sunil-saini"
REPO="astat"
BINARY_NAME="astat"

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "🚀 Installing $BINARY_NAME for $OS/$ARCH..."

# Get latest release tag
LATEST_TAG=$(curl -s "https://api.github.com/repos/$OWNER/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "❌ Could not find latest release tag."
    exit 1
fi

echo "📦 Latest version: $LATEST_TAG"

TRIMMED_LATEST_TAG=${LATEST_TAG#v}
URL="https://github.com/$OWNER/$REPO/releases/download/$LATEST_TAG/${BINARY_NAME}_${TRIMMED_LATEST_TAG}_${OS}_${ARCH}.tar.gz"

echo "Downloading from $URL..."
curl -L "$URL" -o "${BINARY_NAME}.tar.gz"
tar -xzf "${BINARY_NAME}.tar.gz"
chmod +x "$BINARY_NAME"

# Installation destination
DEST="/usr/local/bin/$BINARY_NAME"

echo "Installing to $DEST..."
if [ -w "/usr/local/bin" ]; then
    mv "$BINARY_NAME" "$DEST"
else
    sudo mv "$BINARY_NAME" "$DEST"
fi

echo "✅ $BINARY_NAME installed successfully to $DEST"
echo "✨ Run 'astat install' to complete the setup (autocomplete, etc.)"
