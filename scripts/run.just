# run.just — Run recipes

out := "../dist"

# Build and run (Docker)
run: _build
    {{out}}/dtui

# Build and run (Podman)
run-podman: _build
    {{out}}/dtui -H unix:///run/user/1000/podman/podman.sock

# Internal: build binary for run recipes
_build:
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o {{out}}/dtui ./cmd/docker-tui
