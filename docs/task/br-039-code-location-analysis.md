# BR-039 只读代码定位分析报告

> **归档说明**: 本文是实施前的历史分析,包含已被主模型反馈纠正的方案,不得作为当前实现说明。最终结果以 `br-039-action-bar-replace-q.md` 与源码为准。

> 建立日期: 2026-08-01
> 来源: 小模型执行主模型 prompt("只读分析,不改代码")
> 目标需求: BR-039 取消 Q 退出,改为每页多功能 Action Bar
> 参考任务卡: [docs/task/br-039-action-bar-replace-q.md](br-039-action-bar-replace-q.md)
> 关联任务卡: [docs/task/br-040-dialog-style-center-on-panel.md](br-040-dialog-style-center-on-panel.md)(已 done)

> 本报告只读分析,未修改任何代码。
> BR-039 任务卡"建议任务分解"中 BR-039-A 阶段(只做弹出/选择/关闭,Execute 为占位)的输入。

## 1. 当前 Q / Ctrl+C / Esc / :command 键盘链路

### 1.1 Q 链路(目标:取消,改为 Action Bar 引导)

```
User 按 Q
  → bubbletea 触发 tea.KeyPressMsg{String()="q"}
  → keyboard.HandleKeyPress (keyboard.go:14)
      rawKey="q", key=keys.Normalize("q")="q"
      → 模式分发不命中(ModeNormal 不在 case 中)
      → handleNestedViewKeys 不命中
      → resolveAction(key, m) 命中 ActionQuit(registry.go:38:ActionQuit ← KeyQ, KeyCtrlC, app)
      → handleAction(ActionQuit, m, cmds) (actions.go:18)
          case ActionQuit: return m, tea.Quit
      → bubbletea 接 tea.Quit,退出应用
```

关键锚点:
- 注册:`keys/registry.go:38` `{ActionQuit, []string{KeyQ, KeyCtrlC}, app}`
- 字符串:`keys/letters.go:20` `KeyQ = "q"`
- 处理:`keyboard/actions.go:18-19` `case ActionQuit: return m, tea.Quit`
- Footer 投影:`ui/action/registry.go:37` `{bindingLabel(model, ActionQuit, KeyQ), i18n.T("key.quit")}`
- Footer 渲染:`ui/widget/footer/footer.go:13` 调用 `action.Global(app)` 后 `action.ForMode` / `action.ForPanel` 拼出两行 shortcut bar
- 全局快捷键文本 keymap 配置:`keys/configuredBindings` 读 `keymap.Quit`(`config/types.go:27`)
- Label 常量:`keys/display.go:61` `ActionLabelQuit = "Quit"`

### 1.2 Ctrl+C 链路(目标:保留退出,EscPending 双段确认机制保留)

```
User 按 Ctrl+C
  → key="ctrl+c"
  → resolveAction → ActionQuit
  → handleAction → tea.Quit
```
Ctrl+C 与 Q 共用 `ActionQuit`(registry.go:38),直接退出,不走 EscPending。EscPending 是**Esc** 的双段确认机制,与 Q/Ctrl+C 无关。

### 1.3 Esc 链路(目标:无影响,继续做 Action 的 mode 退出)

```
User 按 Esc
  → keyboard.HandleKeyPress (keyboard.go:25)
      若 Feedback.ErrorMessage != "" → ClearError
      dispatchByMode 命中(ModeDetail / ModeLogView / ModeFilter / ModeConfirm / ModeImageTransfer / ModeCommand / ModeRuntimeSelect / ...),由各 mode handler 消费
      否则 resolveAction → ActionBack
  → handleAction(ActionBack, m, cmds) → handleBackAction(m) (actions.go:128)
      逐级判断 ModeDetail / ModeLogView / ModeFilter / 子视图 ...
      最后若 EscPending=true → tea.Quit,清 EscPending
      否则 EscPending=true + InfoMessage="Press Esc again to quit" + 起 3s EscTimeout tick
```
Esc 是**双段退出**(EscPending)机制,与 Q/Ctrl+C 独立。Action Bar 应**不接管** Esc,以免打断此机制。

### 1.4 :command 链路(目标:保留,Action Bar 与之并存)

