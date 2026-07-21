# 后续需求与规划讨论纪要

> 建立日期: 2026-07-20  
> 当前阶段: 需求讨论  
> 作用: 保存后续产品需求、技术边界、候选方案和待确认决策。  
> 说明: 本文不是已批准设计，也不表示相关能力已经实现。

## 文档规则

1. 先记录代码事实和用户场景，再讨论方案。
2. 每个主题使用独立编号 `FR-xxx`。
3. 决策状态分为 `discussing`、`approved`、`rejected`、`deferred`。
4. 方向确认后，再同步到正式需求、设计和执行任务文档。
5. 实现完成后，更新 [需求规格](../docs/requirements.md) 与 [架构说明](../docs/architecture.md)。

## 规划总览

| 编号 | 主题 | 优先级 | 状态 | 主要依赖 |
|---|---|---|---|---|
| `FR-001` | 多连接与启动选择统一 | P0 | `discussing` | 无 |
| `FR-002` | Docker Events 实时资源同步 | P1 | `discussing` | FR-001 |
| `FR-003` | Volume / Network 创建 | P2 | `discussing` | FR-001 |
| `FR-004` | 批量操作扩展 | P2 | `discussing` | 审计模型已完成 |
| `FR-005` | 操作历史面板 | P2 | `discussing` | 审计模型已完成 |
| `FR-006` | 鼠标、小屏与交互完善 | P3 | `discussing` | 页面骨架 |
| `FR-007` | 命名、发布与分发统一 | P3 | `discussing` | FR-001 |

## 推荐推进顺序

```text
FR-001 连接体系
  -> FR-002 Events 实时同步
  -> FR-003/004 资源操作补齐
  -> FR-005 操作历史界面
  -> FR-006/007 体验与发布
```

原因:

- Events 订阅必须绑定明确的活动连接和连接生命周期。
- 新资源操作依赖可靠的目标 runtime，且需要复用审计上下文。
- 操作历史的数据基础已经存在，可以在核心资源链路稳定后独立推进。

---

## FR-001 多连接与启动选择统一

### 讨论目标

建立唯一的连接配置与选择模型，使 CLI、配置文件、自动探测、首次连接、运行时切换和 UI 展示使用同一数据来源。

### 用户场景

1. 本地只运行 Docker，程序应直接连接 Docker。
2. 本地只运行 Podman，程序应直接连接 Podman。
3. Docker 与 Podman 同时存在，用户可指定默认连接并用 `F2` 切换。
4. 用户通过配置添加多个本地或远程连接，并选择默认连接。
5. 用户通过 `--host` 临时覆盖本次启动目标，但不修改配置文件。
6. 首选连接失败时，用户应看到失败原因，并按既定策略决定是否回退其他连接。

### 已核对的代码事实

1. CLI `--host` 与 `--podman` 当前只修改 `cfg.Docker.Host`。
2. 启动时连接池只从 `cfg.Runtime.Connections` 构造；为空时注入 `local-docker` 与 `local-podman`。
3. 因此 CLI 写入的 `cfg.Docker.Host` 没有进入连接池主路径。
4. `connectDocker()` 首次连接顺序硬编码为 `local-podman`、`local-docker`。
5. 自定义连接名即使进入连接池，也不会被首次连接流程尝试。
6. `RuntimeConfig.Default` 已定义，但首次连接没有读取它。
7. `ConnectionPool.KnownHostNames()` 从 map 取值，`F2` 的循环顺序不稳定。
8. `HostEntry` 已定义 Runtime / TLS / CertPath，但 `ConnectionPool.AddHost()` 没有完整保留这些字段。
9. 连接成功后 Header、Footer、审计上下文分别读取活动连接名、runtime type 和 host，语义尚未完全收敛。
10. `PingLoop()` 已存在，但当前主程序没有启动该连接健康检查循环。

### 问题定义

这不是单个参数失效问题，而是存在三套并行来源:

- 旧入口: `docker.host` 与 CLI `--host`
- 新入口: `runtime.connections` 与 `runtime.default`
- 启动策略: `connectDocker()` 中的硬编码本地连接名

如果不先统一来源，后续 Events、远程连接、重连和连接选择 UI 都会继续产生分支逻辑。

### 推荐领域模型

建议以有序 `ConnectionSpec` 列表作为唯一连接输入:

```go
type ConnectionSpec struct {
    Name     string
    Host     string
    Runtime  RuntimeType
    TLS      bool
    CertPath string
    Source   ConnectionSource
}
```

其中 `Source` 用于说明连接来自:

- `cli`
- `config`
- `builtin`

连接池负责运行状态，不负责重新解释配置优先级:

```go
type ConnectionState struct {
    Spec      ConnectionSpec
    Status    ConnState
    Client    *Client
    LastError error
}
```

### 推荐配置与覆盖优先级

从高到低:

1. CLI 显式连接参数
2. `runtime.default` 指向的配置连接
3. `runtime.connections` 中的有序连接
4. 内置本地 Docker / Podman 候选

推荐语义:

- `--host` 创建仅对本次进程有效的 `cli` 连接，不写回配置。
- `--podman` 是 runtime hint；未同时提供 `--host` 时使用默认 Podman socket。
- `runtime.default` 表示首选连接名，不等同于 runtime 类型。
- `general.runtime` 仅作为连接未声明 runtime 时的探测提示；后续评估是否废弃。

### 推荐启动策略

推荐采用“首选优先、允许回退”:

1. 若存在 CLI 连接，先尝试 CLI 连接。
2. 否则先尝试 `runtime.default`。
3. 首选失败后，按配置顺序尝试其他允许回退的连接。
4. 全部失败时进入可恢复的 disconnected 状态，而不是退出 TUI。
5. UI 保留每个尝试结果和最后错误，允许用户重试或切换。

