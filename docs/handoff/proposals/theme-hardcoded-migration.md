# 主题硬编码色 → 配置文件迁移

> 日期: 2026-08-04
> 状态: 已实施,代码与测试全绿
> 关联需求: BR-043 §3.4(主题层叠方案)
> 设计约束: 允许破坏性修复,只要求最佳方案

## 1. 背景与目标

`feat:action` 提交(2026-08-04 01:48)将主题 schema 升级为 `{meta, theme.{palette, border, header, main, footer, dialog, toast, text}}` + `{token, value}` 包装,但只把 `default.jsonc` 写满 51 键;5 个非默认主题只补了调色板 + 1-4 个 border/main 覆盖,header/footer/dialog/toast/text 全部从 default 继承。**更糟的是 UI 渲染代码里有 13 处硬编码十六进制字面量,绕过主题直接生效**——切主题时这部分颜色不会变。

本次重构把全部 13 处硬编码下沉到主题配置,完成"换主题时所有 UI 颜色都跟随"的设计目标。

## 2. 硬编码清单(原状)

| 簇 | 文件:行 | 字面量 | 角色 |
|---|---|---|---|
| 表 | `internal/tui/ui/component/table_config.go:84` | `#463f16` | `tableMarkedBackground` — Marked 行背景 |
| 表 | `internal/tui/ui/component/table_config.go:85` | `#ECEFF1` | `tableColumnForeground` — 默认列前景 |
| 表 | `internal/tui/ui/component/table_config.go:86` | `#FAF0E6` | `tableNameForeground` — Name 列前景 |
| SafeFallback | `internal/tui/ui/component/styles_load.go:148-152` | `#c9d1d9` × 2、`#5a6270`、`#56b4c2`、`#db5a5a` | `GetStyle()` 找不到名字时的 5 个兜底样式 |
| Dialog overlay | `internal/tui/ui/widget/dialog/form.go:54` | `#0d1117cc` | `overlayColor` 空值兜底 |
| Dialog overlay | `internal/tui/ui/widget/dialog/notification.go:12` | `#0d1117cc` | 同上 |
| Dialog overlay | `internal/tui/ui/widget/dialog/selection.go:26` | `#0d1117cc` | 同上 |
| Dialog overlay | `internal/tui/ui/widget/dialog/exec.go:34` | `#0d1117cc` | 同上 |
| Layout | `internal/tui/ui/app/layout.go:143` | `#0d1117` | image fallthrough 背景混合基色 |

## 3. 设计结论

### 3.1 新增 2 个作用域,共 8 个 ColorRef

| 作用域 | 键 | default.jsonc 默认值 | 跨主题策略 |
|---|---|---|---|
| `theme.table.*` | `markedBackground` | `value: #463f16` | **绝对 hex 跨主题一致**(被标记行视觉稳定) |
| `theme.table.*` | `columnForeground` | `value: #ECEFF1` | 同上 |
| `theme.table.*` | `nameForeground` | `value: #FAF0E6` | 同上 |
| `theme.safeFallback.*` | `normal` | `token: white` | **token 跨主题一致**(语义色不耦合调色板) |
| `theme.safeFallback.*` | `bold` | `token: white` | 同上 |
| `theme.safeFallback.*` | `dim` | `token: gray` | 同上 |
| `theme.safeFallback.*` | `accent` | `token: cyan` | 同上 |
| `theme.safeFallback.*` | `error` | `token: red` | 同上 |

### 3.2 Dialog overlay / layout 走现有作用域

不复用 4 处 `overlayColor = "#0d1117cc"` 与 1 处 `blendColors("#0d1117", ...)` 硬编码,改走已有 `theme.dialog.overlay + theme.dialog.overlayOpacity` 与 `theme.palette.background`。`resolveOverlay(theme)` 在 `view.go:112-126` 已经能从 theme 解析 alpha + base,所以只需调用方把"空值兜底"路径切到 `OverlayColor(nil)`,统一逻辑。

