# Docker / Podman 统一驱动方案

> 对应任务: `TASK-021`
> 状态: `in progress`（Phase 0 已完成）
> 建立日期: 2026-07-21

## 1. 目标

在数据层建立稳定的容器运行时领域接口，由 Docker 与 Podman adapter 分别完成 SDK 调用、参数转换、结果归一化和错误转换。TUI、状态层和业务命令只依赖统一模型，不判断运行时类型，也不接触任一 SDK 类型。

统一不等于抹平真实能力差异。共同能力使用一致语义；专有或不完全等价的能力通过 capability 声明和强类型 unsupported 错误暴露。

## 2. 当前事实

`go.mod` 已直接依赖：

- `github.com/docker/docker v28.5.2+incompatible`
- `go.podman.io/podman/v6 v6.0.1`

但当前实现不是双驱动：

| 领域 | 当前实现 | 问题 |
|---|---|---|
| 连接、Ping、版本 | Docker SDK，包括 Podman 连接 | 依赖 Podman Docker compatibility API |
| 容器生命周期、Top、Logs、Stats | Docker SDK | Podman 差异只能在上层暴露或被兼容层隐藏 |
| Volume / Network | Docker SDK | 尚未建立 create / prune 的统一结果语义 |
| Inspect / Events | Docker SDK 原始类型或 JSON | SDK 模型泄漏，难以稳定适配 |
| Exec / TTY Resize | TUI 通过 `Client.Raw()` 直接调用 Docker SDK | 数据层边界被绕过，Podman native driver 无法替换 |
| Image List（CGO） | Podman native bindings | 当前唯一真正的 Podman adapter |
| Image List（非 CGO） | 手写 Libpod HTTP | 与 CGO 实现重复，错误和 TLS 行为未统一 |

当前 `Client` 同时承担连接、driver 选择、Docker SDK facade 和少量 Podman 分派；`RuntimeType` 只是标签，不是可替换驱动边界。

## 3. 已确认的 SDK 覆盖

Podman v6 bindings 已提供本项目需要的主要原语：

- Containers: list / inspect / start / stop / restart / kill / pause / unpause / rename / remove / logs / stats / top / exec / resize / wait / prune
- Images: list / remove / pull / prune / push / import / export / history / inspect
- Volumes: create / inspect / list / remove / prune / exists / import / export
- Networks: create / inspect / list / remove / prune / connect / disconnect / exists
- System: info / version / events / disk usage / prune

因此统一 driver 可以使用两端正式 SDK，不需要把 Podman 永久建立在 Docker compatibility API 上。

## 4. 目标分层

```text
TUI / keyboard / update
          |
          v
internal/data/runtime         统一模型、接口、错误、capabilities
          |
     DriverFactory
       /       \
      v         v
runtime/docker runtime/podman  SDK adapter + mapper
      |         |
 Docker SDK   Podman bindings
```

建议最终将当前含义过宽的 `internal/data/docker` 迁移为：

```text
internal/data/runtime/
  engine.go          Engine 与分领域接口
  models.go          Container/Image/Volume/Network 等统一模型
  options.go         create/list/prune 等统一输入
  results.go         create/prune/batch 等统一输出
  errors.go          统一错误分类
  capabilities.go    能力声明
  pool.go            连接池，只持有 Engine
  docker/            Docker SDK adapter
  podman/            Podman bindings adapter
```

迁移期可保留 `internal/data/docker` facade，但它只能委托 `runtime.Engine`，不得继续新增 SDK 调用。

## 5. 接口边界

不建立一个包含几十个方法的巨型接口。按能力拆分，由 `Engine` 聚合：

```go
type Engine interface {
    Identity() Identity
    Capabilities() CapabilitySet
    Ping(ctx context.Context) error
    Close() error

    Containers() ContainerService
    Images() ImageService
    Volumes() VolumeService
    Networks() NetworkService
    Events() EventService
    Exec() ExecService
}
```

所有方法必须：

1. 接收调用方 `context.Context`，禁止 driver 保存永久业务 context。
2. 接收统一 options，禁止上层导入 Docker / Podman options。
3. 返回统一模型，禁止返回 SDK struct、原始 response 或任意 JSON。
4. 将 not found、conflict、invalid、unsupported、permission、connection 等错误转换为统一错误类型。
5. List 方法接收统一 `FilterSet`，adapter 将可支持的条件转换为运行时原生 filters，禁止 UI 为方便而先拉取全部资源再筛选。

