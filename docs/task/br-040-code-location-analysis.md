# BR-040 只读代码定位分析报告

> 建立日期: 2026-08-01
> 来源: [docs/task/next-small-model-brief.md](next-small-model-brief.md)
> 角色: 小模型只读定位助手
> 目标需求: BR-040 Dialog 风格统一 — 四周透明 + panel 居中
> 参考任务卡: [docs/task/br-040-dialog-style-center-on-panel.md](br-040-dialog-style-center-on-panel.md)

> 本报告是 BR-040 实现前的代码定位结果,**未修改任何代码**。
> 输出格式遵循 [docs/ai-prompts.md](../ai-prompts.md) "单任务代码定位 Prompt"。

---

## 1. 当前 dialog / overlay 渲染链路

### 1.1 渲染入口汇总(`internal/tui/ui/app/layout.go:218-256`)

```text
RenderApp
  ├─ 标准 / 紧凑分支判断后,统一构造 result = pad + padLeft(JoinVertical(...))
  ├─ 若 m.Navigation.Mode == ModeConfirm            → dialog.RenderChoiceOverlay(result, m)
  ├─ 若 m.Navigation.Mode == ModeExecShell          → component.RenderShellDialog + PlaceOverlay
  ├─ 若 ModeRename / ModeResourceCreate / ModeImageWorkflow
  │                                              → component.RenderTextInput + PlaceOverlay
  ├─ 若 ModeImageTransfer                        → component.RenderProgressDialog + PlaceOverlay
  ├─ 若 m.Dialog.Kind.IsSelection()               → dialog.RenderOverlay(result, m)
  ├─ 若 m.Dialog.Kind == DialogExec                → dialog.RenderExecOverlay(result, m)
  └─ 若 ModeRuntimeSelect                         → component.PlaceOverlay(termW, termH, renderRuntimeSelector)
```

### 1.2 各类 dialog / overlay 归属

| Mode / Kind | 文件 | 入口函数 | 中心化方式 | 备注 |
|---|---|---|---|---|
| `ModeConfirm`(confirm 多选 / 强制 / 删除 / Force) | `widget/dialog/view.go:243` | `RenderChoiceOverlay` | `dialog.PlaceDialog(content, box, termW, termH, ...)` | 多选项 dialog,选项由 `m.Confirm.Options` 决定 |
| `ModeExecShell`(BR-038 升级前的临时实现) | `ui/component/dialog.go:82` | `RenderShellDialog` | `component.PlaceOverlay(termW, termH, ...)` | 仅被 `ModeExecShell` 调 |
| `ModeRename`(容器重命名) | `ui/component/dialog.go:115` | `RenderTextInput` | `component.PlaceOverlay` | 单行 input + 光标 |
| `ModeResourceCreate`(卷/网络创建) | 同上 | `RenderTextInput` | 同上 | 同 Rename,内容是 create 模板 |
| `ModeImageWorkflow`(Tag/Push/Save/Load) | 同上 | `RenderTextInput` | 同上 | 镜像传输对话框 |
| `ModeImageTransfer`(传输进行中) | `ui/component/dialog.go:135` | `RenderProgressDialog` | `PlaceOverlay` | 进度条 + 当前字节数 |
| `m.Dialog.Kind.IsSelection()` = `DialogImageExport`/`DialogImageDebug` | `widget/dialog/view.go:228` | `RenderOverlay` | `dialog.PlaceDialog` | 保留背景内容,只覆盖 dialog 区 |
| `m.Dialog.Kind == DialogExec` | `widget/dialog/view.go:270` | `RenderExecOverlay` | `PlaceDialog` | shell 选项 + input + 按钮 |
| `ModeRuntimeSelect`(F2 runtime 选择器) | `ui/component/dialog.go:158` | `component.PlaceOverlay` | `lipgloss.Place(termW, termH, Center, Center, ...)` | 表格化 runtime 列表 |
| `ModeFilter` / `ModeSearch` / `ModeCommand` / `ModeImagePull` | `ui/app/rails.go:136` | `renderQueryInput` | 内嵌到 query rail,**不是 dialog** | 由 `filterMode` 函数判定,影响 panel border label |
| `ModeHelp` / `ModeLogView` / `ModeDetail` / `ModeTop` / `ModeExecPassthrough` | `ui/pages/` 各自页面 | 整页渲染,**不是 dialog** | n/a | 由 `state.Mode` 走 page router,不是浮层 |

