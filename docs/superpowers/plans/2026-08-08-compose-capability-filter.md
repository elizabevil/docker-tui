# Compose Capability Filter + Pod Column Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align the compose two-pane view with `docs/discussion/compose-panes-semantics-design.md` §3 (设计原则) by deleting out-of-scope operations (L1 filter), introducing capability-driven UI rendering (L2 filter), and adding the Pod column visibility per §4.3 decision C — all in the latest scheme with zero compatibility shims.

**Architecture:** Capability gating is pushed into the YAML-driven Operations registry so the action bar filters items declaratively; view.go queries `Engine.Capabilities()` directly for the Pod indicator + column visibility. Out-of-scope operations (`build`/`pull`/`push`/`group_*`) are deleted in full (YAML, Go functions, tests) — no dead code, no fallback paths.

**Tech Stack:** Go 1.x, Bubble Tea v2, JSONC config; existing `runtime.CapabilitySet` / `runtime.Capability` primitives.

**Reference spec:** `docs/discussion/compose-panes-semantics-design.md` (already on this branch as design source-of-truth).

## Global Constraints

- **Driver layer off-limits** — `internal/driver/podman/**` and `internal/driver/docker/**` MUST NOT be modified in any task. All changes live in `internal/data/**`, `internal/tui/**`, and YAML config.
- **No backward compatibility** — every deletion is a clean cut. No `if legacy { ... } else { ... }` branches, no deprecated fields, no alias fields kept for old configs.
- **Latest design only** — implementation follows `compose-panes-semantics-design.md` §3.1–§3.6. Pre-§3 patterns (e.g. inline `Capabilities().Supports` checks in action handlers) are removed.
- **TDD** — every new behavior has a failing test written first; the test is the deliverable's contract.
- **Frequent commits** — one commit per task, conventional prefix (`feat:`, `chore:`, `refactor:`, `test:`).

---

## File Structure

```
CREATE
  (none — all changes modify existing files)

MODIFY — domain / config layer
  internal/data/config/operations.go                                   # add RequiresCapabilities field
  internal/data/config/defaults/operations/scopes/compose.jsonc       # delete 6 operations
  internal/data/config/defaults/tables/compose.jsonc                   # add pod column; rename pods→containers
  internal/data/i18n/lang/en.jsonc                                     # add table.pod key
  internal/data/i18n/lang/zh.jsonc                                     # add table.pod key
  internal/data/i18n/lang/ja.jsonc                                     # add table.pod key

MODIFY — TUI layer
  internal/tui/actionbar/registry.go                                   # capability filter in evaluateEnabled
  internal/tui/keyboard/compose_action.go                              # delete 9 functions + 2 helpers
  internal/tui/keyboard/compose_action_test.go                         # delete tests for removed code
  internal/tui/ui/pages/compose/view.go                                # Pod indicator + column filter + rename

ADD — tests
  internal/tui/actionbar/registry_test.go                              # extend: capability filter cases
  internal/tui/ui/pages/compose/view_test.go                           # NEW: Pod indicator + column visibility
```

Each task touches 1–3 files. Each task ends with one `git commit`.

---

## Task 1: Add `RequiresCapabilities` field to OperationSpec

**Files:**
- Modify: `internal/data/config/operations.go:159-242` (extend `OperationSpec` struct)
- Modify: `internal/data/config/operations.go:103-133` (extend `allRequirements` doc comment, NOT the set itself)

**Interfaces:**
- Consumes: `OperationSpec` (existing struct)
- Produces: `OperationSpec.RequiresCapabilities []string` — optional list of capability names. Each entry is a `runtime.Capability` value (e.g. `"compose.pod_scope"`). The action bar evaluator reads this list and hides the operation unless `Engine.Capabilities().Supports(cap)` is true for every entry.

### Step 1: Write the failing test

Add to `internal/data/config/operations_loader_test.go` (or create if absent):

```go
func TestRequiresCapabilitiesFieldRoundtrips(t *testing.T) {
    src := []byte(`
        {
          "scope": "compose",
          "operations": [
            {
              "kind": "future_capability_demo",
              "action": "compose.demo",
              "mode": "async",
              "async": { "body": "wait" },
              "label": "demo",
              "description": "demo op",
              "requires": ["engine"],
              "requiresCapabilities": ["compose.pod_scope"]
            }
          ]
        }
    `)
    var doc operationsScopeDocument
    if err := decodeJSONCStrict(src, &doc); err != nil {
        t.Fatalf("decode: %v", err)
    }
    spec := doc.Operations[0]
    if got, want := spec.RequiresCapabilities, []string{"compose.pod_scope"}; !reflect.DeepEqual(got, want) {
        t.Fatalf("RequiresCapabilities = %v, want %v", got, want)
    }
}
```

Add `"reflect"` to the test file's imports if not present.

### Step 2: Run test to verify it fails

Run: `go test ./internal/data/config/ -run TestRequiresCapabilitiesFieldRoundtrips -v`
Expected: FAIL — `Operations` has no field `RequiresCapabilities`, `DisallowUnknownFields` rejects it.

### Step 3: Add the field to the struct

Edit `internal/data/config/operations.go`, inside `OperationSpec` (after line 212, `DisabledWhen` field):

