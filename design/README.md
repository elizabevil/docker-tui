# dtui Design Documents

当前设计文档:

| 文件                  | 内容                              |
|---------------------|---------------------------------|
| `current-design.md` | **唯一设计参考** — 布局/组件/表格/搜索/快捷键/样式 |
| `bugfix-design.md` | 历史 bug 修复设计、分层策略、变更约束 |
| `bugfix-discussion-notes.md` | 历史 bug 修复讨论纪要，保存尚未完全收敛的设计背景 |
| `future-requirements-discussion.md` | 后续需求与规划讨论纪要，记录候选方案和待确认决策 |
| `page-performance-optimization-design.md` | 页面布局优化、页面模板统一与渲染性能优化设计 |
| `page-performance-approval-proposals.md` | 页面与性能优化审批方案集，供逐项确认方向 |
| `page-performance-execution-plan.md` | 页面与性能优化执行计划，拆分阶段、依赖与验收边界 |
| `page-performance-task-sheet.md` | 页面与性能优化任务单，细化任务、回归矩阵与完成定义 |
| `page-performance-regression.md` | 页面与性能优化轨道映射、回归模板与阶段验证记录 |

旧版 `master.md`, `design-system.md`, `ui/*`, `features/*` 等文件内容已合并至 `current-design.md`，不再单独维护。

说明:

- `current-design.md` 仍是当前 UI / 交互的唯一权威参考。
- `bugfix-design.md` 仅用于历史 bug 修复工作流，不替代 UI 设计主文档。
- `bugfix-discussion-notes.md` 用于保留阶段性讨论，不替代正式设计决策。
- `future-requirements-discussion.md` 用于后续需求讨论；标为 `approved` 前不代表正式设计或实现承诺。
- `page-performance-optimization-design.md` 用于记录未来页面与性能优化方案，不代表当前已经实现。
- `page-performance-approval-proposals.md` 用于逐项审批页面与性能优化方向，不代表已经进入实现。
- `page-performance-execution-plan.md` 用于将已审批方向拆成实际执行阶段，不代表相关阶段已经开始。
- `page-performance-task-sheet.md` 用于继续细化阶段内任务与回归任务单，并记录当前执行状态。
- `page-performance-regression.md` 用于保存固定回归格式和每个执行批次的验证结果。

## 参考资源

| 路径                 | 内容                |
|--------------------|-------------------|
| `internal/tui/ui/` | Go 代码实现           |
| `docs/ui-preview/` | 浏览器预览系统 (Vue 3)  |
| `docs/`            | 当前维护中的实现文档        |