### 1.3 PlaceDialog vs PlaceOverlay 两条工具链

| 函数 | 位置 | 语义 | 中心化算法 |
|---|---|---|---|
| `dialog.PlaceDialog(content, dialogBox, termW, termH, _, cfg)` | `widget/dialog/overlay.go:13` | 保留 content,**只覆盖** dialog 区域 | `dialogPosition(termW, termH, dlgW, dlgH, cfg)`(`overlay.go:144`):`x = termW*XOffset/100 - dlgW/2`, `y = termH*YOffset/100 - dlgH/2`,clamp 到 0 |
| `component.PlaceOverlay(termW, termH, dialog, overlayColor)` | `ui/component/dialog.go:158` | **替换** 整个 termW×termH 区 | `lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, dialog)` |

**共同缺陷**:两者都以 `(termW, termH)`(整个终端)为坐标系,**不是 panel**。dialog 居中在终端而非 panel body。BR-040 的核心修复点。

## 2. 当前居中依据

### 2.1 居中坐标系 = **整个终端**

- `dialog.PlaceDialog` 与 `component.PlaceOverlay` 全部按 `m.Viewport.Width × m.Viewport.Height` 居中。
- `LayoutReport`(`ui/app/mouse.go:53`)目前提供 `Panel {panelTop, bodyTop, bodyRows}`,**没有** `bodyLeft` / `bodyWidth`,所以即便想 panel 居中,目前也无 X 维度数据可用。

### 2.2 compact layout 特殊路径

`renderCompactApp`(`ui/app/compact.go:120`)完整重走 layout,**不**走 `RenderApp` 的 dialog 链。但:

- `layout.go:108-110` 标准分支前先判断 `TerminalCompact`,在 compact 分支里**直接** `return renderCompactApp(m)`,**绕开** 所有 dialog overlay(无 confirm / exec / 等 Mode 检查)。
- compact 模式无 `ModeConfirm` 等模式(键盘层在 compact 分支前就拦截了),所以 compact 不渲染 dialog,无居中风险。
- 仅 standard 模式需要 BR-040 修复。

### 2.3 当前"透明"实现现状

- `PlaceDialog`(`overlay.go:13-48`):**保留 content**,只把 dialog box 字符 splice 进对应行 — 这是 BR-040 要求的"透明 + 中间窗口"已**部分实现**。
- `PlaceOverlay`(`dialog.go:158-162`):**替换** 整个终端区(`lipgloss.Place`),**不保留** 背景内容 — 这部分**未实现**。
- `DialogBox`(`widget/dialog/box.go:21`):画 `Border(Padding(1, 2))` — 边框有 / 但 `Background` 不设,透明是**默认行为**(`Padding` 不画背景)。

所以 BR-040 的核心改动是**统一两条工具链**(PlaceDialog / PlaceOverlay)、**统一按 panel body 居中**、**统一不替换背景**。

## 3. CenterOnPanel 最小实现边界

### 3.1 必改文件

| 文件 | 改动 |
|---|---|
| `internal/tui/ui/widget/dialog/centered.go`(新) | 新增 `CenterOnPanel(content string, dialogBox string, panel panelRect, termW int, cfg dialogConfig) string` 核心函数 |
| `internal/tui/ui/widget/dialog/centered_test.go`(新) | 单元测试(几何 + 渲染快照 + 边界) |
| `internal/tui/ui/widget/dialog/overlay.go` | `PlaceDialog` 重构为薄壳,内部调用 `CenterOnPanel`;`dialogPosition` 改为接收 `panelRect` 而非 `termW/termH` |
| `internal/tui/ui/component/dialog.go:158` | `PlaceOverlay` 改为接受 `panelRect`,内部调用 `CenterOnPanel` 而非 `lipgloss.Place` |
| `internal/tui/ui/app/mouse.go:44` | `panelRect` 结构体加 `bodyLeft int` + `bodyWidth int` 字段;`resolveStandardLayout` / `resolveCompactLayout` 写入这俩字段;**不破坏** `bodyTop` / `bodyRows` 现有消费者 |
| `internal/tui/ui/app/layout.go:228-256` | 所有 `PlaceOverlay(...)` / `Render*Overlay(result, m)` 调用方改为传 `panelRect`(从 `ResolveLayout(m)` 取 `report.Panel`) |

