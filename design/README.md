# dtui Design Documents

设计意图与历史变更的权威档案。当前维护中的设计文档按 **UI 设计** 与 **功能/适配设计** 分类。

## 当前设计文档

### UI 设计

| 文件 | 内容 |
|---|---|
| `current-design.md` | **唯一 UI 权威参考** — 布局 / 组件 / 表格 / 搜索 / 快捷键 / 样式 |
| `ui-i18n-design.md` | UI 国际化、翻译资源、终端显示宽度与中英文布局约束 |

### 功能与适配设计

| 文件 | 内容 |
|---|---|
| `podman-capabilities-analysis.md` | Docker / Podman 能力差异对比,为 Podman 独有 / Docker 独有的产品边界决策提供依据 |

### 修复流程

| 文件 | 内容 |
|---|---|
| `bugfix-design.md` | 历史 bug 修复的分层策略、变更约束、验收关注点;**不替代 UI 设计主文档** |

> 实现进度与设计目标见 [docs/](../docs/):`feature-design.md`(功能设计总览)、`feature-todo-list.md`(待办清单)、`bugfix-requirements.md`(BUG 台账)、`pending-bugs.md`(BUG 待办)。

## 已归档(不再修改)

参见 [archived/](archived/) 目录,包含已完成的实施设计与过期讨论:

- `unified-runtime-driver.md` — TASK-021 实施设计(已完成)
- `podman-rest-migration.md` — TASK-022 适配方案(部分 phase 已 cancelled,最终范围见 [docs/feature-todo-list.md §6.1](../docs/feature-todo-list.md))
- `page-performance-*.md` — 页面性能优化(已完成)
- `app-model-refactoring.md` — TASK-015 AppModel 状态域拆分(已完成)
- `bugfix-discussion-notes.md` — 历史 bug 修复讨论纪要

## 文档分类与边界

| 类别 | 文档 | 描述 |
|---|---|---|
| **UI 权威** | `current-design.md` / `ui-i18n-design.md` | UI / 交互 / 视觉的设计事实 |
| **功能设计** | `podman-capabilities-analysis.md` + `docs/feature-design.md` | 能力目标、对比矩阵、设计决策、不在范围 |
| **修复流程** | `bugfix-design.md` + `docs/bugfix-requirements.md` + `docs/pending-bugs.md` | bug 修复分层策略、台账、待办快照 |
| **任务台账** | `docs/feature-todo-list.md` | TASK 历史 + 设计驱动需求 |
| **实现参考** | `docs/architecture.md` / `docs/navigation.md` / `docs/project-structure.md` | 当前实现架构、交互、目录 |
| **用户参考** | `docs/ui-design.md` / `docs/i18n.md` | UI 设计系统 / i18n 的开发者参考(从 `current-design.md` / `ui-i18n-design.md` 提取) |

## 参考资源

| 路径 | 内容 |
|---|---|
| `internal/tui/ui/` | Go 代码实现 |
| `docs/` | 当前维护中的实现文档与用户参考 |
| `docs/ui-design.md` | UI 设计系统开发者参考(提取自 `current-design.md`) |
| `docs/i18n.md` | 国际化设计开发者参考(提取自 `ui-i18n-design.md`) |
| `docs/architecture.md` | 实现架构说明 |