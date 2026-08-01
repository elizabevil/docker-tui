# 功能实现待办清单 (Feature TODO List)

> 建立日期: 2026-08-01
> 来源:
> - 旧 `design/future-requirements-task-list.md`(TASK-001 ~ TASK-025 历史台账,已于 2026-08-01 整合归档;完整历史保留在 [§6](#6-历史-tasks-task-001--task-025))
> - [docs/feature-design.md](feature-design.md) §4 (待实现设计目标)
> - [docs/pending-bugs.md](pending-bugs.md) (BUG 级待办快照)
>
> 目的: 把"设计目标"对应到"待开发工作项",独立跟踪。
> 与本文的关系:
> - **[feature-design.md](feature-design.md)**: 设计意图 / 决策 / 不在范围。
> - **[pending-bugs.md](pending-bugs.md)**: BUG 级待办(从 bugfix-requirements 摘出)。
> - **[bugfix-requirements.md](bugfix-requirements.md)**: BUG 修复台账(含修复记录)。
> - **[requirements.md](requirements.md)**: 能力清单简表。
>
> 本文只列**未完成**的 TASK + 设计驱动需求,**不重复** pending-bugs.md 的 BUG 列表。

## 状态定义

| 状态 | 含义 |
|---|---|
| `todo` | 范围明确,尚未开始 |
| `in_progress` | 已开始,未满足完成定义 |
| `done` | 已实现、验证并提交(在本文是历史记录) |
| `blocked` | 缺少外部条件或产品决策 |
| `cancelled` | 经决策不再实施,不计入待办 |

## 当前概况

| 状态 | 数量 | 备注 |
|---|---:|---|
| `done` | 24 | TASK-001 ~ TASK-019, TASK-021 ~ TASK-025 (详见 [§6 历史台账](#6-历史-tasks-task-001--task-025)) |
| `in_progress` | 0 | — |
| `todo` | 1 | TASK-020 (Podman 专有能力评估) |
| `blocked` | 0 | — |

## 当前执行顺序

1. 推进 `TASK-020` 的 Docker / Podman 专有能力评估(P3,等用户/产品决策后归档)。

---

## 1. 当前 todo

### TASK-020 — Docker / Podman 专有能力评估

- 状态: `todo`
- 优先级: `P3`
- 关联决策: [design/podman-capabilities-analysis.md](../design/podman-capabilities-analysis.md)
- 依赖: TASK-004 (Docker / Podman 运行时识别)
- 完成定义:
  1. 评估 Podman 独有(pod / secret / kube)与 Docker 独有(buildx / context)的产品边界。
  2. 给出"纳入 / 暂缓 / 不实现"的明确决议。
  3. 决议写入 [docs/feature-design.md §3](feature-design.md) "不在范围"列表或新建 TASK。
  4. 测试 / 文档同步。

---

## 2. 设计驱动需求

来自 [docs/feature-design.md §4](feature-design.md#4-待实现设计目标)。

| 编号 | 能力 | 优先级 | 关联 BR |
|---|---|---:|---|
| 镜像 History 顶层页 | 用户浏览镜像 layer history | high | [BR-034](bugfix-requirements.md#br-034) |
| TASK-019 容器高级动作 | Copy/Update/Diff/Export/Commit/Wait | high | [BR-033](bugfix-requirements.md#br-033) |
| Events 独立面板 + Network Connect/Disconnect | F3 浏览 events / 容器动态网络 | medium | [BR-035](bugfix-requirements.md#br-035) |
| Image Import (tarball) | 扁平 tar → 新镜像 | medium | [BR-036](bugfix-requirements.md#br-036) |
| Registry Login | docker / podman login 私库 | medium | [BR-037](bugfix-requirements.md#br-037) |
| Exec 页面 shell UI | 全屏 terminal + 真实 PTY | medium | [BR-038](bugfix-requirements.md#br-038) |
| Action Bar 取代 Q | 每页多功能面板 | **done** | [BR-039](bugfix-requirements.md#br-039) |
| Dialog 风格统一 | 四周透明 + panel 居中 | medium | [BR-040](bugfix-requirements.md#br-040) |

> 上表能力全部**在 [docs/pending-bugs.md](pending-bugs.md) 跟踪**,本表只列能力摘要与关联 BR。**实现进度请以 pending-bugs.md 为准**。

---

## 3. 已取消 / 不再实施

| 能力 | 取消原因 | 决策时间 |
|---|---|---|
| **Container Attach** | 与 logs 高度重叠;唯一差异(stdin 转发)在 dtui 只读排查场景下不常用;attach 的"接管 PID 1 + 信号转发"语义在 TUI 里易意外终止容器 | 2026-08-01 |
| **Image Build** | 终端 UI 不适合 docker build 的长日志 / 多 stage 调试;走 IDE 或 CI | 2026-08-01 |
| **Image Search** (Docker Hub) | Docker Hub 搜索意义不大,主流场景走私有仓库;走 CLI 即可 | 2026-08-01 |
| **Volume Backup** | 业界也少做;用 `docker run --volumes-from` 临时挂载出 tar 即可 | 2026-08-01 |
| **TASK-022 Phase B** | 补齐全部 DTO 与 `dto.Action` 迁移——不在本轮实施 | 2026-07-30 |
| **TASK-022 Phase C** | Client 全量方法化与破坏性签名迁移——保留现有可工作的调用形态 | 2026-07-30 |
| **TASK-022 Phase D-2** | Docker / Podman mapper 目录拆分——不在本轮实施 | 2026-07-30 |
| **TASK-022 Phase E** | `docker/service` 统一入口——架构决策明确不引入该层 | 2026-07-30 |
| **Compose up / build / pull** | 需要完整编排交互和进度模型 | (历史决议) |
| **Build / buildx** | 参数、上下文、多平台和进度交互范围较大 | (历史决议) |
| **Podman pod / secret / kube** | 等待 TASK-020 的产品边界决策 | (历史决议) |
| **Checkpoint / plugin** | 用户场景优先级低于通用运维能力 | (历史决议) |

> 取消的能力不得在其它任务中重新描述为"未修复"或"待收尾";若重新提出,必须建立新任务并重新验收。

---

## 4. 设计驱动需求(高层路线)

| 阶段 | 目标 |
|---|---|
| **P0** | TASK-019 容器高级动作 (BR-033) |
| **P1** | Image History 顶层页 (BR-034) / Network Connect (BR-035) |
| **P2** | Exec 页面 shell UI (BR-038) / Dialog 统一风格 (BR-040) / Image Import (BR-036) |
| **P3** | Registry Login (BR-037) / Action Bar 取代 Q (BR-039) / Events 独立面板 (BR-035 配套) |

---

## 5. 完成定义

任何任务标记为 `done` 前必须满足:

1. 代码行为与已确认决策一致。
2. Docker / Podman 差异收敛在 runtime/data 层,不散落 runtime 字符串分支。
3. 覆盖主路径和失败路径;双运行时能力包含双侧测试或环境验证记录。
4. 配置在读取后校验,运行期不重复修正非法配置。
5. 状态结构通过方法维护不变量。
6. `git diff --check` 和 `just check` 通过。
7. README、架构、需求状态、bugfix-requirements、feature-design、feature-todo-list、pending-bugs 已同步。
8. 形成独立提交,提交信息包含任务或需求语义。

---

## 6. 历史 TASKS (TASK-001 ~ TASK-025)

> 以下是旧 `design/future-requirements-task-list.md`(已于 2026-08-01 整合归档)的完整台账,**作为历史保留**。新任务不再使用 TASK-xxx 编号,改用 BR-xxx (BUG) 或 FR-xxx (功能需求,在 [feature-design.md](feature-design.md) 内记录)。

| 编号 | 任务 | 优先级 | 状态 | 依赖 | 完成证据 / 剩余范围 |
|---|---|---:|---|---|---|
| `TASK-001` | 连接失败错误投影与断开状态展示 | P0 | `done` | — | 启动失败不再显示虚假运行时,目标连接与错误可见 |
| `TASK-002` | F2 连接选择框 | P0 | `done` | TASK-001 | 支持上下选择、Enter 连接、Esc 取消及失败项错误保留 |
| `TASK-003` | 可配置连接健康检测 | P0 | `done` | TASK-001 | 支持检测间隔、超时、失败阈值与恢复通知,不静默切换 |
| `TASK-004` | Docker / Podman 运行时识别与统一基线 | P1 | `done` | — | 连接池、客户端和 UI 使用统一 `RuntimeType` |
| `TASK-005` | TLS 配置与客户端接线 | P0 | `done` | TASK-001 | CA、客户端证书、私钥和 ServerName 已接入 |
| `TASK-006` | endpoint / runtime identity 去重 | P1 | `done` | TASK-004 | 本地候选与配置连接按规范化标识去重并保持顺序 |
| `TASK-007` | 新配置 schema 强校验 | P1 | `done` | TASK-003、TASK-005 | 配置读取后统一校验,运行期直接使用已验证参数 |
| `TASK-008` | Docker / Podman Events 接入主循环 | P1 | `done` | TASK-003、TASK-021 | 活动连接订阅、取消、退避重连、事件合并与轮询降级已完成 |
| `TASK-009` | Volume / Network 创建与清理 | P1 | `done` | TASK-021 | create、prune、确认、部分失败反馈及双运行时契约测试已完成 |
| `TASK-010` | 批量操作扩展与部分成功反馈 | P2 | `done` | TASK-017 | 统一 `BatchActioned` 汇总、审计与资源刷新 |
| `TASK-011` | Compose / 容器 / 镜像联动刷新 | P2 | `done` | TASK-008 | `handleEventFlush` 按资源依赖图扩展刷新集 |
| `TASK-012` | 审计操作历史面板 | P2 | `done` | TASK-001 | 审计列表、筛选、详情、导航和实时同步已完成 |
| `TASK-013` | 鼠标与小终端降级布局 | P3 | `done` | 页面固定轨道 | 三档终端布局、鼠标命中路由、配置开关和测试已完成 |
| `TASK-014` | 发布、命名和跨平台分发 | P3 | `done` | 配置稳定 | 对外二进制名 `dtui`;`DTUI_VERSION` 版本注入;Linux / macOS / Windows 构建与 tag release |
| `TASK-015` | 拆分 `AppModel` 状态域 | P1 | `done` | — | 顶层模型仅协调,命名状态域维护自身不变量 |
| `TASK-016` | TLS / 证书错误分类与安全状态展示 | P0 | `done` | TASK-005 | CA、客户端证书、主机名、握手和网络错误已结构化展示 |
| `TASK-017` | 高频容器操作 | P0 | `done` | TASK-004 | pause、unpause、rename、top、port 与批量约束已完成 |
| `TASK-018` | 镜像标签与传输工作流 | P1 | `done` | TASK-021 | tag、push、save、load、进度、取消、错误和审计终态已完成 |
| `TASK-019` | 高级容器操作 (Update/Diff/Export/Commit/Wait/Copy) | P2 | **runtime 已 done,handler 待补** | TASK-017、TASK-021 | 双适配器 runtime 实现已完成;**TUI handler 缺失 → BR-033** |
| `TASK-020` | Docker / Podman 专有能力评估 | P3 | `todo` | TASK-004 | 决策 Podman pod/secret/kube 与 Docker buildx/context 的产品边界 |
| `TASK-021` | Docker / Podman 统一 runtime driver | P0 | `done` | TASK-004 | 独立 Engine、统一服务契约、错误、能力与连接池工厂已完成(设计已 archived) |
| `TASK-022` | Podman REST 适配收紧 | P1 | `done` (收缩范围) | TASK-021、TASK-024 | Phase A / D / F 已完成;Phase B / C / D-2 / E 经范围决策取消;设计已 archived |
| `TASK-023` | runtime/podman 代码规范与文档优化 | P2 | `done` | TASK-022 | package 级文档、职责边界、CGO / non-CGO 契约已同步;lint 收敛 |
| `TASK-024` | Podman 镜像详情与连接池工厂化 | P1 | `done` | TASK-021 | Podman Image Inspect、History 降级、工厂注入和 typed-nil 防御已完成 |
| `TASK-025` | lint 规则与代码整洁度治理 | P1 | `done` | — | 核心 lint 类清零,问题数 9473 → 524 |

### 6.1 TASK-022 最终范围

TASK-022 按收缩后的交付范围完成。旧设计中的 G1~G10 是方案目标,不再直接作为完成计数;最终以以下 Phase 决议为准。

| Phase | 内容 | 状态 | 说明 |
|---|---|---|---|
| A | gpgme 仅进入 CGO 构建路径 | `done` | 非 CGO 依赖检查与双构建验证已自动化 |
| B | 补齐全部 DTO 与 `dto.Action` 迁移 | `cancelled` | 不在本轮实施,不作为 TASK-022 缺口 |
| C | Client 全量方法化与破坏性签名迁移 | `cancelled` | 保留现有可工作的调用形态 |
| D | 公共 API 路径匿名结构体具名化 | `done` | stats 类型与对应测试已落地 |
| D-2 | Docker / Podman mapper 目录拆分 | `cancelled` | 不在本轮实施,不作为 TASK-022 缺口 |
| E | `docker/service` 统一入口 | `cancelled` | 架构决策明确不引入该层 |
| F | CGO / non-CGO 构建、测试、vet 与依赖矩阵 | `done` | `scripts/verify-build-matrix.sh` 和 `just check` 已接入 |

**最终架构决议**:
- 上层统一依赖 `runtime.Engine`。
- Docker 与 Podman 保留独立 adapter / service 实现。
- 不增加 `internal/data/docker/service` 统一入口。
- 已取消 Phase 不得在其他任务中被描述为"未修复"或"待收尾"。

### 6.2 已完成重点

#### TASK-013 鼠标与小终端布局
- `Unsupported`:小于 60×14,仅显示终端尺寸错误。
- `Compact`:60×14 至 79×19,单行 header/footer 与紧凑内容区。
- `Standard`:至少 80×20,完整布局。
- 鼠标仅处理左键;列表移动游标,详情页滚动,其他区域不改变键盘路径。
- `UIConfig.EnableMouse` 控制鼠标模式,默认开启。

#### TASK-025 lint 治理
已清零或完成迁移:`ineffassign` / `unused` / `errcheck` / `staticcheck` / `gosimple` / `gofmt`。统计从 9473 → 524。剩余 line length / 复杂度 / 重复代码属持续治理基线。

### 6.3 依赖与执行顺序(历史快照)

```text
TASK-001..013, TASK-015..019, TASK-021, TASK-022, TASK-024, TASK-025 [done]

TASK-022 [done]
  -> TASK-023 runtime/podman 文档与契约整理 [done]

独立后续:
  TASK-020 专有能力评估 [todo]
```

### 6.4 暂不实施 (历史决议)

| 能力 | 原因 |
|---|---|
| Compose up / build / pull | 需要完整编排交互和进度模型 |
| Build / buildx | 参数、上下文、多平台和进度交互范围较大 |
| Podman pod / secret / kube | 等待 TASK-020 的产品边界决策 |
| Checkpoint / plugin | 用户场景优先级低于通用运维能力 |

> 本节是从旧 `future-requirements-task-list.md` 迁移而来的**历史决议**。2026-08-01 后的新决议(Container Attach / Image Build / Image Search / Volume Backup 取消)请参见 [§3 已取消](#3-已取消--不再实施)。
