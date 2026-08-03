# dtui 任务卡目录 (Task Cards)

> 建立日期: 2026-08-01
> 目的: 把 [docs/feature-todo-list.md](../feature-todo-list.md) 中**未完成的需求**展开成可执行任务卡,
>       包含文件索引、操作链路、验收标准。
>
> 每一张任务卡回答两个问题:
> 1. **"这个功能涉及哪些文件"** — 代码结构索引
> 2. **"这个 action 从按键到 runtime 怎么走"** — 操作链路

## 与本文的关系

- **[../feature-todo-list.md](../feature-todo-list.md)**: 高层 TASK + 设计驱动需求汇总 + 已取消决议。
- **[../feature-design.md](../feature-design.md)**: 设计意图 / 决策 / 不在范围 / 对比矩阵。
- **[../bugfix-requirements.md](../bugfix-requirements.md)**: BUG 修复台账(全量)。
- **[../pending-bugs.md](../pending-bugs.md)**: BUG 级待办快照。
- **[../architecture.md](../architecture.md)**: 实现架构说明。

> 本目录只覆盖 feature-todo-list.md §1 / §2 的待办项。已取消项(§3)与 BUG 修复(见 bugfix-requirements.md)不在本目录展开。

## 任务卡索引

| 编号 | 优先级 | 标题 | 关联 |
|---|---|---|---|
| [TASK-020](./task-020-podman-capabilities-evaluation.md) | P3 | Docker / Podman 专有能力评估 | feature-todo-list §1 |
| [BR-033](./br-033-task019-advanced-container-actions.md) | **high** | TASK-019 容器高级动作 (Copy/Update/Diff/Export/Commit/Wait) | feature-todo-list §2 + bugfix-requirements BR-033 |
| [BR-034](./br-034-image-history-top-page.md) | **high** | 镜像 History 顶层页 (H 键) | feature-todo-list §2 + bugfix-requirements BR-034 |
| [BR-035](./br-035-events-panel-and-network-connect.md) | medium | Events 独立面板 (F3) + Network Connect/Disconnect | feature-todo-list §2 + bugfix-requirements BR-035 |
| [BR-036](./br-036-image-import-tarball.md) | medium | 镜像 tarball 导入 (Image Import) | feature-todo-list §2 + bugfix-requirements BR-036 |
| [BR-037](./br-037-registry-login.md) | medium | docker / podman Registry Login(私库认证) | feature-todo-list §2 + bugfix-requirements BR-037 |
| [BR-038](./br-038-exec-shell-ui.md) | medium | 容器 Exec 页面 shell UI | feature-todo-list §2 + bugfix-requirements BR-038 |
| [BR-039](./br-039-action-bar-replace-q.md) | **done** | 取消 Q 退出,改为每页多功能 Action Bar | feature-todo-list §2 + bugfix-requirements BR-039 |
| [BR-040](./br-040-dialog-style-center-on-panel.md) | medium | Dialog 风格统一:四周透明 + panel 居中 | feature-todo-list §2 + bugfix-requirements BR-040 |
| [BR-041](./br-041-unified-form-path-completion.md) | **high** | FORM 导航与路径补全统一 | BR-033 后续交互规范 |
| [BR-042](./br-042-form-runtime-refresh-corrections.md) | **high** | FORM 焦点/光标与 Commit、Wait、Top、Stats 修正 | BR-033 + BR-041 运行时缺陷 |

## 任务卡模板

每张任务卡包含:

```markdown
# 任务标题

## 元信息
- 编号 / 关联 BR / 优先级 / 状态 / 依赖

## 目标
(一段话描述)

## 代码结构索引
- 必须读懂的文件
- 必须修改的文件
- 必须新增的文件
- 不应触碰的文件

## 操作链路
(从用户按键 / UI 触发到 runtime 调用的完整路径)

## 验收标准
(可勾选的验收清单)

## 风险
(影响 Docker / Podman / Help / Footer / i18n / docs / 表格布局 / 详情滚动 / 批量等)

## 建议任务分解
(可分批的最小可交付单元)

## 待确认项
(需要主模型决策的问题)
```

## 工作流

```text
1. 主模型确认需求边界 → 任务卡已生成
2. 小模型按任务卡只读定位 / 补测试 / 机械巡检
3. 主模型实现关键逻辑
4. 小模型同步 docs/bugfix-requirements.md / feature-todo-list.md / navigation.md
5. 主模型审查 + 提交
```

## 参考资源

- [docs/ai-prompts.md](../ai-prompts.md):小模型协作 prompt 模板
- [docs/architecture.md](../architecture.md):实现架构
- [docs/project-structure.md](../project-structure.md):目录结构
