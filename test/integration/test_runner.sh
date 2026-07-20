#!/usr/bin/env bash
# dtui integration test runner
# Auto-detects Podman or Docker and validates container engine integration.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
BINARY="$PROJECT_DIR/docker-tui"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.test.yml"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0; FAIL=0
pass() { echo -e "${GREEN}[PASS]${NC} $*"; ((PASS++)); }
fail() { echo -e "${RED}[FAIL]${NC} $*"; ((FAIL++)); }
info() { echo -e "${YELLOW}[INFO]${NC} $*"; }

CONTAINERS=(dtui-test-postgres dtui-test-nginx dtui-test-redis dtui-test-alpine)

cleanup() {
    info "Cleaning up test containers..."
    for c in "${CONTAINERS[@]}"; do
        podman rm -f "$c" 2>/dev/null || docker rm -f "$c" 2>/dev/null || true
    done
}

detect_engine() {
    command -v podman &>/dev/null && podman info &>/dev/null 2>&1 && { echo "podman"; return; }
    command -v docker &>/dev/null && docker info &>/dev/null 2>&1 && { echo "docker"; return; }
    echo ""
}

setup_containers() {
    ENGINE="$1"
    cleanup
    info "Starting containers with $ENGINE..."

    $ENGINE run -d --name dtui-test-postgres \
        -e POSTGRES_PASSWORD=dtui_test -e POSTGRES_DB=dtui_testdb \
        --label com.docker.compose.project=dtui-test \
        swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/postgres:18.4-alpine

    $ENGINE run -d --name dtui-test-nginx \
        --label com.docker.compose.project=dtui-test nginx:alpine

    $ENGINE run -d --name dtui-test-redis \
        --label com.docker.compose.project=dtui-test redis:7-alpine

    $ENGINE run -d --name dtui-test-alpine \
        --label com.docker.compose.project=dtui-test alpine:3.19 sleep infinity

    sleep 5
    pass "test containers started (postgres, nginx, redis, alpine)"
}

# ─── Main ──────────────────────────────────────────────────────────────

echo "========================================="
echo "  dtui Integration Tests"
echo "========================================="
echo ""

# Phase 1: Binary smoke tests
echo "--- Phase 1: Binary Tests ---"
if [ -f "$BINARY" ]; then
    pass "binary exists"
else
    fail "binary not found - run 'just build' first"; exit 1
fi

"$BINARY" --version 2>&1 | grep -q "dtui version" && pass "version" || fail "version"
"$BINARY" --help 2>&1 | grep -q "dtui" && pass "help" || fail "help"

themes=$("$BINARY" --list-themes 2>&1)
if echo "$themes" | grep -q "default"; then
    pass "themes listed ($(echo "$themes" | wc -l) found)"
else
    fail "theme listing"
fi

# Phase 2: Container engine tests
echo ""
echo "--- Phase 2: Container Integration ---"
ENGINE=$(detect_engine)
if [ -z "$ENGINE" ]; then
    info "No container engine available, skipping integration tests"
else
    info "Using engine: $ENGINE"
    setup_containers "$ENGINE"

    count=$($ENGINE ps -q | wc -l)
    [ "$count" -ge 4 ] && pass "container count ($count)" || fail "container count ($count)"

    running=$($ENGINE ps -q -f "status=running" | wc -l)
    [ "$running" -ge 3 ] && pass "running containers ($running)" || fail "running containers ($running)"

    images=$($ENGINE images -q | wc -l)
    [ "$images" -ge 3 ] && pass "image count ($images)" || fail "image count ($images)"

    cleanup
    pass "cleanup complete"
fi

echo ""
echo "========================================="
echo -e "Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}"
echo "========================================="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
