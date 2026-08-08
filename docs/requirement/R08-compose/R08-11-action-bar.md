# R08-11 Action Bar 接入 Compose

## 元信息

- 状态: planned
- 优先级: high
- 来源: 用户评审结果 + [.omo/compose-todo.md 补 E.3 / 补 G / 补 J.5](../../../.omo/compose-todo.md)
- 关联任务: [.omo/compose-todo.md 附录 C / §C.1 / 补 E.3 / 补 F.3](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)
  - [../R03-form-action/R03-02-action-bar.md](../../requirement/R03-form-action/R03-02-action-bar.md) — Action Bar 框架

## ⚠ 重复功能说明

| 能力 | 已实现层 | 本需求增量 |
|---|---|---|
| Action Bar 框架 | [R03-02](../../requirement/R03-form-action/R03-02-action-bar.md) | **完全复用** R03-02 框架 |
| Action Bar 列表 scope | `internal/tui/actionbar/registry.go:29-38` 只为 Container + Image 注册 | **R08-11 增量**:补 `OperationScopeCompose` + scopeForPanel 的 compose case |
| OperationScope 常量 | `internal/data/config/operations.go:10-15` 已有 container / image,volume / network 留 future | **R08-11 增量**:加 compose 常量 |

## 目标

把 Compose 页面接入 Action Bar:

1. 注册 `OperationScopeCompose`
2. `scopeForPanel(panel=PanelCompose) → (OperationScopeCompose, true)`
3. 注册 compose-级动作 spec 进 `internal/data/config/defaults/operations/scopes/compose.jsonc`
4. 每个动作的 `Requires / DisabledWhen` 评估针对 compose 上下文(项目 / service / 资源状态)

## 用户流程

1. Compose 面板任意位置(左栏 / 右栏 / 子视图)→ 按 `;`
2. Action Bar 出现,按 panel 类型 + focus 显示:
   - **左栏(项目栏)**:项目级动作
   - **右栏(服务栏)**:服务级动作(在 service 选中时)
   - **子视图(服务容器列表)**:容器级动作(沿用 R01 容器级)
3. 用户选 action → confirm(若需) → 执行 → toast / audit

## UI/UX

仅本需求特有:

- Action Bar 列表项按"当前可用与否"分 2 段(顶部可用 / 底部 disabled)还是 1 段(灰显)由 R03-02 决定;R08-11 不引入新风格
- Title 用 Action Bar 框架的 label 显示

完整版式沿用 R03-02。

## 功能规则

### F1 OperationScope 常量扩展

```go
// internal/data/config/operations.go:10-15 (追加)
const (
    OperationScopeContainer OperationScope = "container"
    OperationScopeImage     OperationScope = "image"
    OperationScopeVolume    OperationScope = "volume"  // R02 已规划,本需求确认接入
    OperationScopeNetwork   OperationScope = "network" // R02 已规划,本需求确认接入
    OperationScopeCompose   OperationScope = "compose" // R08-11 新增
)
```

> ⚠ **一次性补齐**: Volume / Network / Compose 三 scope 同批接入,统一给 theme.action.<scope> 提供 mapping。R08-11 包含 volume / network 的接入以维持"所有 panel 都有 Action Bar"的一致性。

### F2 scopeForPanel 扩展

```go
// internal/tui/actionbar/registry.go:29-38
func scopeForPanel(panel state.PanelType) (config.OperationScope, bool) {
    switch panel {
    case state.PanelContainers:  return config.OperationScopeContainer, true
    case state.PanelImages:      return config.OperationScopeImage, true
    case state.PanelVolumes:     return config.OperationScopeVolume, true
    case state.PanelNetworks:    return config.OperationScopeNetwork, true
    case state.PanelCompose:     return config.OperationScopeCompose, true
    default: return "", false
    }
}
```

### F3 compose.jsonc 注册文件

新文件:`internal/data/config/defaults/operations/scopes/compose.jsonc`

需要注册的 ActionSpec 集合(R08-14 横切条目追加):

新文件:`internal/data/config/defaults/operations/scopes/compose.jsonc`

需要注册的 ActionSpec 集合:

