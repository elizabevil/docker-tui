# Docker / Podman 功能对比与 docker-tui 可添加功能清单

> 基于 `docker -h` 和 `podman -h` 文档分析，对比 docker-tui 已有功能，整理可扩展方向。

---

## 一、Docker vs Podman 命令对比

### 容器生命周期

| 命令 | Docker | Podman | docker-tui | 备注 |
|------|:------:|:------:|:----------:|------|
| run | ✓ | ✓ | ✗ | 创建并运行容器 (TUI 场景意义有限) |
| create | ✓ | ✓ | ✗ | 创建容器不启动 |
| start | ✓ | ✓ | ✓ | |
| stop | ✓ | ✓ | ✓ | |
| restart | ✓ | ✓ | ✓ | |
| kill | ✓ | ✓ | ✓ | |
| rm | ✓ | ✓ | ✓ | |
| pause | ✓ | ✓ | ✗ | |
| unpause | ✓ | ✓ | ✗ | |
| rename | ✓ | ✓ | ✗ | |
| update | ✓ | ✓ | ✗ | 更新资源限制 |
| wait | ✓ | ✓ | ✗ | 等待容器停止 |
| attach | ✓ | ✓ | ✗ | 附加到运行中容器 |
| exec | ✓ | ✓ | ✓ | |
| logs | ✓ | ✓ | ✓ | |
| stats | ✓ | ✓ | ✓ | |
| top | ✓ | ✓ | ✗ | 查看容器进程 |
| port | ✓ | ✓ | ✗ | 查看端口映射 |
| cp | ✓ | ✓ | ✗ | 文件复制 |
| diff | ✓ | ✓ | ✗ | 文件系统变更 |
| export | ✓ | ✓ | ✗ | 导出文件系统 |
| commit | ✓ | ✓ | ✗ | 容器提交为镜像 |
| inspect | ✓ | ✓ | ✓ | |
| ps | ✓ | ✓ | ✓ | |

### 镜像生命周期

| 命令 | Docker | Podman | docker-tui | 备注 |
|------|:------:|:------:|:----------:|------|
| pull | ✓ | ✓ | ✓ | |
| push | ✓ | ✓ | ✗ | |
| images | ✓ | ✓ | ✓ | |
| rmi | ✓ | ✓ | ✓ | |
| tag | ✓ | ✓ | ✗ | 打标签 |
| untag | ✗ | ✓ | ✗ | Podman 独有 |
| build | ✓ | ✓ | ✗ | 构建镜像 |
| history | ✓ | ✓ | ✓ | 已内嵌在详情 |
| save | ✓ | ✓ | ✗ | 导出为 tar |
| load | ✓ | ✓ | ✗ | 从 tar 导入 |
| import | ✓ | ✗ | ✗ | Docker 独有 |
| search | ✓ | ✓ | ✗ | 搜索仓库 |
| inspect | ✓ | ✓ | ✓ | |

### 卷 / 网络

| 命令 | Docker | Podman | docker-tui | 备注 |
|------|:------:|:------:|:----------:|------|
| volume create | ✓ | ✓ | ✗ | |
| volume rm | ✓ | ✓ | ✓ | |
| volume ls | ✓ | ✓ | ✓ | |
| volume inspect | ✓ | ✓ | ✓ | |
| volume prune | ✓ | ✓ | ✗ | |
| network create | ✓ | ✓ | ✗ | |
| network rm | ✓ | ✓ | ✓ | |
| network ls | ✓ | ✓ | ✓ | |
| network inspect | ✓ | ✓ | ✓ | |
| network prune | ✓ | ✓ | ✗ | |

### 系统管理

| 命令 | Docker | Podman | docker-tui | 备注 |
|------|:------:|:------:|:----------:|------|
| info | ✓ | ✓ | 部分 | 仅连接信息 toast |
| version | ✓ | ✓ | ✗ | |
| system df | ✓ | ✓ | ✗ | 磁盘使用 |
| system prune | ✓ | ✓ | ✗ | 系统清理 |
| events | ✓ | ✓ | ✓ | 已有事件监听 |
| login | ✓ | ✓ | ✗ | 登录仓库 |
| logout | ✓ | ✓ | ✗ | 登出仓库 |

