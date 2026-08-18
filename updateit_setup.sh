#!/bin/bash
set -e
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"
URL="https://github.com/wxwreak/updateit/releases/latest/download/updateit-$OS-$ARCH"
echo "Installing $URL..."
mkdir -p "$HOME/.local/bin"
mkdir -p "$HOME/.local/share/updateit"
curl -L "$URL" -o "$HOME/.local/bin/updateit"
chmod +x "$HOME/.local/bin/updateit"
touch "$HOME/.local/share/updateit/latest.log"
echo "Updateit successfully installed!"
