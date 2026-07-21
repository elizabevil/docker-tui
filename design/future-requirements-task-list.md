# 后续需求实施任务清单

> 建立日期: 2026-07-20  
> 最近整理: 2026-07-21
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
| `done` | 7 |
| `in_progress` | 1 |
| `todo` | 12 |
| `blocked` | 0 |

当前执行队列：

1. 完成 `TASK-015`，继续拆分 `AppModel` 的独立状态域。
2. 执行 `TASK-016`，补齐 TLS / 证书错误的专门状态展示。
3. 执行 `TASK-017`，补齐高频容器操作。
4. 执行 `TASK-009`，补齐 Volume / Network 创建与清理。
5. 执行 `TASK-008`，将 Events 接入主循环并支持局部刷新。

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

相关提交：`7cf479a`、`948d9d2`、`aa7567d`、`4fe226d`。

## 当前架构任务

| 编号 | 任务 | 优先级 | 状态 | 已完成 / 剩余范围 |
|---|---|---:|---|---|
| `TASK-015` | 拆分 `AppModel` 状态域并封装状态方法 | P1 | `in_progress` | 已拆 `ConnectionState`；待拆 Log、Detail、Dialog、Exec、Compose、Notification，消除共享输入字段和 `state -> ui/component` 依赖 |

`TASK-015` 采用逐域迁移，不进行一次性重写。顶层 `AppModel` 继续负责 Bubble Tea 协调，子状态负责维护自身不变量。

## Runtime 与错误体验

| 编号 | 任务 | 优先级 | 状态 | 依赖 | 验收重点 |
|---|---|---:|---|---|---|
| `TASK-016` | TLS / 证书错误分类与安全状态展示 | P0 | `todo` | TASK-005 | 区分 CA、客户端证书、主机名、握手和网络错误；不得展示证书内容或凭据 |

`TASK-005` 只代表配置和客户端链路已完成。TLS 专项错误分类、选择框提示和安全连接标识归 `TASK-016`，不得混记为已完成。

## 实时同步与资源操作

| 编号 | 任务 | 优先级 | 状态 | 依赖 | 验收重点 |
|---|---|---:|---|---|---|
| `TASK-008` | Docker / Podman Events 接入主循环 | P1 | `todo` | TASK-003、TASK-004 | 生命周期管理、断线恢复、事件合并、局部刷新和无事件降级 |
| `TASK-009` | Volume / Network 创建与清理 | P1 | `todo` | TASK-004 | create、prune、确认交互、部分失败反馈及双运行时测试 |
| `TASK-010` | 批量操作扩展与部分成功反馈 | P2 | `todo` | 审计模型、TASK-017 | 每个目标独立终态、汇总提示和可追溯审计 |
| `TASK-011` | Compose / 容器 / 镜像联动刷新 | P2 | `todo` | TASK-008 | 事件只使相关资源失效，不直接修改复杂 UI 状态 |
| `TASK-017` | 高频容器操作 | P0 | `todo` | TASK-004 | `pause`、`unpause`、`rename`、`top`、`port`，Docker / Podman 行为一致 |
| `TASK-018` | 镜像标签与传输工作流 | P1 | `todo` | TASK-004 | `tag`、`push`、`save`、`load`；进度、取消和错误可见 |
| `TASK-019` | 高级容器操作 | P2 | `todo` | TASK-017 | 评估并分批实现 `update`、`diff`、`export`、`commit`、`wait`、`cp` |

`build` 需要独立输入和进度交互设计，暂不并入 `TASK-018`，待该任务完成后再建立实施项。

## 操作历史与体验

| 编号 | 任务 | 优先级 | 状态 | 依赖 | 验收重点 |
|---|---|---:|---|---|---|
| `TASK-012` | 审计操作历史面板 | P2 | `todo` | TASK-001、审计模型 | trace 查询、结果筛选、敏感信息边界和大文件读取策略 |
| `TASK-013` | 鼠标与小终端降级布局 | P3 | `todo` | 页面固定轨道 | 鼠标不破坏键盘路径，小终端无重叠且核心操作可达 |
| `TASK-014` | 发布、命名和跨平台分发 | P3 | `todo` | 新配置稳定 | 统一命名、构建产物、版本信息、Linux / macOS 发布说明 |
| `TASK-020` | Docker / Podman 专有能力评估 | P3 | `todo` | TASK-004 | Podman pod/secret/kube 与 Docker buildx/context 等能力分开决策 |

## 依赖与顺序

```text
已完成连接基础 (TASK-001..007)
  |-> TASK-015 状态域拆分
  |-> TASK-016 TLS 错误体验
  |-> TASK-017 高频容器操作 -> TASK-010 / TASK-019
  |-> TASK-009  Volume / Network
  |-> TASK-018 镜像工作流
  `-> TASK-008 Events -> TASK-011 联动刷新

独立后续: TASK-012 / TASK-013 / TASK-014 / TASK-020
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
