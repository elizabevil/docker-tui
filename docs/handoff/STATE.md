# STATE — task 任务驱动 / 当前活跃批次

> 上次更新: 2026-08-04
> 当前轮次: 12

## 批次 R12:style.Colors 字段重命名为语义名

### R12.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **触发**:用户指出"token 命名是颜色名不是作用范围名"的哲学同样适用于 `style.Colors` 启动兜底层
- **范围**:
  - `internal/constants/color.go` —— 12 语义常量 + 12 legacy 别名指语义名
  - `internal/tui/ui/style/style.go` —— Palette struct 11 字段 + Colors 默认值 + SyncPalette + Color() fallback
  - `internal/tui/styles.go` —— ApplyTheme 11 行注入映射
  - `internal/tui/ui/widget/dialog/box.go` —— 1 处 (sty.Colors.Cyan → Primary)
  - `internal/tui/ui/widget/header/header.go` —— 3 处 cpuLoadColor
  - `internal/utils/color_test.go` —— 1 测试改用新名
  - 留置:`docs/handoff/proposals/theme-style-colors-renamed.md`(新)+ `GLOBAL.md` 第 11 条决策
- **主模型决策**:见 [GLOBAL.md § 2026-08-04 — R12](./GLOBAL.md#2026-08-04--r12-stylecolors-字段重命名为语义名)

### R12.2 — 子任务清单

| 子任务 | 状态 |
|---|---|
| Style 1/11: style/style.go | done |
| Style 2/11: tui/styles.go | done |
| Style 3/11: widget/header/header.go | done |
| Style 4-10/11: 7 widget reader (BG 保留) | done (0 旧引用遗留) |
| 验证: go vet / go test / 旧字段扫描 | done |
| Handoff 1-3/3: 留置文档 | done |
| **增量 1/1**: legacy 别名删除 + 3 文件 caller 迁移 | done (2026-08-04 后续) |

### R12.3 — 验证

- `go test ./...`:35 包全过
- `go vet ./...`:无诊断
- 旧 `style.Colors.{Green|Cyan|Blue|Red|Yellow|Orange|Purple|White|Gray|Dark|Surface}` 引用扫描:0 处遗留
- `constants.Color*` 原始字符串(`"green"/"cyan"/...`)扫描:0 处遗留
- **增量清理后**: legacy 别名块整块删除,所有 `ColorGreen/Cyan/Blue/...` 引用 0 处遗留

### R12.4 — 最终累计

- R12 完整版 = 6 文件(主改动) + 3 文件(增量清理) = **9 文件**
- R4-R12 总累计 = **~56 文件改动**(未 commit,等你确认)

## 批次 R11:主题 token 重命名为作用范围名(方案 C 破坏性修改)

### R11.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **触发**:用户讨论需求"token 命名是颜色名不是作用范围名"
- **范围**:
  - `internal/data/config/constants_domain.go` —— `ColorToken` enum 13 值重命名 + `FallbackColor*` 12 常量重命名
  - `internal/data/config/theme.go` —— `Palette` 12 字段 + `PalettePatch` + `ResolveColor` switch + `DefaultTheme` + `Apply`
  - `internal/data/config/config.go` —— `ValidateTheme` + `isColorRef`
  - `internal/tui/styles.go` —— `ApplyTheme` 中 palette 引用
  - `internal/tui/ui/component/table_config.go` —— `defaultTableConfig` + `defaultColumnStyles`
  - `internal/data/config/config_test.go` + `internal/data/config/theme_test.go` + `internal/tui/ui/component/styles_load_test.go` + `internal/tui/ui/component/table_test.go`
  - 6 主题 JSONC —— palette 字段名 + 所有 token 引用重命名
  - 留置:`docs/handoff/proposals/theme-semantic-tokens.md`(新)+ `GLOBAL.md` 第 10 条决策
- **主模型决策**:见 [GLOBAL.md § 2026-08-04 — R11](./GLOBAL.md#2026-08-04--r11-主题-token-重命名为作用范围名方案-c-破坏性修改)

### R11.2 — 子任务清单

| 子任务 | 状态 |
|---|---|
| Schema 1/7: ColorToken enum 重命名 | done |
| Schema 2/7: Palette/PalettePatch/ResolveColor/DefaultTheme/Apply | done |
| Schema 3/7: Validator (isColorRef) | done |
| Schema 4/7: styles_load.go / table_config.go / styles.go | done |
| JSONC 6 主题:palette 字段 + token 引用 | done |
| 测试 4 文件引用重命名 | done |
| 验证: go test / vet / JSONC 镜像 / 旧 token 扫描 | done |
| 留置文档 | done |

### R11.3 — 验证

- `go test ./...`:全部通过
- `go vet ./...`:无诊断
- JSONC 6 主题均 47 键 1:1 镜像
- 旧 token 引用扫描(`"green"|"cyan"|...`):0 处遗留

## 批次 R10:主题透明架构实施(Q1-Q4 全 a)

### R10.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **触发**:用户讨论需求"颜色配置统一 transparent + 弹框区分图层 + 终端行为统一 + 图层结构"
- **范围**:
  - `internal/data/config/constants_domain.go` —— 加 `ColorTokenTransparent` + `FallbackColorTransparent`
  - `internal/data/config/theme.go` —— `DialogStyles.BodyBackground` + `ResolveColor` transparent 分支 + `DefaultTheme` 调整
  - `internal/data/config/config.go` —— `ValidateTheme` refs 加 bodyBackground + `isColorRef` transparent 分支
  - `internal/tui/terminal.go`(新) —— `SupportsTruecolor()`
  - `internal/tui/ui/component/styles_load.go` —— `DialogBodyBackground` rawStyles + 投影 + lookup
  - `internal/tui/ui/widget/dialog/box.go` —— `DialogBox` 用 `GetStyle("dialogBodyBackground").GetBackground()`
  - 6 主题 JSONC —— 4 个主背景槽位 transparent + `dialog.bodyBackground` 主背景+5% 加亮
  - 测试:`theme_test.go` 3 新增 + `styles_load_test.go` 1 新增
  - 留置:`docs/handoff/proposals/theme-transparent-architecture.md`(新)+ `docs/handoff/GLOBAL.md` 第 9 条决策
- **主模型决策**:见 [GLOBAL.md § 2026-08-04 — 透明架构 R10](./GLOBAL.md#2026-08-04--主题透明架构实施r10q1-q4-全-a)

### R10.2 — 子任务清单

| 子任务 | 状态 | 范围 |
|---|---|---|
| Schema: ColorTokenTransparent + FallbackColorTransparent | done | constants_domain.go |
| DialogStyles.BodyBackground + ResolveColor 分支 + DefaultTheme 调整 | done | theme.go |
| isColorRef 接受 transparent + ValidateTheme refs | done | config.go |
| SupportsTruecolor() 探测 | done | terminal.go(新) |
| DialogBodyBackground rawStyles + 投影 + lookup | done | styles_load.go |
| DialogBox 用新槽位 | done | box.go |
| 6 主题 JSONC 改 4 槽位 transparent + bodyBackground | done | 6 themes |
| 测试覆盖 transparent + bodyBackground | done | theme_test + styles_load_test |
| 验证 | done | - |
| 留置文档 | done | proposals + GLOBAL + STATE |

### R10.3 — 验证

- `go test ./...`:全部通过
- `go vet ./...`:无诊断
- JSONC 6 主题均 47 键 1:1 镜像

## 批次 R9:C 类启动兜底保留 + 透明/RGBA 设计留置

### R9.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **触发**:用户讨论需求"jsonc 配置也应该支持透明，用户根据需要配置颜色（例如 RGBA）"
- **代码改动**:0(讨论后决定保留 C 类启动兜底)
- **留置**:`docs/handoff/proposals/theme-light-bg-fix.md` §12 + `docs/handoff/GLOBAL.md` 第 8 条决策
- **主模型决策**:见 [GLOBAL.md § 2026-08-04 — C 类保留+透明设计留置](./GLOBAL.md#2026-08-04--c-类启动兜底保留--透明rgba-jsonc-支持设计留置)

### R9.2 — 子任务清单

| 子任务 | 状态 | 范围 |
|---|---|---|
| 盘点 C 类 24 处 hex 字面量 | done | style.go + constants_domain.go |
| 决策:保留启动兜底不动 | done | (无文件改动) |
| 透明/RGBA JSONC 支持设计 | done | proposal §12 |
| 留置文档 | done | proposals + GLOBAL + STATE |

### R9.3 — 验证

- `go test ./...`:全部通过
- `go vet ./...`:无诊断
- JSONC 6 主题均 47 键 1:1 镜像
- 代码改动:0

## 批次 R8:架构级重设计 — "仅主背景"架构

### R8.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **范围**:6 主题 JSONC 改 4 个主背景槽位为 `token: background`(共 24 处)
- **触发**:用户截图反馈 light 主题 header/footer 仍画独立背景
- **主模型决策**:见 [GLOBAL.md § 2026-08-04 — 架构级重设计](./GLOBAL.md#2026-08-04--架构级重设计仅主背景架构用户截图反馈)
- **不可触碰边界**:Go 代码、palette.dark/surface 槽位、dialog.overlay、强调组件背景

### R8.2 — 子任务清单

| 子任务 | 状态 | 范围文件 |
|---|---|---|
| 6 主题 JSONC 改 Header.Background | done | 6 themes |
| 6 主题 JSONC 改 Footer.ShortcutBackground | done | 6 themes |
| 6 主题 JSONC 改 Footer.StatusBackground | done | 6 themes |
| 6 主题 JSONC 改 Toast.Background | done | 6 themes |
| 验证 | done | - |
| 留置文档 | done | proposals + GLOBAL + STATE |

### R8.3 — 验证

- `go test ./...`:全部通过
- `go vet ./...`:无诊断
- JSONC 6 主题均 47 键 1:1 镜像

## 批次 R7:B 类全量重构(dialog 30 处显式化)

### R7.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **范围**:
  - `internal/tui/ui/component/styles_load.go` —— +3 字段/投影/lookup
  - `internal/tui/ui/widget/dialog/{exec,form,form_popup,choice,selection,notification,view}.go` —— 30 处替换 + import 清理
  - 留置:`docs/handoff/proposals/theme-light-bg-fix.md` §10 + `docs/handoff/GLOBAL.md` 第 6 条决策
- **不可触碰边界**:JSONC、Go 业务逻辑、C 类(编译期 hex)
- **主模型决策**:见 [GLOBAL.md § 2026-08-04 — Light R7](./GLOBAL.md#2026-08-04--light-主题-b-类全量重构dialog-30-处显式化)

### R7.2 — 子任务清单

| 子任务 | 状态 | 范围文件 | 提交者 |
|---|---|---|---|
| 新增 3 个 dialog token 槽位 | done | styles_load.go | orchestrator |
| dialog 30 处 style.Colors → GetStyle | done | 7 dialog 文件 | orchestrator |
| 清掉 dialog 包无用 style import | done | 7 文件 | orchestrator |
| 留置文档 | done | proposals + GLOBAL | orchestrator |

### R7.3 — 验证

- `go test ./...`:全部通过
- `go vet ./...`:无诊断
- JSONC 6 主题均 47 键 1:1 镜像
- dialog 包内 `style.Colors.*` 使用数:30 → 0

## 批次 R6:Light 主题 B 类修复(对比度差 13 处)

### R6.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **范围**:
  - `internal/tui/ui/widget/header/header.go` —— keyBadge/keystroke border 2 处
  - `internal/tui/ui/widget/dialog/form_popup.go` —— highlight 背景 2 处
  - `internal/tui/ui/widget/dialog/{choice,exec,selection,form}.go` —— unselected 按钮前景 7 处
  - 留置:`docs/handoff/proposals/theme-light-bg-fix.md` §8/9 增量 + `docs/handoff/GLOBAL.md` 第 5 条决策
- **不可触碰边界**:JSONC、schema、Go 业务逻辑、其它跟着调色板 OK 的 21 处
- **主模型决策**:见 [GLOBAL.md § 2026-08-04 — Light R6](./GLOBAL.md#2026-08-04--light-主题-b-类修复对比度差-13-处)

### R6.2 — 子任务清单

| 子任务 | 状态 | 范围文件 | 提交者 |
|---|---|---|---|
| 精确对比度盘点 | done | (无文件改动) | orchestrator |
| header keyBadge/keystroke border 2 处 | done | header.go | orchestrator |
| form_popup highlight 背景 2 处 | done | form_popup.go | orchestrator |
| dialog 按钮 unselected 7 处 | done | 4 文件 | orchestrator |
| 留置文档 | done | proposals + GLOBAL | orchestrator |

### R6.3 — 验证

- `go test ./...`:全部通过
- `go vet ./...`:无诊断

## 批次 R5:Light 主题背景 R4 增量(A 类全容器补背景)

### R5.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **范围**:
  - `internal/tui/ui/app/layout.go` —— 2 处加背景 + import style
  - `internal/tui/ui/widget/panel/panel.go` —— 1 处加背景 + import style
  - `internal/tui/ui/pages/compose/view.go` —— 3 处加背景
  - `internal/tui/ui/pages/help/view.go` —— 2 处加背景 + import style
  - `internal/tui/ui/pages/logs/view.go` —— 1 处加背景 + import style
  - `internal/tui/ui/pages/detail/view.go` —— 1 处加背景 + import style
  - 留置:`docs/handoff/proposals/theme-light-bg-fix.md` §5/6/7 增量 + `docs/handoff/GLOBAL.md` 第 4 条决策
- **不可触碰边界**:JSONC、schema、Go 业务逻辑、R4 已修的 styles.go / actionbar.go
- **主模型决策**:见 [GLOBAL.md § 2026-08-04 — Light R5](./GLOBAL.md#2026-08-04--light-主题背景一致性-r5a-类全容器补背景)
- **未来扩展点**(留待后续批次):B 类(48 处 `style.Colors.X` 细节) + C 类(12 处 theme-hardcoded-migration §6 保留项) + R5 未覆盖的内部小容器(header logo 列、keystroke 列、query rail)

### R5.2 — 子任务清单

| 子任务 | 状态 | 范围文件 | 提交者 |
|---|---|---|---|
| 根因分析(3 类源头) | done | (无文件改动) | orchestrator |
| A 类 11 处补背景 | done | 6 文件 | orchestrator |
| 留置文档 | done | proposals + GLOBAL | orchestrator |

### R5.3 — 验证

- `go test ./...`:全部通过
- `go vet ./...`:无诊断
- JSONC 6 主题均 47 键 1:1 镜像

## 批次 R4:Light 主题背景一致性修复

### R4.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **范围**:
  - `internal/tui/styles.go` —— 3 个 BorderStyle 加背景
  - `internal/tui/ui/widget/actionbar/actionbar.go` —— 容器加背景 + import style
  - 留置:`docs/handoff/proposals/theme-light-bg-fix.md` + `docs/handoff/GLOBAL.md` 第 3 条决策
- **不可触碰边界**:JSONC、schema、Go 业务逻辑、其它 widget
- **主模型决策**:见 [GLOBAL.md § 2026-08-04 — Light 背景](./GLOBAL.md#2026-08-04--light-主题背景一致性修复)

### R4.2 — 子任务清单

| 子任务 | 状态 | 范围文件 | 提交者 |
|---|---|---|---|
| 根因分析 | done | (无文件改动) | orchestrator |
| BorderStyle 背景修复 | done | styles.go | orchestrator |
| actionbar 容器背景修复 | done | actionbar.go | orchestrator |
| 留置文档 | done | proposals/theme-light-bg-fix.md + GLOBAL | orchestrator |

### R4.3 — 验证

- `go test ./...`:全部通过
- `go vet ./...`:无诊断
- JSONC 6 主题均 47 键 1:1 镜像

## 批次 R3:主题硬编码迁移 Follow-up(shortcutBar + light)

### R3.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **范围**:
  - `internal/tui/ui/component/styles_load.go` —— 补 shortcutBar 投影 3 处
  - `internal/tui/ui/component/styles_load_test.go` —— 增 1 项 ShortcutBar 背景断言
  - `internal/data/config/themes/light.jsonc` —— table 3 键 hex 重调
  - 留置:`docs/handoff/proposals/theme-hardcoded-followups.md` + `docs/handoff/GLOBAL.md` 第 2 条决策
- **不可触碰边界**:其余 5 主题 JSONC、`default.jsonc`、Go 业务逻辑、schema
- **主模型决策**:见 [GLOBAL.md § 2026-08-04 — Follow-up](./GLOBAL.md#2026-08-04--主题硬编码迁移-follow-upshortcutbar-投影--light-表色)
- **未来扩展点**(留待后续批次):见 [proposals/theme-hardcoded-followups.md § 5](./proposals/theme-hardcoded-followups.md#5-已知保留项) — 主要是 `theme-hardcoded-migration.md §7.4` 新增主题接入指南 + `Footer.StatusBackground` 调用方

### R3.2 — 子任务清单

| 子任务 | 状态 | 范围文件 | 提交者 |
|---|---|---|---|
| shortcutBar 投影 3 处 | done | styles_load.go | orchestrator |
| ShortcutBar 测试断言 | done | styles_load_test.go | orchestrator |
| light 表色 3 hex 重调 | done | light.jsonc | orchestrator |
| Footer.StatusBackground 调查 | done(无调用点) | grep scan | orchestrator |
| 留置文档 | done | proposals/theme-hardcoded-followups.md + GLOBAL + STATE | orchestrator |

### R3.3 — 验证

- `go test ./...`:全部通过
- `go vet ./...`:无诊断
- JSONC 嵌套键 1:1 镜像:6 主题均 47 键

## 批次 R2:主题硬编码色全面迁移(13 处 → 配置)

### R2.1 — 批次元信息

- **状态**:done(未 commit,等用户确认)
- **日期**:2026-08-04
- **范围**:
  - schema 扩展:`internal/data/config/{theme.go, constants_domain.go}`
  - 投影重写:`internal/tui/ui/component/{styles_load.go, table_config.go}`
  - 主题文件:`internal/data/config/themes/{default, dark, light, nord, dracula, solarized}.jsonc`
  - 硬编码替换:`internal/tui/ui/widget/dialog/{form, notification, selection, exec}.go` + `internal/tui/ui/app/layout.go`
  - 测试:`internal/tui/ui/component/{styles_load_test, table_test}.go`
  - 留置文档:`docs/handoff/proposals/theme-hardcoded-migration.md` + `docs/handoff/GLOBAL.md` 本日决策日志
- **不可触碰边界**:Go 业务逻辑 / `internal/data/i18n/lang/*` / `internal/data/config/default.jsonc` 以外的所有 config 文件 / `docs/requirement/*` 与 `docs/constraint/*` 真理源
- **主模型决策**:见 [GLOBAL.md § 2026-08-04](./GLOBAL.md#2026-08-04--主题硬编码色全面迁移至配置13-处)
- **未来扩展点**(留待后续批次):见 [proposals/theme-hardcoded-migration.md § 7](./proposals/theme-hardcoded-migration.md#7-未来扩展点)

### R2.2 — 子任务清单

| 子任务 | 状态 | 范围文件 | 提交者 |
|---|---|---|---|
| 盘点硬编码清单(13 处) | done | (无文件改动,2 个 explore agent 报告) | 主模型 |
| 设计 schema 扩展 | done | `theme.go` 加 `TableStyles` / `SafeFallbackStyles` | 主模型 |
| 主题文件 1:1 镜像 | done | 6 份 JSONC 各加 8 键 | 主模型 |
| 投影重写 | done | `styles_load.go` + `table_config.go` | 主模型 |
| 硬编码替换 | done | 4 dialog + 1 layout | 主模型 |
| 测试更新 | done | 2 个 test 文件 | 主模型 |
| 留置文档 | done | proposals/ + GLOBAL/ + STATE | 主模型 |

### R2.3 — 验证

- `go test ./...`:全部通过
- `go vet ./...`:无诊断
- JSONC 嵌套键 1:1 镜像:6 主题均 47 键
- 硬编码字面量扫描:13 处全部消失

## 批次 R1:文档真理源重组与清理

### R1.1 — 批次元信息

- **状态**:done + 已 commit(主模型审查完成,2026-08-03)
- **日期**:2026-08-02 ~ 2026-08-03
- **范围**:`docs/requirement/`(新建)、`docs/constraint/`(新建)、`docs/docs-architecture.md`(新建)、`docs/README.md` / `docs/requirements.md` / `docs/task/README.md`(更新入口)
- **不可触碰边界**:Go 代码、`internal/data/i18n/lang/*`、现有 task/ 文档(除非明确删除)
- **主模型决策**:见 [GLOBAL.md § 2026-08-02](./GLOBAL.md#2026-08-02--文档真理源重组)

### R1.2 — 子任务清单

| 子任务 | 状态 | 范围文件 | 提交者 |
|---|---|---|---|
| §8.1 目录创建 | done | `mkdir docs/requirement/R0{1..5} docs/constraint` | 小模型 |
| §8.2 README 索引(7 个) | done | `requirement/README.md` + 5 R## README + `constraint/README.md` | 小模型 |
| §8.3 constraint 文档(6 个) | done | `constraint/C0{1..6}.md` | 小模型 |
| §8.4 requirement 文档(11 个) | done | `requirement/R##-##/*.md` | 小模型 |
| §8.5 入口索引更新(3 个) | done | `docs/README.md` / `docs/requirements.md` / `docs/task/README.md` | 小模型 |
| Findings 修复(链接 + 事实) | done | R## README + R01-03 + docs/requirements.md:14 + 6 个 constraint | 小模型 |
| 6 个冗余 task 删除 | done | `rm` 6 文件 + 13 文件反向引用清理 | 小模型 |

### R1.3 — 验证

- `go build ./...`:未运行(纯文档操作,不应影响 Go 代码)
- `go test -short ./...`:未运行(同上)
- `git diff --check`:通过
- 主模型审查(2026-08-03)补充修复:
  - 子模型声明"残留引用 0"不实 — 实查发现 5 处指向已删 task 的悬空引用(`docs/task/README.md` ×5、br-039-a、br-039-code-location ×2、br-040-code-location ×2、br-041 ×1),已全部修复
  - 跨组断链 13 处:`R03-form-action/` 路径缺失 ×5、`docs/feature-design.md` 前缀 ×6、R04 `internal/data/runtime/` 相对深度 ×2,已修复
  - 全仓 markdown 链接校验:479 个内部链接全部可解析(109 个 `internal/*` 为 bugfix 文档预存代码锚点约定,非断链)
- 旧 task 文件保留:21 个(去 6 后)

### R1.4 — 主模型审查结论

- 全部新增 30 个 + 修改 13 个 + 删除 6 个文档已 review 并 commit
- 审查发现与修复记录:见上 R1.3;决策记录见 [GLOBAL.md § 2026-08-03 R1 批次审查与 commit](./GLOBAL.md#2026-08-03--r1-批次主模型审查与-commit)
- R04-01 状态 `implementing` 偏乐观,待 R4 批次实施时再校准(不阻塞 R1)

## 批次 R2:BR-043 Form / Action Bar 优化

### R2.1 — 批次元信息

- **状态**:done + 已 commit(主模型决策 + 实施完成,2026-08-03)
- **日期**:2026-08-03
- **范围**:`internal/tui/keyboard/container_form.go` + `internal/tui/ui/widget/dialog/{view,form,exec,selection}.go` + `internal/tui/state/form.go` + `internal/tui/keys/{commands,display}.go`
- **不可触碰边界**:`internal/data/i18n/lang/*`、`internal/data/config/*`、`internal/tui/keyboard` 之外的键盘处理
- **主模型决策**:见 [GLOBAL.md § 2026-08-03 BR-043 六个问题决策](./GLOBAL.md#2026-08-03--br-043-六个问题决策)

### R2.2 — 子任务清单

| 子任务 | 状态 | 范围文件 | commit |
|---|---|---|---|
| 批次 A:3.1 路径绝对化(blur-time) | done | `keyboard/container_form.go` + `state/path.go`(只读) | `1731789` |
| 批次 A:3.2 面板尺寸统一(3/4 宽 + 等高) | done | `dialog/{view,form,exec,selection}.go` + `dialog/form_test.go` | `1731789` |
| 批次 B:3.5 Commit form 条件字段 | done | `state/form.go` + `state/form_test.go` + `keyboard/container_form.go` + `dialog/form.go` | `8ec0e4d` |
| 批次 C:3.3 Action 页面布局重构(方案 B) | done | `state/form.go` + `keyboard/container_form.go` + `dialog/form.go` + 多个测试 | `9e0078f` |
| 批次 D:3.6 C 键连接信息迁 command palette | done | `keyboard/keyboard.go` + `keyboard/command.go` + `keys/{commands,display}.go` | `29d8cfe` |

### R2.3 — 验证

- `go vet ./internal/tui/state/... ./internal/tui/keyboard/... ./internal/tui/keys/... ./internal/tui/ui/widget/dialog/...`:通过
- `go test` 上述包:全过(state/keyboard/keys/dialog + 全包测试 cached)
- `git diff --check`:通过
- 真实断链扫描:0(链接校验 479 个内部链接全部可解析)
- 旧 task 文件保留:21 个(去 6 后)

### R2.4 — 未实施项(后续 followup)

- **3.4 配置拆分**:不本轮,后续另起任务卡(`internal/ui/styles/*.jsonc` 多文件 + 用户目录覆盖)
- **per-form DefaultFocus 配置**:本轮不实施(默认 Cancel 已满足 BR-043 §7.3)
- **DependsValue 枚举扩展**:本轮仅支持 bool(DependsEq bool),枚举/字符串比较留待 `DependsValue any` 扩展
- **3.6 R 键归宿**:已决策保留为全局 refresh,无代码变更

## 批次 R3:待启动

### R3.1 — 候选

- `docs/pending-bugs.md` 18 项 open BUG(详见 pending-bugs.md)
- 3.4 配置拆分(若主模型决定启动)

## 公共范围边界(所有子任务不可触碰)

- `internal/**/*.go`
- `internal/data/i18n/lang/*.jsonc`
- `internal/data/config/*.jsonc`
- 现有 task/ 文档(除非在 STATE.md 显式列入本批次范围)

## 全局链接

- [../docs-architecture.md](../docs-architecture.md) — 文档真理源架构
- [../requirement/](../requirement/) — 需求真理源
- [../constraint/](../constraint/) — 横切约束
- [./GLOBAL.md](./GLOBAL.md) — 全局决策日志
- [./COMPLETED.md](./COMPLETED.md) — 已交接历史