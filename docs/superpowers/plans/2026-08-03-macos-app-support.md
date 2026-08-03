# macOS Application Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `dtui` buildable and shippable as a native macOS application — both as a standalone binary and as a `.app` bundle — with correct default socket paths for Docker Desktop for Mac and Podman machine.

**Architecture:** Two-phase rollout.
- **Phase 1** (this plan, implementable from Linux, no macOS host needed for compile verification): keep common connection-building code in `internal/data/runtime/connection.go`, move platform default socket/resolver definitions into build-tagged files, and give macOS correct Docker Desktop plus dynamic Podman machine discovery.
- **Phase 2** (this plan, implementable from Linux, structural only — visual/runtime verification needs a Mac): add a `build-darwin-app` pipeline that wraps the darwin binary in a standard `.app` bundle with `Info.plist`, `PkgInfo`, and `MacOS/<binary>`.

**Tech Stack:** Go 1.26+ build tags, just recipes, bash bundle script, Apple `.app` bundle spec (Info.plist v1.0 / CFBundle keys), `lipo`-free single-arch `.app` (universal binaries deferred).

## Global Constraints

- macOS socket paths must match what Docker Desktop for Mac and Podman machine actually expose (no guessing — see Task 2 sources).
- No upstream dependency source modifications (same constraint as the just-committed musl-gpgme fixes).
- CGO_ENABLED=0 for all darwin targets in this plan (gpgme/Podman bindings stay Linux-only via `scripts/musl-gpgme/`; a darwin cgo story is out of scope and recorded as future work).
- All changes must keep `just verify-build-matrix` green on Linux amd64 + arm64.
- Existing `scripts/build.just` recipes must keep working unchanged.

---

## File Structure

| Path | Action | Responsibility |
|---|---|---|
| `internal/data/runtime/connection.go` | Modify (remove platform-specific socket constants/resolver only) | Shared runtime types + connection builder |
| `internal/data/runtime/connection_default.go` | Create (`//go:build !darwin`) | Linux + other Unix default socket paths |
| `internal/data/runtime/connection_darwin.go` | Create | macOS-specific socket path constants (Docker Desktop + Podman) |
| `internal/data/runtime/connection_darwin_test.go` | Create | Compile-time assertion that macOS paths differ from Linux paths |
| `internal/data/runtime/podman/discovery.go` | Modify (add `//go:build !darwin`) | Linux + other Unix Podman socket discovery |
| `internal/data/runtime/podman/discovery_darwin.go` | Create | macOS Podman machine socket discovery via `podman machine inspect` |
| `internal/data/runtime/podman/discovery_darwin_test.go` | Create | Unit tests for parsing Podman machine inspect output |
| `scripts/macos/Info.plist` | Create | CFBundle metadata template (executable name, version, identifier, LSUIElement) |
| `scripts/macos/build-app.sh` | Create | Assemble `dtui.app/` from a built binary + Info.plist + PkgInfo |
| `scripts/build.just` | Modify | Add `build-darwin-app` recipe + `build-darwin-app-all` aggregator |
| `README.md` | Modify | Add macOS section (build commands, run-from-`.app` note) |

---

## Phase 1 — macOS socket paths

### Task 1: Extract non-darwin socket defaults from `connection.go`

**Files:**
- Modify: `internal/data/runtime/connection.go`
- Create: `internal/data/runtime/connection_default.go`

**Interfaces:**
- Consumes: current platform-specific defaults in `connection.go`
- Produces: common connection builder stays compiled on every platform; non-darwin defaults move behind `//go:build !darwin`

- [ ] **Step 1: Remove platform-specific declarations from `connection.go`**

Open `internal/data/runtime/connection.go`. Do **not** add a build tag to the whole file: it contains shared types and `BuildConnections`, which must compile on darwin too.

Keep only the runtime type constants in the `const` block:

```go
const (
	// RuntimeDocker selects the Docker Engine backend.
	RuntimeDocker RuntimeType = "docker"
	// RuntimePodman selects the Podman backend.
	RuntimePodman RuntimeType = "podman"
)
```

Remove these declarations from `connection.go`:

```go
DefaultDockerSocket = "/var/run/docker.sock"
const defaultPodmanSocket = "/var/run/podman/podman.sock"
func defaultPodmanUserEndpoint(uid int) string { ... }
```

- [ ] **Step 2: Create non-darwin defaults**

Create `internal/data/runtime/connection_default.go`:

