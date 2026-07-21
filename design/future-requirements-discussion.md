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
| `FR-001` | 多连接与启动选择统一 | P0 | `approved` | 无 |
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
    TLS      RuntimeTLSConfig
    Source   ConnectionSource
}

type RuntimeTLSConfig struct {
    Enabled  bool
    Verify   bool
    CAFile   string
    CertFile string
    KeyFile  string
}
```

对应配置层建议直接扩展 `RuntimeConn`，并将健康检查放在 runtime 全局配置中:

```go
type RuntimeConn struct {
    Name    string
    Addr    string
    Runtime string
    TLS     RuntimeTLSConfig
}

type RuntimeHealthConfig struct {
    IntervalSec      int
    TimeoutSec       int
    FailureThreshold int
}
```

```yaml
runtime:
  default: local-docker
  health:
    intervalSec: 3
    timeoutSec: 2
    failureThreshold: 2
  connections:
    - name: remote-docker
      addr: tcp://docker.example.com:2376
      runtime: docker
      tls:
        enabled: true
        verify: true
        caFile: /etc/docker/certs/ca.pem
        certFile: /etc/docker/certs/cert.pem
        keyFile: /etc/docker/certs/key.pem
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

### 已确认的产品决策

1. 未指定 CLI 连接或 `runtime.default` 时，默认只检测本地 Docker；连接失败时必须同时提供用户提示和可见状态，不允许静默吞掉失败，也不自动切换到本地 Podman。
2. `--host` 表示用户明确指定本次目标，进入单连接模式，不提供 `F2` 连接切换。
3. 未使用 `--host` 时，连接选择范围包含配置中的全部连接，Podman 通过配置或用户操作显式选择。
4. TLS 连接能力纳入 FR-001 第一阶段。
5. `F2` 改为打开连接选择框；使用上下键移动，`Enter` 连接并切换，`Esc` 取消。

### 配置与覆盖优先级

从高到低:

1. CLI `--host` 显式连接，进入单连接模式
2. `runtime.default` 指向的配置连接
3. 内置本地 Docker 候选
4. `runtime.connections` 中的其他有序连接

具体语义:

- `--host` 创建仅对本次进程有效的 `cli` 连接，不写回配置，也不加载其他连接供切换。
- `--podman` 是 runtime hint；未同时提供 `--host` 时使用默认 Podman socket，但只有用户显式指定或配置默认连接为 Podman 时才尝试 Podman。
- `runtime.default` 表示首选连接名，不等同于 runtime 类型。
- `general.runtime` 仅作为连接未声明 runtime 时的探测提示；后续评估是否废弃。

### 已确认启动策略

采用“本地默认、失败显式呈现”:

1. 若存在 CLI `--host`，只尝试该 CLI 连接。
2. 若配置了有效的 `runtime.default`，优先尝试该连接。
3. 未配置显式默认连接时，只尝试内置本地 Docker 候选，不自动切换到本地 Podman。
4. 连接失败后进入可恢复的 `disconnected` / `error` 状态，不退出 TUI。
5. Header、连接信息区和提示消息必须展示当前状态、目标连接和最后错误。
6. 普通配置模式下，用户可打开 `F2` 选择框，手动选择其他配置连接并重试。

不采用静默自动切换，避免用户在不知情的情况下操作另一个 Docker / Podman 环境。

### 已确认切换语义

- `F2` 打开连接选择框，不再按键循环连接。
- 选择框按配置声明顺序展示全部连接，不依赖 map 遍历顺序。
- 上下键移动选择，`Enter` 发起连接并切换，`Esc` 取消。
- 未建立的目标连接采用 lazy connect。
- 切换失败时保留原活动连接和页面数据。
- 切换成功后清理旧资源视图的瞬态状态，并为新连接执行一次全量 Fetch。
- 使用 `--host` 时隐藏或禁用 `F2` 连接选择入口。
- 连接切换继续进入用户操作审计；后台健康检查和自动重连不进入用户审计。

### 已确认 UI 范围

第一阶段:

- Header 显示活动连接名、runtime type 和连接状态。
- `c` 显示活动连接的 name / runtime / host / engine version。
- `F2` 打开连接选择框；选择框支持上下键、Enter 和 Esc。
- `--host` 模式明确显示“用户指定连接”，隐藏连接切换入口。
- 失败通知显示连接名和简化错误；选择框保持打开并在对应连接项展示错误，状态区保留 `disconnected` / `error` 状态。
- TLS 连接显示安全连接状态，但不得展示证书内容或凭据。

后续阶段:

