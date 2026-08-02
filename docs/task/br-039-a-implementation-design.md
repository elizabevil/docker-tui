# BR-039-A 实施设计文档

> 角色: 小模型(只读分析 + 设计输出)
> 范围: BR-039-A 框架 + typed action registry 基础(不接完整 UI / 不接全部页面动作)
> 输入: 主模型简报 [docs/task/next-small-model-brief.md](next-small-model-brief.md) + 反馈 [docs/task/br-039-main-model-feedback.md](br-039-main-model-feedback.md) + 任务卡 [br-039-action-bar-replace-q.md](br-039-action-bar-replace-q.md) + 上一轮只读分析 [br-039-code-location-analysis.md](br-039-code-location-analysis.md) + BR-040 [br-040-dialog-style-center-on-panel.md](br-040-dialog-style-center-on-panel.md)
> 输出: 本文件(唯一交付物)
> 状态: 已实施并由主模型修正;本文作为设计过程归档,最终结果以任务卡与源码为准。

---

## 1. 概述

BR-039-A 目标:

1. 取消 Q 退出 → Q 改为 toast 引导 "按 ; 打开 Action Bar,Ctrl+C 退出"
2. 新增 `;` 触发 Action Bar(`ModeActionBar` + 浮层)
3. 浮层显示当前 panel 的 typed `ActionItem` 列表,接**已实现** handler
4. j/k 移动 / Enter 执行 / Esc 关闭 / 数字键跳转 / `/` 触发 filter
5. Footer / Help 同步删除 "Q Quit" 行

复用 BR-040 已实现的 `dialog.PlaceDialogInPanel` / `PanelBody`,Action Bar 是 dialog-like widget。

---

## 2. 文件改动清单

### 2.1 新增文件 (5)

| 文件 | 作用 | 大约行数 |
|---|---|---|
| `internal/tui/state/actionbar.go` | `ActionBarState` 纯状态结构 + 方法 | ~60 |
| `internal/tui/state/actionbar_test.go` | 状态机测试 | ~80 |
| `internal/tui/keyboard/actionbar_keys.go` | Action Bar 键盘 handler | ~80 |
| `internal/tui/keyboard/actionbar_keys_test.go` | 按键分支测试 | ~100 |
| `internal/tui/actionbar/registry.go` | `ActionItem` + `ActionsFor(m)` typed registry | ~150 |
| `internal/tui/actionbar/registry_test.go` | registry 测试 | ~120 |
| `internal/tui/actionbar/actionbar.go` | `RenderBar` 通过 `PlaceDialogInPanel` 居中 | ~100 |
| `internal/tui/actionbar/actionbar_test.go` | RenderBar + 居中测试 | ~100 |

### 2.2 修改文件 (9)

| 文件 | 改动 |
|---|---|
| `internal/tui/state/app.go` | 加 `ModeActionBar` 枚举值 |
| `internal/tui/state/navigation.go` | 加 `ActionBar ActionBarState` 字段 |
| `internal/tui/keys/letters.go` | 加 `KeySemicolon = ";"` |
| `internal/tui/keys/action.go` | 加 `ActionActionBar KeyAction = "actionBar"` |
| `internal/tui/keys/registry.go` | `ActionQuit` 移除 `KeyQ`;加 `{ActionActionBar, []string{KeySemicolon}, main}` |
| `internal/tui/keys/display.go` | 保留 `ActionLabelQuit` 占位(等 BR-039-D 删除),不动 |
| `internal/tui/keyboard/actions.go` | `case ActionQuit: return m, tea.Quit` **不动**;加 `case ActionActionBar: doActionBar(m)`;**不要** 把 ActionQuit 改 toast |
| `internal/tui/keyboard/keyboard.go` | `dispatchByMode` 加 `case state.ModeActionBar`;`HandleKeyPress` 早期加 Q 拦截(`keys.IsQuit(rawKey)` → toast) |
| `internal/tui/ui/app/layout.go` | 加 `ModeActionBar` 分支,调 `widget/actionbar.RenderBar` + `dialog.PlaceDialogInPanel` |
| `internal/tui/ui/action/registry.go:37` | 删除 Q Quit 行(`{bindingLabel(model, keys.ActionQuit, keys.KeyQ), i18n.T("key.quit")}`) |

### 2.3 不应触碰的文件 (硬约束)

