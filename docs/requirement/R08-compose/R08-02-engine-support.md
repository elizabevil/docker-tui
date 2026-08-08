# R08-02 ComposeService engine 抽象候选

## 元信息

- 状态: planned-review
- 优先级: high
- 来源: 用户评审结果 2026-08-07("可能要提供单独的 comeposeservice")
- 关联任务: [.omo/compose-todo.md 补 E.4 / 补 F.2 / 补 F.3](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)
  - 历史决议 [`docs/feature-todo-list.md §6.1 TASK-022 最终范围` Phase E](../../feature-todo-list.md) — 取消 `docker/service` 统一入口

## ⚠ 重复功能说明(必须先读)

`ComposeService` 与 TASK-022 Phase E 取消的 `docker/service` **不是同一件事**:

| | `docker/service` 统一入口(取消) | `ComposeService` 领域 service(本需求候选) |
|---|---|---|
| 用途 | 把所有 docker 适配方法收口到一个 service facade | 对接 compose-级能力(项目发现 / 生命周期 / 聚合) |
| 抽象层级 | 位于 Container / Volume / Network / Image **同一层**之上 | 与 Container / Volume / Network / Image **同层**新增 |
| 影响范围 | 改 ContainerService / VolumeService / NetworkService 的包装路径 | 新增一个 service,不改其它 service 形态 |
| TASK-022 取舍 | 决策:不引入 | 决策:**未决议**(本需求提议) |

> 本需求不复活 Phase E,仅讨论"ComposeService 在 Engine 接口层是否新增"。

## 目标

Engine 接口(`internal/data/runtime/engine.go:7`)当前不含 Compose-级方法。本需求提出**候选** `ComposeService` 接口,内置 Compose-级生命周期 / 发现 / 聚合查询方法,供 R08-03 / R08-04 / R08-06 / R08-07 / R08-09 等需求使用。

需在 R08-02 内解决:

1. 服务接口形态(方法集 / 入参 / 返回)
2. Docker 适配器实现路径(API 限制告知)
3. Podman 适配器实现路径(`podman compose` 现状约束)
4. 聚合回退(当该 runtime 不支持 compose-level API 时回退到 label 聚合)
5. 是否落地的最终决策(本需求 `planned-review` 状态,主模型评审后改)

## 用户流程

不直接交互(架构需求)。具体用户流程在 R08-04 / R08-06 等子需求中。

## UI/UX

不涉及。

## 功能规则

### F1 接口形态(候选 v1)

```go
// internal/data/runtime/compose_service.go (新增)
type ComposeProjectSummary struct {
    Name         string
    Services     []ComposeServiceSummary
    WorkingDir   string
    ConfigFiles  []string
    Version      string
    CreatedAt    time.Time
    HasRunning   bool
    HasStopped   bool
    Source       string // "docker" | "podman" | "aggregated"
}

type ComposeServiceSummary struct {
    Name     string
    Image    string
    Replicas int
    Running  int
    Stopped  int
}

// ComposeService 提供 compose-级操作。
//
// 方法实现矩阵(per R08-15 Q3 决策 —— docker 与 podman 端**对称**):
//
//   方法                实现路径                                 失败模式
//   ─────────────────────────────────────────────────────────────────────
//   ListProjects        多 list + label filter 聚合             常规错误
//   InspectProject      list + inspect(Labels)                    常规错误
//   Start / Stop / Restart   拉容器 + 循环 ContainerStart/Stop 常规错误
//   Down                拉容器/卷/网络 + 循环删(支持 -v/--rmi)   常规错误
//   Up                  label-聚合 + 循环 ContainerStart        常规错误(无 yaml 解析)
//   Run                 ContainerCreate(+oneoff label)+Start +Wait +Remove(per R08-08)
//   Exec                复用容器 exec                            常规错误
//   Top / Port / Stats  容器循环                                 常规错误
//   Events              events stream + label filter             常规错误
//
// **永久 ErrComposeUnsupported**(驱动层不自实现):
//
//   Config              daemon 无 yaml 合并端点,dtui 不引入 yaml 解析
//   Build(*)            daemon ImageBuild 需 Dockerfile context(TarReader),无 context 路径
//
// (*) Build / Pull / Push 不在本 interface,在 R08-05 build-transfer 域定义;
//     Build 永久 ErrUnsupported,Pull / Push 走 ImagePull / ImagePush 循环,
//     详见 R08-05。
type ComposeService interface {
    // 项目列表(含空项目)
    ListProjects(ctx context.Context) ([]ComposeProjectSummary, error)
    // 单项目详情
    InspectProject(ctx context.Context, project string) (ComposeProjectSummary, error)
    // 读取 compose 配置文件内容 — daemon 无 yaml 合并端点,driver 不自实现
    // 永远返回 ErrComposeUnsupported
    Config(ctx context.Context, project string) (string, error)
    // 启动已停止的服务
    Start(ctx context.Context, project string, services []string) error
    // 停止服务
    Stop(ctx context.Context, project string, services []string, timeoutSec int) error
    // 重启
    Restart(ctx context.Context, project string, services []string) error
    // 全项目下线(容器 + 网络,可选卷 / 镜像)
    Down(ctx context.Context, project string, opts DownOptions) error
    // 创建并启动(down 后的反向)— label 聚合 + 循环 start,无 yaml
    Up(ctx context.Context, project string, opts UpOptions) error
    // 一次性命令 — ContainerCreate(+oneoff label)+ Start + Wait + Remove
    Run(ctx context.Context, project string, service string, command []string, opts RunOptions) error
    Exec(ctx context.Context, project string, service string, command []string) error
    // 服务级查询
    Top(ctx context.Context, project string, service string) (runtime.ContainerProcesses, error)
    Port(ctx context.Context, project string, service string, port int) (string, error)
    Stats(ctx context.Context, project string, service string) (runtime.ContainerStats, error)
    // 实时事件流
    Events(ctx context.Context, project string) (<-chan ComposeEvent, error)
}

type DownOptions struct {
    RemoveVolumes    bool
    RemoveImages     string // "all" | "local" | ""
    RemoveOrphans    bool
    TimeoutSec       int
    ProjectTimeoutMs int
}

type UpOptions struct {
    Detach           bool
    Build            string // "" | "always" | "never"
    ForceRecreate    bool
    NoDeps           bool
    NoBuild          bool
    RemoveOrphans    bool
    Scale            map[string]int
}
```

