#!/usr/bin/env bash
# dtui compose integration test runner
# Usage: bash run.sh [auto|docker|podman]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TMPDIR="${DTUI_TEST_TMPDIR:-/tmp/dtui-test}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0

pass() { echo -e "${GREEN}[PASS]${NC} $*"; PASS=$((PASS + 1)); }
fail() { echo -e "${RED}[FAIL]${NC} $*"; FAIL=$((FAIL + 1)); }
info() { echo -e "${YELLOW}[INFO]${NC} $*"; }

# Detect available engine
detect-engine() {
    local requested="$1"
    if [[ "$requested" == "auto" ]]; then
        if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
            echo "docker"
        elif command -v podman &>/dev/null && podman info &>/dev/null 2>&1; then
            echo "podman"
        else
            echo ""
        fi
    elif [[ "$requested" == "docker" ]]; then
        if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
            echo "docker"
        else
            echo ""
        fi
    elif [[ "$requested" == "podman" ]]; then
        if command -v podman &>/dev/null && podman info &>/dev/null 2>&1; then
            echo "podman"
        else
            echo ""
        fi
    else
        echo ""
    fi
}

# Get compose command for engine
get-compose-cmd() {
    local engine="$1"
    if [[ "$engine" == "docker" ]]; then
        echo "docker compose"
    else
        echo "podman-compose"
    fi
}

# Ensure tmp directories exist
setup-tmp() {
    mkdir -p "$TMPDIR/www"
    echo "dtui-test-content" > "$TMPDIR/www/index.html"
}

# Cleanup function
cleanup() {
    local engine="$1"
    local compose_cmd
    compose_cmd=$(get-compose-cmd "$engine")
    info "Cleaning up..."
    cd "$SCRIPT_DIR"
    for f in batch-task.yml network.yml web-app.yml; do
        [[ -f "$f" ]] && $compose_cmd -f "$f" down -v 2>/dev/null || true
    done
}

# Run a test suite
run-suite() {
    local engine="$1"
    local suite="$2"
    local compose_file="$3"
    local check_name="$4"
    local compose_cmd
    compose_cmd=$(get-compose-cmd "$engine")

    info "Testing $suite suite with $engine..."

    cd "$SCRIPT_DIR"

    # Up
    $compose_cmd -f "$compose_file" up -d
    sleep 3

    # Verify containers running
    local containers
    containers=$($compose_cmd -f "$compose_file" ps -q 2>/dev/null | wc -l)
    if [[ "$containers" -gt 0 ]]; then
        pass "$suite: containers started ($containers)"
    else
        fail "$suite: no containers started"
        $compose_cmd -f "$compose_file" down -v 2>/dev/null || true
        return 1
    fi

    # Check specific container
    if [[ -n "$check_name" ]]; then
        $engine ps -a --filter "name=$check_name" --format "{{.Names}}" | grep -q "$check_name" && pass "$suite: $check_name running" || fail "$suite: $check_name missing"
    fi

    # Logs test
    $compose_cmd -f "$compose_file" logs --tail=5 &>/dev/null && pass "$suite: logs accessible" || fail "$suite: logs failed"

    # Down
    $compose_cmd -f "$compose_file" down -v
    pass "$suite: down complete"
}

# Main
ENGINE_REQUESTED="${1:-auto}"
ENGINE=$(detect-engine "$ENGINE_REQUESTED")

if [[ -z "$ENGINE" ]]; then
    echo "Error: No container engine available"
    exit 1
fi

info "Using engine: $ENGINE"
info "Tmp directory: $TMPDIR"

setup-tmp
trap "cleanup $ENGINE" EXIT

echo ""
echo "========================================="
echo "  dtui Compose Integration Tests"
echo "  Engine: $ENGINE"
echo "========================================="
echo ""

# Run all suites
run-suite "$ENGINE" "web-app" "web-app.yml" "dtui-webapp-nginx" || true
run-suite "$ENGINE" "network" "network.yml" "dtui-net-nginx" || true
run-suite "$ENGINE" "batch-task" "batch-task.yml" "dtui-batch-once" || true

echo ""
echo "========================================="
echo -e "Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}"
echo "========================================="

[[ "$FAIL" -eq 0 ]] && exit 0 || exit 1