```
User 按 :
  → key=":"(registry.go:49 `{ActionCommand, []string{":"}, app}`)
  → resolveAction → ActionCommand
  → handleAction → ToCommand(m) (navigate.go:122)
      m.Navigation.Mode = state.ModeCommand
      m.Navigation.CommandInput.Reset()
  → dispatchByMode(ModeCommand) → handleCommandInput(key, m) (command.go:12)
      Enter → executeCommand(m):switch cmd 派发 (compose/images/containers/volumes/networks/logs/help/rename/top/port)
      Esc → ModeNormal + CommandInput.Reset()
      Tab → autocomplete
      default → editQueryInput
```
`:` 命令面板走 `ModeCommand` + 独立 `CommandInput` 字段(navigation.go:13)。Action Bar 若要并存,必须使用**独立的 Mode**(如 `ModeActionBar`),**不能**复用 `ModeCommand`,否则 j/k/Enter/Esc 会与命令面板的编辑/补全冲突。

### 1.5 键盘链路总图

```
Tea KeyPress
  │
  ├─ HandleKeyPress (keyboard.go:14)
  │   ├─ Esc 清错 + 清非 Esc 按键的 InfoMessage
  │   ├─ dispatchByMode (按 m.Navigation.Mode 分发)
  │   │   ├─ ModeHelp    → handleSelectionDialog 内 ActionHelp/Back
  │   │   ├─ ModeExecPassthrough
  │   │   ├─ ModeFilter
  │   │   ├─ ModeSearch
  │   │   ├─ ModeImagePull
  │   │   ├─ ModeImageWorkflow
  │   │   ├─ ModeImageTransfer (仅 Esc 走取消)
  │   │   ├─ ModeCommand     → handleCommandInput
  │   │   ├─ ModeRuntimeSelect
  │   │   ├─ ModeRename
  │   │   ├─ ModeResourceCreate
  │   │   ├─ ModeTop
  │   │   ├─ ModeAuditDetail
  │   │   ├─ ModeExec
  │   │   ├─ ModeConfirm
  │   │   └─ ModeMark
  │   ├─ handleNestedViewKeys (image→containers / compose→services)
  │   ├─ resolveAction + handleAction (全局 action 表)
  │   └─ handlePanelFallbacks (image/compose/audit panel 局部)
  └─ handleShortcuts (Space mark / H header / C conn / O / Ctrl+O sort)
```

`ActionActionBar` 应作为新 key 注册到 `registry.go`,新增 `case ActionActionBar: doActionBar(m)` 到 `handleAction`。`ModeActionBar` 加到 `AppMode` 枚举,`dispatchByMode` 加 case 把按键路由到 `handleActionBarKeys`。

## 2. 当前每个 panel 已有 action 来源

### 2.1 按 `ActionLabel` 与 `doXxx` 入口统计

| Panel | 入口函数 | 已实现 action(`registry.go` + 实际 handler) |
|---|---|---|
| **Containers** | `keyboard/actions.go:56-64` `doContainerAction` 模板 | Start(s) / Stop(Ctrl+S) / Restart(Ctrl+R) / Kill(Ctrl+K) / Remove(Ctrl+D) / Logs(l) / Detail(d) / Exec(e) / Inspect(i) / Stats(m) / Pause(p) |
| **Images** | `keyboard/image_action.go:doImageAction` + `image_transfer.go:openImageWorkflow` | Pull(Ctrl+P) / Prune(p) / Remove(Ctrl+D) / Detail(d) / Tag(Ctrl+T) / Push(Ctrl+U) / Save(Ctrl+E) / Load(Ctrl+L) |
| **Volumes** | `keyboard/volume_action.go:doVolumeAction` | Create(c) / Prune(p) / Remove(Ctrl+D) |
| **Networks** | `keyboard/network_action.go:doNetworkAction` | Create(c) / Prune(p) / Remove(Ctrl+D) |
| **Compose** | `keyboard/compose_action.go:doComposeAction` | Start(s) / Stop(Ctrl+S) / Logs(l) / Detail(d) |
| **Audit** | `keyboard/audit_keys.go:handleAuditKeys` | Enter/Detail / Filter level / /Filter |
| **Common** | `keyboard/actions.go` | Help(F1/?) / Filter(/) / Refresh(r) / Tab(>)/Shift+Tab(<) / SwitchRuntime(F2) / RefreshConnections(R,F12) / Command(:) |

### 2.2 TASK-019 高级动作(已注册 handler 缺失)

