# Podman REST 适配与 docker/service 统一方案

> 建立日期: 2026-07-23
> 最近修订: 2026-07-24（用户决策：移除 docker/service 统一层，简化架构）
> 状态: `discussing` → 待 `approved` 后转入执行
> 关联任务: `TASK-022`
> 前置文档: [`unified-runtime-driver.md`](./unified-runtime-driver.md)、[`podman-capabilities-analysis.md`](./podman-capabilities-analysis.md)、[`future-requirements-discussion.md`](./future-requirements-discussion.md)
> 后续任务台账: [`future-requirements-task-list.md`](./future-requirements-task-list.md)

## 背景

`internal/data/runtime/podman` 已具备 REST 适配骨架（`rest.go`、`containers_rest.go` 等），但 `aliases_cgo.go`、`aliases_container_cgo.go`、`aliases_network.go` 三个文件**缺失 `//go:build cgo` 标签**，导致非 CGO 构建也会间接引入 `github.com/proglottis/gpgme`。

此外：

1. 现有包级函数（如 `ListContainersREST(ctx, *Client, options)`）与"方法式"调用语义重复，且签名混用 `runtimeapi.*` 与 `dto.*`。
2. `internal/data/docker` 中仍残留 24 个 `podman_*.go` 生产文件 + 8 个测试文件，Podman 适配代码分散两处。
3. 当前 docker/service 层对 Docker 与 Podman 分别走两条 service 路径，未真正统一。
4. 全仓存在多处匿名内联 `struct{}` 定义（已枚举），违反类型可读性原则。

本方案在 `unified-runtime-driver.md`（TASK-021）的成果上：

- 把 Podman 适配层收口到 `internal/data/runtime/podman`，并改造为"驱动 + REST 双形态但单一 Client 结构、方法式 API、参数与返回全部 `dto.*`"。
- 在 `internal/data/docker/service` 层实现 **Docker / Podman 统一**入口，让上层只持有 `runtime.Engine`，业务侧不再分支判断 `RuntimeType`。
- 隔离 `gpgme` 仅出现在 CGO 构建路径，确保非 CGO 构建链路无 `go.podman.io/...` 间接依赖。
- 消除所有匿名结构体，全类型具名化。

---

## 目标与约束

### 目标（2026-07-24 用户最终修订）

| # | 目标 | 修订内容 |
|---|---|---|
| G1 | `internal/driver` 是 Podman 整合层，**同时承担 Go 依赖（`go.podman.io/*` + `gpgme`）与 REST transport 整合**。`runtime/podman` 仅消费 `internal/driver`，不再单独引 CGO 依赖；`internal/data/docker` 中 Podman 专属代码本轮不删除。 | 同前 |
| G2 | `runtime/podman.Client` 方法签名**全部**为 `dto.Action` / `dto.ActionOptions` / `dto.ActionResult`，**不再保留旧 `string action` 兼容签名**。这是破坏性变更；调用方一次性迁移。 | 用户决策 2026-07-24：不双签名，统一走 dto |
| G3 | 同一方法在 CGO 模式下走驱动 + 内部映射，在非 CGO 模式下走 REST。两套实现由构建标签隔开，**对外接口形态一致**。 | 维持 |
| G4 | **不增加 `docker/service` 统一入口层**。mapper 直接拆到两处：Docker 专属 mapper 落在 `runtime/docker/mapper/`，Podman 专属 mapper 落在 `runtime/podman/mapper/`（已存在 `mappers.go`，向上提一层）。`runtime.Engine` 仍是唯一抽象层。 | 用户决策 2026-07-24：架构简化，省去 docker/service 层 |
| G5 | Podman 原生动作以 `dto.Action` 字符串枚举表达；`runtimeapi.Action` ↔ `dto.Action` 语义映射放在 `runtime/docker/mapper/` 与 `runtime/podman/mapper/` 各自完成。**不通过 `docker/service` 层**。 | 用户决策 2026-07-24：架构简化 |
| G6 | **不允许任何匿名结构体**。本轮不强求全零：仅清理被引入公共 API 路径的匿名 struct；剩余的内部 helper（`ProgressWriter/Reader`、`StatsJSON` 内部子结构等）作为后续清理项保留 TODO。 | 同前 |
| G7 | 三层类型独立互不别名/转换：A 层 Docker SDK（仅 `docker/` 引用）；B 层 Podman 驱动（仅 `runtime/podman/*_cgo.go` 引用）；C 层 Podman REST DTO（`runtime/podman/dto/`，**绝不**引用 `go.podman.io/...`）。 | 维持 |
| G8 | mapper 分两个位置落：Docker 专属 mapper（Docker SDK type ↔ `runtimeapi.*`）落 `internal/data/runtime/docker/mapper/`；Podman 专属 mapper（`dto.*` ↔ `runtimeapi.*`）落 `internal/data/runtime/podman/mapper/`（提升自现有 `mappers.go`）。**不再走 `docker/service/` 与 `internal/driver/podman/mapper/` 双层结构**。 | 用户决策 2026-07-24：架构简化 |
| G9 | **取消** `docker/service` 统一入口（与 G4 / G5 / G8 同步取消）。上层仍只持有 `runtime.Engine`，两个 service 实现分别走 `runtime/docker/*` 与 `runtime/podman/*` 入口。 | 取消 |
| G10 | `CGO_ENABLED=0` 时全仓零 `gpgme`；所有测试在双构建下均通过。**只有被 CGO 路径实际 `import` 的文件**需要加 `//go:build cgo` 标签；未被引用的 `aliases*.go` 不强求加标签。判定标准：`go mod why <package>` + `go list -deps`。 | 维持 |

### 与 `unified-runtime-driver.md` 的关系

TASK-021 已经完成 Docker / Podman runtime 适配层（独立 Engine、`runtime.Engine` 契约、能力 / 错误 / 身份）。本方案：

- **不重做** TASK-021 已确立的 Engine 形态与 capability / error 体系。
- **承接** TASK-021 留下的 `runtime/podman` 实现细节（`Client` 结构、REST 路径、CGO 桩），将其收紧为方法式 + dto-only 签名 + 驱动/REST 双形态。
- **开启** docker/service 层对 Docker 与 Podman 真正统一的新工作流（§6）。

