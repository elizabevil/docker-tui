# R08-12 Keymap 注册表接入 Compose

## 元信息

- 状态: planned
- 优先级: high
- 来源: 用户评审结果 + [.omo/compose-todo.md 附录 C / 补 E.1 / 补 G.1 / 补 G.2](../../../.omo/compose-todo.md)
- 关联任务: [.omo/compose-todo.md §C.2 Ctrl+D bug / §2.3 子视图缺失 / 补 E.1 dead code 追踪](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)
  - [../R01-container/R01-01-advanced-ops.md](../../requirement/R01-container/R01-01-advanced-ops.md)(registry 模式参考)

## ⚠ 重复功能说明

| 能力 | 已实现层 | 本需求增量 |
|---|---|---|
| 全局 ActionBar / Help / Filter / Tab / Refresh 等 | `internal/tui/keys/registry.go:32-91` | **完全复用** |
| ActionSpec View 上下文(`containers` / `images` / etc.) | registry.go:35 | **R08-12 增量**:补 `"compose"` 和 `"compose-containers"` 双 view;学习镜像子视图模式(`image-containers`) |
| Ctrl+D 在 compose 走 Down(dead code bug) | 当前 dead code | **R08-12 增量**:修 `resourceDeleteOverrides`,让 ActionDelete 在 compose 跳过 |

## 目标

让 Compose 面板有完整可工作的快捷键 / Action 解析路径。具体:

1. 修复 Ctrl+D 在 compose 面板死代码 bug(必做 — R08-04 / R08-09 都依赖)
2. 修复 compose 子视图(`compose-containers`)上下文下的容器级动作(doContainer*) 不可达问题
3. 注册 compose-级新 action 到 `internal/tui/keys/registry.go`
4. 修 keymap YAML 配置(`config.KeymapConfig.Compose.*`)支持 compose-级动作覆盖

## 用户流程

R08-12 是纯基础设施需求,无直接用户流程。具体动作入口见 R08-04 / R08-05 / R08-06 / R08-09 / R08-11。

## UI/UX

不涉及(底层架构)。

## 功能规则

### F1 修复 Ctrl+D dead code

`internal/tui/keys/registry.go:146 resourceDeleteOverrides`:

```go
func (r *Resolver) resourceDeleteOverrides(context Context) bool {
    var action KeyAction
    switch context.View {
    case "containers":
        action = ActionContainerRemove
    case "images":
        action = ActionImageRemove
    case "volumes":
        action = ActionVolumeRemove
    case "networks":
        action = ActionNetworkRemove
    case "compose", "compose-containers": // R08-12 补
        action = ActionComposeProjectDown    // 路由到 compose down
    default:
        return false
    }
    return len(r.bindings.ByAction[action]) > 0
}
```

需新增 `KeyAction = "compose_project.down"` 到 action_id.go。

### F2 Compose 子视图容器级动作

镜像子视图模式:`image-containers` 复用 `containers` 视图(R01 已实现)。

Compose 子视图同样:让 `compose-containers` 在 `keyContext` 返回 `containers`(而不是单独的 `compose-containers`)

**或**:维持 `compose-containers`,但 `ActionContainer*` 注册时同时绑定 `containers` 和 `compose-containers` context。

倾向后者(版本演进更精确)。

```go
// registry.go 改 containers 列表
containers := []Context{
    {View: "containers"},
    {View: "image-containers"},
    {View: "compose-containers"},  // R08-12 补
}
```

### F3 新增 KeyAction 定义

```go
// internal/tui/keys/action_id.go 追加:
const (
    // ... existing
    ActionComposeProjectStart    KeyAction = "compose_project.start"
    ActionComposeProjectStop     KeyAction = "compose_project.stop"
    ActionComposeProjectRestart  KeyAction = "compose_project.restart"
    ActionComposeProjectDown     KeyAction = "compose_project.down"
    ActionComposeProjectLogs     KeyAction = "compose_project.logs"
    ActionComposeProjectTop      KeyAction = "compose_project.top"
    ActionComposeProjectPort     KeyAction = "compose_project.port"
    ActionComposeProjectStats    KeyAction = "compose_project.stats"
    ActionComposeProjectBuild    KeyAction = "compose_project.build"
    ActionComposeProjectPull     KeyAction = "compose_project.pull"
    ActionComposeProjectPush     KeyAction = "compose_project.push"
    ActionComposeProjectScale    KeyAction = "compose_project.scale"
    ActionComposeProjectPause    KeyAction = "compose_project.pause"
    ActionComposeProjectUnpause  KeyAction = "compose_project.unpause"
    ActionComposeProjectKill     KeyAction = "compose_project.kill"
    ActionComposeProjectRm       KeyAction = "compose_project.rm"
    ActionComposeProjectPrune    KeyAction = "compose_project.prune"
    ActionComposeProjectEvents   KeyAction = "compose_project.events"
    ActionComposeServiceRun      KeyAction = "compose_service.run"
    ActionComposeServiceExec     KeyAction = "compose_service.exec"
    ActionComposeServiceLogs     KeyAction = "compose_service.logs"
)
```

