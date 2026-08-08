# Compose 左右栏语义统一设计 — docker × podman 概念层级

> **状态**: 设计定稿(2026-08-08)— §3 重写为设计原则, §4.3 Pod 呈现位置决策 C, §6 决策已收敛
> **触发**: 本地 podman 测试 compose, 发现左右栏语义错位、action 未随栏变化
> **关联**: docs/requirement/R08-compose/、docs/discussion/batch-operations-design.md(文档惯例)

---

## 0. 结论摘要

1. **Service 在两边都是 label 聚合, 不是物理层差异点** — docker engine 没有原生 Service 对象, podman 也没有; 唯一的物理差异是 podman 多一个 Pod 归组层, docker 并不"多一个 Service 物理层"。
2. **标签层两边统一, 物理层由 adapter 自处理** — `com.docker.compose.{project,service}` 由 compose 工具链写入, dtui 被动读取, 不试图补救或回填。聚合算法在两个 runtime 上字面等价, 不需要分支。
3. **导航固定 `Project → Service → Container` 三层**; Pod 不作独立导航层级, 而是容器的物理分组元数据(`CoLocatedGroupID` 已承载)。
4. **action 严格随 `ComposeFocus` 变化, capability = TUI 暴露范围 ∩ runtime 支持范围(双层 filter)**; 设计文档只记目标功能, adapter 实现笔记只记实现机制。

---

## 1. 问题背景

### 1.1 用户报告

> 本地使用 podman 测试 compose, 发现左右栏语义有变化, 且 action 也应随栏的变化而改变。重新设计左右栏概念。

用户进一步澄清意图: **重新定义语义, 梳理 podman 与 docker 在 compose 上的概念层级关系, 将两者概念统一, 方便后续功能开发**。

### 1.2 现状(左栏 / 右栏)

| | 当前实现 |
|---|---|
| 左栏 | 项目列表(`ComposeFocus == 0`), 列: project / status / services 数 / pods 数 |
| 右栏 | 选中项目的服务列表(`ComposeFocus == 1`), 列: service / image / pods(容器数) |
| 容器子视图 | 右栏 Enter 进入, 按 service 过滤显示容器 |

Action bar(registry.go `Compose()`)已部分随焦点变化:
- 左栏: `Enter/→ services`、`d detail`、`s start`、`Ctrl+S stop`、`l logs`、`Ctrl+D down`
- 右栏: `Enter expand`、`← projects`、`s start`、`Ctrl+S stop`、`l logs`、`Ctrl+D down`

### 1.3 问题的本质

**当前 UI 是 docker 层级结构的镜像**(左=项目、右=服务), 而 podman 的真实层级是 "项目≡Pod → 容器", Service 只是标签维度。podman 下右栏显示的"服务列表"是标签聚合, 真正的物理分组(pod)无处呈现 → 语义错位。

---

## 2. 概念层级对比

### 2.1 Docker Compose(纯逻辑分层)

```text
Project           纯逻辑命名空间(compose.yaml 所在目录), 无物理实体
 └─ Service       逻辑服务; 可 scale 出 N 个副本; docker engine 内无对应原生物理对象,
                  由容器上的 com.docker.compose.service 标签聚合而成
      └─ Container  物理实体(replica); 相互独立, 同 netns 需 N+1 inspect 探测
```

- 层级: **Project → Service → Container**, 全程是 "逻辑 → 逻辑 → 物理"; Project 与 Service 都是逻辑层, 唯一物理层是 Container
- 标签: `com.docker.compose.{project, service, container-number, image}`
- Co-located 探测: Degraded(N+1 inspect + 5s TTL)

### 2.2 podman-compose(--in-pod=true 为 1.1.x 默认)

```text
Project ≡ Pod     物理实体! 默认生成 pod_<project>(含 infra 容器)
 └─ Container     直接挂到 pod 下
      └─ Service  非独立层级, 仅容器上的 com.docker.compose.service 标签
```

