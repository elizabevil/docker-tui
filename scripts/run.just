# run.just — Run recipes

out := "./dist"

# Build and run (Docker)
run: _build
    {{out}}/docker-tui

# Build and run (Podman)
run-podman: _build
    {{out}}/docker-tui -H unix:///run/user/1000/podman/podman.sock

# Internal: build binary for run recipes
_build:
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o {{out}}/docker-tui ./cmd/docker-tui
