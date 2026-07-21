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
| 预览系统 | `docs/ui-preview` 中的 Vue 3 + Vite |

## 顶层启动流程

`cmd/docker-tui/main.go` 负责：

1. 解析 CLI 参数：`--config`、`--host`、`--theme`、`--podman`、`--list-themes`、`--lang`、`--version`
2. 按 `configVersion: 1` 严格加载配置与主题
3. 初始化 i18n 与尺寸格式化策略
4. 创建 `internal/data/docker.ConnectionPool`
5. 构造 `state.AppModel`
6. 启动 Bubble Tea 程序

初始化命令当前包含：

- 主机资源定时 tick
- Toast 定时 tick
- 首次容器引擎连接尝试

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
- `docker`：Docker / Podman 客户端封装、连接池、资源操作、stats、inspect、events
- `i18n`：`zh` / `en` 文案

### `internal/tui`

- `state`：顶层模型、领域子状态和资源列表模型；`AppModel` 当前已下沉 Connection、Navigation 与 Dialog 输入状态，其他状态域按 `TASK-015` 渐进迁移
- `keyboard`：所有交互入口、模式切换和资源操作命令
- `update*.go`：Bubble Tea 消息分发与状态更新
- `ui/app`：顶层布局
- `ui/pages`：各页面渲染
- `ui/component`、`ui/widget`：复用组件与局部部件
- `tables`：表格/布局 JSONC 配置

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
- 连接池：本地 Podman / Docker 自动尝试，可运行时切换
- 容器操作：启动、停止、重启、Kill、Logs、Exec、Inspect
- 镜像操作：拉取、删除、Prune、Detail、导出/调试入口
- 卷/网络：列表、删除、详情
- Compose：从容器标签聚合项目与服务视图
- UI 能力：搜索、命令模式、帮助页、主题、i18n、Header 开关
- 用户操作审计：资源操作 trace、通知与 Footer 投影、按日 JSONL 落盘

镜像详情优先使用 `InspectImageDetail()` 的结构化模型。普通镜像通过 runtime `ImageHistory` API 加载 layer history；manifest list 使用列表阶段解析的平台变体，不调用不适用的 layer history API。

## 代码验证后的注意点

这些内容在旧文档里容易被写错，这里按当前代码记录：

- `internal/data/docker/events.go` 已实现 `Client.Events()`，但启动流程里没有订阅事件流；当前刷新主路径仍是连接后 `FetchAll()` 和显式刷新。
- 容器 stats 当前不是“仅聚焦项轮询”，而是对当前容器列表逐项请求，轮询间隔来自 `config.Docker.StatsPollSec`，默认 3 秒。
- Compose 面板不是直接解析 `compose.yaml`，而是基于容器上的 `com.docker.compose.*` labels 聚合。
- `config.Keymap` 已通过统一动作注册表接入 `keyboard.HandleKeyPress()`；当前支持动作级默认绑定覆盖，用户级上下文覆盖尚未开放。
- CLI `Use` 名称是 `dtui`，但仓库中仍有部分构建脚本输出文件名 `docker-tui`；文件名和 Cobra `Use` 目前未完全统一。

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
