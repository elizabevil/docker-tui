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
├── R06-operation/                  # Operation 一等公民(横切:贯穿容器/镜像域)
│   └── R06-01-operation-as-domain.md
├── R07-config-ux/                  # 配置系统便利化(CLI info / config init / validate / 首次运行提示)
│   ├── README.md
│   ├── R07-01-cli-info.md          # dtui info — 配置路径与主题列表
│   ├── R07-02-config-init.md       # dtui config init — 生成配置文件模板
│   ├── R07-03-config-validate.md   # dtui config validate — 校验配置文件
│   └── R07-04-first-run-hint.md    # 首次运行提示(toast,不自动生成)
└── R08-compose/                    # Compose 项目级与编排(2026-08-07 立项)
    ├── README.md
    ├── R08-01-aggregation-model.md # 容器标签聚合,Docker/Podman 分轨适配
    ├── R08-02-engine-support.md    # ComposeService engine 抽象候选
    ├── R08-03-discover.md          # 项目 / 服务 / 配置 / 元信息发现
    ├── R08-04-lifecycle.md         # 生命周期 + 标志
    ├── R08-05-build-transfer.md    # build / pull / push
    ├── R08-06-logs.md              # 聚合日志(多服务 / 服务级)
    ├── R08-07-service-queries.md   # 服务级 top / port / stats
    ├── R08-08-exec-run.md          # exec / run
    ├── R08-09-advanced-control.md  # pause/unpause/kill/scale/rm/prune
    ├── R08-10-events.md            # 实时事件流(events)
    ├── R08-11-action-bar.md        # Action Bar 接入
    ├── R08-12-keymap.md            # Keymap 注册表接入
    └── R08-13-i18n.md              # i18n 接入
    └── R08-14-co-located-group.md  # 横切专题:服务与容器中间层(Docker/Podman 差异落点)
    └── R08-15-driver-implementation.md  # runtime driver 适配实现方案(docker / podman adapter 路径)
```

## 角色

每个 `R##-group/README.md` 是该大需求的范围、子需求清单、关联约束、关联任务卡的索引。
每个 `R##-##-topic.md` 是单一小需求的真理源,按 [docs-architecture.md §4](../docs-architecture.md) 模板组织。

## 关联

- 横切约束:[../constraint/](../constraint/)
- 单次实施卡:[../task/](../task/)
- 总需求清单:[../requirements.md](../requirements.md)
- 文档架构:[../docs-architecture.md](../docs-architecture.md)
- 历史/研究底稿:[../../../.omo/compose-todo.md](../../../.omo/compose-todo.md)(R08 前置)

## 待确认

- 状态字段:旧 task 文档的 `done / partial / open` 不可机械套用,需主模型逐条 review 后再标。
- 子需求边界:R01-03-stats-top 与 R03-03-action-dialog-layout 是否合并或独立,待 R03 大需求梳理后定。
- R08 大需求 2026-08-07 立项:`feature-todo-list.md §6.4` 与 `feature-design.md §3` 的"Compose up / build / pull 已取消"由本需求撤销;旧文本按项目政策保留作历史,但不作为实现阻断。重复功能(R08-04..R08-12 与 R01 容器级)的边界由各 R08-## 子需求的"⚠ 重复功能说明"节标注。
- R08-02 是否引入 ComposeService 仍在 `planned-review` 状态,主模型评审后定。