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

`Raw()` 必须删除。Exec create / attach / resize / close 由 `ExecService` 和统一 `ExecSession` 封装。

## 6. 统一语义与差异处理

### 6.1 统一模型

- ID 保留完整值，UI 自己生成 short ID。
- 时间统一为 `time.Time`，大小统一为 bytes。
- 状态映射到强类型枚举，同时保留只读 `NativeState` 供诊断。
- Ports、IPAM、mount、stats、process table 使用结构化模型。
- Inspect 转为结构化 detail DTO；不得让 UI 解析两端原始 JSON。

### 6.2 Capability

能力至少区分：

- `Available`：两端语义一致，可直接使用。
- `Degraded`：可实现但部分字段或行为不同，UI 必须能展示说明。
- `Unsupported`：当前 driver 或 endpoint 不支持，UI 隐藏或禁用入口。

能力由连接成功后的版本与 driver 探测产生，不允许在 UI 使用 `runtime == "podman"` 判断。

### 6.3 已知差异

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

## 7. CGO 与发布约束

仓库 `just check` 使用 `CGO_ENABLED=0`，这是验收门，不得因 native Podman driver 被破坏。

现状中 Podman bindings image 实现受 `//go:build cgo` 限制，非 CGO 使用手写 HTTP。Phase 0 编译探针已确认：即使补齐 `github.com/gorilla/schema`，同时导入 containers/images/volumes/network/system bindings 的 `CGO_ENABLED=0` 构建仍会在 `github.com/proglottis/gpgme` 处失败（其全部 Go 文件受构建约束排除）。因此 native bindings 不能作为本项目默认发布路径。

Phase 0 决策：

1. 默认 Podman driver 使用集中式 remote REST transport，保持纯 Go 与交叉编译能力。
2. bindings 可作为后续可选 transport 或实现参考，但必须实现同一 driver 契约；mapper 和业务语义只保留一份。
3. 不再为单项功能添加分散的 `*_nocgo.go` 实现；现有 image 分叉在 Podman REST adapter 覆盖后删除。

TLS transport、超时和错误分类必须由 driver factory 统一注入，Podman adapter 不得另建绕过 TLS 配置的 `http.Client`。

## 8. 实施阶段

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

## 9. 测试策略

- Interface contract：同一测试套件分别运行 Docker 与 Podman adapter fixture。
- Mapper golden：对两端真实 API fixture 断言统一 DTO。
- Error contract：not found、conflict、unsupported、permission、connection 分类一致。
- Stream contract：logs/events/exec 的顺序、取消、EOF、断线和 goroutine 退出。
- Live integration：可用时分别连接 Docker 与 Podman socket；缺失一端时明确 skip。
- Build matrix：`CGO_ENABLED=0 go test ./...` 为必选，CGO 构建作为 Podman bindings 补充检查。

## 10. 完成标准

1. TUI 与 state 不导入 Docker / Podman SDK。
2. 不存在 `Raw()` 或上层 SDK类型。
3. ConnectionPool 与业务命令只依赖 `runtime.Engine`。
4. Docker 与 Podman 通过独立正式 adapter 实现同一契约。
5. 差异只存在于 adapter mapper、capabilities 和统一错误中。
6. `TASK-017` 的五项操作在两端 native driver contract 下通过。
7. `TASK-009` 所需 create/prune 统一 options/results 已具备。
8. CGO/非 CGO 策略明确，`just check` 与构建矩阵通过。

## 11. 非目标

- 本任务不实现 Podman pod / secret / kube 或 Docker buildx / plugin。
- 不修改现有连接配置 schema。
- 不在第一阶段重写 TUI 页面。
- 不用最低公共能力掩盖 runtime 专有能力；专有能力由后续任务基于 capability 扩展。