这个策略兼顾默认体验和可用性。生产场景若要求“默认连接失败即停止”，可以后续增加 `strictDefault`，不建议第一版默认严格失败。

### 推荐切换语义

- `F2` 按配置顺序循环全部连接，而不是依赖 map 顺序。
- 未建立的目标连接采用 lazy connect。
- 切换失败时保留原活动连接和页面数据。
- 切换成功后清理旧资源视图的瞬态状态，并为新连接执行一次全量 Fetch。
- 连接切换继续进入用户操作审计；后台健康检查和自动重连不进入用户审计。

### 推荐 UI 范围

第一阶段:

- Header 显示活动连接名、runtime type 和连接状态。
- `c` 显示活动连接的 name / runtime / host / engine version。
- `F2` 保留快速循环切换。
- 失败通知明确显示连接名和简化错误。

后续阶段:

- 增加连接选择对话框，展示全部连接状态和最后错误。
- 支持显式重试。
- 是否支持在 TUI 内新增、编辑、删除连接，另行讨论。

### 安全边界

- 不在 Toast、Footer 或审计日志中记录 TLS 私钥、证书正文或凭据。
- TCP 远程连接不得默认关闭 TLS 校验。
- 配置路径可以记录，敏感文件内容不能进入日志。
- 切换连接不得把前一个 runtime 的资源选择或 marked IDs 应用到新 runtime。

### 第一阶段建议范围

包含:

1. 统一 CLI、配置和内置候选的连接解析。
2. 让 `runtime.default` 生效。
3. 保留连接声明顺序。
4. 修复 `--host` / `--podman` 接线。
5. 自定义连接名可参与首次连接。
6. 稳定 `F2` 顺序和失败回退。
7. 补解析、启动选择和切换测试。
8. 同步 README、配置示例和架构文档。

不包含:

- TUI 内编辑连接配置
- SSH tunnel 管理
- 凭据存储
- 集群级资源聚合视图
- 同时展示多个连接的数据
- Events 订阅实现

### 验收草案

1. `--host` 指定的连接进入实际连接池并优先尝试。
2. `--podman` 在没有 `--host` 时生成正确的本地 Podman 连接。
3. `runtime.default` 能选择任意自定义连接名。
4. 首选连接失败后的回退行为可预测并有测试。
5. `F2` 顺序与配置声明顺序一致。
6. 切换失败不会丢失当前可用连接。
7. 活动连接名、runtime type、host 在 Header、Footer、连接信息和审计记录中语义一致。
8. 旧的本地零配置启动方式继续工作。

### 待确认决策

#### D-001 首选连接失败策略

- 推荐: 自动回退，并向用户明确显示首选失败。
- 备选: 首选失败后保持 disconnected，由用户手动选择。
- 状态: `pending`

#### D-002 `--host` 是否完全覆盖配置连接

- 推荐: 本次启动优先使用 CLI 连接，但仍保留配置连接供 `F2` 切换。
- 备选: CLI 模式下连接池只保留 CLI 连接。
- 状态: `pending`

#### D-003 `F2` 的目标集合

- 推荐: 循环全部配置连接，首次切换时 lazy connect。
- 备选: 只循环已经连接成功的连接。
- 状态: `pending`

#### D-004 远程 TLS 是否进入第一阶段

- 推荐: 数据模型保留字段，但第一阶段只保证现有 socket 与普通 host 行为；TLS 作为独立验收批次。
- 备选: 第一阶段同时完成 TLS 远程连接。
- 状态: `pending`

#### D-005 是否立即新增连接选择对话框

- 推荐: 第一阶段保留 `F2` + `c`，连接选择对话框在基础模型稳定后实现。
- 备选: 与连接模型同时交付选择对话框。
- 状态: `pending`

---

## FR-002 Docker Events 实时资源同步

### 初步边界

- 订阅活动连接的 container / image / volume / network events。
- 事件只触发资源失效和局部刷新，不直接修改复杂 UI 状态。
- 连接切换时取消旧订阅并建立新订阅。
- 断线后采用有上限的退避重连。
- 保留低频全量刷新作为一致性兜底。
- 后台事件属于 runtime 状态，不写入用户操作审计。

### 依赖 FR-001 的原因

Events 流必须明确归属于哪个活动连接，也必须在连接切换、断线和关闭时正确取消。连接生命周期没有统一前，不应先接入长期运行的事件 goroutine。

### 后续讨论点

- 不同事件类型对应的最小刷新范围
- 高频事件合并窗口
- Events 与 Stats 轮询的职责边界
- 断线与恢复在 Header / Footer 中的展示
- 如何测试取消、重连和无 goroutine 泄漏

---

## 其他主题待展开

### FR-003 Volume / Network 创建

待讨论创建表单模型、字段验证、Docker / Podman 差异和审计 target。

### FR-004 批量操作扩展

待讨论支持矩阵、部分成功语义、每个资源独立 trace 还是父子 trace。

### FR-005 操作历史面板

待讨论内存历史与文件回放边界、trace 聚合、过滤、导出和保留策略。

### FR-006 鼠标、小屏与交互完善

待讨论最小终端尺寸、降级布局、鼠标点击目标和键盘优先原则。

### FR-007 命名、发布与分发统一

待讨论 `dtui` / `docker-tui` 对外名称、二进制命名、版本注入、CI 和发布制品。

## 本轮讨论输出

- 建立后续规划总览。
- 确认优先讨论 FR-001，再讨论 FR-002。
- 完成 FR-001 的代码事实、候选模型、范围、验收和五个待确认决策草案。
- 尚未批准任何实现方案。