| Action | Key | handler 状态 |
|---|---|---|
| ActionContainerUpdate | Ctrl+W | ✗ 未实现(见 BR-033) |
| ActionContainerDiff | F2 | ✗ 冲突(SwitchRuntime) |
| ActionContainerExport | Ctrl+X | ✗ 未实现 |
| ActionContainerCommit | Ctrl+K | ✗ 冲突(ContainerKill) |
| ActionContainerWait | Ctrl+Y | ✗ 未实现 |
| ActionContainerCopy | Ctrl+O | ✗ 冲突(sort column) |

**Action Bar 是这些无键位 / 冲突动作的入口**。如果按 BR-039 把它们放进 Action Bar 而不绑键,这些动作即可触发且无键位冲突。

### 2.3 Footer / Help 投影来源

`ui/action/registry.go` 已经有完整的 per-panel / per-mode / global shortcut 投影框架:

```go
// ui/action/registry.go:22
func Global(app ...*state.AppModel) []Shortcut
// 列出全局 shortcut:Q Quit / F1 Help / ? / F2 Runtime / Tab / Filter / Refresh / :Command /

// ui/action/registry.go:132
func ForPanel(panel state.PanelType, marked int, app ...*state.AppModel) []Shortcut
// 列出当前 panel 的 shortcut:containers / images / volumes+networks / compose / audit

// ui/action/registry.go:41
func ForMode(app *state.AppModel) []Shortcut
// 列出当前 mode 的 shortcut:ModeFilter / ModeSearch / ModeCommand / ModeMark / ModeConfirm ...
```

**Action Bar 的内容可以直接复用这套投影函数**(或稍加调整):把 `Shortcut{Key, Description}` 转成 `ActionItem{Label: Description, Key: Key, Action: <对应 KeyAction>}`。

### 2.4 UI 调用链

```text
RenderApp (ui/app/layout.go)
  └─ action.Global() + action.Context() (PerPanel + PerMode) → footer.Shortcuts()
  └─ footer.Render() → 固定两行 footer rail (action + shortcut 描述)
```

Action Bar 的渲染可放在 layout.go 类似 dialog 分支的位置(在 ModeActionBar 时替换或叠加内容)。

## 3. Action Bar 最小状态模型建议

参考 `state.ConfirmState` + `state.NavigationState.CommandInput` + `state.DialogState` 的现有模式,**最小状态模型**:

```go
// internal/tui/state/actionbar.go (新文件)
type ActionBarState struct {
    Open       bool   // 浮层是否打开
    Filter     string // 过滤输入
    Selected   int    // 选中条目索引(过滤后)
    Actions    []keys.KeyAction // 当前 panel 的可执行动作(可由 ActionItem 包装)
    // 备选更富信息:
    // Items   []ActionItem{ Label, Key, Action KeyAction, Description }
}

type ActionItem struct {
    Label       string  // e.g. "Tag"
    Key         string  // e.g. "Ctrl+T" 或 "(no key)"
    Action      keys.KeyAction
    Description string  // e.g. "重命名镜像标签"
}
```

要点:
- `Open bool` 而不是新 `Mode`,避免动 `state.AppMode` 枚举?——**不**,新 mode 必须加,否则与现有 mode dispatch 模式不一致(参考 `ModeCommand` / `ModeFilter` 都是新 mode)。
- 用 `NavigationState.ActionBar` 嵌入,跟 `FilterInput` / `CommandInput` 同层(见 `state/navigation.go:7-17`)。
- 关闭时 `Reset()`,类似 `ConfirmState.Close()` 与 `QueryInputState.Reset()`。
- `Filter` 字段实现 fuzzy match:用户在 Action Bar 内按 `/` 后输入字符串,实时过滤。
- 选中项由 `Selected int` 维护,`MoveSelection(delta int)` 方法移动。

参考现有模式:
- `ConfirmState.MoveFocus` (state/confirm.go:30):`if disabled: skip` 跳过 Disabled
- `QueryInputState.Move` (state/navigation.go:72):clamp 到 [0, len]
- `DialogState.MoveFocus` (state/dialog.go:83):`% count`

`ActionBarState` 复用 `Reset()` 模式 + 边界 clamp。

## 4. Action Bar widget 应放在哪个包 + 如何复用 BR-040

### 4.1 建议包位置

**`internal/tui/actionbar/`**(新子包,与 dialog / footer / header / panel 平级)

