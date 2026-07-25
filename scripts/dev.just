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

# === Docker daemon (system-level systemd service) ============================

# Enable and start the Docker daemon
start-docker:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Starting docker.service ==="
    sudo systemctl enable --now docker.service
    sudo systemctl is-active docker.service
    echo "Socket: /var/run/docker.sock"

# Stop the Docker daemon
stop-docker:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Stopping docker.service ==="
    sudo systemctl stop docker.service

# Restart the Docker daemon
restart-docker:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Restarting docker.service ==="
    sudo systemctl restart docker.service
    sudo systemctl is-active docker.service

# Show Docker daemon status
status-docker:
    systemctl status docker.service --no-pager

# === Podman socket (user-level systemd service) ===============================

# Enable and start the Podman API socket
start-podman:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Starting podman.socket (user) ==="
    systemctl --user enable --now podman.socket
    SOCK="/run/user/$(id -u)/podman/podman.sock"
    # systemd may report 'active' before the socket file is materialised;
    # restart once if the file is missing, then verify it answers.
    if [[ ! -S "$SOCK" ]]; then
        echo "Socket file missing — restarting podman.socket to materialise it..."
        systemctl --user restart podman.socket
    fi
    systemctl --user is-active podman.socket
    if [[ -S "$SOCK" ]]; then
        echo "Socket: $SOCK (ready)"
    else
        echo "WARN: socket file still missing at $SOCK" >&2
    fi

# Stop the Podman socket
stop-podman:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Stopping podman.socket (user) ==="
    systemctl --user stop podman.socket

# Restart the Podman socket (also recovers from missing socket-file state)
restart-podman:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Restarting podman.socket (user) ==="
    systemctl --user restart podman.socket
    systemctl --user is-active podman.socket
    SOCK="/run/user/$(id -u)/podman/podman.sock"
    if [[ ! -S "$SOCK" ]]; then
        echo "Socket file missing after restart: $SOCK" >&2
        exit 1
    fi
    echo "Socket: $SOCK (ready)"

# Show Podman socket status
status-podman:
    systemctl --user status podman.socket --no-pager

# === Combined actions =========================================================

# Start both Docker and Podman services
start-services: start-docker start-podman
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Both services started ==="

# Stop both Docker and Podman services
stop-services: stop-docker stop-podman
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Both services stopped ==="

# Show status of both services
status-services:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "--- docker.service ---"
    systemctl status docker.service --no-pager || true
    echo
    echo "--- podman.socket (user) ---"
    systemctl --user status podman.socket --no-pager || true

# === Diagnostics =============================================================

# Smoke test: ping both daemons and report versions
ping-services:
    #!/usr/bin/env bash
    set -uo pipefail
    echo "--- docker ---"
    if ! command -v docker >/dev/null 2>&1; then
        echo "  docker not installed"
    elif out=$(docker version 2>&1); then
        echo "$out" | grep -iE 'version|server|client' | head -8 | sed 's/^/  /'
    else
        echo "  docker unreachable (permission denied? try: just dev::setup-docker)"
        echo "  hint: $out" | head -3
    fi
    echo
    echo "--- podman ---"
    if ! command -v podman >/dev/null 2>&1; then
        echo "  podman not installed"
    elif out=$(podman version 2>&1); then
        echo "$out" | head -8 | sed 's/^/  /'
    else
        echo "  podman unreachable"
        echo "  hint: $out" | head -3
    fi

# === Binary size analysis =====================================================

size-analysis:
    GOEXPERIMENT=jsonv2 go install github.com/Zxilly/go-size-analyzer/cmd/gsa@latest
    gsa --web ./dist/docker-tui
    go install github.com/imjasonh/gobinsize@latest
    gobinsize -svg out.svg ./dist/docker-tui

# === TUI screenshot helper ====================================================

# Capture TUI screenshot via tmux (output file: /tmp/dtui-screenshot.txt)
tui-shot OUT="/tmp/dtui-screenshot.txt" WIDTH="140" HEIGHT="40" WAIT="2":
    ./tui-screenshot.sh --width {{WIDTH}} --height {{HEIGHT}} --wait {{WAIT}} {{OUT}}

# Capture TUI screenshot with ANSI escape sequences stripped
tui-shot-plain OUT="/tmp/dtui-screenshot.txt" WIDTH="140" HEIGHT="40" WAIT="2":
    ./tui-screenshot.sh --width {{WIDTH}} --height {{HEIGHT}} --wait {{WAIT}} --strip-ansi {{OUT}}