```go
    // RequiresCapabilities is a list of runtime.Capability identifiers
    // (e.g. "compose.pod_scope"). The action bar evaluator hides the
    // Operation unless Engine.Capabilities().Supports(c) returns true
    // for every entry. Distinct from Requires / DisabledWhen, which
    // gate on UI state predicates; capability gating is a runtime
    // capability check, evaluated against the live Engine.
    //
    // Capability identifiers are NOT a closed universe at this layer:
    // the loader accepts any string. Validation happens at evaluation
    // time (CapabilitySet lookup is a map miss → Unsupported), so a
    // typo in YAML fails closed (item hidden) instead of panicking.
    RequiresCapabilities []string `json:"requiresCapabilities,omitempty"`
```

### Step 4: Run test to verify it passes

Run: `go test ./internal/data/config/ -run TestRequiresCapabilitiesFieldRoundtrips -v`
Expected: PASS.

### Step 5: Run the full config test suite

Run: `go test ./internal/data/config/...`
Expected: all PASS.

### Step 6: Commit

```bash
git add internal/data/config/operations.go internal/data/config/operations_loader_test.go
git commit -m "feat(config): add RequiresCapabilities field to OperationSpec"
```

---

## Task 2: Add capability filter to action bar evaluator

**Files:**
- Modify: `internal/tui/actionbar/registry.go:122-134` (`evaluateEnabled` function)
- Modify: `internal/tui/actionbar/registry.go:140-189` (`positiveSatisfied` — no change required, but doc-comment may need a note)

**Interfaces:**
- Consumes: `config.OperationSpec.RequiresCapabilities []string` (from Task 1); `runtime.CapabilitySet` via `m.Connection.Engine.Capabilities()`
- Produces: `evaluateEnabled` returns `false` when any required capability is not `Available` / `Degraded`. Pure function of `OperationSpec` + `*state.AppModel`.

### Step 1: Write the failing test

Append to `internal/tui/actionbar/registry_test.go`:

```go
type capEngine struct {
    runtimeapi.Engine // embed for interface satisfaction; unused methods return zero
    caps runtimeapi.CapabilitySet
}

func (e *capEngine) Capabilities() runtimeapi.CapabilitySet { return e.caps }

// TestRequiresCapabilitiesHidesUnsupported verifies an Operation with
// requiresCapabilities=[X] is hidden when X is not in the engine's
// CapabilitySet, and visible when X is Available.
func TestRequiresCapabilitiesHidesUnsupported(t *testing.T) {
    ops := &config.Operations{
        ByKind: map[string]config.OperationSpec{
            "compose::demo": {
                Scope:                 config.OperationScopeCompose,
                Kind:                  "demo",
                Action:                "compose.demo",
                Mode:                  config.OperationModeAsync,
                Async:                 &config.AsyncBody{Body: "wait"},
                Label:                 "Demo",
                Description:           "demo",
                InActionBar:           true,
                RequiresCapabilities: []string{"compose.pod_scope"},
            },
        },
        ByScope: map[config.OperationScope][]config.OperationSpec{
            config.OperationScopeCompose: {{
                Scope: config.OperationScopeCompose, Kind: "demo",
                Action: "compose.demo", Mode: config.OperationModeAsync,
                Async: &config.AsyncBody{Body: "wait"},
                Label: "Demo", Description: "demo", InActionBar: true,
                RequiresCapabilities: []string{"compose.pod_scope"},
            }},
        },
    }

    // engine WITHOUT the capability → hidden
    m1 := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
    m1.Connection.Engine = &capEngine{caps: runtimeapi.CapabilitySet{}}
    if got := VisibleItemsWithOps(m1, ops); len(got) != 0 {
        t.Fatalf("expected hidden, got %#v", got)
    }

    // engine WITH the capability → visible
    m2 := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
    m2.Connection.Engine = &capEngine{caps: runtimeapi.CapabilitySet{
        runtimeapi.CapabilityComposePodScope: {Support: runtimeapi.Available},
    }}
    if got := VisibleItemsWithOps(m2, ops); len(got) != 1 {
        t.Fatalf("expected visible, got %#v", got)
    }
}
```

To make `VisibleItemsWithOps` callable, refactor `VisibleItems` to delegate:

```go
// VisibleItemsWithOps is the same as VisibleItems but takes an explicit
// Operations snapshot. Tests use this to bypass the YAML cache.
func VisibleItemsWithOps(m *state.AppModel, ops *config.Operations) []ActionItem {
    scope, ok := scopeForPanel(m.Navigation.ActivePanel)
    if !ok {
        return nil
    }
    raw := buildItemsForScope(m, ops, scope)
    filter := strings.ToLower(strings.TrimSpace(m.Navigation.ActionBar.Filter))
    if filter == "" {
        return raw
    }
    out := make([]ActionItem, 0, len(raw))
    for _, item := range raw {
        searchable := item.Label + " " + item.Description
        if strings.Contains(strings.ToLower(searchable), filter) {
            out = append(out, item)
        }
    }
    return out
}
```

### Step 2: Run test to verify it fails

Run: `go test ./internal/tui/actionbar/ -run TestRequiresCapabilitiesHidesUnsupported -v`
Expected: FAIL — `buildItemsForScope` does not consult `RequiresCapabilities`.

### Step 3: Add the capability check

Edit `internal/tui/actionbar/registry.go`, inside `buildItemsForScope` (between the existing `if !spec.InActionBar` check and the `items = append`):

