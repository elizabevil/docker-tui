# ISSUE: musl + gpgme static-build failures

**Status**: ✅ RESOLVED — `just musl-gpgme::build-musl-static` passes all 4
stages for both `DTUI_ARCH=amd64` and `DTUI_ARCH=arm64`; smoke test prints
`dtui version` and exits 0.

**Scope**: three independent failures encountered while fixing the original
`gpgrt fatal: sizeof lock obj` runtime crash. Each has its own root cause
and its own local-side fix; all three together are required to ship a
working statically-linked dtui for musl amd64 + arm64.

---

## User-stated constraints (verbatim)

1. **库路径是否正确** — verify library paths.
2. **musl编译是否需要其他依赖?** — does musl cross-compile need extra
   dependencies?
3. **不应该修改官方依赖代码,本地只负责编译** — must NOT modify upstream
   dependency source; local side only handles compilation.

Every fix below complies with all three.

---

## Fix #1 — `gpgrt fatal: sizeof lock obj` runtime crash

### Symptom

```
$ ./dist/dtui-musl-static-amd64 --version
gpgrt fatal: sizeof lock obj
SIGABRT: abort
PC=0x14db01d m=0 sigcode=18446744073709551610
signal arrived during cgo execution
```

Failure is in `github.com/proglottis/gpgme v0.1.6 init()` →
`gpgme_check_version` → `_gpgrt_lock_init` →
`get_lock_object()`. Link stage passed; the runtime check fires on first
cgo global init.

### Root cause

`libgpg-error` ships a public type `gpgrt_lock_t` whose layout is
*generated* by `src/gen-lock-obj.sh` and substituted into
`gpg-error.h` via the `@include:lock-obj@` placeholder. The generator
runs at `make` time and is supposed to probe the target's
`pthread_mutex_t` size and emit the matching `_priv[N]`.

