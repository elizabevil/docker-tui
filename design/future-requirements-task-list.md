# 后续需求实施任务清单

> 建立日期: 2026-07-20
> 最近整理: 2026-07-24
> 依据: [后续需求与规划讨论](future-requirements-discussion.md)、[Docker / Podman 能力分析](podman-capabilities-analysis.md)
> 规则: 本文是后续工作的唯一主任务台账；其他设计文档中的任务编号仅作为来源参考。

## 状态定义

| 状态 | 含义 |
|---|---|
| `todo` | 范围已明确，尚未开始 |
| `in_progress` | 已有未完成的代码或文档改动 |
| `done` | 已实现、通过回归并提交 |
| `blocked` | 缺少外部条件或尚未完成产品决策 |

任务只有满足文末“完成定义”后才能标记为 `done`。分析文档完成不等于代码能力完成。

## 当前概况

| 状态 | 数量 |
|---|---:|
| `done` | 16 |
| `in_progress` | 0 |
| `todo` | 5 |
| `blocked` | 0 |

当前执行队列：

1. 执行 `TASK-011`，完善 Compose / 容器 / 镜像联动刷新。
2. 执行 `TASK-010`，批量操作扩展与部分成功反馈。

最近一次维护说明：TASK-021 已增加 Podman `Image.Inspect` 与连接池结构性修复记录，状态保持 `done`。

## 已完成基础

| 编号 | 任务 | 优先级 | 状态 | 完成证据 |
|---|---|---:|---|---|
| `TASK-001` | 连接失败错误投影与断开状态展示 | P0 | `done` | 启动失败不再显示虚假 Docker，目标与错误可见 |
| `TASK-002` | F2 连接选择框 | P0 | `done` | 支持上下选择、Enter 连接、Esc 取消和失败项错误保留 |
| `TASK-003` | 可配置连接健康检测 | P0 | `done` | 默认 3 秒、失败阈值、恢复通知且不自动切换 |
| `TASK-004` | Docker / Podman 运行时识别与统一基线 | P1 | `done` | `RuntimeType`、连接池和客户端使用统一类型；能力差距已有分析 |
| `TASK-005` | TLS 配置与客户端接线 | P0 | `done` | `RuntimeConn.TLS`、CA、客户端证书和 ServerName 已接入客户端 |
| `TASK-006` | endpoint / runtime identity 去重 | P1 | `done` | 本地候选与配置连接按规范化 key 去重并保持顺序 |
| `TASK-007` | 新配置 schema 强制校验 | P1 | `done` | 解码后按配置子结构校验，运行期直接使用已验证的健康参数 |

相关提交：`7cf479a`、`948d9d2`、`aa7567d`、`4fe226d`、`801eb76`、`f71d405`。

## 当前架构任务

| 编号 | 任务 | 优先级 | 状态 | 已完成 / 剩余范围 |
|---|---|---:|---|---|
| `TASK-015` | [拆分 `AppModel` 状态域并封装状态方法](app-model-refactoring.md) | P1 | `done` | `AppModel` 仅组合 14 个命名状态域；状态生命周期、独立输入和只读渲染边界均已落地 |

`TASK-015` 已按领域逐步迁移完成。顶层 `AppModel` 负责 Bubble Tea 协调，子状态维护自身不变量。

## Runtime 与错误体验

| 编号 | 任务 | 优先级 | 状态 | 依赖 | 验收重点 |
|---|---|---:|---|---|---|
| `TASK-016` | TLS / 证书错误分类与安全状态展示 | P0 | `done` | TASK-005 | 已结构化区分 CA、客户端证书、主机名、握手和网络错误；选择框、Toast 与状态栏使用本地化安全文案，并区分 TLS 已配置/已验证/不安全 |

TLS 配置与客户端链路由 `TASK-005` 完成，错误分类、安全提示与连接标识由 `TASK-016` 完成。

## 实时同步与资源操作