- 层级: **Project(label, 但有物理归组 Pod) → Container(物理)**; 中间层是 Pod, 不是 Service
- 标签: 同时打 `io.podman.compose.*`(原生) 与 `com.docker.compose.*`(兼容)
- Co-located 探测: Available(`inspect.Pod` 字段, 零 RTT)

### 2.3 差异对照

| 维度 | Docker | podman |
|---|---|---|
| Service 层来源 | 标签聚合(`com.docker.compose.service`) | 标签聚合(`com.docker.compose.service` 或 `io.podman.compose.service`) |
| Service 是否物理 | 否, 逻辑层 | 否, 逻辑层 |
| 物理归组 | 无 | Pod(默认 1 项目 1 pod) |
| 右栏"服务列表" | 标签分组 | 标签分组; 物理分组(pod)不可见 |

---

## 3. 设计原则

> 本节取代原"为什么 podman 打 docker 标签"的事实描述, 升格为可复用的设计原则。

### 3.1 标签层统一, 物理层由 adapter 自处理

两边都打 `com.docker.compose.{project,service}`, 值来自同一份 compose.yaml, 因此:

- **聚合算法在两个 runtime 上字面等价, 不需要任何分支**
- 物理归组的差异(docker 无 pod / podman 有 pod)由各 adapter 在 list 阶段映射到 `CoLocatedGroupID` 字段, **不进入聚合层**
- `gatherComposeProjects`(view.go:36)基于这两个 label 聚合, runtime-agnostic

### 3.2 接口契约 = 用户意图 + 必要参数, 实现路径是 adapter 内部细节

`ComposeService` 接口只描述:

1. **操作签名** (`Start(ctx, projectID, options) error`)
2. **参数语义** (`LifecycleOptions.TimeoutSec int` 含义)
3. **能力声明** (adapter 自报 `Capability.ComposeBuild: Unsupported`)
4. **错误码** (统一 `ErrComposeXXXUnsupported` / `ErrInvalidOption`)

**不**描述:

- adapter 内部如何遍历容器 / 调哪些底层 API / 串哪些请求
- runtime 在底层实现该操作的"路径"

**设计文档只记目标功能, 实现笔记只记实现机制**。两者泾渭分明。

### 3.3 接口实现的验收清单 — 以 Down 为例

`ComposeService.Down` 必须达到以下终态(验收标准, **不是**实现描述):

| 终态 | 默认 | 可选参数 |
|---|---|---|
| project 所有容器已停止 | ✅ | — |
| project 所有容器已删除 | ✅ | — |
| project 创建的网络已删除 | ✅ | — |
| project 引用的镜像已删除 | ❌ | `RemoveImages ∈ {None, Local, All}` |
| project 创建的卷已删除 | ❌ | `RemoveVolumes bool` |

adapter 各自怎么达到这些终态(docker 逐个删 vs podman 删整个 pod)属于实现细节, **不进设计文档**, 进 `internal/data/runtime/{docker,podman}/compose.go` 实现笔记。

### 3.4 dtui 是被动消费者

dtui 是 compose 数据的**被动消费者**, 不是作者:

| 动作 | 谁负责 |
|---|---|
| 写 compose.yaml | 用户 |
| 写容器 label `com.docker.compose.*` | docker-compose / podman-compose |
| 创建 / 删除容器, 网络, 卷 | docker engine / podman engine |
| **读 label, 读状态** | **dtui**(只读) |
| **通过 ComposeService 发起 Start / Stop / Down** | **dtui**(调用) |

因此:

- **`com.docker.compose.*` label 在 podman 下也存在是 podman-compose 工具自身的责任**, 是它与生态(watchtower / portainer / lazydocker / dtui)的契约, 不是 dtui 要去补救的事
- dtui **不**写 label, **不**修改 compose.yaml, **不**做"读不到就回填"的兜底逻辑 — 读不到就视为 empty
- `CoLocatedGroupID` 空就显示"(无 Pod 信息)", 而不是"假定存在却没读到"
- 第三方 runtime(nerdctl / lima 等)只需回答一个问题: **是否写 `com.docker.compose.*`?** 写就纳入, 不写就不进 compose 视图(项目数为 0)

### 3.5 Capability = TUI 暴露范围 ∩ runtime 支持范围(双层 filter)

TUI 项目**不**做完整 lifecycle, capability 评估必须先过 L1 = TUI 是否暴露, 再过 L2 = runtime 是否支持:

| operation | L1 TUI 暴露 | docker (L2) | podman (L2) | 来源 |
|---|---|---|---|---|
| List projects | ✅ | ✅ | ✅ | R08-03 |
| Project detail | ✅ | ✅ | ✅ | R08-03 |
| Service queries(top/port/stats) | ✅ | ✅ | ✅(待 podman-compose 验证) | R08-07 |
| Logs(聚合) | ✅ | ✅ | ✅ | R08-06 |
| Start / Stop / Restart | ✅ | ✅ | ✅ | R08-04 |
| Pause / Unpause / Kill | ✅ | ✅ | ✅ | R08-09 |
| Down | ✅ | ✅ | ✅ | R08-04 |
| Scale | ✅ | ✅ | ✅(数字约束) | R08-09 |
| Rm(服务容器) | ✅ | ✅ | ✅ | R08-09 |
| Prune(项目范围) | ✅ | ✅ | ✅ | R08-09 |
| Events(按 project/service 过滤) | ✅ | ✅ | ✅ | R08-10 |
| Exec / Run | ✅ | ✅ | ✅ | R08-08 |

**L1 直接砍掉, 不进入 capability 评估的 operation**:

| operation | 不做的理由 |
|---|---|
| `compose config` | spec 解析工程量过大; R08-02 已定 `ErrComposeUnsupported` |
| `compose build` | Dockerfile 解析 + 构建引擎 ≠ TUI 职责 |
| `compose pull / push` | 镜像级动作走 R02 镜像页 |
| `compose up` | 包含 build/pull, 等同不做 |

capability 表**只列 L1 已通过的 operation**, 其余不进设计讨论。

### 3.6 边界: 镜像动作归 R02, compose 编排动作归 R08

两块边界**不混**:

- **R02 镜像域**: 镜像的 list / pull / push / tag / save / load / delete / history — 都是**单镜像**操作
- **R08 compose 编排**: project / service / 容器子视图 — 都是**项目范围**操作, 但其底层动作**复用** R01(容器级) 和 R02(镜像级) 的 adapter

compose 视角下的"批量 pull" **不做** — 用户应该在镜像页面对单个镜像操作, 不在 compose 范围批量拉取。

---

## 4. 统一概念模型

### 4.1 导航层级(固定三层, 两边一致)

```text
Compose Project ── 逻辑项目(两边统一)
   └── Service ── 逻辑服务(两边统一: docker / podman 都是标签聚合)
        └── Container ── 物理容器(两边统一)
   └── Pod ── podman 专属物理分组 → 不作为独立导航层级,
              作为容器的物理分组元数据(CoLocatedGroupID 字段已存在)
```

**核心决策**:

1. 导航固定 `Project → Service → Container` — 生命周期操作按 Project 走, 操作模型不分裂
2. Pod 是**属性不是层级** — 物理分组信息在容器视图以列 / 标记呈现, 不新增导航维度
3. 标签层聚合逻辑两边共享 — `gatherComposeProjects` 不分支

### 4.2 左右栏语义定义

| | 语义 | 内容 | 注释 |
|---|---|---|---|
| 左栏 | 项目(Project) | 项目名 / 状态 / 服务数 / 容器数 | 两 runtime 一致 |
| 右栏 | 服务(Service) | 服务名 / image / 容器数 / Pod 状态指示符 | 两 runtime 一致(标签聚合) |
| 容器子视图 | 容器 | 按 service 过滤的容器列表; podman 额外显示完整 Pod 列 | 两 runtime 一致 |

