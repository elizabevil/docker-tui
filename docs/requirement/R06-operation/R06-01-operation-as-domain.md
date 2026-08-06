# R06 Operation as First-Class Domain Concept

## Status

✅ **R06-01 — theme.action.<scope> 视觉契约** 落地
✅ **R06-02 — Operations registry loader** 落地
✅ **R06-03 — Single dispatcher (Operation 一等公民)** 落地
✅ **R06-04 — theme 重组 (方案 B)** 落地
✅ **R06-05 — Page/Async 渲染器接 theme chrome** 落地
✅ **R06-06 — imagePrune async Operation** 落地

剩余 follow-ups:
- R06-07 — 状态合并: `m.Form + m.Processes + m.ContainerWait + m.Detail` → `m.Operation` discriminated union
- R06-08 — Rename 走 Form mode(目前还走 legacy ModeRename + Dialog)

---

## Background

容器操作功能最初把三个关注点混在一个主题槽位里:

1. **Style** — 窗口背景 / 边框 / form input / Confirm·Cancel 颜色,由 `FormDialog` 消费
2. **Action** — 八个 `ActionContainer*` 键绑定及其 Go handler(`openTopView`, `openPortDetail`,
   `doContainerDiff`, `doContainerWait`, `openContainerCopyForm` 等)
3. **Composition** — `internal/tui/actionbar/registry.go` 里手写的 `[]ActionItem` 列表

迭代前的主题结构按"组件"分(`border / header / main / footer / dialog / toast / text / table / safeFallback`),
改"主题危险色"得跳 5 个 section。ColorRef 是嵌套对象 `{ "token": "primary" }` / `{ "value": "#3875d7" }`,
每个颜色引用占 4 行,JSONC 里 65+ 个 wrapper 噪音。

## Goal

把 `Operation` 升格为一等公民:

```
Operation
  ├─ Kind:    unique identifier (top, port, copy, ...)
  ├─ Scope:   which resource it acts on (container, image, volume, ...)
  ├─ Mode:    how its body renders (Form | Page | Async)
  ├─ Window:  window chrome from theme.action.<scope>.window
  ├─ Body:    mode-specific (form.kind / page.body / async.body)
  └─ Actions: mode-specific (Form → Confirm + Cancel; Async → Cancel)
```

并把主题从"组件维度"重组为"视觉维度"(方案 B),作者改主题时按用途找字段:

```
theme.chrome      ← 框架 / 边框 / 标题 / 选择
theme.surfaces    ← 背景填充
theme.text        ← 文字语义色
theme.feedback    ← toast + safe fallback
theme.data        ← 表格内容色
theme.action      ← per-scope 窗口契约
```

---

## Design (current state)

### 1. Theme schema (方案 B + ColorRef 扁平化)

主题 JSONC 顶层结构:

```jsonc
{
  "theme": {
    "palette":   { "primary": "...", "success": "...", ... },
    "chrome":    { "borderKind": "rounded", "panelBorderActive": "...", ... },
    "surfaces":  { "panel": "...", "dialogBody": "...", ... },
    "text":      { "info": "...", "dim": "...", "dialogBody": "...", ... },
    "feedback":  { "toastSuccess": "...", "safeNormal": "...", ... },
    "data":      { "markedBackground": "...", "columnForeground": "...", ... },
    "action": {
      "container": { "window": {...}, "formInput": {...}, "confirm": {...}, "cancel": {...} },
      "image":     { "window": {...}, "formInput": {...}, "confirm": {...}, "cancel": {...} }
    }
  }
}
```

#### ColorRef 扁平化

`ColorRef` 是 `type ColorRef string`:
- 不带 `#` 前缀 → palette token,例如 `"primary"`、`"foregroundMuted"`
- 带 `#` 前缀 → 字面值,例如 `"#3875d7"`、`"rgb(250,240,230)"`、`"grey"`、`"63"`

`IsToken()` 用闭集判定(已知 token 名单),`ResolveColor` 按 token 查 palette,
其余按 `utils.ParseColor` 解析。完全无 wrapper,JSONC 体积降 ~60%。

### 2. Operations schema (per-scope JSONC)

