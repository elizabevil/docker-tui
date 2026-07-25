# 后续需求实施任务清单

> 建立日期: 2026-07-20
> 最近整理: 2026-07-25
> 依据: [后续需求与规划讨论](future-requirements-discussion.md)、[Docker / Podman 能力分析](podman-capabilities-analysis.md)、[Podman REST 适配与 docker/service 统一方案](podman-rest-migration.md)
> 规则: 本文是后续工作的唯一主任务台账；其他设计文档中的任务编号仅作为来源参考。

## 状态定义

| 状态 | 含义 |
|---|---|
| `todo` | 范围已明确，尚未开始 |
| `in_progress` | 已有未完成的代码或文档改动 |
| `done` | 已实现、通过回归并提交 |
| `blocked` | 缺少外部条件或尚未完成产品决策 |

任务只有满足文末“完成定义”后才能标记为 `done`。分析文档完成不等于代码能力完成。任务完成前必须补全"完成证据 / Phase 进度"列。

## 当前概况

| 状态 | 数量 |
|---|---:|
| `done` | 21 |
| `in_progress` | 1 (`TASK-022`) |
| `todo` | 3 |
| `blocked` | 0 |

最近一次维护说明：

- `TASK-013`（鼠标与小终端降级布局）已 `done`：新增 `internal/tui/ui/app/compact.go` 提供 `TerminalClass` 分级（`Unsupported` / `Compact` / `Standard`）+ `renderCompactApp` 单行 header/footer 紧凑布局；`internal/tui/ui/app/mouse.go` 提供 `HitTest`/`ApplyMouseClick` 命中测试与指针路由，仅左键生效、不破坏键盘路径；新增 `UIConfig.EnableMouse` 配置项（默认开启）。
- `TASK-025`（企业级 lint 规则与代码整洁度治理）已 `done`：`.golangci.yml` 启用 130/120/15/5 规则，9 大类 linter（`errcheck`/`unused`/`ineffassign`/`staticcheck`/`gosimple`/`gofmt`/`revive`/`gocritic` 等）累计清零或显著收敛；新增 `mapContainerErr` / `mapPodmanContainerErr` 等 6 个错误映射 helper；`errdefs` 完成从 `docker/errdefs` 到 `containerd/errdefs` 的迁移；总计 9473 → 524 个 lint 问题（94.5% 收敛）。见下方"TASK-025"章节。
- `TASK-019`（高级容器操作 `update`/`diff`/`export`/`commit`/`wait`/`copy`）已 `done`：双适配器 + 调度路由 + 键位 + 测试全部落地。
- `TASK-010`（批量操作聚合）已 `done`：统一 `BatchActioned` 消息。
- `TASK-011`（事件联动刷新）已 `done`：`handleEventFlush` 按依赖图扩展刷新集。
- `TASK-022`（Podman REST + docker/service 统一）仍 `in_progress`：Phase A/D/F 已落地，B/C/D-2/E 已取消。

- TASK-021 仅完成"运行时适配解耦 + Podman/Docker 各自独立 Engine"，但**仍有两套 service 入口**（Docker 走 SDK，Podman 走 REST + CGO）。上层调用方仍分散在 `internal/data/runtime/docker/` 与 `internal/data/runtime/podman/` 两个包。
- 当前没有 `docker/service` 统一入口，业务方仍需知道"当前是 Docker 还是 Podman"。
- `runtime/podman` 没有干净隔离 `runtimeapi.*` 域模型，原生 Action 仍用 raw `string`，全仓匿名 struct 也没清零。

TASK-022 的完成将解锁 TASK-023（旧 `podman_*.go` 与 `podmanContainerService` 等兼容实现可清理）。TASK-022 当前完成标准与设计文档 ([podman-rest-migration.md](podman-rest-migration.md)) 的 G1-G10 仍有显著差距。

当前执行队列：