`Raw()` 必须删除。Exec create / attach / resize / close 由 `ExecService` 和统一 `ExecSession` 封装。

## 6. 统一语义与差异处理

### 6.1 统一模型

- ID 保留完整值，UI 自己生成 short ID。
- 时间统一为 `time.Time`，大小统一为 bytes。
- 状态映射到强类型枚举，同时保留只读 `NativeState` 供诊断。
- Ports、IPAM、mount、stats、process table 使用结构化模型。
- Inspect 转为结构化 detail DTO；不得让 UI 解析两端原始 JSON。

### 6.2 筛选下推

- Container、Image、Volume、Network 和 Event 的 list options 都包含统一 `FilterSet`；筛选字段由各 service 定义常量和校验规则。
- adapter 优先使用 Docker SDK filters 或 Podman bindings/Libpod API query filters，由运行时完成筛选，以减少传输、映射与 UI 内存开销。
- 同一字段有不同名称或编码方式时，由 adapter 翻译，调用方不得传入 Docker/Podman 原生 key。
- 运行时不支持某个筛选条件时，driver 可在数据层后置筛选，但必须将 `CapabilityFiltering` 标记为 `Degraded` 并说明未下推的字段；不能静默改变筛选语义。
- 无法保证等价语义的条件返回 `ErrorUnsupported`，不得返回看似完整但实际错误的结果。
- mapper 只负责数据转换；filter 编码、能力判断和必要的后置筛选分别放在 adapter 的 query/filter 层。

### 6.3 Capability

能力至少区分：

- `Available`：两端语义一致，可直接使用。
- `Degraded`：可实现但部分字段或行为不同，UI 必须能展示说明。
- `Unsupported`：当前 driver 或 endpoint 不支持，UI 隐藏或禁用入口。

能力由连接成功后的版本与 driver 探测产生，不允许在 UI 使用 `runtime == "podman"` 判断。

### 6.4 已知差异

| 能力 | 需要归一化的差异 |
|---|---|
| Container Top | Docker 返回 Titles + Processes；Podman bindings 返回字符串行，需要 adapter 解析为动态列模型 |
| Logs | Docker 返回 multiplexed stream；Podman bindings 使用 stdout/stderr channel，需要统一为带 stream/source 的日志事件 |
| Stats | 字段结构、CPU 计算基数与网络接口表示不同，需要 driver 内生成统一 snapshot |
| Events | action/status 命名与 actor attributes 不同，需要统一 event kind，并保留 native action |
| Volume Prune | 删除报告和 reclaimed space 可用性不同；空间应为 optional 字段 |
| Network Create | IPAM、DNS、route、isolate/internal 等选项支持度不同，需要 capability/validation 协作 |
| Remove / Prune | Podman 常返回逐项 report，Docker 可能返回单 error；统一为逐目标 result |
| Image metadata | manifest、architecture、dangling 表示不同，需要统一 mapper |
| Exec attach | Hijacked connection 与 Podman stream API 生命周期不同，必须封装 session，不能暴露 `net.Conn` 或 Docker response |
| List filters | 两端支持的字段名、正则/精确匹配和多值语义不同；统一条件由 adapter 下推，不等价条件必须降级或拒绝 |

## 7. CGO 与发布约束

仓库 `just check` 使用 `CGO_ENABLED=0`，这是验收门，不得因 native Podman driver 被破坏。

现状中 Podman bindings image 实现受 `//go:build cgo` 限制，非 CGO 使用手写 HTTP。Phase 0 编译探针已确认：即使补齐 `github.com/gorilla/schema`，同时导入 containers/images/volumes/network/system bindings 的 `CGO_ENABLED=0` 构建仍会在 `github.com/proglottis/gpgme` 处失败（其全部 Go 文件受构建约束排除）。因此 native bindings 不能作为本项目默认发布路径。

Phase 0 决策：

1. Podman adapter 同时支持两种 transport：CGO 构建使用官方 bindings，非 CGO 构建使用集中式 remote REST transport。
2. 两种 transport 都先转换到 adapter 内部的 Podman wire DTO，再经过同一 mapper 转成统一领域模型；禁止复制 mapper 和业务语义。
3. Podman `pkg/domain/entities/types` 的依赖图同样包含 `gpgme`，非 CGO 路径不能直接导入。REST DTO 应按 Libpod OpenAPI 响应定义并通过 fixture/contract test 校验，而不是让领域层依赖 SDK struct。
4. 不再为单项功能添加分散的 `*_nocgo.go` 业务实现；build-tag 文件只负责选择 transport，HTTP、错误处理与 mapper 均集中复用。

