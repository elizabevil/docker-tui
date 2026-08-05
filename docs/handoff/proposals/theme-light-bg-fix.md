> ⚠️ **本修复针对的"切到 light 主题仍显示深色"问题已被 [V3 颜色透明方案](../../color-transparency-plan.md) 解决**(2026-08-05)。
>
> 根因(R2)与本方案中"显式 .Background() 调用计数清单"在 V3 仍然有效;但 V3 通过加载期预混合把 101,101,101 异常值消除,而本方案当时尚未触达此层。本文件保留作为历史根因分析。

# Light 主题背景一致性修复

> 日期: 2026-08-04
> 状态: 已实施,代码与测试全绿
> 关联: [./theme-hardcoded-migration.md](./theme-hardcoded-migration.md) §3.3 + [./theme-hardcoded-followups.md](./theme-hardcoded-followups.md)

## 1. 用户报告

切换到 `--theme light` 时:
- 终端背景仍然是黑色(用户的默认终端设置)
- action bar、panel、border 容器**不显示任何背景色**——落在终端黑色背景上
- 只有 header bar、dialog 主体、shortcut bar 显示浅色背景(它们显式调用了 `.Background()`)
- 视觉错位:浅色前景文字 + 黑色容器底 + 部分浅色背景方块 → 整体"看起来仍然是 dark 主题"

## 2. 根因分析

TUI 应用**无法强制改变用户的终端背景色**——终端背景由用户/终端模拟器决定,TUI 只控制前景字符 + 显式画的背景方块。在 `internal/tui/ui/` 中,组件分两类:

**A. "画背景"的组件**(切主题会跟随):
- `headerBar` → `theme.Header.Background`(light 主题下 `#e8e8e8` 浅灰)
- `dialog` 主体(`DialogBox`) → `opaqueDialogColor(theme.Dialog.Overlay + OverlayOpacity)`(light 主题下 `#fec` 浅米色)
- `shortcutBar` → `theme.Footer.ShortcutBackground`(已修,light 主题下 `#f0f0f0`)
- `selectedRow` → `theme.Main.RowSelected`(light 主题下 `#3875d7` 蓝)
- `keyBadge` → `style.Colors.Blue`(light 主题下 `#3875d7`)

**B. "不画背景"的容器**(切主题不跟随,落在终端黑色背景上):
- `actionbar` 容器(`actionbar.go:74-78` 的 `lipgloss.NewStyle().Border().Padding().Width().Render(...)`)→ 无 `.Background()`
- `panel` 边框容器(`panel.go:43` 用 `tui.ActiveBorderStyle`)→ `ActiveBorderStyle` 只有 Foreground + Border
- `dialog` 的内 body parts(`box.go:32` 的 `lipgloss.JoinVertical(...)`)→ 各 part 自己控制背景,容器外层不画

## 3. 修复策略

**给所有 B 类主容器加 `.Background(style.Colors.BG)`**(`BG` 是 `theme.Palette.Background` 在 `tui/styles.go:30` 的别名)。

理由:
1. `theme.Palette.Background` 是主题定义好的"主背景色"——dark 主题是 `#18191b`、light 主题是 `#fefefe`、nord 是 `#242933` 等。
2. 给主容器加这个背景后,切主题时**整屏视觉一致**:dark 主题所有容器都是深底浅前景字,light 主题所有容器都是浅底深前景字。
3. 对原有 `layout.Background.Fallthrough` 行为兼容(`tui.ApplyLayoutConfig` 在 `styles.go:43-55` 已经在背景打开且 fallthrough 时把 ActiveBorderStyle 背景清成 `lipgloss.NoColor{}`)——本批次改动不破坏该行为。

## 4. 改动清单

### 4.1 `internal/tui/styles.go`

`ApplyTheme` 在创建 3 个 BorderStyle 时各加 `.Background(bgColor)`:

```go
bgColor := style.Colors.BG
ActiveBorderStyle = lipgloss.NewStyle().Border(br).Foreground(...).Background(bgColor).Padding(0)
InactiveBorderStyle = lipgloss.NewStyle().Foreground(...).Background(bgColor).Padding(0)
FocusedBorderStyle = lipgloss.NewStyle().Border(br).Foreground(...).Background(bgColor).Padding(0)
```