```jsonc
// defaults/operations/scopes/container.jsonc
{
  "scope": "container",
  "operations": [
    {
      "kind":   "copy",
      "action": "containerCopy",
      "mode":   "form",
      "form":   { "kind": "containerCopy" },
      "label":  "Copy",
      "description": "Copy a file or directory out of the container as a tar",
      "requires":     ["engine", "container"],
      "disabledWhen": []
    },
    {
      "kind":   "top",
      "action": "containerTop",
      "mode":   "page",
      "page":   { "body": "processes" },
      "label":  "Top",
      "description": "Open the running container process view",
      "requires":     ["engine", "container", "running"]
    },
    {
      "kind":   "wait",
      "action": "containerWait",
      "mode":   "async",
      "async":  { "body": "wait" },
      "label":  "Wait",
      "description": "Wait until the container is no longer running",
      "requires":     ["engine", "container"]
    }
    // ... rename / port / diff / update / export / commit
  ]
}
```

#### Mode ↔ body sub-object 对称契约

```
mode=form   → 必填 form:   { kind: "<state.FormKind 名字>" }
mode=page   → 必填 page:   { body: "<page renderer 名>"   }
mode=async  → 必填 async:  { body: "<async renderer 名>"  }
```

Loader 在 `validateSpec` 强制该契约,违规立即报错。

#### Requires / DisabledWhen 谓词

闭集: `engine` / `container` / `image` / `running` / `manifest`。

- `requires`: AND-combined 全满足才 enabled
- `disabledWhen`: OR-combined 任一满足就 disabled

```
"requires":     ["engine", "container", "running"]   // 三者全满足
"disabledWhen": ["manifest"]                       // 是 manifest list 就禁
```

### 3. Mode dispatch pipeline

```
KeyAction (keys.ActionContainerXxx)
  ↓ keyboard/actions.go: case ActionActionBar / non-Operation actions only
  ↓
  if dispatchOperation(action, m) ok:
    return m, cmd
  ↓
dispatchOperation(action, m)
  ├─ LoadOperations() → *Operations
  ├─ ops.LookupByAction(action) → OperationSpec
  └─ switch spec.Mode:
       ├─ form   → dispatchForm(spec.Form.Kind) → openXxxForm(m)
       ├─ page   → dispatchPage(spec.Page.Body) → openXxxView(m)
       └─ async  → dispatchAsync(spec.Async.Body) → doXxx(m)
```

当前 container scope 9 个 action + image scope 2 个 action(`history` + `prune`)
全部走这条路径。`keyboard/actions.go` 里删掉了对应的 8 个硬编码 case。

### 4. Render pipeline (R06-05)

| Mode | Body | Chrome |
|---|---|---|
| `form` | `FormDialog` 直接读 `theme.action.<scope>.window/border/formInput/confirm/cancel` | FormDialog 自带 chrome |
| `page` | `processes.Render` / `history.RenderView` / 走 Detail page | `WrapActionWindow(content, scope, width)` 套边框 + 背景 |
| `async` | Wait status indicator 进 message rail | ⏳ Waiting on <short-id>… |

`WrapActionWindow(content, scope, width)` 是 Page chrome 的唯一 seam,
按 scope 选 `theme.action.<scope>.window` 的 bg + border:
- ModeTop → container scope chrome
- ModeHistory → image scope chrome

### 5. Operations registry (R06-02)

`config.LoadOperations()` 读 `defaults/operations/scopes/*.jsonc`,
校验 Mode↔body 对称、Requirement 闭集、Kind/Action 唯一性,索引到:

```go
type Operations struct {
    All      []OperationSpec
    ByKind   map[string]OperationSpec // "<scope>::<kind>" → spec
    ByAction map[string]OperationSpec // KeyAction string     → spec
    ByScope  map[OperationScope][]OperationSpec
}
```

`config.CachedLoadOperations()` 用 `sync.Once` 共享缓存,actionbar / keyboard 都读同一份。

### 6. File map (R06 已落地的所有文件)

