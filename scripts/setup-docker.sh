#!/usr/bin/env bash
# Grant current user access to Docker daemon socket.
# Run once, then log out and back in.

set -euo pipefail

echo "Adding user $USER to docker group..."
sudo usermod -aG docker "$USER"

echo ""
echo "✅ Done. Log out and back in (or run 'newgrp docker') for changes to take effect."
echo "Then verify with: docker ps"