> **决策**: 右栏标题统一保持 "Services", **不**在 podman 下加 "(标签聚合)" — 标签聚合是实现细节, 不应泄露到 UI。

### 4.3 Pod 信息的呈现位置(决策: C, B 为底线)

| 选项 | 评估 |
|---|---|
| A: 右栏 Service 加 Pod 列 | ❌ Service 是逻辑分组, Pod 是物理分组, 同行承担两轴含义, 排序 / 筛选语义会乱 |
| B: 容器视图加 Pod 列 | ✅ 导航层保持"一栏一语义", 物理信息下沉到最近一层; **底线方案** |
| **C: 两者都加(决策)** | ✅ 右栏 Service 行只显示 Pod 状态指示符(前缀字符 / 小色块), 不占列宽; 容器视图展开完整 Pod 列做精细动作 |

**实施细节**: 右栏 Pod 指示符做成纯样式(前缀字符或小色块), 保持右栏列数与列含义不变。

### 4.4 action 随栏变化规则

| 焦点 | action 集 | 过滤 |
|---|---|---|
| 左栏(项目) | Enter/→ 进服务、d 详情、s 启动、Ctrl+S 停止、l 日志、Ctrl+D down | 项目级 capability(§3.5) |
| 右栏(服务) | Enter 进容器、← 回项目、s 启动、Ctrl+S 停止、l 日志、Ctrl+D down | 服务级 capability(§3.5) |
| 容器子视图 | Esc/← 返回、↑/↓ 滚动 | — |

规则:

1. action 提示严格由 `ComposeFocus` 决定(现状已基本满足, 需回归确认)
2. 不可用操作按 capability 隐藏 / 置灰(详见 §3.5)
3. 后续新增项目级 / 服务级操作时, 按本表归位

---

## 5. 与现有代码的关系

| 位置 | 现状 | 需要改动 |
|---|---|---|
| `internal/tui/state/compose.go` | `ComposeFocus` 0/1 已定义 | 基本不变 |
| `internal/tui/ui/pages/compose/view.go` | 左=项目、右=服务 | 按 §4.3 决策加 Pod 指示符 + 容器视图 Pod 列 |
| `internal/tui/ui/action/registry.go` | `Compose()` 已随焦点变化 | 回归确认 + capability 过滤(对接 §3.5) |
| `internal/tui/keyboard/compose_nav.go` | ←/→/Enter 切换焦点 | 基本不变 |
| `internal/data/runtime/podman/co_located.go` | `CoLocatedGroupID="pod_<project>"` | 已具备, 供 Pod 列使用 |
| `internal/data/runtime/containers.go` | `CoLocatedGroupID` 字段 | 已具备 |

---

## 6. 设计决策(原开放问题已收敛)

| # | 问题 | 决策 | 详见 |
|---|---|---|---|
| 1 | Pod 信息呈现位置 | **C, B 为底线** | §4.3 |
| 2 | 右栏标题语义 | 保持 "Services", 不泄露 "(标签聚合)" | §4.2 |
| 3 | 改动范围 | 只做 UI 呈现 + action 跟随, 聚合逻辑不动 | §5 |
| 4 | 是否引入 Pod 过滤 / 排序 | 暂不 | §4.3 |

后续若有新 runtime(nerdctl / lima 等)接入, 按 §3.4 / §3.5 评估 L1 暴露范围与 L2 支持范围。

---

## 7. 后续步骤

1. ✅ 用户决策 §6 开放问题(2026-08-08)
2. ✅ §3 重写为设计原则(2026-08-08)
3. ✅ §4.3 Pod 呈现位置决策 C(2026-08-08)
4. 按 writing-plans 生成实施计划(如需代码改动)