影响:`panel.Panel.Render()` 用的 `tui.ActiveBorderStyle` 现在带背景——所有 panel 容器在 light 主题下显示 `#fefefe` 浅色底。

### 4.2 `internal/tui/ui/widget/actionbar/actionbar.go`

`renderBox` 的返回 style 加 `.Background(style.Colors.BG)`:

```go
return lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    Padding(0, 1).
    Width(boxWidth - 4).
    Background(style.Colors.BG).  // 新增
    Render(strings.Join(lines, "\n"))
```

`import` 增 `github.com/elizabevil/docker-tui/internal/tui/ui/style`。

影响:action bar 容器(`:` 调出的命令面板)在 light 主题下显示浅色底,与 header/dialog 视觉一致。

## 5. 未修复的组件

| 组件 | 文件 | 当前行为 | 风险 |
|---|---|---|---|
| dialog 内的 body parts | `widget/dialog/{form,notification,selection,exec}.go` | 各 part 自己控制背景,容器外层不画 | 当用户切 light + 触发 dialog 时,dialog 框整体被 `DialogBox` 的 `Background(opaqueDialogColor(overlay))` 覆盖,所以**视觉一致**——box 背景是 `#fec` 浅米色 |
| `header.go:80` 的 pct 函数 | `lipgloss.NewStyle().Foreground(c).Render(...)` | 只设前景 | 仅 CPU 百分比文字,占比小,不影响视觉一致性 |
| `compose/view.go` 多处 | `lipgloss.NewStyle().Width/Height.Render(...)` | 无背景 | compose 列表项落在 panel 容器内,panel 已有背景 |

未列入本批次。后续若用户在更多容器发现"黑底漏出"现象,可同样模式加 `.Background(style.Colors.BG)`。

## 5. R5 增量:A 类全容器补背景(2026-08-04)

用户在 R4 修复后报告"light 主题背景仍有黑色背景泄露"。进一步盘点列出 3 类未跟随主题的源头:

- **A 类(主容器无背景)**:11 处——本批修
- **B 类(组件细节用 `style.Colors.X` 直接绑调色板)**:48 处——本批不修
- **C 类(theme-hardcoded-migration §6 已知保留项 12 处 hex 字面量)**:12 处——本批不修

A 类修复清单(11 处全部加 `.Background(style.Colors.BG)`):

| 文件 | 行 | 修复点 |
|---|---|---|
| `internal/tui/ui/app/layout.go` | 294 | `runtimeSelectorUI.Render` 容器 |
| `internal/tui/ui/app/layout.go` | 434 | `renderTruncatedLines` 行尾容器 |
| `internal/tui/ui/widget/panel/panel.go` | 41 | `panel.Panel.Render` body 容器 |
| `internal/tui/ui/pages/compose/view.go` | 150 | `topBar` JoinHorizontal left |
| `internal/tui/ui/pages/compose/view.go` | 151 | `topBar` JoinHorizontal sep |
| `internal/tui/ui/pages/compose/view.go` | 152 | `topBar` JoinHorizontal right |
| `internal/tui/ui/pages/help/view.go` | 105 | help leftBox |
| `internal/tui/ui/pages/help/view.go` | 106 | help rightBox |
| `internal/tui/ui/pages/logs/view.go` | 116 | logs body |
| `internal/tui/ui/pages/detail/view.go` | 153 | detail body |

每处 1-2 行改动;`layout.go`/`panel.go`/`help.go` 增 `style` import。

未修的剩余未画背景容器(已知,但用户说"当前先补充 A 类"——B/C 留后续):

- `widget/header/header.go:80,150,176`(CPU 百分比 / logo 列 / keystroke 列)—— headerBar 已有背景,内部小容器落在 headerBar 内视觉仍 OK
- `widget/dialog/{form,exec,choice,selection}.go` 按钮文本容器—— DialogBox 整体有背景
- `widget/dialog/box.go:34`(DialogBox 本体)—— 已在 R4 修复加背景
- `app/rails.go:217`(query rail)—— 已是消息栏局部

