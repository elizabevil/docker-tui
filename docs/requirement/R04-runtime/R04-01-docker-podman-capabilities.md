# R04-01 Docker / Podman 专有能力评估

## 元信息

- 状态: planned-review
- 优先级: P3
- 来源: [../../task/task-020-podman-capabilities-evaluation.md](../../task/task-020-podman-capabilities-evaluation.md)、[../../task/task-020-capabilities-review.md](../../task/task-020-capabilities-review.md)
- 关联任务: 无
- 关联约束: 无

## 目标

评估 Docker / Podman runtime 在 dtui 各需求场景下的能力差异(目录枚举、exec / logs 流式、network 操作、registry 认证等),明确:
- 双 runtime 都支持的功能:统一接口,无差异
- 仅一端支持的功能:adapter 层标注,UI 显示能力提示
- 双 runtime 都不支持:标 unsupported,保留手工路径

## 用户流程

无需 UI 流程,本需求产出的是 [internal/data/runtime/](../../../internal/data/runtime/) adapter 层能力矩阵与 UI 层降级提示策略。

## UI/UX

依赖执行结果:
- 失败 toast 区分 docker / podman
- 双 runtime 不一致的功能在 Action Bar 灰显 + tooltip 说明

## 功能规则

- 不得在 UI 层假设 docker / podman 私有字段
- 能力差异由 adapter 暴露为常量或查询函数
- 第二阶段 Container Path 补全必须按本评估结论取舍

## 实现设计

- 评估产物:能力矩阵表(每个需求 × 双 runtime)
- adapter 接口约束:仅放双 runtime 都有的方法;差异方法放 docker-only / podman-only 子接口
- UI 层:读 adapter 暴露的 capability,动态过滤 Action Bar 项

## 验收标准

- 能力矩阵文档完成,所有现行需求覆盖
- 双 runtime 接口一致项在 `internal/data/runtime/` 列出
- 不一致项有明确降级策略

## 非目标

- 引入第三种 runtime
- 屏蔽 podman 私有能力

## 迁移记录

- 旧文档:[task/task-020-podman-capabilities-evaluation.md](../../task/task-020-podman-capabilities-evaluation.md)、[task/task-020-capabilities-review.md](../../task/task-020-capabilities-review.md)
- 保留信息:Podman vs Docker 已知能力差异列表
- 待确认状态:`planned-review`,原评估结论未进入实施;主模型确认后推进。