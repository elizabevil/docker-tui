# dtui Design Documents

当前设计文档:

| 文件                  | 内容                              |
|---------------------|---------------------------------|
| `current-design.md` | **唯一设计参考** — 布局/组件/表格/搜索/快捷键/样式 |
| `ui-i18n-design.md` | UI 国际化、翻译资源、终端显示宽度和中英文布局约束 |
| `bugfix-design.md` | 历史 bug 修复设计、分层策略、变更约束 |
| `future-requirements-discussion.md` | 后续需求与规划讨论纪要，记录候选方案和待确认决策 |
| `future-requirements-task-list.md` | 后续需求实施任务、依赖、状态和完成定义 |
| `podman-rest-migration.md` | `TASK-022` Podman REST 适配收紧：gpgme 隔离 + 匿名 struct 清零 + 验证矩阵 |
| `podman-capabilities-analysis.md` | Docker / Podman 能力差异对比分析 |
| `unified-runtime-driver.md` | `TASK-021` Docker / Podman 统一 runtime driver 实施设计 |

已完成的设计文档已移至 `archived/`，包括：

- `app-model-refactoring.md` — `TASK-015` AppModel 状态域拆分
- `page-performance-optimization-design.md` — 页面布局优化设计
- `page-performance-approval-proposals.md` — 页面与性能优化审批方案
- `page-performance-execution-plan.md` — 页面与性能优化执行计划
- `page-performance-task-sheet.md` — 页面与性能优化任务单
- `page-performance-regression.md` — 页面与性能优化回归记录
- `bugfix-discussion-notes.md` — 历史 bug 修复讨论纪要

说明:

- `current-design.md` 仍是当前 UI / 交互的唯一权威参考。
- `bugfix-design.md` 仅用于历史 bug 修复工作流，不替代 UI 设计主文档。
- `future-requirements-discussion.md` 用于后续需求讨论；标为 `approved` 前不代表正式设计或实现承诺。
- `future-requirements-task-list.md` 是后续实施的唯一主任务台账；其他设计文档中的任务编号仅作为分析来源，任务完成必须经过测试和文档同步。
- `podman-rest-migration.md` 用于规划 `runtime/podman` 的方法式 + dto-only 签名 + 驱动/REST 双形态改造。
- `podman-capabilities-analysis.md` 用于 Docker / Podman 能力差异对比，为 TASK-020 等后续任务提供决策依据。

## 参考资源

| 路径                 | 内容                |
|--------------------|-------------------|
| `internal/tui/ui/` | Go 代码实现           |
| `docs/ui-preview/` | 浏览器预览系统 (Vue 3)  |
| `docs/`            | 当前维护中的实现文档        |
| `docs/ui-design.md`| UI 设计系统开发者参考（提取自 `current-design.md`） |
| `docs/i18n.md`     | 国际化设计开发者参考（提取自 `ui-i18n-design.md`） |