### 3.2 可选文件

| 文件 | 改动 |
|---|---|
| `internal/tui/ui/widget/dialog/box.go` | `DialogBox` 不需要改(已不画背景),但 `Padding(1, 2)` 可以显式标注为 BR-040 约定(注释) |
| `internal/tui/ui/widget/dialog/view.go:212-275` | 各 `Render*` 函数签名**不需要改**,它们已经接收 `termW, termH, cfg`;只需调用方传入新参数 |
| `internal/tui/ui/component/dialog.go:50-152` | `RenderConfirmMsg` / `RenderShellDialog` / `RenderTextInput` / `RenderProgressDialog` 仍接收 `termW, termH`,内部调用 `PlaceOverlay(panelRect, ...)` 时**由调用方提供** |
| `internal/tui/ui/widget/dialog/dialog.jsonc` | 配置里 `XOffset / YOffset`(默认 50)在 BR-040 下**失去意义**,可标 `deprecated` 但暂不删(老用户配置可能依赖) |

### 3.3 不应触碰的文件

- `internal/tui/ui/app/compact.go` — compact 模式不渲染 dialog,BR-040 与之无关
- `internal/tui/state/dialog.go` — DialogKind / DialogState 由 BR-040 之外的需求管理
- `internal/tui/keyboard/confirm.go` / `resource_action.go` — confirm 入口逻辑不变,只换渲染链
- `internal/tui/keyboard/command.go` — 命令面板 `/`(`:` 命令)与 dialog 渲染无关
- runtime 层、Help / Footer 文案、`docs/navigation.md` — BR-040 不修改

### 3.4 panelRect 几何补充定义建议

```go
// mouse.go:44
type panelRect struct {
    panelTop  int  // existing
    bodyTop   int  // existing
    bodyRows  int  // existing
    bodyLeft  int  // NEW: X of first row inside panel content (panel left + border)
    bodyWidth int  // NEW: inner width excluding left+right border
}
```

`bodyLeft` / `bodyWidth` 计算:
- Standard:`bodyLeft = padH + 2`(`padH = (m.Viewport.Width - usableW)/2`,2 = left border),`bodyWidth = usableW - 4`
- Compact:`bodyLeft = 2`,`bodyWidth = usableW - 4`
- 与 `mouse.go:175` 的 `row := y - rep.Panel.bodyTop` 计算一致地扩展为 `col := x - rep.Panel.bodyLeft`

## 4. 测试建议

### 4.1 几何计算单测(`centered_test.go`)

```go
func TestCenterOnPanelPosition(t *testing.T) {
    cases := []struct{
        name string
        panel panelRect
        dlgW, dlgH, termW int
        wantX, wantY int
    }{
        {"centered normal", panelRect{bodyLeft: 40, bodyWidth: 100, bodyTop: 6, bodyRows: 20}, 50, 7, 140, 65, 12},
        {"narrow panel", panelRect{bodyLeft: 0, bodyWidth: 40, bodyTop: 4, bodyRows: 10}, 60, 8, 140, 20, 5},
        {"wide dialog in narrow panel", ...}, // overflow → 应截断
    }
    // 调 CenterOnPanel 后用 ansi.StringWidth 验证
}
```

### 4.2 overlay 保留底层内容测试(回归现有 overlay_test.go)

- 改 `PlaceDialog` 实现后,`TestPlaceDialogPreservesContentOutsideDialog`(overlay_test.go:10)需同步更新为新签名
- 新增:`TestCenterOnPanelPreservesContent` 验证透明背景 + dialog 居中

### 4.3 窄宽度 fallback 测试

- `bodyWidth < dlgW` 时:截断 / wrap / 缩小字号?**默认应截断 + 居中靠左对齐**(避免 `lipgloss.Place` 在窄区行为漂移)
- panel 宽度 = 40(最小),dialog 宽度 = 60:应截断到 38,左对齐于 `bodyLeft + 1`

### 4.4 compact layout 测试

- compact 模式下不渲染 dialog,无需新测试
- 但要测 `resolveCompactLayout` 返回的 `panelRect` 字段(`bodyLeft` / `bodyWidth`)正确

### 4.5 跨 dialog 一致性测试(快照)

