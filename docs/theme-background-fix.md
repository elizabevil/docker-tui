# 主题背景色修复方案 (Background Color Refactor)

> 建立日期: 2026-08-05
> 状态: 已落地,等待用户人工 TUI 验证视觉
> 涉及文件: `internal/tui/ui/component/box/*` (新), `internal/tui/ui/widget/header/header.go`, `internal/tui/ui/component/rows.go`, `internal/tui/ui/component/table.go`, `internal/tui/ui/app/layout.go`, `internal/data/config/themes/light.jsonc`

## 1. 问题描述

light 主题截图(用户提供 `/home/debi/Desktop/1.png`)显示:
- **正确**: Header 文字区 (CPU/Memory/Disk/TimeZone 等 lbl/val/pct) 浅色背景 + 深色文字
- **错误**: Header 右侧 (keystroke 提示框 + logo)、Panel 区域、Table 数据行 — **背景为深色**
- **正确**: Panel 主体容器白色
- **正确**: Footer 浅色

根因: 重构过程中把全局 `renderAppBackground` 移除,改用"组件级背景",但实际只有用 `box` 组件包过的少数 cell 有背景,其余用裸 `lipgloss.NewStyle().Width().Render()` 的区域(left padding、列间隙、keystroke 框、logo、panel 主体、table 数据行) 都没有背景,露出终端默认(深)。

## 2. 设计原则

**透明色是兜底,组件级背景是覆盖**

```
app 底层 (Palette.Background, 兜底)
  └─ renderAppBackground (全局 wrap, reset barrier, 三方向 padding 都填)
       └─ JoinVertical(header, message, panel, footer)
            └─ 组件 (Box 组件,默认 transparent → 透出全局)
                 └─ 例外: selected/alt/keystroke 等显式设 Background → 覆盖全局
```

- **默认**: 组件 `Background = ""` (transparent)
- **兜底**: `renderAppBackground` 把整 viewport 用 `Palette.Background` 填满
- **覆盖**: 显式设了 `Background` 的组件仍生效(Selected/Alt 浅灰高亮、Keystroke 边框色等)
- **关键**: `renderAppBackground` 必须用 reset barrier + 三方向 padding,否则兜底会因组件内 `\x1b[0m` 丢失

## 3. 实施方案 (5 Phase)

### Phase 1: `internal/tui/ui/component/box/` 新建组件包

**目的**: 统一"自包含盒"渲染原语,每个组件自带 `Background` 字段。

**文件**:
- `box.go` — 包核心,`styleWithBackground()` 工具
- `labeled.go` — `LabeledValue` (label-value 对,共享背景)
- `badge.go` — `Badge` + `BadgeRow` (键盘徽章,带 separator 背景继承)
- `stat.go` — `Stat` (CPU%/Memory% 阈值颜色 + 背景)
- `bordered_box.go` — `BorderedBox` (圆角边框盒 + 背景 + padding)
- `table_row.go` — `TableRow` (表格行,Width+Background 强制填满)
- `box_test.go` — 17 个单元测试

**API 关键点**:
- `Background string` — hex 颜色,`""` 表示透明
- 内部各部分共享同一背景,不依赖 reset barrier
- `styleWithBackground(base, bg)` — 集中处理"无背景"分支,避免每个组件重复

### Phase 2: 调用点改造

**`internal/tui/ui/widget/header/header.go`**:
- `lbl(s)` → `box.LabeledValue{Label: s, LabelStyle: box.StyleHeaderLabel, Background: ""}`
- `val(s)` → `box.LabeledValue{Value: s, ValueStyle: box.StyleHeaderValue, Background: ""}`
- `pct(v, c)` → `box.Stat{Value: v, Color: hexFromColor(c), Background: ""}`
- `renderKeyStrokeColumn` → `box.BadgeRow` + `box.BorderedBox` (都 Background: "")
- `resolveHeaderBackground()` → `return ""` (始终透明,靠全局兜底)

**`internal/tui/ui/component/rows.go`**:
- `RowRenderer` 加 `Background string` 字段
- `default` case 改用 `r.Background` (transparent → 全局兜底)
- `alt` case 叠加 `GetRowStyle("alt")` 背景

**`internal/tui/ui/component/table.go`**:
- `RowRenderer` 构造时设 `Background: resolveTableBackground()`
- `resolveTableBackground()` → `return ""` (始终透明)

### Phase 3: 主题背景 token 显式化