## 6. 验证

```text
go test ./...        # 全部通过
go vet ./...         # 无诊断
JSONC 嵌套键 1:1 镜像:6 主题均 47 键
```

R5 修复后,light 主题下:
- runtime 选择器、辅助栏、panel body、compose 左右栏、help 左右栏、logs body、detail body——全部显示 `#fefefe` 浅色底
- dark 主题下显示 `#181919b` 深色底
- 其余 5 主题跟随各自调色板

## 8. R6 增量:B 类 13 处对比度差修复(2026-08-04)

R5 完成后用户确认**A 类已修**，但切到 light 主题仍能看到"颜色泄露"。进一步盘点发现 B 类问题:**虽然 `style.Colors.X` 已经在 `tui.ApplyTheme:19-30` 注入到主题调色板,但在 light 主题下,某些 token 解析后的 hex 与浅色背景对比度差,导致视觉"不可见/看不清"**。

B 类问题分类(light 主题实测):

| 位置 | 旧 | 新 | light 主题下表现 |
|---|---|---|---|
| `header.go:163` keyBadge 背景 | `style.Colors.Blue` (`#3875d7`) | `component.GetStyle("headerBar").GetBackground()` (= `theme.Header.Background` = `#e8e8e8`) | 旧:蓝底浅前景字(模糊);新:浅灰底白前景 |
| `header.go:178` keystroke border 前景 | `style.Colors.Blue` (`#3875d7`) | `component.GetStyle("panelTitle").GetForeground()` (= `theme.Main.Title` = token:cyan = `#067dcd`) | 旧:浅背景+浅蓝边框(对比度差);新:浅背景+深青边框(对比 4.5:1 ✓) |
| `form_popup.go:175,336` highlight 背景 | `style.Colors.Surface` (`#f0f0f0`) | `component.GetStyle("headerBar").GetBackground()` | 旧:Surface 与 panel body 背景 `#fefefe` 同色 → highlight 不可见;新:`#e8e8e8` 与 body 有差异 |
| `dialog/{choice,exec,selection,form}.go` 7 处 unselected 按钮前景 | `style.Colors.Gray` (`#9b9b9b`) | `component.GetStyle("dim").GetForeground()` (= `theme.Text.Dim` = token:gray = `#5a6270`) | 旧:对 dialog 浅色背景对比 3.2:1(低于 WCAG AA 4.5:1);新:对比 5.8:1 ✓ |

总 13 处修改,跨 5 文件。`form_popup.go:175,336` 两处 highlight 背景修复最关键——之前用户实际看不到高亮(同色),现在可识别。

未修 B 类(按精确修复原则不动):

- 21 处 `style.Colors.Cyan/Green/Red/Yellow/Orange/White/Blue/Purple` 直接绑调色板 → light 主题下这些 token 解析后的 hex 与浅色背景对比度 OK(4.5:1+),无需改
- 12 处 `dialog.TitleColor: style.Colors.Cyan/Green` 等 → 跟着调色板走,light 下 `#067dcd` 蓝、`#3e8b3e` 绿对比度 OK

未修 B 类影响:**可能与某些自定义主题(用户自定义调色板)对比度变差**——留给用户自定义主题时自行调整。

## 9. 留置生效的变更汇总(R4 + R5 + R6)

R4 (2 文件):
- `internal/tui/styles.go`(3 个 BorderStyle 加背景)
- `internal/tui/ui/widget/actionbar/actionbar.go`(容器加背景 + import style)

R5 (6 文件,11 处):
- `internal/tui/ui/app/layout.go`(2 处加背景 + import style)
- `internal/tui/ui/widget/panel/panel.go`(1 处加背景 + import style)
- `internal/tui/ui/pages/compose/view.go`(3 处加背景)
- `internal/tui/ui/pages/help/view.go`(2 处加背景 + import style)
- `internal/tui/ui/pages/logs/view.go`(1 处加背景 + import style)
- `internal/tui/ui/pages/detail/view.go`(1 处加背景 + import style)