| 编号 | 任务 | 优先级 | 状态 | 依赖 | 验收重点 |
|---|---|---:|---|---|---|
| `TASK-021` | [Docker / Podman 统一 runtime driver](unified-runtime-driver.md) | P0 | `done` | TASK-004 | 独立 Docker/Podman Engine；统一 Container/Image/Volume/Network service；Podman native actions、Logs、Events、Exec、镜像 Inspect（runtime.ImageDetail 对齐）；同字段 AND 筛选；统一 errors/capabilities；TUI/state 仅依赖 `runtime.Engine`；CGO/non-CGO 矩阵通过；连接池通过 `EngineFactory` 注入，typed-nil 接口经反射清理 |
| `TASK-008` | Docker / Podman Events 接入主循环 | P1 | `done` | TASK-003、TASK-021 | 订阅绑定活动连接；切换时取消；1-30 秒退避重连；100ms 事件合并；按资源局部刷新；15 秒轮询降级 |
| `TASK-009` | Volume / Network 创建与清理 | P1 | `done` | TASK-021 | 已完成 create、prune、确认交互、逐资源部分失败反馈、Docker/Podman contract tests 和双构建矩阵 |
| `TASK-010` | 批量操作扩展与部分成功反馈 | P2 | `todo` | 审计模型、TASK-017 | 每个目标独立终态、汇总提示和可追溯审计 |
| `TASK-011` | Compose / 容器 / 镜像联动刷新 | P2 | `todo` | TASK-008 | 事件只使相关资源失效，不直接修改复杂 UI 状态 |
| `TASK-017` | 高频容器操作 | P0 | `done` | TASK-004 | 已实现状态约束的 `pause` / `unpause`、批量跳过汇总、`rename` 输入校验、独立 `top` 页面和结构化 `port` 展示，并通过 Docker / Podman 兼容 API 契约测试 |
| `TASK-018` | 镜像标签与传输工作流 | P1 | `done` | TASK-021 | runtime-neutral transfer service；`tag`、`push`、`save`、`load`；字节/daemon 进度、context 取消、错误展示和审计终态 |
| `TASK-024` | [Podman 镜像详情 + 连接池工厂化重构](unified-runtime-driver.md) | P1 | `done` | TASK-021 | 补齐 TASK-021 Phase 2 中 Podman `ImageService.Inspect` 的 TODO；移除 `global engineFactory` 与 `SetEngineFactory`；`runtimeinit.NewEngineFactory()` 与 `runtime.EngineFactory` 注入到 `runtimeapi.NewPool`；`sanitizeEngine` / `engineIsUsable` 反射防御 typed-nil 接口；APIVersion 回退覆盖所有错误而非仅 404；`RefreshAll` 替换 `Probe`/`pingAll`；`refreshOne` 对 transient 引擎真实 ping。本条目是 TASK-021 收尾增量 |
| `TASK-019` | 高级容器操作 | P2 | `todo` | TASK-017、TASK-021 | 评估并分批实现 `update`、`diff`、`export`、`commit`、`wait`、`cp` |
| `TASK-022` | [Podman REST 适配收紧与 docker/service 统一入口](podman-rest-migration.md) | P1 | `todo` | TASK-021 | `runtime/podman.Client` 改造为方法式 + `dto.*` 签名 + 驱动/REST 双形态；`gpgme` 仅 CGO；`dto/` 具名类型零 `go.podman.io` 依赖；全仓匿名 struct 清零；`docker/service` 建立 Docker/Podman 统一入口（先 4 个核心 service） |
| `TASK-023` | 清理 `internal/data/docker/podman_*.go` 与旧 `podmanContainerService` 等兼容实现 | P2 | `todo` | TASK-022 | TASK-022 完成后统一移除 `docker/podman_*.go` 共 24 个生产文件 + 8 个测试文件；`docker/Client.podmanREST` 字段清理；engine_factory 切到统一 service |

`build` 需要独立输入和进度交互设计，暂不并入 `TASK-018`，待该任务完成后再建立实施项。

## TASK-021 状态快照（2026-07-22）

### 已完成的 Phase

| Phase | 内容 | 状态 |
|---|---|---|
| Phase 0 | 依赖探测 + CGO 约束验证 | ✅ 完成 |
| Phase 1 | 域包 + 连接驱动（`runtime.Engine` 工厂） | ✅ 完成 |
| Phase 2 | 只读资源迁移 | ✅ 完成 |
| Phase 3 | 资源动作迁移 | ✅ 完成 |
| Phase 4 | 流式能力迁移 | ✅ 完成 |
| Phase 5 | 清理 | ✅ 完成 |

### Phase 2 详情：只读资源迁移状态

| 资源 | List | Inspect | 结构化类型 | Podman 双 transport |
|------|:----:|:-------:|:--------:|:------------------:|
| Container | ✅ Engine 接口 | ✅ Engine 接口 | ✅ `ContainerDetail` | ✅ CGO+REST |
| Volume | ✅ Engine 接口 | ✅ Engine 接口 | ✅ `VolumeDetail` | ✅ CGO+REST |
| Network | ✅ Engine 接口 | ✅ Engine 接口 | ✅ `NetworkDetail` | ✅ CGO+REST |
| Image | ✅ Engine 接口 | ✅ Engine 接口（Podman 通过 `/v4.0.0/libpod/images/{id}/json` + `/history` 双接口实现） | ✅ `ImageSummary`/`ImageDetail` | ✅ CGO+REST |

### TASK-021 完成标准检查