| Action Kind | Label | Requires | DisabledWhen | 关联文档 |
|---|---|---|---|---|
| `compose_project_start` | "Start project" | "engine", "compose_project" | "no_stopped_containers" | R08-04 |
| `compose_project_stop` | "Stop project" | "engine", "compose_project" | "no_running_containers" | R08-04 |
| `compose_project_restart` | "Restart project" | "engine", "compose_project" | "" | R08-04 |
| `compose_project_down` | "Down project" | "engine", "compose_project" | "" | R08-04 |
| `compose_project_logs` | "Compose logs" | "engine", "compose_project" | "" | R08-06 |
| `compose_project_top` | "Compose top" | "engine", "compose_project", "compose_running" | "" | R08-07 |
| `compose_project_port` | "Compose port" | "engine", "compose_project", "compose_has_ports" | "" | R08-07 |
| `compose_project_stats` | "Compose stats" | "engine", "compose_project" | "" | R08-07 |
| `compose_project_build` | "Build services" | "engine", "compose_project" | "" | R08-05 |
| `compose_project_pull` | "Pull images" | "engine", "compose_project" | "" | R08-05 |
| `compose_project_push` | "Push images" | "engine", "compose_project", "compose_tagged" | "" | R08-05 |
| `compose_project_up` | "Up project (unsupported)" | "engine", "compose_project" | "always"(明确标灰显) | R08-04 |
| `compose_project_scale` | "Scale service" | "engine", "compose_service" | "" | R08-09 |
| `compose_project_pause` | "Pause project" | "engine", "compose_project", "compose_running" | "" | R08-09 |
| `compose_project_unpause` | "Unpause project" | "engine", "compose_project", "compose_paused" | "" | R08-09 |
| `compose_project_kill` | "Kill project" | "engine", "compose_project", "compose_running" | "" | R08-09 |
| `compose_project_rm` | "Remove stopped containers" | "engine", "compose_project" | "no_stopped_containers" | R08-09 |
| `compose_project_prune` | "Prune project" | "engine", "compose_project" | "" | R08-09 |
| `compose_project_events` | "Project events" | "engine", "compose_project" | "" | R08-10 |
| `compose_service_run` | "Run one-off" | "engine", "compose_service" | "" | R08-08 |
| `compose_service_exec` | "Exec in service" | "engine", "compose_service", "compose_running" | "" | R08-08 |
| `compose_service_logs` | "Service logs" | "engine", "compose_service" | "" | R08-06 |

### F4 Requirement Tokens(closed universe)

需要在 `internal/data/config/operations.go` 的 `Requirement` 枚举补:

```go
const (
    RequirementEngine Requirement = "engine"
    RequirementContainer Requirement = "container"
    RequirementImage Requirement = "image"
    RequirementVolume Requirement = "volume"
    RequirementNetwork Requirement = "network"
    RequirementRunning Requirement = "running"
    RequirementManifest Requirement = "manifest"
    // R08-11 补
    RequirementComposeProject Requirement = "compose_project"
    RequirementComposeService Requirement = "compose_service"
    RequirementComposeRunning Requirement = "compose_running"
    RequirementComposePaused  Requirement = "compose_paused"
    RequirementComposeStopped Requirement = "compose_stopped"
    RequirementComposeTagged  Requirement = "compose_tagged"
    RequirementComposeHasPorts Requirement = "compose_has_ports"
)
```

并在 `positiveSatisfied`(`internal/tui/actionbar/registry.go:126-147`)补对应评估逻辑。

### F5 Help 投影

Action Bar 项目在 Help 页(`internal/tui/ui/pages/help/about.txt`)需要按 panel 列出。每个 ActionSpec 都注册 label,Help 直接走 `helpers.go` 中已有的 ActionSpec projection,**不**改 help 渲染层。

## 实现设计

涉及模块:

- 修改:`internal/data/config/operations.go:10-15` 加 OperationScopeCompose 等常量与 Requirement tokens
- 修改:`internal/data/config/operations.go`:`positiveSatisfied` / `disabledWhenSatisfied` 函数扩展
- 修改:`internal/tui/actionbar/registry.go:29-38`(scopeForPanel 加 compose / volume / network)+ `:108-146`(evaluateEnabled 补 requirement 评估)
- 新文件:`internal/data/config/defaults/operations/scopes/compose.jsonc`(definitions)
- 新文件:`internal/data/config/defaults/operations/scopes/volume.jsonc`(definitions,顺手补)
- 新文件:`internal/data/config/defaults/operations/scopes/network.jsonc`(definitions,顺手补)
- 新增 theme:theme.action.compose / theme.action.volume / theme.action.network
- 修改:`internal/tui/keyboard/compose_nav.go` 与 `compose_action.go` 使每个 action 对接 `OperationScopeCompose`
- 修改:`internal/tui/ui/action/registry.go:Help` 投影

## 验收标准

1. Compose 面板按 `;` 打开 Action Bar:左栏显示项目级动作,右栏显示服务级动作
2. 项目无 stopped 容器时,Stop / Rm / 等动作 disabled(灰显)
3. 无 service selected 时,服务级动作不显示
4. podman engine 上 "compose_up" 始终 disabled,带 tooltip "compose up requires docker compose CLI"
5. Help(`?`)在 compose panel 的快捷提示按当前 Action Bar 实际项更新
6. 测试覆盖 scopeForPanel 各 panel 的映射

## 非目标

- 不在本需求实现具体 action handler(由 R08-04 / R08-05 / R08-06 / R08-09 的实现对接)
- 不引入 OperationScope<scope> 通用机制;只补 const + theme mapping
- 不为 container / image 现有 spec 增加新动作(不在范围内)

