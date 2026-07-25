# build.just — Build & cross-compile recipes

out := "./dist"
version := "0.2.0"

# Get git commit hash (short)
commit:
    @git rev-parse --short HEAD 2>/dev/null || echo "unknown"

# Build binary (CGO disabled, static, stripped)
build:
    CGO_ENABLED=0 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/docker-tui \
        ../cmd/docker-tui

# ── Cross-compile (CGo) ──

build-linux-amd64:
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/docker-tui-linux-amd64 \
        ../cmd/docker-tui

build-linux-arm64:
    GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/docker-tui-linux-arm64 \
        ../cmd/docker-tui

# ── Cross-compile (no CGo, pure Go) ──

build-linux-amd64-nocgo:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/docker-tui-linux-amd64 \
        ../cmd/docker-tui

build-linux-arm64-nocgo:
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/docker-tui-linux-arm64 \
        ../cmd/docker-tui

build-darwin-amd64-nocgo:
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/docker-tui-darwin-amd64 \
        ../cmd/docker-tui

build-darwin-arm64-nocgo:
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/docker-tui-darwin-arm64 \
        ../cmd/docker-tui

# ── Build all ──

build-all-nocgo: build-linux-amd64-nocgo build-linux-arm64-nocgo build-darwin-amd64-nocgo build-darwin-arm64-nocgo

build-all-cgo: build-linux-amd64 build-linux-arm64

build-all: build-all-cgo build-all-nocgo

# ── Clean ──

clean:
    rm -rf {{ out }}
    go clean -cache