| # | 完成标准 | 状态 |
|---|---------|------|
| 1 | TUI & state 不再 import Docker/Podman SDK | ✅ 达标 |
| 2 | 不存在 `Raw()` 或上层 SDK 类型 | ✅ 达标 |
| 3 | ConnectionPool & 业务命令仅依赖 `runtime.Engine` | ✅ 达标 |
| 4 | Docker & Podman 通过独立 adapter 实现相同契约 | ✅ 达标 |
| 5 | 差异仅存在于 adapter mapper、capabilities、unified errors | ✅ 达标 |
| 6 | TASK-017 的 5 个操作在两种 runtime 下通过 | ✅ 达标 |
| 7 | TASK-009 的 create/prune 统一 options/results 就绪 | ✅ 达标 |
| 8 | CGO/non-CGO 策略清晰，build matrix 通过 | ✅ 达标 |
| 9 | 镜像详情双 runtime 结构化对齐（含 layer history） | ✅ 达标（TASK-024 收尾） |
| 10 | ConnectionPool 通过注入工厂取代全局 `engineFactory` | ✅ 达标 |

### 已完成工作

| 工作项 | 状态 | 说明 |
|--------|------|------|
| Logs 加入 Engine 接口 | ✅ | `ContainerService.Logs()` 提供 runtime-neutral 的流式抽象 |
| TUI 持有 `runtime.Engine` 而非 `*docker.Client` | ✅ | 解耦 TUI 与具体适配器 |
| Volume/Network Inspect/Remove 加入 Service 接口 | ✅ | 完善 `VolumeService`/`NetworkService` 的方法集 |
| Container Attach | ✅ | 独立 container attach 能力 |
| Podman Image.Inspect | ✅ | 通过 `/v4.0.0/libpod/images/{id}/json` + `/v4.0.0/libpod/images/{id}/history` 双接口实现；映射到 `runtime.ImageDetail`；History 失败可降级（写入 `detail.HistoryError`） |
| ConnectionPool 工厂化 | ✅ | 删除全局 `engineFactory` 变量；`runtime.EngineFactory` 由 `runtimeinit.NewEngineFactory()` 注入；typed-nil 接口由反射清理 |
| RefreshAll 整合 | ✅ | 删除分散的 `Probe`/`pingAll`，统一在 `RefreshAll` 中实现；transient 引擎也执行真实 ping |
| APIVersion 健壮性 | ✅ | Podman 5.x 任意错误都触发 `/v4.0.0/libpod/version` 回退 |

## 操作历史与体验

| 编号 | 任务 | 优先级 | 状态 | 依赖 | 验收重点 |
|---|---|---:|---|---|---|
| `TASK-012` | 审计操作历史面板 | P2 | `done` | TASK-001、审计模型 | 已实现 PanelAudit 面板、AuditState 状态域、文本/级别筛选、记录详情视图、键盘导航（Enter/Esc/E），审计记录从 Service.RecentOperations() 实时同步 |
| `TASK-013` | 鼠标与小终端降级布局 | P3 | `todo` | 页面固定轨道 | 鼠标不破坏键盘路径，小终端无重叠且核心操作可达 |
| `TASK-014` | 发布、命名和跨平台分发 | P3 | `todo` | 新配置稳定 | 统一命名、构建产物、版本信息、Linux / macOS 发布说明 |
| `TASK-020` | Docker / Podman 专有能力评估 | P3 | `todo` | TASK-004 | Podman pod/secret/kube 与 Docker buildx/context 等能力分开决策 |

## 依赖与顺序

```text
已完成连接基础 (TASK-001..007)
  |-> TASK-015 状态域拆分
  |-> TASK-016 TLS 错误体验
  |-> TASK-017 高频容器操作 -> TASK-010 / TASK-019
  |-> TASK-021 统一 runtime driver [done]
  |     |-> TASK-009 Volume / Network         [done]
  |     |-> TASK-018 镜像工作流               [done]
  |     `-> TASK-008 Events [done] -> TASK-011 联动刷新 [todo]
  `-> TASK-012 审计历史面板 [done]

独立后续: TASK-013 / TASK-014 / TASK-020
```

## 暂不实施

以下能力保留在能力分析中，尚未达到建立实施任务的条件：

| 能力 | 原因 |
|---|---|
| Compose up / build / pull | 需要完整服务编排交互和进度模型 |
| Build / buildx | 参数、上下文、多平台和进度交互范围较大 |
| Podman pod / secret / kube | 属于运行时专有能力，等待 `TASK-020` 决策 |
| Checkpoint / plugin | 用户场景较窄，优先级低于通用运维能力 |

## 完成定义

每个任务完成前必须满足：

1. 代码行为与讨论文档中的已确认决策一致。
2. Docker 与 Podman 共用能力不得通过散落的 runtime 字符串分支实现；差异应收敛到 runtime/data 层。
3. 至少覆盖主路径、失败路径；双运行时能力应包含 Docker / Podman 测试或明确的环境验证记录。
4. 配置参数必须在读取后校验，运行期不得重复修正非法配置。
5. 状态结构应通过方法维护自身不变量，避免将相关字段作为散装参数传递。
6. `git diff --check` 和 `just check` 通过。
7. README、架构、需求状态及本任务台账同步。
8. 独立提交，提交信息包含任务或需求语义。
