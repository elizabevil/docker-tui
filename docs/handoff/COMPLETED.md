# COMPLETED — task 任务驱动 / 已交接历史

> 上次更新: 2026-08-03
> 排序: 时间倒序

## 批次 R1:文档真理源重组与清理

- **完成日期**:2026-08-03
- **子模型产出**:见 [STATE.md R1.2 子任务清单](./STATE.md#r12--子任务清单)
- **主模型审查**:完成(2026-08-03)并已 commit;审查发现 5 处悬空引用 + 13 处断链,已全部修复(详见 [STATE.md R1.3](./STATE.md#r13--验证))
- **变更统计**:新建 30 个 + 修改 13 个 + 删除 6 个 = 49 个文件

### 子任务交付细节

| 子任务 | 交付 | 验证 |
|---|---|---|
| §8.1 目录创建 | `mkdir docs/requirement/R0{1..5} docs/constraint` | `ls docs/requirement docs/constraint` 6 个目录 |
| §8.2 README 索引 | 7 个 README.md(requirement + 5 大需求 + constraint) | `wc -l` 符合预期 |
| §8.3 constraint 文档 | C01-form / C02-dialog / C03-table / C04-keybinding / C05-path / C06-i18n | `grep "## 适用范围" docs/constraint/*.md` 6/6 |
| §8.4 requirement 文档 | R01-01..R01-03 / R02-01..R02-03 / R03-01..R03-03 / R04-01 / R05-01(11 个) | `ls docs/requirement/R##-##/*.md` 11/11 |
| §8.5 入口索引 | docs/README.md + docs/requirements.md + docs/task/README.md | `git diff --check` 通过 |
| Findings 修复 | 相对链接 `../task/` → `../../task/` × 5 个 R## README;R01-03 事实修正;requirements.md:14 状态修正;6 constraint 章节重构 | `grep "../task/\|../constraint/" docs/requirement/R##/README.md` 零残留 |
| 6 个 task 删除 | next-small-model-brief + br-042 + br-036 + br-037 + br-038 + br-040-dialog-style | `git diff --check` 通过 |

### 主模型审查重点(2026-08-03 已逐项处理)

1. 状态字段逐一 review:R04-01 `implementing` 偏乐观,留待 R4 批次校准;其余维持
2. constraint 当前实现 / 目标规范划分:审查通过,无改动
3. 删除的 6 个 task 反向引用:5 处悬空引用已修复为归档链接(见 STATE.md R1.3)
4. `docs-architecture.md` §7 映射表:审查通过,无改动

### 已归档

- 批次 R1 的 6 个删除文件:
  - `task/next-small-model-brief.md`
  - `task/br-042-form-runtime-refresh-corrections.md`
  - `task/br-036-image-import-tarball.md`
  - `task/br-037-registry-login.md`
  - `task/br-038-exec-shell-ui.md`
  - `task/br-040-dialog-style-center-on-panel.md`

### 引用

- 主模型决策:[GLOBAL.md § 2026-08-03 删除 6 个冗余 task 文件](./GLOBAL.md#2026-08-03--删除-6-个冗余-task-文件)
- 架构方案:[docs-architecture.md §7 旧文档迁移映射](../docs-architecture.md#7-旧文档迁移映射)
- 当前状态:[STATE.md R1.3 验证](./STATE.md#r13--验证) + [STATE.md R1.4 待主模型审查](./STATE.md#r14--待主模型审查)

## 历史批次

### 批次 R2:BR-043 Form / Action Bar 优化

- **完成日期**:2026-08-03
- **范围**:BR-043 §3.1/3.2/3.3/3.5/3.6 五项;3.4 配置拆分 + per-form DefaultFocus + DependsValue 枚举扩展三项 followup 不在本轮
- **主模型决策**:BR-043 §7 五项待确认项(见 [GLOBAL.md § 2026-08-03](./GLOBAL.md#2026-08-03--br-043-六个问题决策))
- **变更统计**:4 个 feat commits,12 个文件改动(+/- ~440 行)

### 子任务交付

| 批次 | commit | 范围 |
|---|---|---|
| A 路径绝对化 + 面板尺寸 | `1731789` | `keyboard/container_form.go`(blur-time `state.Absolute`)+ `dialog/{view,form,exec,selection}.go`(panel 3/4 宽 + 等高)+ `dialog/form_test.go` |
| B Commit form 条件字段 | `8ec0e4d` | `state/form.go`(DependsOn/DependsEq/Hidden + RecomputeVisibility)+ `keyboard/container_form.go`(archivePath 条件可见)+ `dialog/form.go`(跳过 Hidden) |
| C Action 页面布局重构(方案 B) | `9e0078f` | `state/form.go`(FieldFocus 线性化,移除 OnConfirm)+ `keyboard/container_form.go`(Enter 按 Confirm/Cancel 分支)+ `dialog/form.go`(Confirm/Cancel 列表行渲染)+ 多个测试 |
| D C 键迁 command palette | `29d8cfe` | `keyboard/keyboard.go`(移除 KeyC fallback)+ `keyboard/command.go`(conn-info 分发)+ `keys/commands.go`(CommandConnInfo)+ `keys/display.go`(清理未用 ActionLabelConn) |

### 引用

- 设计来源:[task/br-043-form-action-bar-redesign.md](../task/br-043-form-action-bar-redesign.md)
- 决策记录:[GLOBAL.md § 2026-08-03 BR-043 六个问题决策](./GLOBAL.md#2026-08-03--br-043-六个问题决策)
- 完成时状态:[STATE.md R2.3 验证](./STATE.md#r23--验证)

(暂无 R0 及之前的归档批次。若需要,可从 git log 提取)