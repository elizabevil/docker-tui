# network.just — Network test suite (network.yml)

# Start network suite: just network [docker|podman]
network engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else echo "no container engine available" >&2; exit 1; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f network.yml up -d && docker compose -f network.yml ps; else podman-compose -f network.yml up -d && podman-compose -f network.yml ps; fi'

# Stop network suite
network-down engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else ENGINE=""; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f network.yml down -v; else podman-compose -f network.yml down -v; fi'

# Show network containers
network-ps engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else ENGINE=""; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f network.yml ps; else podman-compose -f network.yml ps; fi'

# View network logs
network-logs engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else ENGINE=""; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f network.yml logs --tail=20; else podman-compose -f network.yml logs --tail=20; fi'