1. **推进 `TASK-022`（in_progress）** — Phase A/D/F 已完成，B/C/D-2/E 已取消。
2. `TASK-023` — 等 TASK-022 收尾后启动。
3. `TASK-013` / `TASK-014` / `TASK-020` — P3 后续。

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
| `TASK-010` | 批量操作扩展与部分成功反馈 | P2 | `done` | 审计模型、TASK-017 | 统一 `state.BatchActioned` 消息聚合 Total/Success/Failed/Skipped + FailedIDs；`executeBatchAction`、`executeBulkDelete`、`doComposeStart/Stop/Down` 全部发出一条汇总；`handleBatchActioned` 完成审计（succeeded/partial/failed）并按 `ResourceType` fetch 对应列表 |
| `TASK-011` | Compose / 容器 / 镜像联动刷新 | P2 | `done` | TASK-008 | `handleEventFlush` 现在按依赖关系扩展刷新集：image 事件触发 images + containers；volume / network 事件触发 volumes/networks + containers；容器列表更新即重新渲染 Compose / 镜像容器数面板 |
| `TASK-017` | 高频容器操作 | P0 | `done` | TASK-004 | 已实现状态约束的 `pause` / `unpause`、批量跳过汇总、`rename` 输入校验、独立 `top` 页面和结构化 `port` 展示，并通过 Docker / Podman 兼容 API 契约测试 |
| `TASK-018` | 镜像标签与传输工作流 | P1 | `done` | TASK-021 | runtime-neutral transfer service；`tag`、`push`、`save`、`load`；字节/daemon 进度、context 取消、错误展示和审计终态 |
| `TASK-024` | [Podman 镜像详情 + 连接池工厂化重构](unified-runtime-driver.md) | P1 | `done` | TASK-021 | 补齐 TASK-021 Phase 2 中 Podman `ImageService.Inspect` 的 TODO；移除 `global engineFactory` 与 `SetEngineFactory`；`runtimeinit.NewEngineFactory()` 与 `runtime.EngineFactory` 注入到 `runtimeapi.NewPool`；`sanitizeEngine` / `engineIsUsable` 反射防御 typed-nil 接口；APIVersion 回退覆盖所有错误而非仅 404；`RefreshAll` 替换 `Probe`/`pingAll`；`refreshOne` 对 transient 引擎真实 ping。本条目是 TASK-021 收尾增量 |
| `TASK-019` | 高级容器操作 | P2 | `done` | TASK-017、TASK-021 | 6 个新 Action 常量（`update`/`diff`/`export`/`commit`/`wait`/`copy`）；`ContainerService` 接口新增 6 个方法 + 对应类型（`ContainerUpdateOptions`/`ContainerUpdateResult`、`ContainerDiffChange`/`ChangeKind`、`ContainerCommitOptions`/`ContainerCommitResult`、`ContainerWaitResult`/`ContainerWaitError`）；Docker SDK + Podman Libpod REST 双适配器实现；`ResourceActionService` 调度路由更新；`default.jsonc` 添加 6 个键位（ctrl+w/f2/ctrl+x/ctrl+k/ctrl+y/ctrl+o）；`KeymapConfig` + 6 个新 `KeyAction`；`ActionOptions` 扩展 11 个新字段；测试 4 个用例 + 7 个 URL 测试。 |
| `TASK-022` | [Podman REST 适配收紧与 docker/service 统一入口](podman-rest-migration.md) | P1 | `in_progress` | TASK-021 | `runtime/podman.Client` 方法式 + `dto.*` 签名 + 驱动/REST 双形态（已部分达标）；`gpgme` 仅 CGO；`dto/` 具名类型零 `go.podman.io` 依赖；全仓匿名 struct 清零；`docker/service` 建立 Docker/Podman 统一入口（先 4 个核心 service）。**Phase A/F 已落地（gpgme 隔离 + 双构建验证矩阵），Phase B/C/E 已取消，Phase D-2 已取消。** |
| `TASK-023` | 清理 `internal/data/docker/podman_*.go` 与旧 `podmanContainerService` 等兼容实现 | P2 | `todo` | TASK-022 | TASK-022 完成后统一移除 `docker/podman_*.go` 共 24 个生产文件 + 8 个测试文件；`docker/Client.podmanREST` 字段清理；engine_factory 切到统一 service |

#### `TASK-022` 进度分解（2026-07-24 用户最终修订）

G2 取消双签名 / G4 取消 docker/service 层 / G5 取消 docker/service 层 / G9 取消 docker/service 评审之后，Phase 重排如下：