### Compose

| 命令 | Docker | Podman | docker-tui | 备注 |
|------|:------:|:------:|:----------:|------|
| compose up | ✓ | ✓ | ✗ | |
| compose down | ✓ | ✓ | ✓ | |
| compose start | ✓ | ✓ | ✓ | |
| compose stop | ✓ | ✓ | ✓ | |
| compose restart | ✓ | ✓ | ✗ | |
| compose logs | ✓ | ✓ | ✓ | |
| compose ps | ✓ | ✓ | ✗ | |
| compose build | ✓ | ✓ | ✗ | |
| compose pull | ✓ | ✓ | ✗ | |

### Podman 独有

| 命令 | Podman | docker-tui | 备注 |
|------|:------:|:----------:|------|
| pod create/rm/ls/start/stop | ✓ | ✗ | Pod 管理 |
| secret create/rm/ls | ✓ | ✗ | Secret 管理 |
| artifact | ✓ | ✗ | OCI artifact |
| kube play/generate | ✓ | ✗ | YAML 播放 |
| mount/unmount | ✓ | ✗ | 挂载 rootfs |
| init | ✓ | ✗ | 初始化容器 |
| farm | ✓ | ✗ | 远程构建农场 |
| generate | ✓ | ✗ | 生成 YAML |

### Docker 独有

| 命令 | Docker | docker-tui | 备注 |
|------|:------:|:----------:|------|
| buildx | ✓ | ✗ | 多平台构建 |
| plugin | ✓ | ✗ | 插件管理 |
| trust | ✓ | ✗ | 镜像信任 |
| context | ✓ | ✗ | 上下文管理 |
| checkpoint | ✓ | ✗ | 容器检查点 |
| swarm/config/node/service/stack | ✓ | ✗ | Swarm 集群 |

---

## 二、docker-tui 已实现功能

| 类别 | 已实现操作 |
|------|-----------|
| **容器** | start, stop, restart, kill, remove, logs, stats, exec, inspect |
| **镜像** | pull, remove, prune, inspect (含 history), export (shell), debug run (shell) |
| **卷** | list, inspect, remove |
| **网络** | list, inspect, remove |
| **Compose** | start, stop, down, logs, 展开详情 |
| **其他** | 刷新, 过滤, 命令面板, 运行时切换, 审计日志, i18n, 事件监听 |

---

## 三、可添加功能清单 (按优先级)

### P0 — 核心运维高频操作

| 功能 | Docker | Podman | 复杂度 | 说明 |
|------|:------:|:------:|:------:|------|
| **pause / unpause** | ✓ | ✓ | 低 | 暂停/恢复容器进程 |
| **rename** | ✓ | ✓ | 低 | 重命名容器 |
| **top** | ✓ | ✓ | 低 | 查看容器运行进程 |
| **port** | ✓ | ✓ | 低 | 查看端口映射 |
| **volume create** | ✓ | ✓ | 中 | 创建卷 |
| **network create** | ✓ | ✓ | 中 | 创建网络 |
| **volume prune** | ✓ | ✓ | 低 | 清理未使用卷 |
| **network prune** | ✓ | ✓ | 低 | 清理未使用网络 |

### P1 — 镜像生命周期补全

| 功能 | Docker | Podman | 复杂度 | 说明 |
|------|:------:|:------:|:------:|------|
| **tag** | ✓ | ✓ | 低 | 给镜像打标签 |
| **untag** | ✗ | ✓ | 低 | 移除镜像标签 (Podman) |
| **push** | ✓ | ✓ | 中 | 推送镜像到仓库 |
| **save / load** | ✓ | ✓ | 中 | 镜像导出/导入 tar |
| **build** | ✓ | ✓ | 高 | 构建镜像 (需 Dockerfile 路径) |

### P2 — 容器高级操作

