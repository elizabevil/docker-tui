# dev.just — Development environment setup

# Install CGo development headers required for Podman SDK build
install-cgo-deps:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Installing CGo dev headers for Podman SDK ==="
    sudo apt install -y pkg-config libgpgme-dev libbtrfs-dev
    echo "=== Done ==="

# Grant current user access to Docker daemon socket
setup-docker:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Adding user $USER to docker group..."
    sudo usermod -aG docker "$USER"
    echo "Done. Log out and back in, or run: newgrp docker"

# Enable and start Podman socket
start-socket:
    systemctl --user status podman.socket
    systemctl --user enable podman.socket

# Binary size analysis
size-analysis:
    GOEXPERIMENT=jsonv2 go install github.com/Zxilly/go-size-analyzer/cmd/gsa@latest
    gsa --web ./dist/docker-tui
    go install github.com/imjasonh/gobinsize@latest
    gobinsize -svg out.svg ./dist/docker-tui
