# BR-040 实施决策记录

> 建立日期: 2026-08-01
> 来源: 小模型只读分析报告 [docs/task/br-040-code-location-analysis.md](br-040-code-location-analysis.md)
> 状态: BR-040 已完成

## 决策列表

### 1. 采纳:先做 BR-040

BR-040 仍然是下一步最合适的任务。它能给后续 BR-039 Action Bar 提供统一浮层基础。

### 2. 采纳:统一 PlaceDialog 和旧 PlaceOverlay

当前确实有两条链路:一条保留背景,一条整屏替换。BR-040 应该统一为"保留背景,只覆盖 dialog 区域"。

### 3. 收窄:先不做双栏 Compose 当前栏居中

小模型建议按 ComposeFocus 切左右栏,这会扩大布局模型。第一版建议只按主 panel body 居中,Compose 双栏后续单独优化。

### 4. 收窄:不改 component.PlaceOverlay 为复杂公共 API

更稳的做法是新增 `dialog.CenterOnPanel` / `dialog.PlaceDialogInPanel`,让 `ui/app/layout.go` 的 dialog 分支改走 `widget/dialog` 层。旧 `component.PlaceOverlay` 可暂时保留兼容,减少一次性改动。

### 5. 收窄:超出 panel 时先 clamp + 截断,不做内部滚动

"缩小 dialog 高度 + 内容压缩"会牵涉所有 dialog 内容生成。第一版只保证不覆盖 header/footer,内容过大时截断或限制在 panel 区域。

## 实施边界(主模型确认)

### 必改

- 新增 `internal/tui/ui/widget/dialog/centered.go`
- 新增 `internal/tui/ui/widget/dialog/centered_test.go`
- 调整 `internal/tui/ui/widget/dialog/overlay.go`
- 扩展 `internal/tui/ui/app/mouse.go` 的 panel rect,补 `bodyLeft` / `bodyWidth`
- 调整 `internal/tui/ui/app/layout.go` 的 dialog overlay 调用

### 保持不动

- 保留 compact 路径不动(`renderCompactApp` 不渲染 dialog,不受影响)
- 不改 runtime / state / keyboard

## 协作分工

- **生产代码**: 由主模型实现
- **测试补齐 / diff 巡检**: 由小模型接力

## 第一步:BR-040-A

实现 `CenterOnPanel` 核心函数 + 单元测试。

## BR-040-A 完成

- 新增 `internal/tui/ui/widget/dialog/centered.go`:`PanelBody` 结构体、`CenterOnPanel` 核心函数、`PlaceDialogInPanel` 高层包装
- 新增 `internal/tui/ui/widget/dialog/centered_test.go`:7 个 table-driven 测试(普通居中 / 空 dialog / 宽 dialog 截断 / 高 dialog 截断 / ANSI 保留 / 高层包装自动推导 / 边界 clamp)
- 扩展 `internal/tui/ui/app/mouse.go` 的 `panelRect`:`bodyLeft` / `bodyWidth` 字段;`resolveCompactLayout` 写入 `bodyLeft=2, bodyWidth=ViewPort.Width-4`;`resolveStandardLayout` 复用 `contentWidthPct` 计算写入 `bodyLeft=padH+2, bodyWidth=usableW-4`
- 全部测试通过(7/7),无回归

## BR-040-B 完成

- `internal/tui/ui/widget/dialog/view.go`:新增 3 个 `*InPanel` 变体
  - `RenderChoiceOverlayInPanel(content, m, body)`
  - `RenderOverlayInPanel(content, m, body)`(image export / debug 类型)
  - `RenderExecOverlayInPanel(content, m, body)`
  - 每个变体以 `body.Width` / `body.Rows` 作为 dialog 尺寸,再通过 `PlaceDialogInPanel` 居中
- `internal/tui/ui/app/layout.go`:在 dialog 分支前用 `ResolveLayout` 算 `PanelBody`;dialog 分支(ModeConfirm / ModeExecShell / ModeRename / ModeResourceCreate / ModeImageWorkflow / ModeImageTransfer / IsSelection / DialogExec / ModeRuntimeSelect)全部切到 panel body 居中
- `internal/tui/ui/app/mouse_test.go`:新增 `TestPanelBodyGeometryStandard` / `TestPanelBodyGeometryCompact`,覆盖 `bodyLeft` / `bodyWidth` 字段
- 全部 ui 包测试通过(无回归)
- `widget/dialog/overlay.go` 的 `PlaceDialog` 暂未重构(主模型决策:保留兼容)

## BR-040-C 完成

- `internal/tui/ui/component/dialog.go`:拆出 `RenderShellDialogBox` / `RenderTextInputBox` / `RenderProgressDialogBox`,live app 渲染路径可只生成 dialog box,再由 `dialog.CenterOnPanelDefault` 拼回当前 panel body
- 旧 `RenderShellDialog` / `RenderTextInput` / `RenderProgressDialog` / `PlaceOverlay` 保留为兼容入口,不作为 `RenderApp` 当前 dialog 分支的主路径
- `internal/tui/ui/app/rails_test.go`:新增回归测试,覆盖 confirm / rename / image transfer / runtime select overlay 不丢失页面 chrome,并保持固定 terminal 行数
- `go test -short ./...` 通过

## BR-040 完成度

- [x] 新增 `centered.go` / `centered_test.go`
- [x] 调整 component dialog box 生成链路,旧 `PlaceOverlay` 保留兼容
- [x] 扩展 `mouse.go` 的 `panelRect`
- [x] 调整 `layout.go` 的所有当前 live dialog overlay 调用
- [x] 保留 compact 路径不动
- [x] 不改 runtime / state / keyboard

### 后续可优化(非 BR-040 必需)

- `dialog.jsonc` 的 `XOffset` / `YOffset` 标 deprecated(`CenterOnPanel` 不再用)
- 重构 `widget/dialog.PlaceDialog` 为 `CenterOnPanel` 的薄壳(全终端 body)
- 按 `body.Width` 重新调整 `dialogConfig.WidthPercent` / `MinWidth` / `MaxWidth` 默认值(目前直接传 body.Width 给 `ChoiceDialog` 内部使用)