- 对每类 dialog(confirm / selection / exec / text input / progress / runtime select)生成渲染快照
- 固定 viewport(例如 80×24)与 state 输入,断言所有 dialog 视觉风格(边框圆角 / padding / 标题颜色)一致

## 5. 风险点

| 风险 | 详细 |
|---|---|
| **dialog 覆盖 header / footer** | 当前 `PlaceDialog` 与 `PlaceOverlay` 都按 terminal 居中,若 panel 在 terminal 顶部 / 底部附近,dialog 会盖到 header / footer rail。**修复**:CenterOnPanel 严格按 `bodyTop..bodyTop+bodyRows` 限制 Y,超出范围直接 `clamp` |
| **双栏 Compose panel** | `PanelCompose`(`PanelCompose = 4`)在 Compose 详情模式下会切换到左右双栏(`ComposeFocus` 0 / 1)。`LayoutReport.Panel` 当前**单一 rect**,**不会**区分左右栏。CenterOnPanel 仍以单一 panel rect 居中,dialog 跨在左右栏中间(类似 vim 命令模式)。**需要主模型决策**:dialog 是覆盖两栏中间,还是按 ActiveColumn 居中。详见 §6 |
| **terminal 太小时 fallback** | `LayoutReport.Class == TerminalUnsupported` 时直接走 `renderTerminalError`,**不**进入 RenderApp 主链,dialog 不会显示。但 `bodyRows < dlgH` 时需要降级(dialog 高度 > 可用 panel 高度),目前 `PlaceDialog` 直接 `clamp y=0`,dialog 会**溢出 panel 顶部**(破坏 BR-040 约定)。**修复**:`CenterOnPanel` 在 `dlgH > bodyRows` 时按比例缩小 dialog 高度或滚动 dialog 内部(目前不支持滚动) |
| **ANSI 宽度计算** | `dialog.PlaceDialog`(`overlay.go:38-43`)用 `ansi.TruncateWc` / `ansi.TruncateLeftWc` 处理左侧/右侧的 ANSI 序列,**支持**宽字符与控制序列。新 `CenterOnPanel` 必须保留这个机制,不能改用 `lipgloss.Place` 全量替换(后者会破坏 ANSI) |
| **与 BR-039 Action Bar 复用** | BR-039 设计的 Action Bar 浮层会走 `CenterOnPanel`(`widget/actionbar/` 子包),新 API 必须**统一签名**,不能出现 `CenterOnPanel(content, dialogBox, ...)` 与 `CenterOnPanel(content, bar, ...)` 两个版本。**建议**:`CenterOnPanel` 只接受 `(content, dialogBox string, panel panelRect, termW int, cfg dialogConfig) string` 单个签名;Action Bar widget 自行渲染自身内容并 splice 进背景 |
| **运行时切 runtime 时 dialog** | `ModeRuntimeSelect` 期间,其它 dialog 不应该叠加;但当前模型里 dialog 是**串行** 的(`Mode` 字段),无嵌套 dialog 风险 |
| **Help 模式** | `ModeHelp` 走整页渲染,不是浮层;CenterOnPanel 与之无关,但需测试 `ModeHelp` 不会触发 `PlaceDialog` |
| **多次连续 dialog** | 例如 confirm → 取消 → 立刻 exec,先后两个 dialog 是否有残影?`PlaceDialog` 是**替换** 式合成,理论上不会,但需要回归测试 |

## 6. 需要主模型确认的问题