```go
        // Capability gate: hide if the live Engine lacks any required
        // capability. Engine may be nil during early init; in that
        // case capability requirements are treated as unsatisfied
        // (fail closed) so a not-yet-connected Engine never shows an
        // operation it might not support.
        if m.Connection.Engine != nil {
            caps := m.Connection.Engine.Capabilities()
            for _, want := range spec.RequiresCapabilities {
                if !caps.Supports(runtimeapi.Capability(want)) {
                    goto skip
                }
            }
        } else if len(spec.RequiresCapabilities) > 0 {
            goto skip
        }
```

Replace the closing `items = append(...)` block with a labelled block so the `goto skip` works:

```go
        items = append(items, ActionItem{...})
        continue
    skip:
    }
```

Refactor `buildItemsForScope` to use a labelled `for ... range specs` loop with `continue` to a `skip:` label at the bottom of the loop body. Avoid `goto`; restructure as:

```go
func buildItemsForScope(m *state.AppModel, ops *config.Operations, scope config.OperationScope) []ActionItem {
    specs := ops.ForScope(scope)
    items := make([]ActionItem, 0, len(specs))
capsLoop:
    for _, spec := range specs {
        if !spec.InActionBar {
            continue
        }
        if !capabilitiesSatisfied(spec, m) {
            continue capsLoop
        }
        items = append(items, ActionItem{
            Label:       spec.Label,
            Action:      keys.KeyAction(spec.Action),
            Description: spec.Description,
            Disabled:    !evaluateEnabled(spec, m),
        })
    }
    return items
}

// capabilitiesSatisfied returns true when every entry in
// spec.RequiresCapabilities is supported by the live Engine. An empty
// list is trivially satisfied. Engine==nil with non-empty requirements
// fails closed.
func capabilitiesSatisfied(spec config.OperationSpec, m *state.AppModel) bool {
    if len(spec.RequiresCapabilities) == 0 {
        return true
    }
    if m == nil || m.Connection.Engine == nil {
        return false
    }
    caps := m.Connection.Engine.Capabilities()
    for _, want := range spec.RequiresCapabilities {
        if !caps.Supports(runtimeapi.Capability(want)) {
            return false
        }
    }
    return true
}
```

### Step 4: Run test to verify it passes

Run: `go test ./internal/tui/actionbar/ -run TestRequiresCapabilitiesHidesUnsupported -v`
Expected: PASS.

### Step 5: Run the full actionbar test suite

Run: `go test ./internal/tui/actionbar/...`
Expected: all PASS.

### Step 6: Commit

```bash
git add internal/tui/actionbar/registry.go internal/tui/actionbar/registry_test.go
git commit -m "feat(actionbar): hide operations whose required capabilities are unsupported"
```

---

## Task 3: Delete `project_build`, `project_pull`, `project_push` (YAML + Go)

**Files:**
- Modify: `internal/data/config/defaults/operations/scopes/compose.jsonc` (delete the 3 entries, lines 109-135)
- Modify: `internal/tui/keyboard/compose_action.go` (delete `doComposeBuild`, `doComposePull`, `fetchComposePull`, `doComposePush`, `fetchComposePush`, `taggedComposeImages`, `composeProjectImageRefs` — confirmed via grep these are called nowhere else)
- Modify: `internal/tui/keyboard/compose_action_test.go` (remove tests that reference deleted functions)

**Interfaces:**
- Consumes: existing YAML schema; existing keyboard dispatch tables
- Produces: actions `compose_project.build` / `.pull` / `.push` no longer exist as `KeyAction` identifiers; any keyboard binding referencing them is unreachable

### Step 1: Verify no callers outside compose_action.go

Run:
```bash
grep -rn "doComposeBuild\|doComposePull\|doComposePush\|fetchComposePull\|fetchComposePush\|taggedComposeImages\|composeProjectImageRefs" internal/ --include='*.go' --exclude-dir=driver
```
Expected: only `internal/tui/keyboard/compose_action.go` matches. (If anything else matches, investigate — those callers must also be deleted or rewritten.)

### Step 2: Write the failing test

The test should verify that `Operations.ForScope(OperationScopeCompose)` no longer contains `project_build`, `project_pull`, `project_push`. Add to `internal/data/config/operations_loader_test.go`:

```go
func TestComposeScopeHasNoBuildPullPush(t *testing.T) {
    ops, err := LoadOperations()
    if err != nil { t.Fatal(err) }
    for _, spec := range ops.ForScope(OperationScopeCompose) {
        switch spec.Kind {
        case "project_build", "project_pull", "project_push":
            t.Fatalf("out-of-scope kind %q still in compose scope", spec.Kind)
        }
    }
}
```

### Step 3: Run test to verify it fails

Run: `go test ./internal/data/config/ -run TestComposeScopeHasNoBuildPullPush -v`
Expected: FAIL — kinds still present.

### Step 4: Delete YAML entries

Edit `internal/data/config/defaults/operations/scopes/compose.jsonc`. Delete the three entries starting at line 109 (`project_build`), 118 (`project_pull`), 127 (`project_push`). Remove cleanly — no comment placeholder.

### Step 5: Delete Go functions

Edit `internal/tui/keyboard/compose_action.go`. Delete the following blocks in their entirety:

- `doComposeBuild` (lines 668–686)
- `doComposePull` (lines 688–711)
- `fetchComposePull` (lines 713–734)
- `doComposePush` (lines 736–760)
- `fetchComposePush` (lines 762–795)
- `taggedComposeImages` (lines 797–817)
- `composeProjectImageRefs` (lines 819–843)