**`internal/data/config/themes/light.jsonc`** 关键修改:
```jsonc
"header": { "background": { "token": "background" } }   // 之前: transparent
"main": {
  "panelBackground":      { "token": "background" },  // 之前: transparent
  "actionBarBackground":   { "token": "background" },
  "messageRailBackground": { "token": "background" },
  "queryBarBackground":    { "token": "background" }
}
"footer": {
  "statusBackground":   { "token": "background" },
  "shortcutBackground": { "token": "background" }
}
"dialog": { "bodyBackground": { "value": "#e8e8e8" } }  // 已显式
"toast": { "background": { "token": "background" } }
```

**dark 主题 (dark/default/nord/dracula/solarized) 保持 `transparent`** — 终端默认即深色。

**Light 主题实际配色** (确认全浅):
- `background: #fefefe` — 近白
- `backgroundSubtle: #e8e8e8` — 浅灰
- `backgroundDeep: #f0f0f0` — 浅灰
- `foreground: #2b2b2b` — 近黑(深色前景保证可读性)
- `foregroundMuted: #9b9b9b` — 中灰
- 语义色 (success/primary/info/danger/warning/accent) — 较深,用于文字/图标

### Phase 4: 重新引入全局兜底 + 修复 padding 漏洞

**`internal/tui/ui/app/layout.go` 关键修改**:

`renderAppBackground` 必须解决三个 padding 方向:
| 方向 | 来源 | 宽度 |
|---|---|---|
| **左** | `padLeft(content, padH)` | `padH` 空格 |
| **右** | `(Viewport.W - usableW) / 2` | `padH ~ padH+1` 空格 ⚠️ |
| **上** | `marginTop` 行 `\n` | `marginTop` 行 |
| **下** | 短于 `Viewport.H` 的尾部 | 0 ~ `差值` 行 |

**bug 复盘**: `headerStyle.Width(usableW)` 只 pad 到 `usableW`,右 `padH` 列(从 `usableW` 到 `Viewport.W`)的空格**没有任何 ANSI 码**。即使 prepend `bgAnsi` 也不延伸到那里 → 终端默认色(深)透出。

**最终实现**:
```go
func renderAppBackground(m *state.AppModel, content string) string {
    if m == nil || m.Dependencies.Theme == nil || m.Viewport.Width <= 0 || m.Viewport.Height <= 0 {
        return content
    }
    bgHex := string(m.Dependencies.Theme.Palette.Background)
    if bgHex == "" || bgHex == "transparent" {
        return content
    }
    r, g, b := utils.HexToRGB(bgHex)
    bgAnsi := fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
    content = strings.ReplaceAll(content, "\033[0m", "\033[0m"+bgAnsi)
    // 逐行 prepend bgAnsi + 补到 Viewport.Width (含右 padding) + 补到 Viewport.Height
    lines := strings.Split(content, "\n")
    for i, line := range lines {
        pad := m.Viewport.Width - utils.DisplayWidth(line)
        if pad > 0 {
            line += strings.Repeat(" ", pad)
        }
        lines[i] = bgAnsi + line
    }
    for len(lines) < m.Viewport.Height {
        lines = append(lines, bgAnsi+strings.Repeat(" ", m.Viewport.Width))
    }
    return strings.Join(lines, "\n") + "\033[0m"
}
```

**`renderContentLayer`** (image/solid 场景): 移除 reset barrier,只保留每行背景应用。

**移除**: `style` 未使用的 import,`TestRenderAppBackgroundUsesThemePalette` 过时测试。

### Phase 5: 验证

```
go test ./internal/... ./cmd/... -count=1  → 全部 ok
go vet ./...                                → 无输出
go build                                     → 成功
```

## 4. 改动文件清单

**新增** (7):
- `internal/tui/ui/component/box/box.go`
- `internal/tui/ui/component/box/labeled.go`
- `internal/tui/ui/component/box/badge.go`
- `internal/tui/ui/component/box/stat.go`
- `internal/tui/ui/component/box/bordered_box.go`
- `internal/tui/ui/component/box/table_row.go`
- `internal/tui/ui/component/box/box_test.go`

**修改**:
- `internal/tui/ui/widget/header/header.go` — `lbl/val/pct/renderKeyStrokeColumn/resolveHeaderBackground` 改用 box 组件
- `internal/tui/ui/component/rows.go` — `RowRenderer` 加 `Background` 字段,`default`/`alt` case 用之
- `internal/tui/ui/component/table.go` — 加 `resolveTableBackground()` helper
- `internal/data/config/themes/light.jsonc` — 6 个 `*background` token 从 `transparent` 改 `token: "background"`
- `internal/tui/ui/app/layout.go` — 重新引入 `renderAppBackground`,实现三方向 padding 兜底 + reset barrier
- `internal/tui/ui/app/rails_test.go` — 移除过时测试

