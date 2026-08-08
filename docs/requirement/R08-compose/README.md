# R08 Compose(项目级与编排)

> 范围: Compose 项目的发现、生命周期、聚合查询、镜像构建与传输、Action Bar / Keymap / i18n 全套集成。
>
> 真理源: 本目录的 `R08-##-*.md` 子需求文档,旧文档 [.omo/compose-todo.md](../../../.omo/compose-todo.md) 仅作研究底稿与历史台账。

## 重新立项背景

- 项目立项时(`docs/feature-todo-list.md §6.4`)曾将 `Compose up / build / pull` 列进"已取消的能力"档。
- 实际使用中 `dtui` 用户面**确实**存在多容器编排场景:`docs/requirements.md §场景 3` 已把 Compose 项目管理列为标准用户场景之一。
- 用户评审结果:**完整实现** Compose 相关功能,重新立项为 R08 大需求。已取消条目仅作历史记录,不再作为实现阻断依据。

> 重新提出 = 在本目录(R08)新建子需求并重审,这是项目政策的合规路径。

## 范围

### 范围内(in scope)

- 项目级发现:`ps` / `ls` / `config` / 元信息(working_dir / config_files / version)
- 生命周期:`up` / `down` / `start` / `stop` / `restart`
- 镜像动作:`build` / `pull` / `push`
- 聚合查询:`logs`(多服务 / 单服务范围)/ `top` / `port` / `stats`
- 一次性执行:`run` / `exec`
- 高级控制:`pause` / `unpause` / `kill` / `scale` / `rm` / `prune`
- 事件:`events`
- 集成层:Action Bar(scope)/ Keymap / i18n
- 双 runtime 适配:Docker 与 Podman 独立 label 处理、空项目可见性、podman-compose 包装器兼容

### 范围外(out of scope)

- `docker compose convert` / `compose plugin management` / `compose version` 子命令详情 — 留作基础设施
- Podman kube / podman secret / podman pod(暂留 R04-runtime 跟进)
- Docker Swarm / Stack 部署(在 feature-design.md §5 范围内)
- Compose Spec 多文件 `include` 高级语法(单 compose 文件工作流优先)
- 多主机 active/passive 部署(`docs/requirements.md P2`)

## 子需求

| 编号 | 标题 | 状态 | 优先级 | 来源 |
|---|---|---|---|---|
| [R08-01](./R08-01-aggregation-model.md) | 容器标签聚合模型(Docker/Podman 分轨适配) | implementing | high | 用户评审结果 2026-08-07 |
| [R08-02](./R08-02-engine-support.md) | ComposeService engine 抽象候选 | implementing | high | 用户评审结果 2026-08-07 |
| [R08-03](./R08-03-discover.md) | 项目 / 服务 / 配置发现(ps / ls / config / 元信息) | planned | high | R02 镜像 / R04 runtime 推导 || [R08-04](./R08-04-lifecycle.md) | 生命周期(up / down / start / stop / restart + 标志) | planned | high | docker compose CLI 权威 |
| [R08-05](./R08-05-build-transfer.md) | 镜像构建与传输(build / pull / push) | planned | medium | docker compose CLI 权威 |
| [R08-06](./R08-06-logs.md) | 聚合日志(多服务 / 服务级) | planned | high | R01 容器 + compose 增量 |
| [R08-07](./R08-07-service-queries.md) | 服务级联查询(top / port / stats) | implementing | medium | R01 容器动作复用 |
| [R08-08](./R08-08-exec-run.md) | 一次性执行(run / exec) | implementing | medium | R01 容器 exec 复用 |
| [R08-09](./R08-09-advanced-control.md) | 高级控制(pause / unpause / kill / scale / rm / prune) | planned | low | docker compose CLI 权威 |
| [R08-10](./R08-10-events.md) | 实时事件流(events) | implementing | medium | R05 事件流接续 |
| [R08-11](./R08-11-action-bar.md) | Action Bar 接入 Compose | implementing | high | [R03-02](../R03-form-action/R03-02-action-bar.md) 扩展 |
| [R08-12](./R08-12-keymap.md) | Keymap 注册表接入 Compose | implementing | high | [R01-container 关联约束](../../constraint/C04-keybinding.md) 接续 |
| [R08-13](./R08-13-i18n.md) | i18n 接入 Compose(命名空间与键表) | planned | medium | [constraint/C06-i18n.md](../../constraint/C06-i18n.md) 扩展 |
| [R08-14](./R08-14-co-located-group.md) | **横切专题:服务与容器中间层**(CoLocated Group) — Docker / Podman 差异落点 | implementing | high | 用户评审 2026-08-07 跨切决策 |
| [R08-15](./R08-15-driver-implementation.md) | runtime driver 适配实现方案 — docker / podman adapter 路径 | implementing | high | 用户评审 2026-08-07 驱动实现讨论 |

