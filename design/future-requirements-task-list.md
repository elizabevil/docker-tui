# 后续需求实施任务清单

> 建立日期：2026-07-20
> 最近整理：2026-07-30
> 依据：[后续需求与规划讨论](future-requirements-discussion.md)、[Docker / Podman 能力分析](podman-capabilities-analysis.md)、[Podman REST 迁移设计](podman-rest-migration.md)
> 规则：本文是任务状态的唯一台账；设计文档用于说明方案，不单独决定任务状态。

## 状态定义

| 状态 | 含义 |
|---|---|
| `todo` | 范围明确，尚未开始 |
| `in_progress` | 已开始，但尚未满足完成定义 |
| `done` | 已实现、验证并提交 |
| `blocked` | 缺少外部条件或产品决策 |
| `cancelled` | 经决策不再实施，不计入待办 |

## 当前概况

| 状态 | 数量 | 任务 |
|---|---:|---|
| `done` | 24 | TASK-001～013、TASK-014、TASK-015～019、TASK-021～025 |
| `in_progress` | 0 | — |
| `todo` | 1 | TASK-020 |
| `blocked` | 0 | — |

当前执行顺序：

1. 推进 `TASK-020` 的 Docker / Podman 专有能力评估。

## 任务总览

| 编号 | 任务 | 优先级 | 状态 | 依赖 | 完成证据 / 剩余范围 |
|---|---|---:|---|---|---|
| `TASK-001` | 连接失败错误投影与断开状态展示 | P0 | `done` | — | 启动失败不再显示虚假运行时，目标连接与错误可见 |
| `TASK-002` | F2 连接选择框 | P0 | `done` | TASK-001 | 支持上下选择、Enter 连接、Esc 取消及失败项错误保留 |
| `TASK-003` | 可配置连接健康检测 | P0 | `done` | TASK-001 | 支持检测间隔、超时、失败阈值与恢复通知，不静默切换 |
| `TASK-004` | Docker / Podman 运行时识别与统一基线 | P1 | `done` | — | 连接池、客户端和 UI 使用统一 `RuntimeType` |
| `TASK-005` | TLS 配置与客户端接线 | P0 | `done` | TASK-001 | CA、客户端证书、私钥和 ServerName 已接入 |
| `TASK-006` | endpoint / runtime identity 去重 | P1 | `done` | TASK-004 | 本地候选与配置连接按规范化标识去重并保持顺序 |
| `TASK-007` | 新配置 schema 强校验 | P1 | `done` | TASK-003、TASK-005 | 配置读取后统一校验，运行期直接使用已验证参数 |
| `TASK-008` | Docker / Podman Events 接入主循环 | P1 | `done` | TASK-003、TASK-021 | 活动连接订阅、取消、退避重连、事件合并与轮询降级已完成 |
| `TASK-009` | Volume / Network 创建与清理 | P1 | `done` | TASK-021 | create、prune、确认、部分失败反馈及双运行时契约测试已完成 |
| `TASK-010` | 批量操作扩展与部分成功反馈 | P2 | `done` | TASK-017 | 统一 `BatchActioned` 汇总、审计与资源刷新 |
| `TASK-011` | Compose / 容器 / 镜像联动刷新 | P2 | `done` | TASK-008 | `handleEventFlush` 按资源依赖图扩展刷新集 |
| `TASK-012` | 审计操作历史面板 | P2 | `done` | TASK-001 | 审计列表、筛选、详情、导航和实时同步已完成 |
| `TASK-013` | 鼠标与小终端降级布局 | P3 | `done` | 页面固定轨道 | 三档终端布局、鼠标命中路由、配置开关和测试已完成；提交 `9d14ae5` |
| `TASK-014` | 发布、命名和跨平台分发 | P3 | `done` | 配置稳定 | 对外二进制名统一为 `dtui`；`DTUI_VERSION` 支持版本注入，默认值为 `0.2.0`；已补充 Linux / macOS / Windows 构建、CI 与 tag release workflow；README / docs 已同步发布与下载说明 |
| `TASK-015` | 拆分 `AppModel` 状态域 | P1 | `done` | — | 顶层模型仅负责协调，命名状态域维护自身不变量 |
| `TASK-016` | TLS / 证书错误分类与安全状态展示 | P0 | `done` | TASK-005 | CA、客户端证书、主机名、握手和网络错误已结构化展示 |
| `TASK-017` | 高频容器操作 | P0 | `done` | TASK-004 | pause、unpause、rename、top、port 与批量约束已完成 |
| `TASK-018` | 镜像标签与传输工作流 | P1 | `done` | TASK-021 | tag、push、save、load、进度、取消、错误和审计终态已完成 |
| `TASK-019` | 高级容器操作 | P2 | `done` | TASK-017、TASK-021 | update、diff、export、commit、wait、copy 的双适配器实现和测试已完成 |
| `TASK-020` | Docker / Podman 专有能力评估 | P3 | `todo` | TASK-004 | 决策 Podman pod/secret/kube 与 Docker buildx/context 的产品边界 |
| `TASK-021` | Docker / Podman 统一 runtime driver | P0 | `done` | TASK-004 | 独立 Engine、统一服务契约、错误、能力与连接池工厂已完成 |
| `TASK-022` | Podman REST 适配收紧 | P1 | `done` | TASK-021、TASK-024 | Phase A/D/F 已完成；Phase B/C/D-2/E 经范围决策取消 |
| `TASK-023` | runtime/podman 代码规范与文档优化 | P2 | `done` | TASK-022 | package 级文档、职责边界、CGO / non-CGO 契约已同步到架构文档；`git diff --check` 通过，`just verify::check` 的非 CGO 子矩阵通过，CGO 子矩阵受沙箱缺失 `pkg-config` / `btrfs` 头文件限制 |
| `TASK-024` | Podman 镜像详情与连接池工厂化 | P1 | `done` | TASK-021 | Podman Image Inspect、History 降级、工厂注入和 typed-nil 防御已完成 |
| `TASK-025` | lint 规则与代码整洁度治理 | P1 | `done` | — | 核心 lint 类清零或收敛，问题数 9473 → 524；提交序列已落地 |