## 5. 用户验证清单

请用 light 主题重跑 `./dist/dtui`,确认:

- [ ] Header 整行 (左 + 右 + 列间 + keystroke 框 + logo) 统一浅色
- [ ] Panel 区域 (含 table 边框 + title) 浅色
- [ ] Table 数据行浅色,Selected/Alt 行有 `backgroundSubtle` 浅灰高亮
- [ ] Keystroke 提示框圆角边框 + 浅色背景
- [ ] Logo 区与左右 padding 都浅色
- [ ] Dark 主题下整体深色(终端默认),Selected 行仍带浅色高亮

## 6. 已知未处理

- 6 个 dark 主题的 `*background` token 保持 `transparent` — 终端默认即深色,正确
- `renderContentLayer` 的 per-row 背景应用(image/solid 场景)保留,但不再发 reset barrier
- box 组件的 `Background` 字段保留,作为"组件显式背景可覆盖全局兜底"的接口

## 7. 最终发现的非显然 bug:短形式 reset (`\x1b[m`)

实施上述 5 Phase 后,light 主题仍**有黑色子区**(header 右 logo / keystroke 框 / panel 右)。调试过程(见 `docs/debug-light-theme.md`):

- Level 1 终端测试: `xterm-256color` 支持 24/256/16 色背景 → 排除终端
- Level 4 主题测试: `Palette.Background = "#fefefe"` 正确加载 → 排除主题
- Level 2 逐行 debug: `visW=124, pad=1` 正确,`bgAnsi` prepend/append 都在 → 排除函数逻辑
- **Level 5 实际 result 字节打印定位根因**

**根因** (见 debug 输出 `right200` 第 0 行):
```
\x1b[38;2;43;43;43;48;2;254;254;254mno runtime        \x1b[m                                                   \x1b[38;2;6;125;205m
```

- `\x1b[38;2;43;43;43;48;2;254;254;254m` ← 前景 43,43,43 + 背景 254,254,254 ✅
- `no runtime` ← 内容
- `\x1b[m` ← **短形式 reset** (lipgloss 输出的简写,ANSI 等价于 `\x1b[0m`)
- `                                                   ` ← **空格 padding 背景被清掉了** ❌
- `\x1b[38;2;6;125;205m` ← 下一段(logo)前景

**reset barrier 的盲点**: `renderAppBackground` 的替换是 `strings.ReplaceAll(content, "\x1b[0m", "\x1b[0m"+bgAnsi)`。**只覆盖长形式**。lipgloss 输出的短形式 `\x1b[m` **不被替换**,清掉背景,后续 padding 空格露出终端默认(深色)。

**修法** (`internal/tui/ui/app/layout.go:renderAppBackground`):
```go
content = strings.ReplaceAll(content, "\033[0m", "\033[0m"+bgAnsi)
content = strings.ReplaceAll(content, "\033[m",  "\033[m"+bgAnsi)  // ← 关键
```

**复盘教训**:
- 在 lipgloss 输出面前,reset 类 SGR 至少有两种等价形式:`\x1b[0m` (完整) 和 `\x1b[m` (简写)
- 任何"在 X 后注入 Y" 的 hook 模式,都要枚举 X 的所有等价写法
- 更稳的写法:用 `regexp.MustCompile(`\x1b\[(?:\d+;)*0?m`)` 一次匹配所有 reset 形式

## 8. 后续建议

1. **commit 分三个**:
   - commit 1: "refactor(theme): 组件级背景 + box 包 + light 主题背景显式化"
   - commit 2: "fix(layout): renderAppBackground 三方向 padding 兜底 + 长形式 reset barrier"
   - commit 3: "fix(layout): renderAppBackground 补短形式 reset (\x1b[m) 兜底"
2. **如果仍有视觉偏差**: 截一张当前 light 主题截图,指出具体哪个子区仍是深色,我继续定位
3. **测试覆盖**:
   - `renderAppBackground` 的三方向 padding 兜底逻辑没有测试 — 建议补一个
   - 短/长 reset 形式都应被 patch — 建议补一个回归测试(`\x1b[m` 后的字符仍带背景)
