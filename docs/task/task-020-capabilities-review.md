# TASK-020 专项能力评审决议

> TASK-020 只读能力评估 (read-only evaluation) — 由评估模型产出,供主模型 + 用户最终拍板。
> 本文件不改动任何 feature-design / bugfix-requirements / feature-todo-list;落地动作由用户 + 主模型执行。

## 0. 结论摘要

对 5 项**单 runtime 独有**能力逐项核对代码后,评估结论一致:**全部建议「不实现 / No-Go」**。

| # | 能力 | Runtime | 代码现状 | 业务价值 | 实现成本 | 双 runtime 一致性 | 测试成本 | **建议** |
|---|------|---------|----------|:--------:|:--------:|:------------------:|:--------:|----------|
| 1 | Podman `pod` | Podman 独有 | 全栈缺失 | 中 | 高 | ✗ 破坏 | 高 | **不实现 (No-Go)** |
| 2 | Podman `secret` | Podman 独有 | 全栈缺失 | 低-中 | 中-高 | ✗ 破坏 | 中-高 | **不实现 (No-Go)** |
| 3 | Podman `kube` | Podman 独有 | 全栈缺失 | 低 | 高 | ✗ 破坏 | 高 | **不实现 (No-Go)** |
| 4 | Docker `buildx` | Docker 独有 | 全栈缺失 | 低-中 | 高 | ✗ 破坏 | 高 | **不实现 (No-Go)** |
| 5 | Docker `context` | Docker 独有 | 已被 F2 运行时切换覆盖 | 低 | 中 | ✗ 冗余 | 中 | **不实现 (No-Go)** |

> 没有「纳入 (Go)」项,因此不产生任何新 BR / FR;全部应被记入 `docs/feature-design.md §3「不在范围」`。
> 唯一需要用户明确的口径是 #5 Docker context 的表述方式(见 §7)。

---

## 2. 评估前提(ground truth, 全部经代码核实)

1. **Engine 接口无专有能力服务** — 依据 `internal/data/runtime/engine.go:7-20`, 当前 `Engine` 接口的资源服务为 **8 个**:
   `Containers / Volumes / Networks / Images / ImageTransfers / Actions / Exec / Events`。
   **不存在任何 Pod / Secret / Kube / Build / Context 服务接口。**

2. **Capability 标识符清单** — `internal/data/runtime/capabilities.go` 中 `Capability` 常量全集为:
   `containers / images / volumes / networks / events / exec / native_filtering` 及细粒度
   `container.list.filter / image.list.filter / volume.list.filter / network.list.filter /
   events.filter / exec.resize / stats.stream`(**capabilities.go:20-36**)。
   **没有** `pod / secret / kube / buildx / build / context` 任一 capability 标识。

4. **两端 Capability 实现完全对称** — Docker adapter `Capabilities()`
   (`internal/data/runtime/docker/client.go:197-234`)与 Podman engine `Capabilities()`
   (`internal/data/runtime/podman/engine.go:51-88`)返回**同一组** capability, 仅 filter 标注 `Degraded`。
   说明目前没有任何「单 runtime 独有」能力被抽象层感知。

5. **Podman 低层驱动无 pod / secret / kube 端点** — `internal/driver/podman/urls.go` 全部路径常量
   (urls.go:10-151)只覆盖 `containers / images / volumes / networks / exec / version / events`,
   **不存在任何 `/libpod/pods` / `/libpod/secrets` / `/libpod/kube` / `/libpod/build` 端点**。

6. **Docker 适配层无 build / context** — `internal/data/runtime/docker/` 下涉及的 service 文件为
   `service_container / service_image / service_image_transfer / service_network / service_volume /
   service_action / service_event / service_exec`(见各 service_*.go),**没有** `build` 或 `context` service。

7. **Docker context 的实际对应物已存在 = F2 连接切换** — Docker 独有 `context` 概念在 dtui 的落地形态是
   **连接池 / 运行时切换**: `internal/data/runtime/connection_pool.go` + `config.runtime.connections`
   (见 README「Runtime auto-discovery / Multi-host」);快捷键 `F2` 已绑定切换
   (见 ROOT README Keyboard reference)。`docs/feature-design.md:100` 已明确把
   `docker context use` 映射为 `F2` ✅。

### 设计侧基线(Ground)

- `docs/feature-design.md §3`(feature-design.md:108-119)「不在范围」已包含 **Image Build**。
- `docs/feature-todo-list.md §3`(feature-todo-list.md:78-96)「已取消」已含:
  - **Build / buildx**(feature-todo-list.md:91, 历史决议);
  - **Podman pod / secret / kube**(feature-todo-list.md:92,「等待 TASK-020 决策」——即本评估)。
- `design/podman-capabilities-analysis.md` 初始建议(§3.2/§4)也把 pod / secret / kube / buildx
  归入**后批 / 不建议立做**(analysis.md:360-361, :371-379)。

---

## 3. Podman `pod`

