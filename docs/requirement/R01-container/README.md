# R01 容器(Container)

> 范围: 容器资源的浏览、操作、运行时交互、表单与展示的统一约束

## 子需求

| 编号 | 标题 | 状态 | 优先级 | 来源 |
|---|---|---|---|---|
| [R01-01](./R01-01-advanced-ops.md) | 容器高级动作 (Copy / Update / Diff / Export / Commit / Wait) | implementing | high | [task/br-033-task019-advanced-container-actions.md](../../task/br-033-task019-advanced-container-actions.md) |
| [R01-02](./R01-02-exec-shell.md) | 容器 Exec 页面 shell UI | planned | medium | 原 BR-038 已归档 |
| [R01-03](./R01-03-stats-top.md) | Stats / Top / Wait 运行时刷新体验 | planned-review | medium | 状态待主模型确认 |

## 关联约束

- [constraint/C01-form.md](../../constraint/C01-form.md) — Copy / Update / Export / Commit 表单字段、焦点、提交规则
- [constraint/C05-path.md](../../constraint/C05-path.md) — Copy 源路径与 Export / Copy 目标路径语义
- [constraint/C04-keybinding.md](../../constraint/C04-keybinding.md) — Action Bar 分发

## 关联任务卡

- [task/br-033-task019-advanced-container-actions.md](../../task/br-033-task019-advanced-container-actions.md) — TASK-019 / BR-033 高级动作
- [task/br-033-advanced-container-ops-analysis.md](../../task/br-033-advanced-container-ops-analysis.md) — 实施前只读分析
- [task/br-033-small-model-implementation-prompt.md](../../task/br-033-small-model-implementation-prompt.md) — 小模型 prompt
- [task/br-033-035-review-fixes.md](../../task/br-033-035-review-fixes.md) — 主模型 review 修复
- BR-038 / BR-042 已合并归档,内容见 R01-02 / R01-03

## 待确认

- R01-03-stats-top 边界:是否包含 Wait 的 runtime 刷新体验?与 R01-01 重叠。
- Commit form 是否纳入 R01-01(已包含)还是独立 R01-04。