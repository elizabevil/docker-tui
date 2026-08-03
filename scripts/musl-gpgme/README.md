# scripts/musl-gpgme/

Self-contained module for producing a **fully static** `dtui` binary whose
Podman bindings are statically linked against the musl C library and the
upstream GnuPG stack (`libgpg-error` + `libassuan` + `gpgme`).

The cross-compilation, `go build`, and verification gates all live here as
one `just` submodule so the rest of the build matrix
(`scripts/build.just`, `scripts/verify.just`) stays untouched.

---

## Quick start

```bash
# 1. install the musl-cross toolchain (host prefix + sysroot)
#    expected at /opt/musl/x86_64-linux-musl-cross/bin/x86_64-linux-musl-gcc
#    (any other location works if the binary is on $PATH)

# 2. from the repo root:
just musl-gpgme::build-musl-static                  # default: arch=amd64, version from DTUI_VERSION
DTUI_ARCH=arm64 just musl-gpgme::build-musl-static  # cross to aarch64
DTUI_VERSION=v0.3.0 just musl-gpgme::build-musl-static
just musl-gpgme::clean                              # wipe work/ + dist binary
```

Output:

```
dist/dtui-musl-static-amd64      (or dtui-musl-static-arm64)
```

Recipe pipeline (4 stages, each gated). Every stage is a private recipe,
so the pipeline can be re-run one step at a time
(`just musl-gpgme::_c-libs`, `just musl-gpgme::_go-build`, …):

| Stage | Recipe | What | Pass condition |
|---|---|---|---|
| 0 | `_toolchain` | Locate musl gcc, derive sysroot prefix | gcc found; fails fast otherwise |
| 1 | `_c-libs` → `_c-lib ×3` | Cross-compile `libgpg-error` / `libassuan` / `gpgme` against `$MUSL_PREFIX/usr` | `pkg-config --exists --static gpgme` succeeds |
| 2 | `_go-build` | `go build` with `CGO_LDFLAGS="-static $GPGME_LIBS"` | exit 0 |
| 3 | `_verify` | `file` + `ldd` + `readelf -d` | `statically linked`, `not a dynamic executable`, `DT_NEEDED` count == 0 |
| 4 | `_smoke` | `./dist/dtui-musl-static-amd64 --version` | binary prints version (currently fails — see below) |

---

## Files

| File / dir | Purpose |
|---|---|
| `justfile` | The `build-musl-static` pipeline, split into the step recipes above |
| `work/` | **Generated scratch area** — all build side-effects land here (`set working-directory`), so nothing ever pollutes the repo root: extracted sources (`work/src/`), configure/build trees (`work/build/` — incl. any `conftest.*` probes), tarball cache (`work/src-cache/`). Gitignored (kept via `.gitkeep`); wipe with `just musl-gpgme::clean` |
| `ISSUE-lock-obj-size.md` | Diagnosis of the current runtime crash; next-step patch path |
| `README.md` | This file |

The module is loaded by the root `justfile` via:

```just
mod musl-gpgme "scripts/musl-gpgme/justfile"
```

so all recipes are reachable as `just musl-gpgme::<recipe>`.

---

## Cache & environment knobs

| Variable | Default | Effect |
|---|---|---|
| `DTUI_ARCH` | `amd64` | Target arch: `amd64` or `arm64` (selects the matching `<triple>-gcc` on `$PATH`) |
| `DTUI_VERSION` | `0.2.0` | Embedded buildinfo version (`-X internal/buildinfo.version=...`) |
| `SRC_CACHE` | `scripts/musl-gpgme/work/src-cache` | Where downloaded tarballs are stashed; reused on next run |
| `GPGME_SRC_URL` | `https://gnupg.org/ftp/gcrypt` | Primary source tarball base URL |
| `GPGME_MIRROR_URL` | `https://www.mirrorservice.org/sites/ftp.gnupg.org/gcrypt` | Fallback URL tried on primary failure |
| `MUSL_PREFIX` | auto-detected next to the gcc binary | sysroot prefix used by `configure --prefix` |

---

## Why this module exists

`scripts/build.just` already builds a static binary via `CGO_ENABLED=0` —
that is the project's default static path. `dtui` ships a Podman
driver backed by the upstream `go.podman.io/podman/v6/pkg/bindings`
package, which transitively pulls in `github.com/proglottis/gpgme`. Under
`CGO_ENABLED=0` the cgo build gate drops the entire chain — including
the bindings — so gpgme never enters the binary, but neither do
libpod-remote operations (the driver backend is nil in non-CGO builds).

This module is for the *other* static story: keep `CGO_ENABLED=1`, swap
glibc for musl, and static-link libgpgme.a / libassuan.a / libgpg-error.a
into the same binary. End result: one self-contained ELF, no `.so`
dependencies, full Podman bindings including image-signature
verification.

See [`docs/design/unified-runtime-driver.md`](../../docs/design/unified-runtime-driver.md)
for the broader rationale.

---

## Current blocker

The build link stage succeeds (statically linked, 0 DT_NEEDED). The
runtime smoke test crashes with:

```
gpgrt fatal: sizeof lock obj
SIGABRT: abort
```

Full diagnosis, root cause analysis, and the proposed patch are in
[`ISSUE-lock-obj-size.md`](./ISSUE-lock-obj-size.md).