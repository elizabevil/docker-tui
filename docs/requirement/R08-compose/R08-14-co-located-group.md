# R08-14 服务与容器之间:CoLocated Group 轻量中间层

## 元信息

- 状态: planned
- 优先级: high
- 来源: 用户评审结果 2026-08-07("service 与 container 中间是否有 pod-like 层 — 收集差异 + 给方向") + Docker / Podman 实证
- 关联任务: 跨 R08-01 / R08-02 / R08-03 / R08-04 / R08-09 / R08-11 / R08-12(横切专题,**不**作独立功能)
- 关联约束:
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)
  - 无新约束(沿用 R08 大需求既有约束)

## 决策背景(基于 Docker / Podman primary-source 证据)

### D1 compose-spec 顶层不含 `pod` key

[`compose-spec/schema/compose-spec.json`](https://raw.githubusercontent.com/compose-spec/compose-spec/main/schema/compose-spec.json) 顶层 keys: `version / name / include / services / models / networks / volumes / secrets / configs`。**无 pod 顶层实体**。

service 内可触发 pod-化效应的字段只有 `network_mode: "service:[name]"`(共 net NS)、`pid: "service:[name]"`(共 PID)、`ipc: "service:[name]"`(共 IPC)、`container:[id]`(直接挂容器)。

### D2 podman-compose 默认:一个项目一个 Pod

[podman_compose.py](https://github.com/containers/podman-compose) `resolve_pod_name`(line 2543-2562):

```python
if in_pod_arg_parsed is True:
    return f"pod_{self.project_name}"            # <-- 默认
if in_pod_arg_parsed is False:
    return None                                   # <-- --in-pod=false 显式关闭
```

默认 pod args(line 2567-2570):

```python
return self.x_podman.get(
    PodmanCompose.XPodmanSettingKey.POD_ARGS,
    ["--infra=false", "--share="]                 # 无 infra 容器 + 共享所有 namespace
)
```

→ **podman-compose 把整个 compose 项目天然压成 1 个 Pod,所有 service 容器自动共享 net / pid / ipc namespace**。docker compose 没这个默认行为。

### D3 `podman compose` 是薄包装

[podman-compose.1.html](https://docs.podman.io/en/latest/markdown/podman-compose.1.html) 明确:`podman compose` 只是 thin wrapper,内调 docker-compose / podman-compose。Podman 内置命令不引入新实现。

### D4 Docker 端"pod-like"是作者显式声明

[Dille 2020](https://dille.name/blog/2020/09/20/use-docker-compose_to-manage-pods-on-docker/):"Docker does not implement the concept of a pod. Pods must be constructed by creating multiple containers with a shared network namespace." 且 `scale` 在 pod 模式**不工作**(端口冲突)。

### D5 Podman 引擎原生 Pod(脱离 compose)

Podman 有独立 `/libpod/pods/json` 端点,pod 是 first-class 实体。dtui 走 label 聚合,看不到这些原生 pod,需要专用的 `Pod 面板`(见 `feature-todo-list.md §6.4` 与 `design/podman-capabilities-analysis.md`),**不**在本 R08-14 处理。

## ⚠ 重复功能说明

| 能力 | 已存在层 | 本需求处理 |
|---|---|---|
| Podman 原生 Pod 面板 | `design/podman-capabilities-analysis.md § 184 + 244-257` 已规划 | **不重复**;本需求聚焦 compose-managed 容器 |
| docker compose `network_mode: service:` 隐式 pod | compose-spec 支持,dtui 现有不显式建模 | **本需求显式建模** |
| podman-compose 默认 1 项目 1 Pod | podman_compose.py 默认 True | **本需求在 UI 显示** group ID = `pod_<project>` |

## 目标

把 chat-level 结论"在 service 与 container 之间引入轻量中间层"沉到 R08 大需求。具体定义:

> **CoLocated Group**: dtui UI 用于标识"运行时层面被同一资源(Podman Pod / 共享 net/pid/ipc namespace 的 service group)绑定的一组 services"的**仅供 UI 聚合**的中间层。

**关键约束**:

1. **不是 k8s Pod** — 不与 docker compose spec / podman pod spec 冲突
2. **不是新引擎实体** — 不改 `runtime.Engine` 接口(R08-02 不动)
3. **派生而非申报** — 所有数据从 container 元数据推导,不需要解析 compose yaml
4. **运行时分支但 rt-agnostic UI** — 同 R08-01 原则,DTO 设计为 runtime-agnostic,只在 adapter 实现里分支

## 用户流程

1. Compose 面板左栏选中项目 → 项目 detail header 增字段 "Groups: 2"(只显示非空)
2. 右栏 services 列表,组内 service 同行末尾显示折叠标签 `[shared-net]`,hover 展开组成员
3. drill-in service 子视图顶部增 group 元信息 line:`本 service 属于 group shared-net,包含 api + worker`
4. Action Bar 加 group 维度条目:`Group Down` / `Group Restart`(详见 R08-11)
5. scale 操作触发:group 内 service scale 必须同步(详见 R08-09)

## UI/UX

仅本需求特有:

- group 标签默认折叠:每行末尾 `<2>` 这种小标记
- 展开状态:在 services 子表上方开一条 "Groups ([N])" 子表
- 颜色:与 Podman / Docker 引擎色保持中性,不暗示"运行时原生"概念

详细版式留 UI overhaul 阶段,本需求只规定字段与交互边界。

## 功能规则

### F1 CoLocatedGroup DTO

```go
// R08-14 跨切类型,Append 到 Compose 域
//
// 与 ContainerSummary 同级,与 ComposeLabels 同级。
// 与 ComposeProjectSummary 内嵌展示区分,DTO 独立。
type CoLocatedGroup struct {
    ID            string   // 推荐:Podman="pod_<project>"; Docker="<inferred-deterministic-id>"
    Runtime       string   // "docker" | "podman" | ""
    Source        string   // 来源枚举,见 F3
    Services      []string // 参与该 group 的 service 名
    ContainerIDs  []string // 实时快照,UI 渲染用
}
```

### F2 触发与识别

| 运行时 | Source | 识别方法 | Group ID 规则 |
|---|---|---|---|
| Podman | `podman_default_pod` | 容器 inspect 包含 Pod 字段(podman REST `/containers/json` 列表 + inspect 含 `.Pod` 字段) | `pod_<project-name>`(与 podman-compose 默认行为对齐) |
| Podman | `podman_explicit_in_pod_false` | `x-podman.in_pod: false` 或 `--in-pod=false`(通过 inspect 反推:Pod 字段但共享 namespace 不一致,或用户断定为 single-container pod) | **不**显示 group(等同 docker 默认) |
| Docker | `docker_network_mode_service` | 某 service 出现 `network_mode: "service:[X]"`,且 X 被其它 service 引用 | group ID = `<project>/<X>`(X 是被引用方) |
| Docker | `docker_pid_service` | 出现 `pid: "service:[X]"`,且 X 被引用 | 同上规则,与 net_mode 取并集 |
| Docker | `docker_ipc_service` | 出现 `ipc: "service:[X]"`,且 X 被引用 | 同上规则 |
| 不识别 | `unrecognized` | 其余普通 service | 不显示 group |

**关键判定**:单 service 自我引用或无其它 service 引用 → **不**算 group(避免单 service 也加个标签噪音)。

### F3 Inspect 读取范围

Docker 端识别需要 **inspect 数据**,不在 list 阶段;Podman 端 list / inspect 都行。

为避免 inspect 风暴(每容器一次 RTT),R08-14 引入增量:

- List 阶段:仅 Podman `Pod` 字段(快,无额外 RTT)
- Drill-in / detail 阶段:按需查 docker inspect(`NetworkSettings.NetworkMode` + `Pid` + `Ipc`)
- 缓存:inspect 结果存 `coLocatedCache`(TTL 30s 或资源变更失效)

### F4 group 内 service 视图

- services 列表同行末尾:折叠态 `<N>`(N=组内 service 数),hover 展开
- drill-in 子视图:顶部 group info line,展示 group ID + 成员 service 名列表
- Project detail:Header 增 `Groups: <count>`

### F5 scale 行为

| 运行时 | group 内 scale 行为 |
|---|---|
| Docker | group 内 service scale 必须 = 1(network_mode: service: 强制 1:1,UI 灰显 scale Action) |
| Podman | group 内 service 可独立 scale,但 group ID 共享;scale 一个组内 service 不影响其它 |

### F6 down 行为

- down CoLocated group = 对该 group 内所有 service 执行 down = 对每个 service 的容器 + 该 service 独占卷 + 该 service 独占网络 delete
- "down 整个项目"覆盖 group 语义
- audit trace 含 `group_id` 字段

### F7 audit 类型扩展

`ComposeTarget.Meta`(R08-01 已定义)加字段:

```go
type ComposeMeta struct {
    Containers int      `json:"containers,omitempty"`
    Volumes    int      `json:"volumes,omitempty"`
    Networks   int      `json:"networks,omitempty"`
    GroupID    string   `json:"group_id,omitempty"` // R08-14
    Runtime    string   `json:"runtime,omitempty"`  // R08-14
    Source     string   `json:"source,omitempty"`   // R08-14 来源
}
```

## 实现设计

### I1 数据模型扩展(跨 R08-01 / R08-02)

**R08-01** (`internal/data/runtime/containers.go` `ContainerSummary`) 增字段:

```go
type ContainerSummary struct {
    // ... existing
    CoLocatedGroupID  string  // R08-14: 派生 group ID
    CoLocatedGroupSrc string  // R08-14: 来源枚举
}
```

**R08-02** (`internal/data/runtime/compose_aggregator.go` `ComposeProjectSummary`) 增字段:

```go
type ComposeProjectSummary struct {
    // ... existing
    CoLocatedGroups []CoLocatedGroup `json:"co_located_groups,omitempty"`
}
```

### I2 Docker adapter(`internal/data/runtime/docker/containers.go`)

list 阶段不识别(group 需要 inspect);drill-in 时由调用方触发 inspect,识别逻辑放在新文件:

- 新文件:`internal/data/runtime/docker/co_located.go`
- 公共函数:`IdentifyCoLocatedGroup(ctr *container.InspectResponse, project string) CoLocatedGroup`
- 读 `ctr.NetworkMode`(若以 `service:` 开头,提取被引用方)
- 读 `ctr.PidMode` / `ctr.IpcMode`
- 聚合:仅当 2+ service 引用同一目标时,生成 group

### I3 Podman adapter(`internal/data/runtime/podman/mappers.go`)

list 阶段即可识别,因为 `/containers/json` Podman 端返回项含 `Pod` 字段:

- 新文件:`internal/data/runtime/podman/co_located.go`
- 公共函数:`IdentifyCoLocatedGroup(container PodmanContainerList, project string) CoLocatedGroup`
- 若 `.Pod` 非空 → group ID = `pod_<project>`,Source = `podman_default_pod`(注:不区分是否同 pod,因为容器列表无法确定;降级时:`x-podman.in_pod: false` 通过 inspect `.Pod=""` 与 `.HostConfig.NetworkMode != ""` 区分;不识别时归类为 docker 模型)

### I4 双 adapter 独立,R08-01 原则保持

两份 adapter **互不调用**(per R08-01 F1);只是返回的 DTO 形态一致。

### I5 UI 集成(跨 R08-03 / R08-11)

- `internal/tui/ui/pages/compose/view.go:36` `gatherComposeProjects` 增 group 计算
- `view.go:217` 右栏标题增强:增 group badge
- drill-in 子视图:group info line
- Action Bar:见 R08-11 增补条目(由 R08-11 引用本需求 F7 audit 字段)

### I6 Keymap 接入(跨 R08-12)

新增 KeyAction(详 R08-12):

- `ActionComposeGroupDown` (Shift+Ctrl+D)
- `ActionComposeGroupRestart` (Shift+Ctrl+R)
- `ActionComposeGroupExec` (Shift+Ctrl+E) — group 内选 1 个容器 exec

### I7 i18n keys(跨 R08-13)

```text
compose.co_located.group_label
compose.co_located.group_members
compose.co_located.no_group
compose.co_located.source_docker_network_mode
compose.co_located.source_podman_default
compose.co_located.source_unrecognized
compose.co_located.scale_blocked_docker
```

## 验收标准

1. Docker `network_mode: service:db` + 2 个其它 service 引用 → dtui 显示 group ID `<project>/db`,services 列表含 2 个引用方
2. Docker 无 `network_mode: service:` → dtui 无 group 显示(默认行为,group 列为空)
3. Podman podman-compose 默认行为(无 `x-podman.in_pod: false`)→ 所有 service 共享 `pod_<project>` group ID
4. Podman `x-podman.in_pod: false` → dtui 不显示 group(等同 Docker 默认)
5. Docker `network_mode: service:` 的 service scale 操作被 UI 灰显
6. Podman group 内 service scale 操作可独立触发,audit trace 含 group_id
7. group down / restart 操作触发单条 audit trace,Meta 含 group_id + runtime + source
8. 双 runtime 各自的 co_located.go 文件存在,**互不调用**
9. Unit test 覆盖 docker / podman 的 list / inspect 数据 → CoLocatedGroup 推导

## 非目标

- 不实现 Podman 原生 Pod 面板(`design/podman-capabilities-analysis.md §184` 已在另一独立需求规划)
- 不解析 compose yaml(本需求所有数据从 container 元数据推导)
- 不引入 yaml / compose-spec 解析器
- 不支持多 podman 实例(每个 connection 各自聚合)
- 不在 docker compose stack deploy (Swarm) 模式处理

## 迁移记录

- 旧决策:沿用 R08 大需求"不引入新 Engine 接口"原则,**不复活** TASK-022 Phase E
- 旧代码:`internal/data/runtime/docker/containers.go`、`internal/data/runtime/podman/mappers.go`(无 group 字段)
- 复用:R08-01 / R08-02 数据模型扩展
- 废弃信息:无
- 待确认:
  1. Docker inspect 在 dtui 已有调用路径吗?(R01 inspect 已有)→ **是**,复用路径
  2. Podman container list 是否所有版本都包含 `.Pod` 字段? → 检查 `internal/driver/podman/testdata/`
  3. dtui 是否愿意为 Podman 端加 list 阶段 inspect 二次调用? → 倾向不加,因为 `/containers/json` 已经包含 Pod 字段
