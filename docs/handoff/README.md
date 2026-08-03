# handoff/ — 主模型 ↔ 子模型 通信

> 建立日期: 2026-08-03
> 状态: 形成中
> 范围: 仅新建文档,不修改现有结构

## 角色

| 文件 | 角色 |
|---|---|
| [README.md](./README.md) | 本文件:索引 + 使用约定 |
| [GLOBAL.md](./GLOBAL.md) | **全局信息**:主模型决策日志 + 子模型待确认问题(append-only) |
| [STATE.md](./STATE.md) | **task 任务驱动 — 当前**:活跃批次、子任务范围与进度 |
| [COMPLETED.md](./COMPLETED.md) | **task 任务驱动 — 历史**:已交接归档批次 |

## 使用约定

### 主模型

1. 决策追加到 `GLOBAL.md`(带日期与影响范围)
2. 启动批次时更新 `STATE.md`(批次范围、子任务、不可触碰边界)
3. 整合完成后把批次移到 `COMPLETED.md`
4. 不在子模型期间被要求时修改 `STATE.md`

### 子模型

1. 开工前先读 `STATE.md` 获取当前批次范围与边界
2. 跨会话延续时也读 `COMPLETED.md` 了解历史决策
3. 提问追加到 `GLOBAL.md`(问题章节),不要散落在外部
4. 完成时在 `STATE.md` 标 `done`,主模型整合后归档

### 写入规则

- 增量追加,不覆盖历史
- 时间戳: `YYYY-MM-DD HH:MM`(本地时区)
- 章节按时间倒序(最新在上)
- STATE.md 是单当前视图(主模型维护)

## 与其他文档的关系

- [../requirement/](../requirement/):需求真理源,本目录不复制需求内容,只引用
- [../constraint/](../constraint/):横切约束真理源,本目录引用而不复制
- [../task/](../task/):单次实施卡,本目录跟踪批次但不替代单卡
- [../ai-prompts.md](../ai-prompts.md):会话 prompt 模板,与本目录互补(模板 vs 状态)
- [../docs-architecture.md](../docs-architecture.md):文档真理源架构说明

## 待确认

- 命名 `handoff/` 是否合适,或用 `coordination/` / `comm/` / 其他
- STATE.md 与 GLOBAL.md 的更新职责边界(目前主模型维护)
- COMPLETED.md 是否需要按月/年归档子目录