#!/usr/bin/env bash
# verify-build-matrix.sh — TASK-022 Phase F verification matrix.
#
# Confirms that the project builds and tests cleanly under both
# CGO_ENABLED=0 and CGO_ENABLED=1, and that the non-CGO binary has no
# transitive dependency on github.com/proglottis/gpgme (the upstream
# source of libgpgme, required only by Podman's signature-verification
# path which lives behind the //go:build cgo gate).
#
# Usage: scripts/verify-build-matrix.sh
# Exits non-zero on any failure. Suitable for CI.

set -euo pipefail

cd "$(dirname "$0")/.."

require_clean() {
    local label="$1"
    shift
    echo "==> $label"
    "$@"
}

echo "=== TASK-022 Phase F: build matrix verification ==="

# 1. CGO_ENABLED=0 build + tests + zero gpgme transitive deps
require_clean "CGO_ENABLED=0 build" env CGO_ENABLED=0 go build ./...
require_clean "CGO_ENABLED=0 test"  env CGO_ENABLED=0 go test ./...

echo "--- Non-CGO dependency inspection ---"
if CGO_ENABLED=0 go list -deps ./cmd/docker-tui | grep -E "gpgme|keybase/go-keychain"; then
    echo "FAIL: non-CGO binary still pulls gpgme transitively" >&2
    exit 1
fi
echo "ok: non-CGO binary has zero gpgme / keybase dependency"

# 2. CGO_ENABLED=1 build + tests
require_clean "CGO_ENABLED=1 build" env CGO_ENABLED=1 go build ./...
require_clean "CGO_ENABLED=1 test"  env CGO_ENABLED=1 go test ./...

# 3. Static analysis under both modes
echo "--- go vet under both build modes ---"
require_clean "go vet (CGO=0)" env CGO_ENABLED=0 go vet ./...
require_clean "go vet (CGO=1)" env CGO_ENABLED=1 go vet ./...

# 4. go mod why confirms gpgme still routes through CGO-only files
echo "--- gpgme dependency chain (informational) ---"
go mod why github.com/proglottis/gpgme || true

echo
echo "=== build matrix PASSED ==="