## 当前任务

### TASK-014：发布、命名和跨平台分发

状态：`done`

已完成：

- 对外二进制名统一为 `dtui`。
- `DTUI_VERSION` 已接入构建链路；默认版本为 `0.2.0`，tag 发布可以覆盖嵌入版本号。
- `README.md`、`docs/README.md`、构建脚本、截图脚本、CI 和 release workflow 已统一使用 `dtui` 产物命名。
- 已补充 Linux / macOS / Windows 的构建、发布和下载说明。

验证：

- `git diff --check` 通过。
- `just build` 通过。

不属于 TASK-014：

- 不改源码目录名 `cmd/docker-tui`。
- 不迁移 Go module path。
- 不把当前仓库内所有历史文档一次性重写成 `dtui`，只更新对外交付和当前维护文档。

## TASK-022 最终范围

TASK-022 已按收缩后的交付范围完成。旧设计中的 G1～G10 是方案目标，不再直接作为完成计数；最终以以下 Phase 决议为准。

| Phase | 内容 | 状态 | 说明 |
|---|---|---|---|
| A | gpgme 仅进入 CGO 构建路径 | `done` | 非 CGO 依赖检查与双构建验证已自动化 |
| B | 补齐全部 DTO 与 `dto.Action` 迁移 | `cancelled` | 不在本轮实施，不作为 TASK-022 缺口 |
| C | Client 全量方法化与破坏性签名迁移 | `cancelled` | 保留现有可工作的调用形态 |
| D | 公共 API 路径匿名结构体具名化 | `done` | stats 类型与对应测试已落地 |
| D-2 | Docker / Podman mapper 目录拆分 | `cancelled` | 不在本轮实施，不作为 TASK-022 缺口 |
| E | `docker/service` 统一入口 | `cancelled` | 架构决策明确不引入该层 |
| F | CGO / non-CGO 构建、测试、vet 与依赖矩阵 | `done` | `scripts/verify-build-matrix.sh` 和 `just check` 已接入 |

最终架构决议：

- 上层统一依赖 `runtime.Engine`。
- Docker 与 Podman 保留独立 adapter / service 实现。
- 不增加 `internal/data/docker/service` 统一入口。
- 已取消 Phase 不得在其他任务中被描述为“未修复”或“待收尾”；若重新提出，必须建立新任务并重新验收。

## 已完成重点

### TASK-013：鼠标与小终端布局

- `Unsupported`：小于 60×14，仅显示终端尺寸错误。
- `Compact`：60×14 至 79×19，使用单行 header/footer 与紧凑内容区。
- `Standard`：至少 80×20，使用完整布局。
- 鼠标仅处理左键；列表移动游标，详情页滚动，其他区域不改变键盘路径。
- `UIConfig.EnableMouse` 控制鼠标模式，默认开启。
- 对应分级、布局、命中、边界和空列表行为均有测试。

### TASK-025：lint 治理

已清零或完成迁移：

- `ineffassign`
- `unused`
- `errcheck`
- `staticcheck`
- `gosimple`
- `gofmt`

统计从 9473 个问题收敛到 524 个。剩余 line length、复杂度、重复代码等属于持续治理基线，不影响 TASK-025 已完成状态；后续新增治理应建立独立任务。

相关提交：

```text
37dc3b3 chore(lint): remove unused code, fix unused/gosimple/ineffassign
7b0c756 fix(lint): handle errcheck across cleanup paths
c8bc551 fix(lint): resolve staticcheck SA1019 deprecations
5460224 refactor(audit): split Begin/Finish into Record builders
4fc0998 refactor(docker): extract error-mapping helpers and split literals
b909a87 refactor(podman): extract error-mapping helpers
7ba2320 refactor(tui): tighten containers.go for line-length
e0879b1 style: misc gofmt + small lint cleanups
```

## 依赖与执行顺序

```text
TASK-001..013, TASK-015..019, TASK-021, TASK-022, TASK-024, TASK-025 [done]

TASK-022 [done]
  `-> TASK-023 runtime/podman 文档与契约整理 [done]

独立后续：
  TASK-020 专有能力评估 [todo]
```

## 暂不实施

| 能力 | 原因 |
|---|---|
| Compose up / build / pull | 需要完整编排交互和进度模型 |
| Build / buildx | 参数、上下文、多平台和进度交互范围较大 |
| Podman pod / secret / kube | 等待 TASK-020 的产品边界决策 |
| Checkpoint / plugin | 用户场景优先级低于通用运维能力 |

## 完成定义

任务标记为 `done` 前必须满足：

1. 代码行为与已确认决策一致。
2. Docker / Podman 差异收敛在 runtime/data 层，不散落 runtime 字符串分支。
3. 覆盖主路径和失败路径；双运行时能力包含双侧测试或环境验证记录。
4. 配置在读取后校验，运行期不重复修正非法配置。
5. 状态结构通过方法维护不变量。
6. `git diff --check` 和 `just check` 通过。
7. README、架构、需求状态和本台账已同步。
8. 形成独立提交，提交信息包含任务或需求语义。