### F4 Registry 注册

```go
// internal/tui/keys/registry.go:38 补:
compose := []Context{{View: "compose"}}
composeContainers := []Context{{View: "compose-containers"}}

return []ActionSpec{
    // ... existing
    {ActionComposeProjectStart,    []string{KeyS},       compose}, // 已在 compose_nav.go:60,但需统一到 registry
    {ActionComposeProjectStop,     []string{KeyCtrlS},   compose},
    {ActionComposeProjectRestart,  []string{KeyCtrlR},   compose},
    {ActionComposeProjectDown,     []string{KeyCtrlD},   compose},
    {ActionComposeProjectLogs,     []string{KeyL},       compose},
    {ActionComposeProjectTop,      []string{KeyCtrlT},   compose},
    {ActionComposeProjectPort,     []string{KeyComma},   compose},
    {ActionComposeProjectStats,    nil,                  compose},  // 与 R01 容器 `m` 冲突,留 nil 由 keymap 配置
    {ActionComposeProjectBuild,    []string{KeyCtrlB},   compose},
    {ActionComposeProjectPull,     nil,                  compose},
    {ActionComposeProjectPush,     nil,                  compose},
    {ActionComposeProjectScale,    nil,                  compose},
    {ActionComposeProjectPause,    nil,                  compose},
    {ActionComposeProjectUnpause,  nil,                  compose},
    {ActionComposeProjectKill,     []string{KeyCtrlK},   compose},
    {ActionComposeProjectRm,       nil,                  compose},
    {ActionComposeProjectPrune,    nil,                  compose},
    {ActionComposeProjectEvents,   []string{KeyCtrlF3},  compose},
    {ActionComposeServiceRun,      nil,                  compose},
    {ActionComposeServiceExec,     []string{KeyE},       compose},
    {ActionComposeServiceLogs,     nil,                  compose},
    {ActionFilter,                 []string{KeySlash}, main},
    // ...
}
```

> ⚠ 与 R01 容器级冲突时,`context.VIEW = compose` 优先级 > `containers`,自然区分(因为 view 不同)。

### F5 handleAction 路由

`internal/tui/keyboard/actions.go:17-71`:

```go
func handleAction(action keys.KeyAction, m *state.AppModel, cmds []tea.Cmd) (...) {
    switch action {
    // ... existing
    case keys.ActionComposeProjectStart:
        return doComposeStart(m)
    case keys.ActionComposeProjectStop:
        return doComposeStop(m)
    case keys.ActionComposeProjectDown:
        return doComposeDown(m) // 路径已修通(R08-12 F1)
    // ... 所有 ActionComposeProject* 类似
    }
}
```

### F6 config.KeymapConfig.Compose 字段

`internal/data/config/types.go` — `KeymapConfig` 加 Compose 嵌套结构:

```go
type ComposeKeymap struct {
    Start    MultiKey `yaml:"start"`
    Stop     MultiKey `yaml:"stop"`
    Restart  MultiKey `yaml:"restart"`
    Down     MultiKey `yaml:"down"`
    Logs     MultiKey `yaml:"logs"`
    Top      MultiKey `yaml:"top"`
    Port     MultiKey `yaml:"port"`
    Stats    MultiKey `yaml:"stats"`
    Build    MultiKey `yaml:"build"`
    Pull     MultiKey `yaml:"pull"`
    Push     MultiKey `yaml:"push"`
    Scale    MultiKey `yaml:"scale"`
    Pause    MultiKey `yaml:"pause"`
    Unpause  MultiKey `yaml:"unpause"`
    Kill     MultiKey `yaml:"kill"`
    Rm       MultiKey `yaml:"rm"`
    Prune    MultiKey `yaml:"prune"`
    Events   MultiKey `yaml:"events"`
    Run      MultiKey `yaml:"run"`
    Exec     MultiKey `yaml:"exec"`
    // Service 级
    ServiceLogs MultiKey `yaml:"service_logs"`
}
```