- 支持显式重试动作。
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
4. 修复 `--host` / `--podman` 接线，并在 `--host` 模式禁用 F2。
5. 自定义连接名可参与选择框和首次连接；零配置启动不自动探测 Podman。
6. 使用连接选择框替代 F2 循环。
7. 接通 TLS 参数、证书路径和安全错误提示。
8. 补解析、启动选择、选择框、TLS、Docker/Podman 驱动和切换测试。
9. 同步 README、配置示例和架构文档。

不包含:

- TUI 内编辑连接配置
- SSH tunnel 管理
- 凭据存储
- 集群级资源聚合视图
- 同时展示多个连接的数据
- Events 订阅实现

### 验收草案

1. `--host` 指定的连接进入实际连接池并独占本次会话。
2. `--host` 模式下 `F2` 不打开连接选择框。
3. `--podman` 在没有 `--host` 时生成正确的本地 Podman 连接，并且只有显式选择时才尝试。
4. `runtime.default` 能选择任意自定义连接名。
5. 本地默认连接失败时，状态区和提示消息都能显示目标和错误。
6. 普通配置模式下选择框按配置顺序展示全部连接。
7. 上下键、Enter、Esc 的选择框行为有测试。
8. TLS 连接成功、证书错误和连接失败均有可区分状态，Docker 与 Podman 驱动分别有测试覆盖。
9. 切换失败不会丢失当前可用连接。
10. 活动连接名、runtime type、host 在 Header、Footer、连接信息和审计记录中语义一致。
11. 旧的本地零配置启动方式继续工作。

### 已确认决策

#### D-001 首选连接失败策略

- 决定: 首选连接失败后保持 `disconnected` / `error`，展示提示和状态，由用户明确选择其他连接；未显式指定首选时使用本地连接。
- 状态: `approved`

#### D-002 `--host` 是否完全覆盖配置连接

- 决定: CLI 连接独占本次会话，`F2` 不提供其他连接切换。
- 状态: `approved`

#### D-003 `F2` 的目标集合

- 决定: 选择框展示全部配置连接，首次选中时 lazy connect。
- 状态: `approved`

#### D-004 远程 TLS 是否进入第一阶段

- 决定: TLS 纳入 FR-001 第一阶段，包含配置解析、连接建立、证书错误状态和安全提示。
- 状态: `approved`

#### D-005 是否立即新增连接选择对话框

- 决定: 第一阶段实现连接选择对话框；`F2` 打开，上下键选择，`Enter` 确认，`Esc` 取消。
- 状态: `approved`

### 下一轮待确认细节

#### D-006 本地 Docker / Podman 优先级

- 决定: 零配置默认只尝试本地 Docker，不自动切换到本地 Podman；Podman 必须由 `--podman`、`runtime.default` 或选择框显式选择。
- 状态: `approved`

#### D-007 TLS 配置形式

- 决定: 在 `RuntimeConn` 中添加独立 TLS 配置，包含 `enabled`、`verify`、`caFile`、`certFile`、`keyFile`；远程 TLS 默认必须校验证书。
- 兼容: 可兼容 Docker 常用证书目录中的 `ca.pem`、`cert.pem`、`key.pem`。
- 安全约束: 私钥路径可以进入配置，但私钥内容和凭据不得进入 UI、普通日志或审计记录。
- 状态: `approved`

#### D-008 选择连接失败后的对话框行为

- 决定: 对话框保持打开，失败项显示 error 和简化错误；原活动连接继续工作，用户可以改选或按 `Esc` 退出。
- 状态: `approved`

#### D-009 连接健康状态来源

- 决定: 健康检查默认每 3 秒执行一次并允许配置；状态至少区分 connecting / connected / disconnected / error，健康检查不自动切换活动连接。
- 状态: `approved`

#### D-010 Docker / Podman 驱动适配

- 决定: 连接声明统一使用 `ConnectionSpec`，由驱动工厂分别创建 Docker 与 Podman client；不得只通过 socket 名称猜测运行时行为。
- 验收: Docker socket、Podman socket、TCP/TLS Docker 和 TCP/TLS Podman 的连接解析、Ping、资源 Fetch 和切换分别覆盖测试。
- 状态: `approved`

#### D-011 健康检查失败与恢复阈值

- 推荐: 单次 Ping 超时 2 秒，连续 2 次失败后标记 error / disconnected，成功 1 次即可恢复。
- 推荐: 只在连接状态变化时发送提示，不对每次失败重复 Toast。
- 推荐: 活动连接按配置间隔检查；非活动连接不持续高频 Ping，在选择框打开或用户选中时按需检查。
- 理由: 3 秒间隔下，2 次阈值约 6 秒即可发现断线，同时能过滤一次瞬时超时；对所有远程连接每 3 秒 Ping 会随连接数线性增加负载。
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
- 完成 FR-001 的代码事实、候选模型、范围和验收草案。
- D-001～D-005 已根据本轮讨论确认，FR-001 进入 `approved` 状态。
