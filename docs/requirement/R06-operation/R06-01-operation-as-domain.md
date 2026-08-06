# R06 Operation as First-Class Domain Concept

## Background

The container-action feature currently mixes three concerns into one theme
slot:

1. **Style** — window background / border / form input / Confirm·Cancel
   colours consumed by `FormDialog`.
2. **Action** — the eight `ActionContainer*` key bindings and their
   respective Go handlers (`openTopView`, `openPortDetail`, `doContainerDiff`,
   `doContainerWait`, `openContainerCopyForm`, `openContainerUpdateForm`,
   `openContainerExportForm`, `openContainerCommitForm`).
3. **Composition** — the manual `[]ActionItem` list in
   `internal/tui/actionbar/registry.go` that wires each Action into the
   Action Bar.

The previous round renamed the theme slot to `containerOperation` and
applied it only to the four form-based actions. This left three problems
unresolved:

- **Name scope**: `containerOperation` is hardcoded to one resource. The
  Image panel already needs its own chrome (e.g. `Image History` action
  has a dedicated page), and Volume / Network / Compose will follow.
- **Mode abstraction**: only `Form` mode was wired. `Top` / `Port` /
  `Filesystem Diff` open full pages (`ModeTop`, `ModeDetail`) and `Wait`
  is purely async (toast-only). They share no rendering pipeline with
  `Form`.
- **Configuration locality**: action metadata is duplicated across Go
  code (`ActionItem` literals, `switch` cases in `keyboard/actions.go`)
  rather than living in a single declarative file.

## Goal

Promote `Operation` to a first-class domain concept whose structure,
lifecycle and theme contract are uniform across all resource scopes.

```
Operation
  ├─ Kind:    unique identifier (top, port, copy, ...)
  ├─ Scope:   which resource it acts on (container, image, volume, ...)
  ├─ Mode:    how its body renders (Form | Page | Async)
  ├─ Window:  window chrome from theme.action.<scope>.window
  ├─ Body:    mode-specific renderer
  └─ Actions: mode-specific (Form → Confirm + Cancel; Async → Cancel)
```

## Non-Goals (this iteration)

- Replacing the eight hardcoded `openXxx` / `doXxx` Go handlers with a
  single dispatcher. That change lives in a follow-up because it requires
  new state fields (`m.Operation`) and cross-package refactors beyond
  the theme layer.
- Loading Operation metadata from JSONC. Same reason: requires a new
  loader and registry wiring, separate concern.
- Wiring `Page` / `Async` modes to actual renderers. Same reason.

This iteration only migrates the theme contract from
`theme.containerOperation.{window,formInput,confirm,cancel}` to
`theme.action.<scope>.{window,formInput,confirm,cancel}` with two scopes
(`container`, `image`) seeded. The renderer side continues to apply the
container scope for the four form-based actions as before.

## Design

### Layer separation

| Layer | File family | Responsibility |
|---|---|---|
| Skin | `themes/*.jsonc` | Colours, borders, tokens — only |
| Skeleton | `operations/registry.jsonc` (future) | Kind ↔ KeyAction mapping |
| Skeleton | `operations/scopes/<scope>.jsonc` (future) | Per-scope Operation metadata |
| Render | `internal/tui/ui/widget/dialog/form.go` | Apply Skin via Mode dispatch |

Themes do **not** reference operations; operations do **not** reference
themes. They are joined only through the Go-side style resolution in
`internal/tui/ui/component/styles_load.go`.

### Theme schema (this iteration)

```jsonc
{
  "theme": {
    "palette": { ... },
    "action": {
      "container": {
        "window":    { "background": "...", "border": "..." },
        "formInput": { "foreground": "...", "background": "..." },
        "confirm":   { "foreground": "...", "background": "..." },
        "cancel":    { "foreground": "...", "background": "..." }
      },
      "image": {
        "window":    { "background": "...", "border": "..." },
        "formInput": { "foreground": "...", "background": "..." },
        "confirm":   { "foreground": "...", "background": "..." },
        "cancel":    { "foreground": "...", "background": "..." }
      }
    }
  }
}
```

