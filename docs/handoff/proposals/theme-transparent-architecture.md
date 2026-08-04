# 主题透明架构(Q1-Q4 全 a 方案实施)

> 日期: 2026-08-04
> 状态: 已实施,代码与测试全绿
> 关联: [./theme-light-bg-fix.md](./theme-light-bg-fix.md)(R4-R9 历史)
> 关联: [./theme-hardcoded-migration.md](./theme-hardcoded-migration.md)(R0-R3 历史)

## 1. 用户需求(2026-08-04 讨论)

> 1. 颜色配置应该统一 使用 `transparent` 表示透明 避免空字符串歧义。
> 2. 非弹框/选择框 默认均使用透明,而弹框/选择框要区分图层 需要添加与主题颜色类似的背景色。
> 3. 尽可能保证终端行为统一(待详细讨论)。
> 4. 需要给出文字版 图层结构。

## 2. 最终架构决策(Q1-Q4 全选 a)

| 决策点 | 方案 | 实施 |
|---|---|---|
| **Q1** transparent token 实现 | 新增 `ColorTokenTransparent` 调色板 token | `constants_domain.go` 加 `ColorTokenTransparent = "transparent"`;`Theme.ResolveColor` 增加 `case ColorTokenTransparent: return string(FallbackColorTransparent)`(字面量 `"transparent"`);`isColorRef` validator 接受 transparent |
| **Q2** 弹框主体背景 | 新增 `theme.Dialog.BodyBackground` 槽位 | `DialogStyles` 加 `BodyBackground ColorRef`;`DefaultTheme()` 默认 `TokenRef(ColorTokenBackground)`;6 主题 JSONC 覆盖 `value: <主背景 +5% 加亮 hex>` |
| **Q3** 终端行为统一 | per-terminal 探测 truecolor + alpha 降级 | 新增 `internal/tui/terminal.go` `SupportsTruecolor()`(读 `COLORTERM`/`TERM`/`TERM_PROGRAM`);`tui/styles.go` alpha 降级(未来实施) |
| **Q4** transparent 默认范围 | 4 个主背景槽位 + 其它非弹框组件 | JSONC 6 主题改 `header.background`/`footer.statusBackground`/`footer.shortcutBackground`/`toast.background` 为 `token: transparent` |

## 3. 文字版图层结构

```
终端默认背景 (终端/模拟器控制,TUI 不画)
│
├─ Layer 1: 主背景层 (theme.Palette.Background)
│   - light=#fefefe / dark=#0e0f10 / nord=#242933 等
│   - 由 tui.ApplyTheme 注入到 style.Colors.BG
│   - 切主题唯一改变的"基底色"
│
├─ Layer 2: 非弹框组件层 (theme.*.Background = transparent)
│   ├─ Header / Footer / Toast
│   ├─ Panel body / ActionBar / MessageRail / QueryBar
│   └─ 其它非弹框层组件
│
├─ Layer 3: 强调块层 (保留独立颜色)
│   ├─ SelectedRow ─ token: blue (背景)
│   ├─ MarkedRow ─ value: #463f16
│   ├─ KeyBadge ─ token: header.value
│   ├─ DetailSelection ─ token: rowSelected
│   └─ Toast icon/text ─ token: text.success/error
│
└─ Layer 4: 弹框/选择框层 (区别于主背景)
   ├─ 4a: 遮罩 (theme.Dialog.Overlay = Palette.Background + alpha 0.8)
   ├─ 4b: 弹框主体 (theme.Dialog.BodyBackground = 主背景 +5% 加亮)
   └─ 4c: 弹框标题/按钮 (强调色 token)
```

### 3.1 终端行为映射

| Layer | truecolor 终端 | 256 色终端 | 16 色终端 |
|---|---|---|---|
| 1 主背景 | `\033[48;2;R;G;Bm` 真彩 | `\033[48;5;Nm` 最近色 | `\033[4Xm` 16 色 |
| 2 transparent | `lipgloss.NoColor{}` 不画 | 同 left | 同 left |
| 3 accent | 真彩 | 最近色 | 16 色 |
| 4a overlay alpha | 真彩 alpha 混合 | **alpha 降级为不透明**(遮罩仍可见) | 同 left |
| 4b dialog body | 真彩 +5% 加亮 | 最近色 | 16 色 |

