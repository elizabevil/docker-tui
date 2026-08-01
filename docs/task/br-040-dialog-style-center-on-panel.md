# Dialog 风格统一:四周透明 + panel 居中

> 任务卡 / BR-040

## 元信息

- **关联编号**:BR-040
- **优先级**:medium
- **状态**:`done`
- **依赖**:BR-039(Action Bar 也是 dialog,优先统一)
- **关联设计**:[docs/feature-design.md §5.8](../feature-design.md)

## 目标

统一所有 dialog(confirm / shell / filter / exec / notification / selection / 自定义表单)的渲染规范:

- **四周透明**(不是全屏覆盖)
- **中间窗口**(dialog 内容)
- **位置:当前 panel 的居中**(作用域是 panel)
- 统一边框 / 内边距 / 标题位置

## 代码结构索引

### 必须读懂的文件

| 文件 | 作用 |
|---|---|
| `internal/tui/ui/widget/dialog/overlay.go` | `PlaceOverlay` 当前工具函数 |
| `internal/tui/ui/widget/dialog/view.go` | 各 dialog 渲染分发 |
| `internal/tui/ui/widget/dialog/confirm.go` | confirm 对话框 |
| `internal/tui/ui/widget/dialog/notification.go` | 通知 dialog |
| `internal/tui/ui/widget/dialog/selection.go` | selection 多选项 dialog |
| `internal/tui/ui/widget/dialog/exec.go` | exec dialog(将由 BR-038 升级) |
| `internal/tui/ui/widget/dialog/choice.go` | choice dialog(下拉选择) |
| `internal/tui/ui/app/layout.go:104` | `resolveStandardLayout` 面板几何(`LayoutReport.Panel`) |
| `internal/tui/ui/app/mouse.go` | `LayoutReport.Panel` bodyTop / bodyRows 计算 |
| `internal/tui/ui/widget/dialog/view.go` | 各 dialog 入口 |

### 必须修改的文件

| 文件 | 改动 |
|---|---|
| `internal/tui/ui/widget/dialog/centered.go`(新) | 新增 `CenterOnPanel(panelRect Rect, content string, w, h int) string` |
| `internal/tui/ui/widget/dialog/overlay.go` | `PlaceOverlay` 改为调用 `CenterOnPanel` 或保留作兼容入口 |
| `internal/tui/ui/widget/dialog/view.go` | 所有 dialog 渲染改走 `CenterOnPanel` |
| `internal/tui/ui/widget/dialog/confirm.go` | confirm 渲染走统一规范(透明边框 / 居中) |
| `internal/tui/ui/widget/dialog/notification.go` | notification 同上 |
| `internal/tui/ui/widget/dialog/selection.go` | selection 同上 |
| `internal/tui/ui/widget/dialog/exec.go` | exec 同上(BR-038 协同) |
| `internal/tui/ui/widget/dialog/choice.go` | choice 同上 |

### 必须新增的文件

- `internal/tui/ui/widget/dialog/centered.go`:`CenterOnPanel` 实现 + 测试
- `internal/tui/ui/widget/dialog/centered_test.go`:单元测试 + 渲染快照

### 不应触碰的文件

- 渲染风格之外的功能(键盘 / 状态)
- 全局布局
- dialog state(由 widget/actionbar 等已有结构管理)

## 操作链路

```text
Dialog renderer(各组件)
  → 统一调用 ui/widget/dialog/centered.go:CenterOnPanel
      panelRect := LayoutReport.Panel (来自 app.LayoutReport)
      w, h := 计算 dialog 内容尺寸(lipgloss.Width / Height)
      x := panelRect.bodyLeft + (panelRect.bodyWidth - w) / 2
      y := panelRect.bodyTop + (panelRect.bodyRows - h) / 2
      渲染:panel 内除 dialog 区域外透明(不画背景填充)
  → ui/app/layout.go:在 panel 渲染前调用 CenterOnPanel
      或每个 dialog 在自己的 Render 函数中传入 panelRect
```

## 验收标准

- [x] 所有当前 live dialog(confirm / selection / exec / shell / input / progress / runtime select)在 panel body 内居中显示。
- [x] 四周透明:dialog 边框外不画背景填充,只画边框 + 居中窗口内容。
- [x] panel 缩窄 / 拉宽时,dialog 跟着居中且不溢出 panel。
- [x] 同类 dialog(例如两个 confirm)风格一致:边框 / 内边距 / 标题位置相同。
- [x] dialog 内的按钮 / 输入框 / 文本行版式统一。
- [x] dialog 不覆盖 header / footer / message rail。
- [x] 多 panel 布局下 dialog 居中的是当前 `ActivePanel` 的 panel body。
- [x] 单元测试覆盖:不同 panel 宽度 / 高度下,CenterOnPanel 计算正确。

## 风险

| 风险 | 说明 |
|---|---|
| **panel 几何失效** | BR-012 wontfix 后,`LayoutReport.Panel` 是单一 rect;多 panel 布局下需要按 `ActivePanel` 重新计算 |
| **dialog 内容溢出** | 长文本 / 多行输入框在窄 panel 中可能溢出;需要滚动或自动 wrap |
| **影响既有功能** | 改 dialog 渲染后,所有现有 dialog 的视觉与交互不能回归 |
| **测试覆盖** | 现有 dialog 测试可能依赖具体位置 / 尺寸;需要更新测试 |
| **BR-039 协同** | Action Bar 也是 dialog,必须用 `CenterOnPanel` |

## 建议任务分解

1. **TASK-BR040-A:`CenterOnPanel` 实现 + 单测**
   - `widget/dialog/centered.go` + `centered_test.go`
   - 几何计算 + 渲染逻辑
   - **预估**:小模型独立可完成
2. **TASK-BR040-B:逐 dialog 改造**
   - confirm / notification / selection / exec / filter / choice 等全部走 `CenterOnPanel`
   - **预估**:小模型机械迁移
3. **TASK-BR040-C:BR-039 Action Bar 集成 `CenterOnPanel`**
   - Action Bar widget 使用统一接口
   - **预估**:等 BR-039
4. **TASK-BR040-D:视觉回归测试 + 截图对比**
   - 在 test/preview/ 添加 dialog 截图,与旧版对比
   - **预估**:小模型独立

## 待确认项

- [x] **panel rect 来源**:使用 `ResolveLayout` 统一计算 `LayoutReport.Panel` 的 `bodyLeft` / `bodyWidth` / `bodyTop` / `bodyRows`
- [x] **dialog 居中算法**:水平与垂直都在 panel body 内居中,不包含标题栏 / 边框 / 全局 rails
- [x] **窄 panel fallback**:第一版 clamp + ANSI-aware 截断,不做内部滚动
- [x] **多 dialog 叠加**:沿用现有单一 dialog state,不新增堆叠模型
- [x] **滚动 / 拖动**:当前不支持,超出内容按 panel body 截断