理由:
- Action Bar 是**浮层 widget**,与 dialog 同类
- 与 dialog 不同的只是内容(命令面板 vs 操作列表)
- BR-040 已就绪,直接 `dialog.PanelBody` 传入 `dialog.PlaceDialogInPanel` 复用居中逻辑

**不需要**新增顶层 `internal/tui/ui/actionbar/` 目录,因为:
- widget 概念上是渲染组件
- 已有 `internal/tui/ui/action/` 包,但它是 shortcut label 投影服务,不是 widget

### 4.2 复用 BR-040

Action Bar 渲染流程:

```text
internal/tui/actionbar/actionbar.go:RenderBar(m, body)
  → 生成 ActionItem 列表(根据 m.Navigation.ActivePanel / m.ActionBar.Filter)
  → 计算高度(行数 = len(items) + 1 标题行)
  → 渲染为 lipgloss 风格的多行 list(类似 fuzzy finder)
  → 调 dialog.PlaceDialogInPanel(content, box, body, cfg)
```

即 Action Bar 的"内容生成"与"位置居中"完全分离:
- 内容由 actionbar widget 负责(类似 ChoiceDialog / SelectionDialog)
- 居中由 BR-040 的 `PlaceDialogInPanel` 负责(已在 widget/dialog 包)

不需要复制 `CenterOnPanel` 的数学。

### 4.3 触发键

