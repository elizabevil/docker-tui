# dtui 架构

本文描述当前实现，不记录历史规划。

## 技术栈

| 领域 | 当前实现 |
|---|---|
| 语言 | Go |
| CLI | `cobra` |
| TUI | `charm.land/bubbletea/v2` |
| 样式 | `charm.land/lipgloss/v2` |
| 容器引擎接入 | `github.com/docker/docker/client` |
| 配置 | YAML + 内嵌 JSONC 默认值 |

## 顶层启动流程

`cmd/docker-tui/main.go` 负责：

1. 解析 CLI 参数：`--config`、`--host`、`--theme`、`--podman`、`--list-themes`、`--lang`、`--version`
2. 按 `configVersion: 1` 严格加载配置与主题
3. 初始化 i18n 与尺寸格式化策略
4. 通过 `runtimeapi.BuildConnections(&cfg.Runtime, WithHostOverride(...))` 组合本地候选与配置连接，再交由 `runtimeapi.NewPool(runtimeinit.NewEngineFactory())` 创建连接池
5. 构造 `state.AppModel`
6. 启动 Bubble Tea 程序

初始化命令当前包含：

- 主机资源定时 tick
- Toast 定时 tick
- 首次容器引擎连接尝试
- `probeAllOnStart`：进入 runtime selector 前先并发探测所有连接，把结果合并成 `RuntimeProbeResult`，让首屏显示真实延迟

## 分层

```
cmd/docker-tui/
  main.go
    └── state.AppModel
          ├── internal/data/config
          ├── internal/data/docker
          ├── internal/data/i18n
          └── internal/tui
                ├── keyboard
                ├── state
                ├── ui
                └── update*.go
```

### `internal/data`

- `audit`：用户操作 trace、强类型目标、通知/操作历史投影和按日 JSONL sink
- `config`：配置默认值、加载、保存、主题加载
- `runtime`：SDK 无关的 Engine、Container/Volume/Network service、领域 DTO、错误、capability、筛选 options；`ConnectionPool` 持有注入的 `EngineFactory`，统一负责连接接入、ping 探测、健康刷新和反射层 typed-nil 接口清理
- `runtime/docker`：Docker 适配器（实现 `runtime.Engine`），通过 Docker SDK 与 docker-compatible REST API 提供容器/镜像/卷/网络等资源
- `runtime/podman`：Podman 运行时适配器，实现 `runtime.Engine`，负责把 `runtimeapi.*` 域对象映射到 Podman DTO、做后过滤、错误归一化和能力声明；具体 HTTP / CGO 传输由 `internal/driver/podman` 承担
- `i18n`：`zh` / `en` 文案
- `internal/runtimeinit`：连接池工厂注册点，导出 `NewEngineFactory()`，内部依赖 `runtime/docker` 与 `runtime/podman` 实现 Docker/Podman 派发，并通过反射处理 typed-nil 接口
- `internal/driver/podman`：Podman 传输与可选 CGO driver 边界，包含 `RESTClient`、`DriverBackend`、`dto` wire types 和错误封装；`gpgme` 仅 CGO 路径通过 `aliases_*.go` 与 `driver_cgo.go` build tag 隔离

### TASK-022 进度（2026-07-24 用户最终修订：取消 docker/service 统一入口层）

**当前架构**：`UI → runtime.Engine → {PodmanEngine (runtime/podman) / DockerEngine (runtime/docker)} → driver (仅 Podman 走) → transport`

TASK-022 Phase E（docker/service 统一入口）整 Phase **取消**。两个 mapper 直接拆双，各自落：

- `internal/data/runtime/docker/mapper/`：Docker SDK type ↔ `runtimeapi.*`
- `internal/data/runtime/podman/mapper/`：`dto.*` ↔ `runtimeapi.*`（提升自现有 `runtime/podman/mappers.go`）

**用户决策（2026-07-24）**：
- **G2/D6**：取消 `string action` 双签名；一次性破坏性迁移到 `dto.Action`
- **G4**：跳过 `docker/service` 中间抽象层；mapper 直接进两侧
- **G5**：不再经 `docker/service` 层；`dto.Action` ↔ `runtimeapi.Action` 由各 mapper 内部完成
- **G9**：docker/service 统一入口取消

