#!/bin/bash

set -e

# Determine OS and ARCH
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# If the OS is Linux, convert it to "linux"
if [ "$OS" == "Linux" ]; then
    OS="linux"
fi

# If the ARCH is x86_64, convert it to "amd64"
if [ "$ARCH" == "x86_64" ]; then
    ARCH="amd64"
fi

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

echo "Downloading $BINARY_NAME $LATEST_RELEASE for $OS-$ARCH"

# Download the binary to the temporary directory
curl -sL "$DOWNLOAD_URL" -o "$TEMP_DIR/$BINARY_NAME"

# Make the binary executable
chmod +x "$TEMP_DIR/$BINARY_NAME"

# Copy the binary to /usr/local/bin
sudo cp "$TEMP_DIR/$BINARY_NAME" /usr/local/bin/

# Clean up the temporary directory
rm -rf "$TEMP_DIR"

echo "$BINARY_NAME installed successfully in /usr/local/bin"