TLS transport、超时和错误分类必须由 driver factory 统一注入，Podman adapter 不得另建绕过 TLS 配置的 `http.Client`。

TLS 证书采用懒加载：配置读取阶段只校验字段组合，首次连接时才读取 CA、客户端证书和私钥并构造 transport。默认必须校验证书；仅当连接显式配置 `insecureSkipVerify: true` 时允许跳过服务端证书校验，并将连接安全状态标为 insecure。`verify` 与 `insecureSkipVerify` 不得同时为 true。

筛选语义确定为：不同字段之间使用 AND；同一字段的多个值也使用 AND。UI 只提供简单条件，adapter 必须检查底层 driver 的多值语义；若原生 API 对同字段只能表达 OR，则不得直接下推为错误语义，应采用可下推的最小条件加数据层后置筛选，并将筛选能力标记为 degraded。

## 8. 已确认的实现决策

### 8.1 筛选

- 不同字段之间为 AND，同一字段的多个值同样为 AND。
- UI 只提供简单筛选，不暴露 Docker/Podman 原生 filter key、正则表达式方言或复杂布尔表达式。
- adapter 优先将条件下推给 driver；底层只支持同字段 OR 时，先下推不会改变结果集的条件，再在数据层完成剩余 AND 筛选。
- 后置筛选必须通过细粒度 capability 和诊断原因标记为 degraded；无法保证等价语义时返回 `ErrorUnsupported`。

### 8.2 API 版本

- Docker 使用 SDK API negotiation；Podman REST 在建立连接时查询 version endpoint，并选择服务端支持的 Libpod API 版本。
- 默认自动协商，连接配置允许可选的 API 版本覆盖，用于兼容和问题诊断。
- 版本低于某项能力要求时降低对应 capability，不允许靠请求失败后猜测能力。
- CGO bindings 与非 CGO REST transport 必须根据相同的服务端版本生成一致 capability。

### 8.3 超时与 context

- 健康检测默认间隔 3 秒，单次超时 2 秒，继续允许配置。
- 普通 list/inspect 查询默认超时 10 秒；生命周期操作默认超时 30 秒。
- Pull、Push、Logs、Events、Stats stream 和 Exec 不设置固定短超时，由调用方 context 控制取消。
- driver 不保存永久业务 context；应用退出和连接切换时必须取消该连接的全部流式任务。

### 8.4 流式接口

- Logs、Events、Stats、镜像传输进度和 Exec 使用强类型事件或 session，不向上层暴露 SDK stream、hijacked connection 或原始 JSON。
- 统一定义 stdout/stderr 来源、时间、进度、EOF、取消、断线和关闭语义。
- Events 可在临时连接错误时按退避策略重连；交互式 Exec 不自动重放，断线后明确结束 session。
- 创建 stream 的一方负责返回关闭句柄，消费方负责调用 Close；driver 必须保证 context 取消后 goroutine 退出。

### 8.5 批量操作与状态

- Remove、Prune 等批量操作返回逐资源 `OperationResult`，允许部分成功，不用首个错误覆盖其余结果。
- 容器状态、健康状态和 event action 转换为统一枚举，同时保留 `NativeState`/`NativeAction` 用于诊断。
- 未识别的新状态映射为 `Unknown`，不得导致 mapper 或 UI 失败。

### 8.6 Capability 粒度

- capability 按资源和操作声明，例如 `container.list.filter`、`image.list.filter`、`events.filter`、`exec.resize`，不使用一个全局布尔值代表全部能力。
- 每项能力包含 `Available`、`Degraded` 或 `Unsupported` 以及稳定 reason code；UI 根据状态禁用入口或展示提示。
- capability 来自 driver 类型、服务端版本与连接探测，不在 UI 中判断 Docker/Podman 类型。

### 8.7 TLS

- TLS 默认验证服务端证书，支持 CA、客户端证书、私钥和 ServerName。
- 显式支持 `insecureSkipVerify`；它默认 false，且不能与 `verify: true` 同时设置。启用时 UI 和连接状态必须显示 insecure，不得显示为 verified。
- TLS 材料懒加载：读取配置时只校验字段组合，首次连接时读取文件并创建 transport。配置变更或重连时创建新 transport，不复用旧证书状态。
- Unix socket 配置 TLS 视为无效配置并提前报错，禁止静默忽略。
- 日志和 UI 不输出证书正文、私钥内容或认证凭据。

### 8.8 重试与健康检测