---

### 与 `unified-runtime-driver.md` 的关系

TASK-021 已经完成 Docker / Podman runtime 适配层（独立 Engine、`runtime.Engine` 契约、能力 / 错误 / 身份）。本方案（2026-07-24 用户修订版）：

- **不重做** TASK-021 已确立的 Engine 形态与 capability / error 体系。
- **承接** TASK-021 留下的 `runtime/podman` 实现细节（`Client` 结构、REST 路径、CGO 桩），将其收紧为方法式 + `dto.*` 签名 + 驱动/REST 双形态。
- **简化为两适配器直接并列**：上层只持有 `runtime.Engine`，业务侧不感知 Docker / Podman 差异。**不引入 `docker/service` 层**；两个 mapper 分别落在 `internal/data/runtime/docker/mapper/` 与 `internal/data/runtime/podman/mapper/`（提升自现有 `mappers.go`）。
- 一次性破坏性变更：所有 `action string` 调用点切到 `action dto.Action`。

### G4 架构最终方案（2026-07-24 用户修订）

候选方案简化为二选一：

| 方案 | 数据流 | 是否采纳 | 说明 |
|---|---|---|---|
| **A. 拆 mapper 到两侧（采纳）** | UI → runtime.Engine → {DockerContainerService (runtime/docker)、PodmanContainerService (runtime/podman)} → driver → transport | ✅ | 维持两层（runtime.Engine + 双 adapter），每个 adapter 内部自带 mapper；无 docker/service |
| **B. 引入 docker/service 统一层** | UI → runtime.Engine → docker/service (UnifiedContainerService) → {DockerAdapter / PodmanAdapter} → driver → transport | ❌ | 引入额外抽象；目前必要性不足；后续如发现大量跨 adapter 共享逻辑再启动 |

**优势**：
1. 现有 `runtime.Engine` 抽象足够上层消费，不需要再叠一层
2. docker mapper 与 podman mapper 隔离在各自 adapter 包内，互不干扰
3. 取消 docker/service 后，包路径变浅，新增概念更少

**潜在风险**：跨 adapter 共享代码（如通用错误处理、性能优化）需要由 `runtimeapi.*` 的 helper 函数提供，而不是再抽 service 层。后续评估此风险是否真实存在。

### `internal/driver` 整合层（G1 角色）

**整合层定位**：`internal/driver` 是 Podman Go 依赖（`go.podman.io/*` + `gpgme`）与 REST transport 的整合包，对外暴露 `Client`（仅 REST + CGO 字段），给 `runtime/podman` 消费。

物理分层：
```
runtime/podman/                      ← UI 入口，runtime.Engine 实现
    └─ PodmanEngine { Client }
            └─ Client = internal/driver/podman.Client
                    ├─ client.REST  (RESTClient)
                    ├─ client.Driver (cgoDriver，可选)
                    └─ dto.*  共享类型
```

**判定 build tag 的最小集（G10 修订）**：
```
go list -deps ./internal/driver/podman | grep gpgme          # 在 CGO 时命中；在非 CGO 时为空
go mod why github.com/proglottis/gpgme                        # 锁定传染源
```

只有被 deps 实际传递引用的源文件需要 `//go:build cgo`：
- `internal/driver/podman/driver_cgo.go` 必须 `//go:build cgo`
- `client_default_driver_cgo.go` 必须 `//go:build cgo`
- `internal/driver/podman/aliases*.go` 等未被实际 import 的文件**不加 tag**

### G8 mapper 双位置（用户最终决策）

```
internal/data/runtime/docker/mapper/       ← Docker SDK type ↔ runtimeapi.*
    ├── container.go     # MapContainerSummaries（已存在）
    ├── image.go         # MapImageSummaries（已存在）
    ├── volume.go        # MapVolumes
    ├── network.go       # MapNetworks
    └── action.go        # dto.Action ↔ runtimeapi.Action 映射（决策项 D6/D8 简化版）

internal/data/runtime/podman/mapper/        ← Podman dto.* ↔ runtimeapi.*
    ├── container.go     # MapContainerSummaries（从当前 mappers.go 提取）
    ├── image.go         # MapImageSummaries（含 ImageInspect 映射）
    ├── volume.go        # MapVolumes（含 VolumeDetail）
    ├── network.go       # MapNetworks（含 NetworkDetail）
    └── action.go        # dto.Action 包装 / 字符串化

# 不再有 internal/driver/podman/mapper/、internal/data/docker/service/mapper/ 双层结构
```

---

## 三层类型边界