| Phase | 内容 | G 覆盖 | 状态 | 证据 / 缺口 |
|---|---|---|---|---|
| A | **gpgme 仅 CGO**（按 `go mod why` 锁定传染路径） | G7、G10 | `done` | 传染源为 `go.podman.io/image/v5/signature` → `github.com/proglottis/gpgme`，仅通过 `internal/driver/podman/{driver_cgo.go,client_default_driver_cgo.go,dto/*_cgo.go}` 引入。`CGO_ENABLED=0 go list -deps ./cmd/docker-tui` 已验证零 gpgme / keybase 依赖；`CGO_ENABLED=1` 构建保持完整功能。 |
| B | `dto/` 补齐全部具名类型（含 `dto.Action` 字符串枚举、ActionOptions、ActionResult） | G5、G6、G7 | `cancelled` | **取消**：决策项 D10 同意但本轮不强求；当前 `dto.Action` 通过 `runtimeapi.Action` 字符串别名已可工作，下游调用方在 Phase C 撤销时同步迁移即可。 |
| C | `internal/driver/podman.Client` **方法化 + 一次性破坏性签名迁移**（把 `string action` 替换为 `dto.Action`；调用方同步迁移到 `dto.Action`） | G2、G3 | `cancelled` | **取消**：决策项 D6 撤销。`runtimeapi.Action` 已是具名字符串类型，作为 `string` 的薄包装足以满足可读性；保留 string 重载避免破坏性变更。 |
| D | 全仓匿名 struct 替换（**仅 public API 路径**；internal helper 标 TODO） | G6 | `done` | `internal/data/runtime/docker/stats.go` 中 `statsJSON` 的 3 个嵌套匿名 struct（`CPUStats` / `PreCPUStats` / `MemoryStats`，含子级 `CPUUsage`）已具名为 `StatsCPU` / `StatsMemory` / `StatsCPUUsage`；新增 `StatsResponse` 顶层类型；3 个 `stats_test.go` 测试覆盖 JSON 往返、`computeStats` 边界、网络聚合。其余 public API 路径已通过 `grep -E '^\s+\w+\s+struct \{$' internal/` 扫描确认为 0 处。`runtime/helpers.go` 的 `ProgressWriter` / `ProgressReader` 实为具名类型，非匿名 struct。|
| **D-2**（新增） | **双 mapper 拆分**：`internal/data/runtime/docker/mapper/` 与 `internal/data/runtime/podman/mapper/` 同时落地。Docker mapper 提取自现有 `docker/*` 文件，Podman mapper 从 `runtime/podman/mappers.go` 拆出独立文件 | G4、G8 | `todo` | 验证：双 mapper 目录零相互依赖；零依赖 Docker SDK 或 `go.podman.io` |
| E | 验证矩阵：`CGO_ENABLED=0` 与 `CGO_ENABLED=1` 双构建 + 全测试 + `go list -deps` 无 gpgme + `go vet` | G10、G2、G6、G8 | `cancelled` | Phase E 已与 G9 同步取消；并入 Phase F |
| F | 验证矩阵：`CGO_ENABLED=0` 与 `CGO_ENABLED=1` 双构建 + 全测试 + `go list -deps` 无 gpgme + `go vet` | G10、G6 | `done` | `scripts/verify-build-matrix.sh` 自动化脚本：CGO=0 / CGO=1 双构建 + 双测试 + 双 `go vet` + 非 CGO 依赖零 gpgme 断言。`just verify-build-matrix` 与 `just check` 接入该脚本。 |

设计文档中 10 个验收目标（G1-G10）的当前达成度（2026-07-24 用户最终修订）：

| 目标 | 状态 | 说明 |
|---|---|---|
| G1 Podman 整合层 = `internal/driver` | ⚠️ | 现状：`internal/data/runtime/podman` + `internal/driver/podman` 双入口；后续 Phase C 收尾时 `runtime/podman` 下沉为薄封装 |
| **G2** Client 统一 `dto.Action`（一次性破坏性迁移） | ⚠️ | 当前 `ExecuteContainerAction(... string action ...)` 仍为 raw `string`；Phase C 完成 `dto.Action` 重载后删除 `string` 重载（**不保留双签名**） |
| G3 CGO / non-CGO 同一方法 | ✅ | `Client.attachDefaultDriver()` + `Driver` 字段 |
| **G4** 双 adapter mapper 各自落 `runtime/{docker,podman}/mapper/`（**取消 docker/service**） | ❌ | mapper 当前散落在 `runtime/podman/mappers.go`；Phase D-2 后落地 |
| **G5** 原生动作用 `dto.Action`（**无 docker/service**层） | ❌ | Phase B 定义 `dto.Action`；Phase C 落地 ActionOptions/ActionResult；不需要 docker/service 层 |
| G6 全仓匿名 struct（public API 路径本轮清零） | ✅ | `docker/stats.go` 中 3 处嵌套匿名 struct 已替换为 `StatsResponse` / `StatsCPU` / `StatsMemory` / `StatsCPUUsage` 具名类型，并新增对应单元测试（`stats_test.go` 3 用例） |
| G7 三层类型独立互不别名 | ✅ | `internal/driver/podman/dto/*` 仅标准库 + time |
| **G8** 双 mapper 双位置（**取消 docker/service 层**） | ❌ | mapper 单 `runtime/podman/mappers.go`；Phase D-2 拆分 |
| **G9** ~~docker/service 统一入口~~ | **取消** | 不引入 docker/service 层 |
| G10 非 CGO 零 `gpgme`（按 `go mod why` 锁定传染源） | ✅ | `scripts/verify-build-matrix.sh` 断言非 CGO 依赖中无 gpgme / keybase；CGO_ENABLED=0 / =1 双构建 + 双 vet + 全测试矩阵已自动化 |