- 仅对 connection、timeout、unavailable 和明确的临时网络错误自动重试。
- permission、authentication、invalid、conflict 和 unsupported 不自动重试。
- 健康检测使用失败阈值、指数退避和 jitter，避免全部连接每 3 秒同时请求；成功后恢复正常间隔。
- 用户主动选择连接时立即尝试，不等待后台健康检测；失败后保持选择框打开并展示统一错误状态。

### 8.9 错误与测试

- Docker SDK、Podman bindings 和 REST 错误统一映射为 `runtime.Error`；UI 只根据稳定 ErrorKind 和 operation 生成 i18n 文案，原始 cause 仅进入安全诊断日志。
- contract tests 必须覆盖 API 版本、filter 编码和 AND 语义、错误映射、TLS verified/insecure、超时取消、断线、批量部分失败及未知状态。
- mapper fixtures 分别取自 Docker 和 Podman API；Podman CGO bindings 与非 CGO REST 对同一语义必须得到一致领域结果。
- `CGO_ENABLED=0` 与 `CGO_ENABLED=1` 测试矩阵均为验收项。

## 9. 实施阶段

### Phase 0：依赖与契约探针

- 补齐 Podman bindings 所需依赖校验。
- 在 `CGO_ENABLED=0/1` 下编译 containers/images/volumes/network/system 最小调用集。
- 固化 Docker 与 Podman adapter contract tests。
- 输出 capability baseline，不改 UI 行为。

### Phase 1：领域包与连接驱动

- 建立 `runtime` 统一模型、错误、capabilities 和 factory。
- 分别创建 Docker / Podman connection 与 Ping / Version adapter。
- ConnectionPool 改持有 `runtime.Engine`。
- 保持当前配置 schema 不变。

### Phase 2：只读资源

- 迁移 container/image/volume/network list、inspect、stats、top。
- 为各资源建立统一 list options 和筛选字段，添加 native query 编码及降级 contract tests。
- adapter 内完成 mapper；删除 UI 对原始 JSON 的依赖。
- 用 golden/contract fixtures 覆盖两端字段差异。

### Phase 3：资源动作

- 迁移已有 start/stop/restart/kill/pause/unpause/rename/remove/pull/prune。
- 然后实施 `TASK-009` 的 volume/network create 与 prune。
- 统一逐目标结果、optional reclaimed space 和错误语义。

### Phase 4：流式能力

- 迁移 logs、events、exec/attach/resize。
- 删除 `Client.Raw()` 以及 `internal/tui` 对 Docker SDK 的导入。
- 完成取消、断线和资源释放测试。

### Phase 5：清理

- 删除旧 Docker facade、手写 Podman image 分叉和重复 mapper。
- 重命名 package/import alias，更新架构与开发文档。
- 检查生产代码中不再存在上层 `RuntimeType` 行为分支。

## 10. 测试策略

- Interface contract：同一测试套件分别运行 Docker 与 Podman adapter fixture。
- Mapper golden：对两端真实 API fixture 断言统一 DTO。
- Error contract：not found、conflict、unsupported、permission、connection 分类一致。
- Stream contract：logs/events/exec 的顺序、取消、EOF、断线和 goroutine 退出。
- Live integration：可用时分别连接 Docker 与 Podman socket；缺失一端时明确 skip。
- Build matrix：`CGO_ENABLED=0 go test ./...` 与 `CGO_ENABLED=1 go test ./...` 均为必选，分别覆盖 REST 与 bindings transport。

## 11. 完成标准

1. TUI 与 state 不导入 Docker / Podman SDK。
2. 不存在 `Raw()` 或上层 SDK类型。
3. ConnectionPool 与业务命令只依赖 `runtime.Engine`。
4. Docker 与 Podman 通过独立正式 adapter 实现同一契约。
5. 差异只存在于 adapter mapper、capabilities 和统一错误中。
6. `TASK-017` 的五项操作在两端 native driver contract 下通过。
7. `TASK-009` 所需 create/prune 统一 options/results 已具备。
8. CGO/非 CGO 策略明确，`just check` 与构建矩阵通过。

## 12. 非目标

- 本任务不实现 Podman pod / secret / kube 或 Docker buildx / plugin。
- 除可选 API 版本覆盖和 `tls.insecureSkipVerify` 外，不扩展现有连接配置 schema；这两个字段必须同步更新默认配置、README 和配置校验测试。
- 不在第一阶段重写 TUI 页面。
- 不用最低公共能力掩盖 runtime 专有能力；专有能力由后续任务基于 capability 扩展。