| 问题 | 选项 | 推荐 |
|---|---|---|
| **panel rect 来源**(双栏 Compose 居中策略) | (A) 单一 `panelRect`,dialog 跨双栏中间 (B) 按 `ComposeFocus` 切左右 rect,dialog 居中当前活跃栏 | (B) 与 BR-040 验收标准对齐("dialog 居中的是当前 `ActivePanel`") |
| **垂直居中范围** | (A) panel body 内严格居中 (B) 整个 panel(含 title 与 border)居中 (C) terminal 居中但限制 panel body 上下边界 | (A) — `bodyTop + bodyRows/2` 居中,符合 BR-040 §"位置:当前 panel 的居中" |
| **dialog 超出 panel body 时** | (A) 截断(可能丢内容) (B) wrap(可能断词) (C) 允许溢出(可能盖 footer / header) (D) 缩小到 panel body 大小 | (D) — 缩小 dialog 高度 + 内部内容压缩,避免溢出 |
| **多 dialog 叠加**(确认后立刻另一个 dialog) | (A) 当前串行 model 已天然避免 (B) 需要支持 modal stack | (A) — 当前 `m.Navigation.Mode` 是单一字段,无栈需求;若未来需要 modal stack 单独立 BR |
| **`panelRect.bodyLeft / bodyWidth` 计算口径** | (A) 从 `LayoutReport` 推导(集中) (B) 由每个调用方传入(分散) | (A) — `ResolveLayout` 唯一计算入口,BR-040 改造只改这里 |
| **`XOffset / YOffset` config 字段命运** | (A) 保留兼容(老用户配置) (B) 删除(破坏配置) | (A) — `dialog.jsonc` 标 deprecated,CenterOnPanel 内部忽略 `XOffset` / `YOffset`,改为从 panelRect 推导 |
| **Action Bar 浮层与现有 dialog 是否共用签名** | (A) `CenterOnPanel(content, dialogBox, panel, termW, cfg)` 单一签名 (B) 提供更通用的 `Overlay(content, contentBox, panel, termW, cfg)` + widget 自渲染 | (B) — CenterOnPanel 是底层原语,Action Bar widget 自行渲染(类似 DialogBox / ChoiceDialog 的关系) |
| **是否在 compact 模式下也启用 BR-040** | (A) 否(compact 不渲染 dialog,无意义) (B) 是(防御性) | (A) — compact 走 `renderCompactApp` 不进 dialog 链 |

## 7. 不确定项(需要主模型重新读代码确认)

- `internal/tui/ui/widget/dialog/overlay.go:144-162` 的 `dialogPosition` 是否被**任何**单元测试覆盖?(`grep -l 'dialogPosition' internal/tui/ui/widget/dialog/` —— 若无覆盖,改它风险低)
- `component.PlaceOverlay`(`component/dialog.go:158`)被**多少处**调用?(`grep -rn 'component\.PlaceOverlay' internal/` —— 需列全调用方,改签名时同步更新)
- `LayoutReport.Panel` 的 `panelTop / bodyTop / bodyRows` 三字段**全部**被消费?(`grep -rn 'rep\.Panel\.' internal/tui/` —— 增 `bodyLeft / bodyWidth` 后确保已有消费者仍工作)
- `ModeCommand`(命令面板 `/`)走 `queryKindCommand`,**不是 dialog**,但 Help 文案会把它与 BR-039 Action Bar 一起描述;BR-040 不动它
- `ModeImageTransfer`(传输进度)涉及后台 goroutine + `tea.Tick` 状态,`PlaceOverlay` 改造时不能阻塞 progress 更新路径(若是同步渲染需要测试 race condition)
- `ModeExecShell` 的临时实现与 BR-038 的全屏 exec 升级如何协调 — CenterOnPanel 改造不应影响 BR-038 的全屏方向

## 8. 输出文件建议

按 BR-040 验收标准:`internal/tui/ui/widget/dialog/centered.go` 与 `centered_test.go` 必须新增;`overlay.go` / `mouse.go` / `dialog.go` (component) / `layout.go` 必须修改;**不修改** runtime / state / keyboard / docs。

---

## 9. 给主模型的总结

- **现状**:两条工具链(`dialog.PlaceDialog` 与 `component.PlaceOverlay`)都按 `termW × termH` 居中,**不**按 panel。`LayoutReport.Panel` 缺 `bodyLeft` / `bodyWidth`,需要扩展。
- **修复**:新增 `CenterOnPanel` 单函数,在 `panelRect` 提供的 body 矩形内居中;`PlaceDialog` 与 `PlaceOverlay` 都改为薄壳调用 `CenterOnPanel`;`renderCompactApp` 不受影响。
- **风险**:ANSI 宽度处理(必须保留 `ansi.TruncateWc` 机制)、双栏 Compose 居中策略、超出 panel body 时的降级、与 BR-039 Action Bar 的统一签名。
- **不修改**:runtime / state / keyboard / docs;`ModeCommand` / `ModeHelp` / `ModeLogView` 等非 dialog Mode 不受影响。
- **需要确认**:panel rect 扩展范围(双栏 Compose)、XOffset/YOffset 配置 fate、Action Bar 共用签名设计。

报告完成。**未修改任何代码**。