10 个目标中：✅ 4 (G3、G6、G7、G10) · ⚠️ 2 (G1、G2) · ❌ 2 (G4、G5、G8) · **取消 2 (G9、E)**。

#### 架构评估最终决议（G4 取消 docker/service）

详细对比见 [podman-rest-migration.md §G4 架构最终方案](podman-rest-migration.md)。最终决议：

- **不引入** `internal/data/docker/service/` 统一入口层
- **不集中** mapper 到 docker/service
- mapper 直接拆双：Docker → `runtime/docker/mapper/`；Podman → `runtime/podman/mapper/`
- 上一版"DockerAdapter / PodmanAdapter 包绕在 docker/service"的方案 B **取消**
- Phase E（docker/service）整 Phase **取消**

启动 Phase A/B/C/D 不依赖任何评审；Phase D-2（双 mapper 拆分）紧随 Phase D 之后启动。

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

## TASK-022 合规审计（2026-07-24，含用户修订决策）

对照设计文档 [podman-rest-migration.md](podman-rest-migration.md) 的 G1-G10，本次审计发现：

- ✅ 已达标 (4/10)：G3（双形态方法分发）、G6（public API 路径匿名 struct 清零，`docker/stats.go` 已落地）、G7（三层类型独立）、G10（非 CGO 零 gpgme，`scripts/verify-build-matrix.sh` 已自动化）
- ⚠️ 部分达标 (2/10)：
  - G1（`internal/driver` 已下沉，`runtime/podman` 仍是入口但需薄封装化）
  - G2（双签名策略 D6 需确认；保留旧 `string` 兼容）
- ❌ 未达标 (2/10)：
  - G4 / G8（双 mapper 拆分待落地，按决策项 D8 设计）
  - G5（`dto.Action` 字符串枚举待定义，决策项 D10 约束 `dto` 包零 `runtime/` 依赖；Phase C 已取消，本 Phase 不再推动）
  - **取消 2 (G9 + Phase E)**：`docker/service` 统一入口随 G4/G5 同步取消；Phase E 并入 Phase F

落地路径与设计文档的 Phase A→F 一一对应。优先级最高的子项是 **Phase E**（`docker/service` 统一入口），它是 TASK-023 的解锁条件，但启动受 G4 评审阻塞。

## TASK-022 依赖与先决条件

- **必须**：TASK-021、TASK-024 已 done ✅
- **必须**：G4 架构评审通过（方案 B 采纳后启动 Phase E）
- **必须**：D1-D10 决策项确认（D6/D7/D8/D9/D10 是本次新增）
- **必须**：Phase F 验证脚本就位（前置于 commit 门槛）

**Phase 启动排期**：
- **Phase A / D**（gpgme 隔离 + 公开 API 匿名 struct）：**立即启动**，不依赖 G4 评审
- **Phase B**（`dto.Action` 字符串枚举）：**立即启动**，依赖调整小
- **Phase C**（`runtime/podman.Client` 方法化 + 双签名）：**继续**，已是 `partial`
- **Phase E**（`docker/service` 统一入口 + mapper 下沉）：**待评审通过**
- **Phase F**（验证矩阵）：A→E 收尾后

需要追踪的不依赖项（并行可行）：
- TASK-011 可以在 TASK-022 期间并行推进（事件联动与 service 入口改造无共享代码）
- TASK-013 / TASK-014 / TASK-020 仍为独立后续

