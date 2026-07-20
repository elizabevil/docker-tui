# dtui — Docker/Podman TUI Manager
out := "./dist"

default: build

build:
    CGO_ENABLED=0 go build -o {{ out }}/docker-tui ./cmd/docker-tui

# ── Cross-compile (CGo) ──

build-linux-amd64:
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o {{ out }}/docker-tui-linux-amd64 ./cmd/docker-tui

build-linux-arm64:
    GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CGO_ENABLED=0 go build -o {{ out }}/docker-tui-linux-arm64 ./cmd/docker-tui

# ── Cross-compile (no CGo, pure Go) ──

build-linux-amd64-nocgo:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o {{ out }}/docker-tui-linux-amd64 ./cmd/docker-tui

build-linux-arm64-nocgo:
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o {{ out }}/docker-tui-linux-arm64 ./cmd/docker-tui

build-darwin-amd64-nocgo:
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o {{ out }}/docker-tui-darwin-amd64 ./cmd/docker-tui

build-darwin-arm64-nocgo:
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o {{ out }}/docker-tui-darwin-arm64 ./cmd/docker-tui

# ── Build all ──

build-all-nocgo: build-linux-amd64-nocgo build-linux-arm64-nocgo build-darwin-amd64-nocgo build-darwin-arm64-nocgo

build-all-cgo: build-linux-amd64 build-linux-arm64

build-all: build-all-cgo build-all-nocgo

# ── Test ──

# Run the complete Go test suite (integration tests are opt-in).
test:
    go test ./... -count=1

# Run package-level unit tests only.
test-unit:
    go test ./cmd/... ./internal/... -count=1

# Run Docker client tests.
test-docker:
    CGO_ENABLED=0 go test ./internal/data/docker/ -v -run 'TestManifest|TestList|TestImage' -count=1

# Build the binary and run container-engine integration tests.
test-integration: build
    bash test/integration/run.sh

# Run image-size formatting diagnostics without a container engine.
test-image-size:
    go test ./test/diagnostics/image-size/ -v -count=1

# Compare image sizes against a live Podman API.
test-image-size-live:
    bash test/diagnostics/image-size/compare.sh

# Run serialization benchmarks.
bench:
    go test ./test/benchmarks/ -run '^$' -bench . -benchmem

# Run Go tests followed by integration tests.
test-all: test test-integration

# ── Verify ──

# Run static analysis and the complete Go test suite.
check:
    CGO_ENABLED=0 go vet ./...
    go test ./... -count=1

# ── Run ──

run: build
    {{ out }}/docker-tui

run-podman: build
    {{ out }}/docker-tui -H unix:///run/user/1000/podman/podman.sock

# ── Clean ──

clean:
    rm -rf {{ out }}
    go clean -cache