Also delete the doc comment block immediately preceding `doComposeBuild` (lines 668–671, "doComposeBuild implements `docker compose build` ...").

### Step 6: Delete or fix tests for removed functions

In `internal/tui/keyboard/compose_action_test.go`, search for test functions that reference any removed identifier. Use:
```bash
grep -n "doComposeBuild\|doComposePull\|doComposePush\|fetchComposePull\|fetchComposePush\|taggedComposeImages\|composeProjectImageRefs" internal/tui/keyboard/compose_action_test.go
```

If any test references them, delete the entire test function. (If the test is mixed and only some subtests reference them, delete only those subtests.)

### Step 7: Build + run tests

Run: `go build ./...`
Expected: exit 0.

Run: `go test ./internal/data/config/... ./internal/tui/keyboard/...`
Expected: all PASS, including the new `TestComposeScopeHasNoBuildPullPush`.

### Step 8: Commit

```bash
git add internal/data/config/defaults/operations/scopes/compose.jsonc \
        internal/tui/keyboard/compose_action.go \
        internal/tui/keyboard/compose_action_test.go \
        internal/data/config/operations_loader_test.go
git commit -m "chore(r08): delete compose build/pull/push (out of TUI scope per §3.5 L1)"
```

---

## Task 4: Delete `group_down`, `group_restart`, `group_exec` (YAML + Go)

**Files:**
- Modify: `internal/data/config/defaults/operations/scopes/compose.jsonc` (delete lines 227–252)
- Modify: `internal/tui/keyboard/compose_action.go` (delete `doComposeGroupDown`, `doComposeGroupRestart`, `doComposeGroupExec`, `composeGroupContainers` — confirmed via grep the helper is only used by the 3 deleted functions)

**Interfaces:**
- Produces: actions `compose_group.down` / `.restart` / `.exec` no longer exist. CoLocated metadata stays as a `ContainerSummary` field but is not exposed as a separate group action.

### Step 1: Verify no callers outside compose_action.go

Run:
```bash
grep -rn "doComposeGroupDown\|doComposeGroupRestart\|doComposeGroupExec\|composeGroupContainers" internal/ --include='*.go' --exclude-dir=driver
```
Expected: only `compose_action.go` matches.

### Step 2: Write the failing test

Append to `internal/data/config/operations_loader_test.go`:

```go
func TestComposeScopeHasNoGroupOperations(t *testing.T) {
    ops, err := LoadOperations()
    if err != nil { t.Fatal(err) }
    for _, spec := range ops.ForScope(OperationScopeCompose) {
        switch spec.Kind {
        case "group_down", "group_restart", "group_exec":
            t.Fatalf("group kind %q still in compose scope; CoLocated must not promote to group ops", spec.Kind)
        }
    }
}
```

### Step 3: Run test to verify it fails

Run: `go test ./internal/data/config/ -run TestComposeScopeHasNoGroupOperations -v`
Expected: FAIL.

### Step 4: Delete YAML entries

Edit `internal/data/config/defaults/operations/scopes/compose.jsonc`. Delete the three `group_*` entries (the block starting at line 227, including the `group_down` / `group_restart` / `group_exec` definitions).

### Step 5: Delete Go functions

Edit `internal/tui/keyboard/compose_action.go`. Delete in entirety:

- `doComposeGroupDown` (and its doc comment)
- `doComposeGroupRestart` (and its doc comment)
- `doComposeGroupExec` (and its doc comment)
- `composeGroupContainers` (helper — only used by the 3 above)

### Step 6: Delete or fix tests

Same procedure as Task 3 Step 6: grep for the deleted identifiers in `compose_action_test.go` and remove any test that references them.

### Step 7: Build + run tests

Run: `go build ./... && go test ./internal/data/config/... ./internal/tui/keyboard/...`
Expected: exit 0; all PASS.

### Step 8: Commit

```bash
git add internal/data/config/defaults/operations/scopes/compose.jsonc \
        internal/tui/keyboard/compose_action.go \
        internal/tui/keyboard/compose_action_test.go \
        internal/data/config/operations_loader_test.go
git commit -m "chore(r08): delete compose group_* operations (CoLocated is metadata, not a separate group)"
```

---

## Task 5: Add `pod` column to `pods_sub` YAML + i18n

**Files:**
- Modify: `internal/data/config/defaults/tables/compose.jsonc:90-114` (`pods_sub` array — append a new column entry)
- Modify: `internal/data/i18n/lang/en.jsonc` (add `table.pod`)
- Modify: `internal/data/i18n/lang/zh.jsonc` (add `table.pod`)
- Modify: `internal/data/i18n/lang/ja.jsonc` (add `table.pod`)

**Interfaces:**
- Produces: column key `"pod"` available in `pods_sub` profile; i18n key `table.pod` available for header translation

### Step 1: Write the failing test

Append to a new file `internal/tui/tables/compose_columns_test.go` (create if absent):

```go
package tables

import "testing"

func TestPodsSubHasPodColumn(t *testing.T) {
    tc := MustLoad("compose")
    cols := tc.Columns.Get("pods_sub")
    var found bool
    for _, c := range cols {
        if c.Key == "pod" {
            found = true
            if c.Header != "table.pod" {
                t.Fatalf("pod column header = %q, want %q", c.Header, "table.pod")
            }
        }
    }
    if !found {
        t.Fatalf("pods_sub missing \"pod\" column; got %d cols", len(cols))
    }
}
```

