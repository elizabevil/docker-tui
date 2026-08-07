# batch-task.just — Batch Task suite (batch-task.yml)

# Start batch-task suite: just batch-task [docker|podman]
batch-task engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else echo "no container engine available" >&2; exit 1; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f batch-task.yml up -d && docker compose -f batch-task.yml ps; else podman-compose -f batch-task.yml up -d && podman-compose -f batch-task.yml ps; fi'

# Stop batch-task suite
batch-task-down engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else ENGINE=""; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f batch-task.yml down -v; else podman-compose -f batch-task.yml down -v; fi'

# Show batch-task containers
batch-task-ps engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else ENGINE=""; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f batch-task.yml ps; else podman-compose -f batch-task.yml ps; fi'

# View batch-task logs
batch-task-logs engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else ENGINE=""; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f batch-task.yml logs --tail=20; else podman-compose -f batch-task.yml logs --tail=20; fi'
