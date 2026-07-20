# Tests

Package-level unit tests live next to the code they exercise under `cmd/` and
`internal/`. This directory contains cross-package tests and supporting assets:

- `benchmarks/`: serialization benchmarks and their input data.
- `diagnostics/image-size/`: offline and live image-size comparisons.
- `fixtures/`: configuration and media used for manual testing.
- `integration/`: binary smoke tests and Docker/Podman integration fixtures.

Use `just --list` to see all test commands. The usual entry points are:

```bash
just test             # all Go tests; no container engine required
just test-unit        # package-level unit tests only
just test-integration # binary and container-engine integration tests
just bench            # serialization benchmarks
just test-all         # Go and integration tests
```

Integration tests may pull images and create temporary containers named with
the `dtui-test-` prefix. The runner removes those containers when it exits.
