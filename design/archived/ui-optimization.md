# dtui UI 优化方案

## 现状分析

已完成：

- ✅ 页面统一边框尺寸（所有模式使用 ActiveBorderStyle.Width().Height()）
- ✅ 列配置 JSON 化（config.json 中定义 wide/more/pct/compact 四级配置）
- ✅ 灵活列布局（百分比宽度 + fill 列 + -1 隐藏列）
- ✅ 镜像卡片展开态查看容器列表
- ✅ keyboard 独立包
- ✅ 组件结构优化（FlexTable 方法化、StateDot 辅助函数）

待优化：

### 1. Footer 快捷键栏

当前问题：`footer/footer.go` 中快捷键描述是硬编码的英文字符串，没有使用 i18n。

```go
{string(key.KeyJ) + key.KDown, "Down"}
```

优化方案：使用 i18n.T("key.down") 等已有 i18n key。

### 2. Header 信息栏

当前问题：`header/header.go` 中连接信息、引擎信息显示可以更丰富。
优化方案：

- 显示容器总数（running/total）
- 显示当前引擎类型 + 版本
- 更紧凑的布局

### 3. 空状态提示

当前问题：不同页面空状态提示不一致。

- 网络："Loading..."（纯字符串）
- 容器：""（空的 FlexRender）
- 镜像：""（空的 FlexRender）
- 卷：""（空的 FlexRender）
  优化方案：统一使用 `(empty)` 风格 + i18n 支持。

### 4. Dialog 弹窗

当前问题：`dialog/view.go` 中 export/debug/exec 弹窗样式可以更统一。
优化方案：复用 ActiveBorderStyle + 固定宽度。

### 5. Confirm 确认框

当前问题：`layout.go` 中 `renderConfirmDialog` 是独立渲染的，样式与主面板不一致。
优化方案：与 Dialog 弹窗风格统一。

### 6. Help 帮助页

当前问题：`help/view.go` 中快捷键列表硬编码，没有利用 `key.HelpEntries()`。
优化方案：改用 `key.HelpEntries()` 生成快捷键列表。

### 7. 行选中样式

当前问题：`FlexLine` 中选中行使用 `rowSelected` 样式（blue + bold），但其他行的 `rowPrefix` 是硬编码空格。
优化方案：统一使用配置化的行前缀样式。

## 执行优先级

P0: Footer i18n — 最简单的改进，影响面广
P0: 空状态统一 — 所有页面一致
P1: Confirm 弹窗样式统一 — 与 Dialog 一致
P1: Help 改用 HelpEntries — 消除硬编码
P2: Dialog 样式增强
P2: Header 信息增强
P3: 行选中样式优化