R6 (5 文件,13 处):
- `internal/tui/ui/widget/header/header.go`(2 处 keyBadge/keystroke border)
- `internal/tui/ui/widget/dialog/form_popup.go`(2 处 highlight 背景)
- `internal/tui/ui/widget/dialog/choice.go`(1 处 optionStyle)
- `internal/tui/ui/widget/dialog/exec.go`(3 处 confirmBtn/cancelBtn/optionBtns)
- `internal/tui/ui/widget/dialog/selection.go`(2 处 confirmBtn/cancelBtn)
- `internal/tui/ui/widget/dialog/form.go`(5 处 confirm/cancel/value/renderBtn/base)

## 10. R7 增量:B 类全量重构(dialog 30 处 + 新增 3 个 token 槽位)

R6 修了"对比度差"的 13 处;剩 21 处 `style.Colors.{Cyan,Green,Red,Yellow,Orange,White,Purple}` 在 dialog/header 中直接绑调色板——它们**跟主题走**(因为 `tui.ApplyTheme:19-30` 把 `theme.Palette.X` 注入到 `style.Colors.X`),所以本身不构成"颜色泄露",但代码上**隐式依赖注入绑定**,用户自定义主题时不易追踪。

R7 决定:**显式化**——把 30 处 `style.Colors.X` 全部替换为 `component.GetStyle(name).GetForeground()`,新增 3 个 token 槽位让所有 dialog 按钮色都走主题 token。

### 10.1 新增 token 槽位(在 styles_load.go)

| 槽位 | 投影 | 用途 |
|---|---|---|
| `DialogConfirm` | `theme.Text.Success` (前景) + `theme.Toast.Background` (背景) | confirm/ok 按钮 + focused 按钮 |
| `DialogError` | `theme.Text.Error` (前景) + `theme.Toast.Background` (背景) | error 提示 |
| `DialogWarning` | `theme.Text.Warning` (前景) + `theme.Toast.Background` (背景) | warning + image debug dialog title |

复用现有槽位:
- `panelTitle` = `theme.Main.Title` (token:cyan) → 替换所有 `style.Colors.Cyan`
- `dim` = `theme.Text.Dim` (token:gray) → 替换所有 `style.Colors.Gray`
- `helpDesc` = `theme.Text.HelpDescription` (token:white) → 替换所有 `style.Colors.White`

### 10.2 改动(7 文件,30 处)

| 文件 | 替换数 | 备注 |
|---|---|---|
| `internal/tui/ui/component/styles_load.go` | +3 字段 + 3 投影 + 3 lookup case | schema 扩展,JSONC 不动(用现有 `theme.Text.Success/Error/Warning` token) |
| `internal/tui/ui/widget/dialog/exec.go` | 8 处(Green/Cyan/Gray) | 删除 `style` import |
| `internal/tui/ui/widget/dialog/form.go` | 7 处(Green/Cyan/Red/White) | 删除 `style` import |
| `internal/tui/ui/widget/dialog/form_popup.go` | 7 处(Cyan/Red) | 删除 `style` import |
| `internal/tui/ui/widget/dialog/choice.go` | 1 处(Cyan) | 删除 `style` import |
| `internal/tui/ui/widget/dialog/selection.go` | 1 处(Cyan) | 删除 `style` import |
| `internal/tui/ui/widget/dialog/notification.go` | 1 处(Cyan) | 删除 `style` import |
| `internal/tui/ui/widget/dialog/view.go` | 3 处(Yellow/Orange) | 删除 `style` import + 新增 `component` import |

### 10.3 设计取舍