加到 `KeymapConfig` 主结构:

```go
type KeymapConfig struct {
    Global    GlobalKeymap    `yaml:"global"`
    Container ContainerKeymap  `yaml:"container"`
    Image     ImageKeymap      `yaml:"image"`
    Volume    VolumeKeymap     `yaml:"volume"`
    Network   NetworkKeymap    `yaml:"network"`
    Navigation NavigationKeymap `yaml:"navigation"`
    Compose   ComposeKeymap   `yaml:"compose"`  // R08-12 补
}
```

并在 `internal/tui/keys/registry.go:174-193 configuredBindings` 映射。

### F7 Help 与 Footer 投影

Help(`internal/tui/ui/pages/help/about.txt`)+ Footer 投影需要补 compose-级 key 列表。

> UI overhaul 阶段统一治理;R08-12 仅保证 registry 正确性,Help 文本可在 R08-13 一起补。

## 实现设计

涉及模块:

- 修改:`internal/tui/keys/action_id.go` — 新增 22 个 KeyAction
- 修改:`internal/tui/keys/registry.go:32-91 / :146-161 / :174-193`
- 修改:`internal/tui/keyboard/actions.go:71` 加 case
- 修改:`internal/data/config/types.go`:`KeymapConfig` + `ComposeKeymap`
- 修改:`internal/data/config/defaults/keymap.jsonc`(默认 YAML 模板)
- 修改:`internal/tui/keyboard/compose_nav.go:53-71` 与 `compose_action.go` 旧调用逐步迁移到 `handleAction`
- 修改:`internal/tui/keyboard/keyboard.go:385-404 keyContext`(维持 `compose` / `compose-containers`)

## 验收标准

1. Ctrl+D 在 compose panel 触发达 `doComposeDown`(修复 dead code bug)
2. compose 子视图(`compose-containers`)下 `s / Ctrl+S / Ctrl+R / Ctrl+P` 等容器级动作可达(`doContainer*`)
3. 新增 22 个 KeyAction 全数注册,registry 测试覆盖
4. 用户自定义 `keymap.compose.*` YAML 配置生效
5. F1 修复后,compose panel Help 不再误导用户按 `Ctrl+D` 不会删容器

## 非目标

- 不引入 KeyAction 注册机制的元编程 / DSL(继续用结构化 Slice)
- 不改 R01 / R02 现有 action 定义
- 不实现 compose panel 的 vim 风格二级动作(`g g` 等)— 留 future enhancement

## R08-14 集成

R08-14 横切 3 个 KeyAction(详见 [R08-14 §I6](./R08-14-co-located-group.md)):

```go
// R08-12 F3 已存在的 KeyAction 列表追加
ActionComposeGroupDown    KeyAction = "compose_group.down"
ActionComposeGroupRestart KeyAction = "compose_group.restart"
ActionComposeGroupExec    KeyAction = "compose_group.exec"
```

绑定(R08-12 §F4):

| Action | 默认键 | 上下文 | 备注 |
|---|---|---|---|
| `compose_group.down` | `Shift+Ctrl+D` | `{View: "compose"}` + 光标在 services 列表且 project 含 ≥1 group | 调 `doComposeDown(m)` 但额外过滤 group 内 services |
| `compose_group.restart` | `Shift+Ctrl+R` | `{View: "compose"}` 同上 | 顺序 stop + start group 内 services |
| `compose_group.exec` | `Shift+Ctrl+E` | `{View: "compose"}` + group 含 ≥1 running | pick group 内第一个 running 容器走 R01 exec |

`Registry()` 与 `configuredBindings()` 同步注册,handler 路由在 `handleAction` 加 3 case,委派 `doComposeDown` / `doComposeStop+Start` / 复用 `m.Exec.ExecConn` 入口。

## 迁移记录

- 旧代码:`internal/tui/keys/registry.go:32-91`(当前不含 compose)
- 旧 bug:`doComposeDown` 在 compose panel 按 Ctrl+D 不可达
- 复用:R01 action_id.go 的 KeyAction 模式
- 保留信息:`compose_nav.go:53-71` 现有 KeyS / KeyCtrlS / KeyCtrlD / KeyL / KeyD 调用路径(逐步迁移到 handleAction)
- 废弃信息:无
- 待确认:ActionComposeProjectStats 默认键(`m` 是否冲突 — R01 容器 `m` 用于 stats,compose stats 选 nil 让用户配置)