### 现状
全栈缺失:
- 接口:`engine.go` 无 `PodService`。
- cap:`capabilities.go` 无 `pod` 标识。
- 驱动:`internal/driver/podman/urls.go` 无 `/lib/pods` 端点。
- UI:`podman-capabilities-analysis.md` 提议的 `PanelPods` 不存在。

### 逐项评估
- **业务价值 中**: 本机多容器分组运维, 对 Podman 用户有真实需求; 但 dtui 的 Compose 面板
  (基于 container label 聚合, 见 README「Compose projects」)已覆盖「把相关容器当一组看」的日常诉求,
  pod 的核心增量是 resource sharing(网络 / PID / IPC),TUI 只读排查场景价值有限。
- **实现成本 高**: 需新增 `PodService` 接口 + Docker 侧 stub(or 明确 Unsupported)+
  Podman 低层 `/lib/pods/{create,rm,start,stop,inspect}` 端点 + DTO mapper + `PanelPods` 页面
  + panel 注册 + keymap + i18n + Help/Footer。
- **双 runtime 一致性 ✗**: pod 是 Podman 独有,Docker 无对应物 ⇒ 必须引入 `Unsupported` 分支,
  违反 `feature-design.md:21(§2设计决策)#4 跨 runtime 一致性: 同一套 UI 同时支持 docker 与 podman`
  的核心原则。
- **测试成本 高**: 需 Podman-only integration 测试 + capability 检测(区分 `Available`/`Unsupported`) + 双 runtime 表驱动 + event 订阅兼容。

### 决议
**不实现 (No-Go)**,理由优先落地优先级:
1. 双 runtime 一致性被破坏(设计 §2.4);
2. 与已实现 Compose 面板场景重叠,新增边界价值有限;
3. 需重建一整套低层端点 + 面板, 成本高。
> 若未来用户明确提出, 须先立独立 FR, 再按新 BR 落地。

---

## 4. Podman `secret`

### 能力
全栈缺失: 无 `SecretService`(engine.go)、无 `secret` cap(capabilities.go)、无 `/lib/secrets` 端点(urls.go)、无 `PanelSecrets`(UI 未实现)。

### 逐项评估
- **业务价值 低-中**: K8s 风格 secret 管理, 但 Podman secret 与容器运行绑定较浅, 大多数 Podman 用户走 OCI/compose secrets; 纯 Kubernetes 管理场景使用有限。
- **实现成本 中**: 若只做 list/inspect/rm(读为主)成本略低; 若要 create + 注入到容器则需容器创建流程联动, 需覆盖容器 create 参数(当前容器无 create 路径)。
- **双 runtime 一致性 ✗**: secret 为 Podman 独有,Docker 的 secret 属 swarm(已排除)。UI 需大量 `Degraded/Unsupported` 分支。
- **测试成本 中-高**: 需要 secret 生命周期 integration + capability 分支 + 红色提示。

### 决议
**不实现 (No-Go)**:
- 价值不足以覆盖「双 runtime 一致性破坏 + 新 panel + 容器 create 联动」的成本;
- 与分析 §7.4「pod/secret 属 Podman 专有能力,当前阶段更适合先完成通用能力」一致。

---

## 5. Podman `kube` (Kubernetes YAML)

### 能力
全栈缺失: 无任何 YAML/Kube 解析入口; 驱动无 `/lib/kube` 端点; 产物下有 testdata(`podman 5.4.2` 的 info 等),但无 kube 相关。

### 逐项评估
- **业务价值 低**: 用户多直接走 `kubectl apply` / `podman kube play`(CLI 即足够); TUI 需要完整 YAML 解析/展示/编辑, 与 kubectl 差距太大。
- **实现成本 高**: 需 YAML schema (pod/container/service/volume) 解析 + 类型系统 + play/g generate 双向 + 结果反馈;独立成子系统。
- **双 runtime**: Podman 仅支持 podman-only YAML 子集(不支持原生 K8s), 与 Docker 完全不对齐。
- **测试 高**: YAML fixtures 全语法 + 语义校验 + 错误提示矩阵。

### 决议
**不实现 (No-Go)**。CLI/IDE 已充分覆盖, kubectl 是 K8s 迁移的自然通道; 不值得在 Read-only 管理 TUI 内重造 YAML 播放器。

---

## 6. Docker `buildx`

### 能力
全栈缺失: 无 `BuildService`(engine.go)、无 `build`/`buildx` cap、Docker 适配层无 build service、Podman 低层也无 build 路径(两端都不走引擎, 而是外部 CLI/CI)。且 `feature-design.md §3` 已把 **Image Build** 列为明确不实现(feature-design.md:114), `buildx` 正是 Image Build 的前端。