- **新增 3 个 token 槽位 vs 复用现有**:复用以 `panelTitle/dim/helpDesc` 替代 Cyan/Gray/White 是合适的——它们的语义"通用色"在 dialog 中本就符合;但 Green/Red/Yellow/Orange 是"动作色"(确认/危险/警告/debug),需要独立槽位。`theme.Text.Success/Error/Warning` 已存在 schema,所以**只需新增 rawStyles 字段**,JSONC 6 主题 1:1 镜像保持(47 键未变)。
- **不暴露给 JSONC**:3 个新槽位**不暴露**到主题 JSONC——它们的语义跨主题一致(确认=绿、错误=红、警告=黄),用户不应改;如未来要改,只改 `theme.Text.Success/Error/Warning` 即可。
- **消除隐式绑定**:替换后所有 dialog 按钮/TitleColor 都通过 `component.GetStyle(name)` 显式读主题 token——用户自定义主题修改 `theme.Text.Success` 立刻可见,无需再追 `style.Colors.Green` 这种间接绑定。

### 10.4 验证

```text
go test ./...        # 全部通过
go vet ./...         # 无诊断
JSONC 嵌套键 1:1 镜像:6 主题均 47 键
dialog 包内 style.Colors.* 使用数:从 30 处 → 0 处
```

### 10.5 留置生效的变更汇总(R4 + R5 + R6 + R7)

R4 (2 文件):3 BorderStyle + actionbar 容器加背景
R5 (6 文件,11 处):A 类主容器加背景
R6 (5 文件,13 处):B 类对比度差修复
R7 (8 文件,30 处):B 类全量重构 — 新增 3 token 槽位 + 全部替换为 `component.GetStyle()`

**总 21 文件改动,涉及主题/dialog/panel/actionbar/header/styles 等**

## 11. R8 架构级重设计:"仅主背景"架构(2026-08-04)

### 11.1 用户反馈 + 架构洞察

用户在 R4-R7 完成后用截图反馈:light 主题下 header 和 footer 仍画独立背景色,看起来"仍是 dark 主题"。用户的明确指引:

> "一般情况下都不需要单独指定主题色,直接透明 共用背景色即可。只有部分组件需要用于单独设置背景色"

这是对**整个背景架构**的纠正——之前 R4-R7 我做了"给每个主容器加 `.Background(style.Colors.BG)`",但用户实际期望的是:

- **绝大多数组件不画独立背景**——它们"透明",让外层 `palette.background` 透过
- **只有少数强调组件**(dialog、selectedRow、marked row、toast、keyBadge 等)保留独立背景

### 11.2 实施

修改 6 主题 JSONC,4 个主背景槽位从 `token: dark/surface` 改为 `token: background`:

