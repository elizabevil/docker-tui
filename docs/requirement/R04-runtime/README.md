# R04 Runtime

> 范围: Docker / Podman runtime 能力差异与适配层

## 子需求

| 编号 | 标题 | 状态 | 优先级 | 来源 |
|---|---|---|---|---|
| [R04-01](./R04-01-docker-podman-capabilities.md) | Docker / Podman 专有能力评估 | implementing | P3 | [task/task-020-podman-capabilities-evaluation.md](../../task/task-020-podman-capabilities-evaluation.md) + [task/task-020-capabilities-review.md](../../task/task-020-capabilities-review.md) |

## 关联约束

- 暂无(运行时差异主要在 [internal/data/runtime/](../../../internal/data/runtime/) adapter 层,不直接体现 UI 约束)

## 关联任务卡

- [task/task-020-podman-capabilities-evaluation.md](../../task/task-020-podman-capabilities-evaluation.md)
- [task/task-020-capabilities-review.md](../../task/task-020-capabilities-review.md)

## 待确认

- R04-01 状态:`implementing` 偏乐观;实际只完成评估未进入实施。是否应标 `planned-review`。
- 是否需要新增 R04-02 记录双 runtime 接口规范(adapter 边界)。