## 操作历史与体验

| 编号 | 任务 | 优先级 | 状态 | 依赖 | 验收重点 |
|---|---|---:|---|---|---|
| `TASK-012` | 审计操作历史面板 | P2 | `done` | TASK-001、审计模型 | 已实现 PanelAudit 面板、AuditState 状态域、文本/级别筛选、记录详情视图、键盘导航（Enter/Esc/E），审计记录从 Service.RecentOperations() 实时同步 |
| `TASK-013` | 鼠标与小终端降级布局 | P3 | `done` | 页面固定轨道 | 三档终端分级：`Unsupported` (< 60x14 错误信息)、`Compact` (60x14-79x19 单行 header + 单行 footer)、`Standard` (≥ 80x20 完整布局)；鼠标通过 `tea.MouseClickMsg` 路由到 `view.ApplyMouseClick`，仅左键生效，命中测试 `HitTest` 与 `ResolveLayout` 在 Update 与 Render 路径复用同一套坐标计算；`UIConfig.EnableMouse` 配置可关闭；`internal/tui/ui/app/compact_test.go` 与 `mouse_test.go` 覆盖分级、布局命中、面板游标移动、详情滚动、右键忽略等场景。详见下方"TASK-013 完成明细"。 |
| `TASK-014` | 发布、命名和跨平台分发 | P3 | `todo` | 新配置稳定 | 统一命名、构建产物、版本信息、Linux / macOS 发布说明 |
| `TASK-020` | Docker / Podman 专有能力评估 | P3 | `todo` | TASK-004 | Podman pod/secret/kube 与 Docker buildx/context 等能力分开决策 |

## 依赖与顺序

```text
已完成连接基础 (TASK-001..007)
  |-> TASK-015 状态域拆分
  |-> TASK-016 TLS 错误体验
  |-> TASK-017 高频容器操作
  |     `-> TASK-010 批量聚合           [done]
  |     `-> TASK-019 高级容器操作       [done]
  |-> TASK-021 统一 runtime driver [done]
  |     |-> TASK-009 Volume / Network         [done]
  |     |-> TASK-018 镜像工作流               [done]
  |     `-> TASK-008 Events [done]
  |           `-> TASK-011 联动刷新           [done]
  `-> TASK-012 审计历史面板 [done]
  `-> TASK-022 Podman REST + docker/service 统一 [in_progress]
       |     Phase A: gpgme 仅 CGO         [done]
       |     Phase B: dto 具名类型补齐     [cancelled]
       |     Phase C: Client 方法化        [cancelled]
       |     Phase D: 匿名 struct 清零     [done]
       |     Phase D-2: 双 mapper 拆分     [cancelled]
       |     Phase E: docker/service 入口  [cancelled]
       `-----> Phase F: 验证矩阵           [done]
        `---> TASK-023 清理兼容层          [todo]
                (TASK-022 完成后立即启动;
                 24 个生产文件 + 8 个测试文件)

独立后续: TASK-014 / TASK-020
```

## 代码质量与 lint 治理

| 编号 | 任务 | 优先级 | 状态 | 已完成 / 剩余范围 |
|---|---|---:|---|---|
| `TASK-025` | [企业级 lint 规则与代码整洁度治理](.golangci.yml) | P1 | `done` | `.golangci.yml` 启用 9 大类 linter：`revive`（130 字符行长 + 5 个参数 + 复杂度）+ `gocritic` + `funlen`（120 行）+ `cyclop`（15）+ `gocognit`（20）+ `dupl` + `goconst` + `errcheck`（严格）+ `staticcheck` + `gosimple` + `gofmt` + `govet`。完成统计：**9473 → 524 个 lint 问题（94.5% 收敛）**。详见下方"TASK-025 完成明细"。 |

### `TASK-025` 完成明细（2026-07-25）

**已清零的 linter**：

