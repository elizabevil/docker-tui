# requirement/ — 需求树真理源

> 建立日期: 2026-08-02
> 状态: 形成中(按 [docs-architecture.md §8](../docs-architecture.md) 第一阶段)
> 范围: 仅整理,不修改代码

## 目录结构

```text
requirement/
├── README.md                       # 本文件
├── R01-container/                   # 容器大需求
│   ├── README.md
│   ├── R01-01-advanced-ops.md       # Copy/Update/Diff/Export/Commit/Wait
│   ├── R01-02-exec-shell.md         # Exec shell UI
│   └── R01-03-stats-top.md          # Stats/Top/Wait 运行时刷新体验
├── R02-image/                      # 镜像大需求
│   ├── README.md
│   ├── R02-01-history.md
│   ├── R02-02-import-tarball.md
│   └── R02-03-registry-login.md
├── R03-form-action/                # 表单与动作展示
│   ├── README.md
│   ├── R03-01-form-pattern.md
│   ├── R03-02-action-bar.md
│   └── R03-03-action-dialog-layout.md
├── R04-runtime/                    # 运行时能力
│   ├── README.md
│   └── R04-01-docker-podman-capabilities.md
├── R05-events-network/             # 事件流与网络
│   ├── README.md
│   └── R05-01-events-network-connect.md
└── R06-operation/                  # Operation 一等公民(横切:贯穿容器/镜像域)
    └── R06-01-operation-as-domain.md
```

## 角色

每个 `R##-group/README.md` 是该大需求的范围、子需求清单、关联约束、关联任务卡的索引。
每个 `R##-##-topic.md` 是单一小需求的真理源,按 [docs-architecture.md §4](../docs-architecture.md) 模板组织。

## 关联

- 横切约束:[../constraint/](../constraint/)
- 单次实施卡:[../task/](../task/)
- 总需求清单:[../requirements.md](../requirements.md)
- 文档架构:[../docs-architecture.md](../docs-architecture.md)

## 待确认

- 状态字段:旧 task 文档的 `done / partial / open` 不可机械套用,需主模型逐条 review 后再标。
- 子需求边界:R01-03-stats-top 与 R03-03-action-dialog-layout 是否合并或独立,待 R03 大需求梳理后定。