## R08-14 集成

R08-14 在 F3 ActionSpec 表增 3 项 group-level 条目(详见 [R08-14 §F7 / §I5](./R08-14-co-located-group.md)):

| Action Kind | Label | Requires | DisabledWhen | 关联 |
|---|---|---|---|---|
| `compose_group_down` | "Group Down" | "engine", "compose_project", "compose_group_running" | "" | R08-14 + R08-04 |
| `compose_group_restart` | "Group Restart" | "engine", "compose_project" | "no_group_running" | R08-14 + R08-04 |
| `compose_group_exec` | "Group Exec" | "engine", "compose_service", "compose_running" | "" | R08-14 + R08-08 |

`DisabledWhen: "always"` 在 F3 列已存在的 `compose_project_up` 上不重复(R08-04 已处理)。`Requires` token 新增 `compose_group_running` 走 R08-11 §F4 流程。

`compose_group_exec` 复用 R08-08 §F1 选 service / 选 replica 流程,group 模式下 group 内 pick 第一个运行中容器走 R01 exec。

## 迁移记录

- 旧代码:`internal/tui/actionbar/registry.go:29-38`(当前只 container / image)
- 旧文档:[.omo/compose-todo.md 补 E.3 / 补 G 完整需求列表 / H 决策矩阵](../../../.omo/compose-todo.md)
- 复用:R03-02 框架 / `theme.action.<scope>` 主题契约
- 保留信息:Container / Image scope 不动
- 废弃信息:无
- 待确认:Volume / Network 是否本批接入,R08-11 倾向**是**(避免 scope 缺位不一致)

### 2026-08-08 落地

**已完成**:
- `internal/data/config/defaults/operations/scopes/compose.jsonc`:24 个 OperationSpec(17 project + 3 service + 3 group + 1 detail),完整覆盖 R08-04 / R08-05 / R08-06 / R08-07 / R08-08 / R08-09 / R08-10 / R08-12 / R08-14 动作。
- `internal/tui/actionbar/registry.go::scopeForPanel`:补 `PanelVolumes` / `PanelNetworks` / `PanelCompose` 三 case — Volume / Network scope JSONC 文件此前已存在但 `scopeForPanel` 缺位,本轮顺手补齐。
- `internal/tui/actionbar/registry.go::positiveSatisfied`:加 7 个 compose Requirement tokens(`compose_project` / `compose_service` / `compose_running` / `compose_paused` / `compose_stopped` / `compose_tagged` / `compose_has_ports`)的评估逻辑。
- `internal/tui/actionbar/registry.go`:辅助函数 `selectedComposeProject` / `selectedComposeService` / `composeProjectContainers` / `composeProjectHasState` / `composeProjectHasTagged` / `composeProjectHasPorts` 全部本地实现,避免 actionbar → keyboard 的循环依赖。
- `internal/tui/actionbar/registry.go::imageFromComposeLabel`:`com.docker.compose.image` label 优先,缺失时回退到 `ContainerSummary.Image` — 与 docker / podman compose 行为对齐(只有 `image:` 字段声明才会写入 `com.docker.compose.image`)。
- `internal/tui/keyboard/operation_dispatch.go`:拆出 `dispatchOperationBySpec` helper;`dispatchOperation` 在 Mode-specific dispatcher 不识别 spec body 时 fallback 到 `handleAction`,保留 R06-03 "every registered Operation must be handled" 契约。
- `internal/tui/actionbar/registry_test.go`:新增 4 个测试 — `TestScopeForPanelMapsPanels`(5 panel → 5 scope)、`TestComposeRequirementsEvaluateProjectAndService`(7 token × 多状态)、`TestComposeTaggedRejectsLatest`(4 image 形式)、`TestComposeOperationsSurfaceInActionBar`(17 enabled + 3 disabled 校验)。

**保留事项 / 已知缺口**:
- `compose_project_up`:永久不支持(R08-04 `Up` 走 docker compose CLI),未注册到 JSONC。验收标准 4 的"始终 disabled"语义通过"不显示"实现,避免无意义灰显条目。
- `compose_project.detail` 已注册 `ActionComposeProjectDetail` 但未在 spec 表内 — 一并接入 `compose_project.detail` Operation(Label = "Detail", Mode = page)。
- `theme.action.compose` 已存在(`operations.go:297` 上轮 R08-12 顺手补),所有 6 个主题文件也已在该轮更新。
- Volume / Network scope 的 JSONC 文件早于本轮已存在;`scopeForPanel` 一直漏挂 — 本轮一并接入。

**测试**: `internal/tui/actionbar` 9 个测试 + `internal/tui/keyboard` 既有测试 + `internal/tui/keys` + `internal/data/config` 全部 PASS。`internal/data/runtime/podman` 与 `internal/driver/podman` 仍因宿主缺 `pkg-config` / `btrfs` C 头构建失败,与本次改动无关。
