#!/bin/bash

set -e

# Determine OS and ARCH
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map ARCH to Go ARCH
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Define GitHub repo and binary name
REPO="EdwinPhilip/gh-pr-commenter"
BINARY_NAME="ghpc"

# Fetch the latest release tag from GitHub API
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | jq -r .tag_name)
if [ "$LATEST_RELEASE" == "null" ]; then
    echo "Failed to fetch the latest release"
    exit 1
fi

# Construct download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$BINARY_NAME-$OS-$ARCH"

# Create a temporary directory
TEMP_DIR=$(mktemp -d)

# Download the binary to the temporary directory
curl -L "$DOWNLOAD_URL" -o "$TEMP_DIR/$BINARY_NAME"

# Make the binary executable
chmod +x "$TEMP_DIR/$BINARY_NAME"

# Copy the binary to /usr/local/bin
sudo cp "$TEMP_DIR/$BINARY_NAME" /usr/local/bin/

# Clean up the temporary directory
rm -rf "$TEMP_DIR"

echo "$BINARY_NAME installed successfully in /usr/local/bin"