| 功能 | Docker | Podman | 复杂度 | 说明 |
|------|:------:|:------:|:------:|------|
| **update** | ✓ | ✓ | 中 | 更新容器资源限制 |
| **diff** | ✓ | ✓ | 低 | 查看文件系统变更 |
| **export** | ✓ | ✓ | 低 | 导出容器文件系统 |
| **commit** | ✓ | ✓ | 中 | 容器提交为镜像 |
| **wait** | ✓ | ✓ | 低 | 等待容器停止 |
| **cp** | ✓ | ✓ | 高 | 文件复制 |
| **healthcheck** | ✓ | ✓ | 中 | 健康检查管理 |

### P3 — 系统级操作

| 功能 | Docker | Podman | 复杂度 | 说明 |
|------|:------:|:------:|:------:|------|
| **system df** | ✓ | ✓ | 低 | 磁盘使用情况 |
| **system prune** | ✓ | ✓ | 中 | 系统级清理 |
| **info** | ✓ | ✓ | 低 | 完整系统信息 |
| **version** | ✓ | ✓ | 低 | 版本信息 |
| **login / logout** | ✓ | ✓ | 中 | 仓库认证 |

### P4 — Podman 独有功能

| 功能 | 复杂度 | 说明 |
|------|:------:|------|
| **pod 管理** | 高 | 创建/删除/启停/列表 pod |
| **secret 管理** | 中 | 创建/删除/列表 secret |
| **artifact 管理** | 高 | OCI artifact 管理 |
| **kube** | 高 | 从 YAML 播放容器/pod |
| **generate** | 中 | 根据容器/pod 生成 YAML |
| **mount / unmount** | 中 | 挂载容器 rootfs |

### P5 — Docker 独有功能

| 功能 | 复杂度 | 说明 |
|------|:------:|------|
| **buildx** | 高 | 多平台构建 |
| **context** | 低 | 上下文管理 |
| **checkpoint** | 中 | 容器检查点 |
| **plugin** | 高 | 插件管理 |

### P6 — Compose 增强

| 功能 | Docker | Podman | 复杂度 | 说明 |
|------|:------:|:------:|:------:|------|
| **compose up** | ✓ | ✓ | 中 | 启动完整栈 |
| **compose restart** | ✓ | ✓ | 低 | 重启服务 |
| **compose build** | ✓ | ✓ | 高 | 构建服务镜像 |
| **compose pull** | ✓ | ✓ | 中 | 拉取服务镜像 |
| **compose ps** | ✓ | ✓ | 低 | 列出服务状态 |

---

## 四、推荐实现顺序

### 第一批 — 低成本高价值 (1-2 天)

1. **pause / unpause** — 容器生命周期补全
2. **rename** — 容器重命名
3. **top** — 查看容器进程
4. **port** — 查看端口映射
5. **volume create** — 创建卷
6. **network create** — 创建网络
7. **volume prune / network prune** — 资源清理

### 第二批 — 镜像完善 (2-3 天)

1. **tag / untag** — 镜像标签管理
2. **push** — 推送镜像
3. **save / load** — 镜像归档
4. **system df** — 磁盘使用

### 第三批 — 高级功能 (3-5 天)

1. **update** — 容器资源限制
2. **diff** — 文件系统变更
3. **export** — 文件系统导出
4. **commit** — 容器提交为镜像
5. **healthcheck** — 健康检查管理

### 第四批 — Podman 独有 (5-7 天)

1. **pod 管理** — Pod 面板
2. **secret 管理** — Secret 面板
3. **kube** — YAML 播放

---

## 五、UI 设计建议

### 新增面板

| 面板 | 说明 | 适用 |
|------|------|------|
| `PanelPods` | Pod 列表 | Podman only |
| `PanelSecrets` | Secret 列表 | Podman only |

### 容器面板新增操作键

| 按键 | 操作 | 说明 |
|------|------|------|
| `Ctrl+P` | pause | 暂停容器 |
| `Ctrl+U` | unpause | 恢复容器 |
| `Ctrl+N` | rename | 重命名 (弹出输入框) |
| `Shift+T` | top | 查看进程 |
| `Shift+P` | port | 查看端口 |

