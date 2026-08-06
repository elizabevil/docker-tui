#!/usr/bin/env bash
# dtui integration test runner
# Auto-detects Podman or Docker and validates container engine integration.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
BINARY="$PROJECT_DIR/dist/docker-tui"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0; FAIL=0
pass() { echo -e "${GREEN}[PASS]${NC} $*"; PASS=$((PASS + 1)); }
fail() { echo -e "${RED}[FAIL]${NC} $*"; FAIL=$((FAIL + 1)); }
info() { echo -e "${YELLOW}[INFO]${NC} $*"; }

CONTAINERS=(dtui-test-postgres dtui-test-nginx dtui-test-redis dtui-test-alpine)

cleanup() {
    [ -n "${ENGINE:-}" ] || return 0
    info "Cleaning up test containers..."
    for c in "${CONTAINERS[@]}"; do
        "$ENGINE" rm -f "$c" 2>/dev/null || true
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
    trap cleanup EXIT
    info "Starting containers with $ENGINE..."

    "$ENGINE" run -d --name dtui-test-postgres \
        -e POSTGRES_PASSWORD=dtui_test -e POSTGRES_DB=dtui_testdb \
        --label com.docker.compose.project=dtui-test \
        swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/postgres:18.4-alpine

    "$ENGINE" run -d --name dtui-test-nginx \
        --label com.docker.compose.project=dtui-test nginx:alpine

    "$ENGINE" run -d --name dtui-test-redis \
        --label com.docker.compose.project=dtui-test redis:7-alpine

    "$ENGINE" run -d --name dtui-test-alpine \
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

    count=$("$ENGINE" ps -q | wc -l)
    [ "$count" -ge 4 ] && pass "container count ($count)" || fail "container count ($count)"

    running=$("$ENGINE" ps -q -f "status=running" | wc -l)
    [ "$running" -ge 3 ] && pass "running containers ($running)" || fail "running containers ($running)"

    images=$("$ENGINE" images -q | wc -l)
    [ "$images" -ge 3 ] && pass "image count ($images)" || fail "image count ($images)"

    cleanup
    trap - EXIT
    pass "cleanup complete"
fi

# ─── Phase 3: Multi-arch manifest merge ────────────────────────────
# Merges per-arch alpine base images into one multi-arch manifest list:
#   docker-bkrepo.cwoa.net/h536b8/canway_d/os/alpine:3.23.2-base      (amd64)
#   docker-bkrepo.cwoa.net/h536b8/canway_d_arm/os/alpine:3.23.2-base  (arm64)
# Env overrides:
#   DTUI_MANIFEST_REGISTRY   registry base (default: docker-bkrepo.cwoa.net/h536b8)
#   DTUI_MANIFEST_TAG        tag for both sources and the merged manifest
#                            (default: 3.23.2-base)
#   DTUI_PUSH_MANIFEST=1     push the merged manifest to the registry
echo ""
echo "--- Phase 3: Multi-arch Manifest Merge ---"

merge_manifest() {
    local engine="$1"
    local registry="${DTUI_MANIFEST_REGISTRY:-docker-bkrepo.cwoa.net/h536b8}"
    local tag="${DTUI_MANIFEST_TAG:-3.23.2-base}"
    local amd64="${registry}/canway_d/os/alpine:${tag}"
    local arm="${registry}/canway_d_arm/os/alpine:${tag}"
    local manifest_ref="${registry}/canway_d/os/alpine:${tag}"

    info "amd64 source: ${amd64}"
    info "arm source:   ${arm}"
    info "manifest:     ${manifest_ref}"

    # A stale manifest list at the target ref makes `manifest create` fail.
    if "$engine" manifest inspect "$manifest_ref" >/dev/null 2>&1; then
        info "Removing stale manifest at ${manifest_ref}"
        "$engine" manifest rm "$manifest_ref" >/dev/null 2>&1 || true
    fi

    info "Pulling source images..."
    "$engine" pull -q "$amd64" >/dev/null || { fail "pull ${amd64}"; return 1; }
    "$engine" pull -q "$arm" >/dev/null || { fail "pull ${arm}"; return 1; }

    if "$engine" manifest create "$manifest_ref" "$amd64" "$arm" >/dev/null 2>&1; then
        pass "manifest created: ${manifest_ref}"
    else
        fail "manifest create ${manifest_ref}"
        return 1
    fi

    local inspect_output entry_count
    inspect_output="$("$engine" manifest inspect "$manifest_ref" 2>/dev/null || true)"
    entry_count="$(printf '%s' "$inspect_output" | grep -oE '"architecture"' | wc -l | tr -d ' ' || true)"
    if [ "${entry_count:-0}" -ge 2 ]; then
        pass "manifest contains ${entry_count} platform entries:"
        printf '%s' "$inspect_output" \
            | grep -oE '"architecture"[[:space:]]*:[[:space:]]*"[^"]+"' \
            | sed 's/^/      /' | sort -u
    else
        fail "expected >= 2 platform entries, got ${entry_count:-0}"
        return 1
    fi

    if [ "${DTUI_PUSH_MANIFEST:-0}" = "1" ]; then
        info "Pushing manifest (DTUI_PUSH_MANIFEST=1)..."
        if [ "$engine" = "podman" ]; then
            "$engine" manifest push --all "$manifest_ref" >/dev/null 2>&1 \
                || { fail "manifest push ${manifest_ref}"; return 1; }
        else
            "$engine" manifest push "$manifest_ref" >/dev/null 2>&1 \
                || { fail "manifest push ${manifest_ref}"; return 1; }
        fi
        pass "manifest pushed: ${manifest_ref}"
    else
        info "Skipping push (set DTUI_PUSH_MANIFEST=1 to push to the registry)"
    fi

    "$engine" manifest rm "$manifest_ref" >/dev/null 2>&1 || true
    pass "local manifest cleaned up"
}

if [ -n "${ENGINE:-}" ]; then
    if merge_manifest "$ENGINE"; then
        pass "multi-arch manifest merge complete"
    else
        fail "multi-arch manifest merge"
    fi
else
    info "No container engine available, skipping manifest merge"
fi

echo ""
echo "========================================="
echo -e "Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}"
echo "========================================="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