```go
//go:build !darwin
// +build !darwin

package runtime

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/utils"
)

const (
	// DefaultDockerSocket is the well-known Docker daemon socket path.
	DefaultDockerSocket = "/var/run/docker.sock"
)

// Default podman socket used when PodmanUserEndpoint is unavailable.
const defaultPodmanSocket = "/var/run/podman/podman.sock"

// defaultPodmanUserEndpoint is the production resolver for the per-user podman
// socket. It is package-level so tests can override it via WithPodmanResolver.
func defaultPodmanUserEndpoint(uid int) string {
	if uid <= 0 {
		return utils.SocketURI(defaultPodmanSocket)
	}
	return utils.SocketURI(fmt.Sprintf("/run/user/%d/podman/podman.sock", uid))
}
```

- [ ] **Step 3: Verify Linux build still works**

Run: `go build ./internal/data/runtime/...`
Expected: success, no output.

- [ ] **Step 4: Verify darwin build now FAILS (expected — defaults gone for darwin)**

Run: `GOOS=darwin GOARCH=amd64 go build ./internal/data/runtime/...`
Expected: errors referencing missing `DefaultDockerSocket` and `defaultPodmanUserEndpoint`. Shared types like `RuntimeType`, `ConnectionSpec`, and `BuildConnections` must still be present. This is the expected state — Task 2 will provide darwin defaults.

- [ ] **Step 5: Commit**

```bash
git add internal/data/runtime/connection.go internal/data/runtime/connection_default.go
git commit -m "build(runtime): split platform socket defaults

Move non-darwin Docker/Podman default socket paths and the podman
user endpoint resolver into connection_default.go behind !darwin.
Keep shared connection types and BuildConnections in connection.go
so darwin can reuse the common connection builder."
```

---

### Task 2: Create macOS connection defaults

**Files:**
- Create: `internal/data/runtime/connection_darwin.go`
- Modify: `internal/data/runtime/podman/discovery.go`
- Create: `internal/data/runtime/podman/discovery_darwin.go`

**Interfaces:**
- Consumes: nothing
- Produces: `DefaultDockerSocket`, `defaultPodmanSocket`, `defaultPodmanUserEndpoint` in package `runtime`; `DefaultPodmanSocket`, `PodmanSocketPath`, `PodmanUserEndpoint` in package `runtime/podman`; same existing public names so `runtimeinit` and `connectionBuilder` compile unchanged on darwin.

- [ ] **Step 1: Determine canonical macOS paths from authoritative sources**

Capture the following in the plan (do NOT change during implementation — these are the source of truth):

| Path | Value | Source |
|---|---|---|
| Docker Desktop default socket | `/Users/<USER>/.docker/run/docker.sock` | Docker Desktop for Mac documents that if `/var/run/docker.sock` is not enabled, clients may need `DOCKER_HOST` pointing at `/Users/<user>/.docker/run/docker.sock`. Resolved at runtime via `os.UserHomeDir() + "/.docker/run/docker.sock"`. |
| Podman machine socket | Dynamic: `podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}'` | Podman documents that macOS requires `podman machine`; `podman machine inspect` exposes `.ConnectionInfo.PodmanSocket.Path`. Current official examples show provider-dependent paths under `/var/folders/.../T/podman/...`, so do not hardcode `qemu`, `applehv`, `vz`, or `libkrun` paths. |
| Podman static fallback | Empty path, represented as empty string | There is no reliable static macOS fallback path. If `podman` is absent or `podman machine inspect` fails, return empty from discovery and let explicit `-H unix:///path/to/socket.sock` / config handle custom installations. |