### Step 2: Run test to verify it fails

Run: `go test ./internal/tui/tables/ -run TestPodsSubHasPodColumn -v`
Expected: FAIL — `pods_sub` has no `"pod"` column.

### Step 3: Add column to YAML

Edit `internal/data/config/defaults/tables/compose.jsonc`. In the `pods_sub` array (between `state` and `ip`), insert:

```jsonc
      {
        "key": "pod",
        "fixed": 14,
        "header": "table.pod"
      },
```

### Step 4: Add i18n keys

In `internal/data/i18n/lang/en.jsonc` (alphabetical position — between existing `table.containers` and any following entry, or append if no following entry):

```jsonc
  "table.pod": "POD",
```

In `internal/data/i18n/lang/zh.jsonc`:

```jsonc
  "table.pod": "容器组",
```

In `internal/data/i18n/lang/ja.jsonc`:

```jsonc
  "table.pod": "ポッド",
```

### Step 5: Run test to verify it passes

Run: `go test ./internal/tui/tables/ -run TestPodsSubHasPodColumn -v`
Expected: PASS.

### Step 6: Run all table tests

Run: `go test ./internal/tui/tables/...`
Expected: all PASS.

### Step 7: Commit

```bash
git add internal/data/config/defaults/tables/compose.jsonc \
        internal/data/i18n/lang/en.jsonc \
        internal/data/i18n/lang/zh.jsonc \
        internal/data/i18n/lang/ja.jsonc \
        internal/tui/tables/compose_columns_test.go
git commit -m "feat(tables): add pod column to pods_sub profile"
```

---

## Task 6: Capability-driven column filtering in `renderComposeContainers`

**Files:**
- Modify: `internal/tui/ui/pages/compose/view.go:367-431` (`renderComposeContainers` function)
- Create: `internal/tui/ui/pages/compose/view_test.go` (NEW — first test file in this package)

**Interfaces:**
- Consumes: `m.Connection.Engine.Capabilities()`; `runtime.CapabilityComposePodScope`
- Produces: `colsDef` parameter to `component.BuildRows` and `component.RenderTable` excludes the `"pod"` column when the capability is not supported. Render width rebalanced accordingly.

### Step 1: Write the failing test

Create `internal/tui/ui/pages/compose/view_test.go`:

```go
package compose

import (
    "testing"

    runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
    "github.com/elizabevil/docker-tui/internal/data/runtime/mockengine"
    "github.com/elizabevil/docker-tui/internal/tui/state"
)

// TestRenderComposeContainersHidesPodColumnOnDocker verifies the
// container sub-view's column set excludes "pod" when the engine's
// capability set does NOT include CapabilityComposePodScope.
func TestRenderComposeContainersHidesPodColumnOnDocker(t *testing.T) {
    eng := mockengine.New(runtimeapi.Identity{Type: "test", Name: "docker"})
    eng.SetCapability(runtimeapi.CapabilityComposePodScope, runtimeapi.CapabilityInfo{Support: runtimeapi.Unsupported})

    m := state.NewAppModel(state.DefaultConfig(), eng, "test")
    m.Compose.ComposeContainerViewID = "web"
    m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{
        ID: "abc123", Name: "web-1", State: state.ContainerStateRunning,
        ComposeProject: "demo", ComposeService: "web",
    }}

    out := renderComposeContainers(m, 120, 24)
    // The pod column header should NOT appear in the rendered output
    // when the runtime does not advertise CapabilityComposePodScope.
    if containsRenderedColumn(out, "POD") {
        t.Fatalf("pod column rendered without capability:\n%s", out)
    }
}

// TestRenderComposeContainersShowsPodColumnOnPodman verifies the pod
// column IS rendered when CapabilityComposePodScope is Available.
func TestRenderComposeContainersShowsPodColumnOnPodman(t *testing.T) {
    eng := mockengine.New(runtimeapi.Identity{Type: "test", Name: "podman"})
    eng.SetCapability(runtimeapi.CapabilityComposePodScope, runtimeapi.CapabilityInfo{Support: runtimeapi.Available})

    m := state.NewAppModel(state.DefaultConfig(), eng, "test")
    m.Compose.ComposeContainerViewID = "web"
    m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{
        ID: "abc123", Name: "web-1", State: state.ContainerStateRunning,
        ComposeProject: "demo", ComposeService: "web",
        CoLocatedGroupID: "pod_demo",
    }}

    out := renderComposeContainers(m, 120, 24)
    if !containsRenderedColumn(out, "POD") {
        t.Fatalf("pod column missing with capability Available:\n%s", out)
    }
    if !containsRenderedValue(out, "pod_demo") {
        t.Fatalf("pod value missing in output:\n%s", out)
    }
}

// containsRenderedColumn / containsRenderedValue: thin string checks
// that intentionally tolerate whitespace and ANSI styling.
func containsRenderedColumn(rendered, header string) bool {
    return containsRenderedValue(rendered, header)
}

func containsRenderedValue(rendered, value string) bool {
    // Tolerant substring check; pod column may be truncated.
    for i := 0; i+len(value) <= len(rendered); i++ {
        if rendered[i:i+len(value)] == value {
            return true
        }
    }
    return false
}
```