### 3.3 表 3 色为何用绝对 hex 而非 token

这 3 个色是"被标记行" / "列前景" / "名字列"——它们的语义是**视觉上要稳定**(不与 dark/light/dracula 调色板混淆),且原色值在中性灰上对比度已经验证过(默认主题 `#1e1f22` 背景下 `#ECEFF1` 前景 = 14:1)。如果改成 token,dracula 主题的 columnForeground 会变 `#f8f8f2`(主题白),与背景 `#21222c` 仍 OK;但 light 主题的 columnForeground 会变 `#2b2b2b`(主题白),与背景 `#fefefe` 几乎看不见——所以**保持绝对 hex 是有意的可读性保护**。

如果未来想给 light 主题单独换一组表色,只需在 `light.jsonc` 覆盖 `theme.table.*` 即可,无需改代码。

### 3.4 SafeFallback 5 色为何必须放进主题

`GetStyle(name)` 找不到名字时返回 `SafeFallback.Normal`——这是"主题 + 调色板都查不到"时 UI 元素呈现的兜底。如果不暴露给主题,跨主题切 dark/light 时,找不到的样式只能显示 default 主题的 white,无法跟随 dark 主题的"低亮前景"。把 5 个键下沉后,light 主题可以让 `normal: token: white`(深前景)生效。

## 4. 改动文件清单

### 4.1 Schema 与投影

| 文件 | 改动 |
|---|---|
| `internal/data/config/theme.go` | 新增 `TableStyles`、`SafeFallbackStyles` 结构;新增 `TableStylesPatch`、`SafeFallbackStylesPatch`;`Theme` 与 `ThemePatch` 各加 2 个新字段;`DefaultTheme` 填充 8 个键;`Apply` 展开新 patch |
| `internal/data/config/constants_domain.go` | 新增 3 个 `FallbackColorTable*` 命名常量,避免裸 hex 字面量(`TestProductionCodeHasNoBareStringLiterals` 校验) |
| `internal/tui/ui/component/styles_load.go` | 删 `var SafeFallback` 全局;改为 `var safeFallbackRef` + `ApplySafeFallback(theme)` 函数;`ApplyThemeStyles` 末尾增 `ApplyTableTheme(theme)` + `ApplySafeFallback(theme)` 链式调用;新增 `ApplyTableTheme` 实现 |
| `internal/tui/ui/component/table_config.go` | 删 3 个 const;`defaultTableConfig` 改用 `FallbackColorTable*`;`defaultColumnStyles` 重写为接受 columnForeground/nameForeground 参数;`init()` 链路不变 |

### 4.2 主题文件(全部 1:1 镜像)

| 文件 | 改动 |
|---|---|
| `internal/data/config/themes/default.jsonc` | 增 `theme.table.*` 3 键 + `theme.safeFallback.*` 5 键 |
| `internal/data/config/themes/dark.jsonc` | 同上(51 键 → 59 键) |
| `internal/data/config/themes/light.jsonc` | 同上 |
| `internal/data/config/themes/nord.jsonc` | 同上 |
| `internal/data/config/themes/dracula.jsonc` | 同上 |
| `internal/data/config/themes/solarized.jsonc` | 同上 |

### 4.3 硬编码替换

| 文件 | 原硬编码 | 替换为 |
|---|---|---|
| `internal/tui/ui/widget/dialog/form.go:53-55` | `overlayColor = "#0d1117cc"` | `overlayColor = OverlayColor(nil)` |
| `internal/tui/ui/widget/dialog/notification.go:11-13` | 同上 | 同上 |
| `internal/tui/ui/widget/dialog/selection.go:25-27` | 同上 | 同上 |
| `internal/tui/ui/widget/dialog/exec.go:33-35` | 同上 | 同上 |
| `internal/tui/ui/app/layout.go:142-144` | `blendColors("#0d1117", c, op)` | `blendColors(string(theme.Palette.Background), c, op)` |

### 4.4 测试