TASK-022 最终范围(A / D / F 完成,B / C / D-2 / E 经范围决策取消)详见 [feature-todo-list.md §6.1](feature-todo-list.md);原始设计存档于 [design/archived/podman-rest-migration.md](../design/archived/podman-rest-migration.md)。

### `internal/tui`

- `state`：顶层模型、领域子状态和资源列表模型；`AppModel` 只组合命名状态域，且不依赖 UI 组件包
- `keyboard`：所有交互入口、模式切换和资源操作命令
- `update*.go`：Bubble Tea 消息分发与状态更新
- `ui/app`：顶层布局
- `ui/pages`：各页面渲染
- `ui/component`、`ui/widget`：复用组件与局部部件
- `tables`：表格/布局 JSONC 配置

### AppModel 状态树

```text
AppModel
├── Dependencies
├── Connection
├── Navigation
├── Dialog
├── Feedback
├── Log
├── Detail
├── Exec
├── Compose
├── Confirm
├── Selection
├── Resources
├── Metrics
├── Viewport
└── Processes
```

所有状态域均使用命名字段访问。状态对象维护同步转换和自身不变量；`keyboard` 与 `update*.go` 协调异步命令、runtime IO 和跨域更新；`ui` 只读取状态并渲染。

## 实际数据流

### 启动和刷新

```
connectDocker()
  -> state.DockerConnected
  -> handleDockerConnected()
  -> keyboard.FetchAll()
  -> ContainersLoaded / ImagesLoaded / VolumesLoaded / NetworksLoaded
  -> RenderApp()
```

### 键盘输入

```
tea.KeyPressMsg
  -> keyboard.HandleKeyPress()
  -> tea.Cmd (可选)
  -> XxxLoaded / XxxActioned / ExecOutput ...
  -> tui.Update()
  -> ui/app.RenderApp()
```

### Stats

```
state.StatsTick
  -> handleStatsTick()
  -> 对当前容器列表逐个发起 FetchStats()
  -> state.StatsReceived
```

### 用户操作审计

```
keyboard action
  -> audit.Service.Begin() (requested / started)
  -> async XxxActioned carries audit.Trace
  -> audit.Service.Finish() (succeeded / failed / cancelled)
  -> NotificationProjector + OperationLogProjector + FileSink
```

审计目标使用 `type / id / name` 统一字段和资源专属强类型 meta。顶部 Toast 消费终态通知，Footer 显示最近操作记录；普通 UI 提示只经过通知投影。日志默认写入 `~/.config/docker-tui/logs/audit-YYYY-MM-DD.jsonl`。

## 当前已实现能力

- 多资源面板：容器、镜像、卷、网络、Compose
- 连接池：本地 Podman / Docker 自动尝试，可运行时切换；`RefreshAll` 统一延迟与状态探测，Podman 5.x 的版本协商自动回退到 `/v4.0.0/libpod/version`
- TLS 连接：默认验证证书，支持显式 `insecureSkipVerify`；UI 区分 TLS configured、verified 和 insecure，证书材料在首次连接时加载
- 统一只读资源：Container/Image/Volume/Network list 使用统一 options 和筛选契约；Podman CGO 使用 bindings，非 CGO 或 TLS/API override 使用共享 REST transport
- 高频容器操作：Pause/Unpause、Rename、Top 和结构化 Port bindings 当前仍主要通过 Docker-compatible API，等待 action/top service 迁移
- 容器操作：启动、停止、重启、Kill、Logs、Exec、Inspect
- 镜像操作：拉取、删除、Prune、Tag、Push、Save、Load、Inspect（Docker 与 Podman 都有结构化详情）；Podman 适配器已对齐 `runtime.ImageDetail` 字段（ID、RepoTags、Architecture、OS、Driver、LayerCount、Runtime config、History）
- 卷/网络：列表、删除、详情
- Compose：从容器标签聚合项目与服务视图
- UI 能力：搜索、命令模式、帮助页、主题、i18n、Header 开关
- 用户操作审计：资源操作 trace、通知与 Footer 投影、按日 JSONL 落盘