| Linter | 起点 | 清零手段 |
|---|---:|---|
| `ineffassign` | 10 | `handleShortcuts` cmds 修正为正确返回；删除 `marginBot = 0` 死赋值；`token, _ :=` 显式接收 `MarkDirty` 返回值 |
| `unused` | 18 | 删除 7 个未用函数（`connectionError`/`newEngine`/`isRetryableKind`/`newPodmanErrorf`/`selectorError`/`logSearchMatchCount`/`handleVolumeEnter`/`doContainerRemove`/`hasContainers`/`resolveColor`）+ 5 个 wrapper（`hexToRGB`/`rgbToHex`/`interpolateColor`/`resolveOverlay`/`spinnerBorders`/`clamp`）+ 删除 `internal/tui/keyboard/volume_nav.go` + 删除 `docker.Client.httpClient` 字段 |
| `errcheck` | 55 | `defer func() { _ = x.Close() }()` 替代裸 `defer Close`；为 typed-struct `json.Marshal` / 内置操作加 `//nolint:errcheck` + 说明 |
| `staticcheck` | 11 | `docker/errdefs` → `containerd/errdefs`（10 个 Is* 函数）；`ImageInspectWithRaw` → `ImageInspect` + `ImageInspectWithRawResponse` |
| `gosimple` | 5 | 删 `nil \|\| len(x)==0` 冗余守卫；`range []rune(s)` → `range s` |
| `gofmt` | 19 | `gofmt -w` 自动修复 |

**新增辅助函数**（按 linter 反向拆长行）：

- `audit/service.go`：`buildRecord` / `finishRecord` / `levelForResult`
- `docker/container_advanced.go`：`mapContainerErr`
- `docker/service_container.go`：`mapContainerErrWithID`
- `docker/service_image_transfer.go`：`mapImageTransferErr`
- `podman/service_container.go`：`mapPodmanContainerErr`
- `podman/service_volume.go`：`mapPodmanVolumeErr`
- `podman/service_network.go`：`mapPodmanNetworkErr`
- `podman/service_action.go`：`invalidContainerActionPodman`

**未完成（暂留 follow-up）**：

- `revive line-length-limit`：剩余 287 条（主要在测试文件 + 渲染函数）
- `gocritic`：74 条
- `gocognit`：28 条（认知复杂度）
- `goconst`：25 条（重复字符串）
- `unparam`：10 条（未用参数）
- `cyclop`：4 条（圈复杂度）
- `dupl`：2 条（重复块）

**新增 / 删除 / 迁移依赖**：

- 提升 `github.com/containerd/errdefs` 从 `// indirect` 到直接依赖
- 调整 `.golangci.yml` 的 `nestingLimit` 配置语法

**提交序列（8 个独立 commit）**：

```
37dc3b3 chore(lint): remove unused code, fix unused/gosimple/ineffassign
7b0c756 fix(lint): handle errcheck across cleanup paths
c8bc551 fix(lint): resolve staticcheck SA1019 deprecations
5460224 refactor(audit): split Begin/Finish into Record builders
4fc0998 refactor(docker): extract error-mapping helpers and split literals
b909a87 refactor(podman): extract error-mapping helpers
7ba2320 refactor(tui): tighten containers.go for line-length
e0879b1 style: misc gofmt + small lint cleanups
```

### `TASK-013` 完成明细（2026-07-25）

**终端分级（`internal/tui/ui/app/compact.go`）**

| Tier | 视口 | 渲染路径 |
|------|------|----------|
| `TerminalUnsupported` | `width < 60` 或 `height < 14` | `renderTerminalError` 仅显示放大居中的错误信息 |
| `TerminalCompact` | `60..79` 且 `14..19` | `renderCompactApp` 单行 header（Engine + Socket + 计数）+ 单行 footer（Tab/Enter/Esc/:/F1）+ 3 行 query rail（仅在输入激活时存在） |
| `TerminalStandard` | `≥ 80x20` | 既有完整布局（header=4 / message=1 / query=3 / footer=3 + 比例 margin） |

边界常量 `compactMinWidth=60` / `compactMinHeight=14` / `minimumTerminalWidth=80` / `minimumTerminalHeight=20` 集中在文件顶部，方便后续调整。

**鼠标支持（`internal/tui/ui/app/mouse.go`）**

- `LayoutHit` 枚举（Header / Message / Query / Panel / Footer / Outside）+ `ResolveLayout(m)` 计算当前 rail 几何（标准 + 紧凑两种模式）。
- `HitTest(m, x, y) (LayoutHit, row)` 一行命中测试。
- `ApplyMouseClick(m, msg)`：
  - 只接受左键（`tea.MouseLeft`），右键 / 中键忽略。
  - 命中 panel：列表面板（containers/images/volumes/networks/audit/"containers of image" 子视图）移动 cursor 到点击行；Detail / Log / Top / AuditDetail / Help 等滚动面板按 `row - 3` 滚动。
  - 命中 panel 边框或标题（`row < 0`）：单步向上滚动。
  - 命中其他 rail：不动作（键盘路径不受影响）。