### 逐项评估
- **业务价值 中**: 构建能力常用, 但被 IDE / CI / CLI 抢占; TUI 内的长日志 / 多 build / platform matrix 交互空间差。
- **实现成本 高**: 需 Dockerfile 路径 + buildx 实例发现 + platform 矩阵 UI + 日志 streaming + 上下文; 且两端不对齐(需 Podman buildah 对应)。
- **双 runtime 一致性 ✗**: buildx 是 Docker 独有, Podman 走 `buildah`; 需要抽象 `BuildService` + 两侧差异。
- **测试 高**: integration 需真实 build(pull image + build), 慢且依赖网络。

### 决议
**不实现 (No-Go)**, 且与现有已取消的 **Image Build** 决议一致。若用户未来决定做构建, `buildx` 是其中一环, 需同步重新评估。本结论不承诺任何「build 状态查看」之类半成品(那仍依赖 build 能力)。

> 备注: task-020 待确认项「是否提供 build 状态查看」— 因 build 整体 No-Go, 状态查看无来源, 属未启用, 不应单项 pull 进来。

---

## 7. Docker `context`

### 能力现状
不是「全栈缺失」而是「已被现有形态覆盖」:
- dtui 使用 **F2 运行时切换(connection pool)** 取代 `docker context use`(feature-design.md:100 明确映射 ✅)。
- 连接来源为 `config.runtime.connections`(见 README Configuration「Runtime→connections」), 含 name / driver / endpoint / TLS。
- `internal/data/runtime/connection_pool.go` 已有探测与切换逻辑。

### 逐项评估
- **业务价值 低**: 用户已在 dtui 内通过 F2 在不同 daemon/context 间切换; 「读取 ~/.docker/config.json 展示所有 name context」仅是多一个只读面板, 价值低。
- **实现成本 中**: 读取 ~/.docker/config.json + 新 panel(或并入现有 runtime 切换对话框)。若要让 context 与 F2 联动(切换实随后更新 daemon), 则进入「连接切换」正在覆盖的同一区域, 需决策顺序。
- **双 runtime ✗**: `docker context` 是 Docker 概念; Podman context 指 rootful/rootless 切换, 语义不同 ⇒ 若实现需分支。
- **测试 中**: config.json 解析各平台路径, 易碎。

### 决议
**不实现 (No-Go)**。建议在决策落实时在 `feature-design.md §3` 记录为「**已由 F2 连接切换覆盖, 不单列**」, 并保留弹性, 可未来在运行时切换面板内顺带展示。

---

## 8. 跨能力共性风险(写给主模型/user)

1. **双 runtime 一致性**是本启动最高风险。五个候选**全部破坏** dtui 的核心原则(feature-design.md §2#4)。
   → 保守做法:全部记入「不在范围」, 保持 engine/capabilities 抽象极简。
2. **UI 复杂度 / Help / Footer 漂移**: 任何新面板、新 mode、新 key 都会迫使重排 layout 与 Help/Footer/导航文档; 当前 BR-039 Action Bar、BR-040 Dialog 等正在一般布局递进, 这能吸收新元素。
   ⇒ 即便将来决定复盘一, 也应在当前的页面动量化节奏之后。
3. **测试成本集中在「双 runtime + capability 检测」**: 五个能力需要实现 `Available/Unsupported/Degraded` 三态 + 双 runtime 表驱动, 而当前 tests 已假设两侧对称(见两处 `Capabilities()` 返回完全对齐)。引入任何独有能力都会打破现有对称测试假设。
4. **每次「纳入」必须挂新的 FR/BR 复选 + 重新评估。当前评估全 No-Go ⇒ 不产生新 BR。**

---

## 9. 需要用户 + 主模型确认的待定项

- [ ] 是否把以上五个 No-Go 全量写入 `feature-design.md §3`, 并将「Podman pod/secret/kube」从 feature-todo-list.md「已取消」变更为「TASK-020 正式决议 No-Go」?
- [ ] Docker `context` 的表述: 建议写「已由 F2 运行时切换覆盖 (已在 feature-design.md:100)」。
- [ ] (可选) 若后续有「容器多组分组」独立 FR 出现, 不要自动落回 Podman pod, 另开 FR 评估。

---

## 文件参考索引(所有结论均可在以下文件核证)

- `internal/data/runtime/engine.go:7-20` — `Engine` 接口仅读 `8` 个资源服务
- `internal/data/runtime/capabilities.go:20-36` — capability 标识全集(无 pod/secret/kube/buildx/context)
- `internal/data/runtime/docker/client.go:197-234` + `internal/data/runtime/podman/engine.go:51-87` — 对称 Capabilities
- `internal/data/runtime/docker/service_*.go` 列表 — 无 build/context
- `internal/data/runtime/podman/engine.go` + `internal/driver/podman/urls.go:1-151` — 无 pod/secret/kube/build
- `docs/feature-design.md:100` — `docker context use` ↔ `F2` ✅
- `docs/feature-design.md:108-119` + `docs/feature-todo-list.md:78-96` — 基线「不在范围」/「已取消」
- `design/podman-capabilities-analysis.md` — 初始分析与推荐顺序