## 架构约束(用户评审结论)

1. **Docker / Podman label 必须独立处理**,不允许单 adapter 共享 label 检测逻辑。`Adapter.composeLabels()` 在两个 runtime 中各自实现,在 aggregator 处合并去重。
   - 现状:`internal/data/runtime/docker/containers.go:84-89` 与 `internal/data/runtime/podman/mappers.go:81-82` 各自手写 label 读取。**R08-01** 把它们重构成形态一致但内容独立的适配器。

2. **可能引入独立的 `ComposeService`**(`runtime.Engine.Compose() ComposeService`)。
   - 与 TASK-022 Phase E 取消的 `docker/service` 统一入口**不同**:`ComposeService` 是面向**特定**领域的 service 接口,不是"统一入口"。两件事可以同时为真。
   - 决策候选:**有条件引入**(详见 R08-02 §实现设计)。

3. **不引入** `exec.Command("docker", "compose", ...)` 路径。dtui 的"Engine API 单交互面"原则保留。Compose-级命令全部走 `ComposeService` + 容器 / 卷 / 网络 API 聚合。

## 重复功能标注政策(用户要求)

下列能力在 Compose 层级与容器 / 镜像 / 卷 / 网络层**同时存在**。每个 R08-## 中包含 "**重复功能说明**" 节,记录:

- 已实现层:位置 + 入口
- 重复层级的**增量价值**:为什么仍要在 compose 层级做
- 重复层级**不做**的项:明确收紧

| 能力 | 已实现层 | 重复层级的处理 |
|---|---|---|
| start / stop / restart / pause / unpause / kill | R01 容器动作 | R08-04 / R08-09:Compose-级一次性触发项目所有 / 单 service 容器,功能等价但调度范围不同 |
| logs | R01 容器 | R08-06:**多服务聚合**(容器层没有);**保留**服务级范围 |
| exec | R01 容器 | R08-08:Compose-级 = 选 service → 复用容器 exec。exec 仍走容器 API |
| top / port / stats | R01 容器 | R08-07:Compose-级 = 一次查整个项目各服务的对应信息(逐 service 表格),否则等于"循环 Tab 到容器" |
| build / pull / push | R02 镜像 | R08-05:Compose-级 = 按 service 列表循环镜像动作 + 进度聚合 |
| rm / prune | R05 events-network(含 Network 删删)+ R03 卷删删 | R08-09:Compose-级 rm / prune 提供"批量语义"(项目范围) |
| config | R04 runtime 间接 | R08-03:`compose config` 解析结果在 dtui 可读 |
| events | R05 事件流 | R08-10:事件按 project / service 维度过滤,**不**新建独立面板 |

> **明确收紧原则**: 凡 Compose-级只是"循环到容器层级"的能力,如果 Compose-级**不**提供增量价值,**不**做。例如 `compose run` 在 `docker compose` 是"创建一次性容器并启动"——容器 Run 动作已存在,但要新建"compose-managed"一次性容器需要的额外参数(create + start + label 注入)。**R08-08** 需要评估该参数集并给出最小方案。

