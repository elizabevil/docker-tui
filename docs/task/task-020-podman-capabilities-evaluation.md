# TASK-020: Docker / Podman 专有能力评估

> 任务卡 / TASK-020

## 元信息

- **关联编号**:TASK-020
- **优先级**:P3
- **状态**:`todo`
- **依赖**:TASK-004(Docker / Podman 运行时识别)
- **关联文档**:[design/podman-capabilities-analysis.md](../../design/podman-capabilities-analysis.md)

## 目标

评估 Podman 独有(pod / secret / kube)与 Docker 独有(buildx / context)的产品边界,给出"纳入 / 暂缓 / 不实现"的明确决议,并把决议落到 [docs/feature-design.md §3](../feature-design.md) "不在范围"列表或新建 BR / FR 条目。

## 范围

- Podman pod / secret / kube:Podman 独有概念,Docker 不支持
- Docker buildx / context:Docker 独有,Podman 不支持
- 决定哪些能力纳入 dtui 路线图
- 决定哪些能力明确不实现

## 代码结构索引

### 必须读懂的文件

| 文件 | 作用 |
|---|---|
| `design/podman-capabilities-analysis.md` | 已有能力分析(可作为输入) |
| `internal/data/runtime/engine.go` | 当前 `Engine` 接口(11 个 service) |
| `internal/data/runtime/podman/engine.go` | Podman engine 能力(目前无 pod / secret / kube) |
| `internal/data/runtime/docker/service_*.go` | Docker 各 service(无 buildx / context) |
| `docs/feature-design.md §3` | 当前"不在范围"列表(基线) |

### 必须修改的文件

| 文件 | 改动 |
|---|---|
| `docs/feature-design.md §3` | 把决议(纳入 / 暂缓 / 不实现)写入"不在范围"列表 |
| `docs/feature-todo-list.md §3` | 更新"已取消"列表,新增"纳入" |
| `docs/feature-design.md §2.5` | 系统能力矩阵更新(如纳入了新能力) |

### 必须新增的文件

无(本任务为评估 + 决策落地,不涉及代码改动)。如决定实现某项能力,新建 BR 条目到 [docs/bugfix-requirements.md](../bugfix-requirements.md) 或 FR 条目到 [docs/feature-design.md](../feature-design.md)。

## 操作链路(评估流程)

```text
主模型 + 用户讨论
  1. 阅读 design/podman-capabilities-analysis.md(已有分析)
  2. 评估每项能力的:
     - 业务价值(用户日常使用频次)
     - 实现成本(runtime 适配 + UI 集成)
     - 测试复杂度(双 runtime + capability 检测)
     - 文档成本
  3. 输出决议:纳入(高/中/低) / 暂缓 / 不实现
  4. 决议落地:
     - "纳入" → 新建 BR / FR 条目到 feature-design.md §4 + bugfix-requirements.md + 本目录 task/
     - "暂缓" → 写入 feature-todo-list.md §3 已取消
     - "不实现" → 写入 feature-design.md §3 不在范围
  5. 通知 docs/README.md 更新
```

## 验收标准

- [ ] 设计文档 `feature-design.md §3` 反映最新决议(纳入 / 暂缓 / 不实现)。
- [ ] 设计文档 `feature-todo-list.md §3` 反映"暂缓"列表。
- [ ] 若有"纳入"项,新建对应 BR / FR 条目并入 task/ 目录。
- [ ] 决策文档化:每个能力的决议理由(`why this?` / `why not?`)。
- [ ] `podman-capabilities-analysis.md` 中"候选编号"列表与新决议保持一致。
- [ ] 双 runtime capability 检查策略明确(`Engine.Capabilities()` 返回每项支持情况)。

## 候选能力评估(初始)

| 能力 | 引擎 | 业务价值 | 实现成本 | 决议建议 |
|---|---|---|---|---|
| Podman pod | Podman 独有 | 中(开发者本地多容器组) | 中(新增 `PodService` + UI) | 暂缓(超出当前 5 资源 page 范围) |
| Podman secret | Podman 独有 | 中(K8s-like secrets) | 中-高(K8s 风格,需新 secret 类型) | 不实现(脱离 dtui 场景) |
| Podman kube (K8s YAML) | Podman 独有 | 低(用户多走 kubectl) | 高(整个 YAML 解析) | 不实现 |
| Docker buildx | Docker 独有 | 中(CI 常用,但 IDE/CI 替代) | 高(多 stage / 上下文 / 进度) | 不实现(见 feature-design.md §3) |
| Docker context | Docker 独有 | 低(已被 F2 运行时切换覆盖) | 低(读 ~/.docker/config.json 即可) | 不实现 |
| Docker compose v2 CLI | 共享 | 中 | 高(独立子进程,需要 shim) | 暂缓 |
| Docker swarm | Docker 独有 | 低(用户多走 K8s) | 高 | 不实现 |
| Podman machine | Podman 独有 | 低(WSL/macOS VM 管理) | 中 | 不实现 |

> 上述为初始评估,实际决议需主模型与用户确认。

## 风险

| 风险 | 说明 |
|---|---|
| **范围蔓延** | 评估可能产生"也要这个"的声音,需要严格按价值 / 成本排序 |
| **双 runtime 一致性** | 任何纳入能力都必须 Docker / Podman 都接或明确 capability 缺失处理 |
| **UI 复杂度** | 新能力意味着新 page / panel / Mode,需要重新评估布局 |
| **Help / Footer 漂移** | 任何新能力都需要文档同步 |

## 建议任务分解

1. **TASK-020-A:能力清单整理**
   - 列出所有候选能力(基于 `podman-capabilities-analysis.md`)
   - 每项标注:业务价值 / 实现成本 / 测试复杂度
2. **TASK-020-B:逐项评估 + 决议**
   - 主模型 + 用户讨论
   - 输出决议表(纳入 / 暂缓 / 不实现)
3. **TASK-020-C:决议落地**
   - 纳入项 → 新建 BR / FR
   - 暂缓项 → feature-todo-list.md §3
   - 不实现项 → feature-design.md §3
4. **TASK-020-D:文档同步**
   - docs/README.md + docs/feature-design.md 更新

## 待确认项

- [ ] Podman pod:是否纳入?若纳入,放在哪个 panel?(新建 `PanelPods`?)
- [ ] Podman secret:是否纳入?secret 注入到容器需要 runtime 支持,docker 也可模拟
- [ ] Docker buildx:即使不实现 Image Build,是否提供"build 状态查看"?(与 Image History 类似)
- [ ] Docker context:dtui 当前用 F2 runtime 切换已覆盖;是否需要"读取 ~/.docker/config.json 显示所有可用 context"?
- [ ] 评估结果是否需要单独的"决议纪要"文档,还是直接更新 feature-design.md?

## 参考

- [design/podman-capabilities-analysis.md](../../design/podman-capabilities-analysis.md):已有 Docker / Podman 能力对比分析
- [docs/feature-design.md §3](../feature-design.md):当前"不在范围"列表基线
- [docs/feature-todo-list.md §3](../feature-todo-list.md):当前"已取消"列表基线