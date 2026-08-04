# R11: 主题 token 重命名为作用范围名（方案 C，破坏性修改）

> 日期: 2026-08-04
> 状态: 已实施，代码与测试全绿
> 关联: R4-R10 累计 34 文件改动
> 用户决策（2026-08-04 讨论）: 方案 C；Q1 重命名不缩写+允许破坏性修改+不遗留兼容；Q2 用方案 b；Q3 用方案 a；Q4 不要单独映射层

## 1. 目标

把 12 调色板槽位从"颜色名"改为"作用范围名"，让用户写 `token: success` 直接表达"成功语义"而非"绿色"。

**关键设计**：
- 调色板槽位 = 主题语言锚点（hex 值），主题切换时基础色可改
- 作用域 token 直接引用调色板槽位（无中间 semantic 层）
- 主题作者可为每个语义槽位选任意 hex（dracula 可让 `success = 紫`）

## 2. Token 重命名映射

| 旧 | 新 | 旧 JSONC 字面量 | 新 JSONC 字面量 |
|---|---|---|---|
| ColorTokenGreen | ColorTokenSuccess | "green" | "success" |
| ColorTokenCyan | ColorTokenPrimary | "cyan" | "primary" |
| ColorTokenBlue | ColorTokenInfo | "blue" | "info" |
| ColorTokenRed | ColorTokenDanger | "red" | "danger" |
| ColorTokenYellow | ColorTokenWarning | "yellow" | "warning" |
| ColorTokenOrange | ColorTokenAccent | "orange" | "accent" |
| ColorTokenPurple | ColorTokenAccentSecondary | "purple" | "accentSecondary" |
| ColorTokenWhite | ColorTokenForeground | "white" | "foreground" |
| ColorTokenGray | ColorTokenForegroundMuted | "gray" | "foregroundMuted" |
| ColorTokenDark | ColorTokenBackgroundSubtle | "dark" | "backgroundSubtle" |
| ColorTokenSurface | ColorTokenBackgroundDeep | "surface" | "backgroundDeep" |
| ColorTokenBackground | ColorTokenBackground | "background" | "background"（不变） |
| ColorTokenTransparent | ColorTokenTransparent | "transparent" | "transparent"（不变） |

**Palette 字段名也对应重命名**：Palette.Green → Palette.Success 等。

## 3. 改动文件清单（~13 文件）

### 3.1 Go 代码（7 文件）

| 文件 | 改动 |
|---|---|
| `internal/data/config/constants_domain.go` | `ColorToken` enum 13 值重命名；`FallbackColor*` 12 个 hex 兜底常量重命名（值不变，名字改） |
| `internal/data/config/theme.go` | `Palette` struct 12 字段重命名；`PalettePatch` 12 字段重命名；`Theme.ResolveColor` switch 13 case 重命名；`DefaultTheme()` 全部引用重命名；`Apply` 内部 `assign(&target.Palette.X, ...)` 重命名 |
| `internal/data/config/config.go` | `ValidateTheme` 中 `theme.Palette.Green/...` 引用重命名；`isColorRef` switch case 13 值重命名 |
| `internal/tui/styles.go` | `ApplyTheme` 中 `c.Green/Cyan/...` 引用重命名为 `c.Success/Primary/...` |
| `internal/tui/ui/component/table_config.go` | `defaultTableConfig` + `defaultColumnStyles` 中 `ColorTokenOrange/Green/Yellow/White/Blue/Red/Cyan/Dark` 引用重命名 |
| `internal/tui/ui/component/styles_load_test.go` | 测试中 `ColorTokenPurple/Orange/Cyan/Surface/White/Red` + `FallbackColorPurple/Orange/Cyan/Surface` 重命名 |
| `internal/data/config/theme_test.go` | 测试中 `ColorTokenPurple/Red` + `FallbackColorRed` 重命名；`PalettePatch{Cyan: ...}` 重命名；`theme.Palette.Cyan/Green` 重命名 |
| `internal/data/config/config_test.go` | 测试中 `token: cyan` / `token: purple` 重命名为 `token: primary` / `token: accentSecondary`；`Palette.Cyan` 引用重命名 |

### 3.2 JSONC 6 主题

每个主题改 2 处：
1. palette 12 字段名重命名（值不变）：`"green": "#xxx"` → `"success": "#xxx"` 等
2. 所有作用域 `token: "green"` → `token: "success"` 等

**未改**：palette hex 值、`dialog.bodyBackground`、`safeFallback.*` 内部 token 引用（已对齐）、`table.*` 三色 hex（值）。

### 3.3 测试（4 文件）

已覆盖上述。

## 4. JSONC 配置示例（new）

```jsonc
{
  "theme": {
    "palette": {
      "primary":          "#42a5f5",
      "success":          "#2ecc71",
      "warning":          "#ffca28",
      "danger":           "#ef5350",
      "info":             "#00bcd4",
      "accent":           "#ff9800",
      "accentSecondary":  "#ab47bc",
      "foreground":       "#ffffff",
      "foregroundMuted":  "#546e7a",
      "background":       "#0d1117",
      "backgroundSubtle": "#1a1a2e",
      "backgroundDeep":   "#16213e"
    },
    "border": {
      "active":   { "token": "primary" },
      "inactive": { "token": "foregroundMuted" },
      "focused":  { "token": "success" }
    },
    "text": {
      "info":    { "token": "info" },
      "success": { "token": "success" },
      "error":   { "token": "danger" },
      "warning": { "token": "warning" },
      "dim":     { "token": "foregroundMuted" }
    }
  }
}
```

## 5. 验证

```text
go test ./...        # 全部通过（35+ 包，含 6 主题级联、token 投影、dialog body background、transparent）
go vet ./...         # 无诊断
JSONC 嵌套键 1:1 镜像:6 主题均 47 键
旧 token 引用扫描:0 处遗留
```

## 6. 已知保留项

- `style.Colors` 12 字段名（`Green/Cyan/Blue/...`）保留为"启动兜底调色板"独立命名空间——与 `theme.Palette` 12 槽位语义对齐但字段名不强制一致
- `FallbackColor*` 12 编译期 hex 兜底值与 `style.Colors` 配套保留
- 24 处 C 类启动兜底（`style.Colors` + `FallbackColor*`）按 R9 设计不变

## 7. 留置生效的变更汇总（R4-R11 累计）

R4-R10: 34 文件改动（详见 `theme-light-bg-fix.md` §13）
**R11**: ~13 文件改动（7 Go + 6 JSONC + 测试），破坏性 token 重命名

**总 ~47 文件改动（R4-R11 累计）**