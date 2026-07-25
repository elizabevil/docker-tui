# test.just — Test & benchmark recipes

out := "./dist"

# Run the complete Go test suite (integration tests are opt-in)
test:
    go test ./... -count=1

# Run package-level unit tests only
test-unit:
    go test ./cmd/... ./internal/... -count=1

# Verify both Podman transports selected by the CGO build tag
test-runtime-matrix:
    CGO_ENABLED=0 go test ./... -count=1
    CGO_ENABLED=1 go test ./... -count=1

# Run Docker client tests
test-docker:
    CGO_ENABLED=0 go test ./internal/data/docker/ -v -run 'TestManifest|TestList|TestImage' -count=1

# Build the binary and run container-engine integration tests
test-integration: _build
    bash test/integration/run.sh

# Run image-size formatting diagnostics without a container engine
test-image-size:
    go test ./test/diagnostics/image-size/ -v -count=1

# Compare image sizes against a live Podman API
test-image-size-live:
    bash test/diagnostics/image-size/compare.sh

# Run serialization benchmarks
bench:
    go test ./test/benchmarks/ -run '^$' -bench . -benchmem

# Run Go tests followed by integration tests
test-all: test test-integration

# Internal: build binary for test-integration
_build:
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o {{out}}/docker-tui ./cmd/docker-tui