```
┌──────────────────────────────────────────────────────────────┐
│ A 层: Docker SDK                                              │
│   import: github.com/docker/docker/api/types/*               │
│   仅 internal/data/runtime/docker/ 引用                      │
└──────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────┐
│ C 层: Podman REST DTO                                         │
│   路径: internal/data/runtime/podman/dto/*                   │
│   import: 仅标准库 + time                                     │
│   引用方: runtime/podman/*.go (无 CGO)、runtime/docker/mapper/*（仅允许 dto.*）│
└──────────────────────────────────────────────────────────────┘
                            ▲
                            │
┌──────────────────────────────────────────────────────────────┐
│ B 层: Podman 驱动（CGO only，由 internal/driver/podman 整合）│
│   路径: internal/driver/podman/driver_cgo.go 等              │
│   import: go.podman.io/podman/v6/pkg/domain/entities/types   │
│           go.podman.io/common/libnetwork/types                │
│           go.podman.io/podman/v6/libpod/define                │
│           go.podman.io/podman/v6/pkg/bindings.*               │
│           github.com/proglottis/gpgme                         │
│   仅 //go:build cgo 文件引用                                  │
│   运行时: B → C 内部映射（mapper_driver_cgo.go 需新建）         │
└──────────────────────────────────────────────────────────────┘

**禁止**：
- B 与 C 之间互相 `type A = B` 别名。
- C 包导入 B 包任何路径（`go list -deps ./dto | grep podman` 必须为空）。
- B 层类型出现在非 CGO 文件。
- A 层类型出现在 `runtime/podman/`、`internal/driver/podman/`。
- mapper（`internal/data/runtime/{docker,podman}/mapper/`）不得依赖 A 层（Docker SDK）或 B 层（`go.podman.io/*`）— mapper 只在 `dto.*` ↔ `runtimeapi.*` 之间互转。

### 关于"驱动参数与 REST 参数保持一致时使用别名"

如果某个 B 层类型的 JSON tag 与 C 层 DTO 一一对应（这是 Podman 项目自身的现实：REST 响应字段与 bindings 返回的 Go 结构字段完全一致），那么 C 层**不需要**重复复制全部字段，而是用 `type` 别名：

```go
// dto/container.go （无构建标签）
type ContainerItem = types.ListContainer          // 仅 CGO 有效
type ContainerItem = dtoStructContainerItem       // 仅 !CGO
```

但是别名必须按构建标签分支（B 层不可在非 CGO 文件中可见）。因此：

- 当 B 层类型在 `//go:build cgo` 下完整可用时，C 层在 CGO 文件里直接 `type ContainerItem = types.ListContainer`。
- 在 `//go:build !cgo` 文件里，必须用一份独立具名 DTO（用 `go doc` 复制 JSON tag）。

别名只用于**字段与 JSON tag 完全等价**的场景；JSON tag 不一致或存在需要本地裁剪的字段（如部分隐藏方法、私有字段）时，必须用独立具名结构体。

**判定流程**：

```
B 层类型 → go doc 字段列表 → 与 C 层 DTO 字段 + JSON tag 比对
   ├─ 完全等价 → type 别名（CGO）+ 独立结构（非 CGO）双文件
   └─ 有差异   → C 层独立具名结构（两个文件相同定义）
```

---

## `runtime/podman/dto` 包

`dto/` 包**仅**引用标准库与 `time`。**绝不**引用 `go.podman.io/...`、相关 buildah / image / common 包、`github.com/proglottis/gpgme`。

验证：

```bash
go list -deps ./internal/data/runtime/podman/dto | grep -E 'podman|buildah|common|gpgme'   # 必须为空
```

### 类型清单

| 文件 | 类型 |
|---|---|
| `dto/container.go` | `ContainerItem`（别名双形态）、`ContainerPort`、`ContainerInspect`、`ContainerInspectState`、`ContainerInspectConfig`、`ContainerInspectHostConfig`、`ContainerInspectRestartPolicy`、`ContainerInspectNetSettings`、`ContainerInspectNetworkEntry`、`ContainerInspectPortBinding`、`ContainerInspectMount`、`ContainerTopResponse`、`ContainerStatsResponse`（含嵌套具名子结构）、`ContainerProcesses`、`ContainerStats`、`ContainerLogOptions`、`ContainerMount`、`ContainerNetwork`、`ContainerPortBinding`、`ContainerState`、`ContainerResources`、`ContainerConfig`、`ContainerListOptions` |
| `dto/image.go` | `ImageItem`（别名双形态）、`ImageInspect`、`ImageHistoryLayer`、`ImageHistorySource`、`ImageManifestEntry`、`ImageRuntimeConfig`、`ManifestPlatform`、`ImagePruneReport`、`ImagePullStreamMessage`、`ImageTransferProgress`、`ImageTransferResult`、`ImageListOptions` |
| `dto/network.go` | `Network`（别名双形态）、`NetworkSubnet`、`NetworkInterface`、`NetworkAddress`、`NetworkInspect`（别名双形态）、`NetworkContainerInfo`、`NetworkCreateInput`、`NetworkCreateOptions`、`NetworkPruneReport`、`NetworkListOptions` |
| `dto/volume.go` | `VolumeItem`（别名双形态）、`VolumeInspect`、`VolumeCreateInput`、`VolumeCreateOptions`、`VolumePruneReport`、`VolumeListOptions` |
| `dto/exec.go` | `ExecCreateRequest`、`ExecCreateResponse`、`ExecStartRequest`、`ExecStartOptions`、`ExecOptions`、`TerminalSize`、`ExecSession`（接口） |
| `dto/event.go` | `Event`、`EventItem`、`EventOptions`、`EventFilter` |
| `dto/version.go` | `Version`、`ComponentVersion`、`ComponentDetails` |
| `dto/image_transfer.go` | `ImageTransferRequest`、`ImageTransferResult`、`ImageTransferEvent`、`ImageTransferOperation`、`ImageTransferProgress` |
| `dto/error.go` | `ErrorPayload`（解 REST 错误 JSON） |
| `dto/prune.go` | `PruneOptions`、`PruneResult`、`PruneReport`、`ResourceResult` |
| `dto/action.go` | `Action`（字符串）、`ActionOptions`、`ActionResult` |

### `dto.Action` 定义

```go
// dto/action.go
type Action string

const (
    ActionStart    Action = "start"
    ActionStop     Action = "stop"
    ActionRestart  Action = "restart"
    ActionKill     Action = "kill"
    ActionPause    Action = "pause"
    ActionUnpause  Action = "unpause"
    ActionRemove   Action = "remove"
    ActionRename   Action = "rename"
    ActionPull     Action = "pull"
    ActionPrune    Action = "prune"
    ActionTag      Action = "tag"
    ActionPush     Action = "push"
    ActionSave     Action = "save"
    ActionLoad     Action = "load"
    ActionExec     Action = "exec"
)

type ActionOptions struct {
    Force     bool
    Timeout   time.Duration
    Signal    string
    Name      string
    Reference string
}

type ActionResult struct {
    SpaceReclaimed int64
    References     []string
}
```

`docker/service` 层负责 `dto.Action` ↔ `runtimeapi.Action` 的双向映射，并提供动作描述（i18n 文案键）。

### 匿名结构体替换表

| 文件 | 现匿名 struct | 替换为 |
|---|---|---|
| `runtime/podman/rest.go:110-118` | `version { ApiVersion, Components[]{ Details{} } }` | `dto.Version`、`dto.ComponentVersion`、`dto.ComponentDetails` |
| `runtime/podman/rest.go:143-148` | `extractLibpodVersion([]struct{...}{...}{...})` | `[]dto.ComponentVersion` |
| `runtime/podman/stats.go` 全字段匿名嵌套 | `StatsJSON{CPUStats{CPUUsage{}}, MemoryStats{...}}` | `dto.ContainerStatsResponse{StatsCPU{StatsCPUUsage{}}, StatsMemory{}}`（全部具名） |
| `runtime/podman/containers_read.go:32` | `var raw struct{ Titles, Processes }` | `dto.ContainerTopResponse` |
| `runtime/podman/images_transfer.go:86` | `var report struct{ ID, Error, ... }` | `dto.LoadReportItem` |
| `internal/data/docker/podman_container_read_rest.go:29` | Top 结构 | `dto.ContainerTopResponse` |
| `internal/data/docker/podman_exec_rest.go:22/32/42` | request/created/startBody | `dto.ExecCreateRequest`、`dto.ExecCreateResponse`、`dto.ExecStartRequest` |
| `internal/data/docker/podman_image_actions_rest.go:29` | `[]struct{ID, Size, Err}` | `[]dto.ImagePruneReport` |
| `internal/data/docker/podman_image_transfer_rest.go:105` | `var report struct{...}` | `dto.PushProgressItem` |
| `internal/data/docker/podman_networks_rest.go:62` | `input := struct{...}` | `dto.NetworkCreateInput` |
| `internal/data/docker/podman_volumes_rest.go:65` | `input := struct{...}` | `dto.VolumeCreateInput` |

终态命令：

```bash
grep -rn 'struct {' internal/ | grep -v _test.go | grep -vE '^[^:]+:\s*//'   # 空
```

---

## `runtime/podman.Client` 形态（方法式）

### Client 结构

```go
// runtime/podman/client.go （无构建标签）
type Client struct {
    REST *RESTClient          // 始终存在
    Host string
    TLS  bool
    // Driver 由 driver_cgo.go / driver_nocgo.go 提供方法访问
}

