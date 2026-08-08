# R08-15 runtime driver 适配实现方案

## 元信息

- 状态: implementing
- 优先级: high
- 来源: 用户评审结果 2026-08-07 —"讨论 compose 底层驱动实现;podman 侧已确认,docker 侧怎么接?"
- 关联任务: 跨 R08-02(ComposeService API 规格)、R08-14(CoLocated 中间层)
- 关联约束: 无新约束
- 实现现状: 方案已落地 — 双 adapter `ComposeService` / `PodService` / `co_located.go` 已实现并接线(docker / podman),`CapabilityCompose*` 已报告;遗留:capability 状态复审、TUI Action Bar 集成

## ⚠ 重复功能说明

| 能力 | 已实现层 | 本需求处理 |
|---|---|---|
| docker SDK 标准容器/卷/网络/镜像操作 | `internal/data/runtime/docker/{containers,volumes,networks,...}.go` | **复用** |
| podman REST 标准容器/卷/网络/镜像操作 | `internal/data/runtime/podman/{service_container,service_volume,...}.go` | **复用** |
| podman REST `/libpod/pods/...` | swagger 端点见 §实现设计 1,**未在 dtui driver 层实现** | 本需求规划实现路径 |
| docker engine `/pods` | **不存在** | **不实现**,CoLocated 由 inspect 数据推导 |

## 目标

R08-02 是 `ComposeService` 接口的形态规格,R08-14 是 CoLocated group 的数据规格。本需求回答"**这两个规格在 docker 与 podman 两套**独立的**底层依赖上,driver 层怎么各自实现、runtime 层怎么统一抽象**"。

docker / podman **不是** dtui 的"两个孪生 backend" — 它们是两套**完全独立的底层依赖**:
- 各自有不同的 API 表面(docker Engine API vs podman REST/libpod)
- 各自有不同的能力(Pod 一等公民 vs 显式 namespace 共享)
- podman driver 内部 HTTP / CGO bindings 是同一 REST transport 的两种 client 实现选项(`CGO_ENABLED=0/1` 决定选哪种,与 daemon 无关)

但对 dtui 而言,**`ComposeService` 接口必须对称** — 不论底层是 docker 还是 podman,调用者看到一致语义;不能因为底层差异暴露在接口上。

具体:
1. docker / podman 各自 driver 层 **接口对称**(都暴露 `ComposeService`)、**实现路径不对称**(podman 多 Pod 事实源,docker 必须 label 聚合自实现)
2. podman driver 层 **单 transport = REST**;HTTP client 与 CGO bindings client 是该 transport 的两种实现,daemon 端无差别
3. docker driver 层 **单 transport = Docker SDK**,缺失能力靠 label 聚合 + 循环自实现
4. runtime 层结构体 / 字段 / action 名命名一致性(避免 carry-the-wrong-field 隐性 bug)

读者:写 docker 或 podman adapter 代码的人;review R08-02 / R08-14 实现设计的人。

## 用户流程

按 runtime 划分,dtui 用户在不同 runtime 下面临的实际可用操作不同:

### docker runtime

```
打开 dtui → Connect to docker socket
  → Compose 面板显示聚合项目(project=label)
    → 选中项目 → 详情显示 working_dir / config_files / version
    → Action Bar:
       ✅ Start / Stop / Restart / Down / Top / Port / Stats / Logs / Events
       ✅ Down (with -v / --rmi all / --remove-orphans)
       ⛔ Build / Pull / Push(单镜像可,batch 不支持)
       ⛔ Up / Run / Config → "用 docker compose CLI"
    → drill-in service 子视图:
       ✅ 容器级 start / stop / exec / logs / stats
       ✅ Group badge(若存在 NetworkMode 显式 declare 的 group)
```

### podman runtime

```
打开 dtui → Connect to podman socket
  → Compose 面板显示聚合项目
    → 选中项目:
       ✅ 全部 docker 端可用操作
       ✅ Pod badge(`pod_<project>` 自动)
       ⛔ Up / Build / Run / Config → 同 docker 端(原则统一,不偷懒)
    → Pod 面板(若 dtui 实现)直接列 `/libpod/pods/json`
```

→ 两个 runtime UI 看起来 90% 一致,关键差异在 Pod/CoLocated group 标识源(自动 vs 显式)+ Build/Up/Run 的硬限制。

## UI/UX

仅限本需求特有的 capability-driven 行为:

- Action Bar 在 docker 端 mount point:`compose_project_build` / `compose_project_pull` 项灰显 + tooltip 提示 "Docker engine has no compose project-level action; use docker compose CLI"
- `compose_project_up` / `compose_project_run` / `compose_project_config` 永远灰显 + tooltip "Use docker compose CLI"
- `compose_group_*` 项(R08-14)在 docker 端 visible 但大部分会因 `CapabilityComposePodScope = Unsupported` 而自动 hidden
- podman 端相同项 visible,go-enabled(实际行为依赖具体 runtime capability)

UI 行为用 R08-11 §F4 Already-Implemented 的 `Requirements / DisabledWhen` 机制,而**不**专门写 "docker 端不允许" 的硬编码规则。

## 功能规则

### F1 docker 侧"硬规则"

> **docker 引擎没有 `compose up / build / run / pull(policy=batch) / push(policy=batch)` 等 yaml 化命令**。
>
> dtui 不 shelling out 到 `docker compose` CLI(违反 Engine API 单交互面原则,与 R08-02 决策 B 一致)。
>
> 这几个方法在 docker adapter 上**永远**返回 `ErrComposeUnsupported`,**无逃生口**。

### F2 podman 侧"硬规则"

Podman 同:也不做 yaml 解析,`Up / Build / Run / Config` 返回 `ErrComposeUnsupported`。但 `Pods()` 完整可用(因为 swagger 端点齐)。

### F3 CoLocated group 触发条件

- docker 端:任一 `HostConfig.NetworkMode / PidMode / IpcMode` 以 `"service:<name>"` 或 `"container:<id>"` 开头,且被 2+ 容器引用时建 group
- podman 端:所有容器 `inspect.Pod` 字段非空时,自动归属 `pod_<project>`
- 详见 [R08-14 §F2](./R08-14-co-located-group.md)

### F4 inspect 性能 → Q2

