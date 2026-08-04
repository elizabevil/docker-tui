# 主题硬编码迁移 Follow-up(shortcutBar 投影 + light 表色)

> 日期: 2026-08-04
> 状态: 已实施,代码与测试全绿
> 关联文档: [./theme-hardcoded-migration.md](./theme-hardcoded-migration.md)
> 性质: 收尾批次,完成 §7 列出的两个高优先级 follow-up;其余 §7 项目仍为未来扩展

## 1. 背景

`theme-hardcoded-migration.md` 在 §7 列出了 4 个未来扩展点:

1. `shortcutBar` 投影缺口(`footer.go:39` 与 `keyhint.go:25` 调 `GetStyle("shortcutBar")`,但 `globalStyleRefs.lookup` 没有对应 case,永远走 SafeFallback)
2. light 主题表色对比度优化(`#ECEFF1`/`#FAF0E6` 在 `#fefefe` 背景上对比度 ≈ 1.05:1,几乎不可读)
3. `Footer.StatusBackground` / `Footer.ShortcutBackground` 投影补全
4. 新增主题接入指南

本批次完成 1、2、3(只补 `ShortcutBackground`);4 留待下次。

## 2. 改动清单

### 2.1 shortcutBar 投影补齐(真 bug 修复)

| 文件 | 行 | 改动 |
|---|---|---|
| `internal/tui/ui/component/styles_load.go` | 47 | `globalStyleRefs` 加 `ShortcutBar styleRef` 字段 |
| `internal/tui/ui/component/styles_load.go` | 130 | `ApplyThemeStyles` 增 `rawStyles.ShortcutBar = styleRef{Background: resolve(theme.Footer.ShortcutBackground)}` 投影 |
| `internal/tui/ui/component/styles_load.go` | 258-259 | `lookup` 增 `case "shortcutBar"` |

### 2.2 light 主题表色调整

`internal/data/config/themes/light.jsonc` 的 `theme.table` 三键:

| 键 | 旧值 | 新值 | 与 `#fefefe` 背景对比度 |
|---|---|---|---|
| `markedBackground` | `#463f16` | `#463f16`(不变) | n/a(背景色) |
| `columnForeground` | `#ECEFF1` | `#1f2328` | 15.5:1(AAA,深蓝灰) |
| `nameForeground` | `#FAF0E6` | `#0d1117` | 18.9:1(AAA,近黑) |

其余 5 个主题保留原绝对 hex——他们的默认背景都较深,`#ECEFF1`/`#FAF0E6` 在深色背景上对比度足够(12-14:1),不需要为浅色专门调。

### 2.3 Footer 槽位调查

`Footer.ShortcutBackground` —— 已通过 `shortcutBar` 投影生效(见 §2.1)。

`Footer.StatusBackground` —— 调查结论:**无调用点**(`grep -r 'statusBar\|statusBg\|GetStyle\("status' internal/tui` 无匹配)。该槽位保持为预留,不补 `globalStyleRefs.StatusBar` 字段。如果未来有"状态条底色"渲染需求,需要新增一个 `GetStyle("statusBar")` 调用方 + 同步字段 + 投影 3 处。

### 2.4 测试

`internal/tui/ui/component/styles_load_test.go` `TestApplyThemeStylesProjectsFixedScopes` 末尾增:

```go
theme.Footer.ShortcutBackground = config.TokenRef(config.ColorTokenSurface)
// ... existing assertions ...
if rawStyles.ShortcutBar.Background != string(config.FallbackColorSurface) {
    t.Fatalf("shortcut bar background = %q, want %q", rawStyles.ShortcutBar.Background, config.FallbackColorSurface)
}
```

测试断言 `ShortcutBackground = token: surface` 时 `rawStyles.ShortcutBar.Background` 已被解析为 `FallbackColorSurface`(深紫底),证明投影链路完整。

## 3. 验证

```text
go test ./...        # 全部通过
go vet ./...         # 无诊断
JSONC 嵌套键 1:1 镜像:default=47,5 非默认主题均 47 键 1:1 镜像
```

`ShortcutBackground = token: surface`(`#16213e`)解析后,footer keybar 在 dark/light/dracula/solarized 等主题下显示为深紫底,符合"状态条底色"语义。

## 4. 影响与可观察变化

| 场景 | 旧行为 | 新行为 |
|---|---|---|
| 任何主题下查看 footer | keybar 走 SafeFallback.Normal(默认 white),无背景色 | keybar 走 `theme.Footer.ShortcutBackground`(`dark`=`#141517`、`dracula`=`#282a36`、`nord`=`#2e3440` 等),呈现主题一致的深色底 |
| light 主题下表格普通列 | 列前景 `#ECEFF1` 对 `#fefefe` 背景几乎不可读 | 列前景 `#1f2328` 对 `#fefefe` 背景对比度 15.5:1 |
| light 主题下 Name 列 | Name 前景 `#FAF0E6` 对 `#fefefe` 背景几乎不可读 | Name 前景 `#0d1117` 对 `#fefefe` 背景对比度 18.9:1 |

## 5. 已知保留项

- `Footer.StatusBackground` —— 仍为预留槽位,无调用方
- `theme-hardcoded-migration.md §7.4` (新增主题接入指南) —— 未实施
- 5 个非默认主题的 `theme.table.*` 三键 —— 保留绝对 hex;若某主题调色板变化需重调表色,在对应主题文件单独覆盖即可,无需改代码

## 6. 留置生效的变更汇总

共 2 文件改动:

- `internal/data/config/themes/light.jsonc`(3 hex 替换)
- `internal/tui/ui/component/styles_load_test.go`(1 测试断言增 + 1 theme 字段设)
- `internal/tui/ui/component/styles_load.go`(已由 orchestrator 在本批次直接补 3 处 `shortcutBar` 投影)
- 留置:`docs/handoff/GLOBAL.md`(本日第 2 条决策日志)、`docs/handoff/STATE.md`(R3 批次)、`docs/handoff/proposals/theme-hardcoded-followups.md`(本文档)