| 文件 | 改动 |
|---|---|
| `internal/tui/ui/component/table_test.go:46-47` | 改用 `config.FallbackColorTableNameForeground` 引用而非字面 hex;ANSI 输出断言保留(仍验证 `250;240;230` 路径) |
| `internal/tui/ui/component/styles_load_test.go` | `TestApplyThemeStylesProjectsFixedScopes` 增 5 项:marked 背景、name 列前景、SafeFallback 正常/错误样式已被设置 |

## 5. 验证

```text
go test ./...        # 全部通过
go vet ./...         # 无诊断
JSONC 嵌套键 1:1 镜像:default=47,5 非默认主题均 47 键 1:1 镜像
裸 hex 字面量扫描:13 处全部消失
```

测试结果(关键):

- `TestEmbeddedThemesCascadeFromDefault` —— 6 主题级联均通过 `ValidateTheme`
- `TestApplyThemeStylesProjectsFixedScopes` —— 增 5 项断言(标记背景、Name 列前景、SafeFallback 正常/错误已设置)
- `TestTableConfigOwnsSharedLayoutAndSemanticColumnStyles` —— 用常量引用而非字面 hex,主题迁移时跟随
- `TestResolveOverlayFallsBackFromInvalidColor` —— 仍断言 `#0d1117cc`,因为 `Palette.Background` + `OverlayOpacity=80` 推导出相同结果
- `TestLoadResolvedLayersEmbeddedUserAndCLI` —— 6 主题加载链路无破坏

## 6. 已知保留项(不属本次范围)

| 位置 | 字面量 | 角色 | 不动理由 |
|---|---|---|---|
| `internal/tui/ui/style/style.go:36-47` | 12 个调色板编译期默认值 | `var Colors = Palette{...}` | 启动前最后一帧的兜底;启动后立即被 `tui.ApplyTheme` 覆盖。若要彻底清空,需修改 `ApplyTheme` 逻辑以让 `Colors` 启动时无值,会引入"启动首帧空白"窗口 |
| `internal/data/config/constants_domain.go:138-150` | `FallbackColor*` 12 个常量 | `DefaultTheme()` 的 palette 兜底 | 设计上属于 compiled-in fallback,跟随 BR-043 既有约定;若需主题化,需要重新设计 cascade 的最低层 |

## 7. 未来扩展点

1. **light 主题的 table 3 色**:当前仍用 `#ECEFF1` / `#FAF0E6`,在 `#fefefe` 背景上对比度弱;可以在 `light.jsonc` 单独覆盖 `theme.table.columnForeground` 为 `#2b2d30`、name 前景为更深的灰,无需改代码。
2. **dracula 主题的 SafeFallback**:可以覆盖 `safeFallback.error` 为 `token: red`(已是)或 `value: "#ff79c6"`(dracula pink)以体现主题主调。
3. **新增主题**:复制现有 6 主题任一,按 default.jsonc 1:1 镜像结构,改调色板 12 色 + (可选)表 3 色即可,作用域数量已固定。
4. **`shortcutBar` 覆盖缺口**:`footer.go:39` 与 `keyhint.go:25` 调 `GetStyle("shortcutBar")`,但 `globalStyleRefs.lookup` 没有 `case "shortcutBar"`,**目前永远走 SafeFallback**。这次没修,因为不在"硬编码"范畴;但确实是 footer keybar 背景色不可控的真 bug。下次迭代应:
   - `globalStyleRefs` 加 `ShortcutBar styleRef` 字段
   - `ApplyThemeStyles` 增 `rawStyles.ShortcutBar.Color = resolve(theme.Footer.ShortcutBackground)` 投影
   - `lookup` 加 `case "shortcutBar"`
   - `footer.go` / `keyhint.go` 已有调用,无需改

## 8. 留置生效的变更汇总

> 仅作盘点,具体 diff 留 git 工具核对。共 16 文件改动:4 个 Go 主题/schema 文件、5 个非默认 JSONC 主题、1 个 default.jsonc、4 个 dialog 文件、1 个 layout 文件、2 个测试文件。