`action.<scope>` is the **skin contract** for any Operation whose `Scope`
equals `<scope>`. The `image` scope is seeded but currently has no
form-based action that consumes it; it exists so that adding
`Image Save` / `Image Load` forms later only requires adding the action
in Go + adding the scope's section to a theme.

### Style resolution

`internal/tui/ui/component/styles_load.go` exposes one StyleName per
slot per scope:

```go
StyleActionContainerWindow         = "actionContainerWindow"
StyleActionContainerWindowBorder   = "actionContainerWindowBorder"
StyleActionContainerFormInput      = "actionContainerFormInput"
StyleActionContainerConfirm        = "actionContainerConfirm"
StyleActionContainerCancel         = "actionContainerCancel"
StyleActionImageWindow             = "actionImageWindow"
StyleActionImageWindowBorder       = "actionImageWindowBorder"
StyleActionImageFormInput          = "actionImageFormInput"
StyleActionImageConfirm            = "actionImageConfirm"
StyleActionImageCancel             = "actionImageCancel"
```

Per-scope constants are explicit rather than parameterised so the
existing `StyleName → styleRef` lookup stays a flat table; parameterising
through scope would require a map-based refactor in `globalStyleRefs`
that is out of scope here.

### Form renderer behaviour

`FormDialog` continues to apply the **container scope** for the four
form-based `FormKind` values (`FormContainerCopy`, `FormContainerUpdate`,
`FormContainerExport`, `FormContainerCommit`). When image forms are
added later, a parallel branch selecting the image scope is the
one-line extension point — no theme or constants work is required.

The `theme.containerOperation.*` keys are removed; any user theme file
still carrying them is migrated by this iteration's JSONC rewrite.

### Light-theme test plan

1. Override `theme.action.container.window.border` to a value distinct
   from the default (`primary` token → explicit hex) in
   `themes/light.jsonc`.
2. Override `theme.action.container.confirm.foreground` similarly.
3. Run `go test ./internal/data/config/...` to confirm the patch is
   parsed and resolved.
4. Render `FormDialog` with `FormKind = FormContainerCopy` against the
   `light` theme in a unit test and assert the rendered ANSI output
   contains the light-theme-specific border colour SGR.

Other themes are not customised this round; they pick up the defaults
from `config.DefaultTheme()` and continue to render unchanged.

## Follow-ups

| ID | Scope | Notes |
|---|---|---|
| **R06-02 ✅** | Operations registry loader | `internal/data/config/{operations.go, operations_loader.go, operations_cache.go, defaults/operations/registry.jsonc, defaults/operations/scopes/{container,image}.jsonc}` — `LoadOperations()` parses registry + scope files, merges, validates; `CachedLoadOperations()` shares one `sync.Once` across actionbar / keyboard / future callers. |
| **R06-03 ✅** | Single dispatcher | `internal/tui/actionbar/registry.go` rebuilt against the loaded `Operations`; `internal/tui/keyboard/operation_dispatch.go` introduces `dispatchOperation(action, m)` that resolves KeyAction → `OperationSpec` → Mode-specific handler; `keyboard/actions.go` loses 8 hardcoded `case` branches (Rename / Top / Port / Diff / Wait / Copy / Update / Export / Commit) and falls through to the dispatcher. |
| R06-04 | State consolidation | Migrate `m.Form + m.Processes + m.ContainerWait + m.Detail` into a single `m.Operation` discriminated union. Rename currently still routes through `ModeRename + Dialog`; once R06-04 lands the dispatcher arm collapses too. |
| R06-05 | Page / Async mode renderers | Wire `Mode=Page` to `ModeTop` / `ModeDetail` page chrome and `Mode=Async` to the existing toast / status pipeline, both consuming `theme.action.<scope>.window`. |