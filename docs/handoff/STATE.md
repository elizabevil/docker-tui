# STATE — task 任务驱动 / 当前活跃批次

> 上次更新: 2026-08-04
> 当前轮次: 2

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