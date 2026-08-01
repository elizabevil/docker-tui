# BR-039 主模型反馈记录(2026-08-01)

> 用途: 记录主模型对 BR-039 只读代码定位报告的审查反馈
> 受众: BR-039-A 实施阶段(包括主模型自身 / 小模型)
> 状态: 反馈已采纳,BR-039 已完成;本文作为审查记录保留

## 反馈总览

主模型确认:BR-039 整体链路定位(Q / Ctrl+C / Esc / :command / 各 panel action 入口)正确。
但**不能直接按报告实施**,存在 5 个关键设计错误需要修正。

## 5 个关键设计错误(必须修正)

### 1. ActionQuit handler 不能改成 toast

**原报告错误**: 在 `keyboard/actions.go:18` 建议把 `case ActionQuit: return m, tea.Quit` 改为 `ShowToastNow(...)`。

**问题**: `ActionQuit` 在 `keys/registry.go:38` 同时绑定 `KeyQ` 和 `KeyCtrlC`。如果改 handler,Ctrl+C 也只 toast 不能退出。

**正确做法**:
- `keyboard/actions.go:18` `case ActionQuit: return m, tea.Quit` **保持不变**
- 只从 `keys/registry.go:38` 移除 `KeyQ`,保留 `KeyCtrlC`
- `actions.go` 加独立的 `case "q"` 路径处理 Q → toast 引导 "按 ; 打开 Action Bar,Ctrl+C 退出"
- 或者在 `HandleKeyPress` 早期(`keyboard.go:14-30` 间)用 `keys.IsQuit(rawKey)` 拦截

**最终语义**(主模型拍板):
- Q → 显示 toast 引导("按 ; 打开 Action Bar"),不退出
- `;` → 打开 Action Bar
- Ctrl+C → 直接退出(tea.Quit,沿用现有)

### 2. state.ActionBarState 不能保存 []keys.KeyAction

**原报告错误**: 建议 `Actions []keys.KeyAction` 字段。

**问题**: Go import cycle。`internal/tui/keys` 已 import `internal/tui/state`(比如 `keyboard/actions.go:6` `state "..."`),反向引用会循环:
```
state  ←→  keys   (cycle)
```

**正确做法**:
- `state.ActionBarState` 只保存**纯状态**:`Open bool / Selected int / Filter string`
- 动作 ID 映射在 `ui/widget/actionbar/` 子包完成(那里可以 import keys,因为 ui 包依赖 keys 不依赖 state)
- 或者用 `int` 索引(指向 actionbar 维护的 []ActionItem 列表),不直接持有 `KeyAction`

**state/actionbar.go 形状**:
```go
type ActionBarState struct {
    Open     bool
    Selected int
    Filter   string
}
```

### 3. ui/action.Shortcut 不能反推动作

**原报告错误**: 建议直接复用 `action.Global(app) / ForPanel(...) []Shortcut` 作为 Action Bar 内容。

**问题**: `Shortcut{Key, Description}` 只有**显示文本**,没有 `KeyAction` ID,无法在用户 Enter 时反推调用哪个 `handleAction(action, m, cmds)` 分支。

**正确做法**:
- 新建 **带 KeyAction 的统一 Action Bar 条目**结构,例如:
  ```go
  // internal/tui/ui/widget/actionbar/registry.go
  type ActionItem struct {
      Key         string         // 显示键位(e.g. "Ctrl+T" / "Q" / "—")
      Label       string         // 显示名(e.g. "Tag image")
      Action      keys.KeyAction  // 触发的动作 ID
      Description string         // 副标题(可选)
      Disabled    bool
  }
  func ActionsFor(m *state.AppModel, panel state.PanelType) []ActionItem
  ```
- 渲染走 `ActionItem` 列表(不是 `Shortcut`)
- `ui/action/registry.go` 的 `Shortcut` **保留**(给 footer/help 投影用),不动
- footer 的 `Q Quit` 行在 BR-039-A 同步删除(见 §4)

### 4. Footer / Help "Q Quit" 必须 BR-039-A 同步移除

**原报告错误**: 建议"BR-039-D 与 i18n / Help 同步"。

**问题**: 如果留到 BR-039-D,BR-039-A 到 BR-039-D 之间(可能数天 / 数周)用户按 Q 后会**无声**退出而 Help 还显示 "Q Quit",**界面与行为不一致**,误触发的退出无法解释。

**正确做法**:
- BR-039-A 同步改 `ui/action/registry.go:37`:
  ```go
  // 删除这一行:
  // {bindingLabel(model, keys.ActionQuit, keys.KeyQ), i18n.T("key.quit")},
  ```
- `ui/widget/footer/footer.go` 的渲染逻辑**不动**(它读 action.Global,Global 没 Quit 后 footer 自动没)
- i18n key 文本保留(后续 BR-039-D 再删)