## 关联约束

- [constraint/C01-form.md](../../constraint/C01-form.md) — build / 路径类 Form 字段
- [constraint/C02-dialog.md](../../constraint/C02-dialog.md) — Confirm / Detail 弹层
- [constraint/C03-table.md](../../constraint/C03-table.md) — 服务列表、容器子视图表格
- [constraint/C04-keybinding.md](../../constraint/C04-keybinding.md) — Action Bar 与直接键分层
- [constraint/C05-path.md](../../constraint/C05-path.md) — config_files 路径展示、tar 导出 / 加载
- [constraint/C06-i18n.md](../../constraint/C06-i18n.md) — i18n 命名空间前缀 `compose.*`

## 关联任务卡 / 历史台账

- [.omo/compose-todo.md](../../../.omo/compose-todo.md) — 用户进行评审前的研究底稿与代码验证结论,保留作历史
- [.omo/compose-requirements-discussion.md](../../../.omo/compose-requirements-discussion.md) — Label 规范讨论,信息已被 R08-01 / R08-11 摘要合并
- [docs/bugfix-requirements.md](../../bugfix-requirements.md#br-017-compose-详情页不应把快捷键写进正文) — BR-017 关联 R08-12(footer 上下文投影)
- [docs/requirements.md §场景 3](../../requirements.md) — Compose 项目管理用户场景
- [docs/feature-todo-list.md §6.4](../../feature-todo-list.md) — 历史已取消条目,仅作历史不再阻断

## 待确认

1. **R08-02 ComposeService** 是否确认引入 —— **已定(2026-08-07):确认引入**。`runtime.Engine.Compose() ComposeService` 已实现(接口含 ListProjects / InspectProject / Config / Start / Stop / Restart / Down / Up / Run / Exec / Top / Port / Stats / Events),docker / podman 双 adapter 独立实现,`Config` 永久 `ErrComposeUnsupported`。
2. **重复功能收紧边界** —— **已定(2026-08-08):允许 Compose-级单 service 动作,与容器页面区分**。语义:容器页面调容器级 API(如容器 Stop),Compose 页面调 Compose-级 API(项目/服务范围 Stop),**API 路径不同,仅用户可感知的结果相同**;两者并存,UI 上区分粒度。
3. **`compose ls` 空项目语义** —— **已定(2026-08-08):无容器则隐藏**。与 `docker compose ls` 权威行为一致(`pkg/compose/ls.go` 仅按容器 `com.docker.compose.project` label 聚合,容器是项目存在的唯一载体;`--all` 仅包含 stopped 容器,不产生"无容器项目")。R08-03 **不**补 empty-project 来源。`ComposeService.ListProjects` 以 `ListProjectsOptions{All bool}` 暴露 stopped 容器范围,默认 false 与 CLI 默认一致。

## 实施阶段建议(参考,主模型决策优先)

| 阶段 | 子需求 | 备注 |
|---|---|---|
| **阶段 A:基础正确性** | R08-01 / R08-11 / R08-12 | 修复 Ctrl+D dead code、补 Action Bar scope、Keymap 接入 |
| **阶段 B:项目级运转** | R08-04 (含 down -v/--rmi) / R08-06 / R08-03 | 项目级生命周期 + 多服务日志聚合 + 项目元信息发现 |
| **阶段 C:服务级动作** | R08-07 / R08-09 / R08-08 | 服务级 top / port / stats / pause / unpause / kill / scale + run / exec |
| **阶段 D:镜像集成** | R08-05 | build / pull / push 与服务循环 |
| **阶段 E:事件流** | R08-10 | 事件流 project / service 过滤 |
| **阶段 F:i18n** | R08-13 | 命名空间补齐与多语翻译 |
