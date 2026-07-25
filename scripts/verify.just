# verify.just — Lint & verification recipes

# TASK-022 Phase F: confirm non-CGO binary pulls zero gpgme + vet under
# both build modes. Exits non-zero if gpgme leaks into the non-CGO path.
verify-build-matrix:
    bash scripts/verify-build-matrix.sh

# Run static analysis and the complete Go test suite
check: verify-build-matrix
    CGO_ENABLED=0 go vet ./...
    CGO_ENABLED=1 go vet ./...
    just test-runtime-matrix