### Step 2: Verify test setup

The test file imports `mockengine`. Inspect that package's API to confirm the constructors exist:

Run:
```bash
grep -n "func New\|func.*SetCapability\|func.*Capability" internal/data/runtime/mockengine/mockengine.go | head -10
```

If the mockengine does NOT yet expose `New(...)` / `SetCapability(...)`, the implementation must add those primitives as part of this task (no backward-compat alias needed; this is fresh test scaffolding). Add them:

In `internal/data/runtime/mockengine/mockengine.go`:

```go
// New constructs an Engine stub for tests. All methods return zero
// values; tests configure specific behaviours via setters.
func New(identity runtimeapi.Identity) *Engine {
    return &Engine{identity: identity, caps: runtimeapi.CapabilitySet{}}
}

// SetCapability records one Capability's support level in the
// stub's CapabilitySet. Tests use this to simulate a runtime with
// or without a given capability.
func (e *Engine) SetCapability(c runtimeapi.Capability, info runtimeapi.CapabilityInfo) {
    if e.caps == nil {
        e.caps = runtimeapi.CapabilitySet{}
    }
    e.caps[c] = info
}
```

Adjust the test to match the actual mockengine field names after reading `mockengine.go`.

### Step 3: Run test to verify it fails

Run: `go test ./internal/tui/ui/pages/compose/ -run TestRenderComposeContainers -v`
Expected: FAIL — `renderComposeContainers` does not yet filter colsDef.

### Step 4: Implement the filter

Edit `internal/tui/ui/pages/compose/view.go`. In `renderComposeContainers` (around line 368), after loading `colsDef`, insert:

```go
    if !podScopeSupported(m) {
        colsDef = filterOutColumn(colsDef, "pod")
    }
```

Add the helpers (near the bottom of the file):

```go
// podScopeSupported returns true when the active engine advertises
// CapabilityComposePodScope at any non-Unsupported level. nil engine
// returns false (fail closed — do not show the pod column when no
// runtime is connected).
func podScopeSupported(m *state.AppModel) bool {
    if m == nil || m.Connection.Engine == nil {
        return false
    }
    return m.Connection.Engine.Capabilities().Supports(runtime.CapabilityComposePodScope)
}

// filterOutColumn returns cols minus any ColumnDef whose Key matches
// the supplied target. Order is preserved.
func filterOutColumn(cols []tables.ColumnDef, key string) []tables.ColumnDef {
    out := make([]tables.ColumnDef, 0, len(cols))
    for _, c := range cols {
        if c.Key == key {
            continue
        }
        out = append(out, c)
    }
    return out
}
```

Add `"github.com/elizabevil/docker-tui/internal/data/runtime"` (aliased as `runtime`) to the imports if not present. Verify `tables.ColumnDef` is the type of `colsDef`; if `tables` is already imported, no change.

### Step 5: Run test to verify it passes

Run: `go test ./internal/tui/ui/pages/compose/ -run TestRenderComposeContainers -v`
Expected: PASS.

### Step 6: Run all compose tests

Run: `go test ./internal/tui/ui/pages/compose/... ./internal/data/runtime/mockengine/...`
Expected: all PASS.

### Step 7: Commit

```bash
git add internal/tui/ui/pages/compose/view.go \
        internal/tui/ui/pages/compose/view_test.go \
        internal/data/runtime/mockengine/mockengine.go
git commit -m "feat(compose): capability-driven pod column visibility in container sub-view"
```

---

## Task 7: Add `pod` cell rendering in `renderComposeContainers`

**Files:**
- Modify: `internal/tui/ui/pages/compose/view.go:393-410` (the switch statement inside the `component.BuildRows` closure)
- Modify: `internal/tui/ui/pages/compose/view_test.go` (extend with value-rendering test)

**Interfaces:**
- Consumes: `dockerclient.ContainerSummary.CoLocatedGroupID string`
- Produces: when `col.Key == "pod"`, cell text is `c.CoLocatedGroupID`; empty value renders as `component.StrDash` (existing constant for "—")

### Step 1: Extend the failing test

Append to `internal/tui/ui/pages/compose/view_test.go`:

```go
func TestRenderComposeContainersPodValue(t *testing.T) {
    eng := mockengine.New(runtimeapi.Identity{Type: "test", Name: "podman"})
    eng.SetCapability(runtimeapi.CapabilityComposePodScope, runtimeapi.CapabilityInfo{Support: runtimeapi.Available})

    m := state.NewAppModel(state.DefaultConfig(), eng, "test")
    m.Compose.ComposeContainerViewID = "web"
    m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{
        ID: "abc123", Name: "web-1", State: state.ContainerStateRunning,
        ComposeProject: "demo", ComposeService: "web",
        CoLocatedGroupID: "pod_demo",
    }}

    out := renderComposeContainers(m, 120, 24)
    if !containsRenderedValue(out, "pod_demo") {
        t.Fatalf("pod value missing:\n%s", out)
    }
}
```

### Step 2: Run test to verify it fails

Run: `go test ./internal/tui/ui/pages/compose/ -run TestRenderComposeContainersPodValue -v`
Expected: FAIL — current switch has no `case "pod"` branch; default returns `""`.

### Step 3: Add the case

Edit `internal/tui/ui/pages/compose/view.go`, in the closure passed to `component.BuildRows` (around line 393), inside the switch:

```go
            case "pod":
                if c.CoLocatedGroupID == "" {
                    return component.StrDash
                }
                return c.CoLocatedGroupID
```

### Step 4: Run test to verify it passes

Run: `go test ./internal/tui/ui/pages/compose/ -v`
Expected: PASS.

### Step 5: Commit

```bash
git add internal/tui/ui/pages/compose/view.go internal/tui/ui/pages/compose/view_test.go
git commit -m "feat(compose): render CoLocatedGroupID as pod cell value"
```

---

## Task 8: Add Pod indicator (`●`) prefix in service cell when podman

**Files:**
- Modify: `internal/tui/ui/pages/compose/view.go:255-266` (the `case "service"` branch in `renderServicePanel`)
- Modify: `internal/tui/ui/pages/compose/view_test.go` (extend with indicator test)

**Interfaces:**
- Consumes: `m.Connection.Engine.Capabilities()`; `runtime.CapabilityComposePodScope`
- Produces: when capability is supported and `ComposeFocus == 1`, the service name cell is prefixed with `"● "`; otherwise unchanged

### Step 1: Write the failing test

Append to `internal/tui/ui/pages/compose/view_test.go`:

```go
func TestRenderServicePanelPodIndicatorOnPodman(t *testing.T) {
    eng := mockengine.New(runtimeapi.Identity{Type: "test", Name: "podman"})
    eng.SetCapability(runtimeapi.CapabilityComposePodScope, runtimeapi.CapabilityInfo{Support: runtimeapi.Available})

    m := state.NewAppModel(state.DefaultConfig(), eng, "test")
    m.Compose.ComposeFocus = 1
    m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{
        ID: "abc", Name: "web-1", State: state.ContainerStateRunning,
        ComposeProject: "demo", ComposeService: "web",
    }}

    out := RenderPanel(m, 120, 24)
    if !containsRenderedValue(out, "● web") {
        t.Fatalf("pod indicator missing on podman:\n%s", out)
    }
}

func TestRenderServicePanelNoIndicatorOnDocker(t *testing.T) {
    eng := mockengine.New(runtimeapi.Identity{Type: "test", Name: "docker"})
    eng.SetCapability(runtimeapi.CapabilityComposePodScope, runtimeapi.CapabilityInfo{Support: runtimeapi.Unsupported})

    m := state.NewAppModel(state.DefaultConfig(), eng, "test")
    m.Compose.ComposeFocus = 1
    m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{
        ID: "abc", Name: "web-1", State: state.ContainerStateRunning,
        ComposeProject: "demo", ComposeService: "web",
    }}

    out := RenderPanel(m, 120, 24)
    if containsRenderedValue(out, "●") {
        t.Fatalf("pod indicator leaked on docker:\n%s", out)
    }
}
```

### Step 2: Run tests to verify they fail

Run: `go test ./internal/tui/ui/pages/compose/ -run TestRenderServicePanel -v`
Expected: FAIL on both — `renderServicePanel` does not yet prefix.

### Step 3: Implement the prefix

Edit `internal/tui/ui/pages/compose/view.go`, in `renderServicePanel`, the `case "service"` branch (around line 259):

```go
            case "service":
                if podScopeSupported(m) {
                    cells[j] = "● " + name
                } else {
                    cells[j] = name
                }
```

### Step 4: Run tests to verify they pass

Run: `go test ./internal/tui/ui/pages/compose/ -v`
Expected: all PASS.

### Step 5: Commit

```bash
git add internal/tui/ui/pages/compose/view.go internal/tui/ui/pages/compose/view_test.go
git commit -m "feat(compose): add ● pod indicator prefix in service cell when pod scope supported"
```

---

## Task 9: Rename `services_sub` column `pods` → `containers` (YAML + view.go)

**Files:**
- Modify: `internal/data/config/defaults/tables/compose.jsonc:39-61` (`services_sub` array)
- Modify: `internal/tui/ui/pages/compose/view.go:255-266` (`case "pods"` → `case "containers"`)
- Modify: `internal/tui/ui/pages/compose/view.go:340-349` (`RenderProjectDetailTable` — same rename; the `pods` key there has header `table.containers` and shows `info.running/info.total` already, but the YAML key itself is also `pods`)

**Interfaces:**
- Consumes: column key `"pods"` (existing)
- Produces: column key `"containers"` everywhere it previously was `"pods"`; case statements match the new key; header i18n reference is `table.containers`

### Step 1: Find every occurrence of the old key

Run:
```bash
grep -rn '"pods"\|case "pods":' internal/ --include='*.go' --include='*.jsonc' --exclude-dir=driver
```
Expected hits:
- `internal/data/config/defaults/tables/compose.jsonc` (services_sub column; detail column)
- `internal/tui/ui/pages/compose/view.go` (case statements in `renderServicePanel` and `RenderProjectDetailTable`)
- (No other code references the YAML column key directly.)

### Step 2: Write the failing test

Append to `internal/tui/tables/compose_columns_test.go`:

```go
func TestServicesSubContainersColumn(t *testing.T) {
    tc := MustLoad("compose")
    cols := tc.Columns.Get("services_sub")
    var hasContainers, hasLegacyPods bool
    for _, c := range cols {
        switch c.Key {
        case "containers":
            hasContainers = true
            if c.Header != "table.containers" {
                t.Fatalf("containers column header = %q, want %q", c.Header, "table.containers")
            }
        case "pods":
            hasLegacyPods = true
        }
    }
    if !hasContainers {
        t.Fatalf("services_sub missing \"containers\" column")
    }
    if hasLegacyPods {
        t.Fatalf("services_sub still has legacy \"pods\" column")
    }
}
```