候选:
- `;` (vim 风格,与现有 `:` / `/` 不冲突,registry.go 无该键)
- `\` (反斜杠,Linux 习惯)
- `a` (action 首字母)
- `b` (bar 首字母)

**推荐 `;`**:
- 与现有 `:` `/` `F1` `F2` `R` `Esc` `Tab` 等无冲突
- vim 用户熟悉
- 易于手按(右手小指,无需移动)

需主模型确认(任务卡"待确认项"已列)。

## 5. BR-039-A 最小实现边界

> 主模型提示:"只做 Action Bar 框架 + Q 打开 + Esc 关闭 + Enter 执行动作占位 + Ctrl+C 保留退出。"

### 5.1 BR-039-A 范围(本次)

| 功能 | 是否纳入 BR-039-A |
|---|---|
| Q 键:取消退出,改显示 toast 提示"按 ; 打开 Action Bar" | ✓ |
| ; 键:打开 Action Bar 浮层 | ✓ |
| Esc 键:关闭 Action Bar(ModeActionBar → ModeNormal) | ✓ |
| Enter 键:执行当前选中项(占位:调 `handleAction(action, m, cmds)`) | ✓ |
| j/k 键:上下移动选中 | ✓ |
| 数字键 1-9:跳到第 N 个 | ✓ |
| / 键:filter 输入(简化:后续 BR 增强) | ⏸ 留到 BR-039-B |
| Esc 双段确认机制:不动 | ✓ 不影响 |
| Ctrl+C 退出:不动 | ✓ 保留 |
| :command 面板:不动 | ✓ 保留 |
| 各 panel 动作列表(per-panel 完整):留到 BR-039-B | ⏸ 占位用"无 actions" |
| Footer / Help 文案更新:留到 BR-039-D | ⏸ 后续 |
| i18n key 全量:留到 BR-039-D | ⏸ 后续 |
| TASK-019 高级动作接入:留到 BR-033 | ⏸ 后续 |

### 5.2 BR-039-A 必改文件

| 文件 | 改动 |
|---|---|
| `internal/tui/keys/letters.go` | 加 `KeySemicolon = ";"`(若 keys.go 不在 letters.go 集中管理) |
| `internal/tui/keys/action.go` | 加 `ActionActionBar KeyAction = "actionBar"` |
| `internal/tui/keys/registry.go` | 加 `{ActionActionBar, []string{KeySemicolon}, main}` |
| `internal/tui/keys/registry.go:38` | 改 `{ActionQuit, []string{KeyQ, KeyCtrlC}, app}` → `{ActionQuit, []string{KeyCtrlC}, app}`(移除 KeyQ) |
| `internal/tui/state/app.go:38-61` | 加 `ModeActionBar` 枚举值 |
| `internal/tui/state/navigation.go:7-17` | 加 `ActionBar ActionBarState` 字段 |
| `internal/tui/state/actionbar.go`(新) | `ActionBarState` struct + `Open` / `Close` / `MoveSelection` / `SetFilter` 方法 |
| `internal/tui/keyboard/actions.go:18` | `case ActionQuit` 改为 `ShowToastNow` + 引导 Q 改为 Action Bar(不再 tea.Quit) |
| `internal/tui/keyboard/actions.go:16` | 加 `case ActionActionBar: return doActionBar(m)` |
| `internal/tui/keyboard/keyboard.go:60-112` | `dispatchByMode` 加 `case state.ModeActionBar: ...` 调用 `handleActionBarKeys` |
| `internal/tui/keyboard/actionbar_keys.go`(新) | `handleActionBarKeys(key, m)` 函数 |
| `internal/tui/actionbar/`(新子包) | 至少一个 `actionbar.go` 含 `RenderBar(m, content, body) string` |
| `internal/tui/actionbar/registry.go`(新) | `ActionsFor(m) []keys.KeyAction` 占位 — BR-039-A 只返回 hardcoded 列表,BR-039-B 改为动态 |
| `internal/tui/actionbar/state.go`(新) | 简单 state helper(如果不用 state 包) |
| `internal/tui/ui/app/layout.go` | 加 `if m.Navigation.Mode == state.ModeActionBar { ... }` 分支调 `widget/actionbar.RenderBar(m, body)` + `dialog.PlaceDialogInPanel(content, box, body, cfg)` |
| `internal/tui/ui/action/registry.go:37` | 删 `ActionQuit` 行(或保留,等 BR-039-D) |
| `internal/tui/keys/display.go:61` | 保留 `ActionLabelQuit` 给 BR-039-D 改文案 |

### 5.3 BR-039-A 不应触碰的文件

- `runtime/*` / `state/runtime/*` —— 完全不碰(主模型明确)
- `widget/dialog/centered.go` / `overlay.go` —— BR-040 已完,只复用 `dialog.PlaceDialogInPanel`
- `keyboard/actions.go:56-122` —— container/image/volume/network/composition action handler 不动
- `widget/footer/footer.go` —— footer 内容不调整(BR-039-D 处理)

## 6. 需要新增或修改哪些测试

### 6.1 新增

| 文件 | 测试内容 |
|---|---|
| `internal/tui/state/actionbar_test.go`(新) | `TestActionBarOpen/Close/Reset/MoveSelection/Filter` 状态机测试 |
| `internal/tui/keyboard/actionbar_keys_test.go`(新) | 测 `handleActionBarKeys` 各按键分支:Enter → handleAction / j → MoveSelection / Esc → Close / 数字键 → JumpToN |
| `internal/tui/actionbar/actionbar_test.go`(新) | `TestRenderBarPanelCentered` —— 验证 Action Bar 浮层用 BR-040 `PlaceDialogInPanel` 居中于 panel body |
| `internal/tui/actionbar/registry_test.go`(新) | `TestActionsFor` 占位测试(返回 hardcoded 列表) |

### 6.2 修改

| 文件 | 修改 |
|---|---|
| `internal/tui/keys/registry_test.go:14` | `ActionQuit` 不再含 `q` 键,断言改为 `len(got) == 0 \|\| got[0] != "ctrl+c"` |
| `internal/tui/keys/judge_test.go`(若有) | `IsQuit` 现在永远 false(Q 已不是 quit 键),改为 `TestIsQuitAfterQRemoval` |
| `internal/tui/ui/action/registry_test.go`(若有) | `Global` 列表中 `Quit` 项应删除,断言调整 |

### 6.3 不动

- `internal/tui/keyboard/actions_test.go` (若有) —— ActionQuit case 改了语义但其他 case 不动
- `internal/tui/ui/widget/dialog/centered_test.go` —— BR-040 测试,BR-039 不动

## 7. 风险点

| 风险 | 说明 | 缓解 |
|---|---|---|
| **Q 键完全取消** | 旧用户习惯"按 Q 退出"被破坏;`ActionQuit` 仍含 KeyCtrlC,真正退出路径保留 | Q 改为 toast 提示"按 ; 打开 Action Bar,Ctrl+C 退出" |
| **Ctrl+C 与 EscPending 关系** | Ctrl+C 直接退出(不走 EscPending);若用户期待"Ctrl+C 也是双段"会困惑 | 帮助文案明确说明 |
| **命令面板冲突** | Action Bar 用 `;`,命令面板用 `:`,互不冲突;但若用户预期 Action Bar 是 `:` 的替代品会困惑 | 帮助文案说清"`: /` 命令,/ Action Bar" |
| **Footer 显示 Q Quit** | `action/registry.go:37` 当前在 Global 列表里有 Q Quit,需要删 | BR-039-A 不改 footer(BR-039-D 处理),但可考虑在 BR-039-A 加 TODO 注释 |
| **空页面行为** | detail / logs / help 等无 panel action 的页面,Action Bar 显示什么?任务卡建议"空 list" | BR-039-A 占位:空 list 时不打开 Action Bar(`ActionsFor(m)` 返回空时 no-op) |
| **dialog.jsonc 配置** | Action Bar 走 BR-040 的 `dialogConfig`,WidthPercent=25 / MinWidth=40 是否合适 panel 尺寸? | BR-039-A 用默认 cfg;BR-039-D 调整 |
| **i18n 漂移** | "按 ; 打开 Action Bar"等提示需多语言 | BR-039-A 用英文 + i18n.T() key 占位;BR-039-D 补全 |
| **Help 漂移** | Help 页(ForMode)需要加 Action Bar 段落 | BR-039-D 处理 |
| **TASK-019 动作未接** | BR-033 的 Copy/Update/Diff 等未实现,Action Bar 即使列出也无法执行 | BR-033 接 handler,BR-039-A 占位用 stub("Action not yet implemented") |
| **键位冲突(`;`)** | 现有快捷键无 `;`,但需确认未来 `keymap.ActionActionBar` 字段可被用户重映射 | `configuredBindings` 加 `ActionActionBar: keymap.ActionBar` |
| **focus 状态机** | Action Bar 打开时,j/k / Enter / Esc 需"被 Action Bar 消费"而不是触发全局 j/k(向下移动光标) | `dispatchByMode(ModeActionBar)` 优先返回 handled=true |
| **panel rect 失效** | 若 `ResolveLayout(m)` 返回 `TerminalUnsupported`(`< 60×14`),`bodyLeft/bodyWidth` 可能为 0,`CenterOnPanel` 会把对话框 clamp 到 1 列 | 已在 BR-040 的 maxW/maxH fallback 中处理 |
| **ModeActionBar 与子视图冲突** | image→containers / compose→services 子视图打开时,Action Bar 是否可用? | 建议:子视图打开时不进入 Action Bar(`ModeNormal` 即可),符合 BR-039 任务卡"依赖:所有 panel 已有功能" |

## 8. 需要主模型确认的问题

| 问题 | 选项 | 倾向 |
|---|---|---|
| **触发键** | (A) `;` vim 风格 (B) `\` 反斜杠 (C) `a` action (D) `b` bar (E) 复用 `:` | (A) `;` 与现有 `:` / `/` / `F1` / `F2` / `R` 全无冲突;vim 用户熟悉 |
| **Q 键重映射** | (A) 取消 Q,按 Q 显示 toast 引导 (B) Q 静默 no-op (C) Q 仍 quit | (A) 符合任务卡"显示 Q 重映射到 Action Bar" |
| **动作列表来源** | (A) 硬编码占位(BR-039-A) (B) 从 `ActionRegistry` 动态推导(每行 `Shortcut{Key, Description, Action KeyAction}`) (C) 完全硬编码每个 panel 的 []ActionItem | BR-039-A 用 (A),BR-039-B 用 (B),BR-039-D 覆盖 (C) |
| **空页面行为** | (A) 空 list + 提示 "No additional actions" (B) 直接 no-op(不打开) (C) 仍打开但显示通用 actions(Refresh / Switch Runtime / Help / Quit) | 任务卡建议 (A),但 (C) 实际更实用 —— 镜像页和 detail 页都需 Refresh / Help |
| **多 tab(多 panel 同时挂)** | (A) 不支持,一次只能一个 panel 的 Action Bar (B) 支持栈,`Esc` 逐级退出 | (A) 简化复杂度,任务卡"待确认项"建议不 |
| **与命令面板交互** | (A) Action Bar 内 `:` 打开命令面板(嵌套) (B) Action Bar 内 `:` 被忽略,只 `: `command` 可独立打开 | (A) 嵌套;但复杂度高。BR-039-A 选 (B) |
| **数字键跳转** | (A) 1-9 全支持 (B) 仅 1 (C) 0-9 全支持 | (A) 与 vim command palette 一致 |
| **Footer "Q Quit" 何时移除** | (A) BR-039-A 同步删 (B) BR-039-D 与 i18n / Help 同步 | (A) 与 BR-039-A 一致:删 footer Quit 行 → 防止误导用户 |
| **Action Bar 关闭后焦点** | (A) 回到 panel 列表 (B) 不变(panel 光标位置不动) | (B) 默认行为,光标不动 |
| **Action Bar 调起时的 panel 状态** | (A) 记录 open 时的 panel / cursor,close 后恢复 (B) 始终按当前 m.Navigation.ActivePanel 重新计算 | (B) 简单,Action Bar 永远反映当前 panel |
| **Enter 选中"无键位"动作** | (A) 调 `handleAction(action, m, cmds)` 即使没绑键 (B) 跳 toast "Action not yet implemented" | (A) 选 (A);handleAction 已存在,只需 case 命中 |
| **panel 切换时是否关闭 Action Bar** | (A) 切 panel 自动 close Action Bar (B) Action Bar 内容随 panel 变化(open 状态不变) | (B) 简单;用户能用 Action Bar 在两个 panel 间切换动作 |

## 9. 关键文件路径汇总

| 类别 | 文件 |
|---|---|
| **任务卡** | `docs/task/br-039-action-bar-replace-q.md`(目标 / 范围 / 风险) |
| **关联已完** | `docs/task/br-040-dialog-style-center-on-panel.md` + `internal/tui/ui/widget/dialog/centered.go`(`PlaceDialogInPanel`) |
| **键盘主分发** | `internal/tui/keyboard/keyboard.go:14`(`HandleKeyPress`) + `60-112`(`dispatchByMode`) |
| **action 注册** | `internal/tui/keys/registry.go:38`(`ActionQuit`) + `49`(`ActionCommand`) + `40-85`(全部 default keys) |
| **action 处理** | `internal/tui/keyboard/actions.go:16-126`(`handleAction` switch) |
| **按键字符串** | `internal/tui/keys/letters.go:20`(`KeyQ`) + `keys.go:19`(`KeySlash`) + `KeyColon` 不存在(`:` 直接用字面量) |
| **Mode 枚举** | `internal/tui/state/app.go:38-61`(新增 `ModeActionBar` 在此插入) |
| **Nav 字段** | `internal/tui/state/navigation.go:7-17`(新增 `ActionBar` 字段在此) |
| **现有 dialog-like state** | `internal/tui/state/confirm.go`(参考 `Open/MoveFocus/Close` 模式) + `state/dialog.go` |
| **panel action 入口** | `internal/tui/keyboard/container_action.go:19`(`doContainerAction` 模板) + `image_action.go` / `volume_action.go` / `network_action.go` / `compose_action.go` / `audit_keys.go` |
| **shortcut 投影框架** | `internal/tui/ui/action/registry.go:22`(`Global`) + `41`(`ForMode`) + `132`(`ForPanel`) |
| **footer 渲染** | `internal/tui/ui/widget/footer/footer.go:12`(`Shortcuts`) |
| **layout 路由** | `internal/tui/ui/app/layout.go:101-256`(`RenderApp`,`ModeConfirm` / `IsSelection` / `DialogExec` 三个 dialog 分支是参考) |

## 10. 给主模型的总结

- **现状清晰**:Q 走 `registry.go:38` → `actions.go:18-19` → `tea.Quit`;`:` 走独立 `ModeCommand` 路径(`command.go:12`);Esc 走 `handleBackAction` 双段确认;Ctrl+C 与 Q 共用 `ActionQuit` 但**不**走 EscPending。各 panel action 入口都集中在 `keyboard/actions.go:56-122`,模式清晰。
- **最小实现**:BR-039-A 只做框架 + Q→toast + ; 打开 + Esc 关闭 + Enter 执行(调 `handleAction`)+ j/k 移动 + 数字键跳转。动作列表用 hardcoded 占位(每 panel 返回固定 []KeyAction),BR-039-B 改为动态。
- **包位置**:`internal/tui/actionbar/` 新子包,通过 `dialog.PlaceDialogInPanel(content, box, body, cfg)` 复用 BR-040 居中数学。
- **Mode 复用**:新 `ModeActionBar` 枚举,`dispatchByMode` 加 case;`ActionBarState` 嵌入 `NavigationState`。
- **8 个待确认问题**中,触发键 `;` + 动作列表 hardcoded + 空页面通用 actions + 单 panel 焦点是 BR-039-A 关键决策。

报告完成。**未修改任何代码**。
