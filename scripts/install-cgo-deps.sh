#!/usr/bin/env bash
# Install CGo development headers required for Podman SDK build.
# Run this once before: just build-sdk

set -euo pipefail

echo "=== Installing CGo dev headers for Podman SDK ==="

sudo apt install -y pkg-config

# libgpgme-dev — GPGME bindings (needed by containers/image)

sudo apt install -y libgpgme-dev

# libbtrfs-dev — btrfs storage driver (needed by containers/storage)
sudo apt install -y libbtrfs-dev

echo "=== CGo headers installed ==="
echo ""
echo "Now you can build with:"
echo "  just build-sdk"
echo ""
echo "Or manually:"
echo "  CGO_ENABLED=1 CGO_CFLAGS=\"-I$HOME/.local/include\" go build -tags podmansdk -o ./dist/docker-tui ./cmd/docker-tui"