// DriverBackend 接口在 driver_iface.go（无标签）声明
type DriverBackend interface {
    ListContainers(ctx context.Context, options dto.ContainerListOptions) ([]dto.ContainerItem, error)
    InspectContainer(ctx context.Context, id string) (*dto.ContainerInspect, error)
    ContainerTop(ctx context.Context, id string) (dto.ContainerTopResponse, error)
    ContainerStats(ctx context.Context, id string) (dto.ContainerStatsResponse, error)
    ContainerLogs(ctx context.Context, id string, options dto.ContainerLogOptions) (io.ReadCloser, error)
    ExecuteContainerAction(ctx context.Context, id string, action dto.Action, options dto.ActionOptions) error

    ListImages(ctx context.Context, options dto.ImageListOptions) ([]dto.ImageItem, error)
    InspectImage(ctx context.Context, summary dto.ImageItem) (*dto.ImageInspect, error)
    ExecuteImageAction(ctx context.Context, id string, action dto.Action, options dto.ActionOptions, result *dto.ActionResult) error
    TagImage(ctx context.Context, source, destination string) error
    PushImage(ctx context.Context, request dto.ImageTransferRequest, output chan<- dto.ImageTransferEvent) error
    SaveImage(ctx context.Context, request dto.ImageTransferRequest, output chan<- dto.ImageTransferEvent) error
    LoadImage(ctx context.Context, request dto.ImageTransferRequest, output chan<- dto.ImageTransferEvent) ([]string, error)

    ListNetworks(ctx context.Context, options dto.NetworkListOptions) ([]dto.Network, error)
    InspectNetwork(ctx context.Context, id string) (*dto.NetworkInspect, error)
    CreateNetwork(ctx context.Context, options dto.NetworkCreateOptions) (*dto.Network, error)
    RemoveNetwork(ctx context.Context, id string) error
    PruneNetworks(ctx context.Context, options dto.PruneOptions) (dto.PruneResult, error)

    ListVolumes(ctx context.Context, options dto.VolumeListOptions) ([]dto.VolumeItem, error)
    InspectVolume(ctx context.Context, name string) (*dto.VolumeInspect, error)
    CreateVolume(ctx context.Context, options dto.VolumeCreateOptions) (*dto.VolumeItem, error)
    RemoveVolume(ctx context.Context, name string, force bool) error
    PruneVolumes(ctx context.Context, options dto.PruneOptions) (dto.PruneResult, error)

    Close() error
}

// Driver() 由 driver_cgo.go / driver_nocgo.go 分别返回 *cgoDriver 或 nil
func (c *Client) Driver() DriverBackend
```

### 公开方法签名（全部 dto.*）

```go
// 容器
func (c *Client) ListContainers(ctx context.Context, options dto.ContainerListOptions) ([]dto.ContainerItem, error)
func (c *Client) InspectContainer(ctx context.Context, id string) (*dto.ContainerInspect, error)
func (c *Client) ContainerTop(ctx context.Context, id string) (dto.ContainerTopResponse, error)
func (c *Client) ContainerStats(ctx context.Context, id string) (dto.ContainerStatsResponse, error)
func (c *Client) ContainerLogs(ctx context.Context, id string, options dto.ContainerLogOptions) (io.ReadCloser, error)
func (c *Client) ExecuteContainerAction(ctx context.Context, id string, action dto.Action, options dto.ActionOptions) error

// 镜像
func (c *Client) ListImages(ctx context.Context, options dto.ImageListOptions) ([]dto.ImageItem, error)
func (c *Client) InspectImage(ctx context.Context, summary dto.ImageItem) (*dto.ImageInspect, error)
func (c *Client) ExecuteImageAction(ctx context.Context, id string, action dto.Action, options dto.ActionOptions, result *dto.ActionResult) error
func (c *Client) TagImage(ctx context.Context, source, destination string) error
func (c *Client) PushImage(ctx context.Context, request dto.ImageTransferRequest, output chan<- dto.ImageTransferEvent) error
func (c *Client) SaveImage(ctx context.Context, request dto.ImageTransferRequest, output chan<- dto.ImageTransferEvent) error
func (c *Client) LoadImage(ctx context.Context, request dto.ImageTransferRequest, output chan<- dto.ImageTransferEvent) ([]string, error)

