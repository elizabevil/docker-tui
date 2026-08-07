# run.just — Test runner and cleanup recipes

tmpdir := env("DTUI_TEST_TMPDIR", "/tmp/dtui-test")

# Run integration tests (auto-detect engine)
test engine="":
    bash run.sh {{if engine == "" { "auto" } else { engine }}}

# Run tests on docker
test-docker:
    bash run.sh docker

# Run tests on podman
test-podman:
    bash run.sh podman

# Full cleanup: down all suites and remove tmp dirs
clean:
    @echo "Cleaning up all suites..."
    -podman-compose -f batch-task.yml down -v 2>/dev/null || true
    -podman-compose -f worker-job.yml down -v 2>/dev/null || true
    -podman-compose -f microservice.yml down -v 2>/dev/null || true
    -podman-compose -f web-app.yml down -v 2>/dev/null || true
    -docker compose -f batch-task.yml down -v 2>/dev/null || true
    -docker compose -f worker-job.yml down -v 2>/dev/null || true
    -docker compose -f microservice.yml down -v 2>/dev/null || true
    -docker compose -f web-app.yml down -v 2>/dev/null || true
    @rm -rf {{tmpdir}}
    @echo "Cleanup complete"

# Show test usage help
help:
    @echo "Compose Integration Test Suite"
    @echo ""
    @echo "Suites:"
    @echo "  just web-app      # Web Application (nginx, redis, postgres, alpine)"
    @echo "  just microservice # Microservice Architecture (custom networks, multi-container)"
    @echo "  just worker-job   # Worker/Job (bind mounts, tmpfs, env_file)"
    @echo "  just batch-task   # Batch Task (exited containers, restart policies)"
    @echo ""
    @echo "Commands:"
    @echo "  just test        # Run full test suite (auto-detect engine)"
    @echo "  just test-docker # Run tests on Docker"
    @echo "  just test-podman # Run tests on Podman"
    @echo "  just clean       # Full cleanup (down + rm tmp dirs)"
    @echo ""
    @echo "Examples:"
    @echo "  just web-app podman      # Start web-app on podman"
    @echo "  just microservice docker # Start microservice on docker"
    @echo "  just worker-job-logs    # View worker-job logs"