### 镜像面板新增操作键

| 按键 | 操作 | 说明 |
|------|------|------|
| `Ctrl+T` | tag | 打标签 (弹出输入框) |
| `Ctrl+Shift+P` | push | 推送镜像 |

### 卷/网络面板新增操作键

| 按键 | 操作 | 说明 |
|------|------|------|
| `Shift+C` | create | 创建资源 (弹出输入框) |
| `Shift+P` | prune | 清理未使用资源 |

---

## 六、API 映射表

### 容器操作

| 操作 | Docker SDK | Podman SDK |
|------|-----------|------------|
| pause | `ContainerPause(ctx, id)` | `ContainerPause(id)` |
| unpause | `ContainerUnpause(ctx, id)` | `ContainerUnpause(id)` |
| rename | `ContainerRename(ctx, id, name)` | `ContainerRename(id, name)` |
| top | `ContainerTop(ctx, id, args)` | `ContainerTop(id, args)` |
| update | `ContainerUpdate(ctx, id, config)` | `ContainerUpdate(id, config)` |
| cp | `CopyFromContainer(ctx, id, path)` | `CopyFromContainer(id, path)` |
| export | `ContainerExport(ctx, id)` | `ExportContainer(id)` |
| diff | `ContainerDiff(ctx, id)` | `ContainerDiff(id)` |
| commit | `ContainerCommit(ctx, id, ref)` | `CommitContainer(id, ref)` |
| wait | `ContainerWait(ctx, id, cond)` | `WaitContainer(id, cond)` |

### 镜像操作

| 操作 | Docker SDK | Podman SDK |
|------|-----------|------------|
| tag | `ImageTag(ctx, id, ref)` | `TagImage(id, ref)` |
| push | `ImagePush(ctx, ref, opts)` | `PushImage(id, ref)` |
| save | `ImageSave(ctx, ids)` | `SaveImage(ids)` |
| load | `ImageLoad(ctx, reader)` | `LoadImage(reader)` |
| build | `ImageBuild(ctx, reader, opts)` | `BuildImage(opts)` |

### 卷/网络操作

| 操作 | Docker SDK | Podman SDK |
|------|-----------|------------|
| volume create | `VolumeCreate(ctx, opts)` | `VolumeCreate(opts)` |
| volume prune | `VolumesPrune(ctx, filters)` | `PruneVolumes(filters)` |
| network create | `NetworkCreate(ctx, name, opts)` | `NetworkCreate(name, opts)` |
| network prune | `NetworksPrune(ctx, filters)` | `PruneNetworks(filters)` |

---

## 七、当前未实现能力与前置条件

> 这一节把“命令对比”转成“当前项目还缺什么”。  
> 其中一部分可以直接做，另一部分需要先完成连接模型统一、TLS 接线和 Docker / Podman 统一适配。

### 7.1 连接与运行时前置

这类工作不是业务命令本身，但它决定后续功能能否稳定落地。

| 任务 | 说明 | 关系到的能力 |
|------|------|-------------|
| 统一 `RuntimeConn` / `ConnectionSpec` | 把 Docker / Podman / 远程连接收敛到同一模型 | 所有命令 |
| TLS 独立配置 | 连接参数不再混在旧配置里 | 远程连接、证书错误提示 |
| 本地 Docker / Podman 默认发现 | 无论是否配置，都先探测本地驱动 | 首次启动、切换、F2 选择框 |
| 去重与有序追加 | 本地候选先展示，配置连接再追加，避免重复 | 多连接、连接选择框 |
| 健康检测配置化 | 默认 3s，可调整 | 自动重连、连接状态、失败恢复 |
| 单连接模式 | `--host` 指定时不进入连接切换 | 启动流程、F2 行为 |

### 7.2 可直接进入实现的业务缺口

这些能力不依赖新的交互模型，可以在连接体系稳定后按批次推进。