// 网络
func (c *Client) ListNetworks(ctx context.Context, options dto.NetworkListOptions) ([]dto.Network, error)
func (c *Client) InspectNetwork(ctx context.Context, id string) (*dto.NetworkInspect, error)
func (c *Client) CreateNetwork(ctx context.Context, options dto.NetworkCreateOptions) (*dto.Network, error)
func (c *Client) RemoveNetwork(ctx context.Context, id string) error
func (c *Client) PruneNetworks(ctx context.Context, options dto.PruneOptions) (dto.PruneResult, error)

// 卷
func (c *Client) ListVolumes(ctx context.Context, options dto.VolumeListOptions) ([]dto.VolumeItem, error)
func (c *Client) InspectVolume(ctx context.Context, name string) (*dto.VolumeInspect, error)
func (c *Client) CreateVolume(ctx context.Context, options dto.VolumeCreateOptions) (*dto.VolumeItem, error)
func (c *Client) RemoveVolume(ctx context.Context, name string, force bool) error
func (c *Client) PruneVolumes(ctx context.Context, options dto.PruneOptions) (dto.PruneResult, error)

// Exec
func (c *Client) ExecCreate(ctx context.Context, containerID string, options dto.ExecOptions) (string, error)
func (c *Client) ExecStart(ctx context.Context, execID string, options dto.ExecStartOptions) (io.ReadWriteCloser, error)
func (c *Client) ExecResize(ctx context.Context, execID string, size dto.TerminalSize) error

// Events
func (c *Client) SubscribeEvents(ctx context.Context, options dto.EventOptions) (<-chan dto.Event, error)

// 元信息
func (c *Client) Version(ctx context.Context) (dto.Version, error)
func (c *Client) APIVersion(ctx context.Context) (string, error)
func (c *Client) Ping(ctx context.Context) error
func (c *Client) Close() error
```

### 单一方法 + 内部派发（避免重复签名）

```go
// runtime/podman/containers_list.go （无标签）
func (c *Client) ListContainers(ctx context.Context, options dto.ContainerListOptions) ([]dto.ContainerItem, error) {
    if d := c.Driver(); d != nil {
        return d.ListContainers(ctx, options)
    }
    return c.restListContainers(ctx, options)
}

// runtime/podman/driver_cgo.go （//go:build cgo）
type cgoDriver struct {
    conn *bindings.Connection
    host string
    tls  bool
}

func (c *Client) Driver() DriverBackend { return c.cgoDriver }

func (d *cgoDriver) ListContainers(ctx context.Context, options dto.ContainerListOptions) ([]dto.ContainerItem, error) {
    list, err := containers.List(d.conn, listOptsFrom(options))
    return mapContainersListToDTO(list), err   // 内部 B → C 映射
}

// runtime/podman/driver_nocgo.go （//go:build !cgo）
func (c *Client) Driver() DriverBackend { return nil }
```

`RESTClient` 保持单文件 `rest.go` 定义，与 `Driver` 互不耦合。

---

## `runtime/podman` 文件结构

```
internal/data/runtime/podman/
  client.go                       // Client + REST 字段 + Driver() 接口（无标签）
  rest.go                         // RESTClient 单文件定义
  driver_iface.go                 // DriverBackend 接口（无标签）
  driver_cgo.go                   // //go:build cgo  cgoDriver + bindings 包装
  driver_nocgo.go                 // //go:build !cgo 返回 nil

  # 容器
  containers_list.go              // ListContainers 方法体（无标签）
  containers_list_rest.go         // restListContainers
  containers_inspect.go
  containers_inspect_rest.go
  containers_top.go
  containers_top_rest.go
  containers_stats.go
  containers_stats_rest.go
  containers_logs.go
  containers_logs_rest.go
  containers_action.go
  containers_action_rest.go

  # 镜像
  images_list.go
  images_list_rest.go
  images_inspect.go
  images_inspect_rest.go
  images_action.go
  images_action_rest.go
  images_transfer.go              // Tag/Push/Save/Load
  images_transfer_rest.go

  # 网络
  networks_list.go
  networks_list_rest.go
  networks_inspect.go
  networks_inspect_rest.go
  networks_create.go
  networks_create_rest.go
  networks_remove.go
  networks_remove_rest.go
  networks_prune.go
  networks_prune_rest.go

  # 卷
  volumes_list.go
  volumes_list_rest.go
  volumes_inspect.go
  volumes_inspect_rest.go
  volumes_create.go
  volumes_create_rest.go
  volumes_remove.go
  volumes_remove_rest.go
  volumes_prune.go
  volumes_prune_rest.go

  # Exec
  exec.go
  exec_rest.go

  # Events
  events.go
  events_rest.go

  # 元信息
  version.go
  version_rest.go

  # 辅助
  paths.go            // URL 路径
  urls.go             // 端点常量
  helpers.go          // JSON 编解码、过滤器拼装
  filter_post.go      // 容器/镜像/卷/网络的过滤逻辑
  error_mapping.go    // 错误包装
  event_util.go       // 事件管道发送工具
  stats.go            // 统计字段公式

  # CGO 双形态桩（保留兼容；新代码不再使用）
  containers_cgo.go / containers_nocgo.go
  images_cgo.go / images_nocgo.go
  networks_cgo.go / networks_nocgo.go
  volumes_cgo.go / volumes_nocgo.go

  # 旧 REST 入口（保留兼容；新代码不再使用）
  containers_rest.go
  images_rest.go
  networks_rest.go
  volumes_rest.go
  containers_actions.go
  images_actions.go
  images_transfer.go
  containers_read.go
  services.go
  service_adapters.go
  engine.go

  # 别名（CGO + 非 CGO 双形态）
  aliases_cgo.go                 // //go:build cgo
  aliases_nocgo.go               // //go:build !cgo
  aliases_container_cgo.go       // //go:build cgo
  aliases_container_nocgo.go     // //go:build !cgo
  aliases_network_cgo.go         // //go:build cgo
  aliases_network_nocgo.go       // //go:build !cgo

  # 内部 B → C 映射（仅 CGO）
  mapper_driver_cgo.go           // //go:build cgo  Podman 驱动类型 → dto.*