### 2026-08-08 阶段 A 落地

**已完成**:
- `internal/tui/keys/action.go`:22 个 `ActionCompose*` + 3 个 `ActionComposeGroup*`(`R08-14` 横切)KeyAction 定义到位。
- `internal/tui/keys/registry.go`:
  - `compose` / `compose-containers` context 切片注册。
  - `{View: "compose-containers"}` 加入 `containers` 切片(F2:子视图复用容器动作)。
  - `resourceDeleteOverrides` 增加 `compose` / `compose-containers` case,F1 dead code bug 修通。
  - `configuredBindings` 接入 `config.ComposeKeymap` 全部字段。
- `internal/tui/keyboard/actions.go`:25 个 compose action 的 `handleAction` case 分派到 `doCompose*` 实现。
- `internal/tui/keyboard/compose_action.go`:`doComposeStart` / `doComposeStop` / `doComposeDown` / `doComposeLogs` 已存在并直接复用;其余 19 个 `doCompose*` 助手实现为 `composeActionPending` stub,toast 提示对应 R08 子需求阶段(R08-04 / R08-05 / R08-06 / R08-07 / R08-08 / R08-09 / R08-10 / R08-14)。
- `internal/data/config/types.go`:`ComposeKeymap` 新增 `Detail` 字段(对应 `KeyD` 在 compose 面板的 detail 跳转)。
- `internal/data/config/app_patch.go`:`ComposeKeymapPatch` 同步加 `Detail` 字段并在 apply 中复制。
- `internal/data/config/defaults/keymap.jsonc`:`compose` 段新增 `detail` 默认绑定。
- `internal/tui/keys/keys.go`:`KeyComma` / `KeyCtrlF3` / `KeyShiftCtrlD` / `KeyShiftCtrlR` / `KeyShiftCtrlE` 常量补齐。
- `internal/tui/keys/registry_test.go`:新增 5 个测试覆盖 F1/F2/F3/F4/F6。

**Driver/REST 参数完整性修复**(R08-02 / R08-15 旁路):
- `docker/compose_service.go::Stop`:`timeoutSec int` 参数原先以 `_ = timeoutSec` 丢弃,实现硬编码 10 秒。改为传入 `runContainerAction` 的 `timeoutSec` 参数,Stop / Restart 同步生效。
- `docker/compose_service.go::Stats`:原先命中容器后 `return ContainerStats{}` 空结构,数据字段未填充。改为调 `s.Client.containerStatsContext(ctx, c.ID)` 复用容器 stats 路径,得到 `CPUPercent` / `MemoryUsage` / `NetworkRx/Tx` 等完整快照。
- `docker/compose_service.go::Exec`:原先 `runContainerAction` 对 `dockerContainerExec` 直接 `continue`(纯 no-op),`command []string` 形参未使用。改为独立路径,定位服务首个容器后通过 `ContainerExecCreate` + `ContainerExecStart` 真正执行命令;同时删除 `dockerContainerExec` 枚举(已无引用)。
- `docker/compose_service.go::Events`:`ComposeEvent.Service` 字段未填充,与 podman 不一致。改为从 `ev.Actor.Attributes[runtimeapi.ComposeLabelService]` 取 service label。

**测试**: `./internal/tui/keys/... ./internal/tui/keyboard/... ./internal/data/runtime/docker/... ./internal/data/runtime/... ./internal/data/config/...` 全部 PASS。5 个新增 compose 断言 + 11 个 docker 既有测试全数通过。`internal/data/runtime/podman` 因宿主缺 `pkg-config` / `btrfs` C 头构建失败,与本次改动无关。

**保留迁移**:
- `compose_nav.go:53-71` 旧 switch 路径(`KeyS` / `KeyCtrlS` / `KeyCtrlD` / `KeyL` / `KeyD`)暂保留为冗余入口,后续可由 R08-11 Action Bar 接入统一收敛。
- `templates/template.golden.yml` 暂未渲染 `keymap.compose` 段,留作 R08-13 i18n / 文档同步时统一处理。
- `docker/compose_service.go::Up` 的 `UpOptions` 整组参数(`Build` / `ForceRecreate` / `NoDeps` / `NoBuild` / `Scale` ...)按 Q3 决策表保留 `engine API 单交互面` 约束,实现仅做 `start all`(逐容器 start),后续 R08-05 阶段可补 build/pull 与按 service 重建语义。
