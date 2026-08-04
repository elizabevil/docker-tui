# R12: style.Colors 字段重命名为语义名

> 日期: 2026-08-04
> 状态: 已实施，代码与测试全绿
> 关联: R11 主题 token 重命名（已实施）

## 1. 目标

`internal/tui/ui/style/style.go` 的 `var Colors = Palette{...}` 12 字段名从颜色名改为语义名，与 R11 的 `config.Palette` 一致。

**理由**：`style.Colors.Green` 字段名误导——实际存的是 `theme.Palette.Success`（成功语义色），但字段名说"绿色"。这正是 R11 想消除的问题，不应在启动兜底层残留双标。

## 2. 重命名映射

| style.Colors 字段 | 对应 config.Palette | hex 值（不变） |
|---|---|---|
| `Green` → `Success` | `Success` | `#499c54` |
| `Cyan` → `Primary` | `Primary` | `#56b4c2` |
| `Blue` → `Info` | `Info` | `#589df6` |
| `Red` → `Danger` | `Danger` | `#db5a5a` |
| `Yellow` → `Warning` | `Warning` | `#c8a35e` |
| `Orange` → `Accent` | `Accent` | `#cc7832` |
| `Purple` → `AccentSecondary` | `AccentSecondary` | `#a962b5` |
| `White` → `Foreground` | `Foreground` | `#c9d1d9` |
| `Gray` → `ForegroundMuted` | `ForegroundMuted` | `#5a6270` |
| `Dark` → `BackgroundSubtle` | `BackgroundSubtle` | `#1e1f22` |
| `Surface` → `BackgroundDeep` | `BackgroundDeep` | `#2b2d30` |
| `BG` 保留 | `Background` | `#18191b` |

## 3. constants.Color* 语义别名

`internal/constants/color.go` 同步重命名：

- 新增 12 个语义常量：`ColorSuccess/Primary/Info/Danger/Warning/Accent/AccentSecondary/Foreground/ForegroundMuted/Background/BackgroundSubtle/BackgroundDeep/ColorBG`
- 保留 12 个 legacy 别名指向语义常量：`ColorGreen=ColorSuccess`、`ColorCyan=ColorPrimary`、...、`ColorGrey=ColorForegroundMuted`

```go
const (
    ColorSuccess          = "success"
    ColorPrimary          = "primary"
    ...
    ColorBG               = "bg"
)

// Legacy aliases point to the new semantic palette slots
const (
    ColorGreen   = "success"
    ColorCyan    = "primary"
    ...
    ColorGrey    = "foregroundMuted"
)
```

## 4. 改动文件清单（4 文件）

| 文件 | 改动 |
|---|---|
| `internal/constants/color.go` | 12 个语义常量 + 12 个 legacy 别名（指语义名） |
| `internal/tui/ui/style/style.go` | `Palette` struct 11 字段重命名（`BG` 保留）+ `var Colors` 默认值 + `SyncPalette` 11 行 + `Color()` fallback `Colors.Foreground` |
| `internal/tui/styles.go` | `ApplyTheme` 注入映射 11 行（`style.Colors.Green = style.Color(string(c.Success))` 等） |
| `internal/tui/ui/widget/dialog/box.go` | `sty.Colors.Cyan` → `sty.Colors.Primary`（1 处） |
| `internal/tui/ui/widget/header/header.go` | `cpuLoadColor` 3 处：`Red/Yellow/Green` → `Danger/Warning/Success` |
| `internal/utils/color_test.go` | `TestParseColorPaletteAndANSI` 改用新名：`SetPaletteColor(ColorForeground/ColorForegroundMuted)` + `ParseColor("foreground"/"FOREGROUND "/"foregroundMuted")` |

## 5. 验证

```text
go vet ./...                          # 无诊断
go test ./...                         # 35 包全过（含 TestParseColorPaletteAndANSI 改用新名）
旧 style.Colors.* 引用扫描             # 0 处遗留（仅 BG 保留）
旧 constants.Color* 原始字符串扫描    # 0 处遗留（所有 legacy 别名指向语义名）
JSONC 嵌套键 1:1 镜像                 # 6 主题均 47 键（R11 保持）
```

## 6. 已知保留项

- `style.Colors.BG` 字段名不变（已语义化 "Background"）
- 12 个编译期 hex 值不变（与 R11 一致）

## 7. 不可触碰边界

- 不动 `style.Colors.BG` 字段
- 不动 12 个编译期 hex 值
- 不动 `config.Palette` schema（R11 已完成）
- 不动 JSONC 6 主题（R11 已完成）
- 不提交（等用户确认）

## 8. 留置生效的变更汇总（R4-R12 累计）

R4-R11: 47 文件改动（详见各 proposal）
**R12**: 4 文件改动（constants/color.go + style.go + styles.go + box.go + header.go + color_test.go = 6 文件）

**总 ~53 文件改动（R4-R12 累计）**

## 9. 增量清理:Legacy 别名与 caller 全部移除

用户反馈"明显不符合要求"指出：尽管 R12 已为 12 字段改为语义名，legacy 别名 `ColorGreen = "success"`、`ColorCyan = "primary"` 等仍残留——`internal/tui/ui/component/stateicon.go` 与 `dialog.go` 仍引用旧名（通过别名间接指向新名）。

### 9.1 决策

**完全删除 legacy 别名块**——所有 caller 已迁移到新语义名，legacy 别名已无 caller。删除后语义哲学完整：
- `constants.ColorSuccess` 是唯一的 "success" 名
- `stateicon.go` / `dialog.go` 等全部用 `ColorSuccess/Warning/ForegroundMuted/Info/Accent/Primary`

破坏性变更：`ParseColor("white")`、`ParseColor("grey")` 等历史名不再工作（与 R12 §3 一致——解析路径走语义名）。

### 9.2 改动（3 文件增量）

| 文件 | 改动 |
|---|---|
| `internal/constants/color.go` | 删除 12 个 legacy 别名块（`ColorGreen/Cyan/Blue/Red/Yellow/Orange/Purple/White/Gray/Dark/Surface/Grey`） |
| `internal/tui/ui/component/dialog.go` | 4 处：`ColorYellow/Cyan` → `ColorWarning/Primary` |
| `internal/tui/ui/component/stateicon.go` | 7 处：`ColorGreen/Yellow/Gray/Blue/Orange` → `ColorSuccess/Warning/ForegroundMuted/Info/Accent` |

### 9.3 验证

```
go test -count=1 ./...                                          → 35 包全过
go vet ./...                                                     → 干净
旧 ColorGreen/Cyan/Blue/Red/Yellow/Orange/Purple/White/Gray/Dark/Surface/Grey 引用扫描 → 0 处遗留
```

### 9.4 影响

- `internal/constants/color.go` 现在只有 12 个语义色名 + 8 个 CSS 标准色名
- 所有 Go 代码引用统一用语义名（`ColorSuccess/Primary/Info/Danger/Warning/Accent/AccentSecondary/Foreground/ForegroundMuted/Background/BackgroundSubtle/BackgroundDeep/BG`）
- R12 完整版 = 6 文件（主改动）+ 3 文件（增量清理）= 9 文件