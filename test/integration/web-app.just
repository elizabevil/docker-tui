# web-app.just — Web Application suite (web-app.yml)

# Start web-app suite: just web-app [docker|podman]
web-app engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else echo "no container engine available" >&2; exit 1; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; mkdir -p /tmp/dtui-test/www && echo "dtui-test-content" > /tmp/dtui-test/www/index.html; if [ "$ENGINE" = "docker" ]; then docker compose -f web-app.yml up -d && docker compose -f web-app.yml ps; else podman-compose -f web-app.yml up -d && podman-compose -f web-app.yml ps; fi'

# Stop web-app suite
web-app-down engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else ENGINE=""; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f web-app.yml down -v; else podman-compose -f web-app.yml down -v; fi'

# Show web-app containers
web-app-ps engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else ENGINE=""; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f web-app.yml ps; else podman-compose -f web-app.yml ps; fi'

# View web-app logs
web-app-logs engine="":
    @bash -c 'if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then ENGINE=docker; elif command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then ENGINE=podman; else ENGINE=""; fi; if [ -n "{{engine}}" ]; then ENGINE="{{engine}}"; fi; if [ "$ENGINE" = "docker" ]; then docker compose -f web-app.yml logs --tail=20; else podman-compose -f web-app.yml logs --tail=20; fi'