| 优先级 | 能力 | 说明 |
|------|------|------|
| P0 | `pause` / `unpause` | 容器生命周期补全，回报高、实现简单 |
| P0 | `rename` | 常用运维动作，UI 交互简单 |
| P0 | `top` | 查看容器进程，适合放入详情或快捷动作 |
| P0 | `port` | 查看端口映射，和容器列表联动价值高 |
| P0 | `volume create` | 卷管理最基础的创建能力 |
| P0 | `network create` | 网络管理最基础的创建能力 |
| P0 | `volume prune` / `network prune` | 清理类操作，和资源管理场景一致 |
| P1 | `tag` / `push` | 镜像分发常用，适合镜像面板扩展 |
| P1 | `save` / `load` | 离线迁移与归档，适合镜像面板与命令面板共用 |
| P1 | `build` | 能力强但交互复杂，建议单独设计输入流程 |
| P2 | `update` | 资源限制调整，需补充表单或参数面板 |
| P2 | `diff` / `export` / `commit` / `wait` / `cp` | 属于高级容器操作，适合第二阶段补齐 |
| P2 | `system df` / `system prune` | 系统级能力，适合放到全局命令面板 |
| P3 | `pod` / `secret` / `kube` / `generate` / `mount` | Podman 独有，建议在统一模型稳定后再做 |
| P3 | `buildx` / `context` / `checkpoint` / `plugin` | Docker 独有，优先级低于通用运维能力 |

### 7.3 适合当前仓库的落地顺序

1. 先完成连接模型与运行时统一，保证 Docker 和 Podman 走同一条主链路。
2. 再补最常用的容器操作：`pause`、`unpause`、`rename`、`top`、`port`。
3. 然后补资源创建与清理：`volume create`、`network create`、`volume prune`、`network prune`。
4. 接着补镜像工作流：`tag`、`push`、`save`、`load`、`build`。
5. 最后再推进高级容器能力和 Podman / Docker 独有功能。

### 7.4 当前不建议立即做的项

| 能力 | 原因 |
|------|------|
| `compose up` / `compose build` / `compose pull` | 需要更完整的服务编排交互，不适合在基础连接模型未收敛前推进 |
| `buildx` | 多平台和上下文管理会显著增加 UI 和参数复杂度 |
| `pod` / `secret` | 属于 Podman 专有能力，当前阶段更适合先完成通用能力 |
| `checkpoint` | 用户场景偏窄，优先级低于通用运维动作 |

---

## 八、建议的下一批任务拆分

> 下面这组任务适合直接落到后续任务清单，作为“可实施”的版本。
> `TASK-RUN-*` 是分析阶段候选编号，实施状态统一以
> [后续需求实施任务清单](future-requirements-task-list.md) 中的主编号为准。

| 候选编号 | 主任务编号 | 任务 | 当前状态 |
|---|---|---|---|
| `TASK-RUN-001` | `TASK-004`、`TASK-006` | 连接模型统一与本地驱动去重 | `done` |
| `TASK-RUN-002` | `TASK-005`、`TASK-016` | TLS 接线与错误展示 | `done` |
| `TASK-RUN-003` | `TASK-017` | 容器 `pause` / `unpause` / `rename` / `top` / `port` | `done` |
| `TASK-RUN-004` | `TASK-009` | 卷 / 网络创建与清理 | `todo` |
| `TASK-RUN-005` | `TASK-018` | 镜像标签与传输工作流 | `todo` |
| `TASK-RUN-006` | `TASK-019` | 高级容器操作 | `todo` |
| `TASK-RUN-007` | `TASK-020` | Docker / Podman 专有能力评估 | `todo` |
| `TASK-RUN-008` | `TASK-021` | Docker / Podman 正式双 adapter 与统一 runtime driver | `todo` |

### 建议结论

如果目标是尽快把 docker-tui 的“通用价值”做厚，优先级应该是:

1. 连接模型统一
2. TLS 与连接去重
3. 高频容器运维动作
4. 卷 / 网络创建与清理
5. 镜像工作流补齐

Podman 专有能力可以保留在后续迭代，不要和通用能力抢第一批资源。