internal/data/runtime/podman/dto/
  container.go
  image.go
  network.go
  volume.go
  exec.go
  event.go
  version.go
  image_transfer.go
  error.go
  prune.go
  action.go
```

---

## `docker/service` 层统一（Docker / Podman 单一入口）

目标：在 `internal/data/docker/service` 下建立**统一 service 层**，对上层（`runtime.Engine` 消费者）暴露同一接口，内部根据 `RuntimeType` 分派到 Docker SDK 或 Podman SDK；上层不再 `if RuntimeType == Podman`。

### 现有状态

`engine_factory.go` 提供 `dockerEngine`（Docker SDK）与 `podmanEngine`（Podman 双 transport），都实现 `runtime.Engine`。上层通过 `runtime.Engine` 消费，但是 `dockerEngine.Containers()` 等返回的是 `containerService{client: c}`，调用 `c.listContainersDocker(...)` 或 `c.listContainersPodman(...)`，**两条路径在 docker/ 内部并存**。

### 目标架构

```
runtime.Engine 消费者（TUI / state）
    │
    ▼
runtimeapi.ContainerService / VolumeService / NetworkService / ImageService
    │
    ▼
docker/service.UnifiedContainerService (接口实现)
    │
    ├── RuntimeDocker  → DockerAdapter (走 Docker SDK)
    │
    └── RuntimePodman  → PodmanAdapter (走 runtime/podman.Client)
                            │
                            └── dto.* → runtimeapi.* mapper（也位于 docker/service）
```

### 目录

```
internal/data/docker/service/
  unified_container.go            // UnifiedContainerService 公开类型
  unified_image.go
  unified_volume.go
  unified_network.go
  unified_exec.go
  unified_events.go
  unified_image_transfer.go
  unified_resource_action.go

  adapter_docker.go               // DockerAdapter（Docker SDK）
  adapter_podman.go               // PodmanAdapter（runtime/podman.Client）

  mapper/
    container.go                  // dto.* → runtimeapi.*
    image.go
    volume.go
    network.go
    action.go                     // dto.Action ↔ runtimeapi.Action，含描述

  adapter_docker_test.go
  adapter_podman_test.go
  mapper_test.go
```

### PodmanAdapter 示例

```go
// internal/data/docker/service/adapter_podman.go
package service

import (
    "context"
    runtimepodman "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
    "github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
    runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type PodmanAdapter struct {
    client *runtimepodman.Client
}

func (a *PodmanAdapter) ListContainers(ctx context.Context, options runtimeapi.ContainerListOptions) ([]runtimeapi.ContainerSummary, error) {
    items, err := a.client.ListContainers(ctx, dto.ContainerListOptions{
        All: options.All, Limit: options.Limit, Filters: convertFilters(options.Filters),
    })
    if err != nil {
        return nil, mapRuntimeError(err, "container.list", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer}, runtimeapi.Podman)
    }
    return mapper.MapContainerSummaries(items), nil
}

func (a *PodmanAdapter) ExecuteContainer(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
    dtoAction, _ := mapper.DTOFromRuntimeAction(action)
    return a.client.ExecuteContainerAction(ctx, id, dtoAction, dto.ActionOptions{
        Force: options.Force, Timeout: options.Timeout, Signal: options.Signal, Name: options.Name,
    })
}

func (a *PodmanAdapter) ExecuteResource(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions) (runtimeapi.ActionResult, error) {
    dtoAction, desc := mapper.DTOFromRuntimeAction(action)
    _ = desc   // 调用方从 service 层取描述用于 UI / 审计
    result := runtimeapi.ActionResult{}
    dtoResult := dto.ActionResult{}
    var err error
    switch ref.Type {
    case runtimeapi.ResourceContainer:
        err = a.client.ExecuteContainerAction(ctx, ref.ID, dtoAction, mapper.DTOFromActionOptions(options), &dtoResult)
    case runtimeapi.ResourceImage:
        err = a.client.ExecuteImageAction(ctx, ref.ID, dtoAction, mapper.DTOFromActionOptions(options), &dtoResult)
    case runtimeapi.ResourceVolume:
        err = a.client.RemoveVolume(ctx, ref.ID, options.Force)  // 不需要 dtoResult
    case runtimeapi.ResourceNetwork:
        err = a.client.RemoveNetwork(ctx, ref.ID)
    default:
        return result, runtimeapi.UnsupportedError(string(ref.Type) + "." + string(action))
    }
    result.SpaceReclaimed = dtoResult.SpaceReclaimed
    result.References = dtoResult.References
    return result, err
}
```

### DockerAdapter 示例

```go
// internal/data/docker/service/adapter_docker.go
type DockerAdapter struct {
    client *docker.Client
}

func (a *DockerAdapter) ListContainers(ctx context.Context, options runtimeapi.ContainerListOptions) ([]runtimeapi.ContainerSummary, error) {
    // 走 Docker SDK，原有逻辑迁入此处
}
```

### 工厂选择

```go
// internal/data/docker/service/factory.go
func NewAdapter(rt runtimeapi.Type, dockerClient *docker.Client, podmanClient *runtimepodman.Client) runtimeapi.Engine {
    switch rt {
    case runtimeapi.Podman:
        return &PodmanAdapter{client: podmanClient}
    default:
        return &DockerAdapter{client: dockerClient}
    }
}
```

### 迁移策略（不破坏现有接口）

`engine_factory.go` 中的 `dockerEngine` / `podmanEngine` 不立刻删除，而是在其内部改为调用新 service：

```go
func (e *dockerEngine) Containers() runtimeapi.ContainerService {
    return service.NewUnifiedContainerService(e.Client)
}
```

旧 `containerService{client: c}.List(...)` 与 `podmanContainerService{client: c}.List(...)` 保留为**兼容实现**，标注 `// Deprecated:`。后续清理阶段统一移除。

---

## 阶段化执行（增量化，旧路径保留）

### Phase A — `gpgme` 仅 CGO（2026-07-24 修订）