Notes for the implementer:
- We cannot resolve the per-user home at `const` time — Go `const` cannot call functions. Use `var` instead, initialised via `os.UserHomeDir()` in an `init()` function. This means the value is computed once at process start (acceptable: the home dir doesn't change at runtime).
- Do not hardcode Podman provider directories. Podman provider defaults change across versions and host capabilities; authoritative discovery is `podman machine inspect`.
- `defaultPodmanUserEndpoint` must return a `unix://...` URI when a Podman socket is known, because `BuildConnections` stores endpoint URIs in `ConnectionSpec.Host`.

- [ ] **Step 2: Write the file**

Create `internal/data/runtime/connection_darwin.go` with this exact content:

```go
//go:build darwin
// +build darwin

package runtime

import (
	"os"
	"path/filepath"

	podmandiscovery "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

// macOS default socket paths.
//
// Docker Desktop for Mac: since 4.13 (Feb 2023) the daemon socket is
// exposed to the host at <HOME>/.docker/run/docker.sock — a symlink
// into the Docker Desktop VM's /var/run/docker.sock. Source: Docker
// Desktop release notes + Docker for Mac "Settings → Advanced → Allow
// the default Docker socket to be used".
//
// Podman machine: macOS socket paths are provider- and version-dependent.
// Resolve them through internal/data/runtime/podman discovery, which shells
// out to `podman machine inspect` on darwin. There is no reliable static
// provider path to hardcode.
//
// Note: Go const cannot call os.UserHomeDir(), so these are `var`
// initialised in init() below. The home dir is read once at startup;
// if it changes mid-run (it doesn't), callers would need to re-init.

var DefaultDockerSocket string
var defaultPodmanSocket string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to /tmp if $HOME is unreadable. Connection builders
		// will surface the failure visibly — we never want a silent
		// empty string that resolves to a nonsense URI.
		home = "/tmp"
	}
	DefaultDockerSocket = filepath.Join(home, ".docker", "run", "docker.sock")
	defaultPodmanSocket = podmandiscovery.PodmanSocketPath()
}

// defaultPodmanUserEndpoint is a no-op on macOS with respect to uid:
// podman machine exposes the active machine socket, not /run/user/<uid>.
// It returns a unix:// URI when discovery succeeds, or "" when no podman
// machine socket can be discovered.
func defaultPodmanUserEndpoint(uid int) string {
	return podmandiscovery.PodmanUserEndpoint(uid)
}
```

- [ ] **Step 3: Split Podman discovery for darwin**

Add a build tag to `internal/data/runtime/podman/discovery.go` so it remains the Linux/Unix implementation:

```go
//go:build !darwin
// +build !darwin

package podman
```

Create `internal/data/runtime/podman/discovery_darwin.go`:

```go
//go:build darwin
// +build darwin

package podman

import (
	"os/exec"
	"strings"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// DefaultPodmanSocket is empty on macOS because podman machine sockets are
// provider- and version-dependent. Use PodmanSocketPath/PodmanUserEndpoint to
// discover the active machine socket through `podman machine inspect`.
var DefaultPodmanSocket = ""

// PodmanSocketPath returns the active Podman machine socket path on macOS.
func PodmanSocketPath() string {
	return detectPodmanMachineSocket()
}

// PodmanUserEndpoint returns the active Podman machine socket URI on macOS.
// The uid is ignored: macOS has no /run/user/<uid>/podman/podman.sock.
func PodmanUserEndpoint(uid int) string {
	path := PodmanSocketPath()
	if path == "" {
		return ""
	}
	return utils.SocketURI(path)
}

func detectPodmanMachineSocket() string {
	podmanBin, err := exec.LookPath("podman")
	if err != nil {
		return ""
	}
	out, err := exec.Command(
		podmanBin, "machine", "inspect",
		"--format", "{{.ConnectionInfo.PodmanSocket.Path}}",
	).Output()
	if err != nil {
		return ""
	}
	return parsePodmanSocketPath(out)
}

func parsePodmanSocketPath(out []byte) string {
	return strings.TrimSpace(string(out))
}
```

- [ ] **Step 4: Verify both platforms compile**

Run:
```bash
go build ./internal/data/runtime/...
GOOS=darwin GOARCH=amd64 go build ./internal/data/runtime/...
GOOS=darwin GOARCH=arm64 go build ./internal/data/runtime/...
```

Expected: all three succeed, no output.

- [ ] **Step 5: Verify the wider module compiles on both**

Run:
```bash
go build ./...
GOOS=darwin GOARCH=amd64 go build ./...
GOOS=darwin GOARCH=arm64 go build ./...
```

Expected: all three succeed. If any consumer of `DefaultDockerSocket` / `defaultPodmanSocket` / `defaultPodmanUserEndpoint` was depending on the old compile-time-`const` semantics, this will surface it — fix any caller that now sees a `var` instead (the read-side API is identical; only the declaration type changed).

- [ ] **Step 6: Commit**

```bash
git add internal/data/runtime/connection_darwin.go internal/data/runtime/podman/discovery.go internal/data/runtime/podman/discovery_darwin.go
git commit -m "feat(runtime): add macOS default socket paths

Docker Desktop for Mac exposes the daemon at ~/.docker/run/docker.sock
(or via /var/run/docker.sock when the privileged symlink is enabled).
Podman machine socket paths are provider-dependent, so darwin uses
podman machine inspect instead of hardcoding qemu/applehv/vz/libkrun.

Build-tagged darwin-only; constants are var + init() because Go
const cannot call os.UserHomeDir().

Connection-builder code is unchanged: DefaultDockerSocket,
defaultPodmanSocket, defaultPodmanUserEndpoint keep the same names
and types (string). No caller-side adjustments required."
```

---

### Task 3: Add a darwin compile-time assertion test

**Files:**
- Create: `internal/data/runtime/connection_darwin_test.go`
- Create: `internal/data/runtime/podman/discovery_darwin_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: darwin-only tests that are excluded on Linux and run on macOS

The point of this test is not to run logic at runtime — it's a guard rail that someone doesn't accidentally regress the darwin values to the Linux ones. Because it's `//go:build darwin`, on Linux the file is excluded from the build, so the assertions are never even seen.

- [ ] **Step 1: Write the test file**

Create `internal/data/runtime/connection_darwin_test.go`:

```go
//go:build darwin
// +build darwin

package runtime

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// TestDarwinDockerSocketPath verifies the default Docker socket
// points at Docker Desktop for Mac's host-side symlink, NOT the
// linux /var/run/docker.sock. This guards against accidentally
// regressing to the linux default.
func TestDarwinDockerSocketPath(t *testing.T) {
	if DefaultDockerSocket == "/var/run/docker.sock" {
		t.Fatalf("DefaultDockerSocket is the linux path; darwin should " +
			"use ~/.docker/run/docker.sock (Docker Desktop for Mac)")
	}
	if !strings.HasSuffix(DefaultDockerSocket, "/.docker/run/docker.sock") {
		t.Fatalf("DefaultDockerSocket = %q; expected suffix /.docker/run/docker.sock",
			DefaultDockerSocket)
	}
}

// TestDarwinPodmanSocketPath verifies the default Podman socket is not
// hardcoded to a Linux path. The actual macOS machine socket is resolved
// by internal/data/runtime/podman via `podman machine inspect`.
func TestDarwinPodmanSocketPath(t *testing.T) {
	if defaultPodmanSocket == "/var/run/podman/podman.sock" {
		t.Fatalf("defaultPodmanSocket is the linux path; darwin should " +
			"resolve the podman machine socket dynamically")
	}
}

// TestDarwinPodmanUserEndpointMatchesSystem verifies the per-user
// podman endpoint on macOS is the same as the system endpoint
// (podman machine is system-wide; there is no per-user socket).
func TestDarwinPodmanUserEndpointMatchesSystem(t *testing.T) {
	if defaultPodmanSocket == "" {
		if got := defaultPodmanUserEndpoint(501); got != "" {
			t.Fatalf("darwin defaultPodmanUserEndpoint = %q; want empty when no machine socket is known", got)
		}
		return
	}
	if got := defaultPodmanUserEndpoint(501); got != utils.SocketURI(defaultPodmanSocket) {
		t.Fatalf("darwin defaultPodmanUserEndpoint(%d) = %q; want %q "+
			"(podman machine is system-wide on macOS, no per-user socket)",
			501, got, utils.SocketURI(defaultPodmanSocket))
	}
}
```

Create `internal/data/runtime/podman/discovery_darwin_test.go`:

```go
//go:build darwin
// +build darwin

package podman

import "testing"

func TestDarwinDefaultPodmanSocketIsNotLinuxPath(t *testing.T) {
	if DefaultPodmanSocket == "/run/podman/podman.sock" {
		t.Fatal("DefaultPodmanSocket is the linux system path; darwin must discover podman machine")
	}
}

func TestParsePodmanSocketPath(t *testing.T) {
	got := parsePodmanSocketPath([]byte("/var/folders/abc/T/podman/podman-machine-default-api.sock\n"))
	want := "/var/folders/abc/T/podman/podman-machine-default-api.sock"
	if got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}
```

- [ ] **Step 2: Run on linux (the file is excluded; test count unchanged)**

Run: `go test ./internal/data/runtime/...`
Expected: passes (the `_darwin_test.go` file is excluded by build tag, so it's invisible to the linux test run).

- [ ] **Step 3: Run on darwin (the assertions are evaluated)**

Run on a macOS host: `go test ./internal/data/runtime/...`
Expected: `ok` results for `internal/data/runtime` and `internal/data/runtime/podman`; darwin-only assertions are evaluated.

Do not run `GOOS=darwin go test` on Linux: Go will build a Mach-O test binary and then try to execute it, which fails with `exec format error`. If you don't have a macOS host, the most you can do is confirm the darwin test binary compiles:

```bash
GOOS=darwin GOARCH=amd64 go test -c ./internal/data/runtime
GOOS=darwin GOARCH=amd64 go test -c ./internal/data/runtime/podman
GOOS=darwin GOARCH=amd64 go vet ./internal/data/runtime/...
```

Expected: no errors. The assertions themselves can only be runtime-verified on a Mac; mark this as "verified by build on darwin" in the commit message.

- [ ] **Step 4: Commit**

```bash
git add internal/data/runtime/connection_darwin_test.go internal/data/runtime/podman/discovery_darwin_test.go
git commit -m "test(runtime): compile-time guard for darwin socket paths

Build-tagged darwin-only; on linux the file is excluded entirely.
Asserts that DefaultDockerSocket ends in /.docker/run/docker.sock
(Docker Desktop for Mac), Podman does not regress to linux socket
paths, and the per-user endpoint returns a unix URI only when a
machine socket is known."
```

---

### Task 4: Verify the end-to-end darwin nocgo build

**Files:**
- No file changes. This is a verification task.

- [ ] **Step 1: Cross-compile darwin amd64 + arm64**

Run:
```bash
just build-darwin-amd64-nocgo
just build-darwin-arm64-nocgo
```

Expected: both produce `dist/dtui-darwin-amd64` and `dist/dtui-darwin-arm64` with exit 0.

- [ ] **Step 2: Verify static-linkage gate**

Run:
```bash
file dist/dtui-darwin-amd64 dist/dtui-darwin-arm64
```

Expected: both report `Mach-O 64-bit executable x86_64` and `Mach-O 64-bit executable arm64` respectively. No dynamic loader warnings.

- [ ] **Step 3: Verify the binaries reference the new darwin paths (string-level check)**

Run:
```bash
strings dist/dtui-darwin-amd64 | grep -E '\.docker/run/docker\.sock|ConnectionInfo\.PodmanSocket\.Path|machine inspect' | head -3
```

Expected: at least one match for `.docker/run/docker.sock` or `podman machine inspect`. This confirms the darwin build actually contains the new Docker path and Podman dynamic discovery code. Do not require `podman/machine/qemu`, because provider paths must not be hardcoded.

- [ ] **Step 4: Confirm linux still builds cleanly**

Run: `just verify-build-matrix` (or the subset you can run locally)
Expected: linux amd64 + arm64 still pass; the darwin entries in the matrix either pass (if you can run them) or are noted as "darwin cross-compile verified on linux host".

No commit — verification only.

---

## Phase 2 — `.app` bundle packaging

### Task 5: Create Info.plist template

**Files:**
- Create: `scripts/macos/Info.plist`

The Info.plist is the manifest macOS reads for the `.app` bundle. `dtui` is still a terminal TUI: Finder/`open dtui.app` will not provide an interactive tty. The `.app` bundle is a distribution artifact for users/tools that expect macOS bundle layout; supported interactive use is launching `Contents/MacOS/dtui` from Terminal or using the standalone binary.

- [ ] **Step 1: Create the directory and file**

Create `scripts/macos/Info.plist` with this exact content:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <!-- Required keys -->
    <key>CFBundleExecutable</key>
    <string>dtui</string>
    <key>CFBundleIdentifier</key>
    <string>com.github.elizabevil.dtui</string>
    <key>CFBundleName</key>
    <string>dtui</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>__DTUI_VERSION__</string>
    <key>CFBundleVersion</key>
    <string>__DTUI_VERSION__</string>

    <!-- UI behaviour: this is a terminal TUI bundle, not a Cocoa app.
         Finder/open will not provide an interactive tty; launch the
         inner Contents/MacOS/dtui binary from Terminal. -->
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>

    <!-- Minimum macOS supported (10.13 High Sierra covers all
         Docker Desktop + Podman machine supported versions). -->
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>

    <!-- Optional but recommended for terminal apps -->
    <key>CFBundleDocumentTypes</key>
    <array/>
    <key>NSHumanReadableCopyright</key>
    <string>MIT-style (see project LICENSE if/when present)</string>
</dict>
</plist>
```

The `__DTUI_VERSION__` placeholders are substituted at build time by the script in Task 7.

- [ ] **Step 2: Validate the XML**

Run: `xmllint --noout scripts/macos/Info.plist` (if xmllint is available)
Expected: no output (= XML is well-formed).

If `xmllint` is not installed, skip this step — `plutil -lint scripts/macos/Info.plist` works on macOS but isn't available on Linux.

- [ ] **Step 3: Commit**

```bash
git add scripts/macos/Info.plist
git commit -m "feat(macos): add Info.plist template for .app bundle

CFBundleExecutable=dtui, identifier=com.github.elizabevil.dtui,
LSUIElement=true (agent/terminal TUI), LSMinimumSystemVersion=10.13.
Version placeholders __DTUI_VERSION__ substituted at build time."
```

---

### Task 6: Create the bundle assembly script

**Files:**
- Create: `scripts/macos/build-app.sh`

This script takes a built darwin binary and produces a `.app` bundle next to it. Pure bash, no extra tooling required.

- [ ] **Step 1: Create the script**

Create `scripts/macos/build-app.sh` with this exact content:

```bash
#!/usr/bin/env bash
# build-app.sh — wrap a darwin dtui binary in a .app bundle
#
# Usage: build-app.sh <binary-path> <output-app-path> [version]
#
# Output:
#   <output-app-path>/Contents/MacOS/dtui    <- the binary
#   <output-app-path>/Contents/Info.plist   <- metadata
#   <output-app-path>/Contents/PkgInfo      <- "APPL????" (8 bytes)
#
# Why a script (not a just recipe): the bundle is a flat directory
# tree; a recipe would just shell out to mkdir/cp anyway. Keeping
# it standalone makes it reusable from CI runners, manual debugging,
# and the just recipe alike.

set -euo pipefail

if [[ $# -lt 2 || $# -gt 3 ]]; then
    echo "usage: $0 <binary-path> <output-app-path> [version]" >&2
    exit 64
fi

BIN="$1"
APP="$2"
VERSION="${3:-0.2.0}"

if [[ ! -f "$BIN" ]]; then
    echo "error: binary not found: $BIN" >&2
    exit 66
fi

# Validate it's actually a Mach-O binary before we wrap it. Catches
# the obvious mistake of passing the linux binary by accident.
file_check="$(file -b "$BIN")"
if [[ "$file_check" != Mach-O* ]]; then
    echo "error: $BIN is not a Mach-O binary ($file_check)" >&2
    exit 65
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Clean any prior bundle so we never ship a half-staged one.
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS"
mkdir -p "$APP/Contents/Resources"

# 1. Copy the binary.
cp "$BIN" "$APP/Contents/MacOS/dtui"
chmod +x "$APP/Contents/MacOS/dtui"

# 2. Substitute version placeholders in the Info.plist template.
sed -e "s/__DTUI_VERSION__/$VERSION/g" \
    "$SCRIPT_DIR/Info.plist" > "$APP/Contents/Info.plist"

# 3. Write PkgInfo (legacy but Finder still expects it).
printf 'APPL????' > "$APP/Contents/PkgInfo"

echo "==> packaged $APP"
echo "    Contents:"
find "$APP" -mindepth 2 | sed 's|^|      |'
```

Make the script executable:

```bash
chmod +x scripts/macos/build-app.sh
```

- [ ] **Step 2: Smoke-test the script on the existing darwin binary**

Run:
```bash
scripts/macos/build-app.sh dist/dtui-darwin-amd64 /tmp/dtui-app-test/dtui-darwin-amd64.app
```

Expected: `==> packaged /tmp/dtui-app-test/dtui-darwin-amd64.app` followed by the `find` listing showing `MacOS/dtui`, `Info.plist`, `PkgInfo`.

- [ ] **Step 3: Inspect the produced bundle**

Run:
```bash
find /tmp/dtui-app-test/dtui-darwin-amd64.app -type f | sort
echo "--- Info.plist version substituted ---"
grep -A1 CFBundleVersion /tmp/dtui-app-test/dtui-darwin-amd64.app/Contents/Info.plist | head -4
echo "--- PkgInfo (must be exactly 8 bytes 'APPL????') ---"
xxd /tmp/dtui-app-test/dtui-darwin-amd64.app/Contents/PkgInfo
```

Expected:
- Three files: `Contents/Info.plist`, `Contents/MacOS/dtui`, `Contents/PkgInfo`.
- `CFBundleVersion` shows `0.2.0` (or the version passed), not `__DTUI_VERSION__`.
- `PkgInfo` shows `4150 504c 3f3f 3f3f 3f3f` (= ASCII "APPL????"), 8 bytes.

- [ ] **Step 4: Verify the bundled binary still works (file-level only — runtime needs a Mac)**

Run: `file /tmp/dtui-app-test/dtui-darwin-amd64.app/Contents/MacOS/dtui`
Expected: `Mach-O 64-bit executable x86_64` (or `arm64`). Confirms the cp preserved the binary unchanged.

- [ ] **Step 5: Cleanup + commit**

```bash
rm -rf /tmp/dtui-app-test
git add scripts/macos/build-app.sh
git commit -m "feat(macos): add .app bundle assembly script

scripts/macos/build-app.sh wraps a darwin binary in a named standard
.app bundle (Contents/{MacOS/dtui, Info.plist, PkgInfo}).

Validates the input is actually Mach-O before wrapping (catches
the mistake of passing a linux binary). Substitutes version
placeholders in Info.plist. Output layout matches the macOS bundle
directory structure; interactive use remains Terminal-based."
```

---

### Task 7: Wire `build-darwin-app` into `scripts/build.just`

**Files:**
- Modify: `scripts/build.just` (add new recipes at the end of the file)

- [ ] **Step 1: Read the current end of `scripts/build.just`**

Run: `tail -20 scripts/build.just`
Expected: the file ends with `build-all-nocgo` (or similar aggregator). We append after it.

- [ ] **Step 2: Append the new recipes**

Append to `scripts/build.just`:

```just
# ── macOS .app bundle ──
# Wraps dist/dtui-darwin-{amd64,arm64} into dist/dtui-darwin-{amd64,arm64}.app
# using scripts/macos/build-app.sh. Requires the matching build-darwin-*-nocgo
# step to have already produced the binary (this recipe depends on it).

build-darwin-amd64-app: build-darwin-amd64-nocgo
    bash macos/build-app.sh \
        {{ out }}/dtui-darwin-amd64 \
        {{ out }}/dtui-darwin-amd64.app \
        {{ version }}

build-darwin-arm64-app: build-darwin-arm64-nocgo
    bash macos/build-app.sh \
        {{ out }}/dtui-darwin-arm64 \
        {{ out }}/dtui-darwin-arm64.app \
        {{ version }}

build-darwin-app-all: build-darwin-amd64-app build-darwin-arm64-app
```

Notes:
- `macos/build-app.sh` is relative to the `scripts/` working directory used by `scripts/build.just`, matching the existing `../cmd/docker-tui` and `../dist` paths.
- Use same-file just dependencies (`build-darwin-amd64-app: build-darwin-amd64-nocgo`) instead of shelling out to `just`; this works both from `just --justfile scripts/build.just ...` and from the root module as `just build::build-darwin-amd64-app`.
- `bash` (not `sh`) is required because the script uses `set -euo pipefail` and bash-isms like `[[ ]]`.

- [ ] **Step 3: Verify the recipes parse**

Run: `just --justfile scripts/build.just --list | grep darwin`
Expected: a line each for `build-darwin-amd64-app`, `build-darwin-arm64-app`, `build-darwin-app-all`. If `--list` errors with "could not find recipe", the justfile syntax is broken — fix before committing.

- [ ] **Step 4: Smoke-test the bundling step (no cross-compile, just the wrap)**

First, ensure `dist/dtui-darwin-amd64` exists (from Task 4 or a previous run). Then:

```bash
just build::build-darwin-amd64-app
```

Expected: output includes `==> packaged ../dist/dtui-darwin-amd64.app` and a `find` listing.

- [ ] **Step 5: Verify the bundle**

```bash
file dist/dtui-darwin-amd64.app/Contents/MacOS/dtui
ls -la dist/dtui-darwin-amd64.app/Contents/
```

Expected:
- `Mach-O 64-bit executable x86_64` for the inner binary.
- `Contents/` has `Info.plist`, `MacOS/`, `PkgInfo`, `Resources/`.

- [ ] **Step 6: Commit**

```bash
git add scripts/build.just
git commit -m "feat(macos): add build-darwin-{amd64,arm64}-app recipes

Wraps the darwin nocgo binary in a .app bundle using
scripts/macos/build-app.sh. Aggregator build-darwin-app-all runs
both architectures."
```

---

### Task 8: Update README with macOS build instructions

**Files:**
- Modify: `README.md` (Build & run table + a new macOS subsection)

- [ ] **Step 1: Add rows to the build table**

In the `## Build & run` table, after the existing `just build` row, add:

```markdown
| Build macOS binary (amd64) | `just build-darwin-amd64-nocgo` |
| Build macOS binary (arm64 / Apple Silicon) | `just build-darwin-arm64-nocgo` |
| Build macOS `.app` bundle (amd64) | `just build-darwin-amd64-app` |
| Build macOS `.app` bundle (arm64) | `just build-darwin-arm64-app` |
```

- [ ] **Step 2: Add a macOS subsection**

After the existing build instructions (and before the `## Configuration` section), add:

```markdown
### macOS

dtui runs as a terminal application on macOS with Docker
Desktop for Mac or Podman machine. Build from Linux (cross-compile)
or from macOS natively:

```bash
just build-darwin-arm64-nocgo     # or -amd64 for Intel Macs
just build-darwin-arm64-app       # wrap into dist/dtui-darwin-arm64.app
```

Default socket paths (used when no `-H` flag is given and no
explicit `runtime.Connections` are configured):

- Docker Desktop for Mac: `~/.docker/run/docker.sock`
- Podman machine: discovered with `podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}'`

The `.app` bundle is a macOS packaging artifact. Because dtui is a TUI,
launch the standalone binary or `dist/dtui-darwin-arm64.app/Contents/MacOS/dtui`
from Terminal for interactive use. If your installation has a custom socket
path, override via `-H unix:///path/to/socket.sock` or via
`runtime.connections` in the config file.
```

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: document macOS build recipes and default socket paths

Add build-darwin-{amd64,arm64}-{nocgo,app} to the build table
and a new macOS subsection explaining Docker Desktop default paths,
dynamic Podman machine socket discovery, and Terminal-based .app use."
```

---

## Future Work (NOT in this plan — recorded for follow-up)

These are deliberately deferred. They each warrant their own plan.

| Item | Why deferred | Scope estimate |
|---|---|---|
| **macOS cgo / gpgme / Podman bindings** | `scripts/musl-gpgme/` is Linux-only. Adding darwin cgo requires an osxcross toolchain OR a macOS host runner, and native gpgme builds (via brew) — a multi-day investigation. | 3-5 days + Mac CI runner |
| **GitHub Actions macOS runner** | Verify-build-matrix currently only runs linux. Adding `macos-latest` runner lets us actually run the darwin binary (`./dtui-darwin-arm64 --version`). | Half day |
| **Homebrew formula** | Tap repo + formula.rb referencing the GitHub release tarball. Trivial once releases publish darwin assets. | Half day, after release workflow ships darwin |
| **Code signing + notarization** | Needs Apple Developer ID ($99/yr), notarization credentials, CI integration with `xcrun notarytool`. Notarization adds a per-release minute of wall time. | 1-2 days, gated on credentials |
| **Universal binary (lipo)** | Combine amd64 + arm64 Mach-O into one file via `lipo -create`. Useful for distribution but not for development. | 1 hour once single-arch pipeline is stable |
| **App icon (`icon.icns`)** | No PNG/icns asset exists in the repo. Either generate from a source SVG or skip (macOS will show a generic icon). | Asset-dependent |

---

## Self-Review Checklist

**1. Spec coverage:**
- ✅ Phase 1: socket paths for Docker Desktop + Podman — Tasks 1-3
- ✅ Phase 1: verify darwin nocgo still builds — Task 4
- ✅ Phase 2: Info.plist — Task 5
- ✅ Phase 2: bundle script — Task 6
- ✅ Phase 2: justfile integration — Task 7
- ✅ Phase 2: README — Task 8
- ✅ Future work recorded (not in this plan): cgo, CI, Homebrew, signing

**2. Placeholder scan:**
- No "TBD"/"TODO"/"implement later" — every step has the actual code/commands.
- Every file path is exact.
- Every commit command is the actual `git` invocation.

**3. Type consistency:**
- `RuntimeType`, `ConnectionSpec`, `BuildConnections`, `ConnectionKey`, and other shared APIs stay in `connection.go` with no build tag, so darwin does not lose the common runtime API.
- `DefaultDockerSocket` and `defaultPodmanSocket` change from `const` to `var` only on darwin — documented in Task 2 Step 5 that any caller depending on const-ness would surface as a build error. The known callers in `internal/runtimeinit/` only read these values, never take their address or use them in `const` expressions, so the change is safe.
- `defaultPodmanUserEndpoint` keeps the same signature `(uid int) string` on both platforms and returns endpoint URIs, not raw filesystem paths.
- `internal/data/runtime/podman` discovery is split too, so `runtimeinit.firstLivePodmanSocket()` does not keep probing Linux `/run/...` paths on macOS.

**4. Constraint compliance:**
- No upstream dependency source modified — all changes are in dtui-owned packages.
- CGO_ENABLED=0 for darwin in this plan — `scripts/musl-gpgme/` is untouched.
- Linux build matrix unchanged.
- `.app` output is architecture-named (`dtui-darwin-amd64.app`, `dtui-darwin-arm64.app`) so amd64 and arm64 bundles do not overwrite each other.