### F2 不走子进程

禁止方法内实现包含 `exec.Command("docker", "compose", ...)` / `exec.Command("podman", "compose", ...)` / `exec.Command("podman-compose", ...)`。所有方法走 socket 上的 HTTP / REST API。

如某能力官方不支持(socket API 无对应端点),允许 `ErrComposeUnsupported` 错误返回。

### F3 适配器实现策略

| 能力 | Docker Engine API | Podman REST / libpod | 兜底 |
|---|---|---|---|
| `ListProjects` | 无原生 API(项目是 label 概念) | 无原生 API | **完全靠 label 聚合**(当前 dtui 已实现) |
| `InspectProject` | 同上 | 同上 | 同上 |
| `Start / Stop / Restart` | 无批量端点 | 无批量端点 | 容器 ID 循环 |
| `Down` | 无批量端点 | 无批量端点 | 容器 + 网络 + 卷 ID 循环 |
| `Up` | **无原生**:Compose CLI 本身解析 yaml + 调用 container API,**dtui 不做 yaml 解析** | 同上 | **声明 ErrComposeUnsupported**;UI 提示用户回到 docker compose CLI |
| `Run` | 同上 | 同上 | 同上 |
| `Top` | 循环容器 Top | 循环容器 Top | 容器级动作复用 |
| `Port` | 循环容器 + PortBindings 解析 | 同上 | 容器级 |
| `Stats` | 循环容器 Stats | 同上 | 容器级 |
| `Config` | 无:Compose CLI 解析 yaml | 同上 | **声明 ErrComposeUnsupported**(R08-03 给出只读元信息替代) |
| `Events` | docker events via filter `type=container,label=com.docker.compose.project=<p>` | podman events 同源 | R05 事件流机制 |

### F4 决策候选

| 编号 | 内容 | 优 | 劣 |
|---|---|---|---|
| **决策 A** | 不引入 `ComposeService`,R08-03 / R08-04 / R08-06 等实现继续在 `compose_*` 文件内**直接**拼装 Runtime 调用 | 不增加抽象 | UI 层重复拼装逻辑;up / config 这类需要新端点的能力,以后还会回来引入 |
| **决策 B(本推荐)** | 引入 `ComposeService`,但只覆盖"靠 label 聚合 + 容器级 API 循环可以完成"的能力(`ListProjects / Inspect / Start / Stop / Restart / Down / Top / Port / Stats`);`Up / Config / Run / Build` 返回 `ErrComposeUnsupported` 并在 UI 明示 | 抽象落地,有边界 | 抽象不完整,未来 up/config 还要讨论 |
| **决策 C** | 引入完整 `ComposeService` 实现 yaml 解析 + 真原生 up/config | 完整 | 巨大工程量,与 docker compose CLI 大量重叠语义,需要 yaml/v1+v2/v3+include + 调参表 |

### F5 推荐方案

**采纳决策 B**。理由:

- 解决 `Engine.Containers/Volumes/Networks/Images` 抽象 vs "Compose 是跨这些资源的批处理"之间的语义落差
- 不引入 yaml 解析,不与 docker compose CLI 形成语义二义
- up / build / run 这类依赖 yaml 的命令明确"unsupported",避免"看似能做实则半成品"
- 后续如决定支持 up,可创建独立 `R08-04-a compose-up-native` 决策而不是改动本接口

## 实现设计

涉及模块(决策 B 采纳后):

- 新文件:`internal/data/runtime/compose_service.go` — 接口与错误类型
- 新文件:`internal/data/runtime/docker/compose_service.go` — Docker 实现(全部走容器 / 卷 / 网络 API + label 聚合)
- 新文件:`internal/data/runtime/podman/compose_service.go` — Podman 实现(同策略)
- 新文件:`internal/data/runtime/compose_aggregator.go` — 兜底聚合实现(供 runtime 不支持时回退)
- 修改:`internal/data/runtime/engine.go:7-15` 加 `Compose() ComposeService` 方法
- 修改:`internal/data/runtime/docker/client.go`、`podman/engine.go` 接入 ComposeService
- 修改:`internal/data/runtime/capabilities.go` 加 `CapabilityComposeService` 标识

调用链:

```text
Engine.Compose() -> ComposeService
                       |
                       +-> DockerComposeService (label aggregation + container-level loops)
                       +-> PodmanComposeService (label aggregation + libpod container-level loops)
                       +-> AggregatedComposeService (fallback)
```

## 验收标准

1. `ComposeService` 接口与 `internal/data/runtime/capabilities.go` 的 capability 匹配
2. Docker / Podman 双 adapter 完整实现 `ListProjects / Inspect / Start / Stop / Restart / Down / Top / Port / Stats`,单元测试覆盖
3. `Up / Config / Run` 返回 `ErrComposeUnsupported`,UI 检测到后展示特定提示
4. `Down` 的 `RemoveVolumes / RemoveImages / RemoveOrphans` 三标志独立测试
5. 与 R08-11 Action Bar 配合:`OperationScopeCompose` 引入后,所有上层动作通过 `Engine.Compose()` 调用
6. 不引入 `exec.Command` 子进程

## 非目标

- yaml 解析(留给 R08-03 元信息子集)
- podman-compose 二进制检测(保留现状:`podman compose` 是包装器,docker-compose / podman-compose 同时存在)
- Docker Compose plugin 管理 / `compose version` 子命令
- Compose `config --resolve-image-` / `compose images` / `compose wait` 等次要能力

## R08-14 集成

`ComposeProjectSummary` 加 `CoLocatedGroups []CoLocatedGroup` 字段,详见 [R08-14](./R08-14-co-located-group.md) §I1。`ComposeService.InspectProject` 返回的 DTO 含此字段;`ComposeService.ListProjects` 聚合时填好 group 索引(`Source` 枚举见 R08-14 §F2)。

不修改 `runtime.Engine.Compose() ComposeService` 接口本身(group 只是 DTO 承载,新方法不加)。

## R08-15 集成(驱动实现指引)

`ComposeService` 是 API 规格,R08-15 是 docker / podman adapter 的实现路径。

| 主题 | R08-02(API) | R08-15(驱动) |
|---|---|---|
| Interface 形态 | 本需求 F1 定义 | (略) |
| docker adapter 实现 | (略) | [R08-15 §3](./R08-15-driver-implementation.md) |
| podman adapter 实现 | (略) | [R08-15 §3](./R08-15-driver-implementation.md) |
| Podman REST `/libpod/pods/*` 端点映射 | (略) | [R08-15 §1](./R08-15-driver-implementation.md) |
| docker inspect.N/P/IpcMode group 推导 | (略) | [R08-15 §4](./R08-15-driver-implementation.md) |
| docker 端 `Up/Build/Run` `ErrComposeUnsupported` 硬规则 | (略) | [R08-15 §3.5](./R08-15-driver-implementation.md) |
| Capability 报告 | (略) | [R08-15 §6](./R08-15-driver-implementation.md) |

R08-02 与 R08-15 是**接口-实现**配对。R08-02 评审通过 API 后,实现细节按 R08-15 §3 / §4 / §7 清单落地(在 docker / podman adapter 同名目录下新建文件)。

## 迁移记录

- 旧任务:`compose-todo.md 补 E.4` 验证 `Engine` 不含 ComposeService
- 旧决议:TASK-022 Phase E 取消 `docker/service` 统一入口(本需求不复活)
- 保留信息:Engine 单一上层抽象原则
- 废弃信息:无
- 待确认:**决策 A / B / C** 三选一(主模型评审)