On musl cross-compile, the generator silently fails (`$CC` not exported
into the script's subshell, `mawk` output buffering, or similar), and
configure re-reads the partially-generated `lock-obj-pub.native.h`,
yielding a broken `_priv[0]` form:

```c
volatile char _priv[0];   /* wrong — should be _priv[40] for x86_64-musl */
```

With `_priv[0]`, the installed `gpgrt_lock_t` is 16 bytes, but the
*compiled* internal `_gpgrt_lock_t` (which wraps `pthread_mutex_t` =
40 bytes on musl) is 48 bytes. The runtime check
`sizeof(gpgrt_lock_t) < sizeof(_gpgrt_lock_t)` fires and aborts.

### Fix — Option C from original analysis (recommended + implemented)

`libgpg-error`'s source ships hand-curated syscfg templates under
`src/syscfg/`. The file
`src/syscfg/lock-obj-pub.x86_64-unknown-linux-musl.h` contains the
correct `_priv[40]` form. mkheader, in non-cross mode (when
`force_use_syscfg` is not set), prefers `./lock-obj-pub.native.h` over
re-running `gen-lock-obj.sh` — so we just overwrite the configure-generated
file with the curated template before `make`:

```just
# in _c-lib recipe, between configure and make, for libgpg-error only:
cp "$src_dir/src/syscfg/lock-obj-pub.x86_64-unknown-linux-musl.h" \
   "$build_dir/src/lock-obj-pub.native.h"
```

**Constraint compliance**:
- #1 (path): no change to library path; the fix is purely a template
  substitution within the build tree.
- #2 (deps): no extra system packages needed.
- #3 (no upstream edits): we `cp` upstream's own curated file — zero
  modifications to any upstream source.

### Reused for arm64

`sizeof(pthread_mutex_t) == 40` on BOTH musl amd64 and musl arm64
(measured). The x86_64-musl syscfg template's `_priv[40]` is therefore
correct for arm64 too — upstream has no `aarch64-unknown-linux-musl`
syscfg file, but the musl-side layout only depends on
`pthread_mutex_t` size, which matches.

### Verification

- Installed `gpg-error.h` now contains `_priv[40]` and matching
  `GPGRT_LOCK_INITIALIZER`.
- `ar x libgpg-error.a libgpg_error_la-posix-lock.o` → object contains
  0 occurrences of the fatal string (`48 < 48` is folded out at compile
  time, gcc -O2 DCEs the `fprintf + abort`).
- C test program linked against the fresh `libgpg-error.a`:
  `sizeof(gpgrt_lock_t) = 48`, `gpgrt_lock_init` rc = 0, lock/unlock
  roundtrip OK.

---

## Fix #2 — stale `funopen.o` (x86-64) in arm64 `libassuan.a`

### Symptom

```
DTUI_ARCH=arm64 just musl-gpgme::build-musl-static
...
ld: ../src/.libs/libassuan.a(funopen.o): Relocations in generic ELF (EM: 62)
ld: ../src/.libs/libassuan.a: error adding symbols: file in wrong format
collect2: error: ld returned 1 exit status
```

`file` on the offending object: `funopen.o: ELF 64-bit LSB relocatable,
**x86-64**, version 1 (SYSV)`. Object's EM is 62 (AARCH64) — but the
*bytes* are amd64, so the arm64 `ld` rejects the file.

### Root cause

`libassuan` declares `funopen.c` as a LIBOBJS autoconf replacement
(`Makefile: LIBOBJS = ${LIBOBJDIR}funopen$U.o`). LIBOBJS is compiled
directly by the host compiler — it does NOT go through libtool, so the
resulting object keeps the bare `funopen.o` name (no
`libassuan_la-` prefix).

When the build tree is shared across arches:

1. amd64 build compiles `funopen.c` with `x86_64-linux-musl-gcc`,
   producing `funopen.o` (x86-64). Timestamped `t1`.
2. arm64 build runs configure (`t2 > t1`), regenerates Makefile, but
   `funopen.c` itself is unchanged → `make` decides `funopen.o` is up to
   date and skips the rebuild.
3. The stale amd64 `funopen.o` gets linked into the arm64 `.a`.

`make`'s dependency tracking only checks source timestamps — it has no
notion of "the compiler changed". This affects any autoconf-generated
project with LIBOBJS; libgpg-error has `isascii`/`memrchr` in similar
positions (they happened to not be needed for this build, but the
failure mode is general).

### Fix — arch-suffixed build directories

In `_c-lib`, append `{{ arch }}` to the build directory:

```just
# was:  build_dir="{{ work_dir }}/build/{{ name }}-{{ ver }}"
build_dir="{{ work_dir }}/build/{{ name }}-{{ ver }}-{{ arch }}"
```

Each arch gets a fresh configure + make with its own cross-compiler;
no stale object can survive across builds. `src/` and `src-cache/`
remain shared (source is arch-independent).

**Constraint compliance**:
- #1 (path): the change is purely about where the build tree lives;
  installed library paths (under `/opt/musl/<arch>-linux-musl-cross/...`)
  are unchanged and already arch-separated.
- #2 (deps): no new dependencies.
- #3 (no upstream edits): zero changes to any upstream source.

### Verification

All `.o` files in the arm64 `libassuan.a` are `aarch64`; the x86-64
`funopen.o` is gone. arm64 `make` runs to completion.

---

## Fix #3 — missing `btrfs/*.h` for arm64 cross-compile

### Symptom

After Fix #2, arm64 C builds succeed. The next failure is in step 2
(`go build`):

```
# go.podman.io/storage/drivers/btrfs
../../go/pkg/mod/go.podman.io/storage@v1.63.0/drivers/btrfs/version.go:6:10:
  fatal error: btrfs/version.h: No such file or directory
```

`go.podman.io/storage/drivers/btrfs/version.go` does:

```go
//go:build linux && cgo
/*
#include <btrfs/version.h>
...
*/
import "C"
```

amd64 musl sysroot ships `btrfs/*.h` (libbtrfs-dev user-space headers).
The arm64 musl sysroot doesn't — so cgo can't resolve
`<btrfs/version.h>`.

Installing `linux-libc-dev-arm64-cross` would fix this, but requires
sudo and is therefore out of scope.

### Root cause

Constraint #2 (extra deps): arm64 cross-compile needs arm64 userspace
headers (`btrfs/*.h`), which aren't in the musl-cross toolchain's
sysroot and aren't installable without sudo.

### Fix — vendored local include dir

The 8 files libbtrfs-dev ships are user-space library headers (Btrfs
data structures + ioctl numbers), not kernel headers. Their content is
architecture-independent — ioctl numbers and struct layouts are the
same on amd64 and arm64. Copying them verbatim into a local include
directory and pointing cgo at it via `CGO_CFLAGS=-I<dir>` works for
both arches.

```
scripts/musl-gpgme/include/btrfs/
├── ctree.h          (83934 bytes)
├── ioctl.h          (32797 bytes)
├── kerncompat.h     ( 5781 bytes)
├── rbtree_types.h   ( 1015 bytes)
├── send.h           ( 3103 bytes)
├── send-stream.h    ( 2831 bytes)
├── send-utils.h     ( 2655 bytes)
└── version.h        (  361 bytes)
```

In `_go-build`:

```just
CGO_CFLAGS="-Os -fno-stack-protector -I/home/debi/IdeaProjects/docker-tui/scripts/musl-gpgme/include" \
```

(Note: `{{ root }}` / `{{ justfile_directory() }}` inside the bash
double-quoted `CGO_CFLAGS` value in this recipe wasn't being expanded
by `just` reliably — the literal path is hardcoded instead. Same
constraint compliance, no functional difference.)

For amd64, gcc searches our `-I` path first and finds the same content
as the musl sysroot's btrfs/*.h, so there's no behavioural change. For
arm64, our copy is the only available source and cgo resolves the
include successfully.

**Constraint compliance**:
- #1 (path): `CGO_CFLAGS` points at a local tracked directory; `ls`
  confirms the headers are where we claim.
- #2 (deps): no system-level installation; the extra "dependency" is
  provided locally inside the module.
- #3 (no upstream edits): zero changes to any upstream source — the
  btrfs C bindings in go.podman.io/storage are untouched; we only
  give the preprocessor a path to satisfy `#include <btrfs/*.h>`.

### Verification

arm64 `go build` compiles the btrfs cgo package successfully. Full
pipeline `==> [step 4] smoke test → build OK` for both arches.

---

## Summary

| # | Problem | Where | Local fix | Upstream code touched |
|---|---|---|---|---|
| 1 | `gpgrt fatal: sizeof lock obj` at runtime | libgpg-error configure | Copy upstream's syscfg template over the broken generated `native.h` in `_c-lib` | None (cp of upstream's own file) |
| 2 | `file in wrong format` for arm64 libassuan | shared build tree | Append `-{{ arch }}` to `build_dir` | None |
| 3 | `btrfs/version.h: No such file or directory` for arm64 go build | arm64 musl sysroot | Vendored `include/btrfs/*.h` + `CGO_CFLAGS=-I` | None (local headers only) |

All three fixes are build-side / local-side, no upstream source
modification. All three are required; any single one leaves a
remaining failure. All three verified on `amd64` and `arm64`.