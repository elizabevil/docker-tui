# build.just — Build & cross-compile recipes

out := "../dist"
version := env("DTUI_VERSION", "0.2.0")

# Get git commit hash (short)
commit:
    @git rev-parse --short HEAD 2>/dev/null || echo "unknown"

# Build binary (CGO disabled, static, stripped)
build:
    CGO_ENABLED=0 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/dtui \
        ../cmd/docker-tui

# ── Cross-compile (CGo) ──

build-linux-amd64:
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/dtui-linux-amd64 \
        ../cmd/docker-tui

build-linux-arm64:
    GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/dtui-linux-arm64 \
        ../cmd/docker-tui

# ── Cross-compile (no CGo, pure Go) ──

build-linux-amd64-nocgo:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/dtui-linux-amd64 \
        ../cmd/docker-tui

build-linux-arm64-nocgo:
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/dtui-linux-arm64 \
        ../cmd/docker-tui

build-darwin-amd64-nocgo:
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/dtui-darwin-amd64 \
        ../cmd/docker-tui

build-darwin-arm64-nocgo:
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/dtui-darwin-arm64 \
        ../cmd/docker-tui

build-windows-amd64-nocgo:
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
        -trimpath \
        -ldflags "-s -w -X internal/buildinfo.version={{ version }}" \
        -o {{ out }}/dtui-windows-amd64.exe \
        ../cmd/docker-tui

# ── Build all ──

build-all-nocgo: build-linux-amd64-nocgo build-linux-arm64-nocgo build-darwin-amd64-nocgo build-darwin-arm64-nocgo build-windows-amd64-nocgo

build-all-cgo: build-linux-amd64 build-linux-arm64

build-all: build-all-cgo build-all-nocgo

# ── Clean ──

clean:
    rm -rf {{ out }}
    go clean -cache

# Static musl build (CGO=1, gpgme linked in) lives in scripts/musl-gpgme/.
# Load it via:  just musl-gpgme::build-musl-static   (arch via DTUI_ARCH, default amd64)
