# dtui 任务卡目录 (Task Cards)

> 建立日期: 2026-08-01
> 状态: 2026-08-02 起,本文档**不再是需求真理源**,仅作为单次实施卡的保留位置。
> 需求真理源已迁移到 [../requirement/](../requirement/)(按 5 大需求 R01-R05 组织),横切规则在 [../constraint/](../constraint/)。

## 角色

- [../requirement/](../requirement/):需求树真理源(目标、用户流程、UI/UX、验收、迁移记录)
- [../constraint/](../constraint/):横切约束真理源(Form / Dialog / Table / 快捷键 / Path / i18n)
- [../task/](./README.md):单次实施卡(代码索引、prompt、review 记录)— **不是真理源**

## 与本文的关系

- **[../requirements.md](../requirements.md)**: 实现能力与状态总表。
- **[../requirement/](../requirement/)**: 需求树真理源。
- **[../constraint/](../constraint/)**: 横切约束真理源。
- **[../feature-design.md](../feature-design.md)**: 历史功能设计,后续逐步收敛到 [../requirement/](../requirement/)。
- **[../feature-todo-list.md](../feature-todo-list.md)**: 历史功能待办,后续逐步拆进 [../requirements.md](../requirements.md) 与 [../requirement/](../requirement/)。
- **[../bugfix-requirements.md](../bugfix-requirements.md)**: BUG 修复台账(全量)。
- **[../pending-bugs.md](../pending-bugs.md)**: BUG 级待办快照。
- **[../docs-architecture.md](../docs-architecture.md)**: 文档真理源架构说明。
- **[../architecture.md](../architecture.md)**: 实现架构说明(代码事实)。
- **[../project-structure.md](../project-structure.md)**: 目录结构。

## 任务卡索引

| 编号 | 优先级 | 标题 | 关联 |
|---|---|---|---|
| [TASK-020](./task-020-podman-capabilities-evaluation.md) | P3 | Docker / Podman 专有能力评估 | feature-todo-list §1 |
| [BR-033](./br-033-task019-advanced-container-actions.md) | **high** | TASK-019 容器高级动作 (Copy/Update/Diff/Export/Commit/Wait) | feature-todo-list §2 + bugfix-requirements BR-033 |
| [BR-034](./br-034-image-history-top-page.md) | **high** | 镜像 History 顶层页 (H 键) | feature-todo-list §2 + bugfix-requirements BR-034 |
| [BR-035](./br-035-events-panel-and-network-connect.md) | medium | Events 独立面板 (F3) + Network Connect/Disconnect | feature-todo-list §2 + bugfix-requirements BR-035 |
| BR-036(已归档→[R02-02](../requirement/R02-image/R02-02-import-tarball.md)) | medium | 镜像 tarball 导入 (Image Import) | feature-todo-list §2 + bugfix-requirements BR-036 |
| BR-037(已归档→[R02-03](../requirement/R02-image/R02-03-registry-login.md)) | medium | docker / podman Registry Login(私库认证) | feature-todo-list §2 + bugfix-requirements BR-037 |
| BR-038(已归档→[R01-02](../requirement/R01-container/R01-02-exec-shell.md)) | medium | 容器 Exec 页面 shell UI | feature-todo-list §2 + bugfix-requirements BR-038 |
| [BR-039](./br-039-action-bar-replace-q.md) | **done** | 取消 Q 退出,改为每页多功能 Action Bar | feature-todo-list §2 + bugfix-requirements BR-039 |
| BR-040(已归档→[R03-03](../requirement/R03-form-action/R03-03-action-dialog-layout.md)) | medium | Dialog 风格统一:四周透明 + panel 居中 | feature-todo-list §2 + bugfix-requirements BR-040 |
| [BR-041](./br-041-unified-form-path-completion.md) | **high** | FORM 导航与路径补全统一 | BR-033 后续交互规范 |
| BR-042(已归档→[R01-03](../requirement/R01-container/R01-03-stats-top.md)) | **high** | FORM 焦点/光标与 Commit、Wait、Top、Stats 修正 | BR-033 + BR-041 运行时缺陷 |

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