- `internal/tui/state/state.go` 之外的 state/* 任何文件
- `internal/tui/state/state.go` 内除 `ModeActionBar` 枚举值外的任何代码
- `internal/tui/widget/footer/footer.go`(footer 渲染逻辑)
- `internal/tui/widget/header/header.go`
- `internal/tui/keys/keys.go` (KeyQ 字符串常量的位置;KeySemicolon 加到 letters.go 而非 keys.go)
- `internal/tui/widget/dialog/centered.go` 与 `overlay.go` (BR-040 已 done)
- `internal/tui/runtime/*` 与 `internal/tui/data/runtime/*` (任何文件)
- `internal/tui/ui/widget/dialog/centered.go` 不动,直接复用 `PlaceDialogInPanel`
- `internal/tui/ui/widget/footer/footer.go` 渲染逻辑不动(Global 没 Quit 行后 footer 自动消失)
- `internal/tui/keyboard/actions.go` 的 `case ActionQuit: return m, tea.Quit` 不动(主模型硬约束 1)
- `internal/tui/keyboard/keyboard.go` 中除 `HandleKeyPress` 入口与 `dispatchByMode` 之外的逻辑不动
- `internal/tui/ui/action/registry.go` 除 `Global()` 列表 Q Quit 之外的任何逻辑不动

### 2.4 严禁

1. **严禁** 修改 `case ActionQuit: return m, tea.Quit`(Ctrl+C 必须仍能直接退出)
2. **严禁** 在 `internal/tui/state/*` 引用 `internal/tui/keys`(避免 Go import cycle)
3. **严禁** 把 `[]keys.KeyAction` 放进 `state.ActionBarState`
4. **严禁** 用 `ui/action.Shortcut` 作为 Action Bar 数据源(没有 `KeyAction` 字段)
5. **严禁** 留 hardcoded 占位 stub(直接接已有 handler 走 `keyboard.handleAction`)

---

## 3. 详细改动(按文件)

### 3.1 `internal/tui/state/actionbar.go` (新)

```go
package state

// ActionBarState owns the runtime state of the per-page Action Bar.
// Pure state only — no key references — so the state package has no
// dependency on internal/tui/keys (avoids import cycle).
//
// Field naming note: Visible (not Open) because Go does not allow a
// field and a method to share the same name on a struct; using
// Visible keeps Open() and Close() as the public lifecycle methods.
//
// The widget layer (internal/tui/actionbar) translates this state into
// the typed []ActionItem list and back.
type ActionBarState struct {
    Visible   bool   // whether the Action Bar overlay is shown
    Filtering bool   // whether the user is currently typing into the filter input
    Filter    string // fuzzy / substring filter; "" means no filter
    Selected  int    // cursor index over the current filtered action list
}

// Open resets transient state and marks the bar visible. Filtering
// is forced to false (browse mode, not filter input mode).
func (s *ActionBarState) Open() {
    s.Visible = true
    s.Filtering = false
    s.Filter = ""
    s.Selected = 0
}

// Close hides the bar and resets transient state. Safe to call
// when already closed.
func (s *ActionBarState) Close() {
    s.Visible = false
    s.Filtering = false
    s.Filter = ""
    s.Selected = 0
}

// EnterFilter switches the bar to filter-input mode without closing it.
// Used by handleActionBarKeys when the user presses '/'.
//
// The bar is rendered above the action list with an inline input.
func (s *ActionBarState) EnterFilter() {
    s.Filtering = true
    s.Filter = ""
    s.Selected = 0
}

// ExitFilter leaves filter-input mode and discards the typed filter.
// The bar stays open (Visible remains true) so the user returns to
// browsing the full action list.
func (s *ActionBarState) ExitFilter() {
    s.Filtering = false
    s.Filter = ""
    s.Selected = 0
}

// Reset clears the entire state to the zero value.
func (s *ActionBarState) Reset() {
    *s = ActionBarState{}
}

// MoveSelection moves the cursor by delta, wrapping inside [0, total).
// When total <= 0, Selected is reset to 0. Only meaningful when
// Filtering == false; while in filter-input mode the keyboard
// handler should route printable keys into Filter instead.
func (s *ActionBarState) MoveSelection(delta, total int) {
    if total <= 0 {
        s.Selected = 0
        return
    }
    s.Selected = (s.Selected + delta + total) % total
}

// SetFilter updates the filter string and resets the cursor to 0.
// The keyboard handler calls this on each character typed in filter
// mode; VisibleItems re-filters the action list from this.
func (s *ActionBarState) SetFilter(s2 string) {
    s.Filter = s2
    s.Selected = 0
}
```

**单元测试**(`state/actionbar_test.go`):

```go
func TestActionBarOpenCloseReset(t *testing.T) { ... }
func TestActionBarMoveSelection(t *testing.T) {
    // MoveSelection 边界:空列表 / 负数 / 超界 / 回卷
}
func TestActionBarSetFilterResetsCursor(t *testing.T) { ... }
```

### 3.2 `internal/tui/state/app.go`

**改动**: 在 `ModeAuditDetail` 之后(`app.go:60`)插入新枚举值:

```go
const (
    ModeNormal AppMode = iota
    // ... 现有所有 Mode ...
    ModeAuditDetail
    ModeActionBar // 新增,BR-039-A
)
```

**不应触碰**: `AppModel` struct 定义、构造函数、其它任何代码。

### 3.3 `internal/tui/state/navigation.go`

**改动**: 在 `NavigationState` struct 中加字段:

```go
type NavigationState struct {
    ActivePanel       PanelType
    PrevPanel         PanelType
    Mode              AppMode
    FilterInput       QueryInputState
    SearchInput       QueryInputState
    CommandInput      QueryInputState
    ActionBar         ActionBarState  // 新增
    FilterExitPending bool
    FilterExitToken   uint64
    EscPending        bool
}
```

### 3.4 `internal/tui/keys/letters.go`

**改动**: 在 `KeyZ = "z"` 之后(`letters.go:24`)插入:

```go
KeySemicolon = ";"
```

### 3.5 `internal/tui/keys/action.go`

**改动**: 在 `ActionRefreshConnections` 之后(`action.go:52`)插入:

```go
ActionActionBar KeyAction = "actionBar"
```

### 3.6 `internal/tui/keys/registry.go`

**改动 A**(`registry.go:38`):移除 `KeyQ` 绑定,保留 `KeyCtrlC`:

```go
{ActionQuit, []string{KeyCtrlC}, app},  // 原本: []string{KeyQ, KeyCtrlC}
```

**改动 B**(`registry.go:39-50` 任意位置):加 `ActionActionBar` 注册:

```go
{ActionActionBar, []string{KeySemicolon}, main},  // 与 ActionFilter / ActionRefresh 同一 main context
```

**改动 C**(`registry.go:172-189` `configuredBindings`):加配置字段映射:

```go
func configuredBindings(keymap config.KeymapConfig) map[KeyAction][]string {
    return map[KeyAction][]string{
        // ... 现有所有映射 ...
        ActionActionBar: keymap.ActionBar,  // 新增
    }
}
```

**注意**: `config.KeymapConfig` 需要在 `internal/data/config/types.go` 加 `ActionBar []string` 字段。如果该改动超出本轮范围,可以**不**改 `configuredBindings`,只加默认键位注册(改动 A 和 B);后续 BR 同步 `config.KeymapConfig`。

### 3.7 `internal/tui/keys/display.go`

**改动**: 无。

保留 `ActionLabelQuit = "Quit"`(line 61)与 `ActionLabelFilter / ActionLabelHelp` 等所有 label。后续 BR-039-D 统一清理文案。

### 3.8 `internal/tui/keyboard/actions.go`

**改动 A**(`actions.go:18-19`):**保持不变**。`case ActionQuit: return m, tea.Quit` 不动。

**改动 B**(`actions.go:120-123` 之后):加 `case ActionActionBar`:

```go
case keys.ActionActionBar:
    return doActionBar(m), nil
```

**改动 C**(`actions.go` 末尾,新增 `doActionBar` 函数):

```go
// doActionBar opens the Action Bar overlay for the active panel.
// Per BR-039-A: pure framework, no per-panel action list; ActionsFor
// returns the typed registry for m.Navigation.ActivePanel.
func doActionBar(m *state.AppModel) (*state.AppModel, tea.Cmd) {
    m.Navigation.ActionBar.Open()
    m.Navigation.Mode = state.ModeActionBar
    return m, nil
}
```

**注意**: `doActionBar` 不传 `keys.KeyAction` 到 `state`(避免 import cycle)。`m.Navigation.ActionBar.Selected` 在 `widget/actionbar` 子包中通过 `ActionsFor(m)` 渲染时取 `len(...)` 来初始化 cursor。

### 3.9 `internal/tui/keyboard/keyboard.go`

**改动 A**(`HandleKeyPress` 函数 `keyboard.go:14-54`):Q 拦截**不能放在全局早期路径**(`dispatchByMode` 之前),否则在 `ModeFilter` / `ModeSearch` / `ModeCommand` 输入框中无法输入 `q`。应放在 **`dispatchByMode` 之后**的"普通模式 fallback"路径中,且仅当 `m.Navigation.Mode == state.ModeNormal` 才触发:

```go
func HandleKeyPress(msg tea.KeyPressMsg, m *state.AppModel) (*state.AppModel, tea.Cmd) {
    rawKey := msg.String()
    key := keys.Normalize(rawKey)

    // 现有 Esc / 非 Esc 副作用处理保留 (lines 18-27)
    if key != keys.KeyEsc { m.Navigation.EscPending = false; m.Feedback.InfoMessage = "" }
    if key == keys.KeyEsc && m.Feedback.ErrorMessage != "" { m.Feedback.ClearError() }

    // Mode-keyed dispatch FIRST (existing, unchanged)
    if mm, cmd, handled := dispatchByMode(rawKey, key, m); handled {
        return mm, cmd
    }
    // Nested view keys (existing, unchanged)
    if mm, cmd := handleNestedViewKeys(key, m); mm != nil || cmd != nil { return mm, cmd }

    // NEW: Q 提示 — 在普通模式 fallback 路径,只在 ModeNormal 触发。
    // 必须在 dispatchByMode 之后,这样 ModeFilter / ModeSearch / ModeCommand
    // 内的 q 输入能正常被自己的 input handler 消费,不会误触发。
    // 同样不会与 ModeActionBar 冲突(那里 handleActionBarKeys 已消费 q)。
    if rawKey == keys.KeyQ && m.Navigation.Mode == state.ModeNormal {
        ShowToastNow(m, "Q → Action Bar 入口已迁移,按 ; 打开;退出请用 Ctrl+C")
        return m, nil
    }

    // 现有 resolveAction + handleAction + handlePanelFallbacks + handleShortcuts
    // (lines 45-53 unchanged)
    ...
}
```

**改动 B**(`dispatchByMode` `keyboard.go:60-112`):在 `case state.ModeMark` 之后加 `case state.ModeActionBar`:

```go
case state.ModeActionBar:
    m, cmd := handleActionBarKeys(rawKey, m)
    return m, cmd, true
```

### 3.10 `internal/tui/keyboard/actionbar_keys.go` (新)

```go
package keyboard

import (
    "github.com/elizabevil/docker-tui/internal/tui/state"
    "github.com/elizabevil/docker-tui/internal/tui/actionbar"

    tea "charm.land/bubbletea/v2"
)

// handleActionBarKeys dispatches keys when the Action Bar overlay is
// active. Routes through two sub-modes based on m.ActionBar.Filtering:
//
//   - Filtering == true:  printable keys (including 'q') go into the
//     filter input; Esc/Enter exit filter mode. j/k/digit do NOT move
//     the selection while in filter mode.
//
//   - Filtering == false: j/k move selection, digit jumps, Enter executes,
//     Esc closes, '/' enters filter mode.
//
// On Enter (in browse mode) we delegate to the package-local
// handleAction(sel.Action, m, nil). The widget/actionbar package only
// provides the typed []ActionItem list; the execution chain stays in
// keyboard so we don't import cycle.
func handleActionBarKeys(rawKey string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
    if m.Navigation.Mode != state.ModeActionBar {
        return m, nil
    }

    // Filter mode: route printable keys into the filter input.
    if m.Navigation.ActionBar.Filtering {
        switch rawKey {
        case "esc":
            m.Navigation.ActionBar.ExitFilter()
            return m, nil
        case "enter":
            // Confirm: exit filter mode, keep typed Filter. VisibleItems
            // re-filters on next render.
            m.Navigation.ActionBar.Filtering = false
            m.Navigation.ActionBar.Selected = 0
            return m, nil
        default:
            // 按键字符写入 Filter (单字符, ASCII;回退到默认 editor for 其它)
            m.Navigation.ActionBar.Filter = m.Navigation.ActionBar.Filter + rawKey
            m.Navigation.ActionBar.Selected = 0
            return m, nil
        }
    }

    // Browse mode.
    items := actionbar.VisibleItems(m)
    switch rawKey {
    case "esc":
        m.Navigation.ActionBar.Close()
        m.Navigation.Mode = state.ModeNormal
        return m, nil
    case "enter":
        if len(items) == 0 {
            return m, nil
        }
        sel := items[m.Navigation.ActionBar.Selected]
        m.Navigation.ActionBar.Close()
        m.Navigation.Mode = state.ModeNormal
        return handleAction(sel.Action, m, nil)
    case "j", "down":
        m.Navigation.ActionBar.MoveSelection(+1, len(items))
        return m, nil
    case "k", "up":
        m.Navigation.ActionBar.MoveSelection(-1, len(items))
        return m, nil
    case "/":
        m.Navigation.ActionBar.EnterFilter()
        return m, nil
    default:
        // 数字键 1-9 跳到第 N 项
        if len(rawKey) == 1 && rawKey >= "1" && rawKey <= "9" {
            idx := int(rawKey[0] - '1')
            if idx < len(items) {
                m.Navigation.ActionBar.Selected = idx
            }
            return m, nil
        }
        return m, nil
    }
}
```

**重要说明 — registry 与 keyboard 职责划分**:
- `ui/actionbar/VisibleItems(m)` 只**返回 typed `[]ActionItem`**,每个 Item 含 `Action KeyAction` 字段。**这只验证 ID 映射**(哪个 panel 配哪些 KeyAction)。
- 真实执行链(`Enter` 后调 `handleAction(sel.Action, m, nil)` → 已有 `case ActionContainerStart` 等)是 **keyboard 包内** 的,本轮需要新增 `keyboard/actionbar_keys_test.go` 端到端验证。
- `widget/actionbar/registry_test.go` 只测 ID 列表正确性;`keyboard/actionbar_keys_test.go` 测 Enter 后真的执行了对应 handler(可通过 mock `*state.AppModel` + 验证 `m.Resources.Containers.Loading` / toast 等副作用)。

**Import 关系**(无循环):
- `keyboard` → `actionbar`(调 `actionbar.VisibleItems`)+ `state` + `keys` + `tea`
- `actionbar` → `state` + `keys`(只用 `keys.KeyAction` 常量)+ `dialog` + `component`
- `state` → 0 个其他包(纯状态,无 keys 引用)
- `keys` ← 已被 state/keyboard/actionbar 三处引用,自身只依赖 data/config

**不存在的旧路径**(本设计禁止):
- `internal/config`、`internal/tui/model`、`internal/data/docker`

### 3.11 `internal/tui/actionbar/` (新子包)

#### 3.11.1 `registry.go`

```go
package actionbar

import (
    "github.com/elizabevil/docker-tui/internal/tui/keys"
    "github.com/elizabevil/docker-tui/internal/tui/state"
)

// ActionItem is a single row in the Action Bar. KeyAction identifies
// the runtime handler to invoke on Enter; Key is the human-readable
// key label (may be "—" for actions without a key binding).
type ActionItem struct {
    Key         string        // e.g. "s" / "Ctrl+T" / "—"
    Label       string        // e.g. "Stop container"
    Action      keys.KeyAction
    Description string        // optional sub-label
    Disabled    bool          // reserved for future BR-033/036/037 disabled actions
}

// VisibleItems returns the per-panel action list after applying the
// current Filter. It MUST be deterministic so the keyboard layer can
// drive cursor moves and Enter with matching indices.
func VisibleItems(m *state.AppModel) []ActionItem {
    if m == nil {
        return nil
    }
    raw := actionsForPanel(m)
    if m.Navigation.ActionBar.Filter == "" {
        return raw
    }
    // 简单子串过滤(MVP;BR-039-B 升级为 fuzzy match)
    out := make([]ActionItem, 0, len(raw))
    f := m.Navigation.ActionBar.Filter
    for _, it := range raw {
        if containsFold(it.Label, f) || containsFold(it.Description, f) {
            out = append(out, it)
        }
    }
    return out
}

func containsFold(haystack, needle string) bool {
    // 简单 ASCII case-insensitive contains (避免 import strings 链麻烦)
    // 实际实现可直接 strings.Contains(strings.ToLower(...))
}

// actionsForPanel dispatches by ActivePanel. Each branch returns the
// typed ActionItem list for the panel, hitting only already-implemented
// handler actions in keyboard.handleAction.
func actionsForPanel(m *state.AppModel) []ActionItem {
    switch m.Navigation.ActivePanel {
    case state.PanelContainers:
        return containerActions(m)
    case state.PanelImages:
        return imageActions(m)
    case state.PanelVolumes:
        return volumeActions(m)
    case state.PanelNetworks:
        return networkActions(m)
    case state.PanelCompose:
        return composeActions(m)
    case state.PanelAudit:
        return auditActions(m)
    default:
        return nil
    }
}

func containerActions(m *state.AppModel) []ActionItem {
    items := []ActionItem{
        {Key: "s",         Label: "Start",         Action: keys.ActionContainerStart},
        {Key: "Ctrl+S",    Label: "Stop",          Action: keys.ActionContainerStop},
        {Key: "Ctrl+R",    Label: "Restart",       Action: keys.ActionContainerRestart},
        {Key: "Ctrl+K",    Label: "Kill",          Action: keys.ActionContainerKill},
        {Key: "Ctrl+D",    Label: "Remove",        Action: keys.ActionContainerRemove},
        {Key: "l",         Label: "Logs",          Action: keys.ActionContainerLogs},
        {Key: "e",         Label: "Exec",          Action: keys.ActionContainerExec},
        {Key: "i",         Label: "Inspect",       Action: keys.ActionContainerInspect},
        {Key: "m",         Label: "Stats",         Action: keys.ActionContainerStats},
        {Key: "p",         Label: "Pause/Unpause", Action: keys.ActionContainerPause},
        {Key: "d",         Label: "Detail",        Action: keys.ActionDetail},
    }
    // 选中行可执行性检查: 容器面板无选中 → 标 Disabled
    if m.Resources.Containers.Selected() == nil {
        for i := range items {
            items[i].Disabled = true
        }
    }
    return items
}

func imageActions(m *state.AppModel) []ActionItem {
    // 类似 containers,接 ActionImagePull / Prune / Remove / Detail / Tag / Push / Save / Load
    // 镜像面板无选中 → Disabled
}

func volumeActions(m *state.AppModel) []ActionItem {
    // ActionVolumeCreate / Prune / Remove / Detail
}

func networkActions(m *state.AppModel) []ActionItem {
    // ActionNetworkCreate / Prune / Remove / Detail
}

func composeActions(m *state.AppModel) []ActionItem {
    // Start / Stop / Logs / Detail (existing compose actions)
}

func auditActions(m *state.AppModel) []ActionItem {
    // Enter (Detail) / E (Filter level) / / (Filter)
}
```

**关键决策**(§4): 每 panel 列表**只列**已实现 handler 的动作。TASK-019 高级动作(BR-033)暂不列。

#### 3.11.2 `actionbar.go`

```go
package actionbar

import (
    "fmt"
    "strings"

    "charm.land/lipgloss/v2"
    "github.com/elizabevil/docker-tui/internal/tui/ui/component"
    "github.com/elizabevil/docker-tui/internal/tui/ui/widget/dialog"
)

// RenderBar formats items as a list, marks the Selected row, and
// splices it into content via BR-040 PlaceDialogInPanel so the overlay
// is centered within the active panel.
func RenderBar(m *state.AppModel, content string, body dialog.PanelBody) string {
    cfg := dialog.LoadDialogConfig()
    items := VisibleItems(m)
    if len(items) == 0 {
        return content
    }
    box := renderBox(items, m.Navigation.ActionBar.Selected, body)
    return dialog.PlaceDialogInPanel(content, box, body, cfg)
}

func renderBox(items []ActionItem, selected int, body dialog.PanelBody) string {
    // lipgloss 风格化: 标题 + 行列表
    // 与 ChoiceDialog / SelectionDialog 风格一致(同一项目)
    var sb strings.Builder
    sb.WriteString(component.GetStyle("panelTitle").Render("[Action Bar]"))
    sb.WriteByte('\n')
    innerW := max(1, body.Width-4)
    for i, it := range items {
        marker := "  "
        if i == selected {
            marker = "▶ "
        }
        line := fmt.Sprintf("%s%-12s  %s", marker, it.Key, it.Label)
        if it.Disabled {
            line = component.GetStyle("dim").Render(line)
        }
        sb.WriteString(component.GetStyle("").Render(line))
        sb.WriteByte('\n')
    }
    if body.Width > 0 {
        sb.WriteString(component.PadVisible("", innerW))
    }
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Padding(0, 1).
        Width(body.Width).
        MaxHeight(body.Rows).
        Render(sb.String())
}
```

**注意**:
- `dialog.PanelBody` 是 BR-040 新增结构(在 `widget/dialog/centered.go` 里)
- `dialog.PlaceDialogInPanel` 与 `dialog.LoadDialogConfig` 已存在(BR-040)

### 3.12 `internal/tui/ui/app/layout.go`

**改动**: 在 `RenderApp` 的 dialog 分支前(`layout.go:225-228` 之间)插入 `ModeActionBar` 分支:

```go
if m.Navigation.Mode == state.ModeActionBar {
    rep := ResolveLayout(m)
    body := dialog.PanelBody{
        Left:  rep.Panel.bodyLeft,
        Top:   rep.Panel.bodyTop,
        Width: rep.Panel.bodyWidth,
        Rows:  rep.Panel.bodyRows,
    }
    return actionbar.RenderBar(m, result, body)
}
```

需要 import `"github.com/elizabevil/docker-tui/internal/tui/actionbar"`。

### 3.13 `internal/tui/ui/action/registry.go`

**改动**(`registry.go:37`):删除 Q Quit 行:

```go
func Global(app ...*state.AppModel) []Shortcut {
    model := firstModel(app)
    return []Shortcut{
        {bindingLabel(model, keys.ActionDown, keys.KJDown), i18n.T("key.down")},
        {bindingLabel(model, keys.ActionUp, keys.KKUp), i18n.T("key.up")},
        {bindingLabel(model, keys.ActionTabNext, keys.KTab), i18n.T("key.panel")},
        {bindingLabel(model, keys.ActionEnter, keys.KEnter), "Enter"},
        {bindingLabel(model, keys.ActionBack, keys.KEsc), i18n.T("key.back")},
        {bindingLabel(model, keys.ActionFilter, keys.KeySlash), i18n.T("key.filter")},
        {bindingLabel(model, keys.ActionRefresh, keys.KeyR), "Refresh"},
        {bindingLabel(model, keys.ActionCommand, ":"), "Command"},
        {bindingLabel(model, keys.ActionSwitchRuntime, keys.KeyF2), "Runtime"},
        {bindingLabel(model, keys.ActionHelp, keys.KeyQmark), i18n.T("key.help")},
        {keys.KeyHUpper, i18n.T("key.header")},
        {keys.KeyCUpper, i18n.T("key.connect")},
        // 删除这一行:
        // {bindingLabel(model, keys.ActionQuit, keys.KeyQ), i18n.T("key.quit")},
    }
}
```

**注意**: `keys.ActionQuit` 与 `keys.KeyQ` 字符串常量保留(`keys/letters.go` `keys/action.go` `keys/display.go` 不动);`ActionLabelQuit` 字符串保留。后续 BR-039-D 统一删除。

---

## 4. 关键设计决策点

### 4.1 决策 1:VisibleItems 的归属

**问题**: `handleActionBarKeys`(`keyboard` 包)需要枚举当前 panel 的 ActionItem 列表,但**不能** import `widget/actionbar`(会循环)。

**决策 A**:**VisibleItems 放在 `internal/tui/actionbar/`**(`registry.go`)。
- `handleActionBarKeys` 通过 `m.Navigation.ActivePanel + m.Navigation.ActionBar.Filter` 调 `internal/tui/actionbar.VisibleItems(m)`,**keyboard 包 import actionbar 子包**。这没问题,因为 keyboard 已经是顶层包(import state / keys / ui),加 `internal/tui/actionbar` 依赖不形成循环(actionbar 不 import keyboard)。
- `widget/actionbar/registry.go` 内调 `keyboard.handleAction` 不可行(同样循环);改用**直接返回 `keys.KeyAction` 列表**,由 keyboard 端 `handleActionBarKeys` 选完后调 `handleAction(action, m, nil)`(本包)。

**决策 B**:`VisibleItems` 拆成两段:
- `internal/tui/actionbar/VisibleItems` → 返回 `[]keys.KeyAction`(纯数据)
- `widget/actionbar/actionbar.go` → 渲染层

keyboard import actionbar → 无循环。

**采用决策 A**:VisibleItems 在 actionbar 包,返回 `[]ActionItem`(含 KeyAction 字段)。handleActionBarKeys 调 `actionbar.VisibleItems(m)`,Enter 后调 `keyboard.handleAction(sel.Action, m, nil)`(本包函数)。

### 4.2 决策 2:Q 拦截的精确位置

`keyboard.go:HandleKeyPress` 早期(`rawKey = msg.String()` 后,`Mode-keyed dispatch` 前):

```go
if rawKey == keys.KeyQ && m.Navigation.Mode == state.ModeNormal {
    ShowToastNow(m, "Q → Action Bar 入口已迁移,按 ; 打开;退出请用 Ctrl+C")
    return m, nil
}
```

**注意**: 只在 `ModeNormal` 拦截。`ModeActionBar` / `ModeFilter` / `ModeCommand` 等模式下 Q 仍走原路径(可能是命令输入字符)。

### 4.3 决策 3:Action Bar 与现有 :command 面板的交互

- `:` (colon) → `ModeCommand`(原路径,不动)
- `;` (semicolon) → `ModeActionBar`(新路径)
- 两者独立 Mode,独立 state(`CommandInput` vs `ActionBarState`)
- Action Bar 内 `:` 不打开命令面板(简化:把 `:` 当作普通字符,不过滤);命令输入仍走 `:` 全局快捷键

### 4.4 决策 4:Action Bar 与子视图(image→containers / compose→services)的交互

子视图打开时(`ContainersViewID != ""` / `ComposeContainerViewID != ""`),Action Bar 的 actions 列表:
- 默认按 `m.Navigation.ActivePanel` 计算
- 如果打开子视图,仍按 outer panel 计算(简化)
- 后续 BR 可优化为按子视图内容调整

### 4.5 决策 5:handleActionBarKeys 的 Selected 与 VisibleItems 索引同步

每次按键(除 Esc / Enter 关闭外),重新计算 `items := actionbar.VisibleItems(m)`,然后 `MoveSelection` 用 `len(items)` 做边界 clamp + wrap。这保证:
- Filter 改变后,Selected 自动 reset 到 0
- items 列表长度变化,Selected 不会越界
- wrap 行为与 `ConfirmState.MoveFocus` 一致

---

## 5. 数据流

### 5.1 打开 Action Bar

```
User 按 ;
  → keyboard.HandleKeyPress(";", m)
  → resolveAction → ActionActionBar
  → handleAction(ActionActionBar, m, cmds)
      case ActionActionBar: doActionBar(m)
  → m.Navigation.Mode = ModeActionBar
  → m.Navigation.ActionBar.Open() (reset Selected/Filter)
  → RenderApp
  → ModeActionBar 分支
  → actionbar.RenderBar(m, content, body)
  → actionbar.VisibleItems(m) → []ActionItem
  → renderBox → lipgloss 风格化
  → dialog.PlaceDialogInPanel(content, box, body, cfg)  ← BR-040 复用
```

### 5.2 执行选中

```
User 按 Enter (ModeActionBar)
  → handleActionBarKeys
  → items = actionbar.VisibleItems(m)
  → sel = items[m.Navigation.ActionBar.Selected]
  → m.Navigation.ActionBar.Close()
  → m.Navigation.Mode = ModeNormal
  → handleAction(sel.Action, m, nil)  ← 已有的 case 分发
  → 调 doContainerAction / openImageWorkflow / etc.
```

### 5.3 关闭 Action Bar

```
User 按 Esc (ModeActionBar)
  → handleActionBarKeys
  → m.Navigation.ActionBar.Close()
  → m.Navigation.Mode = ModeNormal
```

### 5.4 Q 引导

```
User 按 Q (ModeNormal)
  → HandleKeyPress 早期拦截
  → ShowToastNow("Q → Action Bar 入口已迁移,按 ; 打开;退出请用 Ctrl+C")
  → return m, nil (不调后续 resolveAction)
```

### 5.5 Ctrl+C 退出(不变)

```
User 按 Ctrl+C
  → resolveAction → ActionQuit (注册表仍含 KeyCtrlC)
  → handleAction(ActionQuit, m, cmds)
  → case ActionQuit: return m, tea.Quit
  → 应用退出
```

---

## 6. 测试设计

### 6.1 `state/actionbar_test.go`

```go
func TestActionBarOpenResetsState(t *testing.T) {
    s := ActionBarState{Selected: 5, Filter: "x"}
    s.Open()
    if !s.Open { t.Fatal("Open not set") }
    if s.Selected != 0 { t.Errorf("Selected = %d", s.Selected) }
    if s.Filter != "" { t.Errorf("Filter = %q", s.Filter) }
}

func TestActionBarClose(t *testing.T) {
    s := ActionBarState{Open: true, Selected: 3}
    s.Close()
    if s.Open { t.Error("Open still true") }
    if s.Selected != 0 { t.Errorf("Selected = %d", s.Selected) }
}

func TestActionBarMoveSelectionEmpty(t *testing.T) {
    s := ActionBarState{}
    s.MoveSelection(+1, 0)
    if s.Selected != 0 { t.Errorf("expected 0, got %d", s.Selected) }
}

func TestActionBarMoveSelectionWraps(t *testing.T) {
    s := ActionBarState{Selected: 4}
    s.MoveSelection(+1, 5)
    if s.Selected != 0 { t.Errorf("expected wrap to 0, got %d", s.Selected) }
    s.MoveSelection(-1, 5)
    if s.Selected != 4 { t.Errorf("expected wrap to 4, got %d", s.Selected) }
}

func TestActionBarSetFilterResetsCursor(t *testing.T) {
    s := ActionBarState{Selected: 2}
    s.SetFilter("foo")
    if s.Filter != "foo" { t.Errorf("Filter = %q", s.Filter) }
    if s.Selected != 0 { t.Errorf("Selected = %d, want 0", s.Selected) }
}
```

### 6.2 `keyboard/actionbar_keys_test.go`

```go
func TestActionBarKeysOnlyActiveInMode(t *testing.T) {
    m := newModel(...)
    // ModeNormal 时不消费按键
    m, _ = handleActionBarKeys("j", m)
    if m.Navigation.Mode != state.ModeNormal { t.Error("consumed j in ModeNormal") }
    // ModeActionBar 时消费
    m.Navigation.Mode = state.ModeActionBar
    items := []int{3}  // 模拟 len(items)
    _ = items
    m, _ = handleActionBarKeys("j", m)
    // Selected 应增加
}

func TestActionBarKeysEscCloses(t *testing.T) { ... }
func TestActionBarKeysEnterExecutes(t *testing.T) { ... }
func TestActionBarKeysDigitJump(t *testing.T) { ... }
```

### 6.3 `internal/tui/actionbar/registry_test.go`

```go
func TestActionsForContainers(t *testing.T) {
    m := newModel(...)
    m.Navigation.ActivePanel = state.PanelContainers
    items := VisibleItems(m)
    for _, it := range items {
        if it.Action == "" { t.Errorf("empty Action: %+v", it) }
    }
    // 容器面板无选中 → items 应标 Disabled
    for _, it := range items {
        if !it.Disabled { t.Error("expected Disabled when no selection") }
    }
}

func TestActionsForEachPanel(t *testing.T) {
    for _, panel := range []state.PanelType{
        state.PanelContainers, state.PanelImages,
        state.PanelVolumes, state.PanelNetworks,
        state.PanelCompose, state.PanelAudit,
    } {
        m := newModel(...)
        m.Navigation.ActivePanel = panel
        items := VisibleItems(m)
        if len(items) == 0 {
            t.Errorf("panel %v returned empty", panel)
        }
    }
}

func TestVisibleItemsFilter(t *testing.T) {
    m := newModel(...)
    m.Navigation.ActivePanel = state.PanelContainers
    m.Navigation.ActionBar.SetFilter("Stop")
    items := VisibleItems(m)
    for _, it := range items {
        if !strings.Contains(strings.ToLower(it.Label), "stop") {
            t.Errorf("filter missed: %+v", it)
        }
    }
}
```

### 6.4 `internal/tui/actionbar/actionbar_test.go`

```go
func TestRenderBarEmpty(t *testing.T) {
    // ModeNormal / no items → 返回 content 原样
    m := newModel(...)
    body := dialog.PanelBody{Left: 0, Top: 0, Width: 80, Rows: 24}
    got := RenderBar(m, "hello", body)
    if got != "hello" { t.Errorf("expected passthrough, got %q", got) }
}

func TestRenderBarCenteredInPanel(t *testing.T) {
    // BR-040 复用:验证 RenderBar 用 PlaceDialogInPanel 居中
    m := newModel(...)
    m.Navigation.ActivePanel = state.PanelContainers
    m.Navigation.Mode = state.ModeActionBar
    body := dialog.PanelBody{Left: 10, Top: 5, Width: 60, Rows: 15}
    content := buildContent(80, 24, '.')
    got := RenderBar(m, content, body)
    // 验证 content 其它行未被修改
    lines := strings.Split(got, "\n")
    if len(lines) != 24 { t.Fatalf("lines = %d", len(lines)) }
    // 验证 body 外的行(0..4, 20..23)未变
    for _, idx := range []int{0, 4, 20, 23} {
        if lines[idx] != buildContentRow(80, '.') {
            t.Errorf("line %d modified", idx)
        }
    }
}
```

### 6.5 `keys/registry_test.go`(修改)

```go
// 原断言: got[0] != "q"
// 改为:
if got := bindings.ByAction[ActionQuit]; len(got) > 0 && got[0] == "q" {
    t.Errorf("ActionQuit should not bind to Q (per BR-039)")
}
// 加新断言: ActionActionBar 绑定 ;
if got := bindings.ByAction[ActionActionBar]; len(got) == 0 || got[0] != ";" {
    t.Errorf("ActionActionBar should bind to ;")
}
```

### 6.6 `ui/app/layout_test.go`(新增或修改)

如已有该测试,加 ModeActionBar 分支断言;否则新建:

```go
func TestRenderAppModeActionBar(t *testing.T) {
    m := newModel(120, 32)
    m.Navigation.ActivePanel = state.PanelContainers
    m.Navigation.Mode = state.ModeActionBar
    got := RenderApp(m)
    if !strings.Contains(got, "Action Bar") {
        t.Errorf("Action Bar not rendered")
    }
}
```

---

## 7. 实现顺序(自上而下,每步验证编译)

| 步骤 | 文件 | 改动 | 验证命令 |
|---|---|---|---|
| 1 | `keys/letters.go` | 加 `KeySemicolon` | `go build ./internal/tui/keys/` |
| 2 | `keys/action.go` | 加 `ActionActionBar` | 同上 |
| 3 | `keys/registry.go` | 改 `ActionQuit` 移除 `KeyQ`;加 `ActionActionBar` | `go test ./internal/tui/keys/` |
| 4 | `state/app.go` | 加 `ModeActionBar` | `go build ./internal/tui/state/` |
| 5 | `state/navigation.go` | 加 `ActionBar` 字段 | 同上 |
| 6 | `state/actionbar.go` (新) | `ActionBarState` + 方法 | `go test ./internal/tui/state/` |
| 7 | `internal/tui/actionbar/registry.go` (新) | `ActionItem` + `VisibleItems` + 6 panel actions | `go build ./internal/tui/actionbar/` |
| 8 | `internal/tui/actionbar/actionbar.go` (新) | `RenderBar` | 同上 |
| 9 | `keyboard/actions.go` | 加 `case ActionActionBar: doActionBar(m)` | `go build ./internal/tui/keyboard/` |
| 10 | `keyboard/keyboard.go` | `HandleKeyPress` 早期加 Q 拦截;`dispatchByMode` 加 `ModeActionBar` 分支 | `go test ./internal/tui/keyboard/` |
| 11 | `keyboard/actionbar_keys.go` (新) | `handleActionBarKeys` | 同上 |
| 12 | `ui/app/layout.go` | 加 `ModeActionBar` 分支 | `go build ./internal/tui/ui/app/` |
| 13 | `ui/action/registry.go` | 删除 Q Quit 行 | `go test ./internal/tui/ui/action/` |
| 14 | 所有测试一起跑 | `go test ./internal/tui/...` | 全 ok |

**每步后** `go build ./...` 验证编译;**最后** `go test ./internal/tui/...` 验证无回归。

---

## 8. 验证清单(实现完成时主模型逐项验收)

| 验证项 | 期望 | 验证方法 |
|---|---|---|
| Q 行为 | toast,不退出 | 手动 + test |
| Ctrl+C 行为 | tea.Quit 直接退出 | 手动 |
| `;` 行为 | 打开 Action Bar | 手动 + test |
| Esc 行为 | 关闭 Action Bar,无副作用 | 手动 + test |
| Enter 行为 | 执行选中动作,ModeActionBar → ModeNormal | 手动 + test |
| j/k 行为 | 上下移动 Selected,边界 wrap | 手动 + test |
| 数字键 1-9 | 跳到第 N 项 | 手动 + test |
| `/` 行为 | 后续 BR-039-B 处理,本轮占位 no-op |  |
| Action Bar 内容 | 随 `m.Navigation.ActivePanel` 变化 | test(6 panel) |
| Action Bar 居中 | 用 `PlaceDialogInPanel` 居中于 panel body | test |
| Footer "Q Quit" | 不再显示 | `action/registry_test.go` |
| Help "Q Quit" | 不再显示 | 手动 |
| `state` 包 | 不 import `keys` | `grep "tui/keys" internal/tui/state/*.go` |
| `case ActionQuit` | 仍返回 `m, tea.Quit` | `keyboard/actions_test.go` |
| import cycle | 无 | `go build ./...` 通过 |
| 全部测试 | 通过 | `go test ./internal/tui/...` |

---

## 9. 风险与边界

| 风险 | 边界处理 |
|---|---|
| Q 引导在 ModeCommand / ModeActionBar 内触发会破坏输入 | `HandleKeyPress` 早期拦截条件加 `m.Navigation.Mode == ModeNormal` |
| `;` 与现有 `:` `/` 等输入字符冲突 | 现有无 `;` 绑键,无冲突 |
| Action Bar items 列表为空 | `RenderBar` 在空时直接返回 content 原样(不弹浮层) |
| 容器面板无 `Selected()` 时调 action 报错 | `containerActions` 中无选中 → `items[i].Disabled = true`,Enter 时 `handleAction` 由 action 自身处理 no-selection |
| Filter 改变后 Selected 越界 | `SetFilter` 强制 `Selected = 0` |
| 数字键 0 行为 | 不响应("1"-"9" 范围) |
| 子视图打开时 ActivePanel 不变 | VisibleItems 仍按 outer panel 计算,子视图内不重计算 |
| Action Bar 与 `ModeCommand` 冲突 | 各自独立 Mode,各自独立 state;`:` 不在 Action Bar 内打开 `ModeCommand` |
| `keyboard.handleAction` 与 actionbar 包循环 | `handleAction` 在 keyboard 包;actionbar 包**不** import keyboard;VisibleItems 在 actionbar 包,返回 `[]ActionItem` 含 `keys.KeyAction`;keyboard 调 `handleAction(sel.Action, m, nil)` 直接调本包函数,无循环 |
| 配置文件 `config.KeymapConfig` 是否加 `ActionBar` 字段 | 本轮不**必**加;只加默认键位注册;若加,只动 `internal/data/config/types.go` 一个字段,scope 小 |
| i18n 字符串"Q → Action Bar..." | 本轮用硬编码英文字符串;BR-039-D 统一补全 i18n |
| Action Bar 浮层超出 panel body | `PlaceDialogInPanel` 内部 clamp;空 list 不弹 |

---

## 10. 主模型硬约束交叉验证

| 主模型硬约束 | 本设计如何满足 |
|---|---|
| 1. 不改 `case ActionQuit: return m, tea.Quit` | §3.8 改动 B **仅加** `case ActionActionBar`,`case ActionQuit` 保持原样 |
| 2. `state` 不 import `keys` | §3.1 `ActionBarState` 字段类型 `bool/int/string`,无 `keys.KeyAction`;`actionbar.VisibleItems` 在 ui 包而非 state 包 |
| 3. 不用 `ui/action.Shortcut` 作 Action Bar 数据源 | §3.11.1 新建 `ActionItem` typed struct;`registry.go` 在 widget/actionbar 包,与 ui/action 平级 |
| 4. 不留 hardcoded 占位 | §3.11.1 每 panel 返回**已实现** handler 的 ActionItem;空选中 → Disabled(不静默) |
| 5. 不把 Q Quit 留到 BR-039-D | §3.13 立即删 Q Quit 行;`ActionLabelQuit` 字符串保留(后续 BR-039-D 统一删) |
| 6. 不接 BR-033/036/037/038 | §3.11.1 明确只列已实现 handler;TASK-019 / Image Import / Registry Login / Exec UI 不在本轮 |

---

## 11. 给主模型汇报模板

实现完成后小模型按此模板汇报(主模型审):

```text
## BR-039-A 实施完成

### 改动文件清单
[paste `git status` + `git diff --stat`]

### 关键设计点
- ActionBarState 形状:仅 Open/Selected/Filter,不持有 keys.KeyAction(避免 import cycle)
- VisibleItems 在 widget/actionbar 包,返回 []ActionItem{Key, Label, Action KeyAction, ...}
- handleActionBarKeys 通过 handleAction(本包)调选中 action,**不**经过 actionbar 包
- Q 拦截在 HandleKeyPress 早期,仅 ModeNormal,ShowToastNow 后 return
- case ActionQuit 保持 m, tea.Quit 不动
- 删 Footer "Q Quit" 行

### 测试结果
[go test ./internal/tui/... 输出]

### 仍需后续(BR-039-B/C/D)
- 完整 6 panel 动作清单(目前只接已实现)
- TASK-019 / Image Import / Registry Login 接入(等 BR-033/036/037 done)
- i18n 文案 / Help 页 / 完整 footer 文案
- 数字键 / / 过滤的完整实现
- config.KeymapConfig 加 ActionBar 字段

### 额外风险
[...]
```

---

## 12. 不在范围(明确列出)

明确**不做**(等后续 BR):

1. **TASK-019 高级动作接入**(`ActionContainerCopy/Update/Diff/Export/Commit/Wait`):等 BR-033 done
2. **Image Import 入口**(`ActionImageImport`):等 BR-036 done
3. **Registry Login 入口**(`ActionRegistryLogin`):等 BR-037 done
4. **Image History 顶层页**(`H` 键):等 BR-034 done
5. **Events 独立面板**(`F3` 键):等 BR-035 done
6. **i18n 完整 key 文案**:等 BR-039-D
7. **Help 页 "Action Bar" 段落**:等 BR-039-D
8. **fuzzy filter (当前用简单子串包含)**:BR-039-B 增强
9. **Action Bar 浮层滚动(列表 > bodyRows 时)**:本轮默认 lists ≤ bodyRows/2,无滚动需求;BR-039-B 视情况加

设计文档完。