### 5. 不要先 hardcoded 占位

**原报告错误**: 建议"BR-039-A 动作列表 hardcoded 占位,BR-039-B 改为动态"。

**问题**: 第一阶段建 typed registry,接入**已有可执行动作**;不要先放占位再重构,会引入 2 次设计变更。

**正确做法**:
- BR-039-A 就建立 `ActionItem{Key, Label, Action KeyAction, ...}` 的 typed registry
- `ActionsFor(m, panel)` 返回**已实现**的动作(e.g. containers 的 Start/Stop/Restart/Kill/Logs/Exec/Stats/Pause/Remove 等)
- TASK-019 高级动作(仍未实现 handler,BR-033) → 暂时不列入或标 `Disabled=true`
- 不需要做的:空占位 / stub / "Action not yet implemented" 占位

## Q / ; / Ctrl+C 最终语义(主模型拍板)

| 按键 | 行为 | 涉及文件 |
|---|---|---|
| Q | toast 引导 "按 ; 打开 Action Bar" | `keys/registry.go` 移除 KeyQ + `keyboard/actions.go` 加 Q 引导 case |
| `;` | 打开 Action Bar(`m.Navigation.Mode = ModeActionBar`) | 新加 `ActionActionBar` + `KeySemicolon` + 触发 case |
| Ctrl+C | 直接退出(`m, tea.Quit`) | `keys/registry.go` 保留 `KeyCtrlC` + `handleAction.ActionQuit` 保持 `tea.Quit` |
| Esc | Action Bar 打开时关闭(回到 ModeNormal);非 Action Bar 时走原 `handleBackAction` 双段确认 | `keyboard/keyboard.go:dispatchByMode(ModeActionBar)` 优先消费 Esc |

## 修正后的 BR-039-A 范围

### 必做

- 新增 `internal/tui/state/actionbar.go`:`ActionBarState{Open, Selected, Filter}`(**无 keys 引用**)
- 新增 `internal/tui/state/app.go:ModeActionBar` 枚举
- 新增 `internal/tui/keys/letters.go:KeySemicolon = ";"`
- 新增 `internal/tui/keys/action.go:ActionActionBar`
- 新增 `internal/tui/keys/registry.go`:`{ActionActionBar, []string{KeySemicolon}, main}` + 从 `ActionQuit` 移除 `KeyQ`
- 新增 `internal/tui/keyboard/actions.go:case ActionActionBar: doActionBar(m)` + 改 Q 引导 case
- 新增 `internal/tui/keyboard/keyboard.go:dispatchByMode(ModeActionBar)` + 拦截 Q 走 toast
- 新增 `internal/tui/keyboard/actionbar_keys.go`:`handleActionBarKeys`
- 新增 `internal/tui/ui/widget/actionbar/registry.go`:`ActionItem` + `ActionsFor(m, panel) []ActionItem`(typed,**接已有 handler**)
- 新增 `internal/tui/ui/widget/actionbar/actionbar.go`:`RenderBar(m, content, body) string` + 通过 `dialog.PlaceDialogInPanel` 居中
- 改 `internal/tui/ui/app/layout.go`:加 `ModeActionBar` 分支
- 改 `internal/tui/ui/action/registry.go:37`:删除 Q Quit 行(同步 Footer / Help)

### 必不做

- 不写 hardcoded 占位 action,直接接已有可执行动作
- 不动 `handleAction.ActionQuit`(仍走 `tea.Quit`,只通过注册表去掉 Q 绑键)
- 不在 `state` 引用 `keys`(避免 import cycle)
- 不动 `widget/footer/footer.go` 渲染逻辑(Global 没 Quit 后自动消失)
- 不写完整 6 个 panel 的全部动作清单,先建立 typed registry + 接 containers / images / networks / volumes / compose / audit 中已有 handler 的动作;后续 BR-039-B 再补全

### 验收标准(BR-039-A)

- 按 Q 显示 toast 引导,不退出
- 按 `;` 打开 Action Bar,列出当前 panel 可执行动作(typed ActionItem)
- j/k 上下移动 + Enter 执行 + Esc 关闭
- 数字键 1-9 跳到第 N 项
- Action Bar 内容随 `m.Navigation.ActivePanel` 动态变化
- Ctrl+C 仍能直接退出
- Footer / Help 不再显示 "Q Quit"
- 所有现有功能不退化
- 无 Go import cycle

## 与其他 BR 的协同

- BR-040(`PlaceDialogInPanel`):已 done,Action Bar 复用居中数学
- BR-033(TASK-019 高级动作):**BR-039-A 不接入**,BR-039-B 接入(那时 BR-033 已 done)
- BR-034(镜像 History):BR-039-A 不涉及
- BR-036(Image Import):**BR-039-A 不接入**,后续接入
- BR-037(Registry Login):**BR-039-A 不接入**,后续接入