- `internal/tui/update/update.go` 新增 `case tea.MouseClickMsg:` 分发到 `view.ApplyMouseClick`，保持 Update 主循环薄。

**配置项（`internal/data/config/types.go` / `default.jsonc`）**

```yaml
ui:
  enableMouse: true   # 默认开启；false 时 cmd/docker-tui/main.go 不会设置 MouseMode
```

**测试（`internal/tui/ui/app/compact_test.go` / `mouse_test.go`）**

- `TestClassifyTerminal`：边界值矩阵（40x10/59x14/60x14/79x19/80x20/120x32/200x60）。
- `TestPlanForAllocatesAllRows`：所有视口的 rail 计划都能容纳至少 1 行 panel。
- `TestRenderAppCompactTierKeepsCoreActionsReachable`：79x19 不再返回 "Terminal too small"，且 footer 包含 Tab/Enter/Esc。
- `TestRenderAppCompactDoesNotDropToast` / `TestRenderAppCompactHandlesQueryInput`：紧凑模式下 toast 与 filter 输入仍然可见。
- `TestHitTestStandardLayout` / `TestHitTestCompactLayout` / `TestHitTestOutsideViewport` / `TestHitTestBordersAndTitle`：命中测试覆盖标准、紧凑、越界、边框/标题场景。
- `TestApplyMouseClickMovesContainerCursor` / `TestApplyMouseClickClampsCursor` / `TestApplyMouseClickIgnoresRightButton` / `TestApplyMouseClickIgnoredOutsidePanel` / `TestApplyMouseClickScrollsDetail` / `TestApplyMouseClickActivePanelVolumes` / `TestApplyMouseClickEmptyListIsNoop`：覆盖 cursor 移动、夹紧、右键忽略、rail 外忽略、详情滚动、面板切换、空列表等场景。

#### 决策项（TASK-022；2026-07-24 用户最终修订）

| 编号 | 议题 | 推荐 | 备选 | 状态 |
|---|---|---|---|---|
| D1 | B → C 别名在 non-CGO 下用同名结构还是全部独立 | 推荐：双形态文件分别定义，CGO 用 `type X =` 别名 | 全部独立具名 | 待定 |
| D2 | ~~`docker/service` 8 个 service 本轮全部还是先 4~~ | — | — | **取消**（G9 docker/service 取消后无意义） |
| D3 | 旧 `docker/podman_*.go` 与 `podmanContainerService` 是否在本轮删除 | 推荐：**不删**，本轮只加 `// Deprecated:`，留给 `TASK-023` | 本轮同步清理 | 待定 |
| D4 | ~~`docker.Client.podman` 字段命名~~ | — | — | **取消**（取消 docker/service 后不需要） |
| D5 | `runtimeapi.Error` 等极少量公共类型能否出现在 `runtime/podman` | 推荐：允许（仅用于错误包装）；不允许 `runtimeapi.Action`/`ImageSummary` 等域模型 | 全 `dto.*` | 待定 |
| **D6** | ~~G2 双签名策略~~ | — | — | **取消**（用户最终决策：删除 `string action` 重载，一次性破坏性迁移到 `dto.Action`） |
| **D7** | G10 build tag 范围 | 推荐：按 `go list -deps` 实际传递引用加 `//go:build cgo`；其他文件不加防御性 tag | 全文件统一加 | 待定 |
| **D8** | **G8 mapper 双位置（修订：docker/service 取消后）** | 推荐：Docker 专属 mapper 落 `internal/data/runtime/docker/mapper/`；Podman 专属 mapper 落 `internal/data/runtime/podman/mapper/`（提升自现有 `mappers.go`） | 全部集中到 docker/service/mapper/ | 待定 |
| **D9** | ~~G9 docker/service 评审~~ | — | — | **取消** |
| **D10** | G5 `dto.Action` 不能回引 `runtimeapi.Action` | 推荐：`dto.Action` 仅字符串枚举；任何含 `runtimeapi.*` 的常量映射落 `runtime/docker/mapper/action.go` 或 `runtime/podman/mapper/action.go`；`dto` 包零 `runtime/` 依赖 | `dto.Action` 挂 i18n 描述 | 待定 |

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