1. 仅对 `go mod why github.com/proglottis/gpgme` 列出的**实际传染路径**加 `//go:build cgo`；其他 `aliases*.go` 若未被实际 import，**不强求加标签**。
2. 现状传染源：`internal/driver/podman/driver_cgo.go` + `client_default_driver_cgo.go`（携带 `go.podman.io/*`）。`internal/driver/podman/aliases.go` 当前未引 CGO 依赖，可不加 tag（保留为 fallback）。
3. 验证：
   ```bash
   # CGO 路径独立验证
   CGO_ENABLED=0 go list -deps ./internal/driver/podman/transport/rest | grep -E "gpgme|podman|buildah"   # 空
   go list -deps ./internal/driver/podman/transport/driver_cgo | grep gpgme                                          # 命中（仅 CGO 时）
   # 传染范围锁定
   go mod why -m github.com/proglottis/gpgme                                                                        # 只指向 driver_cgo
   CGO_ENABLED=0 go list -deps ./... | grep gpgme                                                                   # 空
   ```

### Phase B — 补齐 `dto/` 具名类型

1. 按 §"dto/ 包"表格逐一新增类型。
2. B 层与 C 层 JSON tag 等价的，使用 `type` 别名（双形态文件分别定义）。
3. 验证：
   ```bash
   CGO_ENABLED=0 go list -deps ./internal/data/runtime/podman/dto | grep podman   # 空
   ```

### Phase C — `runtime/podman.Client` 方法化

1. 新建 `driver_iface.go`、`driver_cgo.go`、`driver_nocgo.go`。
2. 新建各资源方法文件（无标签），内部分发 `Driver() / rest*`。
3. 旧包级函数（`ListContainersREST(ctx, *Client, ...)` 等）保留文件，标记 `// Deprecated:`，不影响旧路径。

### Phase D — 全仓匿名 struct 替换（2026-07-24 修订范围）

按"匿名结构体替换表"逐个替换，**范围仅限被 public API 路径引用的匿名 struct**：
- `runtime/podman/rest.go`：5+ 处（version / 错误响应 / 嵌套 stats / 嵌套 top 等）
- `runtime/helpers.go`：2 处（ProgressWriter/Reader）标 TODO，本 Phase 不处理
- `runtime/podman/containers_read.go`：1 处（top response）已具 dto，零变更

阶段验证（**修订**：仅检查 public API 路径，internal helper 保留）：
```bash
# 仅 public API 相关文件里不应有匿名 struct
grep -rn 'struct {' internal/data/runtime/podman/rest.go internal/data/runtime/podman/containers_read.go \
  | grep -v _test.go | grep -vE '^[^:]+:\s*//' || echo "clean"
```

### Phase D-2 — 双 mapper 拆分（**新增**，2026-07-24 用户决策）

将 `docker/service` 抽象层取消后，mapper 直接拆到两侧：

1. **Docker mapper** 拆出独立目录：
   - 新建 `internal/data/runtime/docker/mapper/` 目录
   - 从 `internal/data/runtime/docker/` 现有代码（`images.go`、`containers.go` 等）中提取 mapper 函数到独立文件
   - 不引入 `docker/service` 抽象

2. **Podman mapper** 提升自 `runtime/podman/mappers.go`：
   - 已有 `mappers.go`，把每一类资源（container/image/volume/network）的 mapper 函数拆到独立文件：`mapper_container.go`、`mapper_image.go` 等
   - 同时维护 `runtime/podman/mapper/` 子目录的命名

3. mapper 函数范围：仅做 `dto.*` ↔ `runtimeapi.*` 互转，不感知 A 层（Docker SDK）或 B 层（`go.podman.io/*`）。

```text
internal/data/runtime/docker/mapper/
    container.go      # Docker SDK type → runtimeapi.ContainerSummary
    image.go          # Docker SDK type → runtimeapi.ImageSummary
    volume.go         # Docker type → runtimeapi.Volume
    network.go        # Docker type → runtimeapi.Network
    action.go         # 字符串 ↔ runtimeapi.Action（含 i18n 描述）

internal/data/runtime/podman/mapper/
    container.go      # dto.* → runtimeapi.ContainerSummary
    image.go          # dto.* → runtimeapi.ImageSummary（含 ImageInspect 映射）
    volume.go         # dto.* → runtimeapi.Volume
    network.go        # dto.* → runtimeapi.Network
    action.go         # dto.Action 包装
```

4. 验证：双 mapper 目录零相互引用。
   ```bash
   grep -rn "internal/data/runtime/docker/mapper" internal/data/runtime/podman/mapper/    # 空
   grep -rn "internal/data/runtime/podman/mapper" internal/data/runtime/docker/mapper/   # 空
   ```

### Phase E — 验证

执行"验证矩阵"全部命令。

---

## 验证矩阵（2026-07-24 用户最终修订）

| 阶段 | 命令 | 期望 |
|---|---|---|
| A | `CGO_ENABLED=0 go list -deps ./internal/driver/podman \| grep -E "gpgme\|podman\|buildah"` | 空 |
| A | `CGO_ENABLED=0 go list -deps ./internal/driver/podman/transport/rest \| grep gpgme` | 空 |
| A | `go mod why github.com/proglottis/gpgme` 列出的引用方 | 只有 `internal/driver/podman/driver_cgo.go` / `client_default_driver_cgo.go`，且必须仅在 CGO 时命中 |
| B | `CGO_ENABLED=0 go list -deps ./internal/data/runtime/podman/dto \| grep podman` | 空 |
| C | `CGO_ENABLED=0 go build ./...` && `CGO_ENABLED=1 go build ./...` | 双通过 |
| C | **G2 验证**：`go vet ./...` 不报 `action string` 与 `dto.Action` 类型不匹配 | 干净（破坏性变更一次性迁移） |
| D | `grep -rn 'struct {' internal/data/runtime/podman/rest.go internal/data/runtime/podman/containers_read.go \| grep -v _test.go` | clean（public API 路径已清零；internal helper 残留另列 TODO） |
| D-2 | `grep -rn "internal/data/runtime/docker/mapper" internal/data/runtime/podman/mapper/` 与反向 | 双方均为空（双 mapper 目录零相互依赖） |
| D-2 | `find internal/data/runtime/{docker,podman}/mapper/ -name "*_test.go"` | 两个 mapper 目录都有独立测试，覆盖 Podman 私有动作与 Docker 通用动作互转 |
| E | `CGO_ENABLED=0 go test ./...` && `CGO_ENABLED=1 go test ./...` | 双通过 |
| F | `CGO_ENABLED=0 go list -deps ./... \| grep gpgme` | 空 |
| 终态 | `go vet ./...` | 通过 |
| 终态 | mapper 全部覆盖：Docker mapper 测试覆盖 SDK → `runtimeapi` 全部路径；Podman mapper 测试覆盖 `dto` → `runtimeapi` 全部路径 | 通过 |