镜像详情优先使用 `InspectImageDetail()` 的结构化模型。普通镜像通过 runtime `ImageHistory` API 加载 layer history；manifest list 使用列表阶段解析的平台变体，不调用不适用的 layer history API。Podman 通过 `/v4.0.0/libpod/images/{id}/json` 与 `/v4.0.0/libpod/images/{id}/history` 双接口实现同等结构。

## 连接池与引擎工厂分层

`ConnectionPool` 仅依赖 `EngineFactory` 函数类型，不再持有任何全局可变状态：

```text
runtimeinit.NewEngineFactory()     ┐
                                  ├─> runtime.EngineFactory (func)
runtime/docker.NewClient          ┘
runtime/podman.NewEngine           ┘
                                       │
                                       ▼
runtimeapi.NewPool(factory)  ─► ConnectionPool.Connect / RefreshAll
                                       │
                                       ▼
                              runtime.Engine (interface)
                                       │
                          ┌────────────┴─────────────┐
                          ▼                          ▼
                 runtime/docker.Client     runtime/podman.PodmanEngine
```

调用方通过注入工厂同时避免：
- 全局 `engineFactory` 变量被任何 init() 副作用污染
- runtime 与 docker/podman 适配器形成导入循环
- typed-nil 接口逃逸到 `ConnectionPool.Connect`/`RefreshAll`（`sanitizeEngine` + `engineIsUsable` 双层防御）

## 代码验证后的注意点

这些内容在旧文档里容易被写错，这里按当前代码记录：

- `internal/driver/podman` 负责 REST transport、CGO driver、wire DTO 和 transport 级错误分类；`internal/data/runtime/podman` 负责 runtime.Engine 适配、资源 service、post-filter、domain mapping 和运行时能力声明，依赖方向是 `runtime → driver`。
- `runtime/podman` 的公开边界是 `runtime.Engine`：上层只看到 service 接口与 domain DTO，不直接接触 Podman REST 细节，也不依赖 CGO 类型。
- `runtime/init` 包负责把 docker / podman adapter 装配成可注入的 `EngineFactory`，并通过反射处理 typed-nil 接口。
- 连接池 `Connect` 只负责单连接的创建与 ping；`RefreshAll` 遍历所有 host，更新 latency/probedAt/state；`PingLoop` 仅定期 ping 已存在引擎的连接。
- `Client.Raw()` 仍被 Stats 和 Exec 路径使用；TUI 尚未完全解除 Docker SDK 依赖。
- 容器 stats 当前不是“仅聚焦项轮询”，而是对当前容器列表逐项请求，轮询间隔来自 `config.Docker.StatsPollSec`，默认 3 秒。
- Compose 面板不是直接解析 `compose.yaml`，而是基于容器上的 `com.docker.compose.*` labels 聚合。
- `config.Keymap` 已通过统一动作注册表接入 `keyboard.HandleKeyPress()`；当前支持动作级默认绑定覆盖，用户级上下文覆盖尚未开放。
- CLI `Use` 名称是 `dtui`，但仓库中仍有部分构建脚本输出文件名 `docker-tui`；文件名和 Cobra `Use` 目前未完全统一。
- Podman 适配器中 Image.Inspect 已完整实现，详情页支持 Runtime（WorkingDir / Cmd / Entrypoint / Env / Labels / ExposedPorts / Volumes）和 History 层。

## 配置入口

- 默认配置路径：`~/.config/docker-tui/config.yml`
- 连接配置：`runtime.default`、`runtime.health`、`runtime.connections`
- 连接字段：`name`、`driver`、`endpoint` 和独立 `tls` 配置
- 本地 Docker / Podman 会始终被探测并加入候选，`runtime.discovery` 仅保留为配置结构字段，不再屏蔽本地驱动
- 旧连接字段不兼容；未知字段与无效连接会在启动前返回配置错误
- 默认配置模板：`internal/data/config/default.jsonc`
- 布局窗口配置：`internal/tui/ui/app/app.jsonc`
- 主题：`internal/data/config/themes/*.jsonc`
- 表格和组件配置：`internal/tui/tables/*.jsonc`、`internal/tui/ui/component/*.jsonc`