| 槽位 | 旧 | 新 | 视觉效果 |
|---|---|---|---|
| `theme.Header.Background` | `token: dark` (light=#e8e8e8, dark=#1e1f22) | `token: background` (light=#fefefe, dark=#181919b) | header 与主背景同色——视觉"透明" |
| `theme.Footer.ShortcutBackground` | `token: dark` | `token: background` | footer keybar 与主背景同色 |
| `theme.Footer.StatusBackground` | `token: surface` | `token: background` | footer status 行与主背景同色 |
| `theme.Toast.Background` | `token: dark` | `token: background` | toast 与主背景同色(仍可见 text/icon) |

### 11.3 设计取舍

- **不动 `theme.Palette.dark` / `theme.Palette.surface` 槽位**:它们仍是"通用浅深背景色"语义,其它代码可能引用(如 `LogHighlight` 在 styles_load.go:125 用 `theme.Header.Background`,R8 之后该 slot 等同 `theme.Palette.Background`,但 `LogHighlight.Background` 用的是 `theme.Text.Warning` ——独立)。
- **不动 dialog.overlay**:dialog 遮罩仍按主题独立配色(由 `theme.Dialog.Overlay` + `OverlayOpacity` 决定),不是"透明"。
- **不动 selectedRow / markedRow / keyBadge / toastIcon**:这些强调块保留独立背景,符合用户"只有部分组件需要单独设置背景色"的指引。
- **不动 detailSelection / logHighlight**:同上。

### 11.4 验证

```text
go test ./...        # 全部通过
go vet ./...         # 无诊断
JSONC 嵌套键 1:1 镜像:6 主题均 47 键
```

### 11.5 留置生效的变更汇总(R4 + R5 + R6 + R7 + R8)

R4 (2 文件):3 BorderStyle + actionbar 容器加背景
R5 (6 文件,11 处):A 类主容器加背景
R6 (5 文件,13 处):B 类对比度差修复
R7 (8 文件,30 处):B 类全量重构——新增 3 token 槽位 + 全部替换为 `component.GetStyle()`
R8 (6 文件):4 个主背景槽位 6 主题统一为 `token: background`——架构级纠正

**总 27 文件改动**

## 12. R9 增量:C 类启动兜底 + 透明/RGBA JSONC 支持设计(2026-08-04,讨论)

### 12.1 C 类兜底现状（保留不动）

C 类 = 24 处编译期 hex 字面量:

- `internal/tui/ui/style/style.go:35-48` —— 12 处 `var Colors = Palette{}` 默认值（与 default 主题一致，启动兜底）
- `internal/data/config/constants_domain.go:139-153` —— 15 处 `FallbackColor*` 常量（`DefaultTheme()` 的 palette 兜底）

**R9 决策：不改**。理由:

1. `style.Colors.*` 是 `init()` 之前 `utils.ParseColor("green")` 解析调色板名的依赖——移除会让 `tui.ApplyTheme` 之前的渲染调用崩溃
2. `FallbackColor*` 是用户没指定主题时的最后兜底——移除会让 `theme.Palette` 在极端启动场景下为空
3. 这 24 处兜底与 R8 架构"透明/共用背景"**不冲突**——它们只在启动最早期生效（毫秒级），用户视觉看不到

### 12.2 透明/RGBA JSONC 支持设计（用户讨论需求，2026-08-04 留置，未实现）

用户明确需求:

> "jsonc 配置也应该支持透明，用户根据需要配置 颜色（例如 RGBA）"

#### 12.2.1 当前支持现状

| 路径 | 实现状态 | 备注 |
|---|---|---|
| `#RRGGBBAA` 9 位 hex | ✅ 已支持 | `isHexColor` validator 接受 4/7/9 位；`utils.ParseColor` 解析；`lipgloss.Color` 在 truecolor 终端输出 `\033[48;2;R;G;B m`（alpha 通常被 SGR 忽略或真混合） |
| 调色板 token 透明 | ❌ 不支持 | `theme.Palette.*` 12 个 token 全是 hex，无 `transparent` 概念 |
| 空字符串 sentinel | ❌ 不支持 | `isHexColor` 不接受空值 |
| `{"value": ""}` | ❌ 不支持 | 同上 |

#### 12.2.2 推荐设计（未来实现）

| 设计点 | 推荐方案 | 理由 |
|---|---|---|
| 透明 sentinel 表达 | 新增 `ColorTokenTransparent = "transparent"` 调色板 token，`Theme.ResolveColor` 遇到返回空字符串；lipgloss 处理空背景=不画 | 与现有 token 体系一致；用户写 `{"token": "transparent"}` 自然 |
| JSONC 配置 `{"value": ""}` | 拒绝 | 空字符串歧义多；走显式 token 更稳 |
| 终端 alpha 降级 | 普通 256 色终端 alpha 被忽略 → 显示全不透明 | 真透明仅在 truecolor 终端生效；需文档说明 |
| 哪些槽位默认透明 | 按 R8 "绝大多数透明"原则，主背景槽位建议 `{"token": "transparent"}`；强调块保留实色 | 与 R8 一致 |

#### 12.2.3 实施清单（如果未来启动）

1. `internal/data/config/constants_domain.go` —— 新增 `ColorTokenTransparent`
2. `internal/data/config/theme.go` —— `Theme.ResolveColor` 增加 `case ColorTokenTransparent: return ""`；`DefaultTheme()` 给 4 个透明槽位默认 `TokenRef(ColorTokenTransparent)`（Header.Background、Footer.StatusBackground、Footer.ShortcutBackground、Toast.Background）
3. `internal/tui/ui/component/styles_load.go` —— `styleRef` 已是 string，Background 空字符串语义需要 lipgloss 适配——检查 `.Background(lipgloss.Color(""))` 是否等价于 NoColor
4. JSONC 6 主题的 4 个槽位从 `token: background` 改为 `token: transparent`（配合 default 默认）
5. 测试：加 `TestResolveColorReturnsEmptyForTransparent` 断言

#### 12.2.4 用户文档要点

- 默认哪些槽位是透明、哪些是实色
- 真透明需要 truecolor 终端
- RGBA 用法示例：`{"value": "#18191bcc"}`（alpha cc = 80%）
- 自定义主题时如何选透明 vs 实色

### 12.3 R9 验证

```text
go test ./...        # 全部通过
go vet ./...         # 无诊断
JSONC 嵌套键 1:1 镜像:6 主题均 47 键
C 类代码改动数:0（按讨论结果保留启动兜底）
```

### 12.4 留置生效的变更汇总(R4 + R5 + R6 + R7 + R8 + R9)

R4 (2 文件):3 BorderStyle + actionbar 容器加背景
R5 (6 文件,11 处):A 类主容器加背景
R6 (5 文件,13 处):B 类对比度差修复
R7 (8 文件,30 处):B 类全量重构——新增 3 token 槽位 + 全部替换为 `component.GetStyle()`
R8 (6 文件):4 个主背景槽位 6 主题统一为 `token: background`——架构级纠正
**R9 (0 文件改动)**:C 类启动兜底保留不动 + 透明/RGBA JSONC 支持设计留置

**总 27 文件改动 + §12 设计留置**

## 13. 链接到 R10 透明架构

R10(2026-08-04 实施 Q1-Q4 全 a 方案)是 R8/R9 "仅主背景 + 透明兜底"架构的完整实现:

- 新增 `ColorTokenTransparent` 调色板 token,JSONC 写 `{"token": "transparent"}` 表示"不画背景"
- 新增 `theme.Dialog.BodyBackground` 槽位区分弹框主体与遮罩
- 新增 `internal/tui/terminal.go` `SupportsTruecolor()` 探测
- 6 主题 JSONC 改 4 个主背景槽位为 `token: transparent` + 新增 dialog.bodyBackground 为主背景+5% 加亮

完整设计与文字版图层结构见 [./theme-transparent-architecture.md](./theme-transparent-architecture.md)

**总 34 文件改动(R4-R10 累计)**

## 7. 留置生效的变更汇总(R4 + R5)

R4 (2 文件):
- `internal/tui/styles.go`(3 个 BorderStyle 加背景)
- `internal/tui/ui/widget/actionbar/actionbar.go`(容器加背景 + import style)

R5 (6 文件):
- `internal/tui/ui/app/layout.go`(2 处加背景 + import style)
- `internal/tui/ui/widget/panel/panel.go`(1 处加背景 + import style)
- `internal/tui/ui/pages/compose/view.go`(3 处加背景)
- `internal/tui/ui/pages/help/view.go`(2 处加背景 + import style)
- `internal/tui/ui/pages/logs/view.go`(1 处加背景 + import style)
- `internal/tui/ui/pages/detail/view.go`(1 处加背景 + import style)

## 6. 验证

```text
go test ./...        # 全部通过
go vet ./...         # 无诊断
JSONC 嵌套键 1:1 镜像:6 主题均 47 键
```

手工验证(用户应能直接观察):
- `--theme light`:action bar / panel 容器 / border 全部显示 `#fefefe` 浅色底,前景文字是 `#1f2328` 深色——视觉一致
- `--theme dark`:容器显示 `#18191b` 深色底,前景文字是 `#c9d1d9` 浅色——视觉一致
- `--theme nord`:容器显示 `#242933` 北欧深色底——视觉一致
- `layout.Background.Fallthrough=true` 时,`tui.ApplyLayoutConfig` 仍会把 ActiveBorderStyle 的背景清成 NoColor——保持透明

## 7. 留置生效的变更汇总

2 文件改动:
- `internal/tui/styles.go`(3 个 BorderStyle 加背景)
- `internal/tui/ui/widget/actionbar/actionbar.go`(容器加背景 + import style 包)

未改动主题 JSONC、schema、Go 业务逻辑。