**修订要点（2026-07-24 用户决策）**：
- A 验证按 `go mod why` 锁定传染源（不依赖普遍 build tag）
- C 验证破坏性迁移完成，不保留双签名（**G2/D6 修订**）
- D 验证仅 public API 路径（**G6 修订**）
- D-2 验证双 mapper 隔离（**G8 修订**）
- 删除 docker/service 验证项（**G9 取消**）

---

## 风险登记

| 风险 | 缓解 |
|---|---|
| `DriverBackend` 接口在两条构建下方法签名需保持一致 | `driver_iface.go`（无标签）声明接口；CGO 与非 CGO 文件分别实现 |
| `dto.Action` 与 `runtimeapi.Action` 枚举映射遗漏 | `docker/service/mapper/action_test.go` 覆盖所有 action 双向 |
| 旧 `docker/podman_*.go` 与新 `docker/service/adapter_podman.go` 共存易混淆 | 旧文件加 `// Deprecated:`；新文件以 `adapter_` 前缀 |
| 匿名 struct 散布多文件 | `grep -rn 'struct {'` 一次列举，逐个迁移；CI 加 vet 规则 |
| B 层 → C 层 JSON tag 不匹配 | `dump_api_test.go` 抓 Podman 实际响应做对照；`mapper_driver_cgo_test.go` 比对 |
| `docker/service` 统一入口若实现不一致，TUI 调用方体验分裂 | 引入 contract test：相同输入序列在 Docker / Podman 下应得到等价的 runtimeapi 域模型 |
| docker/ 旧 podman 调用方路径本轮未删，遗留代码堆积 | 在 Phase E 后单独建立后续清理任务（`TASK-023`），本轮**不**清理 |

---

## 阶段摘要

1. **A**：`aliases_*.go` 加 `//go:build cgo` + 新建 `*_nocgo.go` → `gpgme` 仅 CGO。
2. **B**：`dto/` 补齐全部具名类型，含别名双形态。
3. **C**：`runtime/podman.Client` 方法化，双模式通过 `Driver() DriverBackend` 派发。
4. **D**：全仓匿名 struct 替换为 `dto.*`。
5. **E**：`docker/service/` 建立统一入口；mapper 下沉到 `docker/service/mapper`。
6. **F**：验证矩阵全部通过；旧路径保留；后续清理任务另行排期。

---

## 决策项（2026-07-24 用户最终修订）

| 编号 | 议题 | 推荐 | 备选 | 状态 |
|---|---|---|---|---|
| D1 | B → C 别名在 non-CGO 下用同名结构还是全部独立 | 推荐：双形态文件分别定义，CGO 用 `type X =` 别名 | 全部独立具名 | 待定 |
| D2 | **取消**（与 G9 docker/service 取消同步） | — | — | — |
| D3 | 旧 `docker/podman_*.go` 与 `podmanContainerService` 是否在本轮删除 | 推荐：本轮**不删**，标注 `// Deprecated:`，TASK-023 收尾清理 | 本轮同步清理 | 待定 |
| D4 | **取消**（取消 `docker/service` 后不再需要 docker.Client.podman 字段） | — | — | — |
| D5 | `runtimeapi.Error` 等极少量公共类型能否出现在 `runtime/podman` | 推荐：允许（仅用于错误包装）；不允许 `runtimeapi.Action`/`ImageSummary` 等域模型 | 全 `dto.*` | 待定 |
| **D6** | ~~G2 双签名策略~~（用户决策：取消） | — | — | **取消** |
| **D7** | G10 build tag 范围 | 推荐：按 `go list -deps` 实际传递引用加 `//go:build cgo`；其他文件不加防御性 tag | 全文件统一加 `//go:build cgo` | **待定** |
| **D8** | **G8 mapper 双位置（修订）**：Docker 专属 mapper 落 `internal/data/runtime/docker/mapper/`；Podman 专属 mapper 落 `internal/data/runtime/podman/mapper/`（提升自现有 `mappers.go`） | 推荐：双 mapper 仅在各自 adapter 包内互转 | 两类 mapper 落在中间层 | **待定** |
| **D9** | ~~G9 docker/service 评审~~（用户决策：取消 docker/service 层） | — | — | **取消** |
| **D10** | G5 `dto.Action` 不能回引 `runtimeapi.Action` | 推荐：`dto.Action` 仅字符串枚举；任何含 `runtimeapi.*` 的常量映射落 `runtime/docker/mapper/action.go` 或 `runtime/podman/mapper/action.go`；`dto` 包零 `runtime/` 依赖 | `dto.Action` 挂 i18n 描述 | **待定** |

**用户最终决策（2026-07-24 18:00 后）总览**：
1. **G2/D6 取消**：`runtime/podman.Client` **不再保留**旧 `string action` 签名；一次性破坏性迁移到 `dto.Action`。Phase C 同时完成：
   - 删除 `*Client.ExecuteContainerAction(ctx, id, action string, ...)` 与 `*Client.ExecuteImageAction(ctx, id, action string, ...)` 的 `string` 重载
   - 替换为 `dto.Action` 单一签名
   - 所有调用方（`internal/tui/keyboard/*` 等）一次性同步迁移
2. **G4 简化**：mapper **不**集中在 `docker/service` 上，**直接拆到两侧**：`runtime/docker/mapper/` 与 `runtime/podman/mapper/`。
3. **G5/G9 取消**：**不引入** `internal/data/docker/service/` 统一入口层。`runtime.Engine` 仍是唯一上层抽象。

确认后转入执行阶段。