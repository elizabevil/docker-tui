# dtui

A keyboard-driven terminal UI for managing Docker and Podman from a single
interface. Browse containers, images, volumes, networks and Compose
projects; inspect resources, follow logs, exec into running containers,
and audit every action — all without leaving the keyboard.

Built with Go, [Bubble Tea v2](https://charm.land/bubbletea/v2), and
[Lip Gloss](https://github.com/charmbracelet/lipgloss).

> The CLI binary is `dtui` (built to `dist/dtui` by default).

---

## Features

- **Containers** — list, start, stop, restart, pause / unpause, rename,
  top, ports, kill, delete, logs, stats, inspect, exec
- **Images** — list, pull, prune, tag, push, save, load, delete, structured
  detail with manifest variants and layer history
- **Volumes & Networks** — create, view, prune, delete with cross-resource
  references
- **Compose projects** — aggregated from container labels; start, stop,
  down, log streaming
- **Runtime events** — subscribed and merged per resource type, with
  reconnect-and-poll fallback
- **Instant filtering & log search** — apply filters as you type; jump
  between log matches with `n` / `Ctrl+N`
- **Configurable keymap** — override per-action keys in YAML; Help and
  Footer always reflect the effective bindings
- **User-action audit** — typed traces projected to toast, footer, and
  daily JSONL audit logs
- **Localized UI** — English and Simplified Chinese out of the box, with
  extensible JSONC-based translation files
- **Runtime auto-discovery** — Docker and Podman sockets detected
  automatically; connect via TLS with per-connection verification

Multi-host workflows, image transfers, full batch operations and mouse
interaction are still in progress. See [docs/requirements.md](docs/requirements.md)
for the full status.

---

## Table of contents

- [Requirements](#requirements)
- [Install](#install)
- [Build & run](#build--run)
- [Configuration](#configuration)
- [Keyboard reference](#keyboard-reference)
- [Localization](#localization)
- [Audit log](#audit-log)
- [Architecture](#architecture)
- [Development & testing](#development--testing)
- [Documentation](#documentation)
- [License](#license)

---

## Requirements

| Dependency | Notes |
|------------|-------|
| Go | Use the version pinned in [`go.mod`](go.mod) |
| Docker **or** Podman daemon | Current user must have access to the socket |
| [`just`](https://github.com/casey/just) *(optional)* | Convenience runner for build / test recipes |

Typical Linux socket paths:

```text
Docker: unix:///var/run/docker.sock
Podman: unix:///run/user/1000/podman/podman.sock
```

---

## Install

```bash
# Clone
git clone https://github.com/elizabevil/docker-tui.git
cd docker-tui

# Build (CGO disabled for a static binary)
CGO_ENABLED=0 go build -o ./dist/dtui ./cmd/docker-tui

# Run
./dist/dtui
```

If you prefer the recipe runner:

```bash
just build
just run
```

The build recipes embed a product version from `DTUI_VERSION` and default to
`0.2.0`. Tagged releases publish Linux / macOS / Windows artifacts named
`dtui-*`.

---

## Build & run

| Task | Command |
|------|---------|
| Build binary | `just build` |
| Override embedded version | `DTUI_VERSION=v0.2.0 just build` |
| Run (Docker) | `just run` |
| Run (Podman) | `just run-podman` |
| List built-in themes | `./dist/dtui --list-themes` |
| Show version | `./dist/dtui --version` |

CLI flags:

```text
-f, --config PATH    Override config file path
-t, --theme NAME     Theme name (default, dark, light, nord, dracula, solarized)
-L, --lang zh|en     Interface language
-p, --podman         Connect to local Podman socket instead of Docker
    --list-themes    Print available themes and exit
-v, --version        Print version and exit
-H, --host URI       Daemon socket or remote endpoint URI
```

The program always registers local Docker and Podman candidates. By
default it prefers `local-docker`; if Docker is unreachable and local
Podman is available it falls back automatically and shows a notice.
Override the default with `runtime.default` in the config file, or
explicitly force Podman with `--podman`.

Remote connections require per-connection TLS configuration — see
[`runtime.connections`](#configuration).

---

## Configuration

Default path:

```text
~/.config/docker-tui/config.yml
```

Pass `--config` to use a different YAML file. Missing fields fall back
to the embedded defaults in
[`internal/data/config/default.jsonc`](internal/data/config/default.jsonc).

Minimal example:

```yaml
configVersion: 1

general:
  lang: zh           # en | zh
  sizeFormat: binary # binary | si

runtime:
  default: local-docker
  discovery:
    localDocker: true
    localPodman: true
  health:
    intervalSec: 3
    timeoutSec: 2
    failureThreshold: 2
  connections:
    - name: remote-docker
      driver: docker
      endpoint: tcp://docker.example.com:2376
      apiVersion: ""   # auto-negotiate when empty
      tls:
        enabled: true
        verify: true
        insecureSkipVerify: false
        caFile:    /etc/docker/certs/ca.pem
        certFile:  /etc/docker/certs/cert.pem
        keyFile:   /etc/docker/certs/key.pem

logs:
  since: 1h
  tail: "200"
  timestamps: false

keymap:
  help:           [f1]
  containerStart: [s]
  containerStop:  [ctrl+s]
  containerPause: [p]
  imagePull:      [ctrl+p]
  imageTag:       [ctrl+t]
  imagePush:      [ctrl+u]
  imageSave:      [ctrl+e]
  imageLoad:      [ctrl+l]
```

> Only the `runtime` schema is accepted. Legacy `docker.host`,
> `docker.tlsVerify`, `docker.tlsCertPath` and `general.runtime` keys
> are rejected on load.

All fields are documented in
[`internal/data/config/default.jsonc`](internal/data/config/default.jsonc)
and [`internal/data/config/types.go`](internal/data/config/types.go).

---

## Keyboard reference

### Global

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Switch panel |
| `j` / `k` / `↑` / `↓` | Move cursor |
| `Enter` | Open item or execute primary action |
| `Esc` | Back; double-tap on main view to quit |
| `/` | Filter resources; log search in the Logs page |
| `:` | Command palette |
| `?` / `F1` | Open Help |
| `r` | Refresh all resources |
| `F2` | Switch runtime connection |
| `Space` | Toggle mark mode |
| `Ctrl+D` | Delete current resource |

### Per-resource

| Panel | Keys |
|-------|------|
| Containers | `s` start · `Ctrl+S` stop · `Ctrl+R` restart · `p` pause/unpause · `Ctrl+K` kill · `l` logs · `e` exec · `d` detail · `m` stats |
| Images | `Ctrl+P` pull · `p` prune · `d` detail · `Ctrl+D` delete |
| Volumes | `Enter` expand · `d` detail · `Ctrl+D` delete |
| Networks | `d` detail · `Ctrl+D` delete |
| Compose | `s` start · `Ctrl+S` stop · `Ctrl+D` down · `l` logs · `d` detail |

Filters apply as you type; `Enter` commits and exits edit, double `Esc`
within 5 s clears the filter. In Logs, `n` / `Ctrl+N` jump between
search matches.

The container command palette accepts `:rename`, `:top`, and `:port`.
`Top` requires a running container and supports `j`/`k` scroll plus `r`
refresh. `Port` shows structured container ports, protocols, host IP and
host port. Batch pause / unpause skip state-incompatible containers and
summarize succeeded / skipped / failed counts.

For the full guide see [docs/navigation.md](docs/navigation.md). The
in-app Help and Footer are always authoritative for the current
bindings.

---

## Localization

Translations live as JSONC files under
[`internal/data/i18n/lang/`](internal/data/i18n/lang/):

```text
lang/
├── en.jsonc   # English (default)
├── zh.jsonc   # Simplified Chinese
└── ja.jsonc   # Japanese
```

Each file is a flat key → string map, with `{0}`, `{1}` placeholders for
arguments. Comments (`//` and `/* */`) are allowed and stripped at load
time via [`internal/utils/jsonc.go`](internal/utils/jsonc.go).

Adding a new language:

1. Copy an existing file (e.g. `en.jsonc`) to `lang/<code>.jsonc`.
2. Translate each value.
3. Add the language code to the `tags` map and `parseTag` switch in
   [`internal/data/i18n/lang.go`](internal/data/i18n/lang.go).
4. Rebuild.

Set the active language with `--lang`, the `general.lang` config field,
or the `LANG` environment variable. Missing keys fall back to English
and never panic.

---

## Audit log

Every user-initiated resource action (start, stop, exec, prune, runtime
switch …) emits a typed trace with the action, outcome, runtime, UI
context, and strongly-typed resource target.

Default path:

```text
~/.config/docker-tui/logs/audit-YYYY-MM-DD.jsonl
```

Audit events also drive the toast notification and footer action line.
Pure navigation, filtering and background refresh are not audited.

---

## Architecture

A stable three-section layout — status bar on top, resource workspace
in the middle, query / message / footer row at the bottom — is shared
by all resource pages and overlays. Keyboard input is resolved against
the active context, dispatched as a `tea.Cmd`, and the resulting
message flows back through a single Update loop.

```text
keyboard input
  -> action registry and context resolver
  -> tea.Cmd
  -> resource/action message
  -> state update + audit projection
  -> render
```

Source layout:

```text
cmd/docker-tui/        CLI entrypoint and bootstrap
internal/data/         config, runtime adapters, i18n, audit
internal/tui/          top-level dispatcher + theme/layout bootstrap
internal/tui/state/    application state and message types
internal/tui/keys/     action registry and effective key bindings
internal/tui/keyboard/ interaction and business commands
internal/tui/update/   tea.Msg dispatch + per-message handlers
internal/tui/ui/       pages, components, layout rendering
internal/utils/        shared helpers (JSONC, formatting, TLS, URIs)
test/                  cross-package tests, diagnostics, benchmarks, integration
```

JSONC parsing is centralised in
[`internal/utils/jsonc.go`](internal/utils/jsonc.go) so every config,
theme, table profile and translation file shares one comment-tolerant
loader backed by both `encoding/json` and `bytedance/sonic`.

---

## Development & testing

List all recipes:

```bash
just --list
```

Common checks:

```bash
just check              # go vet + full Go test suite
just test               # all Go tests
just test-unit          # cmd/ and internal/ unit tests only
just test-integration   # build and run container-engine integration tests
just bench              # serialization benchmarks
```

Direct equivalents:

```bash
go vet ./...
go test ./...
go test ./cmd/... ./internal/...
go test -tags=integration ./test/integration/...
go test -bench=. -benchmem ./test/benchmarks/...
```

Integration tests may pull images and create temporary containers
prefixed with `dtui-test-`; the runner cleans them up on exit. See
[test/README.md](test/README.md) for details.

---

## Documentation

| Document | Purpose |
|----------|---------|
| [docs/architecture.md](docs/architecture.md) | Module layering, runtime adapters, message flow |
| [docs/requirements.md](docs/requirements.md) | Functional & non-functional requirements, status |
| [docs/project-structure.md](docs/project-structure.md) | Repository conventions and directory map |
| [docs/navigation.md](docs/navigation.md) | Full key, command and mode reference |
| [docs/ui-design.md](docs/ui-design.md) | UI design system (layout, components, tables, styles) |
| [docs/i18n.md](docs/i18n.md) | Internationalization (translations, terminal width, lang switching) |
| [docs/bugfix-requirements.md](docs/bugfix-requirements.md) | Known issues and remediation notes |
| [design/current-design.md](design/current-design.md) | Full authoritative UI design document |
| [design/ui-i18n-design.md](design/ui-i18n-design.md) | i18n design with future implementation plans |
| [design/unified-runtime-driver.md](design/unified-runtime-driver.md) | Docker / Podman driver unification |
| [docs/README.md](docs/README.md) | Full documentation index |

---

## License

This project does not currently declare a license. Treat the source as
**all rights reserved** until a `LICENSE` file is added.
