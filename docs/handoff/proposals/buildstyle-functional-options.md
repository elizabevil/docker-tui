# buildStyle 函数式选项重构

> 日期: 2026-08-05
> 状态: 计划阶段，待用户确认后实施

## 1. 目标

将 `buildStyle(ref styleRef)` 转换为函数式选项签名 `buildStyle(ref styleRef, opts ...StyleOption) lipgloss.Style`，引入 `StyleOption` + `fromStyle` + `withWidth`，使用新 API 消除三个调用点的重复代码。

**关键设计**：
- 函数式选项模式允许可选参数扩展
- 保留现有的 α=0 透明保护逻辑
- `WrapStyle` 变成 `buildStyle(ref, fromStyle(s))` 的薄封装

## 2. 改动范围

### 2.1 必须完成

| 文件 | 改动 |
|---|---|
| `internal/tui/ui/component/styles_load.go` | 转换 `buildStyle` 签名为函数式选项，添加 `StyleOption` 类型、`fromStyle` 和 `withWidth` 辅助函数。保留 α=0 透明保护 |
| `internal/tui/ui/component/helpers.go` | `WrapStyle(s, ref)` 变成 `buildStyle(ref, fromStyle(s))` |
| `internal/tui/ui/component/rows.go` | 三处调用点更新 |

### 2.2 详细变更

#### 2.2.1 `buildStyle` 签名变更

```go
// 旧
func buildStyle(ref styleRef) lipgloss.Style

// 新
func buildStyle(ref styleRef, opts ...StyleOption) lipgloss.Style
```

#### 2.2.2 新增类型

```go
type StyleOption func(s lipgloss.Style) lipgloss.Style

func fromStyle(s lipgloss.Style) StyleOption

func withWidth(w int) StyleOption
```

#### 2.2.3 rows.go 调用点

| 位置 | 旧代码 | 新代码 |
|---|---|---|
| `joinRow` (lines 100-114) | 内联 `lipgloss.NewStyle().Foreground(c).Bold(b).Faint(f).Render(display)` | `buildStyle(sty).Render(display)` |
| `RenderRow` selected/marked (line 44) | `buildStyle(ref).Width(r.rowWidth)` | `buildStyle(ref, withWidth(r.rowWidth))` |
| `RenderRow` alt-row merge (lines 56-72) | 手动合并 alt style | `buildStyle(GetRowStyle("alt"), fromStyle(style))` |

### 2.3 不得修改

- 不修改 `cellStyleANSI` — 输出 ANSI SGR 字符串，与 lipgloss.Style 不同
- 不修改 `widget/dialog/*.go` 和 `dialog.go` 中的 lipgloss 链 — 单独的重构范围
- 不删除 `WrapStyle` — 保持公共 API 稳定
- 不改变 α=0 透明保护语义
- 不改变字段应用顺序（Color → Background → Bold → Faint）

## 3. 验收标准

- [ ] `go build ./...` 通过
- [ ] `go vet ./...` 通过
- [ ] `go test ./internal/tui/ui/component/...` 通过
- [ ] 现有 `buildStyle` 调用者继续工作（无参数变化）

## 4. 待确认问题

无