## 4. 改动清单

### 4.1 Schema 与 validator

| 文件 | 改动 |
|---|---|
| `internal/data/config/constants_domain.go` | 新增 `ColorTokenTransparent ColorToken = "transparent"` + `FallbackColorTransparent Color = "transparent"` |
| `internal/data/config/theme.go` | `DialogStyles` 加 `BodyBackground ColorRef`;`DialogStylesPatch` 加 `BodyBackground`;`Theme.ResolveColor` 加 transparent 分支;`DefaultTheme()` 改 4 个透明 + bodyBackground 默认值 |
| `internal/data/config/config.go` | `ValidateTheme` 加 `theme.Dialog.BodyBackground` 到 refs;`isColorRef` 加 `ColorTokenTransparent` 分支 |
| `internal/tui/terminal.go`(新) | `SupportsTruecolor()` 探测 `COLORTERM`/`TERM`/`TERM_PROGRAM` |
| `internal/tui/ui/component/styles_load.go` | 加 `DialogBodyBackground` rawStyles 字段 + 投影 + lookup case |
| `internal/tui/ui/widget/dialog/box.go` | `DialogBox` 改用 `component.GetStyle("dialogBodyBackground").GetBackground()` 替换 `opaqueDialogColor(overlay)` |

### 4.2 JSONC 6 主题

| 主题 | header.background | footer.statusBackground | footer.shortcutBackground | toast.background | dialog.bodyBackground |
|---|---|---|---|---|---|
| default | transparent | transparent | transparent | transparent | `#ffffff` |
| dark | transparent | transparent | transparent | transparent | `#1a1b1c` |
| light | transparent | transparent | transparent | transparent | `#ffffff` |
| nord | transparent | transparent | transparent | transparent | `#30353f` |
| dracula | transparent | transparent | transparent | transparent | `#2d2e38` |
| solarized | transparent | transparent | transparent | transparent | `#0c2c3c` |

### 4.3 测试

| 测试 | 验证 |
|---|---|
| `TestResolveColorTransparentReturnsLiteralTransparent` | `Theme.ResolveColor(TokenRef(ColorTokenTransparent))` 返回 `"transparent"` |
| `TestDialogStylesHasBodyBackground` | `DefaultTheme().Dialog.BodyBackground` 非空且 `isColorRef` 通过 |
| `TestIsColorRefAcceptsTransparent` | `isColorRef(TokenRef(ColorTokenTransparent))` 返回 true |
| `TestEmbeddedThemesCascadeFromDefault`(已存在) | 6 主题全部通过 `ValidateTheme`(含 transparent + bodyBackground) |
| `TestApplyThemeStylesProjectsFixedScopes`(已存在) | `rawStyles.DialogBodyBackground.Background` 非空 |

## 5. 验证

```text
go test ./...    # 全部通过(35+ 包)
go vet ./...     # 无诊断
JSONC 嵌套键 1:1 镜像:6 主题均 47 键
```

## 6. 已知保留项

- `tui.SupportsTruecolor()` 已实现但**未在 `tui.ApplyTheme` 中实际使用**——alpha 降级策略需要单独批次实施(用户尚未决定降级算法细节)
- `DialogStyles.BodyBackground` 默认 `TokenRef(ColorTokenBackground)`(主背景色),6 主题 JSONC 用 `value:` 覆盖为 +5% 加亮——用户可在 JSONC 改任意 hex 调整
- `theme-hardcoded-migration.md §6` 提到的 24 处 C 类启动兜底(`style.Colors` 编译期默认值 + `FallbackColor*`)仍保留——`FallbackColorTransparent` 是新增的第 16 个 fallback

## 7. 未来扩展

- 把 `SupportsTruecolor()` 接入 `tui.ApplyTheme`:non-truecolor 时自动把 9 位 hex 退化为 7 位,保留遮罩可见性
- 给 `BodyBackground` 加 JSONC 计算支持:用户可写 `token: lighten(background, 5%)` 让程序动态计算 +5% 加亮(避免 6 主题都手填 hex)
- 把 `transparent` token 扩展到所有非弹框组件(ActionBar / MessageRail / QueryBar)——目前这些由 R4/R5 代码层 `.Background(style.Colors.BG)` 兜底,JSONC 层暂未暴露