```
internal/data/config/
├── theme.go                          # Theme = Palette+Chrome+Surfaces+Text+Feedback+Data+Action
├── operations.go                     # OperationSpec, OperationScope, OperationMode, Requirement
├── operations_loader.go              # LoadOperations(), validateSpec (mode↔body 对称)
├── operations_cache.go               # CachedLoadOperations() sync.Once
├── defaults/operations/scopes/
│   ├── container.jsonc               # 9 个 container Operation
│   └── image.jsonc                   # history + prune
├── theme_test.go                     # 新结构验证
├── light_theme_test.go               # 端到端覆盖
├── themes/*.jsonc                    # 6 个主题,按 chrome/surfaces/text/feedback/data/action 重写

internal/tui/ui/component/
├── styles_load.go                    # ApplyThemeStyles 按新 section 投影
├── action_chrome.go                  # WrapActionWindow(scope, content, width)
├── wait_indicator.go                 # NewWaitIndicator → ⏳ Waiting on …
id

internal/tui/actionbar/
└── registry.go                       # ActionItem 从 Operations 投影

internal/tui/keyboard/
├── operation_dispatch.go             # dispatchOperation → dispatchForm/Page/Async
├── operation_dispatch_test.go        # 6 个测试覆盖 Form/Page/Async + imagePrune
├── actions.go                        # 删 8 个 case,只留全局快捷键

internal/tui/state/
└── container_ops.go                  # ContainerWaitState: TargetID, IsActive()

internal/tui/ui/app/
├── page_templates.go                 # ModeTop/History 调 WrapActionWindow
├── rails.go                          # message rail 加 ContainerWait.IsActive 分支

internal/data/i18n/
└── lang_test.go                      # fmt.Sprintf(i18n.T(key)) MISSING bug 回归测试
```

---

## Follow-ups

| ID | Scope | Notes |
|---|---|---|
| **R06-07** | 状态合并 | `m.Form + m.Processes + m.ContainerWait + m.Detail` → 单一 `m.Operation` discriminated union。Rename 仍走 `ModeRename + Dialog`,并入后 dispatcher arm 一并收敛。 |
| **R06-08** | Port / Diff chrome | Port(`ModeDetail` + portBindings)和 Diff(`ModeDetail` + diffChanges)目前共用 `ModeDetail` 通用页面,没消费 `theme.action.container.window`。需要区分"container-scope Detail"和"普通 Detail"。 |
| **R06-09** | Volume / Network scope | `operations/scopes/volume.jsonc` + `operations/scopes/network.jsonc` + 主题加 `action.volume` / `action.network`。 |

---

## Light-theme acceptance check (R06 终验)

`tests/*.jsonc` 端到端覆盖(在 light 主题下):

| 测试 | 验证 |
|---|---|
| `TestLightActionContainerOverridesApply` | light.jsonc.action.container.{window.border=`#3875d7`, confirm.foreground=`#8a3fa0`, formInput.background=`#f0f0f0`} 三条 hex override 穿透 ThemePatch → ResolveColor |
| `TestLightActionImageScopeSeeded` | action.image 有可解析颜色,即使没 Operation 消费 |
| `TestLightThemeActionContainerStylesResolve` | ApplyThemeStyles 后 `rawStyles.ActionContainer*` 含 hex |
| `TestLightThemeActionImageScopeDefaultsApplied` | image scope 走 default 不 panic |
| `TestFormDialogHonoursLightThemeContainerWindowBorder` | 端到端:`FormContainerCopy` 的 FormDialog 输出含 `56;117;215`(`#3875d7`) |
| `TestFormDialogHonoursLightThemeConfirmForeground` | 端到端:`FormContainerExport` 输出含 `138;63;160`(`#8a3fa0`) |
| `TestFormDialogHonoursLightThemeFormInputBackground` | 端到端:`FormContainerUpdate` 输出含 `240;240;240`(`#f0f0f0`) |
| `TestWrapActionWindowRendersChrome` | ModeTop 渲染含 light-theme window.border SGR |
| `TestWrapActionWindowImageScopeUsesImageTheme` | image scope chrome 不会错拿 container.border |
| `TestWaitIndicatorRendersShortID` | 短 ID 显示,不显示全 hash |

切换到 light 主题后所有 chrome 都按 light palette 渲染;
切回 default / dark / dracula / nord / solarized 时 chrome 跟随各自调色板。

---

## Conventions (固化 lint 防止踩坑)

| Convention | 检查命令 |
|---|---|
| 调用 `i18n.T(key, args...)` 直接传 args,不要嵌套 `fmt.Sprintf(i18n.T(...), ...)` | `rg "fmt\.Sprintf\(i18n\." internal/` 必须返回 0 行 |
| 添加新 Operation: 1 行 JSONC + 1 行 dispatcher case,不动 `actions.go` | review 时人工确认 |
| ColorRef: bare token 还是 `#xxx` 字面值,看是否需要 palette 解析 | 自动校验 `IsToken()` |
| Mode ↔ body sub-object 对称:load-time 校验,违规立即 fail | `validateSpec` 在 `LoadOperations` 阶段报错 |