详见 [Q2](#q2colocated-group-显示性能docker-端-inspect-n1) 决策 C':N+1 inspect + 5s TTL + docker events 失效。**F4 不再做独立决策**。

docker adapter 实现需注意:
- inspect 只取 `HostConfig.NetworkMode/PidMode/IpcMode` 3 字段(其他字段丢弃)
- goroutine fan-out 限 16,errgroup 风格
- cache key = project,LRU max=64,TTL=5s
- events 失败 → 退化 TTL,UI 不卡

podman adapter 不适用此机制(inspect.Pod 在 list 阶段可见,无需 inspect 风暴)。

### F5 Capability 报告(总是)

两侧 client.`Capabilities()` 都必须报告 `CapabilityCompose*` 系列。UI 不读 capability 时**默认最严**(docker 端的 Up 等在 capability 缺失时直接不显示)。

## 实现设计

### 0. 驱动层架构模型(分层硬性)

dtui 在 docker 与 podman 上分层如下:

```
┌─────────────────────────────────────────────────────────────┐
│ dtui 业务层(tui/...)                                          │
│   看到 runtime.Engine 接口,不感知 docker / podman 底层       │
└────────────────────────────┬────────────────────────────────┘
                             │ 调接口
┌────────────────────────────▼────────────────────────────────┐
│ runtime 层(data/runtime/)— 统一接口 + 共享 DTO              │
│   - Engine{ Containers() / Compose() / Pods() / ... }        │
│   - 各 Service 是 interface(ComposeService 等)               │
│   - 共享类型 ContainerSummary / ComposeProjectSummary / ...  │
└────────────────────────────┬────────────────────────────────┘
                             │ 实现
┌────────────────────────────▼────────────────────────────────┐
│ adapter 层(data/runtime/{docker,podman}/)                   │
│   - docker.NewPodmanEngine 不调 daemon / SDK,调 docker driver│
│   - podman.NewPodmanEngine 同,调 podman driver               │
│   - 不重复业务逻辑(委派给 driver),只做接口→driver 适配       │
└────────────────────────────┬────────────────────────────────┘
                             │ 调
┌────────────────────────────▼────────────────────────────────┐
│ driver 层(driver/{docker,podman}/)— 与 daemon 直连           │
│   - docker driver: 单 transport(只走 Docker SDK)             │
│   - podman driver: 单 transport(只走 podman REST API),       │
│     daemon 端 API 等价于 swagger 中所有 /libpod/* 端点         │
│                                                              │
│   单 transport 内的"实现选项"(不影响 dtui 视角):             │
│     podman client 可用 HTTP 或 CGO bindings 调到 daemon,     │
│     两路最终都打 podman REST API;driver 包内自带抽象          │
└────────────────────────────┬────────────────────────────────┘
                             │ Docker SDK / podman REST API
┌────────────────────────────▼────────────────────────────────┐
│ daemon 层                                                  │
│   - docker daemon /var/run/docker.sock                  │
│   - podman daemon /run/podman/podman.sock                 │
└─────────────────────────────────────────────────────────────┘
```

**硬性规则**:

1. runtime 层 **零容错**地不感知 docker / podman 差异 — 只看到 interface
2. adapter 层**不**直接调 SDK / daemon / SDK+daemon,**只**调 driver — 防止 driver 实现策略变化时 adapter 重写
3. docker driver 层 `compose` 语义靠**自实现**(daemon 端不存在原生 compose 原子操作)
4. podman driver 层**单 transport**(REST),`HTTP` 与 `CGO bindings` 是**单 transport 内的两种 client 实现**,都被 driver 包吸收

→ adapter 与 driver 必须**严格分层**;违背这条会把 docker/podman 差异推给上层。

### 1. podman driver 层(单 transport = REST)

#### 1.1 单一 transport:podman REST API

podman daemon **只接受** REST API(`docs/api/podman-swagger.yaml` 已确认),daemon 端**没有**别的 transport。dtui 角度下:**只有一条 transport** — `/libpod/...`。

**客户实现选项**(dtui 视角下不可见):

| Client 实现 | 文件 | 触发 | 内部通道 |
|---|---|---|---|
| `RESTClient`(HTTP) | `internal/driver/podman/rest.go` | 默认,无 CGO 依赖 | HTTP 调用 daemon |
| `bindingsClient`(CGO) | `internal/driver/podman/driver_cgo.go` | `CGO_ENABLED=1` 编译 + 链接 `go.podman.io/podman/v6/pkg/bindings` | in-process 调本机 lib,而 lib 内部仍以 HTTP/Socket 打 daemon |

→ **两个 client 实现都打同一 daemon 同一组 REST 端点**。差别在延迟与依赖(CGO 更省一次进程间跳,但仍走相同的协议)。

#### 1.2 driver 层抽象(就一个 transport,两种 client 实现)

```go
// podman driver 内的 client interface:
type Client interface {
    Containers() ContainersService
    Volumes()    VolumesService
    Networks()   NetworksService
    Images()     ImagesService
    Pods()       PodsService                 // 见 §1.3
    // ... 对接 runtime 层的所有 service
}

// 两种实现 — 单 transport 的两种 client
type RESTClient    struct { ... }            // HTTP → /libpod/...
type bindingsClient struct { ... }           // CGO → bindings → /libpod/...

// 工厂:
func NewClient(config Config) (Client, error) {
    if config.CGOEnabled { return newBindingsClient(config) }
    return newRESTClient(config)
}
```

**硬规则**:

- adapter 层只看到 `Client` interface,**不**知道 client 是 HTTP 还是 bindings
- 这层抽象的目的是 *测试 mock* + *替换 client 实现策略*,**不是** transport 多样性 — 因为 transport 本来就一个
- podman `NewPodmanEngine` 接受 `Client` interface 而非具体类型

#### 1.3 Podman REST 端点(swagger 已确认)

`docs/api/podman-swagger.yaml` 暴露 Pod 一等公民全套:

| REST 端点 | 用途 |
|---|---|
| `GET /libpod/pods/json` | 列 Pod |
| `GET /libpod/pods/{name}` | 检查存在 |
| `GET /libpod/pods/{name}/json` | Pod 详情 |
| `POST /libpod/pods/create` | 创建 Pod |
| `DELETE /libpod/pods/{name}` | 删 Pod |
| `POST /libpod/pods/{name}/start` | 启动 Pod |
| `POST /libpod/pods/{name}/stop` | 停止 Pod |
| `POST /libpod/pods/{name}/restart` | 重启 Pod |
| `POST /libpod/pods/{name}/kill` | 强杀 Pod |
| `POST /libpod/pods/{name}/pause` | 暂停 Pod |
| `POST /libpod/pods/{name}/unpause` | 恢复 Pod |
| `POST /libpod/pods/prune` | 批量清理 |
| `GET /libpod/pods/{name}/top` | Pod 内 top |
| `GET /libpod/pods/stats` | Pod 级 stats |

`bindings` 路径用同样的方法名(call `bindings.pods.List(ctx, ...)`,`bindings.pods.Start(ctx, name, ...)`),只是 GOCALL 替代 HTTP。

#### 1.4 podman docker compose 适配边界

`podman compose` / `podman-compose` 是 daemon 端的**独立工具**(不通过 daemon API),**dtui 不 shelling**:
- podman daemon 提供容器 / 卷 / 网络 / Pod 原语
- compose 语义在 daemon 端没有原生支持,与 docker daemon 同;dtui driver 层同样**自实现**(label 聚合 + 循环)
- 但 podman 比 docker 多了 **Pod 事实源**(容器 `.Pod` 字段),CoLocated group 推导更优

→ **compose 端:docker / podman driver 层实现策略 100% 对称**(都自实现),**唯一差别**:podman 端 CoLocated 推导效率更高(见 §4)。

### 2. docker driver 层(单 transport)

#### 2.1 单 transport 路径

Docker SDK `github.com/docker/docker/client` 是**唯一** transport 路径 — 没有 CGO 等价物,也没有 REST 替代。

| 路径 | 文件 |
|---|---|
| Docker SDK 唯一 | `internal/driver/docker/client.go`(间接通过 Docker Go SDK 包) |

→ docker driver 层**没有 transport 二选一抽象**问题,但仍要保持单一 `dockerClient` interface,方便测试 mock。

#### 2.2 docker driver 与 daemon API 的能力映射

| 能力 | Docker Engine API | 说明 |
|---|---|---|
| 容器按 label filter list | `GET /containers/json?filters={"label":"com.docker.compose.project=X"}` | ✅ 原生 |
| 容器 inspect with `HostConfig.NetworkMode` | `GET /containers/{id}/json` | ✅ `docker/inspect.go:55` 有该字段 |
| 容器 inspect with Pid/Ipc mode | 同上 | ✅ `HostConfig.PidMode / IpcMode` |
| 网络按 label filter list | `GET /networks/json?filters={"label":"com.docker.compose.project=X"}` | ✅ 原生 |
| 卷按 label filter list | `GET /volumes?filters={"label":"com.docker.compose.project=X"}` | ✅ 原生 |
| 事件按 label filter stream | `GET /events?filters=label=X` | ✅ 原生,R05 已用 |
| `/pods` | **不存在** | ❌ 项目级 entity 抽象 |
| 项目级原子 start/stop/restart | **不存在** | ❌ 只能逐容器循环 |
| `compose up / build / run / config` 原生命令 | **不存在** | ❌ compose-spec 行为依赖 yaml |

→ 与 podman 端一样,docker daemon 不提供 compose 语义;dtui docker driver 层**自实现**(label 聚合 + 循环,见 §3)。

### 3. docker 侧 ComposeService 实现路径

#### 3.1 `ListProjects`

完全依靠现有容器/卷/网络 list。

```text
docker.Containers().List(filters={"label": "com.docker.compose.project"})
  → 客户端聚合 project 字段 → ListProjects
docker.Networks().List(same filter)    → 补 networks 计数
docker.Volumes().List(same filter)     → 补 volumes 计数
```

#### 3.2 `InspectProject`

元信息从容器 Inspect 的 `Config.Labels` 读:

- `com.docker.compose.project.working_dir`
- `com.docker.compose.project.config_files`(多值,以 `\\n` 分隔)
- `com.docker.compose.version`

任取一匹配容器 Inspect(`cli.ContainerInspect(ctx, id).Config.Labels`)即可。Docker SDK 直接给。

#### 3.3 `Start / Stop / Restart(project, services)`

按 spec 行为:list 容器 + 循环调 SDK。

```text
Containers.List(filter=project=X AND state=exited)  → for each: ContainerStart(ctx, id)
Containers.List(filter=project=X AND state=running) → for each: ContainerStop(ctx, id)
```

R08-04 当前 `doComposeStart / Stop` 已按此模式写。R08-15 不新增实现,只确认模式。

#### 3.4 `Down(project, opts)`

```text
opts.RemoveVolumes=true:    Volumes.List(label=X)        → for each: VolumeRemove(force)
opts.RemoveImages="all":    ImageList(repos match project containers) → ImageRemove
opts.RemoveOrphans=true:    Containers orphan 检测(见 §5)
                            Containers.List                 → for each: ContainerRemove(force)
Containers:                  for each: ContainerRemove(force)
Networks:                    for each: NetworkRemove(label=X)
```

`--remove-orphans` 的孤儿检测见 §5。

#### 3.5 `Up / Run / Build / Pull / Push / Config` — **ErrUnsupported 硬规则**

> 这几个方法在 docker adapter 上**永远**返回 `ErrComposeUnsupported`。

无逃生口。无 yaml 解析路径。

#### 3.6 `Top / Port / Stats(project, service)`

按容器循环聚合:

```text
top:    for each ContainerTop(ctx, id) → 聚合 title + processes
port:   for each ContainerInspect(ctx, id).NetworkSettings.Ports → 聚合去重
stats:  for each ContainerStats(ctx, id) → 启动 stream + 合并 channel
```

#### 3.7 `Events(project)`

复用 R05 现有 docker events stream + label filter。**无新增实现**。

### 4. docker 侧 CoLocated group 推导(无 Pod 端的实现)

#### 4.1 数据来源

| 阶段 | 字段 | 来源 |
|---|---|---|
| List | `c.Labels["com.docker.compose.project"] / service` | Docker SDK `ContainerList` 已含 |
| Inspect | `HostConfig.NetworkMode` | Docker SDK `ContainerInspect` |
| Inspect | `HostConfig.PidMode` | 同 |
| Inspect | `HostConfig.IpcMode` | 同 |

#### 4.2 推导算法

```text
1. List 一次 project=X 容器集合
2. 触发每容器 Inspect(fan-out goroutine)
3. 收集所有 Inspect.NetworkMode / PidMode / IpcMode 中以 "service:<name>" / "container:<id>" 开头的引用
4. 建索引 target_name → [source_container_ids]
5. 阈值判断:target 上 ≥2 source 引用 → 建 CoLocated group
6. group id 形如 "<project>/<target_name>"
7. Source 枚举:
   - "docker_network_mode_service"   (net 共享)
   - "docker_pid_service"            (pid 共享)
   - "docker_ipc_service"            (ipc 共享)
   - "docker_container_ref"          (NetworkMode == "container:<id>")
```

#### 4.3 性能取舍

| 方案 | RTT 次数 | 实现成本 | 推荐 |
|---|---|---|---|
| 每容器 1 次 Inspect(N+1) | 项目 30+ RTT | 低 | **采用**,goroutine 并发 + 缓存 |
| Engine filter `?include=HostConfig.NetworkMode` | 不存在 | 不可行 | ❌ |
| 5s TTL cache + events trigger 失效 | N+1,cache 后 0 | 中 | 推荐作为常驻 |

→ **接受 N+1 inspect + 5s cache**,不引入新 Docker API。

#### 4.4 docker vs podman CoLocated group 语义对比

| 维度 | Docker | Podman |
|---|---|---|
| 数据来源 | inspect.NetworkMode/PidMode/IpcMode(显式 declare) | inspect.Pod 字段(自动) |
| List 阶段可见 | ❌ 需 Inspect | ✅ Pod 字段直接读 |
| 准确度 | 声明式 | 事实式 |

UI 必须**自适应**这个差异,不假设两 runtime 看到的 group 数一致。

### 5. 高级 — orphan detection

#### 5.1 Compose 端定义

`docker compose down --remove-orphans` 把"配置已变化但仍存在的容器"删掉。

#### 5.2 dtui 当前实现

**完全没有 orphan 检测**。`doComposeDown` 直接删所有匹配 label 的容器。

#### 5.3 推荐的简化算法

> 当项目内某容器的 `com.docker.compose.config-hash` label 与当前 config-hash 不匹配 → orphan。

退化:orphan = "容器带 `com.docker.compose.project=X` label, 但 `com.docker.compose.service` 不在该项目其它容器宣告的 service 集合内"。

**是否引入 yaml 解析**?**不**——继续简化算法,放弃边角情况。

### 6. Capability 报告策略

`CapabilityCompose*` **按 driver 路径 × 端点能力**两维报告。adapter 层做合并上抛 — runtime 层看到的 `CapabilitySet` 由 adapter 决定,与底层 transport(REST vs CGO)无关。

```text
Docker adapter 报告(单 transport):
    CapabilityComposeAggregate         Available      // List/Start/Stop/Restart/Down/Top/Port/Stats/Events
    CapabilityComposeProjectLevel      Degraded        // 项目级只能聚合循环,无原子性
    CapabilityComposeUp                Unsupported     // 硬 unknown(docker daemon 无)
    CapabilityComposeBuild              Unsupported
    CapabilityComposeRun                Unsupported
    CapabilityComposeConfig             Unsupported
    CapabilityComposePull               Degraded         // 单镜像可,batch 是循环
    CapabilityComposePush               Degraded
    CapabilityComposePodScope           Available         // ✓ docker driver 自实现:wrap inspect.NetworkMode/PidMode/IpcMode
    CapabilityComposeCoLocated          Available         // driver 内同源算法,DTO 不同

Podman adapter 报告(transport-agnostic 合并):
    CapabilityComposeAggregate         Available
    CapabilityComposeProjectLevel      Available         // Pods/json + container Pod 反查
    CapabilityComposeUp                Unsupported        // 同原则,daemon 端也不支持 yaml
    CapabilityComposeBuild              Unsupported
    CapabilityComposeRun                Unsupported
    CapabilityComposeConfig             Unsupported
    CapabilityComposePull               Degraded
    CapabilityComposePush               Degraded
    CapabilityComposePodScope           Available          // /libpod/pods
    CapabilityComposeCoLocated          Available          // inspect.Pod 字段直接读,零 RTT
```

注:docker 与 podman **两侧**的 PodScope 都是 Available,**实现路径不同**:
- podman:daemon 原生 `/libpod/pods/...`
- docker:driver 自实现 wrap inspect.NetworkMode

`runtime.Engine.Pods()` 接口**统一** — 上层不感知实现差异。

`Up / Build / Run / Config` 两侧都标 Unsupported(daemon 端都无,driver 层也不自实现 yaml)。

### 7. Adapter 实现清单(只列,不写)

按 §0 分层模型拆:**runtime → adapter → driver**,adapter 只委派不调 SDK。

#### 7.1 runtime 层

| 任务 | 入口文件 | 改动 |
|---|---|---|
| `NetworkMode / PidMode / IpcMode` 暴露到 `ContainerSummary` | `runtime/containers.go` | 加 3 字段 |
| engine 接口加 `Compose() ComposeService` | `runtime/engine.go:7` | 加方法 |
| engine 接口加 `Pods() PodService`(**两端各自实现,接口统一**) | `runtime/engine.go:7` | 加方法 |
| 加 CapabilityCompose 系列常量 | `runtime/capabilities.go` | 加 9 个 Capability 标识 |

#### 7.2 adapter 层 — docker

| 任务 | 入口文件 | 改动 |
|---|---|---|
| `dockerComposeService` | `runtime/docker/compose_service.go` 新文件 | 实现 §3 算法,**只委派给 docker driver**,不调 SDK |
| `dockerCoLocatedResolver` | `runtime/docker/co_located.go` 新文件 | 实现 §4 算法 |
| **`dockerPodService`**(Q6 决策) | `runtime/docker/pod_service.go` 新文件 | **自实现 wrap** `ContainerList` + per-container `ContainerInspect`(NetworkMode/PidMode/IpcMode),不报 ErrUnsupported,接口层统一 |
| `dockerClient` interface 抽象 | `internal/driver/docker/client.go` | 包内小 interface 用于 mock 测试 |
| `docker `Up/Build/Run/Config` 实现 | 同 | hard return `ErrComposeUnsupported`(**这四项** daemon 端真没有,driver 层不自实现 yaml) |

#### 7.3 adapter 层 — podman

| 任务 | 入口文件 | 改动 |
|---|---|---|
| `podmanComposeService` | `runtime/podman/compose_service.go` 新文件 | 调 list /inspect,**复用**容器聚合;只委派给 podman driver |
| `podmanCoLocatedResolver` | `runtime/podman/co_located.go` 新文件 | 读 `inspect.Pod` |
| `podmanPodService` | `runtime/podman/pod_service.go` 新文件 | 薄包装 `/libpod/pods/*`(driver 层已转译 REST 与 bindings) |
| **关键**:Podman engine 构造接受 `podmanClient` interface 而非具体类型 | `runtime/podman/engine.go` | `NewPodmanEngine(client podman.Client)` 让 driver 层选 REST/bindings |

#### 7.4 driver 层 — podman 单 transport + 内部 client 实现

| 任务 | 入口文件 | 改动 |
|---|---|---|
| `Client` interface | `internal/driver/podman/` 新文件如 `interface.go` | 单 transport 下统一抽象,所有 service 一个入口 |
| `RESTClient`(HTTP)实现 | 已有 `rest.go` | 实现 `Client` interface,默认 |
| `bindingsClient`(CGO)实现 | 已有 `driver_cgo.go` | 实现 `Client` interface,`CGO_ENABLED=1` 时启用 |
| `NewClient(config)` 工厂 | `internal/driver/podman/factory.go` 新文件 | 按 `config.CGOEnabled` 选 client;**两条都打同一 daemon 的 REST API** |
| CGO_ENABLED 构建标签隔离 | `internal/driver/podman/driver_cgo.go` 已有 build tag | 维持 |
| Capability 报告 | `runtime/docker/client.go` + `runtime/podman/engine.go` | **adapter 层不暴露 client 实现**(HTTP vs CGO)是 driver 内部选择 |

#### 7.5 docker 层

| 任务 | 入口文件 | 改动 |
|---|---|---|
| Capability 报告加 `CapabilityCompose*` | `runtime/docker/client.go:197` | 加 CapabilityCompose 系列 |

### 8. 命名约定(防止 carry-the-wrong-field)

#### 8.1 类型命名

| 层 | 类型 | 命名示例 |
|---|---|---|
| runtime service interface | `<Resource>Service` | `ContainerService`, `ComposeService`, `PodService` |
| runtime 共享 DTO | `Runtime<Resource>` | `RuntimeContainerSummary`(避免与 driver 私有类型同名) |
| driver interface(内部) | `<runtime>Client` | `podmanClient`, `dockerClient`(小写不 export,与 driver 包绑死) |
| driver SDK 私有类型 | 第三方原样 | `*client.Client`(Docker SDK), `*podman.RESTClient`(本项目自封装) |

#### 8.2 字段命名一致性

跨 driver 流动的字段**只在 `runtime` 包定义一次**,driver 只填充:

```go
// 错误示范 — 每个 driver 自己定义
type docker.ContainerSummary struct{ ComposeProject string }
type podman.ContainerSummary struct{ ComposeProject string }

// 正确 — runtime 包定义,driver 填充
type runtime.ContainerSummary struct {
    ComposeProject      string
    ComposeService      string
    CoLocatedGroupID    string  // R08-14
    CoLocatedGroupSrc    string  // R08-14
    WorkingDir           string  // R08-01
    ConfigFiles          []string
    Version              string
    ConfigHash           string
    ContainerNumber      string
    Oneoff               bool
    // ...
}
```

#### 8.3 网络 / 协议层命名

| 项 | 命名 |
|---|---|
| Compose 项目发现 query key | `com.docker.compose.*`(podman-compose 也写这组 — 见 §1) |
| `docker-compose` 与 `podman-compose` 同时使用 | 统一读 `com.docker.compose.*`,**不**特化 `io.podman.compose.*` |
| Lifecycle action 名(`runtime/actions.go`) | `container.start` / `container.stop`,runtime-agnostic |
| Compose action 名(`runtime/api`) | `compose_project.start` / `compose_service.logs` 等 |
| Pod 资源 namespace(仅 podman) | `/libpod/pods/...` |

#### 8.4 CoLocated Source 枚举(R08-14 §F2)

```text
"podman_default_pod"           # podman 自动 1 项目 1 pod(facts)
"docker_network_mode_service"  # docker 显式 declare net
"docker_pid_service"
"docker_ipc_service"
"docker_container_ref"
"unrecognized"
```

## 待确认(主模型决策)

### Q1:实现顺序

> **架构原则(user 拍板)**:podman / docker adapter **完全依赖各自 driver 或 REST 约束**(不共享统一接口层)。**统一接口只在 runtime 层**(`engine.go` 的 `Compose()` / `Pods()` / `ContainerService` 接口)。

#### 问题场景

| 场景 | 当前协调风险 | 痛点 |
|---|---|---|
| **S1 字段定义先于两侧实现** | docker / podman adapter 并行开发,字段名会自然发散(`composeProject` vs `ComposeProject`,`netMode` vs `NetworkMode`) | 中途发现 carry-the-wrong-field bug,需要双向 PR 修复。增加 review 轮次 |
| **S2 interface 加方法必须 non-breaking** | `engine.go` 当前已稳定 9 个 `*Service` 接口。加 `Compose() / Pods()` 一旦破坏旧调用,会触发全工程回归测试 | 旧 adapter / 旧测试要重写 |
| **S3 docker 与 podman 复杂度差距大** | docker 端需 wrap 自实现(Pods 拉 inspect,Up 拉 labels,Up/Run/Pull/Push 都要做)。podman 端 daemon 直调简单 | 串行会让 podman 阻塞 docker,或反之——浪费其中一个完工作 |
| **S4 driver 层独立原则** | 不应在 driver 层做"统一接口"——这违反 §0 硬规则 | driver 层一旦被统一接口定义约束,podman / docker 独立性丢失 |

#### 需求契约(可验证)

| ID | 需求 | 验证 |
|---|---|---|
| **N1 字段非破坏** | `ContainerSummary` 新字段全部 `omitempty` JSON tag,容器序列化 / 反序列化与旧数据兼容 | go test:序列化旧字段的新 ContainerSummary 不出错 |
| **N2 接口加方法非破坏** | `engine.go` 加 `Compose() ComposeService` 与 `Pods() PodService` 是 additive,旧 Engine 实例仍能编译 / 运行 | go test:旧调用方代码不需修改 |
| **N3 两侧并行不阻塞** | docker adapter 与 podman adapter **无相互依赖**,各自独立 PR 合并 | git log:可分别 commit,合并无序 |
| **N4 driver 层独立** | podman driver / docker driver **互不调用**,无共享 interface | grep:driver 包之间无跨 import |

#### 决策 — **方案 A**(数据先行 → 接口 stub → 两侧并行)

| Phase | 内容 | 满足需求 |
|---|---|---|
| **A1 — 数据基础** | `ContainerSummary` 加字段:`NetworkMode / PidMode / IpcMode / WorkingDir / ConfigFiles / Version / ConfigHash / ContainerNumber / Oneoff / ComposeProject / ComposeService / CoLocatedGroupID / CoLocatedGroupSrc` | N1 |
| **A2 — 字段填充** | docker adapter + podman adapter 同时从各自数据源填字段(docker:`ContainerList.Labels` + `ContainerInspect.NetworkMode`;podman:`ContainersList.Labels` + `InspectContainerResponse.Pod`) | N3,N4 |
| **B1 — engine 接口加方法** | `engine.go` 加 `Compose() ComposeService` 与 `Pods() PodService`,初始返回 `nil` 或 panic 均可 | N2 |
| **B2 — Capability 常量** | `capabilities.go` 加 `CapabilityCompose*` 系列(9 个) | N2 |
| **C — 两侧并行服务实现** | docker 与 podman adapter 同时实现各自的 `ComposeService` + `PodService` | N3 |

#### 非选项(原提案为何错误)

之前的提案 "podman 先 Composes,Pods 后置" — 在 Q6 落地前看是合理的(那时 docker PodsService 报 ErrUnsupported);但 Q6 让 docker PodsService 也必须自实现,**就没有"谁后置 Pods"的必要** — 两 adapter PodsService 必须同时存在。

#### 落地开 task 序

`docs/task/` 下应建 4 个并行 task 卡(可同时开工):
- `br-compose-data-fields.md`(A1)
- `br-engine-compose-pods-interface.md`(B1 + B2)
- `br-docker-compose-pods.md`(C — docker 侧)
- `br-podman-compose-pods.md`(C — podman 侧)

四个 task 卡可**同时开工**(A1 + B1 + 两侧 C 在 PR 上独立,但 CI 必须过)。

#### 与 Q3 的对齐

Q1 C 阶段"两侧 adapter 同时实现"必须包含 Q3 决策表里 6 个命令(Up/Run/Pull/Push wrap-able;Build/Config ErrUnsupported)。**两 adapter 同步给出 Up / Run / Pull / Push 实现**,而不是先 Composes 后 Pods,因为 Q6 已经把 Pods 自实现绑死在两 adapter 必交付项里。

### Q2:CoLocated group 显示性能(docker 端 inspect N+1)

#### 问题场景

| 场景 | 当前实现 | 用户体感 | 痛点 |
|---|---|---|---|
| S1 大项目冷启动(30 services) | 30 inspect RTT 串行 → goroutine fan-out | 服务列表先出,group 标签 ~200-500ms 延迟才出现 | 页面"看起来在加载",鼠标可能反复点;无障碍差 |
| S2 频繁切回 Compose 面板 | 每次重做完整 inspect 风暴 | 累积延迟,用户被迫等 | 切面板的预期是即时反映 |
| S3 活动 compose 实时性(用户在另终端 `docker compose up -d`) | dtui 端缓存到下次完整刷新 | group 标签"落后"daemon 几分钟 | 用户对 dtui 失去信任,去用 docker compose CLI |
| S4 daemon 资源压力 | 30 并发 inspect,daemon 短暂 high load | daemon 端可能 throttle,导致 dtui 限速 | docker engine API 自身保护 |

#### 需求契约(可验证)

| ID | 需求 | 验证方法 |
|---|---|---|
| N1 冷启动延迟 | 30-service 项目,Compose 面板冷启动 ≤ 500ms | benchmark + audit log |
| N2 缓存命中 | 同一项目 5s 内重复切回 ≤ 100ms | benchmark |
| N3 实时性 | 容器变化事件(create/start/die/stop/destroy)→ CoLocated ≤ 1s 重算 | 测试:启动新容器,等 1s,group 出现 |
| N4 鲁棒性 | events handler 故障时,UI 不卡,退化 5s TTL | 注入 events handler 故障,UI 仍 5s 内更新 |
| N5 资源有界 | 缓存按 project 数有界,LRU max=64;并发 inspect ≤ 16 | 压测 |

#### 决策 — **C'(修订)**

**N+1 inspect(docker 端专用,只取 `HostConfig.NetworkMode/PidMode/IpcMode` 3 字段) + 5s TTL 缓存 + docker events 失效**

| 需求满足映射 |
|---|
| N1 冷启动 ≤ 500ms — 5s TTL 内 hot path 直接命中 |
| N2 缓存命中 ≤ 100ms — 同 |
| N3 实时性 ≤ 1s — events 触发失效重算 |
| N4 鲁棒性 — TTL 兜底,events 故障不阻塞 UI |
| N5 资源有界 — TTL+LRU max=64,inspect fan-out max=16 |

#### 实施

- `runtime/docker/co_located_cache.go` — LRU + TTL cache(key = `project`)
- `runtime/docker/co_located_events.go` — 订阅 `/events?filters=type=container`,事件触发 `cache.Invalidate(project)`
- inspect fan-out:`errgroup.WithContext` 限 16 goroutine,单次只取 3 字段(`HostConfig.NetworkMode/PidMode/IpcMode`)
- podman 端**不适用**(inspect.Pod 在 list 阶段已有,见 §1.4 / R08-14 §F2),无需 N+1 inspect 风暴

#### 范围

- F2 / F3 / F4 / F5(§功能规则)的 inspect 性能假设全部交由 Q2 决策
- 验收标准 #15 显式覆盖 Q2 contract

### Q3:Compose-级 生命周期 / 镜像 / 配置 命令的自实现边界

> **原则**(per Q6 + §0 硬规则 #3):daemon 端没有的 ≠ 必然 ErrUnsupported —— driver 层能 wrap 自实现就自实现,只有真正无法 wrap 的能力才永久 ErrUnsupported。

#### 问题场景

| 场景 | 用户动作 | 当前 dtui 行为 | 痛点 |
|---|---|---|---|
| **S1 Up** | 用户在另一终端 `docker compose down`,再回 dtui 想 up | 报 ErrUnsupported,用户被迫回 docker compose CLI | dtui 只支持 down 不支持 up,语义不对称 |
| **S2 Run** | 用户想 exec 一次性命令进 web 服务(如 `web.1` 跑 migrate) | 报 ErrUnsupported | 现有容器 exec 已能跑,但项目级入口缺失 |
| **S3 Pull/Push** | 镜像版本更新(registry tag 推),想批量拉/推 | 报 ErrUnsupported | dtui 已经能单镜像 pull,push service level 不支持 |
| **S4 Build**(Dockerfile) | 改了 Dockerfile,想 rebuild | 报 ErrUnsupported | daemon `ImageBuild` 需 TarReader context,dtui 端无路径 |
| **S5 Config** | 想看 docker compose config 渲染结果(或 `--compatibility` 合并后) | 报 ErrUnsupported | daemon 端无 yaml 合并端点;dtui 不引入 yaml 解析器(per §0 硬规则) |

#### 需求契约(可验证)

| ID | 需求 | 验证 |
|---|---|---|
| **N1 Up** 可执行,不需要 yaml 解析 | daemon 上 label 匹配的全部 stopped 容器 `ContainerStart`;不读 compose.yaml | 单元测试:给一组 label=project 容器,Up 后 0 stopped 容器 |
| **N2 Run** 可执行,带 `com.docker.compose.oneoff=true` label | ContainerCreate + ContainerStart + ContainerWait + ContainerRemove(per R08-08) | 单元测试:R08-08 已规划 |
| **N3 Pull** 批量可执行,逐 service | loop ImagePull(name),去重(同 image 多 service) | 单元测试:R08-05 |
| **N4 Push** 批量可执行,逐 service | loop ImagePush(ref),无认证时跳过 + 提示 | 单元测试:R08-05 |
| **N5 Build** daemon 真没有 yaml 上传端点,自实现不现实 | ❌ 必须返回 ErrUnsupported | 暂不补 |
| **N6 Config** daemon 真没有 yaml 合并端点,dtui 不引入 yaml 解析器 | ❌ 必须返回 ErrUnsupported | 暂不补 |

#### 决策 — **按命令分轨**

| 命令 | 决策 | 依据 |
|---|---|---|
| **Up** | ✅ Available — driver 层 label-聚合 + 循环 ContainerStart | daemon 上 label 仍存,容器仍在(可 start) — **不需要 yaml** |
| **Run** | ✅ Available — driver 层 ContainerCreate(+ oneoff=true label)+ Start + Wait + Remove | R08-08 已规划完整路径 |
| **Pull** | ✅ Available — driver 层循环 ImagePull(service image),去重 | R08-05 已规划 |
| **Push** | ✅ Available — driver 层循环 ImagePush(ref),无认证时跳过 + 提示 | R08-05 已规划 |
| **Build** | ❌ **永久 ErrComposeUnsupported** | daemon ImageBuild 需 Dockerfile context(TarReader),**无 context 路径**;yaml 解析不引入(per §0 硬规则) |
| **Config** | ❌ **永久 ErrComposeUnsupported** | daemon 无 yaml 合并端点;dtui 不引入 yaml 解析器 |

> **原则统一**:能 wrap 的 wrap,不能 wrap 的报 Err。docker 与 podman 端**完全对称**(同一份决策表)。

#### 文档一致性修正(全部已闭环 2026-08-07)

| 文档 | 当前 | 修正 | 状态 |
|---|---|---|---|
| **R08-02 §F1 ComposeService interface** | 把 `Up / Run / Pull / Push` 都标 `ErrComposeUnsupported` | **已修正**:F1 方法可用性矩阵 + docstring 标注 — Up/Run/Exec/Top/Port/Stats/Events/Down/ListProjects/InspectProject/Start/Stop/Restart 全部 wrap-able(Build / Pull / Push 走 R08-05;此处只标永久 Err 是 Config) | ✅ **已合** |
| R08-04 lifecycle.md | 已规划 up/down/start/stop/restart | ✓ 与 Q3 一致,不动 | — |
| R08-05 build-transfer.md | Build ErrUnsupported;Pull/Push wrap-able | ✓ 与 Q3 一致,不动 | — |
| R08-08 exec-run.md | Run + Exec wrap-able | ✓ 与 Q3 一致,不动 | — |

**闭环检查**(自检):R08-15 §Q3 ↔ R08-02 §F1 ↔ R08-04/05/08 之间对 "Up/Run/Pull/Push wrap-able,Build/Config ErrUnsupported" 的表述已统一。**任何上层 adapter 实现前不再有文档/代码冲突风险**。

> ⚠ 此项**已闭环**,实施阶段 A 起点为"R08-02 §F1 已就位,可直接进入 C 阶段两侧 adapter 并行实现"。

#### 受影响代码路径

| Command | docker driver | podman driver |
|---|---|---|
| Up | `Containers.List(filter=project=X, state=exited)` → for each `ContainerStart` | 同(Labels 已包含) |
| Run | `ContainerCreate(--label oneoff=true)` + `ContainerStart` + `ContainerWait` + `ContainerRemove` | 同 |
| Pull | loop `ImagePull(name)` | loop ImagePull |
| Push | loop `ImagePush(ref)` | loop ImagePush |
| Build | `ErrComposeUnsupported`(hard return) | `ErrComposeUnsupported`(原则一致) |
| Config | `ErrComposeUnsupported`(hard return) | `ErrComposeUnsupported`(原则一致) |

### Q4:CoLocated group 触发条件(三模式 OR vs 严格 AND)

#### 问题场景

Q4 决定的关键判定:**docker / podman inspect 返回的 `NetworkMode / PidMode / IpcMode` 任一或全部共享时**,是否建 CoLocated group?

| 场景 | 实际 compose 模式 | 含义 |
|---|---|---|
| **S1 sidecar 模式(最常见)** | `network_mode: service:db` only(Pid / Ipc 不共享) | net namespace 共用 — web 与 db 在同一 net view |
| **S2 full pod-equivalent** | 网络 + PID + IPC 三者都共享 | k8s Pod 标准形态 |
| **S3 PID-only 共享**(少见) | `pid: service:db`,net 各自独立 | 进程命名空间共用,但网络隔离 |
| **S4 IPC-only 共享**(少见) | `ipc: service:db`,net/pid 各自独立 | System V IPC / POSIX shared memory 共用 |

#### 需求契约(可验证)

| ID | 需求 | 验证 |
|---|---|---|
| **N1 sidecar 可见** | "网络 namespace 共享"必须能触发 group(用户场景最常见) | 测试:给定 web + db `network_mode: service:db`,group "db" 出现 |
| **N2 不炸常见 case** | 不应因为引入 mode 限制而让用户看不到 group | 测试:30-service 普通 compose 项目,group 数 ≥ 0 |
| **N3 不污染噪声** | 单模式共享(尤其非 net)不应铺满 group 标签 | UI 默认折叠非 Net 共享的 group;视觉密度控制 |
| **N4 与 compose-spec 一致** | 触发条件不超出 compose-spec 允许的语义 | 比对 compose-spec `network_mode` / `pid` / `ipc` 章节 |

#### 决策 — **三模式 OR**(宽松 + UI 折叠)

| 模式 | 决策 | 依据 |
|---|---|---|
| **三模式 OR** | ✅ **采用** | N1(sidecar)直接受益 — compose-spec 允许 web + db 共享 net 而无 pid/ipc,OR 必捕获 |
| 三模式 AND(k8s-strict) | ❌ 弃 | N2 失败 — S1 这类最常见 sidecar 不会触发 group,违背用户预期 |

→ **Q4 拍板 = 三模式 OR**

#### UI 折叠策略(配套 N3 噪声控制)

| 共享模式 | UI 默认显示 | 备注 |
|---|---|---|
| `network_mode: service:X`(仅 net) | ✅ 显示 group 标签,但**可折叠**(默认 expanded) | 主用 sidecar 模式 |
| `net + pid`(2 / 3) | ✅ 显示,**默认 expanded** | 接近 pod 形态 |
| `net + pid + ipc`(3 / 3,k8s strict Pod) | ✅ 显示,**默认 expanded**,徽章样式不同 | 完整 pod 等价 |
| 仅 `pid` / 仅 `ipc`(无 net) | ⚠️ 显示,**默认 collapsed** | N3 降噪 — 主线是 net |

→ UI 通过"默认展开/折叠" + 徽章样式区分 group 类型(详细 UI 定型推迟到 R08 设计 overhaul 阶段)

#### 与上游 R08-14 对齐

R08-14 §F2 已规定三模式 OR 为推荐,本 Q4 决策与该推荐一致并细化 UI 折叠行为。
**R08-14 不需修改** — 本 Q4 仅在 R08-15 内加 UI 行为备忘。

### Q5:podman driver client 实现 — 构建期选择(static / dynamic)

#### 前置澄清

podman driver 是**单 transport**(REST to daemon)。`HTTP client` 与 `CGO bindings client` 是**单 transport 的两种 client 实现**,两者最终都打同一 daemon 的 `/libpod/...`。**dtui runtime 视角下没有 transport 多样性**。

唯一选择是**构建期**用哪种 client 实现。

#### 问题场景

| 场景 | 谁触发 | 痛点 |
|---|---|---|
| **S1 release 用户 — 静态二进制** | 普通用户跑生产 / 个人开发机 | 希望 dtui 是单文件 release,无 CGO 工具链依赖(README:84-85 已声明 `CGO_ENABLED=0`) |
| **S2 power user — 想要 in-process 性能** | 大项目批量轮询玩家(CGO 优势:进程内调,省一次 HTTP 跳)| 希望 release 之外有可选 in-process 优化 |
| **S3 构建链复杂度** | 发行者 / packager | `CGO_ENABLED=1` 需要 glibc / musl 协调 + 工具链就绪,易出 supply-chain 问题 |
| **S4 多平台分发** | release engineer | macOS / Windows / Linux 三平台分别打;静态二进制能压平矩阵 |

#### 需求契约(可验证)

| ID | 需求 | 验证 |
|---|---|---|
| **N1 主 release 静态** | dtui 主 release 是静态二进制,无 CGO 依赖 | `file dist/dtui`(应当 not dynamic linked + portable) |
| **N2 静态下接口不可降级** | `CGO_ENABLED=0` 编译后**仍能**调 podman daemon(走 HTTP) | 测试:静态 build 后跑 `dtui --host podman.sock`,功能不退化 |
| **N3 CGO 用户可选** | CGO 用户能编出一个 in-process 优化版本 | `just build-cgo` 产物存在 `dist/dtui-cgo` |
| **N4 运行时无歧义** | runtime 视角下不出现"两种 client 切换"导致的连接/状态差异 | integration test:无论静态 / CGO build,功能等价 |

#### 决策 — **静态构建默认 HTTP,可选 CGO build**

| 构建模式 | 入口 | binary 内含 | daemon 通信 |
|---|---|---|---|
| **静态构建** `CGO_ENABLED=0`(默认) | `just build` | `HTTP client`(`internal/driver/podman/rest.go`) | HTTP → daemon REST |
| **动态构建** `CGO_ENABLED=1`(可选) | `just build-cgo` | `HTTP client` + `bindings client`(`internal/driver/podman/driver_cgo.go`) | 默认 HTTP,降级 bindings 可选 |

→ **Q5 拍板 = 静态构建默认 HTTP**(与 dtui release 策略一致,见 README:84-85)

#### 用户扩展路径

| 场景 | 命令 | 产物 |
|---|---|---|
| 普通用户 / 默认 release | `just build`(or `CGO_ENABLED=0 go build`) | `dist/dtui`(静态) |
| CGO 优化 build | `just build-cgo`(or `CGO_ENABLED=1 go build -tags=podman_cgo`) | `dist/dtui-cgo`(动态,含 bindings) |

**无运行时切换** — 两个 build 是两个不同 binary,用户按需选择,N4 满足。

#### 与 Q1 / Q6 对齐

- Q1 阶段 A1 / B1 / C 都不依赖 client 实现选择 — 两套 build 路径对上层透明
- Q6 docker PodsService 自实现 wrap — 与 podman client 选择**互不相关**
- dtui 的 Engine interface 不暴露 HTTP / bindings 差异(per Q1 N4)

### Q6:docker adapter 的 `Pods()` 应实现为 wrap 还是直接报"不支持"

#### 问题场景

| 场景 | 当前 dtui 行为(决定前) | 用户痛点 |
|---|---|---|
| **S1 docker 引擎,网/pid/ipc 共享的服务组** | 报 ErrUnsupported,Pod 面板空白 | 用户看不到"哪些服务是 namespace 共享",失去最重要的 pod-like 关系可视化 |
| **S2 docker + podman 切换连接** | docker 端 Pod 永远空,podman 端有数据 | 切换时用户无法对比;误以为 docker "不支持",转去 docker compose CLI |
| **S3 大型 docker compose 项目调试** | 用户已 down,dtui 看残留容器 | Pod 面板报错 → 用户不能检视 daemon 上 label 表示的项目结构 |

#### 需求契约(可验证)

| ID | 需求 | 验证 |
|---|---|---|
| **N1 接口永不报错** | `runtime.Engine.Pods()` 永远返回非 nil `PodService` 实现,不返回 `ErrUnsupported` | 单元测试:任意 Engine 实例调用 Pods() 不报错 |
| **N2 docker 与 podman 对称** | 两 adapter 都能提供 Pod 数据;UI 不区分能填 | 测试:docker 30 容器项目能展示 ≥ 0 个 pod group |
| **N3 复用既有的 capability 矩阵** | `CapabilityComposePodScope` 两侧都 Available(无 Err 分支) | §6 / §7 不存在"docker Pods = Unsupported"路径 |
| **N4 与 §0 硬规则一致** | "daemon 端没有 → driver 层 wrap,不向底层逃"是 §0 第 3 硬规则的 Pod 域延伸 | grep:`R08-15 §0` 第 3 条 + `R08-15 §F4 / §Q6` 语义不冲突 |

#### 决策 — **driver 层自实现 PodsService**(封装而非报错)

| 选项 | 决策 | 依据 |
|---|---|---|
| 返回 `nil, ErrComposeUnsupported` | ❌ 弃 | **违反 N1 + N3 + §0 硬规则 #3** — 接口不报错的原则,以及"wrap 而非报错"的 §0 原则 |
| 不实现 `Pods()`(panics) | ❌ 弃 | 编译期错;违反 N1 |
| **driver 层自实现 wrap`(推荐)`** | ✅ **采用** | N1-N4 全满足 + 与 R08-14 CoLocated group 同源算法 + Pod DTO 同 interface |

#### 实现要点

| 维度 | 实现 |
|---|---|
| **数据源** | docker `container.Inspect.HostConfig.NetworkMode / PidMode / IpcMode` |
| **算法** | 与 R08-14 CoLocated group 推导**算法同源**(共享引用即分组);Q4 拍板三模式 OR |
| **DTO** | `runtime.PodService.Pod` vs `runtime.RuntimeCoLocatedGroup` — interface 同,DTO 字段有差(见 §命名约定 8.4) |
| **缓存** | per Q2:C' 修订决策 — 5s TTL + docker events 失效 |

#### 与 R08-14 的关系

R08-14 CoLocated group 推导已经是**容器级**(一个 project 内多服务)。docker PodsService wrap 是**项目级**(每个 project 对应一个合成 pod,但 fact 都是同源数据)。DTO 不同但**算法同源**——可共享 `coLocatedResolver` 内部逻辑,DTO 转换层另写。

#### 文档一致性(已在 Q3 处置)

R08-02 §F1 原标 `Up / Run / Pull / Push` 为 ErrUnsupported — 已被 Q3 修正;Capability 矩阵中 `CapabilityComposePodScope` 两侧都标记 Available(已在 R08-15 §6 更新)。

#### 实施入口(per Q1 阶段 C)

- `runtime/docker/pod_service.go` 新文件 — `dockerPodService` struct,实现 `runtime.PodService` 接口
- `runtime/docker/co_located.go` 内部 share `coLocatedResolver`(per 上面"算法同源"提示)
- Unit 测试:mock ContainerList + Inspect → 预期 Pod DTO 列表

## 验收标准

1. engine 接口增加 `Compose() ComposeService`,Docker 与 Podman adapter **各实现一份**(包括 List/Start/Stop/.../Down)
2. engine 接口增加 `Pods() PodService`,**两端各自实现**(per Q6)
3. **podman adapter** `Pods()` 调 swagger `/libpod/pods/*` 端点,直接包装
4. **docker adapter** `Pods()` **自实现 wrap**:`ContainerList` + 每容器 `ContainerInspect`(NetworkMode / PidMode / IpcMode),产出与 R08-14 CoLocated group 算法同源但 DTO 不同的 Pod 数据
5. docker adapter 在 `Up / Run / Build / Config` 上 hard return `ErrComposeUnsupported`
6. podman adapter 在 `Up / Run / Build / Config` 上同样返回 `ErrComposeUnsupported`(原则一致)
7. docker adapter **不**写 yaml;Podman adapter **不**写 yaml;都不 shelling out
8. docker `HostConfig.NetworkMode` / `PidMode` / `IpcMode` 经 inspect 后写入 `ContainerSummary` 新字段
9. podman `inspect.Pod` 字段直接读,无需额外 RTT
10. 双 adapter 各自的 `co_located.go` 文件存在,**互不调用**(per R08-01 F1 原则)
11. `Capabilities()` 报告矩阵正确反映 §6(**双端 PodScope 都是 Available**,实现路径不同)
12. CoLocated group 推导算法按 §4.2,2+ 引用阈值,OR 触发(Q4)
13. orphan 简化算法按 §5.3,纯 label 推导,无 yaml
14. 单元测试覆盖 docker / podman 各自的 list /inspect 数据 → CoLocatedGroup + capabilities 矩阵
15. 单元测试覆盖 docker Pods 自实现 wrap(给定 mock container list + NetworkMode,产出预期 Pod 列表)
16. podman driver client 实现策略:`CGO_ENABLED=0` 默认 HTTP,`CGO_ENABLED=1` 可启用 bindings(per Q5)

## 非目标

- 不引入 yaml 解析器(不读 compose.yaml / compose.yml)
- 不引入 `docker compose` CLI / `podman-compose` CLI / `podman compose` CLI shelling out
- 不实现 Podman 原生 Pod 面板的完整功能(留 `feature-todo-list.md §6.4` 独立 R 需求)
- 不支持 Docker Swarm stack mode(走 docker events 不读 swarm 编排层)
- 不暴露 docker compose `convert` / plugin 管理类次要子命令
- 不实现引擎端不存在的服务级原子动作

## 迁移记录

- 旧文档:`compose-todo.md 补 E.4` "Engine 接口没有 ComposeService"、`.omo/compose-todo.md` 系列
- 旧代码:无(本需求为新建)
- 复用:R08-02(ComposeService API)、R08-14(CoLocated group)
- 保留信息:`docker/inspect.go` 已含 NetworkMode 字段(本研究确认)
- 待确认:Q1-Q4(本需求 待确认 节)

## 参考资料(本文档依据)

- `docs/api/podman-swagger.yaml`(Podman REST 端点完整列表)
- `internal/data/runtime/engine.go:7-19`(Engine 接口现状)
- `internal/data/runtime/capabilities.go`(Capability 现状,无 Compose 类)
- `internal/data/runtime/docker/client.go`(Docker Client 结构)
- `internal/data/runtime/docker/inspect.go:55`(HostConfig.NetworkMode 字段已存在)
- `internal/data/runtime/docker/service_container.go`(容器服务实现)
- `internal/data/runtime/podman/engine.go`(PodmanEngine 结构)
- `internal/data/runtime/podman/mappers.go`(Podman 容器映射,无 Pod 字段处理)
- R08-01 / R08-02 / R08-14(R08 大需求内的相关章节)
- `compose-spec/json-schema`(compose-spec 顶层无 pod key 验证)
- `podman_compose.py:2543-2570`(podman-compose 默认 1 项目 1 pod 验证)
- `docs.podman.io/markdown/podman-compose.1.html`(`podman compose` 薄包装验证)