### Step 3: Run test to verify it fails

Run: `go test ./internal/tui/tables/ -run TestServicesSubContainersColumn -v`
Expected: FAIL — column is still named `pods`.

### Step 4: Update YAML

Edit `internal/data/config/defaults/tables/compose.jsonc`:
- In `services_sub` array, change `{"key": "pods", "fixed": 8, "header": "table.pods"}` → `{"key": "containers", "fixed": 8, "header": "table.containers"}`.
- In `detail` array, change `{"key": "pods", "fixed": 10, "header": "table.containers"}` → `{"key": "containers", "fixed": 10, "header": "table.containers"}`.

### Step 5: Update view.go case statements

Edit `internal/tui/ui/pages/compose/view.go`:
- `renderServicePanel` (around line 263): `case "pods":` → `case "containers":`
- `RenderProjectDetailTable` (around line 347): `case "pods":` → `case "containers":`

### Step 6: Run test to verify it passes

Run: `go test ./internal/tui/tables/ -run TestServicesSubContainersColumn -v && go test ./internal/tui/ui/pages/compose/ -v`
Expected: all PASS.

### Step 7: Build everything

Run: `go build ./... && go test ./...`
Expected: exit 0; all PASS.

### Step 8: Commit

```bash
git add internal/data/config/defaults/tables/compose.jsonc \
        internal/tui/ui/pages/compose/view.go \
        internal/tui/tables/compose_columns_test.go
git commit -m "refactor(tables): rename services_sub/detail pods→containers (fix misleading header)"
```

---

## Task 10: Final integration verification

**Files:** (no edits)

### Step 1: Run the full build matrix

Run:
```bash
go vet ./...
go build ./...
go test ./...
```

Expected: all exit 0; no vet warnings.

### Step 2: Run the integration test (if Docker socket available locally)

Run:
```bash
just test-integration
```

Expected: all PASS, OR skip cleanly with a documented "no docker socket" message. (User owns the runtime-environment side; per D1 design, manual testing is the user's responsibility.)

### Step 3: Manual smoke checklist (for user)

Print this checklist so the user can run it themselves:

```text
[ ] docker:    project list shows, services show, no "●" indicator, container sub-view has 4 cols (no pod)
[ ] docker:    action bar does NOT show build/pull/push/group_*
[ ] podman:    project list shows, services show with "●" indicator, container sub-view has 5 cols (with pod)
[ ] podman:    action bar does NOT show build/pull/push/group_*
[ ] podman:    container sub-view shows CoLocatedGroupID (e.g. "pod_demo") in the pod column
[ ] down on docker:    removes project containers + networks (default)
[ ] down on podman:    same end-state (project containers + networks gone)
```

### Step 4: Tag the integration commit

```bash
git log --oneline -10
git tag feat/compose-capability-filter-test-passed HEAD   # optional
```

Do NOT push or open a PR — user owns those steps.

---

## Self-Review (post-write)

**Spec coverage** (mapped to `compose-panes-semantics-design.md`):

| Spec section | Task |
|---|---|
| §3.1 标签层统一, 物理层由 adapter 自处理 | (informational; no code) |
| §3.2 接口契约 = 用户意图 + 必要参数 | (informational; enforced via existing interface) |
| §3.3 接口实现的验收清单 | (informational; Down unchanged) |
| §3.4 dtui 是被动消费者 | (informational; no CoLocated writing) |
| §3.5 Capability = TUI 暴露范围 ∩ runtime 支持范围 | T1, T2, T3, T4 |
| §3.6 边界: 镜像动作归 R02, compose 编排动作归 R08 | T3 (delete build/pull/push) |
| §4.3 Pod 呈现位置 C (B 为底线) | T6, T7, T8 |
| §6 决策 #1 Pod 呈现位置 | T6, T7, T8 |
| §6 决策 #2 右栏标题保持 "Services" | (verified — no title change in view.go) |
| §6 决策 #3 改动范围 = UI + action 跟随, 聚合逻辑不动 | T6, T7, T8, T9 (UI); `gatherComposeProjects` unchanged |
| §6 决策 #4 不引入 Pod 过滤 / 排序 | (verified — no filter UI added) |

**Placeholder scan**: No "TBD", "TODO", "fill in details" in any task body. Every Step has either concrete code, a concrete command, or a concrete verification expectation.

**Type consistency**:
- `runtime.Capability` (string alias) — used uniformly for `RequiresCapabilities` comparisons.
- `runtime.CapabilityInfo` / `runtime.CapabilitySet` — used in mockengine setter, consistent with capabilities.go.
- `tables.ColumnDef` — used uniformly for colsDef filtering.
- `component.StrDash` — existing constant for empty cell value; used in Task 7.

**File path consistency**: Every `git add` line uses paths relative to repo root and matches the Files header of its task.

---

## Execution Handoff

After completing the tasks above, return to the orchestrator with:

1. Branch name: `feat/compose-capability-filter`
2. Commit log (10 commits expected: T1–T9 + verification tag)
3. `go test ./...` clean exit
4. The 6-item smoke checklist completed by the user